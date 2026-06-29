package aep_test

import "github.com/yueli-fx/aep-parser/internal/aep"

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

// ExampleAddEffect adds a built-in effect to a parsed layer and tunes one of its
// parameters. Use a typed match-name constant (aep.Effect*); discover the full
// addable set with aep.SupportedEffects().
func ExampleAddEffect() {
	var comp *aep.Composition
	layer := comp.LayerByID(1)
	if layer == nil {
		return
	}
	fx, err := aep.AddEffect(layer, aep.EffectGaussianBlur)
	if err != nil {
		return // e.g. unsupported effect, or layer has no Effect Parade
	}
	// Tune a parameter on the freshly added effect (effect params accept
	// SetStaticValue once located by match-name).
	for _, p := range fx.Parameters {
		if len(p.Keyframes) == 0 {
			_ = p.SetStaticValue(15.0)
			break
		}
	}
}

// ExampleSetEffectParam sets effect parameters by full match-name. Unlike raw
// Property.SetStaticValue it also works on default-elided parameters (a
// default effect instance only persists params whose value differs from the
// default) — the missing value stream is materialized first, then written.
func ExampleSetEffectParam() {
	var comp *aep.Composition
	layer := comp.LayerByID(1)
	if layer == nil {
		return
	}
	fx, err := aep.AddEffect(layer, aep.EffectGaussianBlur)
	if err != nil {
		return
	}
	// Works although a fresh Gaussian Blur exposes none of these params yet.
	if _, err := aep.SetEffectParam(layer, fx, "ADBE Gaussian Blur 2-0001", 25.0); err != nil {
		return // e.g. unsupported control type while default-elided
	}
	_, _ = aep.SetEffectParam(layer, fx, "ADBE Gaussian Blur 2-0003", 1.0) // bools are 0/1 float64
}

// ExampleRemoveEffect removes the first effect from a layer's Effect Parade —
// the inverse of AddEffect.
func ExampleRemoveEffect() {
	var comp *aep.Composition
	layer := comp.LayerByID(1)
	if layer == nil {
		return
	}
	if len(layer.Effects) > 0 {
		_ = aep.RemoveEffect(layer, 0)
	}
}
