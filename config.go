package main

import (
	"nixdevkit/internal/cfg"
)

func isConfigPath(abs string) bool {
	return abs == cfg.FilePath(rootDir)
}

func isReadonly() bool {
	return cfg.IsReadonly(rootDir)
}
