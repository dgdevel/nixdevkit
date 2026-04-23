package main

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestReadFile(t *testing.T) {
	setupTestRoot(t)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "read",
			Arguments: map[string]interface{}{
				"path":       "/file1.txt",
				"line_range": ":",
			},
		},
	}
	result, err := readHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("read returned error")
	}
	if textOf(t, result) != "hello\nworld\nfoo" {
		t.Errorf("read /file1.txt: got %q, want %q", textOf(t, result), "hello\nworld\nfoo")
	}
}

func TestReadNested(t *testing.T) {
	setupTestRoot(t)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "read",
			Arguments: map[string]interface{}{
				"path":       "/subdir/nested.txt",
				"line_range": ":",
			},
		},
	}
	result, err := readHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("read returned error")
	}
	if textOf(t, result) != "hello\nbar" {
		t.Errorf("read /subdir/nested.txt: got %q, want %q", textOf(t, result), "hello\nbar")
	}
}

func TestReadEscape(t *testing.T) {
	setupTestRoot(t)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "read",
			Arguments: map[string]interface{}{
				"path":       "../../etc/passwd",
				"line_range": ":",
			},
		},
	}
	result, err := readHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for path escape")
	}
}

func TestReadDir(t *testing.T) {
	setupTestRoot(t)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "read",
			Arguments: map[string]interface{}{
				"path":       "/subdir",
				"line_range": ":",
			},
		},
	}
	result, err := readHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error when reading a directory")
	}
}

func TestReadLineRangeFrom(t *testing.T) {
	setupTestRoot(t)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "read",
			Arguments: map[string]interface{}{
				"path":       "/file1.txt",
				"line_range": "2:",
			},
		},
	}
	result, err := readHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("read returned error")
	}
	if textOf(t, result) != "world\nfoo" {
		t.Errorf("read 2:: got %q, want %q", textOf(t, result), "world\nfoo")
	}
}

func TestReadLineRangeTo(t *testing.T) {
	setupTestRoot(t)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "read",
			Arguments: map[string]interface{}{
				"path":       "/file1.txt",
				"line_range": ":2",
			},
		},
	}
	result, err := readHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("read returned error")
	}
	if textOf(t, result) != "hello\nworld" {
		t.Errorf("read :2: got %q, want %q", textOf(t, result), "hello\nworld")
	}
}

func TestReadLineRangeBoth(t *testing.T) {
	setupTestRoot(t)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "read",
			Arguments: map[string]interface{}{
				"path":       "/file1.txt",
				"line_range": "2:2",
			},
		},
	}
	result, err := readHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("read returned error")
	}
	if textOf(t, result) != "world" {
		t.Errorf("read 2:2: got %q, want %q", textOf(t, result), "world")
	}
}

func TestReadLineRangeInvalid(t *testing.T) {
	setupTestRoot(t)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "read",
			Arguments: map[string]interface{}{
				"path":       "/file1.txt",
				"line_range": "abc:def",
			},
		},
	}
	result, err := readHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("read returned error")
	}
	if textOf(t, result) != "hello\nworld\nfoo" {
		t.Errorf("read invalid range: got %q, want full content", textOf(t, result))
	}
}

func TestReadLineRangeOverflow(t *testing.T) {
	setupTestRoot(t)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "read",
			Arguments: map[string]interface{}{
				"path":       "/file1.txt",
				"line_range": "1:999",
			},
		},
	}
	result, err := readHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("read returned error")
	}
	if textOf(t, result) != "hello\nworld\nfoo" {
		t.Errorf("read 1:999: got %q, want full content", textOf(t, result))
	}
}
