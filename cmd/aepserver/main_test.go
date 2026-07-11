package main

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestRunHelpPrintsUsage(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := runWithIO([]string{"-h"}, &stdout, &stderr, nil)

	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "usage: aepserver") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunRejectsInvalidAddr(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := runWithIO([]string{"-addr", ""}, &stdout, &stderr, nil)

	if code != 2 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "-addr is required") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRunInvokesServe(t *testing.T) {
	var stdout, stderr bytes.Buffer
	called := false

	code := runWithIO([]string{"-addr", "127.0.0.1:0", "-max-body-mb", "32", "-allow-path-input", "-allowed-path-roots", "."}, &stdout, &stderr, func(opts serveOptions) error {
		called = true
		if opts.Addr != "127.0.0.1:0" {
			t.Fatalf("Addr = %q", opts.Addr)
		}
		if opts.MaxBodyBytes != 32<<20 {
			t.Fatalf("MaxBodyBytes = %d", opts.MaxBodyBytes)
		}
		if !opts.AllowPathInput {
			t.Fatal("AllowPathInput = false")
		}
		if opts.MaxConcurrent != 32 || len(opts.AllowedPathRoots) != 1 {
			t.Fatalf("safety options = %+v", opts)
		}
		if opts.ReadHeaderTimeout != 5*time.Second || opts.ReadTimeout != 30*time.Second || opts.WriteTimeout != 60*time.Second || opts.IdleTimeout != 120*time.Second {
			t.Fatalf("timeouts = %+v", opts)
		}
		return nil
	})

	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr.String())
	}
	if !called {
		t.Fatal("serve was not called")
	}
}

func TestRunRequiresAllowedRootsForPathInput(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runWithIO([]string{"-allow-path-input"}, &stdout, &stderr, nil)
	if code != 2 || !strings.Contains(stderr.String(), "requires -allowed-path-roots") {
		t.Fatalf("exit = %d, stderr = %s", code, stderr.String())
	}
}
