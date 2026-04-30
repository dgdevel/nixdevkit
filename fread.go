package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"nixdevkit/internal/cfg"

	"github.com/mark3labs/mcp-go/mcp"
)

func freadHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	blockSize := 100
	config := cfg.MergedRead(rootDir)
	if core, ok := config["core"]; ok {
		blockSize = cfg.Atoi(core["fread_block_size"], 100)
	}
	if blockSize < 1 {
		blockSize = 1
	}

	lines := strings.Split(string(data), "\n")
	from, to := parseLineRange(lineRange, len(lines))
	if from >= len(lines) {
		return mcp.NewToolResultText(""), nil
	}
	if to > len(lines) {
		to = len(lines)
	}
	if from >= to {
		return mcp.NewToolResultText(""), nil
	}

	selected := lines[from:to]
	var b strings.Builder
	for i := 0; i < len(selected); i += blockSize {
		end := i + blockSize
		if end > len(selected) {
			end = len(selected)
		}
		fmt.Fprintf(&b, "----- %s - lines from %d to %d -----\n", p, from+i+1, from+end)
		for _, line := range selected[i:end] {
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}
	if to < len(lines) {
		fmt.Fprintf(&b, "----- %s - remaining lines from %d to %d -----\n", p, to+1, len(lines))
	} else {
		fmt.Fprintf(&b, "----- %s - EOF -----\n", p)
	}
	return mcp.NewToolResultText(b.String()), nil
}
