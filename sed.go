package main

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func sedHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pattern, err := req.RequireString("pattern")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	replacement, err := req.RequireString("replacement")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	pathspec, err := req.RequireString("pathspec")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	var files []string
	filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(rootDir, path)
		if err != nil {
			return nil
		}
		if globMatch(pathspec, rel) {
			files = append(files, path)
		}
		return nil
	})
	var changed []string
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		newData := re.ReplaceAllLiteral(data, []byte(replacement))
		if string(newData) != string(data) {
			if err := os.WriteFile(f, newData, 0644); err != nil {
				continue
			}
			rel, _ := filepath.Rel(rootDir, f)
			changed = append(changed, rel)
		}
	}
	if changed == nil {
		return mcp.NewToolResultText(""), nil
	}
	return mcp.NewToolResultText(strings.Join(changed, "\n")), nil
}
