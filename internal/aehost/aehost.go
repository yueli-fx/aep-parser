package aehost

import (
	"context"
	"io"
)

type CapabilityStatus string

const (
	CapabilityAvailable   CapabilityStatus = "available"
	CapabilityUnavailable CapabilityStatus = "unavailable"
)

type Availability struct {
	Status  CapabilityStatus `json:"status"`
	Reason  string           `json:"reason,omitempty"`
	Version string           `json:"version,omitempty"`
	Path    string           `json:"path,omitempty"`
}

type ScriptRequest struct {
	AEPath     string
	JSXPath    string
	DonePath   string
	TimeoutSec int
	WorkDir    string
	Env        map[string]string
	Stdout     io.Writer
	Stderr     io.Writer
}

type ScriptResult struct {
	ExitCode int
	DonePath string
}

type Host interface {
	Available(ctx context.Context) Availability
	RunScript(ctx context.Context, req ScriptRequest) (ScriptResult, error)
}
