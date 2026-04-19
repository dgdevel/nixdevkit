package main

import (
	"path/filepath"
	"strings"

	"nixdevkit/internal/cfg"
)

func isConfigPath(abs string) bool {
	dir := cfg.DirPath(rootDir)
	return abs == dir || strings.HasPrefix(abs, dir+string(filepath.Separator))
}

func isReadonly() bool {
	return cfg.IsReadonly(rootDir)
}
