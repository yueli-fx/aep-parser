package aehost

import (
	"context"
	"fmt"
)

const DefaultUnavailableReason = "after effects automation is not configured on this platform"

type UnavailableHost struct {
	Reason string
}

func NewUnavailableHost(reason string) UnavailableHost {
	if reason == "" {
		reason = DefaultUnavailableReason
	}
	return UnavailableHost{Reason: reason}
}

func (h UnavailableHost) Available(context.Context) Availability {
	return Availability{
		Status: CapabilityUnavailable,
		Reason: h.Reason,
	}
}

func (h UnavailableHost) RunScript(_ context.Context, req ScriptRequest) (ScriptResult, error) {
	reason := h.Reason
	if reason == "" {
		reason = DefaultUnavailableReason
	}
	return ScriptResult{ExitCode: 1, DonePath: req.DonePath}, fmt.Errorf("%s", reason)
}
