package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// TestApplyPseudoEffect_AEShipGate_* verifies that AE accepts a pseudo effect
// (.ffx Animation Preset) spliced into a layer's Effect Parade by the pure-Go
// ApplyPseudoEffect — the offline equivalent of applyPreset — and reads it back
// LIVE (match-name correct, not a "Missing:" placeholder, all controls present,
// enabled) in a fresh AE process that never registered the preset. This is the
// self-contained-acceptance crux: the pseudo control definitions travel inside
// the .ffx bytes, so AE renders them from the saved .aep without registration.
//
// The fixture pseudo_scribe.ffx is the rendertom/PseudoEffect "Scribe" sample
// (RIFX FaFX form). ApplyPseudoEffect applies it at its pard defaults (authored
// .ffx values are not yet preserved — see ApplyPseudoEffect docs).
func TestApplyPseudoEffect_AEShipGate_AE2020(t *testing.T) {
	runApplyPseudoEffectGate(t, ae2020(), aep.TargetAE2020)
}

func TestApplyPseudoEffect_AEShipGate_AE2025(t *testing.T) {
	runApplyPseudoEffectGate(t, ae2025(), aep.TargetAE2025)
}

func runApplyPseudoEffectGate(t *testing.T, aeExe string, target aep.AETarget) {
	t.Helper()
	// displayName "" → applied with the .ffx's own name; nameCodes nil → no
	// custom-name char-code assertion.
	runApplyPseudoEffectGateImpl(t, aeExe, target, "", nil)
}

// TestApplyPseudoEffectNamed_CJK_AEShipGate_* verifies a CJK effect-instance
// display name ("伪效果", U+4F2A U+6548 U+679C) round-trips through Go → .aep →
// AE exactly. CJK names are where naïve byte-pokers break (the Utf8 sub-record's
// length is in bytes, not characters); we encode bytes, so AE reads "伪效果"
// back as 3 chars with the right code points on both AE versions.
func TestApplyPseudoEffectNamed_CJK_AEShipGate_AE2020(t *testing.T) {
	runApplyPseudoEffectGateImpl(t, ae2020(), aep.TargetAE2020, "伪效果", []rune{'伪', '效', '果'})
}

func TestApplyPseudoEffectNamed_CJK_AEShipGate_AE2025(t *testing.T) {
	runApplyPseudoEffectGateImpl(t, ae2025(), aep.TargetAE2025, "伪效果", []rune{'伪', '效', '果'})
}

// TestBuildPseudoEffect_AEShipGate_* verifies a pseudo effect synthesized
// entirely in Go (no .ffx, no AE, no cloned template bytes — every pard built
// field-by-field) is accepted by AE 2020 + 2025 and reads back live with all its
// controls (Slider/Color/Checkbox/Angle/Point + Compositing Options).
func TestBuildPseudoEffect_AEShipGate_AE2020(t *testing.T) {
	runBuildPseudoEffectGate(t, ae2020(), aep.TargetAE2020)
}

func TestBuildPseudoEffect_AEShipGate_AE2025(t *testing.T) {
	runBuildPseudoEffectGate(t, ae2025(), aep.TargetAE2025)
}

func runBuildPseudoEffectGate(t *testing.T, aeExe string, target aep.AETarget) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/pseudo_effect_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_pseudo_effect.jsx`
	const matchName = "Pseudo/aepgo01/Demo"
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "Main", 400, 400, 30, 5)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := aep.NewShapeLayer(comp, "S"); err != nil {
		t.Fatal(err)
	}
	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	var l *aep.Layer
	for _, c := range rp.Compositions {
		for _, cl := range c.Layers {
			if cl.Name == "S" {
				l = cl
			}
		}
	}
	controls := []aep.PseudoControl{
		{Kind: aep.PseudoSlider, Name: "Strength"},
		{Kind: aep.PseudoColor, Name: "Tint"},
		{Kind: aep.PseudoCheckbox, Name: "Enabled"},
		{Kind: aep.PseudoAngle, Name: "Rotation"},
		{Kind: aep.PseudoPoint, Name: "Center"},
	}
	if _, err := aep.BuildPseudoEffect(l, "aepgo01", "Demo", "Demo Effect", controls); err != nil {
		t.Fatalf("BuildPseudoEffect: %v", err)
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "build_in.aep")
	doneFile := filepath.Join(tempDir, "build.done")
	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	// 6 = 5 controls + Compositing Options.
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"matchName":%q,"minParams":6}`,
		toFwd(inputAEP), toFwd(doneFile), matchName)
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	if body := string(content); !strings.HasPrefix(body, "PASS") {
		t.Fatalf("AE rejected Go-synthesized pseudo effect:\n%s", body)
	}
}

func runApplyPseudoEffectGateImpl(t *testing.T, aeExe string, target aep.AETarget, displayName string, nameCodes []rune) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	const ffxPath = `e:/projects/tools/aep-parser/test_data/pseudo_scribe.ffx`
	const argsPath = `e:/projects/tools/aep-parser/test_data/pseudo_effect_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_pseudo_effect.jsx`
	const matchName = "Pseudo/9db0uID/Scribe"
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	ffx, err := os.ReadFile(ffxPath)
	if err != nil {
		t.Fatalf("read fixture .ffx: %v", err)
	}

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "Main", 400, 400, 30, 5)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := aep.NewShapeLayer(comp, "S"); err != nil {
		t.Fatal(err)
	}
	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	var l *aep.Layer
	for _, c := range rp.Compositions {
		for _, cl := range c.Layers {
			if cl.Name == "S" {
				l = cl
			}
		}
	}
	if l == nil {
		t.Fatal("reopened project: layer S not found")
	}
	if _, err := aep.ApplyPseudoEffectNamed(l, ffx, displayName); err != nil {
		t.Fatalf("ApplyPseudoEffectNamed: %v", err)
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "pseudo_in.aep")
	doneFile := filepath.Join(tempDir, "pseudo.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	nameCodesJSON := ""
	if len(nameCodes) > 0 {
		parts := make([]string, len(nameCodes))
		for i, r := range nameCodes {
			parts[i] = fmt.Sprintf("%d", r)
		}
		nameCodesJSON = fmt.Sprintf(`,"nameCodes":[%s]`, strings.Join(parts, ","))
	}
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"matchName":%q,"minParams":8%s}`,
		toFwd(inputAEP), toFwd(doneFile), matchName, nameCodesJSON)
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	if body := string(content); !strings.HasPrefix(body, "PASS") {
		t.Fatalf("AE rejected Go-spliced pseudo effect:\n%s", body)
	}
}
