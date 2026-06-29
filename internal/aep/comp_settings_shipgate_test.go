// internal/aep/comp_settings_shipgate_test.go
//
// AE ship gate for batch-6 Composition setters. Builds a from-scratch project
// with 5 comps, each exercising a slice of the comp-level setters (cdta fields),
// WriteAEP, and has AE read them back from CompItem DOM + resave so Go confirms
// survival. From-scratch is a valid carrier here: comp settings live in cdta
// (present on every comp, no elision), and NewComposition is already a
// double-version ae-accept path.
//
// Layout (decoupled so displayStart never collides with workArea):
//
//	CfgA — scalars/bools + SetWorkArea(seconds)
//	CfgB — SetWorkAreaStartFrame + SetWorkAreaEndFrame
//	CfgC — SetWorkAreaDurationFrame
//	CfgD — SetDisplayStartFrame
//	CfgE — SetDisplayStartTime
//
// SetComment is NOT here (item-comment idta flag needs its own RE — see plan).
//
// Gated by AE_SHIP_GATE.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func runCompSettingsGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/comp_settings_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_comp_settings.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := aep.NewProject(target)
	must := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	mkComp := func(tmpName string) *aep.Composition {
		c, err := aep.NewComposition(p, tmpName, 1280, 720, 30, 10)
		if err != nil {
			t.Fatalf("NewComposition(%s): %v", tmpName, err)
		}
		return c
	}

	// CfgA — scalars/bools + work area in seconds. Shutter setters touch the
	// work-area cdta bytes as a cosmetic side-effect, so set shutter BEFORE
	// SetWorkArea (incidents/shutter-side-effect-divisors.md).
	a := mkComp("TMP_A")
	must("A.SetName", a.SetName("CfgA"))
	must("A.SetSize", a.SetSize(1600, 900))
	// SetDuration at the creation 30fps (NO SetFrameRate here): SetDuration writes
	// a frame count = round(sec×FrameRate); SetFrameRate does NOT rescale the
	// duration ticks, so doing both on one comp makes AE read the count at the
	// wrong rate (8s→6.4s, real interaction bug). SetFrameRate is gated alone (CfgF).
	must("A.SetDuration", a.SetDuration(8))
	must("A.SetPixelAspect", a.SetPixelAspect(2.0))
	must("A.SetResolutionFactor", a.SetResolutionFactor(2, 2))
	must("A.SetBGColor", a.SetBGColor([3]uint8{255, 128, 0}))
	// SetLabel/SetComment write the item-level idta chunk, which a from-scratch
	// comp lacks ("no idta chunk reference") — gated separately in 6b on a
	// parsed comp carrier.
	must("A.SetShutterAngle", a.SetShutterAngle(360))
	must("A.SetShutterPhase", a.SetShutterPhase(-90))
	must("A.SetMotionBlurSamplesPerFrame", a.SetMotionBlurSamplesPerFrame(32))
	must("A.SetMotionBlurAdaptiveSampleLimit", a.SetMotionBlurAdaptiveSampleLimit(256))
	must("A.SetCompMotionBlur", a.SetCompMotionBlur(true))
	// SetDraft3D (@0x8A bit0) NOT gated: AE 2025 comp.draft3d DOM does not reflect
	// that bit (the 5 @0x8B flag bits below all do) — left roundtrip, needs RE.
	must("A.SetFrameBlending", a.SetFrameBlending(true))
	must("A.SetHideShyLayers", a.SetHideShyLayers(true))
	must("A.SetPreserveNestedFrameRate", a.SetPreserveNestedFrameRate(true))
	must("A.SetPreserveNestedResolution", a.SetPreserveNestedResolution(true))
	must("A.SetWorkArea", a.SetWorkArea(2.0, 6.0))

	// CfgB — work area via start/end frame (30fps → 1.0s / 4.0s).
	b := mkComp("TMP_B")
	must("B.SetName", b.SetName("CfgB"))
	must("B.SetWorkAreaStartFrame", b.SetWorkAreaStartFrame(30))
	must("B.SetWorkAreaEndFrame", b.SetWorkAreaEndFrame(120))

	// CfgC — work area via duration frame (45 frames → 1.5s).
	c := mkComp("TMP_C")
	must("C.SetName", c.SetName("CfgC"))
	must("C.SetWorkAreaDurationFrame", c.SetWorkAreaDurationFrame(45))

	// CfgD — display start via frame (15 → 0.5s).
	d := mkComp("TMP_D")
	must("D.SetName", d.SetName("CfgD"))
	must("D.SetDisplayStartFrame", d.SetDisplayStartFrame(15))

	// CfgE — display start via seconds.
	e := mkComp("TMP_E")
	must("E.SetName", e.SetName("CfgE"))
	must("E.SetDisplayStartTime", e.SetDisplayStartTime(2.0))

	// CfgF — SetFrameRate alone (decoupled from SetDuration). frameRate=24.
	f := mkComp("TMP_F")
	must("F.SetName", f.SetName("CfgF"))
	must("F.SetFrameRate", f.SetFrameRate(24))

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "comp_in.aep")
	resavedAEP := filepath.Join(tempDir, "comp_resaved.aep")
	doneFile := filepath.Join(tempDir, "comp.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP))
	if err := writeGeneratedArgs(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	body := string(content)
	t.Logf("%s comp settings AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("%s comp settings ship gate FAIL:\n%s", ver, body)
	}

	// Preservation proof: AE's resave kept CfgA's cdta scalar fields.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("re-open AE resave: %v", err)
	}
	var ra *aep.Composition
	for _, cc := range re.Compositions {
		if cc.Name == "CfgA" {
			ra = cc
		}
	}
	if ra == nil {
		t.Fatalf("%s resaved: CfgA missing", ver)
	}
	chkU16(t, ver, "Width", ra.Width, 1600)
	chkU16(t, ver, "Height", ra.Height, 900)
	chkF64(t, ver, "Duration", ra.Duration, 8) // CfgA kept at 30fps
	chkF64(t, ver, "PixelAspect", ra.PixelAspect, 2.0)
	if ra.ResolutionFactor[0] != 2 || ra.ResolutionFactor[1] != 2 {
		t.Errorf("%s resaved: ResolutionFactor = %v, want [2 2]", ver, ra.ResolutionFactor)
	}
	if ra.BGColor != [3]uint8{255, 128, 0} {
		t.Errorf("%s resaved: BGColor = %v, want [255 128 0]", ver, ra.BGColor)
	}
	if ra.ShutterAngle != 360 {
		t.Errorf("%s resaved: ShutterAngle = %d, want 360", ver, ra.ShutterAngle)
	}
	if ra.ShutterPhase != -90 {
		t.Errorf("%s resaved: ShutterPhase = %d, want -90", ver, ra.ShutterPhase)
	}
	if ra.MotionBlurSamplesPerFrame != 32 {
		t.Errorf("%s resaved: MotionBlurSamplesPerFrame = %d, want 32", ver, ra.MotionBlurSamplesPerFrame)
	}
	if ra.MotionBlurAdaptiveSampleLimit != 256 {
		t.Errorf("%s resaved: MotionBlurAdaptiveSampleLimit = %d, want 256", ver, ra.MotionBlurAdaptiveSampleLimit)
	}
	chkF64(t, ver, "WorkAreaStart", ra.WorkAreaStart, 2.0)
	chkF64(t, ver, "WorkAreaEnd", ra.WorkAreaEnd, 6.0)

	// SetFrameRate verified on CfgF (decoupled from SetDuration).
	var rf *aep.Composition
	for _, cc := range re.Compositions {
		if cc.Name == "CfgF" {
			rf = cc
		}
	}
	if rf == nil {
		t.Fatalf("%s resaved: CfgF missing", ver)
	}
	chkF64(t, ver, "F.FrameRate", rf.FrameRate, 24)
}

func chkU16(t *testing.T, ver, label string, got, want uint16) {
	t.Helper()
	if got != want {
		t.Errorf("%s resaved: %s = %d, want %d", ver, label, got, want)
	}
}

func chkF64(t *testing.T, ver, label string, got, want float64) {
	t.Helper()
	if got-want > 0.02 || want-got > 0.02 {
		t.Errorf("%s resaved: %s = %g, want %g", ver, label, got, want)
	}
}

func TestCompSettings_AEShipGate_AE2020(t *testing.T) {
	runCompSettingsGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}

func TestCompSettings_AEShipGate_AE2025(t *testing.T) {
	runCompSettingsGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
