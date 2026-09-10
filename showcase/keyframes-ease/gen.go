// showcase/keyframes-ease/gen.go — from-scratch (no AE) showcase of
// temporal-ease keyframes. Three dots travel the same span (x 200→1700 over 4s)
// at different rows with different easing; rendered at the MID time (t=2s) the
// ease shows up as a horizontal position difference in a single still:
//   - LIN  (linear)      → at t=2s sits at the midpoint x≈950
//   - EOUT (ease-out start, influence 0.9) → lags far behind (x≪950)
//   - EIO  (ease in+out)  → eased both ends, sits near but not at 950
// Writes keyframes_ease.aep next to this file. render.jsx renders time=2s.
// Run from repo root: `go run ./showcase/keyframes-ease`.
package main

import (
	"fmt"
	"os"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/codec"
)

const outPath = "showcase/keyframes-ease/keyframes_ease.aep"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	p := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(p, "KeyframesEase", 1920, 1080, 30, 4)
	must(err)

	bg, err := aep.NewShapeLayer(comp, "BG")
	must(err)
	bgr, err := bg.RootGroup().AddRect()
	must(err)
	must(bgr.SetSize([2]float64{2200, 1300}))
	bgf, err := bg.RootGroup().AddFill()
	must(err)
	must(bgf.SetColor([4]float64{0.07, 0.08, 0.11, 1}))
	must(bg.Position().SetStaticValue([2]float64{960, 540}))

	dot := func(name string, color [4]float64) *aep.ShapeLayer {
		dl, err := aep.NewShapeLayer(comp, name)
		must(err)
		e, err := dl.RootGroup().AddEllipse()
		must(err)
		must(e.SetSize([2]float64{84, 84}))
		f, err := dl.RootGroup().AddFill()
		must(err)
		must(f.SetColor(color))
		return dl
	}

	// LIN — linear baseline. At t=2s x=950.
	lin := dot("LIN_linear", [4]float64{1.0, 0.55, 0.1, 1})
	must(lin.Position().AddKeyframeLinear(0, [2]float64{200, 300}))
	must(lin.Position().AddKeyframeLinear(4, [2]float64{1700, 300}))

	// EOUT — slow ease out of the start keyframe (influence 0.9). At t=2s it
	// still lags far left of 950.
	eout := dot("EOUT_easeOut", [4]float64{0.25, 0.85, 1.0, 1})
	slowOut := codec.TemporalEase{Speed: 0, Influence: 0.9}
	must(eout.Position().AddKeyframeWithEase(0, [2]float64{200, 540}, codec.TemporalEase{}, slowOut))
	must(eout.Position().AddKeyframeWithEase(4, [2]float64{1700, 540}, codec.TemporalEase{Speed: 0, Influence: 0.1}, codec.TemporalEase{}))

	// EIO — eased into BOTH keyframes (classic ease-in-out). At t=2s near centre.
	eio := dot("EIO_easeInOut", [4]float64{1.0, 0.2, 0.55, 1})
	soft := codec.TemporalEase{Speed: 0, Influence: 0.6}
	must(eio.Position().AddKeyframeWithEase(0, [2]float64{200, 780}, codec.TemporalEase{}, soft))
	must(eio.Position().AddKeyframeWithEase(4, [2]float64{1700, 780}, soft, codec.TemporalEase{}))

	rp, err := aep.Reopen(p)
	must(err)
	must(aep.MoveToEnd(rp.Compositions[0].LayerByName("BG")))

	out, err := os.Create(outPath)
	must(err)
	defer out.Close()
	must(rp.WriteAEP(out))
	fmt.Printf("wrote %s (%d layers)\n", outPath, len(rp.Compositions[0].Layers))
}
