package aepmigrate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func BuildAEHostMap(installRoot string) map[string]string {
	hosts := map[string]string{}
	if strings.TrimSpace(installRoot) == "" {
		return hosts
	}
	for year := 2020; year <= 2025; year++ {
		label := fmt.Sprintf("AE%d", year)
		path := filepath.Join(installRoot, fmt.Sprintf("Adobe After Effects %d", year), "Support Files", "AfterFX.exe")
		if _, err := os.Stat(path); err == nil {
			hosts[label] = path
		}
	}
	return hosts
}
