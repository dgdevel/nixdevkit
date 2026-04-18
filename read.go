package main

import (
	"context"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func readHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	p, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	lineRange, err := req.RequireString("line_range")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	abs, err := resolve(p)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if isConfigPath(abs) {
		return mcp.NewToolResultError("access denied"), nil
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	lines := strings.Split(string(data), "\n")
	from, to := parseLineRange(lineRange, len(lines))
	if from >= len(lines) {
		return mcp.NewToolResultText(""), nil
	}
	if to > len(lines) {
		to = len(lines)
	}
	return mcp.NewToolResultText(strings.Join(lines[from:to], "\n")), nil
}
