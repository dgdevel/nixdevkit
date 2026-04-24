package main

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestDiffStringsIdentical(t *testing.T) {
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "diff_strings",
			Arguments: map[string]interface{}{
				"string1": "hello\nworld",
				"string2": "hello\nworld",
			},
		},
	}
	result, err := diffStringsHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("diff_strings returned error")
	}
	if textOf(t, result) != "" {
		t.Errorf("expected empty, got %q", textOf(t, result))
	}
}

func TestDiffStringsChanged(t *testing.T) {
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "diff_strings",
			Arguments: map[string]interface{}{
				"string1": "hello\nworld\nfoo",
				"string2": "hello\nearth\nfoo",
			},
		},
	}
	result, err := diffStringsHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("diff_strings returned error")
	}
	text := textOf(t, result)
	if !containsLine(text, "-world") || !containsLine(text, "+earth") {
		t.Errorf("missing expected changes in %q", text)
	}
	if !containsLine(text, " hello") {
		t.Errorf("missing context in %q", text)
	}
}

func TestDiffStringsAdded(t *testing.T) {
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "diff_strings",
			Arguments: map[string]interface{}{
				"string1": "hello\nworld\nfoo",
				"string2": "hello\nworld\nfoo\nbar",
			},
		},
	}
	result, err := diffStringsHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	text := textOf(t, result)
	if !containsLine(text, "+bar") {
		t.Errorf("missing +bar in %q", text)
	}
}

func TestDiffStringsRemoved(t *testing.T) {
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "diff_strings",
			Arguments: map[string]interface{}{
				"string1": "hello\nworld\nfoo",
				"string2": "hello\nfoo",
			},
		},
	}
	result, err := diffStringsHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	text := textOf(t, result)
	if !containsLine(text, "-world") {
		t.Errorf("missing -world in %q", text)
	}
}

func TestDiffStringsEmpty(t *testing.T) {
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "diff_strings",
			Arguments: map[string]interface{}{
				"string1": "",
				"string2": "hello\nworld",
			},
		},
	}
	result, err := diffStringsHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	text := textOf(t, result)
	if !containsLine(text, "+hello") || !containsLine(text, "+world") {
		t.Errorf("missing additions in %q", text)
	}
}
