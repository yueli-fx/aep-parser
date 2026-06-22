// flightdeck/showcase/booyah-clone/gen_glitch_text.go — comp ⑩ "グリッチテキスト".
// The MONSTER comp: 27 layers, ~100 masks, 11 Displacement Maps — the work/risk
// centre of the whole replication (Phase 3, plan Task 3.1). Built in stages:
//   STEP 1 (this file's first commit): no-mask SKELETON — 27 layers with correct
//     source refs, blend modes, timing (start/in/out) and visibility. Anchors the
//     layer structure and tests that 27 layers survive AE (multi-layer-silent-drop).
//   STEP 2 (later): batch masks (copyMasksFromOriginal) on L0–L20.
//   STEP 3 (later): effect chains (Geometry2 / Glo2 / Displacement Map / Fractal
//     Noise / Gaussian Blur), incl. the Displacement-Map layer references.
//
// Source map — driven by reading the original comp's layers via the oracle:
//   precomp layers      → the clone comp of the same name (⑦ ここは開けない / ②
//                         テキスト / ⑧ RGBズレ / ③ マップ用ノイズ), built earlier
//   footage-backed adj  → own NewAdjustmentLayer (footage-share = own solid, a
//                         render-neutral structural delta, same call as ③)
//   footage-backed av   → own NewSolidLayer (L24/L26; Fractal Noise paints pixels)
//
// Two-phase: buildGlitchText creates the comp + 27 layers (sources only; needs the
// child comps ⑦②⑧③ to exist already); finishGlitchText applies blend/timing/
// visibility after the reopen (those ldta edits are unreliable on un-Reopened
// embed-template clones — same discipline as ⑤⑥⑦).
package main

import (
	"fmt"

	aep "github.com/example/aep-parser/internal/aep"
)

const glitchTextCompName = "グリッチテキスト"

func buildGlitchText(p *aep.Project, orc *oracle) {
	comp, err := aep.NewComposition(p, glitchTextCompName, 1920, 1080, cloneFps, 6)
	must(err)
	orig := orc.mustComp(glitchTextCompName)

	for i, ol := range orig.Layers {
		switch {
		case ol.SourceComposition() != nil:
			// Precomp layer → the clone comp of the same name (built earlier in main).
			src := ol.SourceComposition()
			child := p.CompositionByName(src.Name)
			if child == nil {
				panic(fmt.Sprintf("comp ⑩ L%d: child comp %q not found in clone", i, src.Name))
			}
			_, err := aep.NewPrecompLayer(comp, child, layerName(ol, i))
			must(err)
		case ol.IsAdjust:
			// Footage-backed adjustment layer → own adjustment (own solid source).
			_, err := aep.NewAdjustmentLayer(comp, layerName(ol, i))
			must(err)
		default:
			// Footage-backed AV solid (L24/L26) → own black solid; Fractal Noise paints it.
			_, err := aep.NewSolidLayer(comp, layerName(ol, i), 1920, 1080, [3]float64{0, 0, 0})
			must(err)
		}
	}
	fmt.Printf("  ⑩ グリッチテキスト: skeleton %d layers built (sources only; masks+effects deferred)\n", len(orig.Layers))
}

// layerName returns the original layer's display name, or a synthetic stable name
// when the original is empty (AE needs a non-empty solid/adjustment source name; the
// glitch comp's adjustment layers are all unnamed in the original).
func layerName(ol *aep.Layer, i int) string {
	if ol.Name != "" {
		return ol.Name
	}
	return fmt.Sprintf("L%02d", i)
}

func finishGlitchText(rp *aep.Project, orc *oracle) {
	comp := rp.CompositionByName(glitchTextCompName)
	if comp == nil {
		panic("comp ⑩: not found after reopen")
	}
	orig := orc.mustComp(glitchTextCompName)
	if len(comp.Layers) != len(orig.Layers) {
		panic(fmt.Sprintf("comp ⑩: layer count %d != original %d (silent drop?)", len(comp.Layers), len(orig.Layers)))
	}

	for i, l := range comp.Layers {
		ol := orig.Layers[i]
		// Timing: source-offset start (@0x0C) + in/out trim. The short adjustment
		// spans (e.g. L0 [1.2→1.3]) are the staccato glitch slices — set in then out
		// so the duration recompute lands on the trimmed span.
		must(l.SetStartTime(ol.StartTime))
		must(l.SetInPoint(ol.InPoint()))
		must(l.SetOutPoint(ol.OutPoint()))
		// Blend mode (L7 = Add; every other layer Normal).
		must(l.SetBlendingMode(ol.BlendingMode))
		// Visibility (L24/L25/L26 are hidden — video switch off — in the original).
		must(l.SetVisible(ol.Visible))
	}
	fmt.Printf("  ⑩ グリッチテキスト: + blend/timing/visibility on %d layers\n", len(comp.Layers))
}
