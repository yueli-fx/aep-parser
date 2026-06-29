package aep_test

import "github.com/yueli-fx/aep-parser/internal/aep"

// Doc examples for docgen (compile-checked, no "// Output:").

func ExampleMask_SetMode() {
	var comp *aep.Composition
	layer := comp.LayerByID(1)
	if layer == nil {
		return
	}
	for _, m := range layer.Masks {
		_ = m.SetMode(aep.MaskModeSubtract)
	}
}

func ExampleMask_SetColor() {
	var comp *aep.Composition
	if layer := comp.LayerByID(1); layer != nil {
		for _, m := range layer.Masks {
			_ = m.SetColor([3]uint8{0xFF, 0x88, 0x00}) // timeline label color
		}
	}
}
