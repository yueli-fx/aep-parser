package aep_test

import (
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// TestCompositionViews_ReCameraLight exercises every filter view against
// re_cameralight.aep, which contains a known mix of layer kinds:
//   - "Comp 1": 16 layers (1 shape, 1 text, rest AV layers with footage source)
//   - "RE_CL":   3 layers (2 light, 1 camera) — all Is3D=true
//   - "RE_TEXT": 12 text layers
func TestCompositionViews_ReCameraLight(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present: %v", err)
	}

	findComp := func(name string) *aep.Composition {
		for _, c := range proj.Compositions {
			if c.Name == name {
				return c
			}
		}
		return nil
	}

	t.Run("Comp1", func(t *testing.T) {
		c := findComp("Comp 1")
		if c == nil {
			t.Fatal("Comp 1 not found")
		}
		checkLen(t, "ShapeLayers", c.ShapeLayers(), 2) // 自定义曲线 + 形状测试
		checkLen(t, "TextLayers", c.TextLayers(), 1)   // 文字测试
		checkLen(t, "CameraLayers", c.CameraLayers(), 0)
		checkLen(t, "LightLayers", c.LightLayers(), 0)
		checkLen(t, "NullLayers", c.NullLayers(), 0)
		checkLen(t, "ThreeDLayers", c.ThreeDLayers(), 0)
		checkLen(t, "GuideLayers", c.GuideLayers(), 0)
		checkLen(t, "SoloLayers", c.SoloLayers(), 0)
		// 13 AV layers, all with solid footage sources
		checkLen(t, "AVLayers", c.AVLayers(), 13)
		checkLen(t, "FootageLayers", c.FootageLayers(), 13)
		checkLen(t, "CompositionLayers", c.CompositionLayers(), 0)
		checkLen(t, "FileLayers", c.FileLayers(), 0)
		checkLen(t, "SolidLayers", c.SolidLayers(), 13)
		checkLen(t, "PlaceholderLayers", c.PlaceholderLayers(), 0)
	})

	t.Run("RE_CL", func(t *testing.T) {
		c := findComp("RE_CL")
		if c == nil {
			t.Fatal("RE_CL not found")
		}
		checkLen(t, "CameraLayers", c.CameraLayers(), 1)
		checkLen(t, "LightLayers", c.LightLayers(), 2)
		checkLen(t, "ThreeDLayers", c.ThreeDLayers(), 3)
		checkLen(t, "AVLayers", c.AVLayers(), 0)
		checkLen(t, "ShapeLayers", c.ShapeLayers(), 0)
		checkLen(t, "TextLayers", c.TextLayers(), 0)
	})

	t.Run("RE_TEXT", func(t *testing.T) {
		c := findComp("RE_TEXT")
		if c == nil {
			t.Fatal("RE_TEXT not found")
		}
		checkLen(t, "TextLayers", c.TextLayers(), 12)
		checkLen(t, "AVLayers", c.AVLayers(), 0)
		checkLen(t, "ShapeLayers", c.ShapeLayers(), 0)
	})
}

// TestCompositionViews_EmptyComp ensures filter views on an empty comp
// return non-nil empty slices (safe to range over without nil checks).
func TestCompositionViews_EmptyComp(t *testing.T) {
	c := &aep.Composition{Name: "Empty"}
	views := map[string][]*aep.Layer{
		"TextLayers":        c.TextLayers(),
		"ShapeLayers":       c.ShapeLayers(),
		"CameraLayers":      c.CameraLayers(),
		"LightLayers":       c.LightLayers(),
		"NullLayers":        c.NullLayers(),
		"AdjustmentLayers":  c.AdjustmentLayers(),
		"ThreeDLayers":      c.ThreeDLayers(),
		"GuideLayers":       c.GuideLayers(),
		"SoloLayers":        c.SoloLayers(),
		"AVLayers":          c.AVLayers(),
		"CompositionLayers": c.CompositionLayers(),
		"FootageLayers":     c.FootageLayers(),
		"FileLayers":        c.FileLayers(),
		"SolidLayers":       c.SolidLayers(),
		"PlaceholderLayers": c.PlaceholderLayers(),
	}
	for name, v := range views {
		if v == nil {
			t.Errorf("%s returned nil; want non-nil empty slice", name)
		}
		if len(v) != 0 {
			t.Errorf("%s len = %d; want 0 on empty comp", name, len(v))
		}
	}
}

func checkLen(t *testing.T, name string, got []*aep.Layer, want int) {
	t.Helper()
	if len(got) != want {
		names := make([]string, len(got))
		for i, l := range got {
			names[i] = l.Name
		}
		t.Errorf("%s len = %d, want %d (names=%v)", name, len(got), want, names)
	}
}
