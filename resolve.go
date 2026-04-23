package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var rootDir string
var ignoreRe *regexp.Regexp

func isIgnored(abs string) bool {
	if ignoreRe == nil {
		return false
	}
	rel, err := filepath.Rel(rootDir, abs)
	if err != nil {
		return false
	}
	return ignoreRe.MatchString(rel)
}

func maskPath(s string) string {
	if rootDir == "" {
		return s
	}
	s = strings.ReplaceAll(s, rootDir+string(os.PathSeparator), string(os.PathSeparator))
	s = strings.ReplaceAll(s, rootDir, string(os.PathSeparator))
	return s
}

func resolve(p string) (string, error) {
	p = filepath.Clean(p)
	if p == "." {
		p = ""
	}
	abs, err := filepath.Abs(filepath.Join(rootDir, p))
	if err != nil {
		return "", err
	}
	r, err := filepath.Abs(rootDir)
	if err != nil {
		return "", err
	}
	if abs != r && !strings.HasPrefix(abs, r+string(os.PathSeparator)) {
		return "", fmt.Errorf("path escapes root")
	}
	return abs, nil
}

func globMatch(pattern, name string) bool {
	pp := strings.Split(pattern, "/")
	np := strings.Split(name, "/")
	var match func(int, int) bool
	match = func(pi, ni int) bool {
		for pi < len(pp) && ni < len(np) {
			if pp[pi] == "**" {
				pi++
				for i := ni; i <= len(np); i++ {
					if match(pi, i) {
						return true
					}
				}
				return false
			}
			ok, _ := filepath.Match(pp[pi], np[ni])
			if !ok {
				return false
			}
			pi++
			ni++
		}
		for pi < len(pp) && pp[pi] == "**" {
			pi++
		}
		return pi == len(pp) && ni == len(np)
	}
	return match(0, 0)
}

func parseLineRange(s string, total int) (int, int) {
	parts := strings.SplitN(s, ":", 2)
	from := 0
	to := total
	if parts[0] != "" {
		if v, err := strconv.Atoi(parts[0]); err == nil && v >= 0 {
			if v > 0 {
				from = v - 1
			}
		}
	}
	if len(parts) > 1 && parts[1] != "" {
		if v, err := strconv.Atoi(parts[1]); err == nil && v >= 0 {
			to = v
		}
	}
	return from, to
}
