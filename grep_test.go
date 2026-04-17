package main

import (
	"context"
	"sort"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestGrepStarGlob(t *testing.T) {
	setupTestRoot(t)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "grep",
			Arguments: map[string]interface{}{
				"pattern":  "hello",
				"pathspec": "*.txt",
			},
		},
	}
	result, err := grepHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("grep returned error")
	}
	text := textOf(t, result)
	if text != "file1.txt:1:hello" {
		t.Errorf("grep hello *.txt: got %q, want %q", text, "file1.txt:1:hello")
	}
}

func TestGrepGlobstar(t *testing.T) {
	setupTestRoot(t)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "grep",
			Arguments: map[string]interface{}{
				"pattern":  "hello",
				"pathspec": "**/*.txt",
			},
		},
	}
	result, err := grepHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("grep returned error")
	}
	lines := strings.Split(textOf(t, result), "\n")
	sort.Strings(lines)
	want := []string{"file1.txt:1:hello", "subdir/nested.txt:1:hello"}
	if len(lines) != len(want) {
		t.Fatalf("grep hello **/*.txt: got %v, want %v", lines, want)
	}
	for i, l := range lines {
		if l != want[i] {
			t.Errorf("grep hello **/*.txt: got %v, want %v", lines, want)
		}
	}
}

func TestGrepRegex(t *testing.T) {
	setupTestRoot(t)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "grep",
			Arguments: map[string]interface{}{
				"pattern":  "^w",
				"pathspec": "*.txt",
			},
		},
	}
	result, err := grepHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("grep returned error")
	}
	text := textOf(t, result)
	if text != "file1.txt:2:world" {
		t.Errorf("grep ^w: got %q, want %q", text, "file1.txt:2:world")
	}
}

func TestGrepNoMatch(t *testing.T) {
	setupTestRoot(t)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "grep",
			Arguments: map[string]interface{}{
				"pattern":  "nonexistent",
				"pathspec": "*.txt",
			},
		},
	}
	result, err := grepHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("grep returned error")
	}
	if textOf(t, result) != "" {
		t.Errorf("grep nonexistent: expected empty, got %q", textOf(t, result))
	}
}
