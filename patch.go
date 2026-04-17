package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

type hunk struct {
	start1, count1 int
	body           []string
}

func patchHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	patch, err := req.RequireString("patch")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	lines := strings.Split(patch, "\n")
	if len(lines) < 2 || !strings.HasPrefix(lines[0], "--- ") || !strings.HasPrefix(lines[1], "+++ ") {
		return mcp.NewToolResultError("invalid patch format"), nil
	}
	path1 := strings.TrimPrefix(lines[0], "--- ")
	abs, err := resolve(path1)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	fileLines := splitLines(string(data))

	hunks := parseHunks(lines[2:])
	for hi := len(hunks) - 1; hi >= 0; hi-- {
		h := hunks[hi]
		start := h.start1 - 1
		var newLines []string
		for _, l := range h.body {
			if len(l) == 0 {
				continue
			}
			switch l[0] {
			case ' ':
				newLines = append(newLines, l[1:])
			case '+':
				newLines = append(newLines, l[1:])
			}
		}
		end := start + h.count1
		result := make([]string, 0, len(fileLines)-h.count1+len(newLines))
		result = append(result, fileLines[:start]...)
		result = append(result, newLines...)
		result = append(result, fileLines[end:]...)
		fileLines = result
	}

	out := strings.Join(fileLines, "\n") + "\n"
	if err := os.WriteFile(abs, []byte(out), 0644); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText("ok"), nil
}

func parseHunks(lines []string) []hunk {
	var hunks []hunk
	i := 0
	for i < len(lines) {
		if !strings.HasPrefix(lines[i], "@@ ") {
			i++
			continue
		}
		var s1, c1 int
		fmt.Sscanf(lines[i], "@@ -%d,%d", &s1, &c1)
		i++
		var body []string
		f1 := 0
		for i < len(lines) && f1 < c1 {
			if strings.HasPrefix(lines[i], "@@ ") || len(lines[i]) == 0 {
				break
			}
			body = append(body, lines[i])
			if lines[i][0] == ' ' || lines[i][0] == '-' {
				f1++
			}
			i++
		}
		for i < len(lines) && len(lines[i]) > 0 && lines[i][0] == '+' {
			body = append(body, lines[i])
			i++
		}
		hunks = append(hunks, hunk{s1, c1, body})
	}
	return hunks
}
