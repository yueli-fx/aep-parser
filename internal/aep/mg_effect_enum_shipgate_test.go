// internal/aep/mg_effect_enum_shipgate_test.go
//
// AE ship gate for CROSS-EFFECT enum parameter materialization. SetEffectParam's
// generic per-control-type path materializes a default-elided enum param from
// the Gaussian-Blur-extracted enum template (GB Blur Dimensions -0002), patched
// with the host effect's own pard metadata. That round-trips for GB itself
// (set_effect_param gate), but the cross-effect enum case — materializing an
// enum on a DIFFERENT effect — was never separately render-gated (the known
// caveat in incidents/effect-param-elision-synthesis-lite.md).
//
// This closes it at the render surface (red line 4): two from-scratch Go shape
// cards, each a solid (0.9,0.1,0.1) fill + Invert effect with a non-default
// Channel enum:
//
//	card  Channel        invert       rendered fill   R-channel
//	RED   2 (Red)   → (0.1,0.1,0.1)   dark gray        LOW
//	GRN   3 (Green) → (0.9,0.9,0.1)   yellow           HIGH
//
// Both values are non-default (default Channel = RGB = 1), so both exercise the
// generic-enum splice on Invert (not Gaussian Blur). The R channel swings the
// full 0.1↔0.9 between the two cards, so one rendered frame proves AE honored
// each distinct enum value — gamma-robust around the 128 midpoint. The resave
// reopen proves the materialized enum survives AE's own re-encode.
//
// Gated by AE_SHIP_GATE. Uses test_data/generators/verify_effect_enum.jsx.
package aep_test

import (
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

type enumCell struct {
	name    string
	cx, cy  int
	channel float64 // ADBE Invert-0001 enum value
	wantRHi bool    // expected: R channel high (untouched) vs low (inverted)
}

var effectEnumCells = []enumCell{
	{"RED", 560, 540, 2, false}, // invert Red → R low
	{"GRN", 1360, 540, 3, true}, // invert Green → R untouched → high
}

func buildEffectEnumDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "EFFECTENUM", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	for _, c := range effectEnumCells {
		card, err := aep.NewShapeLayer(comp, c.name)
		if err != nil {
			t.Fatalf("NewShapeLayer %s: %v", c.name, err)
		}
		r, err := card.RootGroup().AddRect()
		if err != nil {
			t.Fatalf("%s AddRect: %v", c.name, err)
		}
		if err := r.SetSize([2]float64{600, 600}); err != nil {
			t.Fatalf("%s SetSize: %v", c.name, err)
		}
		fill, err := card.RootGroup().AddFill()
		if err != nil {
			t.Fatalf("%s AddFill: %v", c.name, err)
		}
		if err := fill.SetColor([4]float64{0.9, 0.1, 0.1, 1}); err != nil {
			t.Fatalf("%s SetColor: %v", c.name, err)
		}
		if err := card.Position().SetStaticValue([2]float64{float64(c.cx), float64(c.cy)}); err != nil {
			t.Fatalf("%s Position: %v", c.name, err)
		}
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	for _, c := range effectEnumCells {
		l := rp.Compositions[0].LayerByName(c.name)
		if l == nil {
			t.Fatalf("reopened: layer %s missing", c.name)
		}
		fx, err := aep.AddEffect(l, aep.EffectInvert)
		if err != nil {
			t.Fatalf("%s AddEffect(Invert): %v", c.name, err)
		}
		if _, err := aep.SetEffectParam(l, fx, "ADBE Invert-0001", c.channel); err != nil {
			t.Fatalf("%s SetEffectParam(Channel=%v): %v", c.name, c.channel, err)
		}
	}
	return rp
}

func runMGEffectEnumGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/effect_enum_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_effect_enum.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildEffectEnumDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "effect_enum_in.aep")
	resavedAEP := filepath.Join(tempDir, "effect_enum_resaved.aep")
	doneFile := filepath.Join(tempDir, "effect_enum.done")
	framePNG := filepath.Join(tempDir, "effect_enum_frame.png")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"png":%q,"expect":[`+
		`{"layer":"RED","param":"ADBE Invert-0001","value":2},`+
		`{"layer":"GRN","param":"ADBE Invert-0001","value":3}]}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP), toFwd(framePNG))
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
	t.Logf("effect enum %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("effect enum %s ship gate FAIL:\n%s", ver, body)
	}

	f, err := os.Open(framePNG)
	if err != nil {
		t.Fatalf("%s rendered frame missing: %v", ver, err)
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		t.Fatalf("%s decode rendered frame: %v", ver, err)
	}
	const win = 10
	const loT, hiT = 110, 150
	for _, c := range effectEnumCells {
		r, g, b := avgRGB(img, c.cx, c.cy, win)
		t.Logf("%s %s (Channel=%v): avgRGB=(%d,%d,%d)", ver, c.name, c.channel, r, g, b)
		if c.wantRHi {
			// GRN: green inverted → (0.9,0.9,0.1): R untouched high, G high.
			if r < hiT || g < hiT {
				t.Errorf("%s %s (Channel=%v): avgRGB=(%d,%d,%d), want R&G high (>%d) — green-invert wrong", ver, c.name, c.channel, r, g, b, hiT)
			}
		} else {
			// RED: red inverted → (0.1,0.1,0.1): R low, G low.
			if r > loT || g > loT {
				t.Errorf("%s %s (Channel=%v): avgRGB=(%d,%d,%d), want R&G low (<%d) — red-invert wrong", ver, c.name, c.channel, r, g, b, loT)
			}
		}
	}
	// The R channel must split across the 128 midpoint between the two cards —
	// the direct proof the two distinct enum values both took effect.
	rRed, _, _ := avgRGB(img, effectEnumCells[0].cx, effectEnumCells[0].cy, win)
	rGrn, _, _ := avgRGB(img, effectEnumCells[1].cx, effectEnumCells[1].cy, win)
	if !(rRed < 128 && rGrn > 128) {
		t.Errorf("%s R-channel did not split: RED.R=%d GRN.R=%d (want RED<128<GRN)", ver, rRed, rGrn)
	}

	// Resave proof: the materialized cross-effect enum survives AE's re-encode.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	for _, c := range effectEnumCells {
		l := re.Compositions[0].LayerByName(c.name)
		if l == nil {
			t.Errorf("resaved: %s layer missing", c.name)
			continue
		}
		var got any
		for _, e := range l.Effects {
			for _, prm := range e.Parameters {
				if prm.MatchName == "ADBE Invert-0001" {
					got = prm.StaticValue
				}
			}
		}
		if !staticValueEq(got, c.channel) {
			t.Errorf("resaved %s Channel = %v, want %v (AE dropped/reverted the enum)", c.name, got, c.channel)
		}
	}
}

// TestMGEffectEnum_GoRoundTrip verifies the from-scratch build path (shapes →
// Reopen → AddEffect(Invert) → SetEffectParam enum) survives a Go write + parse
// round-trip — a cheap pre-check before the AE ship-gate burns cold-starts.
func TestMGEffectEnum_GoRoundTrip(t *testing.T) {
	rp := buildEffectEnumDemo(t, aep.TargetAE2020)
	dir := t.TempDir()
	fpath := filepath.Join(dir, "enum_rt.aep")
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
	for _, c := range effectEnumCells {
		l := re.Compositions[0].LayerByName(c.name)
		if l == nil {
			t.Fatalf("reopened: layer %s missing", c.name)
		}
		var got any
		found := false
		for _, e := range l.Effects {
			if e.MatchName != aep.EffectInvert {
				continue
			}
			for _, prm := range e.Parameters {
				if prm.MatchName == "ADBE Invert-0001" {
					got = prm.StaticValue
					found = true
				}
			}
		}
		if !found {
			t.Errorf("%s: Invert-0001 not materialized after round-trip", c.name)
			continue
		}
		if !staticValueEq(got, c.channel) {
			t.Errorf("%s: round-trip Channel = %v, want %v", c.name, got, c.channel)
		}
	}
}

func TestMGEffectEnum_AEShipGate_AE2020(t *testing.T) {
	runMGEffectEnumGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGEffectEnum_AEShipGate_AE2025(t *testing.T) {
	runMGEffectEnumGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
