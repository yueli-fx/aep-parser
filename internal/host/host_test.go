package host

import (
	"context"
	"os"
	"testing"
)

func TestDefaultPlatformCanInspectCurrentProcess(t *testing.T) {
	platform := DefaultPlatform()
	if !platform.ProcessInspector.IsRunning(os.Getpid()) {
		t.Fatalf("current process pid %d should be running", os.Getpid())
	}
}

func TestRunnerReportsCommandExitCode(t *testing.T) {
	platform := DefaultPlatform()
	result := platform.Runner.Run(context.Background(), Command{
		Name:   os.Args[0],
		Args:   []string{"-test.run=TestHelperProcessExit3", "--"},
		Env:    append(os.Environ(), "GO_WANT_HELPER_PROCESS=1"),
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	if result.ExitCode != 3 {
		t.Fatalf("exit code = %d, want 3", result.ExitCode)
	}
}

func TestHelperProcessExit3(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	os.Exit(3)
}
