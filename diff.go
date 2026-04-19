package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

type editOp struct {
	kind    byte
	line    string
	lineNo1 int
	lineNo2 int
}

func splitLines(data string) []string {
	lines := strings.Split(data, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func computeOps(a, b []string) []editOp {
	m, n := len(a), len(b)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}
	var ops []editOp
	i, j := m, n
	for i > 0 || j > 0 {
		if i > 0 && j > 0 && a[i-1] == b[j-1] {
			ops = append([]editOp{{' ', a[i-1], i - 1, j - 1}}, ops...)
			i--
			j--
		} else if j > 0 && (i == 0 || dp[i][j-1] >= dp[i-1][j]) {
			ops = append([]editOp{{'+', b[j-1], -1, j - 1}}, ops...)
			j--
		} else {
			ops = append([]editOp{{'-', a[i-1], i - 1, -1}}, ops...)
			i--
		}
	}
	return ops
}

func unifiedDiff(path1, path2 string, a, b []string) string {
	ops := computeOps(a, b)

	type changeRange struct{ start, end int }
	var changes []changeRange
	inChange := false
	for i, op := range ops {
		if op.kind != ' ' {
			if !inChange {
				changes = append(changes, changeRange{start: i})
				inChange = true
			}
			changes[len(changes)-1].end = i + 1
		} else {
			inChange = false
		}
	}
	if len(changes) == 0 {
		return ""
	}

	ctx := 3
	type hunkRange struct{ startOp, endOp int }
	var hunks []hunkRange
	for _, ch := range changes {
		hs := ch.start - ctx
		if hs < 0 {
			hs = 0
		}
		he := ch.end + ctx
		if he > len(ops) {
			he = len(ops)
		}
		if len(hunks) > 0 && hs <= hunks[len(hunks)-1].endOp {
			hunks[len(hunks)-1].endOp = he
		} else {
			hunks = append(hunks, hunkRange{hs, he})
		}
	}

	var buf strings.Builder
	buf.WriteString("--- " + path1 + "\n")
	buf.WriteString("+++ " + path2 + "\n")
	for _, h := range hunks {
		count1, count2 := 0, 0
		start1, start2 := 0, 0
		for i := h.startOp; i < h.endOp; i++ {
			if ops[i].lineNo1 >= 0 {
				if count1 == 0 {
					start1 = ops[i].lineNo1
				}
				count1++
			}
			if ops[i].lineNo2 >= 0 {
				if count2 == 0 {
					start2 = ops[i].lineNo2
				}
				count2++
			}
		}
		buf.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n", start1+1, count1, start2+1, count2))
		for i := h.startOp; i < h.endOp; i++ {
			buf.WriteByte(ops[i].kind)
			buf.WriteString(ops[i].line + "\n")
		}
	}
	return buf.String()
}

func diffHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	p1, err := req.RequireString("path1")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	p2, err := req.RequireString("path2")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	abs1, err := resolve(p1)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if isConfigPath(abs1) || isIgnored(abs1) {
		return mcp.NewToolResultError("access denied"), nil
	}
	abs2, err := resolve(p2)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if isConfigPath(abs2) || isIgnored(abs2) {
		return mcp.NewToolResultError("access denied"), nil
	}
	data1, err := os.ReadFile(abs1)
	if err != nil {
		return mcp.NewToolResultError(maskPath(err.Error())), nil
	}
	data2, err := os.ReadFile(abs2)
	if err != nil {
		return mcp.NewToolResultError(maskPath(err.Error())), nil
	}
	if string(data1) == string(data2) {
		return mcp.NewToolResultText(""), nil
	}
	a := splitLines(string(data1))
	b := splitLines(string(data2))
	return mcp.NewToolResultText(unifiedDiff(p1, p2, a, b)), nil
}
