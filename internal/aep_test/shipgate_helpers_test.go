package aep_test

import (
	"os"
	"path/filepath"
)

func writeGeneratedArgs(path string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, perm)
}
