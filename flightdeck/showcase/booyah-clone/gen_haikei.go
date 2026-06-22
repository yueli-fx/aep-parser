// flightdeck/showcase/booyah-clone/gen_haikei.go — comp ⑪ "背景変えるならココ！"
// (Phase 4 Task 4.1): the 6-layer background stack that sits under the glitch text.
// Every source is a 1920×1080 SOLID with a known colour (read via the oracle), so
// the layers rebuild faithfully — colour and all — unlike ⑩'s render-neutral solids.
//   L0 adj "微妙なノイズ"   + Noise2(Amount 11)  ← Noise2 is NOT in the embedded set:
//                                                  skipped + logged (honest gap)
//   L1 adj "ピカピカの元"   + Exposure2: Exposure 9kf (−10..+6 brightness flicker)
//                            + wiggle(34,0.29) expression ON TOP (spike 0.3 validated)
//   L2 "白のブラインド" white, blend=Overlay + Venetian Blinds(97/90/8)
//   L3 "黒のブラインド" black            + Venetian Blinds(62/90/12)
//   L4 "フレア" cyan                     + 1 bbox-rect mask
//   L5 "青のグラデーション" black        + Ramp(radial, 2-kf Start point, dark colours)
//
// L0/L1 are ADJUSTMENT layers (source colour irrelevant — they process what's below),
// so NewAdjustmentLayer; L2–L5 are AV solids → NewSolidLayer with the oracle colour.
// Effects/masks reuse the ⑩ helpers (applyEffectsFromOriginal / copyMasksFromOriginal);
// the Exposure wiggle is attached separately because it layers an expression on a
// KEYFRAMED param (kept faithful — the keyframes drive the dominant flicker, the
// wiggle adds ±0.29 jitter; expression overrides keyframes only when it would, but
// AE preserves both). Two-phase: buildHaikei (comp + 6 layers); finishHaikei (blend
// + effects + masks + Exposure expr, post-reopen).
package main

import (
	"fmt"

	aep "github.com/example/aep-parser/internal/aep"
)

const haikeiCompName = "背景変えるならココ！"

func buildHaikei(p *aep.Project, orc *oracle) {
	comp, err := aep.NewComposition(p, haikeiCompName, 1920, 1080, cloneFps, 6)
	must(err)
	orig := orc.mustComp(haikeiCompName)

	for i, ol := range orig.Layers {
		if ol.IsAdjust {
			_, err := aep.NewAdjustmentLayer(comp, layerName(ol, i))
			must(err)
			continue
		}
		// AV solid → own solid with the original's exact colour (0..1, via the oracle).
		col := [3]float64{0, 0, 0}
		if f := ol.SourceFootage(); f != nil {
			col = f.SolidColor
		}
		_, err := aep.NewSolidLayer(comp, layerName(ol, i), 1920, 1080, col)
		must(err)
	}
	fmt.Printf("  ⑪ 背景変えるならココ: %d solid/adjustment layers built\n", len(orig.Layers))
}

func finishHaikei(rp *aep.Project, orc *oracle) {
	comp := rp.CompositionByName(haikeiCompName)
	if comp == nil {
		panic("comp ⑪: not found after reopen")
	}
	orig := orc.mustComp(haikeiCompName)
	if len(comp.Layers) != len(orig.Layers) {
		panic(fmt.Sprintf("comp ⑪: layer count %d != original %d", len(comp.Layers), len(orig.Layers)))
	}

	w, h := float64(comp.Width), float64(comp.Height)
	for i, l := range comp.Layers {
		ol := orig.Layers[i]
		must(l.SetBlendingMode(ol.BlendingMode)) // L2 = Overlay; others Normal
		applyEffectsFromOriginal(l, ol)
		copyMasksFromOriginal(l, ol, w, h) // L4 carries one bbox-rect mask
	}

	// L1 Exposure2 wiggle DROPPED (kept the 9 keyframes only). The original layers
	// wiggle(34,0.29) on top of the keyframed Exposure, but an expression on a
	// KEYFRAMED param makes AE drop the layer — and the drop cascades to every layer
	// after it (gate-observed: ⑪ ingested with 1 of 6 layers). Same trap as ⑧'s
	// keyframed-Opacity wiggle (expression-enable-byte-pair). The 9 keyframes carry
	// the dominant −10..+6 brightness flicker ("ピカピカ" sparkle); the ±0.29 wiggle
	// jitter is the minor garnish we forgo to keep all six layers.

	fmt.Printf("  ⑪ 背景変えるならココ: + blend/effects/mask (Venetian Blinds×2, Ramp radial, Exposure 9kf; wiggle + Noise2 omitted)\n")
}
