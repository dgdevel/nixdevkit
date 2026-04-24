package main

import (
	"context"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func replaceRangeHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	p, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	lineRange, err := req.RequireString("line_range")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	content, err := req.RequireString("content")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	abs, err := resolve(p)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if isConfigPath(abs) || isIgnored(abs) {
		return mcp.NewToolResultError("access denied"), nil
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return mcp.NewToolResultError(maskPath(err.Error())), nil
	}
	lines := strings.Split(string(data), "\n")
	from, to := parseLineRange(lineRange, len(lines))
	if from > len(lines) {
		from = len(lines)
	}
	if to > len(lines) {
		to = len(lines)
	}
	var newLines []string
	if content != "" {
		newLines = strings.Split(content, "\n")
	}
	result := make([]string, 0, len(lines[:from])+len(newLines)+len(lines[to:]))
	result = append(result, lines[:from]...)
	result = append(result, newLines...)
	result = append(result, lines[to:]...)
	out := strings.Join(result, "\n")
	if err := os.WriteFile(abs, []byte(out), 0644); err != nil {
		return mcp.NewToolResultError(maskPath(err.Error())), nil
	}
	return mcp.NewToolResultText("ok"), nil
}
