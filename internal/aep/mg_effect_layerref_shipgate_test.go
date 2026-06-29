// internal/aep/mg_effect_layerref_shipgate_test.go
//
// Go round-trip + AE render gate for the LAYER-REFERENCE effect family
// (Displacement Map / Compound Blur / CC Vector Blur via SetEffectLayerParam —
// same tdpi-rewrite mechanism as Set Matte). On a 100% Go-built file:
//
//	HOST shape = white rect covering the LEFT half (sharp white|black vertical
//	seam at x=960, right half transparent→black). MAP shape = white rect over
//	the BOTTOM half, video OFF (referenced by the effect but not composited).
//	The effect's layer-ref param points at MAP, so the effect acts only where
//	MAP is white (bottom) — proving the reference took spatially:
//	  · Displacement Map: the bottom seam shifts horizontally vs the (untouched)
//	    top seam → big top/bottom luminance delta along the seam columns.
//	  · Compound Blur: the bottom seam blurs (white bleeds into the black side)
//	    while the top seam stays sharp → black-side bottom brighter than top.
//	Without the layer ref the effect would act uniformly (top==bottom) — red line 4.
//
// CC Vector Blur ships accept+round-trip+resave only (its gradient-driven blur
// has no clean spatial pixel proof here); render-pixel is honestly deferred.
//
// Gated by AE_SHIP_GATE. Uses test_data/generators/verify_effect_layerref.jsx.
package aep_test

import (
	"encoding/binary"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// buildLayerRefDemo builds the HOST(white-left) + MAP(white-bottom, hidden) comp
// with fxMatch on HOST whose refParam points at MAP; amtParam (if set) tunes the
// effect strength. Returns the project + MAP layer ID.
func buildLayerRefDemo(t *testing.T, target aep.AETarget, compName, fxMatch, refParam, amtParam string, amtVal float64) (*aep.Project, uint32) {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, compName, 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	host, err := aep.NewShapeLayer(comp, "HOST")
	if err != nil {
		t.Fatalf("NewShapeLayer HOST: %v", err)
	}
	hr, err := host.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("HOST AddRect: %v", err)
	}
	if err := hr.SetSize([2]float64{960, 1080}); err != nil {
		t.Fatalf("HOST SetSize: %v", err)
	}
	hf, err := host.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("HOST AddFill: %v", err)
	}
	if err := hf.SetColor([4]float64{1, 1, 1, 1}); err != nil {
		t.Fatalf("HOST SetColor: %v", err)
	}
	if err := host.Position().SetStaticValue([2]float64{480, 540}); err != nil { // white left half
		t.Fatalf("HOST Position: %v", err)
	}

	// BG = full-frame opaque black: hides MAP behind it so HOST's transparent
	// right half composites to black, not MAP-white showing through.
	if _, err := aep.NewSolidLayer(comp, "BG", 1920, 1080, [3]float64{0, 0, 0}); err != nil {
		t.Fatalf("NewSolidLayer BG: %v", err)
	}
	// MAP = full-frame opaque white solid: the effect reads it as a uniform
	// max-strength map (white=255). It stays video-ON (a hidden layer reads as
	// empty — RE'd: an invisible map yields zero displacement), parked at the
	// bottom behind BG so it does not composite over HOST.
	if _, err := aep.NewSolidLayer(comp, "MAP", 1920, 1080, [3]float64{1, 1, 1}); err != nil {
		t.Fatalf("NewSolidLayer MAP: %v", err)
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	c := rp.Compositions[0]
	hl := c.LayerByName("HOST")
	bg := c.LayerByName("BG")
	ml := c.LayerByName("MAP")
	if hl == nil || bg == nil || ml == nil {
		t.Fatal("reopened: HOST/BG/MAP missing")
	}
	// Force stack order HOST(top) · BG(mid) · MAP(bottom).
	if err := aep.MoveToBeginning(hl); err != nil {
		t.Fatalf("MoveToBeginning HOST: %v", err)
	}
	if err := aep.MoveToEnd(ml); err != nil {
		t.Fatalf("MoveToEnd MAP: %v", err)
	}
	if err := aep.MoveBefore(bg, ml); err != nil { // BG just above MAP
		t.Fatalf("MoveBefore BG: %v", err)
	}
	fx, err := aep.AddEffect(hl, fxMatch)
	if err != nil {
		t.Fatalf("AddEffect(%s): %v", fxMatch, err)
	}
	if err := aep.SetEffectLayerParam(hl, fx, refParam, ml); err != nil {
		t.Fatalf("SetEffectLayerParam(%s): %v", refParam, err)
	}
	if amtParam != "" {
		if _, err := aep.SetEffectParam(hl, fx, amtParam, amtVal); err != nil {
			t.Fatalf("SetEffectParam(%s): %v", amtParam, err)
		}
	}
	return rp, ml.ID
}

func layerRefRoundTrip(t *testing.T, compName, fxMatch, refParam, amtParam string, amtVal float64) {
	rp, mapID := buildLayerRefDemo(t, aep.TargetAE2020, compName, fxMatch, refParam, amtParam, amtVal)
	dir := t.TempDir()
	fpath := filepath.Join(dir, "layerref_rt.aep")
	out, err := os.Create(fpath)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	root := parseAEP(t, fpath)
	tdpi := findParamTdpi(root, refParam)
	if len(tdpi) < 4 {
		t.Fatalf("%s tdpi not found in written file", refParam)
	}
	if got := binary.BigEndian.Uint32(tdpi); got != mapID {
		t.Errorf("%s tdpi = %d, want MAP layer ID %d", refParam, got, mapID)
	}
}

func TestLayerRef_GoRoundTrip(t *testing.T) {
	layerRefRoundTrip(t, "DISPMAP", aep.EffectDisplacementMap, aep.EffectDisplacementMapLayer, "ADBE Displacement Map-0003", 180)
	layerRefRoundTrip(t, "CMPBLUR", aep.EffectCompoundBlur, aep.EffectCompoundBlurLayer, "ADBE Compound Blur-0002", 90)
	layerRefRoundTrip(t, "VECBLUR", aep.EffectCCVectorBlur, aep.EffectCCVectorBlurMap, "CC Vector Blur-0002", 120)
}

// Wave 11 layer-ref effects: each materialized layer-ref param's tdpi must
// round-trip to the MAP layer ID after AddEffect (retarget→host) +
// SetEffectLayerParam (→MAP). 3D Glasses / Timewarp expose two each — both tested.
func TestLayerRefWave11_GoRoundTrip(t *testing.T) {
	layerRefRoundTrip(t, "WARPSTAB", aep.EffectWarpStabilizer, aep.EffectWarpStabilizerRefLayer, "", 0)
	layerRefRoundTrip(t, "GLASSESL", aep.Effect3DGlasses, aep.Effect3DGlassesLeftView, "", 0)
	layerRefRoundTrip(t, "GLASSESR", aep.Effect3DGlasses, aep.Effect3DGlassesRightView, "", 0)
	layerRefRoundTrip(t, "TWMATTE", aep.EffectTimewarp, aep.EffectTimewarpMatteLayer, "", 0)
	layerRefRoundTrip(t, "TWSOURCE", aep.EffectTimewarp, aep.EffectTimewarpSourceLayer, "", 0)
	layerRefRoundTrip(t, "PARTWORLD", aep.EffectCCParticleWorld, aep.EffectCCParticleWorldTexture, "", 0)
}

func lum(r, g, b int) int { return (r + g + b) / 3 }

func runLayerRefGate(t *testing.T, aeExe, ver string, target aep.AETarget, compName, fxMatch, refParam, amtParam string, amtVal float64, pixelCheck func(*testing.T, image.Image, string)) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/effect_layerref_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_effect_layerref.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	rp, _ := buildLayerRefDemo(t, target, compName, fxMatch, refParam, amtParam, amtVal)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "layerref_in.aep")
	resavedAEP := filepath.Join(tempDir, "layerref_resaved.aep")
	doneFile := filepath.Join(tempDir, "layerref.done")
	framePNG := filepath.Join(tempDir, "layerref_frame.png")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"png":%q,"comp":%q,"host":"HOST","param":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP), toFwd(framePNG), compName, refParam)
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
	t.Logf("%s %s AE readback:\n%s", compName, ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("%s %s ship gate FAIL:\n%s", compName, ver, body)
	}

	if pixelCheck != nil {
		f, err := os.Open(framePNG)
		if err != nil {
			t.Fatalf("%s %s rendered frame missing: %v", compName, ver, err)
		}
		img, _, err := image.Decode(f)
		f.Close()
		if err != nil {
			t.Fatalf("%s %s decode frame: %v", compName, ver, err)
		}
		pixelCheck(t, img, ver)
	}

	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	hl := re.Compositions[0].LayerByName("HOST")
	if hl == nil {
		t.Fatal("resaved: HOST missing")
	}
	has := false
	for _, e := range hl.Effects {
		if e.MatchName == fxMatch {
			has = true
		}
	}
	if !has {
		t.Errorf("resaved: HOST lost its %s effect", fxMatch)
	}
}

// displacement pixel proof: a uniform white map shifts the whole HOST seam
// horizontally (white-left edge at x=960). Sampling ±90px around the seam, the
// shift (≥180px either direction) flips at least one point: 1050 black→white
// (positive shift) OR 870 white→black (negative shift). No displacement (effect
// dead) leaves 1050 black + 870 white → both fail.
func dispMapPixelCheck(t *testing.T, img image.Image, ver string) {
	rr, rg, rb := avgRGB(img, 1050, 540, 10) // 90px right of seam, originally black
	lr, lg, lb := avgRGB(img, 870, 540, 10)  // 90px left of seam, originally white
	right := lum(rr, rg, rb)
	left := lum(lr, lg, lb)
	t.Logf("%s DISPMAP x1050=%d (orig black) x870=%d (orig white)", ver, right, left)
	if !(right > 180 || left < 70) {
		t.Errorf("%s DISPMAP x1050=%d x870=%d: seam did not move (no point flipped) — displacement not rendered", ver, right, left)
	}
}

// compound-blur pixel proof: a uniform white map drives max blur, softening the
// sharp white|black seam at x=960 into a gradient — white bleeds into the black
// side (x=1010 lifts above pure black) and the white side dims (x=910 drops
// below pure white). A dead effect leaves a hard seam (1010=0, 910=255).
func compoundBlurPixelCheck(t *testing.T, img image.Image, ver string) {
	fr, fg, fb := avgRGB(img, 300, 540, 10) // far white, sanity
	wr, wg, wb := avgRGB(img, 910, 540, 8)  // white side near seam
	br, bg, bb := avgRGB(img, 1010, 540, 8) // black side near seam
	farWhite := lum(fr, fg, fb)
	nearWhite := lum(wr, wg, wb)
	nearBlack := lum(br, bg, bb)
	t.Logf("%s CMPBLUR farWhite=%d nearWhite=%d nearBlack=%d", ver, farWhite, nearWhite, nearBlack)
	if farWhite < 150 {
		t.Errorf("%s CMPBLUR far white=%d (<150): HOST not rendering as expected", ver, farWhite)
	}
	if !(nearBlack > 20 && nearWhite < 248) {
		t.Errorf("%s CMPBLUR nearBlack=%d (want >20) nearWhite=%d (want <248): seam not blurred — compound blur not rendered", ver, nearBlack, nearWhite)
	}
}

func TestLayerRefDispMap_AEShipGate_AE2020(t *testing.T) {
	runLayerRefGate(t, ae2020(), "AE2020", aep.TargetAE2020, "DISPMAP", aep.EffectDisplacementMap, aep.EffectDisplacementMapLayer, "ADBE Displacement Map-0003", 180, dispMapPixelCheck)
}
func TestLayerRefDispMap_AEShipGate_AE2025(t *testing.T) {
	runLayerRefGate(t, ae2025(), "AE2025", aep.TargetAE2025, "DISPMAP", aep.EffectDisplacementMap, aep.EffectDisplacementMapLayer, "ADBE Displacement Map-0003", 180, dispMapPixelCheck)
}
func TestLayerRefCompoundBlur_AEShipGate_AE2020(t *testing.T) {
	runLayerRefGate(t, ae2020(), "AE2020", aep.TargetAE2020, "CMPBLUR", aep.EffectCompoundBlur, aep.EffectCompoundBlurLayer, "ADBE Compound Blur-0002", 90, compoundBlurPixelCheck)
}
func TestLayerRefCompoundBlur_AEShipGate_AE2025(t *testing.T) {
	runLayerRefGate(t, ae2025(), "AE2025", aep.TargetAE2025, "CMPBLUR", aep.EffectCompoundBlur, aep.EffectCompoundBlurLayer, "ADBE Compound Blur-0002", 90, compoundBlurPixelCheck)
}

// CC Vector Blur: accept + round-trip + resave only (render-pixel deferred).
func TestLayerRefVectorBlur_AEShipGate_AE2020(t *testing.T) {
	runLayerRefGate(t, ae2020(), "AE2020", aep.TargetAE2020, "VECBLUR", aep.EffectCCVectorBlur, aep.EffectCCVectorBlurMap, "CC Vector Blur-0002", 120, nil)
}
func TestLayerRefVectorBlur_AEShipGate_AE2025(t *testing.T) {
	runLayerRefGate(t, ae2025(), "AE2025", aep.TargetAE2025, "VECBLUR", aep.EffectCCVectorBlur, aep.EffectCCVectorBlurMap, "CC Vector Blur-0002", 120, nil)
}
