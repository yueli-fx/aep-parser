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
	"encoding/binary"
	"fmt"
	"math"

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

// finishGlitchTextMasks (STEP 2) rebuilds every original mask on the matching
// clone layer. All 131 masks in ⑩ are axis-aligned rectangles (audited: 4
// vertices, Add mode, unit-square ldat) — the tearing slices that gate the
// displacement/scale glitch. AE stores a mask outline as a bbox (shph @0x04 =
// L,T,R,B fractions of the SOURCE pixel space) plus a unit-square ldat; the
// parser's Mask.Vertices surfaces only the normalized ldat (all masks read
// identically), so the real per-mask geometry must be reconstructed from the
// shph bbox. We map each bbox to a pixel rectangle and feed it to AddMask, which
// re-divides by the (1920×1080) source dims back to the original fractions —
// a faithful round-trip for these rectangular masks.
func finishGlitchTextMasks(rp *aep.Project, orc *oracle) {
	comp := rp.CompositionByName(glitchTextCompName)
	if comp == nil {
		panic("comp ⑩: not found after reopen (masks)")
	}
	orig := orc.mustComp(glitchTextCompName)
	w, h := float64(comp.Width), float64(comp.Height) // source = own 1920×1080 solid
	total := 0
	for i, l := range comp.Layers {
		total += copyMasksFromOriginal(l, orig.Layers[i], w, h)
	}
	fmt.Printf("  ⑩ グリッチテキスト: + %d masks (bbox-rect tearing slices) across L0–L20\n", total)
}

// copyMasksFromOriginal rebuilds origLayer's masks on newLayer as bbox-rect
// pixel paths and returns the count added. w/h are the layer source pixel dims.
func copyMasksFromOriginal(newLayer, origLayer *aep.Layer, w, h float64) int {
	n := 0
	for _, om := range origLayer.Masks {
		rect, ok := maskBBoxRect(om, w, h)
		if !ok {
			panic(fmt.Sprintf("comp ⑩: mask %q has no decodable shph bbox", om.Name))
		}
		nm, err := aep.AddMask(newLayer, om.Name, rect)
		must(err)
		if om.Mode != aep.MaskModeAdd {
			must(nm.SetMode(om.Mode))
		}
		if om.Inverted {
			must(nm.SetInverted(true))
		}
		n++
	}
	return n
}

// maskBBoxRect decodes a mask's shph bounding box (@0x04 = four big-endian
// float32 L,T,R,B fractions of the source pixel space) into a closed pixel
// rectangle (top-left → top-right → bottom-right → bottom-left). Returns false
// when the shph is too short to carry a bbox.
func maskBBoxRect(m *aep.Mask, w, h float64) (aep.BezierPath, bool) {
	d := m.ShphRaw
	if len(d) < 0x14 {
		return aep.BezierPath{}, false
	}
	f := func(off int) float64 {
		return float64(math.Float32frombits(binary.BigEndian.Uint32(d[off : off+4])))
	}
	l, t, r, b := f(0x04), f(0x08), f(0x0C), f(0x10)
	x0, y0, x1, y1 := l*w, t*h, r*w, b*h
	return aep.BezierPath{
		Vertices: [][2]float64{{x0, y0}, {x1, y0}, {x1, y1}, {x0, y1}},
		Closed:   true,
	}, true
}
