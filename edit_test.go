package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestEditReplaceMiddle(t *testing.T) {
	setupTestRoot(t)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "edit",
			Arguments: map[string]interface{}{
				"path":       "/file1.txt",
				"line_range": "2:2",
				"content":    "REPLACED",
			},
		},
	}
	result, err := editHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("edit returned error")
	}
	data, _ := os.ReadFile(filepath.Join(rootDir, "file1.txt"))
	if string(data) != "hello\nREPLACED\nfoo" {
		t.Errorf("edit: got %q, want %q", string(data), "hello\nREPLACED\nfoo")
	}
}

func TestEditDelete(t *testing.T) {
	setupTestRoot(t)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "edit",
			Arguments: map[string]interface{}{
				"path":       "/file1.txt",
				"line_range": "2:3",
				"content":    "",
			},
		},
	}
	result, err := editHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("edit returned error")
	}
	data, _ := os.ReadFile(filepath.Join(rootDir, "file1.txt"))
	if string(data) != "hello" {
		t.Errorf("edit delete: got %q, want %q", string(data), "hello")
	}
}

func TestEditPrepend(t *testing.T) {
	setupTestRoot(t)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "edit",
			Arguments: map[string]interface{}{
				"path":       "/file1.txt",
				"line_range": "0:0",
				"content":    "new_top",
			},
		},
	}
	result, err := editHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("edit returned error")
	}
	data, _ := os.ReadFile(filepath.Join(rootDir, "file1.txt"))
	if string(data) != "new_top\nhello\nworld\nfoo" {
		t.Errorf("edit prepend: got %q, want %q", string(data), "new_top\nhello\nworld\nfoo")
	}
}

func TestEditMultiline(t *testing.T) {
	setupTestRoot(t)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "edit",
			Arguments: map[string]interface{}{
				"path":       "/file1.txt",
				"line_range": "2:2",
				"content":    "a\nb\nc",
			},
		},
	}
	result, err := editHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("edit returned error")
	}
	data, _ := os.ReadFile(filepath.Join(rootDir, "file1.txt"))
	if string(data) != "hello\na\nb\nc\nfoo" {
		t.Errorf("edit multiline: got %q, want %q", string(data), "hello\na\nb\nc\nfoo")
	}
}

func TestEditEscape(t *testing.T) {
	setupTestRoot(t)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "edit",
			Arguments: map[string]interface{}{
				"path":       "../../etc/passwd",
				"line_range": ":",
				"content":    "x",
			},
		},
	}
	result, err := editHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for path escape")
	}
}
