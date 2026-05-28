package aep_test

import (
	"bytes"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// Tests for Layer.SetTrackMatteSource — the *Layer-arg parity wrapper
// over SetTrackMatteLayer. Underlying setter has its own round-trip and
// refuse-invalid tests in layer_test.go; these tests verify wrapper-
// specific behavior (nil / cross-comp / self) + parity guarantee.

func TestSetTrackMatteSource_Alpha(t *testing.T) {
	setTrackMatteSourceHappy(t, aep.TrackMatteAlpha)
}

func TestSetTrackMatteSource_AlphaInverse(t *testing.T) {
	setTrackMatteSourceHappy(t, aep.TrackMatteAlphaInverse)
}

func TestSetTrackMatteSource_Luma(t *testing.T) {
	setTrackMatteSourceHappy(t, aep.TrackMatteLuma)
}

func TestSetTrackMatteSource_LumaInverse(t *testing.T) {
	setTrackMatteSourceHappy(t, aep.TrackMatteLumaInverse)
}

func setTrackMatteSourceHappy(t *testing.T, mode aep.TrackMatteType) {
	t.Helper()
	proj, comp := openTrackMatteAE24(t)
	target := layerBySourceName(proj, comp, "mt_baseline")
	src := layerBySourceName(proj, comp, "solidA")
	if target == nil || src == nil {
		t.Fatalf("fixture layers missing (target=%v src=%v)", target, src)
	}
	if err := target.SetTrackMatteSource(src, mode); err != nil {
		t.Fatalf("SetTrackMatteSource(%v): %v", mode, err)
	}
	if target.TrackMatteLayerID != src.ID {
		t.Errorf("TrackMatteLayerID = %d, want %d", target.TrackMatteLayerID, src.ID)
	}
	if target.TrackMatte != mode {
		t.Errorf("TrackMatte = %v, want %v", target.TrackMatte, mode)
	}
	if got := target.TrackMatteLayer(); got == nil || got.ID != src.ID {
		t.Errorf("TrackMatteLayer() resolve = %v, want solidA (id=%d)", got, src.ID)
	}
}

func TestSetTrackMatteSource_IntentWithoutMode(t *testing.T) {
	proj, comp := openTrackMatteAE24(t)
	target := layerBySourceName(proj, comp, "mt_baseline")
	src := layerBySourceName(proj, comp, "solidA")
	if err := target.SetTrackMatteSource(src, aep.TrackMatteNone); err != nil {
		t.Fatalf("SetTrackMatteSource(None): %v", err)
	}
	if target.TrackMatteLayerID != src.ID {
		t.Errorf("TrackMatteLayerID = %d, want %d (source pointer must persist when mode=None)", target.TrackMatteLayerID, src.ID)
	}
	if target.TrackMatte != aep.TrackMatteNone {
		t.Errorf("TrackMatte = %v, want None", target.TrackMatte)
	}
}

func TestSetTrackMatteSource_RefuseNilSource(t *testing.T) {
	proj, comp := openTrackMatteAE24(t)
	target := layerBySourceName(proj, comp, "mt_baseline")
	err := target.SetTrackMatteSource(nil, aep.TrackMatteAlpha)
	if err == nil {
		t.Fatal("nil source: want error, got nil")
	}
	if !strings.Contains(err.Error(), "nil") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSetTrackMatteSource_RefuseSelf(t *testing.T) {
	proj, comp := openTrackMatteAE24(t)
	target := layerBySourceName(proj, comp, "mt_baseline")
	err := target.SetTrackMatteSource(target, aep.TrackMatteAlpha)
	if err == nil {
		t.Fatal("self-matte: want error, got nil")
	}
	if !strings.Contains(err.Error(), "self-matte") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSetTrackMatteSource_RefuseCrossComp(t *testing.T) {
	projA, compA := openTrackMatteAE24(t)
	projB, compB := openTrackMatteAE24(t)
	target := layerBySourceName(projA, compA, "mt_baseline")
	src := layerBySourceName(projB, compB, "solidA")
	if target == nil || src == nil {
		t.Fatalf("fixture layers missing (target=%v src=%v)", target, src)
	}
	err := target.SetTrackMatteSource(src, aep.TrackMatteAlpha)
	if err == nil {
		t.Fatal("cross-comp: want error, got nil")
	}
	if !strings.Contains(err.Error(), "cross-comp") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSetTrackMatteSource_RefuseStub(t *testing.T) {
	stub := &aep.Layer{Name: "stub"}
	proj, comp := openTrackMatteAE24(t)
	src := layerBySourceName(proj, comp, "solidA")
	if err := stub.SetTrackMatteSource(src, aep.TrackMatteAlpha); err == nil {
		t.Error("stub layer (no comp): want error, got nil")
	}
}

func TestSetTrackMatteSource_RoundTrip(t *testing.T) {
	proj, comp := openTrackMatteAE24(t)
	target := layerBySourceName(proj, comp, "mt_baseline")
	src := layerBySourceName(proj, comp, "solidA")
	if err := target.SetTrackMatteSource(src, aep.TrackMatteLuma); err != nil {
		t.Fatalf("SetTrackMatteSource: %v", err)
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	var comp2 *aep.Composition
	for _, c := range proj2.Compositions {
		if c.Name == "RE_TRACKMATTE" {
			comp2 = c
			break
		}
	}
	target2 := layerBySourceName(proj2, comp2, "mt_baseline")
	src2 := layerBySourceName(proj2, comp2, "solidA")
	if target2.TrackMatteLayerID != src2.ID {
		t.Errorf("roundtrip TrackMatteLayerID = %d, want %d", target2.TrackMatteLayerID, src2.ID)
	}
	if target2.TrackMatte != aep.TrackMatteLuma {
		t.Errorf("roundtrip TrackMatte = %v, want Luma", target2.TrackMatte)
	}
}

func TestSetTrackMatteSource_ParityWithSetTrackMatteLayer(t *testing.T) {
	projWrap, compWrap := openTrackMatteAE24(t)
	projID, compID := openTrackMatteAE24(t)

	tWrap := layerBySourceName(projWrap, compWrap, "mt_baseline")
	sWrap := layerBySourceName(projWrap, compWrap, "solidA")
	tID := layerBySourceName(projID, compID, "mt_baseline")
	sID := layerBySourceName(projID, compID, "solidA")

	if err := tWrap.SetTrackMatteSource(sWrap, aep.TrackMatteAlphaInverse); err != nil {
		t.Fatalf("wrapper: %v", err)
	}
	if err := tID.SetTrackMatteLayer(sID.ID, aep.TrackMatteAlphaInverse); err != nil {
		t.Fatalf("id setter: %v", err)
	}

	var bufWrap, bufID bytes.Buffer
	if err := projWrap.WriteAEP(&bufWrap); err != nil {
		t.Fatalf("WriteAEP wrap: %v", err)
	}
	if err := projID.WriteAEP(&bufID); err != nil {
		t.Fatalf("WriteAEP id: %v", err)
	}
	if !bytes.Equal(bufWrap.Bytes(), bufID.Bytes()) {
		t.Errorf("wrapper output (%d bytes) != id-setter output (%d bytes); parity broken",
			bufWrap.Len(), bufID.Len())
	}
}
