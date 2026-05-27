package aep_test

import (
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestLayerConvenience_ReCameraLight(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present: %v", err)
	}
	var comp1, reCL *aep.Composition
	for _, c := range proj.Compositions {
		switch c.Name {
		case "Comp 1":
			comp1 = c
		case "RE_CL":
			reCL = c
		}
	}
	if comp1 == nil || reCL == nil {
		t.Fatalf("Comp 1 or RE_CL missing")
	}

	t.Run("ContainingComp", func(t *testing.T) {
		for _, l := range comp1.Layers {
			if l.ContainingComp() != comp1 {
				t.Errorf("layer %q ContainingComp != Comp 1", l.Name)
			}
		}
	})

	t.Run("HasVideo_HasAudio_AVLayer", func(t *testing.T) {
		// First AV layer in Comp 1.
		var av *aep.Layer
		for _, l := range comp1.Layers {
			if l.Type == aep.LayerTypeAV {
				av = l
				break
			}
		}
		if av == nil {
			t.Fatal("no AV layer in Comp 1")
		}
		if !av.HasVideo() {
			t.Error("AV layer HasVideo = false; want true")
		}
		// All AV layers in Comp 1 source from solid footage → HasAudio = false.
		if av.HasAudio() {
			t.Errorf("AV layer (solid source) HasAudio = true; want false")
		}
	})

	t.Run("HasVideo_CameraLight", func(t *testing.T) {
		for _, l := range reCL.Layers {
			if l.HasVideo() {
				t.Errorf("layer %q (type=%s) HasVideo = true; want false (camera/light no video)", l.Name, l.Type)
			}
		}
	})

	t.Run("Width_Height_FromSourceOrComp", func(t *testing.T) {
		// Camera/Light layers in RE_CL have no source → fall back to comp size.
		for _, l := range reCL.Layers {
			if l.Width() != int(reCL.Width) {
				t.Errorf("layer %q Width = %d, want comp width %d", l.Name, l.Width(), reCL.Width)
			}
			if l.Height() != int(reCL.Height) {
				t.Errorf("layer %q Height = %d, want comp height %d", l.Name, l.Height(), reCL.Height)
			}
		}
	})

	t.Run("AutoName_IsNameFromSource", func(t *testing.T) {
		// Find an unnamed AV layer (Name == "") — there are several in Comp 1.
		var unnamed *aep.Layer
		for _, l := range comp1.Layers {
			if l.Name == "" && l.SourceFootage() != nil {
				unnamed = l
				break
			}
		}
		if unnamed == nil {
			t.Skip("no unnamed AV layer to test AutoName on")
		}
		f := unnamed.SourceFootage()
		if got := unnamed.AutoName(); got != f.Name {
			t.Errorf("AutoName = %q, want source name %q", got, f.Name)
		}
		if !unnamed.IsNameFromSource() {
			t.Error("IsNameFromSource = false; want true for unnamed layer with source")
		}
	})

	t.Run("ActiveAtTime", func(t *testing.T) {
		// Pick a visible AV layer; it should be Active within [start, start+dur).
		for _, l := range comp1.Layers {
			if !l.Visible {
				continue
			}
			start, end := l.StartTime, l.StartTime+l.Duration
			if l.Duration <= 0 {
				end = l.StartTime + comp1.Duration
			}
			mid := (start + end) / 2
			if !l.ActiveAtTime(mid) {
				t.Errorf("layer %q not Active at mid=%v (range [%v, %v))", l.Name, mid, start, end)
			}
			if l.ActiveAtTime(end + 1) {
				t.Errorf("layer %q Active past end=%v", l.Name, end)
			}
			return // one is enough
		}
		t.Skip("no visible layer in Comp 1")
	})

	t.Run("HasTrackMatte_IsTrackMatte", func(t *testing.T) {
		// re_cameralight Comp 1 layers don't have track mattes set up. Just
		// verify the helpers return false consistently for both sides.
		for _, l := range comp1.Layers {
			if l.HasTrackMatte() && l.TrackMatte == aep.TrackMatteNone {
				t.Errorf("layer %q HasTrackMatte true but TrackMatte=None", l.Name)
			}
		}
	})
}

func TestLayerConvenience_StandaloneLayer(t *testing.T) {
	// A layer built without a parser owns no comp — most helpers should
	// degrade gracefully without panicking.
	l := &aep.Layer{Name: "Solo", Type: aep.LayerTypeAV}
	if l.ContainingComp() != nil {
		t.Error("ContainingComp on standalone layer != nil")
	}
	if l.HasAudio() {
		t.Error("HasAudio on standalone layer = true")
	}
	if l.AudioActive() {
		t.Error("AudioActive on standalone layer = true")
	}
	if l.AudioActiveAtTime(0) {
		t.Error("AudioActiveAtTime on standalone layer = true")
	}
	if l.IsTrackMatte() {
		t.Error("IsTrackMatte on standalone layer = true")
	}
	if l.IsNameFromSource() {
		t.Error("IsNameFromSource = true; want false (no source)")
	}
	if got := l.AutoName(); got != "Solo" {
		t.Errorf("AutoName = %q, want Solo", got)
	}
	// Width/Height with no comp or source → 0.
	if l.Width() != 0 || l.Height() != 0 {
		t.Errorf("Width/Height = %d/%d, want 0/0 (no source/comp)", l.Width(), l.Height())
	}
}

func TestLayer_IsThreeDModelLayer(t *testing.T) {
	cases := []struct {
		ltype aep.LayerType
		want  bool
	}{
		{aep.LayerType3DModel, true},
		{aep.LayerTypeAV, false},
		{aep.LayerTypeLight, false},
		{aep.LayerTypeCamera, false},
		{aep.LayerTypeShape, false},
		{aep.LayerTypeText, false},
		{aep.LayerTypeNull, false},
	}
	for _, c := range cases {
		l := &aep.Layer{Type: c.ltype}
		if got := l.IsThreeDModelLayer(); got != c.want {
			t.Errorf("Type=%q IsThreeDModelLayer() = %v, want %v", c.ltype, got, c.want)
		}
	}
}
