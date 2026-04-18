package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func grepHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pattern, err := req.RequireString("pattern")
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
		if isConfigPath(path) {
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
	var out []string
	for _, f := range files {
		rel, _ := filepath.Rel(rootDir, f)
		fh, err := os.Open(f)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(fh)
		ln := 0
		for scanner.Scan() {
			ln++
			if re.MatchString(scanner.Text()) {
				out = append(out, fmt.Sprintf("%s:%d:%s", rel, ln, scanner.Text()))
			}
		}
		fh.Close()
	}
	if out == nil {
		return mcp.NewToolResultText(""), nil
	}
	return mcp.NewToolResultText(strings.Join(out, "\n")), nil
}
