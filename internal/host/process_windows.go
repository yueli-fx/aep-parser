//go:build windows

package host

import (
	"os/exec"
	"strconv"
	"strings"
)

type processInspector struct{}

func NewProcessInspector() ProcessInspector {
	return processInspector{}
}

func (processInspector) IsRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	out, err := exec.Command("tasklist", "/FI", "PID eq "+strconv.Itoa(pid), "/FO", "CSV", "/NH").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), strconv.Itoa(pid))
}
