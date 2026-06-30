package aehost

import (
	"context"
	"os"
	"strconv"

	"github.com/yueli-fx/aep-parser/internal/host"
)

type PowerShellHost struct {
	Runner     host.Runner
	ScriptPath string
}

func NewPowerShellHost(runner host.Runner, scriptPath string) PowerShellHost {
	if runner == nil {
		runner = host.ExecRunner{}
	}
	return PowerShellHost{
		Runner:     runner,
		ScriptPath: scriptPath,
	}
}

func (h PowerShellHost) Available(context.Context) Availability {
	return Availability{
		Status: CapabilityAvailable,
		Path:   h.ScriptPath,
	}
}

func (h PowerShellHost) RunScript(ctx context.Context, req ScriptRequest) (ScriptResult, error) {
	runner := h.Runner
	if runner == nil {
		runner = host.ExecRunner{}
	}
	env := os.Environ()
	for key, value := range req.Env {
		env = append(env, key+"="+value)
	}
	result := runner.Run(ctx, host.Command{
		Name: "pwsh",
		Args: []string{
			"-NoProfile",
			"-File", h.ScriptPath,
			"-AeExe", req.AEPath,
			"-Jsx", req.JSXPath,
			"-Done", req.DonePath,
			"-TimeoutSec", strconv.Itoa(req.TimeoutSec),
		},
		Dir:    req.WorkDir,
		Env:    env,
		Stdout: req.Stdout,
		Stderr: req.Stderr,
	})
	return ScriptResult{ExitCode: result.ExitCode, DonePath: req.DonePath}, nil
}
