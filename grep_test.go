package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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
	want := "file1.txt:1:hello\nfile1.txt:2:world\nfile1.txt:3:foo"
	if text != want {
		t.Errorf("grep hello *.txt: got %q, want %q", text, want)
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
	text := textOf(t, result)
	want := "file1.txt:1:hello\nfile1.txt:2:world\nfile1.txt:3:foo\n--\nsubdir/nested.txt:1:hello\nsubdir/nested.txt:2:bar"
	if text != want {
		t.Errorf("grep hello **/*.txt: got %q, want %q", text, want)
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
	want := "file1.txt:1:hello\nfile1.txt:2:world\nfile1.txt:3:foo"
	if text != want {
		t.Errorf("grep ^w: got %q, want %q", text, want)
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

func TestGrepLimit500(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < 600; i++ {
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
	if !strings.HasPrefix(got, "Output cut at 500 lines, refine the search pattern\n") {
		t.Fatalf("expected cut prefix, got %q", got[:80])
	}
	contentLines := 0
	for _, line := range strings.Split(got, "\n") {
		if line != "Output cut at 500 lines, refine the search pattern" && line != "--" && line != "" {
			contentLines++
		}
	}
	if contentLines != 500 {
		t.Errorf("expected 500 content lines, got %d", contentLines)
	}
}

func TestGrepContextOverlap(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "test.txt"), []byte("a\nb\nc\nmatch1\ne\nf\ng\nmatch2\ni\nj"), 0644)
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
	text := textOf(t, result)
	if strings.Contains(text, "--") {
		t.Errorf("expected no -- separator for overlapping context, got %q", text)
	}
	lines := strings.Split(text, "\n")
	if lines[0] != "test.txt:1:a" {
		t.Errorf("expected first line to be context from start, got %q", lines[0])
	}
}

func TestGrepVisualTab(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "tabs.txt"), []byte("a\tb\nmatch\tc"), 0644)
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
	text := textOf(t, result)
	if !strings.Contains(text, "a\tb") {
		t.Errorf("expected raw tab in context line, got %q", text)
	}
	if !strings.Contains(text, "match\tc") {
		t.Errorf("expected raw tab in match line, got %q", text)
	}
}

func TestGrepVisualTrailingSpace(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "spaces.txt"), []byte("match   \nother"), 0644)
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
	text := textOf(t, result)
	if !strings.Contains(text, "match   ") {
		t.Errorf("expected raw trailing spaces, got %q", text)
	}
}
