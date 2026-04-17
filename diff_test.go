package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestDiffIdentical(t *testing.T) {
	setupTestRoot(t)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "diff",
			Arguments: map[string]interface{}{
				"path1": "/file1.txt",
				"path2": "/file1.txt",
			},
		},
	}
	result, err := diffHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("diff returned error")
	}
	if textOf(t, result) != "" {
		t.Errorf("diff identical: expected empty, got %q", textOf(t, result))
	}
}

func TestDiffChanged(t *testing.T) {
	root := setupTestRoot(t)
	os.WriteFile(filepath.Join(root, "file2.txt"), []byte("hello\nearth\nfoo"), 0644)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "diff",
			Arguments: map[string]interface{}{
				"path1": "/file1.txt",
				"path2": "/file2.txt",
			},
		},
	}
	result, err := diffHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("diff returned error")
	}
	text := textOf(t, result)
	if !containsLine(text, "-world") || !containsLine(text, "+earth") {
		t.Errorf("diff changed: missing expected changes in %q", text)
	}
	if !containsLine(text, " hello") {
		t.Errorf("diff changed: missing context in %q", text)
	}
}

func TestDiffAdded(t *testing.T) {
	root := setupTestRoot(t)
	os.WriteFile(filepath.Join(root, "file2.txt"), []byte("hello\nworld\nfoo\nbar"), 0644)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "diff",
			Arguments: map[string]interface{}{
				"path1": "/file1.txt",
				"path2": "/file2.txt",
			},
		},
	}
	result, err := diffHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("diff returned error")
	}
	text := textOf(t, result)
	if !containsLine(text, "+bar") {
		t.Errorf("diff added: missing +bar in %q", text)
	}
}

func TestDiffRemoved(t *testing.T) {
	root := setupTestRoot(t)
	os.WriteFile(filepath.Join(root, "file2.txt"), []byte("hello\nfoo"), 0644)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "diff",
			Arguments: map[string]interface{}{
				"path1": "/file1.txt",
				"path2": "/file2.txt",
			},
		},
	}
	result, err := diffHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("diff returned error")
	}
	text := textOf(t, result)
	if !containsLine(text, "-world") {
		t.Errorf("diff removed: missing -world in %q", text)
	}
}

func containsLine(text, line string) bool {
	for _, l := range splitLines(text) {
		if l == line {
			return true
		}
	}
	return false
}
