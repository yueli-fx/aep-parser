package aep_test

import (
	"bytes"
	"math"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestMaskDecoding(t *testing.T) {
	rb := &rifxBuilder{}
	verts := []aep.MaskVertex{
		{Anchor: [2]float64{0, 0}, InTangent: [2]float64{0, 0}, OutTangent: [2]float64{1, 0}},
		{Anchor: [2]float64{1, 0}, InTangent: [2]float64{1, 0}, OutTangent: [2]float64{1, 1}},
		{Anchor: [2]float64{1, 1}, InTangent: [2]float64{1, 1}, OutTangent: [2]float64{0, 0}},
	}
	atom := buildMaskAtom(rb, verts, true, "test mask")

	// Outer "ADBE Mask Parade" tdgp wrapping the atom.
	var paradeBody []byte
	paradeBody = append(paradeBody, atom...)
	paradeBody = append(paradeBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
	paradeTdgp := rb.listChunk("LIST", "tdgp", paradeBody)

	var tdgpBody []byte
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Mask Parade"))...)
	tdgpBody = append(tdgpBody, paradeTdgp...)
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)

	data := wrapAsLayer(tdgpBody)
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	layer := proj.Compositions[0].Layers[0]
	if len(layer.Masks) != 1 {
		t.Fatalf("Masks = %d, want 1", len(layer.Masks))
	}
	mask := layer.Masks[0]
	if mask.Name != "test mask" {
		t.Errorf("Name = %q, want %q", mask.Name, "test mask")
	}
	if !mask.Closed {
		t.Error("Closed = false, want true")
	}
	if len(mask.Vertices) != 3 {
		t.Fatalf("Vertices = %d, want 3", len(mask.Vertices))
	}
	for i, want := range verts {
		got := mask.Vertices[i]
		// float32 round-trip: compare with tolerance.
		approx := func(a, b float64) bool { return math.Abs(a-b) < 1e-5 }
		if !approx(got.Anchor[0], want.Anchor[0]) || !approx(got.Anchor[1], want.Anchor[1]) {
			t.Errorf("vertex[%d] Anchor = %v, want %v", i, got.Anchor, want.Anchor)
		}
		if !approx(got.OutTangent[0], want.OutTangent[0]) || !approx(got.OutTangent[1], want.OutTangent[1]) {
			t.Errorf("vertex[%d] OutTangent = %v, want %v", i, got.OutTangent, want.OutTangent)
		}
	}
	if len(mask.MkifRaw) != 48 {
		t.Errorf("MkifRaw size = %d, want 48", len(mask.MkifRaw))
	}
	if len(mask.ShphRaw) != 24 {
		t.Errorf("ShphRaw size = %d, want 24", len(mask.ShphRaw))
	}
}

func TestMaskFeatherOpacityExpansion(t *testing.T) {
	rb := &rifxBuilder{}
	verts := []aep.MaskVertex{
		{Anchor: [2]float64{0, 0}, InTangent: [2]float64{0, 0}, OutTangent: [2]float64{1, 0}},
		{Anchor: [2]float64{1, 0}, InTangent: [2]float64{1, 0}, OutTangent: [2]float64{0, 0}},
	}
	feather := rb.leafStatic2D("ADBE Mask Feather", 30.0, 30.0)
	opacity := rb.leafStatic("ADBE Mask Opacity", 0x01, 0.6)
	expansion := rb.leafStatic("ADBE Mask Offset", 0x01, 90.0)
	mkifBytes := buildMkif(aep.MaskModeAdd, false, 1, 0xE4, 0xD8, 0x4C)
	atom := buildMaskAtomWithProps(rb, verts, true, "props mask", mkifBytes,
		feather, opacity, expansion)

	var paradeBody []byte
	paradeBody = append(paradeBody, atom...)
	paradeBody = append(paradeBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
	paradeTdgp := rb.listChunk("LIST", "tdgp", paradeBody)

	var tdgpBody []byte
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Mask Parade"))...)
	tdgpBody = append(tdgpBody, paradeTdgp...)
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)

	data := wrapAsLayer(tdgpBody)
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	layer := proj.Compositions[0].Layers[0]
	if len(layer.Masks) != 1 {
		t.Fatalf("Masks = %d, want 1", len(layer.Masks))
	}
	m := layer.Masks[0]
	if m.Feather != [2]float64{30.0, 30.0} {
		t.Errorf("Feather = %v, want (30, 30)", m.Feather)
	}
	if m.Opacity != 0.6 {
		t.Errorf("Opacity = %v, want 0.6", m.Opacity)
	}
	if m.Expansion != 90.0 {
		t.Errorf("Expansion = %v, want 90.0", m.Expansion)
	}
	// Properties slice should also include them.
	names := map[string]bool{}
	for _, p := range m.Properties {
		names[p.MatchName] = true
	}
	for _, n := range []string{"ADBE Mask Feather", "ADBE Mask Opacity", "ADBE Mask Offset"} {
		if !names[n] {
			t.Errorf("Mask.Properties missing %q", n)
		}
	}
}

func TestMaskMkifDecode(t *testing.T) {
	rb := &rifxBuilder{}
	verts := []aep.MaskVertex{
		{Anchor: [2]float64{0, 0}, InTangent: [2]float64{0, 0}, OutTangent: [2]float64{1, 0}},
		{Anchor: [2]float64{1, 0}, InTangent: [2]float64{1, 0}, OutTangent: [2]float64{0, 0}},
	}
	// Two masks: mask 1 default Add yellow, mask 2 Darken+Inverted peach.
	atom1 := buildMaskAtomWithMkif(rb, verts, true, "m1",
		buildMkif(aep.MaskModeAdd, false, 1, 0xE4, 0xD8, 0x4C))
	atom2 := buildMaskAtomWithMkif(rb, verts, true, "m2",
		buildMkif(aep.MaskModeDarken, true, 2, 0xE7, 0xC1, 0x9E))

	var paradeBody []byte
	paradeBody = append(paradeBody, atom1...)
	paradeBody = append(paradeBody, atom2...)
	paradeBody = append(paradeBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
	paradeTdgp := rb.listChunk("LIST", "tdgp", paradeBody)

	var tdgpBody []byte
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Mask Parade"))...)
	tdgpBody = append(tdgpBody, paradeTdgp...)
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)

	data := wrapAsLayer(tdgpBody)
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	layer := proj.Compositions[0].Layers[0]
	if len(layer.Masks) != 2 {
		t.Fatalf("Masks = %d, want 2", len(layer.Masks))
	}

	m1 := layer.Masks[0]
	if m1.Mode != aep.MaskModeAdd {
		t.Errorf("m1.Mode = %s, want add", m1.Mode)
	}
	if m1.Inverted {
		t.Errorf("m1.Inverted = true, want false")
	}
	if m1.Index != 1 {
		t.Errorf("m1.Index = %d, want 1", m1.Index)
	}
	if m1.Color != [3]uint8{0xE4, 0xD8, 0x4C} {
		t.Errorf("m1.Color = %v, want (0xE4,0xD8,0x4C)", m1.Color)
	}

	m2 := layer.Masks[1]
	if m2.Mode != aep.MaskModeDarken {
		t.Errorf("m2.Mode = %s, want darken", m2.Mode)
	}
	if !m2.Inverted {
		t.Errorf("m2.Inverted = false, want true")
	}
	if m2.Index != 2 {
		t.Errorf("m2.Index = %d, want 2", m2.Index)
	}
	if m2.Color != [3]uint8{0xE7, 0xC1, 0x9E} {
		t.Errorf("m2.Color = %v, want (0xE7,0xC1,0x9E)", m2.Color)
	}

	// Byte-perfect roundtrip — mkif bytes should survive untouched.
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), data) {
		t.Errorf("roundtrip length mismatch: in=%d out=%d", len(data), buf.Len())
	}
}

func TestMaskPathEasing(t *testing.T) {
	rb := &rifxBuilder{}
	v := []aep.MaskVertex{
		{Anchor: [2]float64{0, 0}, InTangent: [2]float64{0, 0}, OutTangent: [2]float64{1, 0}},
		{Anchor: [2]float64{1, 0}, InTangent: [2]float64{1, 0}, OutTangent: [2]float64{0, 0}},
	}
	snaps := []struct {
		Time     float64
		Vertices []aep.MaskVertex
	}{
		{Time: 0.0, Vertices: v},
		{Time: 2.0, Vertices: v},
	}
	atom := buildAnimatedMaskAtomEased(rb, snaps, true, "eased")

	var paradeBody []byte
	paradeBody = append(paradeBody, atom...)
	paradeBody = append(paradeBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
	paradeTdgp := rb.listChunk("LIST", "tdgp", paradeBody)

	var tdgpBody []byte
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Mask Parade"))...)
	tdgpBody = append(tdgpBody, paradeTdgp...)
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)

	data := wrapAsLayer(tdgpBody)
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	layer := proj.Compositions[0].Layers[0]
	if len(layer.Masks) != 1 || len(layer.Masks[0].PathKeyframes) != 2 {
		t.Fatalf("expected 1 mask with 2 path kfs")
	}
	for i, kf := range layer.Masks[0].PathKeyframes {
		if kf.InInterp != aep.InterpBezier || kf.OutInterp != aep.InterpBezier {
			t.Errorf("kf[%d] interp = %s/%s, want bezier/bezier", i, kf.InInterp, kf.OutInterp)
		}
		if kf.InTemporalEase.Influence != 0.333 || kf.OutTemporalEase.Influence != 0.333 {
			t.Errorf("kf[%d] influences = %v / %v, want 0.333", i, kf.InTemporalEase, kf.OutTemporalEase)
		}
	}
}

func TestMaskAnimation(t *testing.T) {
	rb := &rifxBuilder{}
	v1 := []aep.MaskVertex{
		{Anchor: [2]float64{0, 0}, InTangent: [2]float64{0, 0}, OutTangent: [2]float64{1, 0}},
		{Anchor: [2]float64{1, 0}, InTangent: [2]float64{1, 0}, OutTangent: [2]float64{0, 0}},
	}
	v2 := []aep.MaskVertex{
		{Anchor: [2]float64{0.5, 0.5}, InTangent: [2]float64{0.5, 0.5}, OutTangent: [2]float64{1, 1}},
		{Anchor: [2]float64{1, 1}, InTangent: [2]float64{1, 1}, OutTangent: [2]float64{0.5, 0.5}},
	}
	snaps := []struct {
		Time     float64
		Vertices []aep.MaskVertex
	}{
		{Time: 1.0, Vertices: v1},
		{Time: 3.0, Vertices: v2},
	}
	atom := buildAnimatedMaskAtom(rb, snaps, true, "anim mask")

	var paradeBody []byte
	paradeBody = append(paradeBody, atom...)
	paradeBody = append(paradeBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
	paradeTdgp := rb.listChunk("LIST", "tdgp", paradeBody)

	var tdgpBody []byte
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Mask Parade"))...)
	tdgpBody = append(tdgpBody, paradeTdgp...)
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)

	data := wrapAsLayer(tdgpBody)
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	layer := proj.Compositions[0].Layers[0]
	if len(layer.Masks) != 1 {
		t.Fatalf("Masks = %d, want 1", len(layer.Masks))
	}
	mask := layer.Masks[0]
	if mask.Name != "anim mask" {
		t.Errorf("Name = %q, want %q", mask.Name, "anim mask")
	}
	if len(mask.PathKeyframes) != 2 {
		t.Fatalf("PathKeyframes = %d, want 2", len(mask.PathKeyframes))
	}
	approx := func(a, b float64) bool { return math.Abs(a-b) < 1e-5 }
	if !approx(mask.PathKeyframes[0].Time, 1.0) || !approx(mask.PathKeyframes[1].Time, 3.0) {
		t.Errorf("times = %v / %v, want 1.0 / 3.0",
			mask.PathKeyframes[0].Time, mask.PathKeyframes[1].Time)
	}
	if len(mask.PathKeyframes[0].Vertices) != 2 || len(mask.PathKeyframes[1].Vertices) != 2 {
		t.Errorf("snapshot vertex counts wrong: %d / %d",
			len(mask.PathKeyframes[0].Vertices), len(mask.PathKeyframes[1].Vertices))
	}
	// Snapshot 1's anchor should match v2[0].
	if !approx(mask.PathKeyframes[1].Vertices[0].Anchor[0], 0.5) {
		t.Errorf("snapshot 1 vertex 0 anchor X = %v, want 0.5",
			mask.PathKeyframes[1].Vertices[0].Anchor[0])
	}
	// Mask.Vertices should mirror first snapshot.
	if !approx(mask.Vertices[0].Anchor[0], 0) || !approx(mask.Vertices[0].Anchor[1], 0) {
		t.Errorf("Vertices[0] = %v, want first snapshot (0,0)", mask.Vertices[0].Anchor)
	}
}

// TestMaskSetters drives Mask.SetMode / SetInverted / SetColor /
// SetClosed against a synthetic single-mask fixture and confirms the
// values survive a WriteAEP → FromReader roundtrip.
func TestMaskSetters(t *testing.T) {
	rb := &rifxBuilder{}
	verts := []aep.MaskVertex{
		{Anchor: [2]float64{0, 0}, OutTangent: [2]float64{1, 0}},
		{Anchor: [2]float64{1, 1}, InTangent: [2]float64{1, 1}, OutTangent: [2]float64{0, 0}},
	}
	atom := buildMaskAtom(rb, verts, true, "the mask")
	var paradeBody []byte
	paradeBody = append(paradeBody, atom...)
	paradeBody = append(paradeBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
	paradeTdgp := rb.listChunk("LIST", "tdgp", paradeBody)
	var tdgpBody []byte
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Mask Parade"))...)
	tdgpBody = append(tdgpBody, paradeTdgp...)
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
	data := wrapAsLayer(tdgpBody)

	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	masks := proj.Compositions[0].Layers[0].Masks
	if len(masks) != 1 {
		t.Fatalf("expected 1 mask, got %d", len(masks))
	}
	m := masks[0]

	mustNoErr := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	mustNoErr("SetMode", m.SetMode(aep.MaskModeSubtract))
	mustNoErr("SetInverted", m.SetInverted(true))
	mustNoErr("SetColor", m.SetColor([3]uint8{0x11, 0x22, 0x33}))
	mustNoErr("SetClosed(false)", m.SetClosed(false))

	if m.Mode != aep.MaskModeSubtract || !m.Inverted || m.Color != [3]uint8{0x11, 0x22, 0x33} || m.Closed {
		t.Errorf("in-mem after Set*: mode=%s inverted=%v color=%v closed=%v",
			m.Mode, m.Inverted, m.Color, m.Closed)
	}

	// Round-trip
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	m2 := proj2.Compositions[0].Layers[0].Masks[0]
	if m2.Mode != aep.MaskModeSubtract {
		t.Errorf("roundtrip Mode = %s", m2.Mode)
	}
	if !m2.Inverted {
		t.Errorf("roundtrip Inverted = false")
	}
	if m2.Color != [3]uint8{0x11, 0x22, 0x33} {
		t.Errorf("roundtrip Color = %v", m2.Color)
	}
	if m2.Closed {
		t.Errorf("roundtrip Closed = true, want false")
	}
}

func TestMaskSettersRejectMissingChunks(t *testing.T) {
	m := &aep.Mask{Name: "standalone"}
	if err := m.SetMode(aep.MaskModeAdd); err == nil {
		t.Error("SetMode on standalone mask: expected error")
	}
	if err := m.SetInverted(true); err == nil {
		t.Error("SetInverted on standalone mask: expected error")
	}
	if err := m.SetColor([3]uint8{1, 2, 3}); err == nil {
		t.Error("SetColor on standalone mask: expected error")
	}
	if err := m.SetClosed(true); err == nil {
		t.Error("SetClosed on standalone mask: expected error")
	}
}

// TestMaskExtraSetters covers Mask.SetLocked + SetMaskMotionBlur.
func TestMaskExtraSetters(t *testing.T) {
	rb := &rifxBuilder{}
	verts := []aep.MaskVertex{
		{Anchor: [2]float64{0, 0}}, {Anchor: [2]float64{1, 1}},
	}
	atom := buildMaskAtom(rb, verts, true, "test mask")
	var paradeBody []byte
	paradeBody = append(paradeBody, atom...)
	paradeBody = append(paradeBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
	paradeTdgp := rb.listChunk("LIST", "tdgp", paradeBody)
	var tdgpBody []byte
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Mask Parade"))...)
	tdgpBody = append(tdgpBody, paradeTdgp...)
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
	data := wrapAsLayer(tdgpBody)

	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	m := proj.Compositions[0].Layers[0].Masks[0]

	mustNoErr := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	mustNoErr("SetLocked", m.SetLocked(true))
	mustNoErr("SetMaskMotionBlur(On)", m.SetMaskMotionBlur(aep.MaskMotionBlurOn))

	// In-mem: setters now sync the Go-level Locked / MotionBlur fields.
	if !m.Locked {
		t.Errorf("in-mem Locked = false, want true")
	}
	if m.MotionBlur != aep.MaskMotionBlurOn {
		t.Errorf("in-mem MotionBlur = %d, want MaskMotionBlurOn", m.MotionBlur)
	}

	// Round-trip — assert via both raw mkif bytes and decoded fields.
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	m2 := proj2.Compositions[0].Layers[0].Masks[0]
	if m2.MkifRaw[0x01] != 1 {
		t.Errorf("roundtrip mkif @0x01 = %#x, want 1 (locked)", m2.MkifRaw[0x01])
	}
	if m2.MkifRaw[0x02] != 2 {
		t.Errorf("roundtrip mkif @0x02 = %#x, want 2 (MaskMotionBlurOn)", m2.MkifRaw[0x02])
	}
	if !m2.Locked {
		t.Errorf("roundtrip Locked = false, want true")
	}
	if m2.MotionBlur != aep.MaskMotionBlurOn {
		t.Errorf("roundtrip MotionBlur = %d, want MaskMotionBlurOn", m2.MotionBlur)
	}
}
