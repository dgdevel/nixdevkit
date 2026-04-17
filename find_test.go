package main

import (
	"context"
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
