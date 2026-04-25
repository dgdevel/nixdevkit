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

func TestCatBLimit200(t *testing.T) {
	setupTestRoot(t)
	var buf strings.Builder
	for i := 0; i < 300; i++ {
		if i > 0 {
			buf.WriteByte('\n')
		}
		fmt.Fprintf(&buf, "line%d", i)
	}
	path := filepath.Join(rootDir, "big.txt")
	os.WriteFile(path, []byte(buf.String()), 0644)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "cat-b",
			Arguments: map[string]interface{}{
				"path":       "/big.txt",
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
	got := result.Content[0].(mcp.TextContent).Text
	if !strings.HasPrefix(got, "Output cut at 200 lines starting from 0\n") {
		t.Errorf("expected cut prefix, got %q", got[:60])
	}
	lines := strings.Count(got, "\n")
	// 1 cut header + 200 data lines
	if lines != 201 {
		t.Errorf("expected 201 lines (1 header + 200 data), got %d", lines)
	}
}

func TestCatBLimit200WithOffset(t *testing.T) {
	setupTestRoot(t)
	var buf strings.Builder
	for i := 0; i < 300; i++ {
		if i > 0 {
			buf.WriteByte('\n')
		}
		fmt.Fprintf(&buf, "line%d", i)
	}
	path := filepath.Join(rootDir, "big.txt")
	os.WriteFile(path, []byte(buf.String()), 0644)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "cat-b",
			Arguments: map[string]interface{}{
				"path":       "/big.txt",
				"line_range": "50:",
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
	got := result.Content[0].(mcp.TextContent).Text
	if !strings.HasPrefix(got, "Output cut at 200 lines starting from 49\n") {
		t.Errorf("expected cut prefix with offset 49, got %q", got[:70])
	}
}
