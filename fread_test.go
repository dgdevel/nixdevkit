package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nixdevkit/internal/cfg"

	"github.com/mark3labs/mcp-go/mcp"
)

func freadReq(path, lineRange string) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "fread",
			Arguments: map[string]interface{}{
				"path":       path,
				"line_range": lineRange,
			},
		},
	}
}

func TestFreadBasic(t *testing.T) {
	setupTestRoot(t)

	result, err := freadHandler(context.Background(), freadReq("/file1.txt", ":"))
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("fread returned error")
	}
	want := "----- /file1.txt - line from 1 to 3 -----\nhello\nworld\nfoo\n----- /file1.txt - EOF -----\n"
	got := result.Content[0].(mcp.TextContent).Text
	if got != want {
		t.Errorf("fread basic: got %q, want %q", got, want)
	}
}

func TestFreadRange(t *testing.T) {
	setupTestRoot(t)

	result, err := freadHandler(context.Background(), freadReq("/file1.txt", "2:2"))
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("fread returned error")
	}
	want := "----- /file1.txt - line from 2 to 2 -----\nworld\n----- /file1.txt - EOF -----\n"
	got := result.Content[0].(mcp.TextContent).Text
	if got != want {
		t.Errorf("fread range: got %q, want %q", got, want)
	}
}

func TestFreadEmptyLines(t *testing.T) {
	setupTestRoot(t)
	os.WriteFile(filepath.Join(rootDir, "emptylines.txt"), []byte("a\n\nb"), 0644)

	result, err := freadHandler(context.Background(), freadReq("/emptylines.txt", ":"))
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("fread returned error")
	}
	want := "----- /emptylines.txt - line from 1 to 3 -----\na\n\nb\n----- /emptylines.txt - EOF -----\n"
	got := result.Content[0].(mcp.TextContent).Text
	if got != want {
		t.Errorf("fread empty lines: got %q, want %q", got, want)
	}
}

func TestFreadPathEscape(t *testing.T) {
	setupTestRoot(t)

	result, err := freadHandler(context.Background(), freadReq("../../etc/passwd", ":"))
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for path escape")
	}
}

func TestFreadConfigPath(t *testing.T) {
	setupTestRoot(t)
	cfgDir := cfg.DirPath(rootDir)
	os.MkdirAll(cfgDir, 0755)
	os.WriteFile(filepath.Join(cfgDir, "config.ini"), []byte("[core]\nreadonly=true\n"), 0644)

	result, err := freadHandler(context.Background(), freadReq("/.nixdevkit/config.ini", ":"))
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for config path")
	}
}

func TestFreadOutOfRange(t *testing.T) {
	setupTestRoot(t)

	result, err := freadHandler(context.Background(), freadReq("/file1.txt", "100:"))
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("fread returned error")
	}
	got := result.Content[0].(mcp.TextContent).Text
	if got != "" {
		t.Errorf("fread out of range: expected empty, got %q", got)
	}
}

func TestFreadNoTransform(t *testing.T) {
	setupTestRoot(t)
	os.WriteFile(filepath.Join(rootDir, "raw.txt"), []byte("a\tb\tc\nhello   \nworld  "), 0644)

	result, err := freadHandler(context.Background(), freadReq("/raw.txt", ":"))
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("fread returned error")
	}
	want := "----- /raw.txt - line from 1 to 3 -----\na\tb\tc\nhello   \nworld  \n----- /raw.txt - EOF -----\n"
	got := result.Content[0].(mcp.TextContent).Text
	if got != want {
		t.Errorf("fread no transform: got %q, want %q", got, want)
	}
}

func TestFreadBlocks(t *testing.T) {
	setupTestRoot(t)
	var buf strings.Builder
	for i := 0; i < 75; i++ {
		if i > 0 {
			buf.WriteByte('\n')
		}
		fmt.Fprintf(&buf, "line%d", i)
	}
	os.WriteFile(filepath.Join(rootDir, "blocked.txt"), []byte(buf.String()), 0644)

	result, err := freadHandler(context.Background(), freadReq("/blocked.txt", ":"))
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("fread returned error")
	}
	got := result.Content[0].(mcp.TextContent).Text

	if !strings.Contains(got, "----- /blocked.txt - line from 1 to 30 -----\n") {
		t.Error("missing block 1 header")
	}
	if !strings.Contains(got, "----- /blocked.txt - line from 31 to 60 -----\n") {
		t.Error("missing block 2 header")
	}
	if !strings.Contains(got, "----- /blocked.txt - line from 61 to 75 -----\n") {
		t.Error("missing block 3 header")
	}
	if !strings.Contains(got, "----- /blocked.txt - EOF -----\n") {
		t.Error("missing EOF marker")
	}
	if !strings.Contains(got, "line0\n") {
		t.Error("missing first line content")
	}
	if !strings.Contains(got, "line74\n") {
		t.Error("missing last line content")
	}
}

func TestFreadCustomBlockSize(t *testing.T) {
	setupTestRoot(t)

	cfgDir := cfg.DirPath(rootDir)
	os.MkdirAll(cfgDir, 0755)
	os.WriteFile(filepath.Join(cfgDir, "config.ini"), []byte("[core]\nfread_block_size=10\n"), 0644)

	var buf strings.Builder
	for i := 0; i < 25; i++ {
		if i > 0 {
			buf.WriteByte('\n')
		}
		fmt.Fprintf(&buf, "line%d", i)
	}
	os.WriteFile(filepath.Join(rootDir, "custom.txt"), []byte(buf.String()), 0644)

	result, err := freadHandler(context.Background(), freadReq("/custom.txt", ":"))
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("fread returned error")
	}
	got := result.Content[0].(mcp.TextContent).Text

	if !strings.Contains(got, "----- /custom.txt - line from 1 to 10 -----\n") {
		t.Error("missing block 1 header with custom size")
	}
	if !strings.Contains(got, "----- /custom.txt - line from 11 to 20 -----\n") {
		t.Error("missing block 2 header with custom size")
	}
	if !strings.Contains(got, "----- /custom.txt - line from 21 to 25 -----\n") {
		t.Error("missing block 3 header with custom size")
	}
}

func TestFreadExactBlockBoundary(t *testing.T) {
	setupTestRoot(t)

	cfgDir := cfg.DirPath(rootDir)
	os.MkdirAll(cfgDir, 0755)
	os.WriteFile(filepath.Join(cfgDir, "config.ini"), []byte("[core]\nfread_block_size=3\n"), 0644)

	os.WriteFile(filepath.Join(rootDir, "exact.txt"), []byte("a\nb\nc\nd\ne\nf"), 0644)

	result, err := freadHandler(context.Background(), freadReq("/exact.txt", ":"))
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("fread returned error")
	}
	want := "----- /exact.txt - line from 1 to 3 -----\na\nb\nc\n----- /exact.txt - line from 4 to 6 -----\nd\ne\nf\n----- /exact.txt - EOF -----\n"
	got := result.Content[0].(mcp.TextContent).Text
	if got != want {
		t.Errorf("fread exact boundary: got %q, want %q", got, want)
	}
}

func TestFreadNestedPath(t *testing.T) {
	setupTestRoot(t)

	result, err := freadHandler(context.Background(), freadReq("/subdir/nested.txt", ":"))
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("fread returned error")
	}
	want := "----- /subdir/nested.txt - line from 1 to 2 -----\nhello\nbar\n----- /subdir/nested.txt - EOF -----\n"
	got := result.Content[0].(mcp.TextContent).Text
	if got != want {
		t.Errorf("fread nested: got %q, want %q", got, want)
	}
}
