package util

import (
	"os"
	"path/filepath"
)

func EnsureDir(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return nil
}
