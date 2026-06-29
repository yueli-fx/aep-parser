package aep_test

import (
	"bytes"
	"math"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// TestPixelAspectReadReal verifies the cdta @0x90/@0x94 PAR read
// against re_batch3.aep's RE_B3_D_pixelAspect comp (PAR=2.0).
func TestPixelAspectReadReal(t *testing.T) {
	proj, err := aep.Open("../../test_data/fixtures/re_batch3.aep")
	if err != nil {
		t.Skipf("re_batch3.aep not present")
	}
	for _, c := range proj.Compositions {
		if c.Name == "RE_B3_D_pixelAspect" {
			if math.Abs(c.PixelAspect-2.0) > 1e-6 {
				t.Errorf("RE_B3_D_pixelAspect: PixelAspect = %g, want 2.0", c.PixelAspect)
			}
			return
		}
	}
	t.Skip("RE_B3_D_pixelAspect comp not found")
}

// TestResolutionFactorReal verifies cdta @0x00 / @0x02 read against
// re_cdta_probe.aep, where comps were explicitly set to varying
// resolutionFactor pairs.
func TestResolutionFactorReal(t *testing.T) {
	proj, err := aep.Open("../../test_data/fixtures/re_cdta_probe.aep")
	if err != nil {
		t.Skipf("re_cdta_probe.aep not present; run test_data/generators/re_cdta_probe.jsx in AE")
	}
	want := map[string][2]uint16{
		"A_baseline":           {1, 1},
		"B_resolution_half":    {2, 2},
		"C_resolution_quarter": {4, 4},
		"D_resolution_3x4":     {3, 4},
	}
	got := map[string][2]uint16{}
	for _, c := range proj.Compositions {
		if _, ok := want[c.Name]; ok {
			got[c.Name] = c.ResolutionFactor
		}
	}
	for name, w := range want {
		g, ok := got[name]
		if !ok {
			t.Errorf("comp %q not found in fixture", name)
			continue
		}
		if g != w {
			t.Errorf("comp %q: ResolutionFactor = %v, want %v", name, g, w)
		}
	}
}

// TestSetResolutionFactorRoundtrip writes a new resolution factor pair,
// re-serializes, re-parses, and confirms persistence + JSON output.
func TestSetResolutionFactorRoundtrip(t *testing.T) {
	data := buildMinimalAEP()
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	c := proj.Compositions[0]
	if err := c.SetResolutionFactor(3, 5); err != nil {
		t.Fatalf("SetResolutionFactor: %v", err)
	}
	if c.ResolutionFactor != [2]uint16{3, 5} {
		t.Errorf("in-mem after set: %v, want [3 5]", c.ResolutionFactor)
	}
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if got := re.Compositions[0].ResolutionFactor; got != [2]uint16{3, 5} {
		t.Errorf("after roundtrip: %v, want [3 5]", got)
	}
	// Reject 0 factors.
	if err := c.SetResolutionFactor(0, 1); err == nil {
		t.Error("SetResolutionFactor(0, 1) should error")
	}
	if err := c.SetResolutionFactor(1, 0); err == nil {
		t.Error("SetResolutionFactor(1, 0) should error")
	}
	// Reject on missing cdta.
	standalone := &aep.Composition{Name: "standalone"}
	if err := standalone.SetResolutionFactor(2, 2); err == nil {
		t.Error("SetResolutionFactor on missing cdta should error")
	}
}

// TestCompositionRendererReal verifies Composition.Renderer decode
// from PRin LIST → prin chunk @offset 4 (ASCII NUL-padded match-name)
// against re_renderer.aep, which has 3 comps with different renderers.
func TestCompositionRendererReal(t *testing.T) {
	proj, err := aep.Open("../../test_data/fixtures/re_renderer.aep")
	if err != nil {
		t.Skipf("re_renderer.aep not present; run test_data/generators/re_renderer.jsx in AE")
	}
	// In AE 2025, "ADBE Standard 3d" (Classic) and "ADBE Picasso" (Advanced)
	// both end up as "ADBE Escher" internally; "ADBE Ernst" stays Cinema 4D.
	want := map[string]string{
		"RDR_classic":  "ADBE Escher",
		"RDR_advanced": "ADBE Escher",
		"RDR_cinema4d": "ADBE Ernst",
	}
	for _, c := range proj.Compositions {
		if expected, ok := want[c.Name]; ok {
			if c.Renderer != expected {
				t.Errorf("comp %q: Renderer = %q, want %q", c.Name, c.Renderer, expected)
			}
		}
	}
}

// TestCompositionActiveCamera verifies ActiveCamera() picks the
// topmost enabled camera layer using re_cameralight.aep.
func TestCompositionActiveCamera(t *testing.T) {
	proj, err := aep.Open("../../test_data/fixtures/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present")
	}
	for _, c := range proj.Compositions {
		if c.Name != "RE_CL" {
			continue
		}
		cam := c.ActiveCamera()
		if cam == nil {
			t.Fatal("ActiveCamera() = nil, expected a camera layer")
		}
		if cam.Type != aep.LayerTypeCamera {
			t.Errorf("ActiveCamera type = %s, want camera", cam.Type)
		}
		if cam.Name != "MyCamera" {
			t.Errorf("ActiveCamera name = %q, want MyCamera", cam.Name)
		}
		return
	}
	t.Skip("RE_CL comp not found")
}

// TestCompositionFlagSetters covers the 4 cdta flag bit setters
// (Draft3D, HideShyLayers, CompMotionBlur, PreserveNestedFrameRate).
func TestCompositionFlagSetters(t *testing.T) {
	proj, err := aep.FromReader(bytes.NewReader(buildMinimalAEP()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	comp := proj.Compositions[0]

	mustNoErr := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	mustNoErr("SetHideShyLayers", comp.SetHideShyLayers(true))
	mustNoErr("SetCompMotionBlur", comp.SetCompMotionBlur(true))
	mustNoErr("SetPreserveNestedFrameRate", comp.SetPreserveNestedFrameRate(true))
	mustNoErr("SetDraft3D", comp.SetDraft3D(true))

	// Each flag occupies a separate bit — confirm they coexist by
	// reading back the cdta bytes.
	// (No public getter for these yet; just confirm WriteAEP + reparse
	// doesn't blow up.)
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	_, err = aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}

	// Flip back off; should be idempotent.
	mustNoErr("SetHideShyLayers(false)", comp.SetHideShyLayers(false))
	mustNoErr("SetCompMotionBlur(false)", comp.SetCompMotionBlur(false))
}

func TestDisplayStartTimeReal(t *testing.T) {
	proj := openWave2AE24(t)
	var baseline, dsf120 *aep.Composition
	for _, c := range proj.Compositions {
		switch c.Name {
		case "RE_CDTA_DSF":
			if baseline == nil || c.ID > baseline.ID {
				baseline = c
			}
		case "RE_CDTA_DSF_120":
			if dsf120 == nil || c.ID > dsf120.ID {
				dsf120 = c
			}
		}
	}
	if baseline == nil || dsf120 == nil {
		t.Fatalf("baseline / dsf120 missing")
	}
	if baseline.DisplayStartTime != 0 {
		t.Errorf("baseline DisplayStartTime = %g, want 0", baseline.DisplayStartTime)
	}
	// dsf120 was set to displayStartFrame=120 at 29.97 fps ≈ 4 sec.
	// AE 25's encoding produces ~3.98 sec (rounding quirks).
	if dsf120.DisplayStartTime < 3.9 || dsf120.DisplayStartTime > 4.1 {
		t.Errorf("dsf120 DisplayStartTime = %g, want ~4.0", dsf120.DisplayStartTime)
	}
}

func TestSetDisplayStartTimeRoundtrip(t *testing.T) {
	proj := openWave2AE24(t)
	var comp *aep.Composition
	for _, c := range proj.Compositions {
		if c.Name == "RE_CDTA_DSF" && (comp == nil || c.ID > comp.ID) {
			comp = c
		}
	}
	if comp == nil {
		t.Fatalf("RE_CDTA_DSF missing")
	}
	compID := comp.ID
	// Set to 2.5 seconds
	if err := comp.SetDisplayStartTime(2.5); err != nil {
		t.Fatalf("SetDisplayStartTime: %v", err)
	}
	if math.Abs(comp.DisplayStartTime-2.5) > 0.01 {
		t.Errorf("after Set: DisplayStartTime = %g, want ~2.5", comp.DisplayStartTime)
	}

	// Roundtrip
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	var found *aep.Composition
	for _, c := range proj2.Compositions {
		if c.ID == compID {
			found = c
			break
		}
	}
	if found == nil {
		t.Fatalf("roundtrip: comp (id=%d) missing", compID)
	}
	if math.Abs(found.DisplayStartTime-2.5) > 0.01 {
		t.Errorf("roundtrip DisplayStartTime = %g, want ~2.5", found.DisplayStartTime)
	}

	// Clear back to 0
	if err := found.SetDisplayStartTime(0); err != nil {
		t.Fatalf("clear: %v", err)
	}
	if found.DisplayStartTime != 0 {
		t.Errorf("after clear: DisplayStartTime = %g, want 0", found.DisplayStartTime)
	}

	// Frame helper
	if err := found.SetDisplayStartFrame(60); err != nil {
		t.Fatalf("SetDisplayStartFrame: %v", err)
	}
	wantSec := 60.0 / found.FrameRate
	if math.Abs(found.DisplayStartTime-wantSec) > 0.01 {
		t.Errorf("SetDisplayStartFrame(60): DisplayStartTime = %g, want ~%g", found.DisplayStartTime, wantSec)
	}

	// Reject negative
	if err := found.SetDisplayStartTime(-1); err == nil {
		t.Error("negative: expected error")
	}
}
