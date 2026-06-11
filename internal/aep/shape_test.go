package aep_test

import (
	"bytes"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestShapeLayerDetection(t *testing.T) {
	rb := &rifxBuilder{}
	// Empty Root Vectors Group is enough to flip the IsShapeLayer flag.
	innerTdgp := rb.listChunk("LIST", "tdgp", rb.chunk("tdmn", []byte("ADBE Group End")))
	var tdgpBody []byte
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Root Vectors Group"))...)
	tdgpBody = append(tdgpBody, innerTdgp...)
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)

	data := wrapAsLayer(tdgpBody)
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	layer := proj.Compositions[0].Layers[0]
	if !layer.IsShapeLayer {
		t.Error("IsShapeLayer = false, want true")
	}
	if layer.Type != aep.LayerTypeShape {
		t.Errorf("Type = %q, want %q", layer.Type, aep.LayerTypeShape)
	}
}

func TestShapePathExtraction(t *testing.T) {
	rb := &rifxBuilder{}
	// One Vector Shape - Group containing a single Vector Shape with a 3-vert path.
	verts := []aep.MaskVertex{
		{Anchor: [2]float64{0, 0}, InTangent: [2]float64{0, 0}, OutTangent: [2]float64{0.5, 0.1}},
		{Anchor: [2]float64{1, 0}, InTangent: [2]float64{0.5, 0.1}, OutTangent: [2]float64{1, 1}},
		{Anchor: [2]float64{1, 1}, InTangent: [2]float64{1, 1}, OutTangent: [2]float64{0, 0}},
	}
	data := buildShapeLayerData(rb, verts, true, "my path")
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	layer := proj.Compositions[0].Layers[0]
	if !layer.IsShapeLayer {
		t.Fatal("IsShapeLayer = false, want true")
	}
	if len(layer.ShapePaths) != 1 {
		t.Fatalf("ShapePaths = %d, want 1", len(layer.ShapePaths))
	}
	sp := layer.ShapePaths[0]
	if sp.Name != "my path" {
		t.Errorf("Name = %q, want %q", sp.Name, "my path")
	}
	if !sp.Closed {
		t.Error("Closed = false, want true")
	}
	if len(sp.Vertices) != 3 {
		t.Errorf("Vertices = %d, want 3", len(sp.Vertices))
	}
}

// buildShapeLayerData wraps a single Vector Shape path (verts/closed/name) in
// the Root Vectors Group scaffolding shape layers use, returning layer bytes
// for FromReader. om-s/omks/shap wrapping is identical to masks.
func buildShapeLayerData(rb *rifxBuilder, verts []aep.MaskVertex, closed bool, name string) []byte {
	omksList := rb.listChunk("LIST", "omks", buildMaskShapBytes(rb, verts, closed, name))
	tdb4 := rb.chunk("tdb4", buildTdb4(0x01))
	tdbsList := rb.listChunk("LIST", "tdbs", append([]byte(nil), tdb4...))
	var omSInner []byte
	omSInner = append(omSInner, tdbsList...)
	omSInner = append(omSInner, omksList...)
	omSList := rb.listChunk("LIST", "om-s", omSInner)

	var vsBody []byte
	vsBody = append(vsBody, rb.chunk("tdmn", []byte("ADBE Vector Shape"))...)
	vsBody = append(vsBody, omSList...)
	vsBody = append(vsBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
	vsTdgp := rb.listChunk("LIST", "tdgp", vsBody)

	var rootBody []byte
	rootBody = append(rootBody, rb.chunk("tdmn", []byte("ADBE Vector Shape - Group"))...)
	rootBody = append(rootBody, vsTdgp...)
	rootBody = append(rootBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
	rvgTdgp := rb.listChunk("LIST", "tdgp", rootBody)

	var tdgpBody []byte
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Root Vectors Group"))...)
	tdgpBody = append(tdgpBody, rvgTdgp...)
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
	return wrapAsLayer(tdgpBody)
}

// TestShapePathOpenIsNotClosed is a regression guard: an OPEN shape path
// (shph[3]=0x09, bit3 set) must decode to Closed=false. The old parser read
// shph[0x14] — a constant 0x01 on every AE-native path — and reported every
// shape path as closed (it only coincided on closed paths). Ground truth:
// v2_2_shape_path_re.aep carries open shaps with shph[3]=0x09, shph[0x14]=0x01.
func TestShapePathOpenIsNotClosed(t *testing.T) {
	rb := &rifxBuilder{}
	verts := []aep.MaskVertex{
		{Anchor: [2]float64{0, 0}},
		{Anchor: [2]float64{1, 0}},
		{Anchor: [2]float64{1, 1}},
	}
	data := buildShapeLayerData(rb, verts, false, "open path") // closed=false → shph[3]=0x09
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	layer := proj.Compositions[0].Layers[0]
	if len(layer.ShapePaths) != 1 {
		t.Fatalf("ShapePaths = %d, want 1", len(layer.ShapePaths))
	}
	if layer.ShapePaths[0].Closed {
		t.Error("Closed = true, want false (open shape path; shph[3]=0x09, [0x14]=0x01)")
	}
}

func TestShapePathExcludesMasks(t *testing.T) {
	// Build a shape layer that ALSO has an ADBE Mask Parade with a shap.
	// The mask shap must end up in Masks, not ShapePaths.
	rb := &rifxBuilder{}
	maskVerts := []aep.MaskVertex{
		{Anchor: [2]float64{0, 0}, InTangent: [2]float64{0, 0}, OutTangent: [2]float64{1, 1}},
		{Anchor: [2]float64{1, 1}, InTangent: [2]float64{1, 1}, OutTangent: [2]float64{0, 0}},
	}
	maskAtom := buildMaskAtom(rb, maskVerts, true, "mask in shape layer")
	var paradeBody []byte
	paradeBody = append(paradeBody, maskAtom...)
	paradeBody = append(paradeBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
	paradeTdgp := rb.listChunk("LIST", "tdgp", paradeBody)

	// Empty Root Vectors Group → just flips IsShapeLayer.
	emptyRVG := rb.listChunk("LIST", "tdgp", rb.chunk("tdmn", []byte("ADBE Group End")))

	var tdgpBody []byte
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Mask Parade"))...)
	tdgpBody = append(tdgpBody, paradeTdgp...)
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Root Vectors Group"))...)
	tdgpBody = append(tdgpBody, emptyRVG...)
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)

	data := wrapAsLayer(tdgpBody)
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	layer := proj.Compositions[0].Layers[0]
	if !layer.IsShapeLayer {
		t.Error("IsShapeLayer = false, want true")
	}
	if len(layer.Masks) != 1 {
		t.Errorf("Masks = %d, want 1", len(layer.Masks))
	}
	if len(layer.ShapePaths) != 0 {
		t.Errorf("ShapePaths = %d, want 0 (mask shap should NOT be counted as shape path)", len(layer.ShapePaths))
	}
}

// TestShapePrimitivesReal verifies Rect / Ellipse / Star primitive
// decoding against the AE 2020 fixture re_shapes.aep (built by
// /tmp/re_shapes.jsx). Skips silently if absent.
func TestShapePrimitivesReal(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_shapes.aep")
	if err != nil {
		t.Skipf("re_shapes.aep not present; rerun re_shapes.jsx")
	}
	var comp *aep.Composition
	for _, c := range proj.Compositions {
		if c.Name == "RE_SHAPES" {
			comp = c
			break
		}
	}
	if comp == nil {
		t.Fatal("RE_SHAPES comp missing")
	}
	var shape *aep.Layer
	for _, l := range comp.Layers {
		if l.IsShapeLayer {
			shape = l
			break
		}
	}
	if shape == nil {
		t.Fatal("no shape layer in fixture")
	}
	if len(shape.ShapePrimitives) != 3 {
		t.Fatalf("ShapePrimitives = %d, want 3", len(shape.ShapePrimitives))
	}

	getStatic := func(p *aep.Property) any {
		if p == nil {
			return nil
		}
		return p.StaticValue
	}

	rect := shape.ShapePrimitives[0]
	if rect.Kind != aep.ShapePrimitiveRect || rect.GroupName != "RectGroup" {
		t.Errorf("rect: kind=%s group=%q", rect.Kind, rect.GroupName)
	}
	if v, ok := getStatic(rect.Size).([]float64); !ok || v[0] != 400 || v[1] != 200 {
		t.Errorf("rect Size = %v", getStatic(rect.Size))
	}
	if v, ok := getStatic(rect.Roundness).(float64); !ok || v != 20 {
		t.Errorf("rect Roundness = %v", getStatic(rect.Roundness))
	}

	ellipse := shape.ShapePrimitives[1]
	if ellipse.Kind != aep.ShapePrimitiveEllipse || ellipse.GroupName != "EllipseGroup" {
		t.Errorf("ellipse: kind=%s group=%q", ellipse.Kind, ellipse.GroupName)
	}
	if v, ok := getStatic(ellipse.Size).([]float64); !ok || v[0] != 300 || v[1] != 150 {
		t.Errorf("ellipse Size = %v", getStatic(ellipse.Size))
	}
	if v, ok := getStatic(ellipse.Position).([]float64); !ok || v[0] != 100 || v[1] != 50 {
		t.Errorf("ellipse Position = %v", getStatic(ellipse.Position))
	}

	star := shape.ShapePrimitives[2]
	if star.Kind != aep.ShapePrimitiveStar || star.GroupName != "StarGroup" {
		t.Errorf("star: kind=%s group=%q", star.Kind, star.GroupName)
	}
	if v, ok := getStatic(star.Points).(float64); !ok || v != 7 {
		t.Errorf("star Points = %v", getStatic(star.Points))
	}
	if v, ok := getStatic(star.Rotation).(float64); !ok || v != 45 {
		t.Errorf("star Rotation = %v", getStatic(star.Rotation))
	}
	if v, ok := getStatic(star.OuterRadius).(float64); !ok || v != 150 {
		t.Errorf("star OuterRadius = %v", getStatic(star.OuterRadius))
	}
	if v, ok := getStatic(star.InnerRoundness).(float64); !ok || v != 30 {
		t.Errorf("star InnerRoundness = %v", getStatic(star.InnerRoundness))
	}
	if v, ok := getStatic(star.OuterRoundness).(float64); !ok || v != 60 {
		t.Errorf("star OuterRoundness = %v", getStatic(star.OuterRoundness))
	}

	// The primitives' Property objects share chunk refs with the flat
	// layer.Properties list — exercise the write path via SetStaticValue
	// and confirm both views see the change after roundtrip.
	if err := rect.Size.SetStaticValue([]float64{500, 250}); err != nil {
		t.Fatalf("SetStaticValue: %v", err)
	}
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	var rt *aep.Composition
	for _, c := range proj2.Compositions {
		if c.Name == "RE_SHAPES" {
			rt = c
			break
		}
	}
	for _, l := range rt.Layers {
		if !l.IsShapeLayer {
			continue
		}
		if len(l.ShapePrimitives) == 0 {
			t.Fatal("after roundtrip: no primitives")
		}
		got, _ := getStatic(l.ShapePrimitives[0].Size).([]float64)
		if got == nil || got[0] != 500 || got[1] != 250 {
			t.Errorf("after roundtrip: rect Size = %v, want [500 250]", got)
		}
	}
}
