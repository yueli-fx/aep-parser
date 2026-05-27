package aep

import (
	"encoding/binary"
	"strings"
	"testing"

	"github.com/example/aep-parser/internal/rifx"
)

// newSyntheticLayer builds a minimal Layer with a 0xA4-byte ldta chunk
// (covers all offsets the writers touch) and the light-source sentinel
// pre-written into @0x28. Used by light-source / replace-source tests.
func newSyntheticLayer(typ LayerType, name string, id uint32) *Layer {
	chunk := &rifx.Chunk{ID: rifx.IDLdta, Data: make([]byte, 0xA4)}
	binary.BigEndian.PutUint32(chunk.Data[0x28:0x2C], lightSourceUndefined)
	return &Layer{
		Name:     name,
		Type:     typ,
		ID:       id,
		ldta:     chunk,
		SourceID: lightSourceUndefined,
	}
}

func TestLightSource_RoundtripAndClear(t *testing.T) {
	comp := &Composition{}
	light := newSyntheticLayer(LayerTypeLight, "L1", 10)
	av := newSyntheticLayer(LayerTypeAV, "AV1", 20)
	light.comp = comp
	av.comp = comp
	comp.Layers = []*Layer{light, av}

	// Initially: sentinel 0xFFFFFFFF → no source.
	if got := light.LightSource(); got != nil {
		t.Errorf("initial LightSource = %v, want nil", got)
	}

	// Set → av.
	if err := light.SetLightSource(av); err != nil {
		t.Fatalf("SetLightSource(av): %v", err)
	}
	if got := light.LightSource(); got != av {
		t.Errorf("after Set, LightSource = %v, want %p (av)", got, av)
	}
	if light.SourceID != 20 {
		t.Errorf("after Set, SourceID = %d, want 20", light.SourceID)
	}
	if got := binary.BigEndian.Uint32(light.ldta.Data[0x28:0x2C]); got != 20 {
		t.Errorf("after Set, ldta@0x28 = %d, want 20", got)
	}

	// Clear.
	if err := light.SetLightSource(nil); err != nil {
		t.Fatalf("SetLightSource(nil): %v", err)
	}
	if got := light.LightSource(); got != nil {
		t.Errorf("after Clear, LightSource = %v, want nil", got)
	}
	if light.SourceID != lightSourceUndefined {
		t.Errorf("after Clear, SourceID = %#x, want %#x", light.SourceID, lightSourceUndefined)
	}
}

func TestLightSource_Validation(t *testing.T) {
	comp := &Composition{}
	otherComp := &Composition{}
	light := newSyntheticLayer(LayerTypeLight, "L1", 10)
	av := newSyntheticLayer(LayerTypeAV, "AV1", 20)
	cam := newSyntheticLayer(LayerTypeCamera, "C1", 30)
	light2 := newSyntheticLayer(LayerTypeLight, "L2", 40)
	av3d := newSyntheticLayer(LayerTypeAV, "AV3D", 50)
	av3d.Is3D = true
	avOther := newSyntheticLayer(LayerTypeAV, "AVOther", 60)
	light.comp, av.comp, cam.comp, light2.comp, av3d.comp = comp, comp, comp, comp, comp
	avOther.comp = otherComp
	comp.Layers = []*Layer{light, av, cam, light2, av3d}
	otherComp.Layers = []*Layer{avOther}

	cases := []struct {
		name    string
		caller  *Layer
		target  *Layer
		wantErr string // substring
	}{
		{"non-light caller", av, light, "only valid for Light"},
		{"target self", light, light, "self-source"},
		{"target camera", light, cam, "cannot be Light/Camera"},
		{"target light", light, light2, "cannot be Light/Camera"},
		{"target 3d", light, av3d, "cannot be a 3D layer"},
		{"target cross-comp", light, avOther, "same composition"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.caller.SetLightSource(c.target)
			if err == nil {
				t.Fatalf("err = nil, want substring %q", c.wantErr)
			}
			if !strings.Contains(err.Error(), c.wantErr) {
				t.Errorf("err = %q, want substring %q", err.Error(), c.wantErr)
			}
		})
	}
}

func TestLightSource_NonLightReadReturnsNil(t *testing.T) {
	comp := &Composition{}
	av := newSyntheticLayer(LayerTypeAV, "AV1", 20)
	av.comp = comp
	av.SourceID = 99 // would normally be a footage ID
	comp.Layers = []*Layer{av}
	if got := av.LightSource(); got != nil {
		t.Errorf("LightSource on non-light = %v, want nil", got)
	}
}

func TestReplaceSource_RoundtripAndWarning(t *testing.T) {
	proj := &Project{}
	comp := &Composition{proj: proj}
	proj.Compositions = []*Composition{comp}

	// Create synthetic footage items with IDs
	footage1 := &Footage{ID: 100, Name: "Footage1"}
	footage2 := &Footage{ID: 200, Name: "Footage2"}
	proj.Footage = []*Footage{footage1, footage2}

	// Create a layer with initial source
	layer := newSyntheticLayer(LayerTypeAV, "AV1", 1)
	layer.comp = comp
	layer.SourceID = 100 // initially points to footage1
	comp.Layers = []*Layer{layer}

	// Test ReplaceSource with footage2
	if err := layer.ReplaceSource(footage2, false); err != nil {
		t.Fatalf("ReplaceSource(footage2, false): %v", err)
	}
	if layer.SourceID != 200 {
		t.Errorf("after ReplaceSource, SourceID = %d, want 200", layer.SourceID)
	}
	if got := binary.BigEndian.Uint32(layer.ldta.Data[0x28:0x2C]); got != 200 {
		t.Errorf("after ReplaceSource, ldta@0x28 = %d, want 200", got)
	}

	// Test ReplaceSource with fixExpressions=true (should add warning)
	proj.Warnings = nil
	if err := layer.ReplaceSource(footage1, true); err != nil {
		t.Fatalf("ReplaceSource(footage1, true): %v", err)
	}
	if len(proj.Warnings) != 1 {
		t.Errorf("expected 1 warning, got %d", len(proj.Warnings))
	}
	if !strings.Contains(proj.Warnings[0], "fixExpressions=true not implemented") {
		t.Errorf("warning = %q, want substring 'fixExpressions=true not implemented'", proj.Warnings[0])
	}

	// Test ReplaceSource with nil target
	err := layer.ReplaceSource(nil, false)
	if err == nil {
		t.Error("ReplaceSource(nil) should return error")
	}
	if !strings.Contains(err.Error(), "target is nil") {
		t.Errorf("err = %q, want substring 'target is nil'", err.Error())
	}
}
