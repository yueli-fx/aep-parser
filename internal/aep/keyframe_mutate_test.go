package aep_test

import (
	"bytes"
	"math"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

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
