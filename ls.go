package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func lsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	p, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	abs, err := resolve(p)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	rel, _ := filepath.Rel(rootDir, abs)
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if isConfigPath(filepath.Join(abs, e.Name())) {
			continue
		}
		n := filepath.Join(rel, e.Name())
		if e.IsDir() {
			n += "/"
		}
		names = append(names, n)
	}
	return mcp.NewToolResultText(strings.Join(names, "\n")), nil
}
