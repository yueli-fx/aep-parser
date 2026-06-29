// internal/aep/mask_feather_falloff_test.go
//
// SetFeatherFalloff writes the mask feather-falloff curve at mkif @0x03
// (RE'd 2026-06-17: byte-diff of two AE-native masks, Smooth vs Linear, showed
// the only mkif delta beyond index/colour was @0x03 = 0/1). Go round-trip proves
// the byte survives parse; the AE gate is the coincidence-proof that @0x03 really
// is the falloff (AE reads it back as FFO_LINEAR), since a new RE field needs AE
// validation, not just a value round-trip (delivery-contract red line 7a).
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func buildMaskFalloffDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "FALLOFF", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	card, err := aep.NewShapeLayer(comp, "CARD")
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	r, err := card.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect: %v", err)
	}
	if err := r.SetSize([2]float64{600, 600}); err != nil {
		t.Fatalf("SetSize: %v", err)
	}
	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	l := rp.Compositions[0].LayerByName("CARD")
	if l == nil {
		t.Fatal("CARD missing after reopen")
	}
	rect := aep.BezierPath{
		Vertices: [][2]float64{{-300, -300}, {300, -300}, {300, 300}, {-300, 300}},
		Closed:   true,
	}
	m, err := aep.AddMask(l, "M", rect)
	if err != nil {
		t.Fatalf("AddMask: %v", err)
	}
	// Feather so the falloff curve is actually meaningful in AE.
	if err := m.SetFeather([2]float64{80, 80}); err != nil {
		t.Fatalf("SetFeather: %v", err)
	}
	if err := m.SetFeatherFalloff(aep.MaskFeatherFalloffLinear); err != nil {
		t.Fatalf("SetFeatherFalloff: %v", err)
	}
	return rp
}

func TestSetMaskFeatherFalloff_RoundTrip(t *testing.T) {
	rp := buildMaskFalloffDemo(t, aep.TargetAE2020)
	dir := t.TempDir()
	fpath := filepath.Join(dir, "falloff.aep")
	out, err := os.Create(fpath)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	re, err := aep.Open(fpath)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	l := re.Compositions[0].LayerByName("CARD")
	if l == nil || len(l.Masks) != 1 {
		t.Fatal("CARD mask missing after round-trip")
	}
	if got := l.Masks[0].FeatherFalloff; got != aep.MaskFeatherFalloffLinear {
		t.Errorf("FeatherFalloff = %d, want Linear (%d)", got, aep.MaskFeatherFalloffLinear)
	}
}

// TestSetMaskFeatherFalloff_DefaultSmooth confirms a fresh AddMask leaves falloff
// at the AE default Smooth (0) — i.e. @0x03 is not accidentally set.
func TestSetMaskFeatherFalloff_DefaultSmooth(t *testing.T) {
	p := aep.NewProject(aep.TargetAE2020)
	comp, _ := aep.NewComposition(p, "FALLOFF", 1920, 1080, 30, 5)
	card, _ := aep.NewShapeLayer(comp, "CARD")
	r, _ := card.RootGroup().AddRect()
	_ = r.SetSize([2]float64{600, 600})
	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	l := rp.Compositions[0].LayerByName("CARD")
	rect := aep.BezierPath{Vertices: [][2]float64{{-300, -300}, {300, -300}, {300, 300}, {-300, 300}}, Closed: true}
	m, err := aep.AddMask(l, "M", rect)
	if err != nil {
		t.Fatalf("AddMask: %v", err)
	}
	if m.FeatherFalloff != aep.MaskFeatherFalloffSmooth {
		t.Errorf("fresh mask FeatherFalloff = %d, want Smooth (0)", m.FeatherFalloff)
	}
}

func runMaskFalloffGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/mask_falloff_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_mask_feather_falloff.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	rp := buildMaskFalloffDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "falloff_in.aep")
	resavedAEP := filepath.Join(tempDir, "falloff_resaved.aep")
	doneFile := filepath.Join(tempDir, "falloff.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"comp":"FALLOFF","layer":"CARD"}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP))
	if err := writeGeneratedArgs(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 240)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	body := string(content)
	t.Logf("mask falloff %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mask falloff %s ship gate FAIL:\n%s", ver, body)
	}

	// Resave proof: AE's own re-encode preserves the Linear falloff (@0x03).
	reSaved, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	l := reSaved.Compositions[0].LayerByName("CARD")
	if l == nil || len(l.Masks) != 1 {
		t.Fatal("resaved: CARD mask missing")
	}
	if got := l.Masks[0].FeatherFalloff; got != aep.MaskFeatherFalloffLinear {
		t.Errorf("resaved FeatherFalloff = %d, want Linear (AE reverted the falloff)", got)
	}
}

func TestMaskFeatherFalloff_AEShipGate_AE2020(t *testing.T) {
	runMaskFalloffGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMaskFeatherFalloff_AEShipGate_AE2025(t *testing.T) {
	runMaskFalloffGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
