package host

import (
	"context"
	"fmt"
	"io"
)

type Command struct {
	Name   string
	Args   []string
	Dir    string
	Env    []string
	Stdout io.Writer
	Stderr io.Writer
}

type Result struct {
	ExitCode int
}

type Runner interface {
	Run(ctx context.Context, cmd Command) Result
	Start(ctx context.Context, cmd Command) (Process, error)
}

type Process interface {
	PID() int
	Release() error
}

type ProcessInspector interface {
	IsRunning(pid int) bool
}

type BrowserOpener interface {
	Open(ctx context.Context, path string) error
}

type Platform struct {
	Runner           Runner
	ProcessInspector ProcessInspector
	BrowserOpener    BrowserOpener
}

func DefaultPlatform() Platform {
	return Platform{
		Runner:           ExecRunner{},
		ProcessInspector: NewProcessInspector(),
		BrowserOpener:    NoopBrowserOpener{},
	}
}

type NoopBrowserOpener struct{}

func (NoopBrowserOpener) Open(_ context.Context, path string) error {
	return fmt.Errorf("browser opener is not configured: %s", path)
}
