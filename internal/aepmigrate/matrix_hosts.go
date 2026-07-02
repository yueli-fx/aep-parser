package aepmigrate

import (
	"os"
	"path/filepath"
	"strings"
)

func BuildAEHostMap(installRoot string) map[string]string {
	hosts := map[string]string{}
	if strings.TrimSpace(installRoot) == "" {
		return hosts
	}
	for _, label := range SupportedVersionStrings() {
		year := strings.TrimPrefix(label, "AE")
		path := filepath.Join(installRoot, "Adobe After Effects "+year, "Support Files", "AfterFX.exe")
		if _, err := os.Stat(path); err == nil {
			hosts[label] = path
		}
	}
	return hosts
}
