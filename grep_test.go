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

func TestGrepLimit200(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < 300; i++ {
		os.WriteFile(filepath.Join(root, fmt.Sprintf("file%03d.txt", i)), []byte("match"), 0644)
	}
	rootDir = root

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "grep",
			Arguments: map[string]interface{}{
				"pattern":  "match",
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
	got := textOf(t, result)
	if !strings.HasPrefix(got, "Output cut at 200 lines, refine the search pattern\n") {
		t.Fatalf("expected cut prefix, got %q", got[:80])
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 201 {
		t.Errorf("expected 201 lines, got %d", len(lines))
	}
}
