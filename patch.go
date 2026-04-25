package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/mark3labs/mcp-go/mcp"
)

type hunk struct {
	start1, count1 int
	body           []string
}

func extractPatchPath(line, prefix string) string {
	p := strings.TrimPrefix(line, prefix)
	if idx := strings.IndexByte(p, '\t'); idx >= 0 {
		p = p[:idx]
	}
	for _, pre := range []string{"a/", "b/"} {
		if strings.HasPrefix(p, pre) {
			p = p[len(pre)-1:]
			break
		}
	}
	return p
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
	path1 := extractPatchPath(lines[0], "--- ")
	path2 := extractPatchPath(lines[1], "+++ ")
	_ = path2
	abs, err := resolve(path1)
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
	fileLines := splitLines(string(data))

	hunks := parseHunks(lines[2:])
	for _, h := range hunks {
		if h.start1 < 1 {
			return mcp.NewToolResultError(fmt.Sprintf(
				"hunk @@ -%d,%d @@ expected context at line %d but file has only %d lines -- file may have shifted after previous edits",
				h.start1, h.count1, h.start1, len(fileLines))), nil
		}
		if h.start1-1+h.count1 > len(fileLines) {
			return mcp.NewToolResultError(fmt.Sprintf(
				"hunk @@ -%d,%d @@ expected context at line %d but file has only %d lines -- file may have shifted after previous edits",
				h.start1, h.count1, h.start1+h.count1, len(fileLines))), nil
		}
		lineIdx := h.start1 - 1
		for _, l := range h.body {
			if len(l) == 0 {
				continue
			}
			switch l[0] {
			case ' ':
				if whitespaceNorm(fileLines[lineIdx]) != whitespaceNorm(l[1:]) {
					return mcp.NewToolResultError(fmt.Sprintf(
						"hunk @@ -%d,%d @@ expected context %q at line %d but found %q -- file may have shifted after previous edits",
						h.start1, h.count1, l[1:], lineIdx+1, fileLines[lineIdx])), nil
				}
				lineIdx++
			case '-':
				lineIdx++
			}
		}
	}
	for hi := len(hunks) - 1; hi >= 0; hi-- {
		h := hunks[hi]
		start := h.start1 - 1
		var newLines []string
		lineIdx := start
		for _, l := range h.body {
			if len(l) == 0 {
				continue
			}
			switch l[0] {
			case ' ':
				newLines = append(newLines, fileLines[lineIdx])
				lineIdx++
			case '-':
				lineIdx++
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
		return mcp.NewToolResultError(maskPath(err.Error())), nil
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

func whitespaceNorm(s string) string {
	var b strings.Builder
	prev := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if !prev {
				b.WriteByte(' ')
				prev = true
			}
			continue
		}
		b.WriteRune(r)
		prev = false
	}
	return strings.TrimSpace(b.String())
}
