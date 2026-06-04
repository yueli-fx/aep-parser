package aep_test

import "github.com/example/aep-parser/internal/aep"

// Doc example for docgen (compile-checked, no "// Output:"). Attaches to the
// Effect.Parameters attribute via the Example<Type>_<Field> naming.

func ExampleEffect_Parameters() {
	var comp *aep.Composition
	layer := comp.LayerByID(1)
	if layer == nil {
		return
	}
	// set Gaussian Blur's Blurriness to 20
	for _, fx := range layer.Effects {
		if fx.MatchName != "ADBE Gaussian Blur 2" {
			continue
		}
		for _, p := range fx.Parameters {
			if p.MatchName == "ADBE Gaussian Blur 2-0001" && len(p.Keyframes) == 0 {
				_ = p.SetStaticValue(20.0)
			}
		}
	}
}
