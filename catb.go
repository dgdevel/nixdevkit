package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func catbHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	if isConfigPath(abs) || isIgnored(abs) {
		return mcp.NewToolResultError("access denied"), nil
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return mcp.NewToolResultError(maskPath(err.Error())), nil
	}
	lines := strings.Split(string(data), "\n")
	from, to := parseLineRange(lineRange, len(lines))
	if from >= len(lines) {
		return mcp.NewToolResultText(""), nil
	}
	if to > len(lines) {
		to = len(lines)
	}
	var b strings.Builder
	requested := to - from
	if requested > 200 {
		fmt.Fprintf(&b, "Showing lines %d-%d of %d. Use line_range to read more.\n", from+1, from+200, len(lines))
		to = from + 200
	}
	for i, line := range lines[from:to] {
		if line == "" {
			b.WriteString("\n")
		} else {
			visible := strings.ReplaceAll(line, "\t", "→")
			visible = strings.TrimRight(visible, " ")
			trailing := len(line) - len(strings.TrimRight(line, " "))
			if trailing > 0 {
				visible += strings.Repeat("·", trailing)
			}
			fmt.Fprintf(&b, "%6d\t%s\n", from+i+1, visible)
		}
	}
	return mcp.NewToolResultText(b.String()), nil
}
