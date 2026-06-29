package aep_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestTrackMatteLayerReal(t *testing.T) {
	proj, comp := openTrackMatteAE24(t)

	// All 4 source / baseline layers have no track matte set
	for _, name := range []string{"solidA", "solidB", "solidC", "mt_baseline"} {
		l := layerBySourceName(proj, comp, name)
		if l == nil {
			t.Errorf("layer with source=%q missing", name)
			continue
		}
		if l.TrackMatteLayerID != 0 {
			t.Errorf("%s TrackMatteLayerID = %d, want 0 (no explicit matte)", name, l.TrackMatteLayerID)
		}
		if l.TrackMatte != aep.TrackMatteNone {
			t.Errorf("%s TrackMatte = %v, want None", name, l.TrackMatte)
		}
		if l.TrackMatteLayer() != nil {
			t.Errorf("%s TrackMatteLayer() = %v, want nil", name, l.TrackMatteLayer())
		}
	}

	type matteCase struct {
		layerSource  string // identify dependent layer by its source name
		wantMode     aep.TrackMatteType
		wantMatteSrc string // source footage name of the matte-source layer
	}
	cases := []matteCase{
		{"mt_alpha_to_solidA", aep.TrackMatteAlpha, "solidA"},
		{"mt_luma_to_solidB", aep.TrackMatteLuma, "solidB"},
		{"mt_alphainv_to_solidC", aep.TrackMatteAlphaInverse, "solidC"},
	}
	for _, c := range cases {
		l := layerBySourceName(proj, comp, c.layerSource)
		if l == nil {
			t.Errorf("layer with source=%q missing", c.layerSource)
			continue
		}
		if l.TrackMatte != c.wantMode {
			t.Errorf("%s TrackMatte = %v, want %v", c.layerSource, l.TrackMatte, c.wantMode)
		}
		if l.TrackMatteLayerID == 0 {
			t.Errorf("%s TrackMatteLayerID = 0, want non-zero", c.layerSource)
			continue
		}
		src := l.TrackMatteLayer()
		if src == nil {
			t.Errorf("%s TrackMatteLayer() = nil, want layer with source=%q", c.layerSource, c.wantMatteSrc)
			continue
		}
		wantSrc := layerBySourceName(proj, comp, c.wantMatteSrc)
		if wantSrc == nil || src.ID != wantSrc.ID {
			t.Errorf("%s TrackMatteLayer().ID = %d, want layer with source=%q (id=%d)",
				c.layerSource, src.ID, c.wantMatteSrc,
				func() uint32 {
					if wantSrc != nil {
						return wantSrc.ID
					}
					return 0
				}())
		}
	}
}

func TestSetTrackMatteLayerRoundtrip(t *testing.T) {
	proj, comp := openTrackMatteAE24(t)
	mtBaseline := layerBySourceName(proj, comp, "mt_baseline")
	solidA := layerBySourceName(proj, comp, "solidA")
	if mtBaseline == nil || solidA == nil {
		t.Fatalf("fixture layers missing (mtBaseline=%v solidA=%v)", mtBaseline, solidA)
	}
	if mtBaseline.TrackMatteLayerID != 0 {
		t.Fatalf("mt_baseline should start with no matte; got TMLayerID=%d", mtBaseline.TrackMatteLayerID)
	}

	// Assign solidA as matte source for mt_baseline with Luma mode.
	if err := mtBaseline.SetTrackMatteLayer(solidA.ID, aep.TrackMatteLuma); err != nil {
		t.Fatalf("SetTrackMatteLayer: %v", err)
	}
	if mtBaseline.TrackMatteLayerID != solidA.ID {
		t.Errorf("after Set: TMLayerID = %d, want %d", mtBaseline.TrackMatteLayerID, solidA.ID)
	}
	if mtBaseline.TrackMatte != aep.TrackMatteLuma {
		t.Errorf("after Set: TrackMatte = %v, want Luma", mtBaseline.TrackMatte)
	}
	if got := mtBaseline.TrackMatteLayer(); got == nil || got.ID != solidA.ID {
		t.Errorf("after Set: TrackMatteLayer() = %v, want solidA", got)
	}

	// Roundtrip.
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
	mt2 := layerBySourceName(proj2, comp2, "mt_baseline")
	solidA2 := layerBySourceName(proj2, comp2, "solidA")
	if mt2 == nil || solidA2 == nil {
		t.Fatalf("roundtrip layers missing")
	}
	if mt2.TrackMatteLayerID != solidA2.ID {
		t.Errorf("roundtrip TMLayerID = %d, want %d", mt2.TrackMatteLayerID, solidA2.ID)
	}
	if mt2.TrackMatte != aep.TrackMatteLuma {
		t.Errorf("roundtrip TrackMatte = %v, want Luma", mt2.TrackMatte)
	}

	// Clear and roundtrip.
	if err := mt2.ClearTrackMatteLayer(); err != nil {
		t.Fatalf("ClearTrackMatteLayer: %v", err)
	}
	if mt2.TrackMatteLayerID != 0 || mt2.TrackMatte != aep.TrackMatteNone {
		t.Errorf("after Clear: TMLayerID=%d TrackMatte=%v", mt2.TrackMatteLayerID, mt2.TrackMatte)
	}
	var buf2 bytes.Buffer
	if err := proj2.WriteAEP(&buf2); err != nil {
		t.Fatalf("WriteAEP after clear: %v", err)
	}
	proj3, err := aep.FromReader(bytes.NewReader(buf2.Bytes()))
	if err != nil {
		t.Fatalf("re-parse after clear: %v", err)
	}
	var comp3 *aep.Composition
	for _, c := range proj3.Compositions {
		if c.Name == "RE_TRACKMATTE" {
			comp3 = c
			break
		}
	}
	mt3 := layerBySourceName(proj3, comp3, "mt_baseline")
	if mt3.TrackMatteLayerID != 0 || mt3.TrackMatte != aep.TrackMatteNone {
		t.Errorf("after clear+roundtrip: TMLayerID=%d TrackMatte=%v", mt3.TrackMatteLayerID, mt3.TrackMatte)
	}
}

func TestSetTrackMatteLayerRejectsInvalid(t *testing.T) {
	proj, comp := openTrackMatteAE24(t)
	mtBaseline := layerBySourceName(proj, comp, "mt_baseline")
	if mtBaseline == nil {
		t.Fatalf("mt_baseline missing")
	}
	// Self-matte rejected.
	if err := mtBaseline.SetTrackMatteLayer(mtBaseline.ID, aep.TrackMatteAlpha); err == nil {
		t.Error("self-matte: expected error")
	}
	// Unknown ID rejected.
	if err := mtBaseline.SetTrackMatteLayer(99999, aep.TrackMatteAlpha); err == nil {
		t.Error("unknown ID: expected error")
	}
	// Non-text layer without ldta returns error.
	stub := &aep.Layer{Name: "stub"}
	if err := stub.SetTrackMatteLayer(1, aep.TrackMatteAlpha); err == nil {
		t.Error("layer without ldta: expected error")
	}
}

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
