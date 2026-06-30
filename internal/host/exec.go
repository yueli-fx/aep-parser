package host

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
)

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, req Command) Result {
	cmd := newExecCommand(ctx, req)
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return Result{ExitCode: exitErr.ExitCode()}
		}
		if req.Stderr != nil {
			fmt.Fprintln(req.Stderr, err)
		}
		return Result{ExitCode: 1}
	}
	return Result{ExitCode: 0}
}

func (ExecRunner) Start(ctx context.Context, req Command) (Process, error) {
	cmd := newExecCommand(ctx, req)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return execProcess{cmd: cmd}, nil
}

func newExecCommand(ctx context.Context, req Command) *exec.Cmd {
	cmd := exec.CommandContext(ctx, req.Name, req.Args...)
	cmd.Dir = req.Dir
	if req.Env != nil {
		cmd.Env = req.Env
	}
	cmd.Stdout = req.Stdout
	cmd.Stderr = req.Stderr
	return cmd
}

type execProcess struct {
	cmd *exec.Cmd
}

func (p execProcess) PID() int {
	if p.cmd == nil || p.cmd.Process == nil {
		return 0
	}
	return p.cmd.Process.Pid
}

func (p execProcess) Release() error {
	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}
	return p.cmd.Process.Release()
}
