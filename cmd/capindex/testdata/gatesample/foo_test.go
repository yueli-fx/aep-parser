package gatesample

import (
	"os"
	"testing"
)

// normal gate — env-conditional skip must NOT count as disabled.
func TestRealGate(t *testing.T) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE")
	}
	_ = 1
}

// unconditionally disabled — top-level skip counts as disabled.
func TestDisabledGate(t *testing.T) {
	t.Skip("TODO: not implemented")
}

// plain test, no skip.
func TestPlain(t *testing.T) {
	_ = 2
}

// not a test function (has args other than *testing.T pattern is irrelevant here;
// helper) — should be ignored by the Test* prefix filter.
func helper() {}
