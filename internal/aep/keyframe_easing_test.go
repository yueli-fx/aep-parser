package aep_test

import (
	"bytes"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

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
