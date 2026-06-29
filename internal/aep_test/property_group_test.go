package aep_test

import (
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// TestPropertyTree_CameralightFixture verifies the hierarchical
// PropertyGroup tree built alongside flat Layer.Properties for a known
// fixture with Camera + Light + Solid layers.
func TestPropertyTree_CameralightFixture(t *testing.T) {
	proj, err := aep.Open("../../test_data/fixtures/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present: %v", err)
	}
	if len(proj.Compositions) == 0 {
		t.Fatal("fixture has no compositions")
	}
	var camera, light *aep.Layer
	for _, c := range proj.Compositions {
		for _, l := range c.Layers {
			switch l.Type {
			case aep.LayerTypeCamera:
				if camera == nil {
					camera = l
				}
			case aep.LayerTypeLight:
				if light == nil {
					light = l
				}
			}
		}
	}
	if camera == nil {
		t.Skip("no camera layer in fixture")
	}

	tree := camera.PropertyTree()
	if tree == nil {
		t.Fatal("Camera.PropertyTree() = nil; want non-nil")
	}
	if tree.NumProperties() == 0 {
		t.Fatal("Camera root group has 0 children")
	}

	// Every layer should expose a Transform group.
	tg := camera.TransformGroup()
	if tg == nil {
		t.Fatal("Camera.TransformGroup() = nil; want present")
	}
	pos := tg.Property(aep.MatchNamePosition)
	if pos == nil {
		t.Fatalf("Transform.Property(Position) = nil; group children = %d", tg.NumProperties())
	}
	if pos.MatchName != aep.MatchNamePosition {
		t.Errorf("Position.MatchName = %q, want %q", pos.MatchName, aep.MatchNamePosition)
	}

	// PropertyByPath: chained access from layer.
	pos2 := camera.PropertyByPath(aep.MatchNameGroupTransform, aep.MatchNamePosition)
	if pos2 != pos {
		t.Errorf("PropertyByPath(Transform, Position) returned different *Property than direct walk")
	}

	// CameraOptionsGroup present on camera, not on light.
	co := camera.CameraOptionsGroup()
	if co == nil {
		t.Error("Camera.CameraOptionsGroup() = nil")
	}
	if light != nil {
		if light.CameraOptionsGroup() != nil {
			t.Error("Light.CameraOptionsGroup() != nil; want nil")
		}
		if light.LightOptionsGroup() == nil {
			t.Error("Light.LightOptionsGroup() = nil; want present")
		}
	}

	// ParentGroup walks up from a leaf's containing group to root.
	parent := tg.ParentGroup()
	if parent == nil {
		t.Fatal("Transform.ParentGroup() = nil; want root")
	}
	if parent.ParentGroup() != nil {
		t.Error("root.ParentGroup() != nil; tree should bottom out at synthetic root")
	}
}

// TestPropertyTree_EffectsParade_TreeMirrorsEffectsSlice verifies that
// when Layer.Effects is non-empty, the property tree's
// "ADBE Effect Parade" subgroup exists and exposes one child group per
// effect.
func TestPropertyTree_EffectsParade_TreeMirrorsEffectsSlice(t *testing.T) {
	proj, err := aep.Open("../../test_data/fixtures/re_batch.aep")
	if err != nil {
		t.Skipf("re_batch.aep not present: %v", err)
	}
	var layerWithEffects *aep.Layer
	for _, c := range proj.Compositions {
		for _, l := range c.Layers {
			if len(l.Effects) > 0 {
				layerWithEffects = l
				break
			}
		}
		if layerWithEffects != nil {
			break
		}
	}
	if layerWithEffects == nil {
		t.Skip("no layer with effects in fixture")
	}

	parade := layerWithEffects.EffectsParade()
	if parade == nil {
		t.Fatal("layer with effects has no EffectsParade group")
	}
	if parade.NumProperties() == 0 {
		t.Errorf("EffectsParade has %d children but Layer.Effects has %d", 0, len(layerWithEffects.Effects))
	}

	// First effect's match-name should be the first effect-parade child.
	firstEffectName := layerWithEffects.Effects[0].MatchName
	if g := parade.Group(firstEffectName); g == nil {
		t.Errorf("EffectsParade.Group(%q) = nil; layer.Effects[0] should appear as a subgroup", firstEffectName)
	}
}

// TestPropertyTree_StandaloneLayer verifies a layer built outside the
// parser has no PropertyTree (nil-safe accessors).
func TestPropertyTree_StandaloneLayer(t *testing.T) {
	l := &aep.Layer{Name: "synth"}
	if l.PropertyTree() != nil {
		t.Error("standalone Layer.PropertyTree() != nil")
	}
	if l.TransformGroup() != nil {
		t.Error("standalone Layer.TransformGroup() != nil")
	}
	if p := l.PropertyByPath("ADBE Transform Group", "ADBE Position"); p != nil {
		t.Error("standalone Layer.PropertyByPath() != nil")
	}
}

// TestAEPropertyGroup_PropertyByPath_NotFound verifies graceful nil
// return when path doesn't resolve.
func TestAEPropertyGroup_PropertyByPath_NotFound(t *testing.T) {
	proj, err := aep.Open("../../test_data/fixtures/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present: %v", err)
	}
	if len(proj.Compositions) == 0 || len(proj.Compositions[0].Layers) == 0 {
		t.Skip("fixture has no layers")
	}
	l := proj.Compositions[0].Layers[0]
	if p := l.PropertyByPath("nonexistent"); p != nil {
		t.Errorf("PropertyByPath('nonexistent') = %v; want nil", p.MatchName)
	}
	if p := l.PropertyByPath(aep.MatchNameGroupTransform, "nonexistent"); p != nil {
		t.Errorf("PropertyByPath(Transform, 'nonexistent') = %v; want nil", p.MatchName)
	}
	if p := l.PropertyByPath(); p != nil {
		t.Error("PropertyByPath() with no args should return nil")
	}
}

// TestAEPropertyGroup_ParentGroupBackRef verifies a leaf's ParentGroup()
// resolves back to the group that lists it as a Child.
func TestAEPropertyGroup_ParentGroupBackRef(t *testing.T) {
	proj, err := aep.Open("../../test_data/fixtures/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present: %v", err)
	}
	var pos *aep.Property
	var transformGroup *aep.AEPropertyGroup
	for _, c := range proj.Compositions {
		for _, l := range c.Layers {
			if tg := l.TransformGroup(); tg != nil {
				if p := tg.Property(aep.MatchNamePosition); p != nil {
					pos = p
					transformGroup = tg
					break
				}
			}
		}
		if pos != nil {
			break
		}
	}
	if pos == nil {
		t.Skip("no Position in fixture")
	}
	got := pos.ParentGroup()
	if got != transformGroup {
		t.Errorf("Position.ParentGroup() = %p, want %p (TransformGroup)", got, transformGroup)
	}
	// Depth of TransformGroup should be 1 (one hop from root).
	if d := transformGroup.Depth(); d != 1 {
		t.Errorf("TransformGroup.Depth() = %d, want 1", d)
	}
	if d := transformGroup.ParentGroup().Depth(); d != 0 {
		t.Errorf("root.Depth() = %d, want 0", d)
	}
	// PropertyIndex finds the leaf in its parent.
	if idx := transformGroup.PropertyIndex(pos); idx < 0 {
		t.Errorf("TransformGroup.PropertyIndex(Position) = %d (not found)", idx)
	}
	if idx := transformGroup.PropertyIndex(nil); idx != -1 {
		t.Errorf("PropertyIndex(nil) = %d, want -1", idx)
	}
}

// TestAEPropertyGroup_LeafResolutionByTdbsIdentity verifies the tree
// nodes are the SAME *Property instances as Layer.Properties (pointer
// identity), so mutations through the flat API are visible through the
// tree and vice versa.
func TestAEPropertyGroup_LeafResolutionByTdbsIdentity(t *testing.T) {
	proj, err := aep.Open("../../test_data/fixtures/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present: %v", err)
	}
	var l *aep.Layer
	for _, c := range proj.Compositions {
		for _, lr := range c.Layers {
			if lr.Position() != nil {
				l = lr
				break
			}
		}
		if l != nil {
			break
		}
	}
	if l == nil {
		t.Skip("no layer with Position property in fixture")
	}

	flat := l.Position()
	tree := l.PropertyByPath(aep.MatchNameGroupTransform, aep.MatchNamePosition)
	if flat == nil || tree == nil {
		t.Fatalf("Position not found: flat=%v tree=%v", flat, tree)
	}
	if flat != tree {
		t.Errorf("flat Position (%p) and tree Position (%p) are different instances", flat, tree)
	}
}
