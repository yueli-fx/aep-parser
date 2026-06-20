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
		{Kind: aep.PseudoSlider, Name: "Strength", Min: -100, Max: 100, Default: 50},
		{Kind: aep.PseudoColor, Name: "Tint", Color: []float64{1, 0, 0, 1}},
		{Kind: aep.PseudoCheckbox, Name: "Enabled", Checked: true},
		{Kind: aep.PseudoAngle, Name: "Rotation", Default: 45},
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

	// 6 = 5 controls + Compositing Options. checks assert AE honored the
	// synthesized pard defaults (1-based property index within the effect):
	// [1] slider value 50 / range -100..100, [3] checkbox checked=1, [4] angle 45°.
	const checks = `[` +
		`{"idx":1,"prop":"value","expect":50},` +
		`{"idx":1,"prop":"min","expect":-100},` +
		`{"idx":1,"prop":"max","expect":100},` +
		`{"idx":3,"prop":"value","expect":1},` +
		`{"idx":4,"prop":"value","expect":45}]`
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"matchName":%q,"minParams":6,"checks":%s}`,
		toFwd(inputAEP), toFwd(doneFile), matchName, checks)
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

// TestBuildPseudoEffectRich_AEShipGate_* verifies the from-scratch synthesizer
// for the structural control kinds — Dropdown (menu), Group (nesting) + a child,
// and Label — are accepted by AE 2020 + 2025 and read back live. The dropdown's
// selected index reads back, and the "Advanced" group nests its child angle.
func TestBuildPseudoEffectRich_AEShipGate_AE2020(t *testing.T) {
	runBuildPseudoEffectRichGate(t, ae2020(), aep.TargetAE2020)
}

func TestBuildPseudoEffectRich_AEShipGate_AE2025(t *testing.T) {
	runBuildPseudoEffectRichGate(t, ae2025(), aep.TargetAE2025)
}

func runBuildPseudoEffectRichGate(t *testing.T, aeExe string, target aep.AETarget) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/pseudo_effect_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_pseudo_effect.jsx`
	const matchName = "Pseudo/aepgo02/Rich"
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
	// slider (1), dropdown (2), a group wrapping one angle child, then a label.
	controls := []aep.PseudoControl{
		{Kind: aep.PseudoSlider, Name: "Strength", Min: -100, Max: 100, Default: 50},
		{Kind: aep.PseudoDropdown, Name: "Mode", Options: []string{"Add", "Screen", "Multiply"}, Default: 2},
		{Kind: aep.PseudoGroupStart, Name: "Advanced"},
		{Kind: aep.PseudoAngle, Name: "Spin", Default: 90},
		{Kind: aep.PseudoGroupEnd},
		{Kind: aep.PseudoLabel, Name: "footer"},
	}
	if _, err := aep.BuildPseudoEffect(l, "aepgo02", "Rich", "Rich Effect", controls); err != nil {
		t.Fatalf("BuildPseudoEffect: %v", err)
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "rich_in.aep")
	doneFile := filepath.Join(tempDir, "rich.done")
	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	// Checks robust to group indexing: slider value/range at top-level idx 1,
	// dropdown selected index (2) at idx 2. minParams modest — structural
	// acceptance (matchName live, not Missing, enabled) is the crux; the log
	// carries a recursive structure dump for the group/label layout.
	const checks = `[` +
		`{"idx":1,"prop":"value","expect":50},` +
		`{"idx":1,"prop":"min","expect":-100},` +
		`{"idx":1,"prop":"max","expect":100},` +
		`{"idx":2,"prop":"value","expect":2}]`
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"matchName":%q,"minParams":4,"checks":%s}`,
		toFwd(inputAEP), toFwd(doneFile), matchName, checks)
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
		t.Fatalf("AE rejected Go-synthesized rich pseudo effect:\n%s", body)
	}
}

// TestBuildPseudoEffectValueEntry_AEShipGate_* verifies the value-entry
// synthesizer: a Point / 3D-Point with a custom default position, and a Layer
// picker bound to a specific layer. AE 2020 + 2025 read back the exact pixel
// coordinates (the fraction-of-coord-space cdat decoded right) and resolve the
// picker to the bound layer's index.
func TestBuildPseudoEffectValueEntry_AEShipGate_AE2020(t *testing.T) {
	runBuildPseudoEffectValueEntryGate(t, ae2020(), aep.TargetAE2020)
}

func TestBuildPseudoEffectValueEntry_AEShipGate_AE2025(t *testing.T) {
	runBuildPseudoEffectValueEntryGate(t, ae2025(), aep.TargetAE2025)
}

func runBuildPseudoEffectValueEntryGate(t *testing.T, aeExe string, target aep.AETarget) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/pseudo_effect_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_pseudo_effect.jsx`
	const matchName = "Pseudo/aepgo03/P2"
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "Main", 400, 400, 30, 5)
	if err != nil {
		t.Fatal(err)
	}
	// Two layers: S hosts the effect, T is the layer-picker's bind target.
	if _, err := aep.NewShapeLayer(comp, "S"); err != nil {
		t.Fatal(err)
	}
	if _, err := aep.NewShapeLayer(comp, "T"); err != nil {
		t.Fatal(err)
	}
	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	var host, target2 *aep.Layer
	for _, c := range rp.Compositions {
		for _, cl := range c.Layers {
			switch cl.Name {
			case "S":
				host = cl
			case "T":
				target2 = cl
			}
		}
	}
	if host == nil || target2 == nil {
		t.Fatal("reopened project: layer S or T not found")
	}
	// Point 0.25/0.125 of a 400-comp → [100,50]; 3D adds z 0.0625 → 25. Picker
	// bound to T by its internal ID → AE's DOM exposes the resolved binding as
	// T's 1-based timeline index. Layer.Index is 0-based (the on-disk Layr order,
	// which matches AE top→bottom), so the AE-visible index is target2.Index+1.
	controls := []aep.PseudoControl{
		{Kind: aep.PseudoPoint, Name: "Center", PointDefault: []float64{0.25, 0.125}},
		{Kind: aep.PseudoPoint3D, Name: "Pos3D", PointDefault: []float64{0.25, 0.125, 0.0625}},
		{Kind: aep.PseudoLayer, Name: "Pick", LayerID: target2.ID},
	}
	if _, err := aep.BuildPseudoEffect(host, "aepgo03", "P2", "P2 Effect", controls); err != nil {
		t.Fatalf("BuildPseudoEffect: %v", err)
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "p2_in.aep")
	doneFile := filepath.Join(tempDir, "p2.done")
	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	checks := fmt.Sprintf(`[`+
		`{"idx":1,"prop":"value","expect":[100,50]},`+
		`{"idx":2,"prop":"value","expect":[100,50,25]},`+
		`{"idx":3,"prop":"value","expect":%d}]`, target2.Index+1)
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"matchName":%q,"minParams":3,"checks":%s}`,
		toFwd(inputAEP), toFwd(doneFile), matchName, checks)
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
		t.Fatalf("AE rejected Go-synthesized pseudo value entries:\n%s", body)
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
