package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestPatchApplyDiff(t *testing.T) {
	root := setupTestRoot(t)
	os.WriteFile(filepath.Join(root, "file2.txt"), []byte("hello\nearth\nfoo"), 0644)

	diffReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "diff",
			Arguments: map[string]interface{}{
				"path1": "/file1.txt",
				"path2": "/file2.txt",
			},
		},
	}
	diffResult, err := diffHandler(context.Background(), diffReq)
	if err != nil {
		t.Fatal(err)
	}
	if diffResult.IsError {
		t.Fatal("diff returned error")
	}
	patchText := textOf(t, diffResult)

	patchReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "patch",
			Arguments: map[string]interface{}{
				"patch": patchText,
			},
		},
	}
	patchResult, err := patchHandler(context.Background(), patchReq)
	if err != nil {
		t.Fatal(err)
	}
	if patchResult.IsError {
		t.Fatalf("patch returned error: %s", textOf(t, patchResult))
	}

	data, _ := os.ReadFile(filepath.Join(root, "file1.txt"))
	if string(data) != "hello\nearth\nfoo\n" {
		t.Errorf("patch round-trip: got %q, want %q", string(data), "hello\nearth\nfoo\n")
	}
}

func TestPatchNoOp(t *testing.T) {
	setupTestRoot(t)

	patchText := "--- /file1.txt\n+++ /file1.txt\n@@ -1,3 +1,3 @@\n hello\n world\n foo\n"
	patchReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "patch",
			Arguments: map[string]interface{}{
				"patch": patchText,
			},
		},
	}
	result, err := patchHandler(context.Background(), patchReq)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("patch returned error")
	}
	data, _ := os.ReadFile(filepath.Join(rootDir, "file1.txt"))
	if string(data) != "hello\nworld\nfoo\n" {
		t.Errorf("patch noop: got %q", string(data))
	}
}

func TestPatchAddLines(t *testing.T) {
	setupTestRoot(t)

	patchText := "--- /file1.txt\n+++ /file1.txt\n@@ -1,3 +1,4 @@\n hello\n world\n+inserted\n foo\n"
	patchReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "patch",
			Arguments: map[string]interface{}{
				"patch": patchText,
			},
		},
	}
	result, err := patchHandler(context.Background(), patchReq)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("patch returned error")
	}
	data, _ := os.ReadFile(filepath.Join(rootDir, "file1.txt"))
	if string(data) != "hello\nworld\ninserted\nfoo\n" {
		t.Errorf("patch add: got %q", string(data))
	}
}

func TestPatchRemoveLines(t *testing.T) {
	setupTestRoot(t)

	patchText := "--- /file1.txt\n+++ /file1.txt\n@@ -1,3 +1,2 @@\n hello\n-world\n foo\n"
	patchReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "patch",
			Arguments: map[string]interface{}{
				"patch": patchText,
			},
		},
	}
	result, err := patchHandler(context.Background(), patchReq)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("patch returned error")
	}
	data, _ := os.ReadFile(filepath.Join(rootDir, "file1.txt"))
	if string(data) != "hello\nfoo\n" {
		t.Errorf("patch remove: got %q", string(data))
	}
}

func TestPatchInvalidFormat(t *testing.T) {
	setupTestRoot(t)

	patchReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "patch",
			Arguments: map[string]interface{}{
				"patch": "not a valid patch",
			},
		},
	}
	result, err := patchHandler(context.Background(), patchReq)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for invalid patch")
	}
}

func TestPatchWithTimestamp(t *testing.T) {
	setupTestRoot(t)

	patchText := "--- /file1.txt\t2024-01-15 10:30:00.000000000 +0000\n+++ /file1.txt\t2024-01-15 10:30:00.000000000 +0000\n@@ -1,3 +1,2 @@\n hello\n-world\n foo\n"
	patchReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "patch",
			Arguments: map[string]interface{}{
				"patch": patchText,
			},
		},
	}
	result, err := patchHandler(context.Background(), patchReq)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("patch with timestamp returned error: %s", textOf(t, result))
	}
	data, _ := os.ReadFile(filepath.Join(rootDir, "file1.txt"))
	if string(data) != "hello\nfoo\n" {
		t.Errorf("patch with timestamp: got %q", string(data))
	}
}

func TestPatchWithGitPrefixes(t *testing.T) {
	setupTestRoot(t)

	patchText := "--- a/file1.txt\n+++ b/file1.txt\n@@ -1,3 +1,2 @@\n hello\n-world\n foo\n"
	patchReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "patch",
			Arguments: map[string]interface{}{
				"patch": patchText,
			},
		},
	}
	result, err := patchHandler(context.Background(), patchReq)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("patch with git prefixes returned error: %s", textOf(t, result))
	}
	data, _ := os.ReadFile(filepath.Join(rootDir, "file1.txt"))
	if string(data) != "hello\nfoo\n" {
		t.Errorf("patch with git prefixes: got %q", string(data))
	}
}

func TestPatchRoundTripAdded(t *testing.T) {
	root := setupTestRoot(t)
	os.WriteFile(filepath.Join(root, "file2.txt"), []byte("hello\nworld\nfoo\nbar"), 0644)

	diffReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "diff",
			Arguments: map[string]interface{}{
				"path1": "/file1.txt",
				"path2": "/file2.txt",
			},
		},
	}
	diffResult, _ := diffHandler(context.Background(), diffReq)
	patchText := textOf(t, diffResult)

	patchReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "patch",
			Arguments: map[string]interface{}{
				"patch": patchText,
			},
		},
	}
	patchResult, err := patchHandler(context.Background(), patchReq)
	if err != nil {
		t.Fatal(err)
	}
	if patchResult.IsError {
		t.Fatalf("patch returned error: %s", textOf(t, patchResult))
	}

	data, _ := os.ReadFile(filepath.Join(root, "file1.txt"))
	if string(data) != "hello\nworld\nfoo\nbar\n" {
		t.Errorf("round-trip added: got %q", string(data))
	}
}

func TestPatchMultiFileDiff(t *testing.T) {
	setupTestRoot(t)

	patchText := "--- /file1.txt\n+++ /file1.txt\n@@ -1,3 +1,2 @@\n hello\n-world\n foo\n--- /other.txt\n+++ /other.txt\n@@ -200,1 +200,1 @@\n-old\n+new"
	patchReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "patch",
			Arguments: map[string]interface{}{
				"patch": patchText,
			},
		},
	}
	result, err := patchHandler(context.Background(), patchReq)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for multi-file patch with out-of-bounds hunk")
	}
}

func TestPatchMalformedHunkHeader(t *testing.T) {
	setupTestRoot(t)

	patchText := "--- /file1.txt\n+++ /file1.txt\n@@ bad\n hello\n"
	patchReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "patch",
			Arguments: map[string]interface{}{
				"patch": patchText,
			},
		},
	}
	result, err := patchHandler(context.Background(), patchReq)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for malformed hunk header")
	}
}

func TestPatchHunkStartZero(t *testing.T) {
	setupTestRoot(t)

	patchText := "--- /file1.txt\n+++ /file1.txt\n@@ -0,1 @@\n hello\n"
	patchReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "patch",
			Arguments: map[string]interface{}{
				"patch": patchText,
			},
		},
	}
	result, err := patchHandler(context.Background(), patchReq)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for hunk with start line 0")
	}
}

func TestPatchHunkStartBeyondFile(t *testing.T) {
	setupTestRoot(t)

	patchText := "--- /file1.txt\n+++ /file1.txt\n@@ -100,1 @@\n hello\n"
	patchReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "patch",
			Arguments: map[string]interface{}{
				"patch": patchText,
			},
		},
	}
	result, err := patchHandler(context.Background(), patchReq)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for hunk start beyond file length")
	}
}

func TestPatchHunkCountExceedsFile(t *testing.T) {
	setupTestRoot(t)

	patchText := "--- /file1.txt\n+++ /file1.txt\n@@ -1,100 @@\n hello\n"
	patchReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "patch",
			Arguments: map[string]interface{}{
				"patch": patchText,
			},
		},
	}
	result, err := patchHandler(context.Background(), patchReq)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for hunk count exceeding file length")
	}
}

func TestPatchNoHunksNoOp(t *testing.T) {
	setupTestRoot(t)

	patchText := "--- /file1.txt\n+++ /file1.txt\n"
	patchReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "patch",
			Arguments: map[string]interface{}{
				"patch": patchText,
			},
		},
	}
	result, err := patchHandler(context.Background(), patchReq)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("expected success for no-hunk patch, got error: %s", textOf(t, result))
	}
	data, _ := os.ReadFile(filepath.Join(rootDir, "file1.txt"))
	if string(data) != "hello\nworld\nfoo\n" {
		t.Errorf("file modified unexpectedly: got %q", string(data))
	}
}

func TestPatchMultipleHunksSameFile(t *testing.T) {
	root := setupTestRoot(t)
	os.WriteFile(filepath.Join(root, "multi.txt"), []byte("a\nb\nc\nd\ne"), 0644)

	patchText := "--- /multi.txt\n+++ /multi.txt\n@@ -2,1 +2,1 @@\n-b\n+B\n@@ -4,1 +4,1 @@\n-d\n+D"
	patchReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "patch",
			Arguments: map[string]interface{}{
				"patch": patchText,
			},
		},
	}
	result, err := patchHandler(context.Background(), patchReq)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("patch returned error: %s", textOf(t, result))
	}
	data, _ := os.ReadFile(filepath.Join(root, "multi.txt"))
	if string(data) != "a\nB\nc\nD\ne\n" {
		t.Errorf("multi-hunk: got %q, want %q", string(data), "a\nB\nc\nD\ne\n")
	}
}

func TestPatchRoundTripRemoved(t *testing.T) {
	root := setupTestRoot(t)
	os.WriteFile(filepath.Join(root, "file2.txt"), []byte("hello\nfoo"), 0644)

	diffReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "diff",
			Arguments: map[string]interface{}{
				"path1": "/file1.txt",
				"path2": "/file2.txt",
			},
		},
	}
	diffResult, _ := diffHandler(context.Background(), diffReq)
	patchText := textOf(t, diffResult)

	patchReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "patch",
			Arguments: map[string]interface{}{
				"patch": patchText,
			},
		},
	}
	patchResult, err := patchHandler(context.Background(), patchReq)
	if err != nil {
		t.Fatal(err)
	}
	if patchResult.IsError {
		t.Fatalf("patch returned error: %s", textOf(t, patchResult))
	}

	data, _ := os.ReadFile(filepath.Join(root, "file1.txt"))
	if string(data) != "hello\nfoo\n" {
		t.Errorf("round-trip removed: got %q", string(data))
	}
}
