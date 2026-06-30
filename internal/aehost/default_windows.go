//go:build windows

package aehost

import (
	"path/filepath"

	"github.com/yueli-fx/aep-parser/internal/host"
)

func DefaultHost() Host {
	return NewPowerShellHost(host.ExecRunner{}, filepath.Join("scripts", "ae-worker", "ae_run.ps1"))
}
