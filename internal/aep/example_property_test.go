package aep_test

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// These Example functions are the documentation source for docgen (see
// cmd/docgen). They carry no "// Output:" line, so `go test` compiles but
// does not run them — they exist to compile-check the API spellings shown in
// docs/property.gen.md. `layer` / `prop` placeholders stand in for values a
// caller obtains from a parsed *aep.Project.

func ExampleProperty_SetStaticValue() {
	var layer *aep.Layer
	if op := layer.Opacity(); op != nil && len(op.Keyframes) == 0 {
		_ = op.SetStaticValue(0.5) // Opacity → 50%
	}
}

func ExampleProperty_SetStaticValue_multiComponent() {
	var layer *aep.Layer
	if pos := layer.Position(); pos != nil && len(pos.Keyframes) == 0 {
		_ = pos.SetStaticValue([]float64{960, 540, 0}) // Position → (960, 540, 0)
	}
}

func ExampleInsertKeyframe() {
	var pos *aep.Property
	kf, idx, err := aep.InsertKeyframe(pos, 2.5, []float64{960, 540, 0})
	if err == nil {
		_ = kf.SetInInterp(aep.InterpBezier)
		_ = kf.SetOutInterp(aep.InterpBezier)
		fmt.Println("inserted at index", idx)
	}
}

func ExampleDeleteKeyframe() {
	var op *aep.Property
	_ = aep.DeleteKeyframe(op, 2) // remove the 3rd keyframe
}

func ExampleProperty_SetExpression() {
	var pos *aep.Property
	_ = pos.SetExpression("wiggle(2, 30)")
	_ = pos.SetExpression("") // clear the expression
}

func ExampleProperty_SetExpressionEnabled() {
	var op *aep.Property
	_ = op.SetExpression("time * 50")
	_ = op.SetExpressionEnabled(false) // keep source, stop evaluating
	_ = op.SetExpressionEnabled(true)  // resume
}

func ExampleSetDimensionsSeparated() {
	var layer *aep.Layer
	pos := layer.Position()
	if err := aep.SetDimensionsSeparated(pos, true); err != nil {
		return
	}
	// access per-axis via layer.PropertyByMatchName("ADBE Position_0"), etc.
	_ = aep.SetDimensionsSeparated(pos, false) // merge back
}

func ExampleKeyframe_SetTime() {
	var layer *aep.Layer
	if pos := layer.Position(); pos != nil && len(pos.Keyframes) > 0 {
		_ = pos.Keyframes[0].SetTime(2.5)
	}
}

func ExampleKeyframe_SetValue() {
	var layer *aep.Layer
	if pos := layer.Position(); pos != nil && len(pos.Keyframes) > 0 {
		_ = pos.Keyframes[0].SetValue([]float64{960, 540, 0})
	}
}

func ExampleKeyframe_SetInTemporalEase() {
	var layer *aep.Layer
	if opa := layer.Opacity(); opa != nil && len(opa.Keyframes) > 0 {
		// 1D property: a single ease
		_ = opa.Keyframes[0].SetInTemporalEase([]aep.TemporalEase{{Speed: 1.5, Influence: 0.25}})
	}
}

func ExampleKeyframe_SetInSpatialTangent() {
	var layer *aep.Layer
	if pos := layer.Position(); pos != nil && len(pos.Keyframes) > 0 {
		_ = pos.Keyframes[0].SetInSpatialTangent([]float64{10, 20, 0})
	}
}
