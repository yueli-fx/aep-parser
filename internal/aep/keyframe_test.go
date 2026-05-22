package aep_test

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// buildKeyframedAEP wraps three test properties (Opacity 1D + Position 3D
// spatial + Glossiness Coefficient 1D static) into a full RIFX file.
// Shared scaffolding for the keyframe roundtrip + insert/delete tests.
func buildKeyframedAEP() []byte {
	rb := &rifxBuilder{}

	// One Opacity keyframe pair: t=1.0s v=1.0, t=3.0s v=0.5.
	opacity := rb.leafKeyframed("ADBE Opacity", 0x01, 48, [][]byte{
		buildKF1D(1.0, 1.0),
		buildKF1D(3.0, 0.5),
	})
	// One Position keyframe pair (spatial): t=0.0s v=[100,200,0], t=2.0s v=[300,400,0].
	position := rb.leafKeyframed("ADBE Position", 0x03, 128, [][]byte{
		buildKF3DSpatial(0.0, 100, 200, 0),
		buildKF3DSpatial(2.0, 300, 400, 0),
	})
	// One static Glossiness Coefficient.
	gloss := rb.leafStatic("ADBE Glossiness Coefficient", 0x01, 80.0)

	// Property group: tdmn entries interleaved with payload, terminated by
	// "ADBE Group End" tdmn.
	var groupBody []byte
	groupBody = append(groupBody, opacity...)
	groupBody = append(groupBody, position...)
	groupBody = append(groupBody, gloss...)
	groupBody = append(groupBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
	tdgpList := rb.listChunk("LIST", "tdgp", groupBody)

	// Layer: ldta (minimal — SourceID at 0x28) + property tdgp.
	ldtaData := make([]byte, 0x2C)
	binary.BigEndian.PutUint32(ldtaData[0x28:], 99) // SourceID
	ldta := rb.chunk("ldta", ldtaData)
	var layerBody []byte
	layerBody = append(layerBody, ldta...)
	layerBody = append(layerBody, tdgpList...)
	layrList := rb.listChunk("LIST", "Layr", layerBody)

	// Comp Item: Utf8 + idta + cdta + Layr.
	compName := rb.chunk("Utf8", []byte("KF Comp"))
	compIdta := rb.chunk("idta", buildIdta(0x04, 1))
	compCdta := rb.chunk("cdta", buildCdta(1920, 1080, 30, 0, 300))
	var compBody []byte
	compBody = append(compBody, compName...)
	compBody = append(compBody, compIdta...)
	compBody = append(compBody, compCdta...)
	compBody = append(compBody, layrList...)
	compItem := rb.listChunk("LIST", "Item", compBody)

	foldList := rb.listChunk("LIST", "Fold", compItem)

	var root bytes.Buffer
	root.WriteString("RIFX")
	_ = binary.Write(&root, binary.BigEndian, uint32(4+len(foldList)))
	root.WriteString("Egg!")
	root.Write(foldList)
	return root.Bytes()
}

// TestKeyframeSetValueUpdatesInMemory pins the regression for the 1D
// SetValue path: bytes AND k.Value must both be updated. Earlier the
// scalar case returned early after writing bytes, leaving k.Value stale.
func TestKeyframeSetValueUpdatesInMemory(t *testing.T) {
	rb := &rifxBuilder{}
	op := rb.leafKeyframed("ADBE Opacity", 0x01, 48, [][]byte{
		buildKF1D(0.0, 1.0),
	})
	var tdgpBody []byte
	tdgpBody = append(tdgpBody, op...)
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
	data := wrapAsLayer(tdgpBody)
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	kf := proj.Compositions[0].Layers[0].Opacity().Keyframes[0]
	if err := kf.SetValue(0.42); err != nil {
		t.Fatalf("SetValue: %v", err)
	}
	if got := kf.Value.(float64); got != 0.42 {
		t.Errorf("after SetValue, k.Value = %v, want 0.42 (in-memory state diverged from bytes)", got)
	}
	var out bytes.Buffer
	if err := proj.WriteAEP(&out); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, _ := aep.FromReader(bytes.NewReader(out.Bytes()))
	got2 := proj2.Compositions[0].Layers[0].Opacity().Keyframes[0].Value.(float64)
	if got2 != 0.42 {
		t.Errorf("after roundtrip, k.Value = %v, want 0.42", got2)
	}
}

func TestKeyframeRoundtrip(t *testing.T) {
	data := buildKeyframedAEP()
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	if len(proj.Compositions) != 1 || len(proj.Compositions[0].Layers) != 1 {
		t.Fatalf("expected 1 comp/1 layer, got %d/%d",
			len(proj.Compositions),
			func() int {
				if len(proj.Compositions) == 0 {
					return 0
				}
				return len(proj.Compositions[0].Layers)
			}())
	}
	layer := proj.Compositions[0].Layers[0]

	props := map[string]*aep.Property{}
	for _, p := range layer.Properties {
		props[p.MatchName] = p
	}
	for _, name := range []string{"ADBE Opacity", "ADBE Position", "ADBE Glossiness Coefficient"} {
		if props[name] == nil {
			t.Fatalf("missing property %q on layer", name)
		}
	}

	op := props["ADBE Opacity"]
	if op.Components != 1 {
		t.Errorf("Opacity Components = %d, want 1", op.Components)
	}
	if len(op.Keyframes) != 2 {
		t.Fatalf("Opacity keyframes = %d, want 2", len(op.Keyframes))
	}
	if got, want := op.Keyframes[0].Time, 1.0; math.Abs(got-want) > 1e-9 {
		t.Errorf("Opacity kf[0].Time = %v, want %v", got, want)
	}
	if got, want := op.Keyframes[0].Value.(float64), 1.0; got != want {
		t.Errorf("Opacity kf[0].Value = %v, want %v", got, want)
	}
	if got, want := op.Keyframes[1].Value.(float64), 0.5; got != want {
		t.Errorf("Opacity kf[1].Value = %v, want %v", got, want)
	}

	pos := props["ADBE Position"]
	if pos.Components != 3 {
		t.Errorf("Position Components = %d, want 3", pos.Components)
	}
	wantPos0 := []float64{100, 200, 0}
	wantPos1 := []float64{300, 400, 0}
	gotPos0 := pos.Keyframes[0].Value.([]float64)
	gotPos1 := pos.Keyframes[1].Value.([]float64)
	for i := range wantPos0 {
		if gotPos0[i] != wantPos0[i] {
			t.Errorf("Position kf[0][%d] = %v, want %v", i, gotPos0[i], wantPos0[i])
		}
		if gotPos1[i] != wantPos1[i] {
			t.Errorf("Position kf[1][%d] = %v, want %v", i, gotPos1[i], wantPos1[i])
		}
	}

	gloss := props["ADBE Glossiness Coefficient"]
	if len(gloss.Keyframes) != 0 {
		t.Errorf("Glossiness should have no keyframes, got %d", len(gloss.Keyframes))
	}
	if got, want := gloss.StaticValue.(float64), 80.0; got != want {
		t.Errorf("Glossiness static = %v, want %v", got, want)
	}

	if err := op.Keyframes[0].SetTime(2.5); err != nil {
		t.Fatalf("Opacity SetTime: %v", err)
	}
	if err := op.Keyframes[1].SetValue(0.25); err != nil {
		t.Fatalf("Opacity SetValue: %v", err)
	}
	if err := pos.Keyframes[1].SetValue([]float64{500, 600, 0}); err != nil {
		t.Fatalf("Position SetValue: %v", err)
	}
	if err := gloss.SetStaticValue(42.0); err != nil {
		t.Fatalf("Glossiness SetStaticValue: %v", err)
	}

	if err := pos.Keyframes[0].SetValue(123.0); err == nil {
		t.Error("expected error setting scalar on 3D property")
	}
	if err := op.Keyframes[0].SetValue([]float64{1, 2, 3}); err == nil {
		t.Error("expected error setting 3D slice on 1D property")
	}
	if err := op.SetStaticValue(0.0); err == nil {
		t.Error("expected error calling SetStaticValue on keyframed property")
	}

	var out bytes.Buffer
	if err := proj.WriteAEP(&out); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	if out.Len() != len(data) {
		t.Errorf("roundtrip length changed: in=%d out=%d (expected byte-preserving)", len(data), out.Len())
	}

	proj2, err := aep.FromReader(bytes.NewReader(out.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	layer2 := proj2.Compositions[0].Layers[0]
	props2 := map[string]*aep.Property{}
	for _, p := range layer2.Properties {
		props2[p.MatchName] = p
	}
	if got, want := props2["ADBE Opacity"].Keyframes[0].Time, 2.5; math.Abs(got-want) > 1e-9 {
		t.Errorf("after roundtrip Opacity kf[0].Time = %v, want 2.5", got)
	}
	if got, want := props2["ADBE Opacity"].Keyframes[1].Value.(float64), 0.25; got != want {
		t.Errorf("after roundtrip Opacity kf[1].Value = %v, want 0.25", got)
	}
	gotPos := props2["ADBE Position"].Keyframes[1].Value.([]float64)
	wantPos := []float64{500, 600, 0}
	for i := range wantPos {
		if gotPos[i] != wantPos[i] {
			t.Errorf("after roundtrip Position kf[1][%d] = %v, want %v", i, gotPos[i], wantPos[i])
		}
	}
	if got, want := props2["ADBE Glossiness Coefficient"].StaticValue.(float64), 42.0; got != want {
		t.Errorf("after roundtrip Glossiness static = %v, want 42", got)
	}
}

func TestKeyframeEasingSpatial(t *testing.T) {
	rb := &rifxBuilder{}
	kf := buildKF3DSpatialBezier(1.0,
		[3]float64{100, 200, 0},
		[3]float64{-10, -20, 0},
		[3]float64{30, 40, 0},
		0.5, 0.25)
	position := rb.leafKeyframed("ADBE Position", 0x03, 128, [][]byte{kf})

	var tdgpBody []byte
	tdgpBody = append(tdgpBody, position...)
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)

	data := wrapAsLayer(tdgpBody)
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	layer := proj.Compositions[0].Layers[0]
	pos := layer.Position()
	if pos == nil || len(pos.Keyframes) != 1 {
		t.Fatalf("Position absent or wrong kf count: %+v", pos)
	}
	k := pos.Keyframes[0]
	if k.InInterp != aep.InterpBezier || k.OutInterp != aep.InterpBezier {
		t.Errorf("interp = %s/%s, want bezier/bezier", k.InInterp, k.OutInterp)
	}
	wantIn := []float64{-10, -20, 0}
	wantOut := []float64{30, 40, 0}
	if len(k.InSpatialTangent) != 3 || len(k.OutSpatialTangent) != 3 {
		t.Fatalf("tangent lengths in=%d out=%d, want 3", len(k.InSpatialTangent), len(k.OutSpatialTangent))
	}
	for i := 0; i < 3; i++ {
		if k.InSpatialTangent[i] != wantIn[i] {
			t.Errorf("InSpatialTangent[%d] = %v, want %v", i, k.InSpatialTangent[i], wantIn[i])
		}
		if k.OutSpatialTangent[i] != wantOut[i] {
			t.Errorf("OutSpatialTangent[%d] = %v, want %v", i, k.OutSpatialTangent[i], wantOut[i])
		}
	}
	if len(k.InTemporalEase) != 1 || k.InTemporalEase[0].Influence != 0.5 {
		t.Errorf("InTemporalEase = %v, want [{Speed:0 Influence:0.5}]", k.InTemporalEase)
	}
	if len(k.OutTemporalEase) != 1 || k.OutTemporalEase[0].Influence != 0.25 {
		t.Errorf("OutTemporalEase = %v, want [{Speed:0 Influence:0.25}]", k.OutTemporalEase)
	}
}

func TestNonSpatial2DKeyframe(t *testing.T) {
	rb := &rifxBuilder{}
	// 2D Feather-style property: tdb4[3]=0x02, bpk=88.
	kf0 := buildKF2DNonSpatial(0.0, 10, 20, 0, 0, 0.1, 0.2, 0, 0, 0.3, 0.4)
	kf1 := buildKF2DNonSpatial(1.0, 60, 80, 0, 0, 0.5, 0.6, 0, 0, 0.7, 0.8)
	prop := rb.leafKeyframed("ADBE Mask Feather", 0x02, 88, [][]byte{kf0, kf1})

	var tdgpBody []byte
	tdgpBody = append(tdgpBody, prop...)
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
	data := wrapAsLayer(tdgpBody)

	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	layer := proj.Compositions[0].Layers[0]
	var p *aep.Property
	for _, pp := range layer.Properties {
		if pp.MatchName == "ADBE Mask Feather" {
			p = pp
		}
	}
	if p == nil {
		t.Fatal("Feather property missing")
	}
	if p.Components != 2 {
		t.Errorf("Components = %d, want 2", p.Components)
	}
	if len(p.Keyframes) != 2 {
		t.Fatalf("Keyframes = %d, want 2", len(p.Keyframes))
	}
	got := p.Keyframes[1].Value.([]float64)
	if got[0] != 60 || got[1] != 80 {
		t.Errorf("kf[1].Value = %v, want [60 80]", got)
	}
	in := p.Keyframes[0].InTemporalEase
	if len(in) != 2 {
		t.Fatalf("InTemporalEase length = %d, want 2 (per-component)", len(in))
	}
	if in[0].Influence != 0.1 || in[1].Influence != 0.2 {
		t.Errorf("InTemporalEase influences = (%v, %v), want (0.1, 0.2)",
			in[0].Influence, in[1].Influence)
	}
	out := p.Keyframes[0].OutTemporalEase
	if out[0].Influence != 0.3 || out[1].Influence != 0.4 {
		t.Errorf("OutTemporalEase influences = (%v, %v), want (0.3, 0.4)",
			out[0].Influence, out[1].Influence)
	}
}

func TestColor4DKeyframe(t *testing.T) {
	rb := &rifxBuilder{}
	// Color property: tdb4[3]=0x04 → Components=4, bpk=152.
	kfs := [][]byte{
		buildKF4DColor(0.0, 255, 255, 0, 0),
		buildKF4DColor(2.0, 100, 50, 200, 128),
	}
	color := rb.leafKeyframed("ADBE Tritone-0001", 0x04, 152, kfs)

	var tdgpBody []byte
	tdgpBody = append(tdgpBody, color...)
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
	data := wrapAsLayer(tdgpBody)

	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	layer := proj.Compositions[0].Layers[0]
	var p *aep.Property
	for _, pp := range layer.Properties {
		if pp.MatchName == "ADBE Tritone-0001" {
			p = pp
		}
	}
	if p == nil {
		t.Fatal("color property not found")
	}
	if p.Components != 4 {
		t.Errorf("Components = %d, want 4", p.Components)
	}
	if len(p.Keyframes) != 2 {
		t.Fatalf("Keyframes = %d, want 2", len(p.Keyframes))
	}
	got1 := p.Keyframes[1].Value.([]float64)
	want1 := []float64{100, 50, 200, 128}
	for i, w := range want1 {
		if got1[i] != w {
			t.Errorf("kf[1].Value[%d] = %v, want %v", i, got1[i], w)
		}
	}
	if len(p.Keyframes[0].InTemporalEase) != 1 {
		t.Errorf("InTemporalEase length = %d, want 1 (scalar for 4D color)",
			len(p.Keyframes[0].InTemporalEase))
	}
	if p.Keyframes[0].InTemporalEase[0].Influence != 0.333 {
		t.Errorf("InInfluence = %v, want 0.333", p.Keyframes[0].InTemporalEase[0].Influence)
	}
	if len(p.Keyframes[0].InSpatialTangent) != 4 {
		t.Errorf("InSpatialTangent length = %d, want 4", len(p.Keyframes[0].InSpatialTangent))
	}
}

func TestHoldInterpolation(t *testing.T) {
	rb := &rifxBuilder{}
	kf0 := buildKF1D(0.0, 1.0)
	kf0[0x04] = byte(aep.InterpBezier)
	kf0[0x05] = byte(aep.InterpHold)
	kf1 := buildKF1D(2.0, 0.0)
	kf1[0x04] = byte(aep.InterpHold)
	kf1[0x05] = byte(aep.InterpHold)

	op := rb.leafKeyframed("ADBE Opacity", 0x01, 48, [][]byte{kf0, kf1})
	var tdgpBody []byte
	tdgpBody = append(tdgpBody, op...)
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
	data := wrapAsLayer(tdgpBody)
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	kfs := proj.Compositions[0].Layers[0].Opacity().Keyframes
	if kfs[0].InInterp != aep.InterpBezier || kfs[0].OutInterp != aep.InterpHold {
		t.Errorf("kf[0] = %s/%s, want bezier/hold", kfs[0].InInterp, kfs[0].OutInterp)
	}
	if kfs[1].InInterp != aep.InterpHold || kfs[1].OutInterp != aep.InterpHold {
		t.Errorf("kf[1] = %s/%s, want hold/hold", kfs[1].InInterp, kfs[1].OutInterp)
	}
	if aep.InterpHold.String() != "hold" {
		t.Errorf("InterpHold.String() = %q, want hold", aep.InterpHold.String())
	}
}

func TestNonSpatialBezierEasing(t *testing.T) {
	rb := &rifxBuilder{}
	op := rb.leafKeyframed("ADBE Opacity", 0x01, 48, [][]byte{
		buildKF1DBezier(0.0, 1.0, 0.0, 0.333, 0.0, 0.333),
		buildKF1DBezier(1.0, 0.5, 0.0, 0.5, 0.0, 0.5),
	})
	scaleVal := [3]float64{1, 1, 1}
	scaleSpd := [3]float64{0, 0, 0}
	scaleInf := [3]float64{0.1, 0.2, 0.3}
	defInf := [3]float64{0.333, 0.333, 0.333}
	sc := rb.leafKeyframed("ADBE Scale", 0x03, 128, [][]byte{
		buildKF3DBezierNonSpatial(0.0, scaleVal, scaleSpd, scaleInf, scaleSpd, defInf),
		buildKF3DBezierNonSpatial(2.0, scaleVal, scaleSpd, defInf, scaleSpd, defInf),
	})

	var tdgpBody []byte
	tdgpBody = append(tdgpBody, op...)
	tdgpBody = append(tdgpBody, sc...)
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
	data := wrapAsLayer(tdgpBody)

	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	layer := proj.Compositions[0].Layers[0]

	opacity := layer.Opacity()
	if opacity == nil || len(opacity.Keyframes) != 2 {
		t.Fatalf("opacity keyframes missing")
	}
	if opacity.Keyframes[0].InInterp != aep.InterpBezier {
		t.Errorf("opacity kf[0].InInterp = %s, want bezier", opacity.Keyframes[0].InInterp)
	}
	if len(opacity.Keyframes[0].InTemporalEase) != 1 {
		t.Fatalf("1D ease should be length 1, got %d", len(opacity.Keyframes[0].InTemporalEase))
	}
	if opacity.Keyframes[0].InTemporalEase[0].Influence != 0.333 {
		t.Errorf("opacity kf[0] inInf = %v, want 0.333", opacity.Keyframes[0].InTemporalEase[0].Influence)
	}
	if opacity.Keyframes[1].OutTemporalEase[0].Influence != 0.5 {
		t.Errorf("opacity kf[1] outInf = %v, want 0.5", opacity.Keyframes[1].OutTemporalEase[0].Influence)
	}

	scale := layer.Scale()
	if scale == nil || len(scale.Keyframes) != 2 {
		t.Fatalf("scale keyframes missing")
	}
	in := scale.Keyframes[0].InTemporalEase
	if len(in) != 3 {
		t.Fatalf("3D ease should be length 3, got %d", len(in))
	}
	want := []float64{0.1, 0.2, 0.3}
	for i, w := range want {
		if in[i].Influence != w {
			t.Errorf("scale kf[0] inInf[%d] = %v, want %v", i, in[i].Influence, w)
		}
	}
}

// TestPropertyInsertAndDeleteKeyframe round-trips InsertKeyframe and
// DeleteKeyframe on both 1D (Opacity) and 3D spatial (Position)
// properties, then re-parses to confirm the lhd3 count and ldat
// stream stay coherent across WriteAEP.
func TestPropertyInsertAndDeleteKeyframe(t *testing.T) {
	proj, err := aep.FromReader(bytes.NewReader(buildKeyframedAEP()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	layer := proj.Compositions[0].Layers[0]

	opa := layer.Opacity()
	if opa == nil {
		t.Fatal("Opacity missing")
	}
	originalCount := len(opa.Keyframes)

	kf, idx, err := opa.InsertKeyframe(1.5, 0.75)
	if err != nil {
		t.Fatalf("InsertKeyframe Opacity: %v", err)
	}
	if len(opa.Keyframes) != originalCount+1 {
		t.Errorf("opacity Keyframes len = %d, want %d", len(opa.Keyframes), originalCount+1)
	}
	if kf == nil || math.Abs(kf.Time-1.5) > 1e-3 {
		t.Errorf("inserted kf Time = %g", kf.Time)
	}
	if got, _ := kf.Value.(float64); got != 0.75 {
		t.Errorf("inserted kf Value = %v", kf.Value)
	}
	if idx <= 0 || idx >= len(opa.Keyframes) {
		t.Errorf("inserted idx = %d, expected interior", idx)
	}

	pos := layer.Position()
	if pos == nil {
		t.Fatal("Position missing")
	}
	posOrig := len(pos.Keyframes)
	_, posIdx, err := pos.InsertKeyframe(0.5, []float64{500, 250, 0})
	if err != nil {
		t.Fatalf("InsertKeyframe Position: %v", err)
	}
	if len(pos.Keyframes) != posOrig+1 {
		t.Errorf("position Keyframes len = %d, want %d", len(pos.Keyframes), posOrig+1)
	}
	got := pos.Keyframes[posIdx].Value.([]float64)
	if got[0] != 500 || got[1] != 250 {
		t.Errorf("position inserted Value = %v", got)
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	rt := proj2.Compositions[0].Layers[0]
	if len(rt.Opacity().Keyframes) != originalCount+1 {
		t.Errorf("roundtrip opacity Keyframes = %d", len(rt.Opacity().Keyframes))
	}
	if len(rt.Position().Keyframes) != posOrig+1 {
		t.Errorf("roundtrip position Keyframes = %d", len(rt.Position().Keyframes))
	}

	if err := rt.Opacity().DeleteKeyframe(idx); err != nil {
		t.Fatalf("DeleteKeyframe: %v", err)
	}
	if len(rt.Opacity().Keyframes) != originalCount {
		t.Errorf("after DeleteKeyframe: opacity len = %d, want %d", len(rt.Opacity().Keyframes), originalCount)
	}

	var buf2 bytes.Buffer
	if err := proj2.WriteAEP(&buf2); err != nil {
		t.Fatalf("WriteAEP post-delete: %v", err)
	}
	proj3, err := aep.FromReader(bytes.NewReader(buf2.Bytes()))
	if err != nil {
		t.Fatalf("re-parse post-delete: %v", err)
	}
	if len(proj3.Compositions[0].Layers[0].Opacity().Keyframes) != originalCount {
		t.Errorf("after roundtrip-delete: opacity len = %d", len(proj3.Compositions[0].Layers[0].Opacity().Keyframes))
	}
}

func TestPropertyInsertKeyframeRejectsEmpty(t *testing.T) {
	p := &aep.Property{MatchName: "stub", Components: 1}
	_, _, err := p.InsertKeyframe(0, 1.0)
	if err == nil {
		t.Error("InsertKeyframe on property with no keyframes: expected error")
	}
}

// TestKeyframeInterpAndEaseSetters covers SetInInterp / SetOutInterp /
// SetInTemporalEase / SetOutTemporalEase on both spatial (Position 3D)
// and non-spatial (Opacity 1D) keyframes, plus SetInSpatialTangent /
// SetOutSpatialTangent on the spatial case. Round-trips via WriteAEP.
func TestKeyframeInterpAndEaseSetters(t *testing.T) {
	proj, err := aep.FromReader(bytes.NewReader(buildKeyframedAEP()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	layer := proj.Compositions[0].Layers[0]

	opa := layer.Opacity()
	if opa == nil || len(opa.Keyframes) == 0 {
		t.Fatal("Opacity keyframes missing")
	}
	okf := opa.Keyframes[0]
	mustNoErr := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	mustNoErr("opacity SetInInterp", okf.SetInInterp(aep.InterpHold))
	mustNoErr("opacity SetOutInterp", okf.SetOutInterp(aep.InterpBezier))
	mustNoErr("opacity SetInTemporalEase", okf.SetInTemporalEase([]aep.TemporalEase{{Speed: 1.5, Influence: 0.25}}))
	mustNoErr("opacity SetOutTemporalEase", okf.SetOutTemporalEase([]aep.TemporalEase{{Speed: 2.5, Influence: 0.75}}))

	if okf.InInterp != aep.InterpHold || okf.OutInterp != aep.InterpBezier {
		t.Errorf("opacity in-mem interp: in=%s out=%s", okf.InInterp, okf.OutInterp)
	}
	if len(okf.InTemporalEase) != 1 || okf.InTemporalEase[0].Speed != 1.5 || okf.InTemporalEase[0].Influence != 0.25 {
		t.Errorf("opacity in-mem InTemporalEase = %v", okf.InTemporalEase)
	}
	if len(okf.OutTemporalEase) != 1 || okf.OutTemporalEase[0].Speed != 2.5 {
		t.Errorf("opacity in-mem OutTemporalEase = %v", okf.OutTemporalEase)
	}

	pos := layer.Position()
	if pos == nil || len(pos.Keyframes) == 0 {
		t.Fatal("Position keyframes missing")
	}
	pkf := pos.Keyframes[0]
	mustNoErr("position SetInInterp", pkf.SetInInterp(aep.InterpBezier))
	mustNoErr("position SetOutInterp", pkf.SetOutInterp(aep.InterpBezier))
	mustNoErr("position SetInTemporalEase", pkf.SetInTemporalEase([]aep.TemporalEase{{Speed: 3.0, Influence: 0.4}}))
	mustNoErr("position SetOutTemporalEase", pkf.SetOutTemporalEase([]aep.TemporalEase{{Speed: 4.0, Influence: 0.6}}))
	mustNoErr("position SetInSpatialTangent", pkf.SetInSpatialTangent([]float64{10, 20, 0}))
	mustNoErr("position SetOutSpatialTangent", pkf.SetOutSpatialTangent([]float64{-10, -20, 0}))

	if pkf.InTemporalEase[0].Speed != 3.0 || pkf.OutTemporalEase[0].Influence != 0.6 {
		t.Errorf("position in-mem ease: in=%v out=%v", pkf.InTemporalEase, pkf.OutTemporalEase)
	}
	if pkf.InSpatialTangent[0] != 10 || pkf.OutSpatialTangent[1] != -20 {
		t.Errorf("position in-mem tangents: in=%v out=%v", pkf.InSpatialTangent, pkf.OutSpatialTangent)
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	rt := proj2.Compositions[0].Layers[0]
	rtOpa := rt.Opacity().Keyframes[0]
	if rtOpa.InInterp != aep.InterpHold || rtOpa.OutInterp != aep.InterpBezier {
		t.Errorf("roundtrip opacity interp: in=%s out=%s", rtOpa.InInterp, rtOpa.OutInterp)
	}
	if rtOpa.InTemporalEase[0].Speed != 1.5 || rtOpa.OutTemporalEase[0].Influence != 0.75 {
		t.Errorf("roundtrip opacity eases: in=%v out=%v", rtOpa.InTemporalEase, rtOpa.OutTemporalEase)
	}
	rtPos := rt.Position().Keyframes[0]
	if rtPos.InSpatialTangent[0] != 10 || rtPos.OutSpatialTangent[1] != -20 {
		t.Errorf("roundtrip position tangents: in=%v out=%v", rtPos.InSpatialTangent, rtPos.OutSpatialTangent)
	}
}
