package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func findHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pattern, err := req.RequireString("pattern")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	var matches []string
	err = filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if isConfigPath(path) {
			return nil
		}
		rel, err := filepath.Rel(rootDir, path)
		if err != nil {
			return nil
		}
		if rel == "." {
			return nil
		}
		if !globMatch(pattern, rel) {
			return nil
		}
		name := rel
		if d.IsDir() {
			name += "/"
		}
		matches = append(matches, name)
		return nil
	})
	if err != nil {
		return mcp.NewToolResultError(maskPath(err.Error())), nil
	}
	if matches == nil {
		return mcp.NewToolResultText(""), nil
	}
	return mcp.NewToolResultText(strings.Join(matches, "\n")), nil
}
