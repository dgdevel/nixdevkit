package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestFindStarGlob(t *testing.T) {
	setupTestRoot(t)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "find",
			Arguments: map[string]interface{}{
				"pattern": "*.txt",
			},
		},
	}
	result, err := findHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("find returned error")
	}
	text := textOf(t, result)
	if text != "file1.txt" {
		t.Errorf("find *.txt: got %q, want %q", text, "file1.txt")
	}
}

func TestFindGlobstar(t *testing.T) {
	setupTestRoot(t)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "find",
			Arguments: map[string]interface{}{
				"pattern": "**/*.txt",
			},
		},
	}
	result, err := findHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("find returned error")
	}
	lines := strings.Split(textOf(t, result), "\n")
	sort.Strings(lines)
	want := []string{"file1.txt", "subdir/nested.txt"}
	if len(lines) != len(want) {
		t.Fatalf("find **/*.txt: got %v, want %v", lines, want)
	}
	for i, l := range lines {
		if l != want[i] {
			t.Errorf("find **/*.txt: got %v, want %v", lines, want)
		}
	}
}

func TestFindDir(t *testing.T) {
	setupTestRoot(t)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "find",
			Arguments: map[string]interface{}{
				"pattern": "sub*",
			},
		},
	}
	result, err := findHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("find returned error")
	}
	text := textOf(t, result)
	if text != "subdir/" {
		t.Errorf("find sub*: got %q, want %q", text, "subdir/")
	}
}

func TestFindNoMatch(t *testing.T) {
	setupTestRoot(t)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "find",
			Arguments: map[string]interface{}{
				"pattern": "*.xyz",
			},
		},
	}
	result, err := findHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("find returned error")
	}
	if textOf(t, result) != "" {
		t.Errorf("find *.xyz: expected empty, got %q", textOf(t, result))
	}
}

func TestFindLimit200(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < 300; i++ {
		os.WriteFile(filepath.Join(root, fmt.Sprintf("file%03d.txt", i)), []byte("x"), 0644)
	}
	rootDir = root

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "find",
			Arguments: map[string]interface{}{
				"pattern": "*.txt",
			},
		},
	}
	result, err := findHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("find returned error")
	}
	got := textOf(t, result)
	if !strings.HasPrefix(got, "Output cut at 200 lines, refine the search pattern\n") {
		t.Fatalf("expected cut prefix, got %q", got[:80])
	}
	lines := strings.Split(got, "\n")
	// 1 cut header line + 200 data lines
	if len(lines) != 201 {
		t.Errorf("expected 201 lines, got %d", len(lines))
	}
}
