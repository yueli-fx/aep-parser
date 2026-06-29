// internal/aep/add_mask_test.go
//
// Go round-trip tests for AddMask (no AE required). Proves the spliced mask
// atom triple survives WriteAEP → re-parse with the new mask present, its path
// bytes decoding back to the requested geometry, and the parade positioned
// per AE's group order. AE acceptance is covered separately by the ship-gate
// (add_mask_shipgate_test.go).
package aep_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// rectPath is the canonical test path: a 10..190 square in a 200×200 layer,
// zero tangents, closed.
func rectPath() aep.BezierPath {
	return aep.BezierPath{
		Vertices: [][2]float64{{10, 10}, {190, 10}, {190, 190}, {10, 190}},
		Closed:   true,
	}
}

// rectNormTriples is rectPath's expected decoded form: anchors bbox-normalized
// to 0..1 in input order; with zero tangents each entry's second pair (this
// vertex's out control) equals the anchor and the third pair (the NEXT
// vertex's in control) equals the next anchor — the mask ldat triple layout
// shared with "ADBE Vector Shape" (see decodeMaskVertices' field naming).
var rectNormAnchors = [][2]float64{{0, 0}, {1, 0}, {1, 1}, {0, 1}}

func assertRectMask(t *testing.T, m *aep.Mask) {
	t.Helper()
	if len(m.Vertices) != 4 {
		t.Fatalf("mask vertices = %d, want 4", len(m.Vertices))
	}
	for i, v := range m.Vertices {
		want := rectNormAnchors[i]
		next := rectNormAnchors[(i+1)%4]
		if v.Anchor != want {
			t.Errorf("v%d anchor = %v, want %v", i, v.Anchor, want)
		}
		if v.InTangent != want { // zero out-control → equals anchor
			t.Errorf("v%d out-control pair = %v, want %v", i, v.InTangent, want)
		}
		if v.OutTangent != next { // zero next-in-control → equals next anchor
			t.Errorf("v%d next-in-control pair = %v, want %v", i, v.OutTangent, next)
		}
	}
}

func maskParadeLayer(proj *aep.Project, id uint32) *aep.Layer {
	for _, c := range proj.Compositions {
		for _, l := range c.Layers {
			if l.ID == id {
				return l
			}
		}
	}
	return nil
}

func TestAddMask_RoundTrip(t *testing.T) {
	proj, err := aep.Open("../../test_data/generated/fixtures/re_property_struct_baseline.aep")
	if err != nil {
		t.Skipf("baseline not present: %v", err)
	}
	l := layerWithEffects(proj)
	if l == nil {
		t.Fatal("no layer with effects in baseline")
	}
	if l.MaskParade() != nil || len(l.Masks) != 0 {
		t.Fatal("baseline layer unexpectedly already has masks")
	}

	m, err := aep.AddMask(l, "Probe Mask", rectPath())
	if err != nil {
		t.Fatalf("AddMask: %v", err)
	}
	if m == nil {
		t.Fatal("AddMask returned nil mask")
	}
	if m.Name != "Probe Mask" || !m.Closed || m.Mode != aep.MaskModeAdd || m.Index != 1 {
		t.Errorf("mask = name %q closed %v mode %d index %d, want Probe Mask/true/Add/1", m.Name, m.Closed, m.Mode, m.Index)
	}
	if len(l.Masks) != 1 || l.Masks[0] != m {
		t.Errorf("layer.Masks should hold exactly the new mask")
	}

	// The new mask's setters must work immediately (back-refs wired).
	if err := m.SetInverted(true); err != nil {
		t.Fatalf("SetInverted on fresh mask: %v", err)
	}

	// Auto-created Mask Parade must precede the existing Effect Parade.
	maskIdx, fxIdx := -1, -1
	for i, c := range l.PropertyTree().Children {
		if g, ok := c.(*aep.AEPropertyGroup); ok {
			switch g.MatchName {
			case aep.MatchNameGroupMaskParade:
				maskIdx = i
			case aep.MatchNameGroupEffectParade:
				fxIdx = i
			}
		}
	}
	if maskIdx < 0 || fxIdx < 0 || maskIdx >= fxIdx {
		t.Errorf("mask parade idx %d / effect parade idx %d: Mask Parade must sit before Effect Parade", maskIdx, fxIdx)
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	rl := maskParadeLayer(re, l.ID)
	if rl == nil {
		t.Fatal("re-parsed: layer not found")
	}
	if len(rl.Masks) != 1 {
		t.Fatalf("re-parsed masks = %d, want 1", len(rl.Masks))
	}
	rm := rl.Masks[0]
	if rm.Name != "Probe Mask" {
		t.Errorf("re-parsed name = %q, want Probe Mask", rm.Name)
	}
	if !rm.Closed || rm.Mode != aep.MaskModeAdd || rm.Index != 1 || !rm.Inverted {
		t.Errorf("re-parsed mask = closed %v mode %d index %d inverted %v", rm.Closed, rm.Mode, rm.Index, rm.Inverted)
	}
	if rm.Opacity != 1.0 {
		t.Errorf("re-parsed opacity = %v, want default 1", rm.Opacity)
	}
	assertRectMask(t, rm)

	// The pre-existing effects are untouched.
	if got, want := paradeChildNames(rl), []string{"ADBE Gaussian Blur 2", "ADBE Tint", "ADBE Fill"}; !eq(got, want) {
		t.Errorf("effect parade after mask add = %v, want %v", got, want)
	}
}

func TestAddMask_SecondMask_IndexAndOrder(t *testing.T) {
	proj, err := aep.Open("../../test_data/generated/fixtures/re_property_struct_baseline.aep")
	if err != nil {
		t.Skipf("baseline not present: %v", err)
	}
	l := layerWithEffects(proj)
	if l == nil {
		t.Fatal("no layer with effects in baseline")
	}
	if _, err := aep.AddMask(l, "First", rectPath()); err != nil {
		t.Fatalf("AddMask #1: %v", err)
	}
	m2, err := aep.AddMask(l, "", rectPath())
	if err != nil {
		t.Fatalf("AddMask #2: %v", err)
	}
	if m2.Index != 2 {
		t.Errorf("second mask index = %d, want 2", m2.Index)
	}
	if m2.Name != "Mask 2" {
		t.Errorf("default name = %q, want Mask 2", m2.Name)
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	rl := maskParadeLayer(re, l.ID)
	if rl == nil || len(rl.Masks) != 2 {
		t.Fatalf("re-parsed masks = %v, want 2", rl)
	}
	if rl.Masks[0].Name != "First" || rl.Masks[1].Name != "Mask 2" {
		t.Errorf("re-parsed mask order = %q, %q", rl.Masks[0].Name, rl.Masks[1].Name)
	}
}

func TestAddMask_OpenPath(t *testing.T) {
	proj, err := aep.Open("../../test_data/generated/fixtures/re_property_struct_baseline.aep")
	if err != nil {
		t.Skipf("baseline not present: %v", err)
	}
	l := layerWithEffects(proj)
	if l == nil {
		t.Fatal("no layer with effects in baseline")
	}
	open := aep.BezierPath{Vertices: [][2]float64{{0, 0}, {100, 50}, {200, 0}}, Closed: false}
	if _, err := aep.AddMask(l, "Open", open); err != nil {
		t.Fatalf("AddMask: %v", err)
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	rl := maskParadeLayer(re, l.ID)
	if rl == nil || len(rl.Masks) != 1 {
		t.Fatal("re-parsed open mask not found")
	}
	if rl.Masks[0].Closed {
		t.Error("re-parsed mask Closed = true, want open")
	}
}

func TestAddMask_FreshLayerRefuse_ReopenWorks(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatal(err)
	}
	sl, err := aep.NewShapeLayer(comp, "S")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := aep.AddMask(sl.Layer, "M", rectPath()); err == nil {
		t.Fatal("AddMask on from-scratch layer: want error, got nil")
	} else if !strings.Contains(err.Error(), "Reopen") {
		t.Errorf("refuse error should mention the Reopen upgrade path, got: %v", err)
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
	m, err := aep.AddMask(l, "M", rectPath())
	if err != nil {
		t.Fatalf("AddMask (auto-create parade on reopened layer): %v", err)
	}
	assertRectMask(t, m)

	// Parade order: Mask Parade before Transform Group (no effect parade here).
	maskIdx, transformIdx := -1, -1
	for i, c := range l.PropertyTree().Children {
		if g, ok := c.(*aep.AEPropertyGroup); ok {
			switch g.MatchName {
			case aep.MatchNameGroupMaskParade:
				maskIdx = i
			case "ADBE Transform Group":
				transformIdx = i
			}
		}
	}
	if maskIdx < 0 || transformIdx < 0 || maskIdx >= transformIdx {
		t.Errorf("mask parade idx %d / transform idx %d: parade must sit before Transform Group", maskIdx, transformIdx)
	}

	var buf bytes.Buffer
	if err := rp.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	rl := maskParadeLayer(re, l.ID)
	if rl == nil || len(rl.Masks) != 1 {
		t.Fatal("re-parsed mask not found on reopened-layer project")
	}
	if rl.Masks[0].Name != "M" {
		t.Errorf("re-parsed name = %q, want M", rl.Masks[0].Name)
	}
	assertRectMask(t, rl.Masks[0])
}

func TestAddMask_RefuseCameraLight(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := aep.NewCameraLayer(comp, "Cam"); err != nil {
		t.Fatal(err)
	}
	if _, err := aep.NewLightLayer(comp, "Light"); err != nil {
		t.Fatal(err)
	}
	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	for _, c := range rp.Compositions {
		for _, l := range c.Layers {
			if _, err := aep.AddMask(l, "M", rectPath()); err == nil {
				t.Errorf("AddMask on %s layer %q: want refuse, got nil", l.Type, l.Name)
			}
		}
	}
}

func TestAddMask_EmptyPathRefused(t *testing.T) {
	proj, err := aep.Open("../../test_data/generated/fixtures/re_property_struct_baseline.aep")
	if err != nil {
		t.Skipf("baseline not present: %v", err)
	}
	l := layerWithEffects(proj)
	if l == nil {
		t.Fatal("no layer with effects in baseline")
	}
	if _, err := aep.AddMask(l, "E", aep.BezierPath{}); err == nil {
		t.Fatal("AddMask with empty path: want error, got nil")
	}
}
