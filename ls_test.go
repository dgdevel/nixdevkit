package main

import (
	"context"
	"sort"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestLsRoot(t *testing.T) {
	setupTestRoot(t)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "ls",
			Arguments: map[string]interface{}{
				"path": "/",
			},
		},
	}
	result, err := lsHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("ls returned error")
	}

	lines := strings.Split(textOf(t, result), "\n")
	sort.Strings(lines)
	want := []string{"file1.txt", "subdir/"}
	if len(lines) != len(want) {
		t.Fatalf("ls /: got %v, want %v", lines, want)
	}
	for i, l := range lines {
		if l != want[i] {
			t.Errorf("ls /: got %v, want %v", lines, want)
		}
	}
}

func TestLsSubdir(t *testing.T) {
	setupTestRoot(t)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "ls",
			Arguments: map[string]interface{}{
				"path": "/subdir",
			},
		},
	}
	result, err := lsHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	text := textOf(t, result)
	if text != "subdir/nested.txt" {
		t.Errorf("ls /subdir: got %q, want %q", text, "subdir/nested.txt")
	}
}

func TestLsEscape(t *testing.T) {
	setupTestRoot(t)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "ls",
			Arguments: map[string]interface{}{
				"path": "../../etc",
			},
		},
	}
	result, err := lsHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for path escape")
	}
}

func TestLsNotFound(t *testing.T) {
	setupTestRoot(t)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "ls",
			Arguments: map[string]interface{}{
				"path": "/nonexistent",
			},
		},
	}
	result, err := lsHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for nonexistent path")
	}
}
