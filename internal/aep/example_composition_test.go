package aep_test

import "github.com/example/aep-parser/internal/aep"

// Doc examples for docgen (compile-checked, no "// Output:").

func ExampleComposition_SetResolutionFactor() {
	var proj *aep.Project
	comp := proj.CompositionByName("Main")
	if comp == nil {
		return
	}
	_ = comp.SetResolutionFactor(2, 2) // half-resolution preview
	_ = comp.SetResolutionFactor(1, 1) // back to full
}

func ExampleComposition_SetWorkArea() {
	var proj *aep.Project
	if comp := proj.CompositionByName("Main"); comp != nil {
		_ = comp.SetWorkArea(1.5, 4.5)         // start, end (seconds)
		_ = comp.SetWorkArea(0, comp.Duration) // reset work area
	}
}

func ExampleComposition_LayerByName() {
	var proj *aep.Project
	if comp := proj.CompositionByName("Main"); comp != nil {
		if logo := comp.LayerByName("Logo"); logo != nil {
			_ = logo.SetVisible(false)
		}
	}
}
