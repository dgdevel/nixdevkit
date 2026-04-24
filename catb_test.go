package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestCatB(t *testing.T) {
	setupTestRoot(t)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "cat-b",
			Arguments: map[string]interface{}{
				"path":       "/file1.txt",
				"line_range": ":",
			},
		},
	}
	result, err := catbHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("cat-b returned error")
	}
	want := "     1\thello\n     2\tworld\n     3\tfoo\n"
	got := result.Content[0].(mcp.TextContent).Text
	if got != want {
		t.Errorf("cat-b full: got %q, want %q", got, want)
	}
}

func TestCatBRange(t *testing.T) {
	setupTestRoot(t)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "cat-b",
			Arguments: map[string]interface{}{
				"path":       "/file1.txt",
				"line_range": "2:2",
			},
		},
	}
	result, err := catbHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("cat-b returned error")
	}
	// line_range "2:2" → parseLineRange returns from=1, to=2 → only index 1 ("world")
	want := "     2\tworld\n"
	got := result.Content[0].(mcp.TextContent).Text
	if got != want {
		t.Errorf("cat-b range: got %q, want %q", got, want)
	}
}

func TestCatBEmptyLines(t *testing.T) {
	setupTestRoot(t)
	path := filepath.Join(rootDir, "emptylines.txt")
	os.WriteFile(path, []byte("a\n\nb"), 0644)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "cat-b",
			Arguments: map[string]interface{}{
				"path":       "/emptylines.txt",
				"line_range": ":",
			},
		},
	}
	result, err := catbHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("cat-b returned error")
	}
	// "a\n\nb" splits to ["a", "", "b"], line 2 is empty → just "\n"
	want := "     1\ta\n\n     3\tb\n"
	got := result.Content[0].(mcp.TextContent).Text
	if got != want {
		t.Errorf("cat-b empty lines: got %q, want %q", got, want)
	}
}

func TestCatBEscape(t *testing.T) {
	setupTestRoot(t)
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "cat-b",
			Arguments: map[string]interface{}{
				"path":       "../../etc/passwd",
				"line_range": ":",
			},
		},
	}
	result, err := catbHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for path escape")
	}
}
