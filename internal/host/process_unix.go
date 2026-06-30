//go:build !windows

package host

import (
	"os"
	"syscall"
)

type processInspector struct{}

func NewProcessInspector() ProcessInspector {
	return processInspector{}
}

func (processInspector) IsRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}
