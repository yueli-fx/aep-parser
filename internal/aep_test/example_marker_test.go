package aep_test

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// Doc examples for docgen (compile-checked, no "// Output:" so `go test` does
// not run them). `proj` stands in for a parsed *aep.Project.

func ExampleMarker_SetComment() {
	var proj *aep.Project
	comp := proj.CompositionByID(1)
	if comp == nil {
		return
	}
	for _, m := range comp.Markers {
		_ = m.SetComment("scene start")
		fmt.Printf("@%.2fs %q\n", m.Time, m.Comment)
	}
}

func ExampleMarker_SetTime() {
	var proj *aep.Project
	if comp := proj.CompositionByID(1); comp != nil && len(comp.Markers) > 0 {
		_ = comp.Markers[0].SetTime(2.5)
	}
}

func ExampleMarker_SetLabel() {
	var proj *aep.Project
	if comp := proj.CompositionByID(1); comp != nil && len(comp.Markers) > 0 {
		_ = comp.Markers[0].SetLabel(9) // timeline label color 0..16
	}
}
