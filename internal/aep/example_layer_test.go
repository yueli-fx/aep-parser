package aep_test

import "github.com/example/aep-parser/internal/aep"

// Doc examples for docgen (compile-checked, no "// Output:").

func ExampleLayer_SetName() {
	var comp *aep.Composition
	if l := comp.LayerByName("Logo"); l != nil {
		_ = l.SetName("Logo (final)")
	}
}

func ExampleLayer_Position() {
	var comp *aep.Composition
	l := comp.LayerByID(1)
	if l == nil {
		return
	}
	if pos := l.Position(); pos != nil && len(pos.Keyframes) == 0 {
		_ = pos.SetStaticValue([]float64{960, 540, 0})
	}
}

func ExampleLayer_SetVisible() {
	var comp *aep.Composition
	if l := comp.LayerByName("BG"); l != nil {
		_ = l.SetVisible(false) // hide the layer (video disabled)
	}
}
