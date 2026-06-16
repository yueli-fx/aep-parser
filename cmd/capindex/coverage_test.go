package main

import (
	"path/filepath"
	"testing"
)

// TestLayerCreateFullyTagged enforces P1 scope: every layer-create New* facade
// symbol carries a valid layer-create aep:cap tag. (Full-facade tag coverage is
// enforced in P2.)
func TestLayerCreateFullyTagged(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	entries, err := extractEntries(filepath.Join(root, "internal", "aep"))
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]Entry{}
	for _, e := range entries {
		byName[e.Symbol] = e
	}
	want := []string{
		"NewShapeLayer", "NewSolidLayer", "NewNullLayer", "NewAdjustmentLayer",
		"NewCameraLayer", "NewLightLayer", "NewTextLayer", "NewPrecompLayer",
	}
	for _, n := range want {
		e, ok := byName[n]
		if !ok {
			t.Errorf("%s not extracted from facade", n)
			continue
		}
		if !e.HasCap || e.Cap.Domain != "layer-create" {
			t.Errorf("%s: expected layer-create tag, got hasCap=%v domain=%q", n, e.HasCap, e.Cap.Domain)
		}
	}
}
