// showcase/procedural-fx/gen.go — from-scratch (no AE) showcase of the
// procedural-FX generator's flame (spec 2026-06-18-procedural-fx-generator). v3 =
// MULTI-LAYER composite for depth (technique T3 additive-depth, docs/fx-techniques.md):
// black bg + 3 Fractal Noise->Tritone->Turbulent Displace fire layers at different
// noise scales, CONCENTRIC masks for outer->inner temperature zones (deep-red wispy
// outside -> orange -> yellow-white hot core), Add-blended so overlaps build the hot
// core, + a top Glo2 Glow adjustment. Motion is SHARED across layers (coherent — per-
// layer rates desync and shimmer late). User-accepted on real machine 2026-06-18.
// Writes flame.aep next to this file. Run from repo root: `go run ./showcase/procedural-fx`.
//
// Same recipe as TestFlameDemo_AEShipGate (double-version gated). The deterministic
// recipe the AI layer (later phases) will parameterize; here every knob is hand-set.
package main

import (
	"fmt"
	"os"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

const outPath = "showcase/procedural-fx/flame.aep"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

type fireCfg struct {
	name                                             string
	contrast, brightness, scaleW, scaleH, complexity float64
	dispAmt, dispSize, maskScale                     float64
	tShadow, tMid, tHigh                             []float64 // [A,R,G,B] 0-255
}

// scalePath shrinks a path toward (cx,cy) by f (concentric flame zones).
func scalePath(p aep.BezierPath, cx, cy, f float64) aep.BezierPath {
	out := aep.BezierPath{Closed: p.Closed}
	for _, v := range p.Vertices {
		out.Vertices = append(out.Vertices, [2]float64{cx + (v[0]-cx)*f, cy + (v[1]-cy)*f})
	}
	return out
}

func main() {
	p := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(p, "FLAME", 1080, 1920, 30, 4)
	must(err)

	// Build top->bottom (new layers append below): Glow adj, core, mid, base, BG.
	_, err = aep.NewAdjustmentLayer(comp, "Glow")
	must(err)
	cfgs := []fireCfg{
		{name: "FireCore", contrast: 185, brightness: -18, scaleW: 30, scaleH: 420, complexity: 6,
			dispAmt: 38, dispSize: 28, maskScale: 0.5,
			tShadow: []float64{255, 30, 0, 0}, tMid: []float64{255, 255, 150, 25}, tHigh: []float64{255, 255, 250, 235}},
		{name: "FireMid", contrast: 158, brightness: -24, scaleW: 50, scaleH: 330, complexity: 6,
			dispAmt: 50, dispSize: 38, maskScale: 0.76,
			tShadow: []float64{255, 16, 0, 0}, tMid: []float64{255, 245, 80, 0}, tHigh: []float64{255, 255, 195, 70}},
		{name: "FireBase", contrast: 128, brightness: -30, scaleW: 82, scaleH: 260, complexity: 5,
			dispAmt: 62, dispSize: 46, maskScale: 1.0,
			tShadow: []float64{255, 14, 0, 0}, tMid: []float64{255, 175, 26, 0}, tHigh: []float64{255, 255, 110, 8}},
	}
	for _, c := range cfgs {
		_, err = aep.NewSolidLayer(comp, c.name, 1080, 1920, [3]float64{0, 0, 0})
		must(err)
	}
	_, err = aep.NewSolidLayer(comp, "BG", 1080, 1920, [3]float64{0, 0, 0})
	must(err)

	rp, err := aep.Reopen(p)
	must(err)
	fc := rp.Compositions[0]
	set := func(l *aep.Layer, fx *aep.Effect, mn string, v any) {
		if _, err := aep.SetEffectParam(l, fx, mn, v); err != nil {
			panic(fmt.Sprintf("%s %s: %v", l.Name, mn, err))
		}
	}

	glowL := fc.LayerByName("Glow")
	gl, err := aep.AddEffect(glowL, "ADBE Glo2")
	must(err)
	set(glowL, gl, "Glow Threshold", 50.0) // threshold
	set(glowL, gl, "Glow Radius", 55.0)    // radius
	set(glowL, gl, "Glow Intensity", 1.5)  // intensity

	flamePath := aep.BezierPath{
		Vertices: [][2]float64{
			{540, 250}, {700, 760}, {812, 1260}, {700, 1700},
			{540, 1785}, {380, 1700}, {268, 1260}, {380, 760},
		},
		Closed: true,
	}
	// SHARED, calm motion (coherent -> no desync shimmer).
	const evoEnd, offEndY, tdEvoEnd = 540.0, 540.0, 360.0

	for _, c := range cfgs {
		l := fc.LayerByName(c.name)
		fn, err := aep.AddEffect(l, aep.EffectFractalNoise)
		must(err)
		set(l, fn, "Contrast", c.contrast)
		set(l, fn, "Brightness", c.brightness)
		set(l, fn, "Uniform Scaling", 0.0)
		set(l, fn, "Scale Width", c.scaleW)
		set(l, fn, "Scale Height", c.scaleH)
		set(l, fn, "Complexity", c.complexity)

		tr, err := aep.AddEffect(l, "ADBE Tritone")
		must(err)
		set(l, tr, "Highlights", c.tHigh)
		set(l, tr, "Midtones", c.tMid)
		set(l, tr, "Shadows", c.tShadow)

		td, err := aep.AddEffect(l, aep.EffectTurbulentDisplace)
		must(err)
		set(l, td, "Amount", c.dispAmt)
		set(l, td, "Size", c.dispSize)

		mp := scalePath(flamePath, 540, 1080, c.maskScale)
		feather := 110 * c.maskScale
		if feather < 60 {
			feather = 60
		}
		mask, err := aep.AddMask(l, c.name+"Mask", mp)
		must(err)
		must(mask.SetFeather([2]float64{feather, feather}))

		must(l.SetBlendingMode(aep.BlendingModeAdd))

		_, err = aep.AnimateEffectParam(l, fn, "Evolution",
			[]aep.ScalarKeyframe{{Time: 0, Value: 0}, {Time: 4, Value: evoEnd}})
		must(err)
		_, err = aep.AnimateEffectParamVec(l, fn, "Offset Turbulence",
			[]aep.VectorKeyframe{{Time: 0, Value: []float64{540, 960}}, {Time: 4, Value: []float64{540, offEndY}}})
		must(err)
		_, err = aep.AnimateEffectParam(l, td, "Evolution",
			[]aep.ScalarKeyframe{{Time: 0, Value: 0}, {Time: 4, Value: tdEvoEnd}})
		must(err)
	}

	f, err := os.Create(outPath)
	must(err)
	defer f.Close()
	must(rp.WriteAEP(f))
	fmt.Println("wrote", outPath)
}
