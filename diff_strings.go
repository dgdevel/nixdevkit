package main

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

func diffStringsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s1, err := req.RequireString("string1")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	s2, err := req.RequireString("string2")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if s1 == s2 {
		return mcp.NewToolResultText(""), nil
	}
	a := splitLines(s1)
	b := splitLines(s2)
	return mcp.NewToolResultText(unifiedDiff("string1", "string2", a, b)), nil
}
