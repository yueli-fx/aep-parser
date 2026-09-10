// showcase/booyah-clone/gen_fractal_map.go — comp ③ "マップ用フラクタルノイズ".
// Two black-solid layers, each carrying one Fractal Noise effect; their composited
// noise is the displacement MAP that comp ⑩ samples. L0 "…A" blends Overlay over
// L1 "…B" (Normal). Both effects share the same tuned params except Scale Width and
// the Evolution expression speed; Offset Turbulence pans horizontally (2-kf 2D).
//
// Footage-share note: the original's two layers reference ONE shared black solid
// (footage id 67). We have no from-scratch API to make a second layer reference an
// existing footage item (only SetSource-retarget, which orphans the donor item, or
// DuplicateLayer, which inherits the effect and needs a fragile post-duplicate
// reopen). Since the solid is plain black 1920×1080 and Fractal Noise GENERATES its
// own pixels, two separate identical solids render identically to one shared solid —
// a render-neutral structural delta, logged in the ledger. 29.97 fps (cloneFps),
// matching the original (every Booyah comp is NTSC 29.97).
//
// Two-phase: buildFractalMap creates the comp + the two solids before reopen;
// finishFractalMap adds + tunes the effects, which AddEffect refuses on un-Reopened
// fresh layers (see the Reopen in main).
package main

import (
	"fmt"
	"github.com/yueli-fx/aep-parser/internal/serializer"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

const fractalMapCompName = "マップ用フラクタルノイズ"

var fractalMapLayerNames = []string{"マップ用フラクタルノイズA", "マップ用フラクタルノイズB"}

func buildFractalMap(p *aep.Project, orc *oracle) {
	comp, err := aep.NewComposition(p, fractalMapCompName, 1920, 1080, cloneFps, 6)
	must(err)
	// Create A then B: NewSolidLayer appends at the bottom of the stack, so A lands
	// at index 0 (top) and B at index 1 — matching the original (L0=A over L1=B).
	black := [3]float64{0, 0, 0}
	for _, name := range fractalMapLayerNames {
		if _, err := aep.NewSolidLayer(comp, name, 1920, 1080, black); err != nil {
			must(err)
		}
	}
}

// finishFractalMap runs after the project is reopened: it operates on the parsed
// solid layers (AddEffect requires a parsed layer).
func finishFractalMap(rp *aep.Project, orc *oracle) {
	comp := rp.CompositionByName(fractalMapCompName)
	if comp == nil {
		panic("comp ③: not found after reopen")
	}
	origComp := orc.mustComp(fractalMapCompName)

	for i, name := range fractalMapLayerNames {
		layer := comp.LayerByName(name)
		if layer == nil {
			panic("comp ③: layer not found after reopen: " + name)
		}
		applyFractalNoise(layer, origComp.Layers[i])
	}

	// L0 "…A" blends Overlay; L1 "…B" stays Normal (default).
	if err := comp.LayerByName(fractalMapLayerNames[0]).SetBlendingMode(aep.BlendingModeOverlay); err != nil {
		must(err)
	}

	fmt.Printf("  ③ マップ用フラクタルノイズ: 2 solids + Fractal Noise (Scale W 211/2847, Evolution time*1200/time*3000, Offset Turbulence 2kf), L0 blend=Overlay\n")
}

// applyFractalNoise adds a Fractal Noise effect to layer and copies the tuned
// params, the 2-kf Offset Turbulence pan, and the Evolution expression from the
// original layer's Fractal Noise (values pulled via the oracle, not byte-copied).
func applyFractalNoise(layer *aep.Layer, orig *aep.Layer) {
	fx, err := aep.AddEffect(layer, "ADBE Fractal Noise")
	must(err)
	of := orig.Effects[0] // the original layer's single Fractal Noise effect

	// Tuned scalar params (match-name → human name): copy each value from the original.
	for _, mn := range []string{
		"ADBE Fractal Noise-0002", // Noise Type   (1, default 3)
		"ADBE Fractal Noise-0004", // Contrast      (254, default 100)
		"ADBE Fractal Noise-0009", // Uniform Scaling (0=off, default 1)
		"ADBE Fractal Noise-0011", // Scale Width   (211 / 2847)
		"ADBE Fractal Noise-0012", // Scale Height  (12)
		"ADBE Fractal Noise-0015", // Complexity    (1, default 6)
	} {
		op := fxParam(of, mn)
		if op == nil {
			panic("comp ③: original Fractal Noise missing param " + mn)
		}
		if _, err := serializer.SetEffectParam(layer, fx, mn, toScalar(op.StaticValue)); err != nil {
			must(err)
		}
	}

	// Offset Turbulence (ADBE Fractal Noise-0013): 2-kf horizontal pan (2D point).
	ot := fxParam(of, "ADBE Fractal Noise-0013")
	vkfs := make([]aep.VectorKeyframe, 0, len(ot.Keyframes))
	for _, kf := range ot.Keyframes {
		vkfs = append(vkfs, aep.VectorKeyframe{Time: kf.Time, Value: toFloats(kf.Value)})
	}
	if _, err := aep.AnimateEffectParamVec(layer, fx, "Offset Turbulence", vkfs); err != nil {
		must(err)
	}

	// Evolution (ADBE Fractal Noise-0023): driven by expression "time*N" (Task 0.3 GO).
	evoSrc := fxParam(of, "ADBE Fractal Noise-0023").Expression
	evo, err := aep.SetEffectParam(layer, fx, "Evolution", 0.0)
	must(err)
	must(evo.SetExpression(evoSrc))
	must(evo.SetExpressionEnabled(true))
}

// fxParam returns the effect's parameter with the given match-name, or nil.
func fxParam(fx *aep.Effect, matchName string) *aep.Property {
	for _, pr := range fx.Parameters {
		if pr.MatchName == matchName {
			return pr
		}
	}
	return nil
}
