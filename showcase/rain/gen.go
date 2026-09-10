// showcase/rain/gen.go — from-scratch, plugin-free rain showcase.
// It validates the rain recipe's native spine: shape/repeater/trim rain streaks,
// keyed falling motion, native mist, and Echo/Blur persistence. Exact Motionbox
// Rain Day render still needs Trapcode Particular + Unmult; this showcase is the
// authored stock-AE fallback route.
//
// Run from repo root: `go run ./showcase/rain`.
package main

import (
	"fmt"
	"os"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

const outPath = "showcase/rain/rain.aep"

type streakCfg struct {
	name         string
	x, y         float64
	length       float64
	width        float64
	copies       float64
	spacing      [2]float64
	color        [4]float64
	startY, endY float64
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	p := aep.NewProject(aep.TargetAE2020)

	streaks, err := aep.NewComposition(p, "RAIN_STREAKS", 1920, 1080, 30, 4)
	must(err)
	mainComp, err := aep.NewComposition(p, "RAIN", 1920, 1080, 30, 4)
	must(err)

	buildStreaks(streaks)
	buildMain(mainComp, streaks)

	rp, err := aep.Reopen(p)
	must(err)
	must(applyMainEffects(rp.CompositionByName("RAIN")))

	out, err := os.Create(outPath)
	must(err)
	defer out.Close()
	must(rp.WriteAEP(out))
	fmt.Printf("wrote %s\n", outPath)
}

func buildStreaks(comp *aep.Composition) {
	cfgs := []streakCfg{
		{name: "NearFast", x: 120, y: -160, length: 330, width: 8.0, copies: 18, spacing: [2]float64{112, -4}, color: [4]float64{0.82, 0.94, 1.0, 0.96}, startY: -170, endY: 1280},
		{name: "MidSheet", x: 60, y: -80, length: 235, width: 5.6, copies: 28, spacing: [2]float64{76, 10}, color: [4]float64{0.58, 0.78, 1.0, 0.84}, startY: -90, endY: 1180},
		{name: "FineBack", x: 35, y: -40, length: 150, width: 3.2, copies: 42, spacing: [2]float64{50, 18}, color: [4]float64{0.42, 0.62, 0.90, 0.66}, startY: -40, endY: 1080},
		{name: "BrightCuts", x: 220, y: -220, length: 275, width: 7.0, copies: 14, spacing: [2]float64{140, 2}, color: [4]float64{0.94, 0.98, 1.0, 0.92}, startY: -220, endY: 1260},
	}
	for _, cfg := range cfgs {
		l, err := aep.NewShapeLayer(comp, cfg.name)
		must(err)
		pth, err := l.RootGroup().AddPath()
		must(err)
		halfW := cfg.width / 2
		slant := cfg.length * 0.11
		must(pth.SetClosed(true))
		must(pth.SetVertices([][2]float64{
			{-halfW + slant, -cfg.length / 2},
			{halfW + slant, -cfg.length / 2},
			{halfW - slant, cfg.length / 2},
			{-halfW - slant, cfg.length / 2},
		}))
		fill, err := l.RootGroup().AddFill()
		must(err)
		must(fill.SetColor(cfg.color))
		rp, err := l.RootGroup().AddRepeater()
		must(err)
		must(rp.SetCopies(cfg.copies))
		must(rp.Transform().SetPosition(cfg.spacing))
		must(rp.Transform().SetStartOpacity(100))
		must(rp.Transform().SetEndOpacity(28))
		must(l.Position().AddKeyframeLinear(0, [2]float64{cfg.x, cfg.startY}))
		must(l.Position().AddKeyframeLinear(4, [2]float64{cfg.x + 360, cfg.endY}))
	}
}

func buildMain(comp, streaks *aep.Composition) {
	_, err := aep.NewPrecompLayer(comp, streaks, "RainStreaks")
	must(err)
	_, err = aep.NewPrecompLayer(comp, streaks, "EchoTrail")
	must(err)
	_, err = aep.NewSolidLayer(comp, "NativeMist", 1920, 1080, [3]float64{0, 0, 0})
	must(err)

	ground, err := aep.NewShapeLayer(comp, "WetGround")
	must(err)
	r, err := ground.RootGroup().AddRect()
	must(err)
	must(r.SetSize([2]float64{2100, 210}))
	f, err := ground.RootGroup().AddFill()
	must(err)
	must(f.SetColor([4]float64{0.08, 0.12, 0.18, 1}))
	must(ground.Position().SetStaticValue([2]float64{960, 995}))

	for i, x := range []float64{310, 520, 760, 1040, 1330, 1580} {
		splash, err := aep.NewShapeLayer(comp, fmt.Sprintf("Splash%02d", i+1))
		must(err)
		e, err := splash.RootGroup().AddEllipse()
		must(err)
		must(e.SetSize([2]float64{95 + float64(i%3)*28, 10 + float64(i%2)*5}))
		st, err := splash.RootGroup().AddStroke()
		must(err)
		must(st.SetColor([4]float64{0.45, 0.68, 0.92, 0.45}))
		must(st.SetWidth(2))
		tr, err := splash.RootGroup().AddTrim()
		must(err)
		must(tr.SetStart(8))
		must(tr.SetEnd(58))
		must(splash.Position().SetStaticValue([2]float64{x, 900 + float64(i%2)*38}))
	}

	bg, err := aep.NewShapeLayer(comp, "BG")
	must(err)
	br, err := bg.RootGroup().AddRect()
	must(err)
	must(br.SetSize([2]float64{2200, 1300}))
	bf, err := bg.RootGroup().AddFill()
	must(err)
	must(bf.SetColor([4]float64{0.025, 0.035, 0.055, 1}))
	must(bg.Position().SetStaticValue([2]float64{960, 540}))
}

func applyMainEffects(comp *aep.Composition) error {
	set := func(l *aep.Layer, fx *aep.Effect, mn string, v any) error {
		_, err := aep.SetEffectParam(l, fx, mn, v)
		return err
	}

	echoL := comp.LayerByName("EchoTrail")
	echo, err := aep.AddEffect(echoL, aep.EffectEcho)
	if err != nil {
		return err
	}
	if err := set(echoL, echo, "Echo Time (seconds)", -0.045); err != nil {
		return err
	}
	blur, err := aep.AddEffect(echoL, aep.EffectGaussianBlur)
	if err != nil {
		return err
	}
	if err := set(echoL, blur, "Blurriness", 2.5); err != nil {
		return err
	}
	glow, err := aep.AddEffect(echoL, "ADBE Glo2")
	if err != nil {
		return err
	}
	if err := set(echoL, glow, "Glow Threshold", 108.0); err != nil {
		return err
	}
	if err := set(echoL, glow, "Glow Radius", 18.0); err != nil {
		return err
	}
	if err := set(echoL, glow, "Glow Intensity", 0.7); err != nil {
		return err
	}
	if err := echoL.SetBlendingMode(aep.BlendingModeAdd); err != nil {
		return err
	}

	mistL := comp.LayerByName("NativeMist")
	fn, err := aep.AddEffect(mistL, aep.EffectFractalNoise)
	if err != nil {
		return err
	}
	if err := set(mistL, fn, "Contrast", 54.0); err != nil {
		return err
	}
	if err := set(mistL, fn, "Brightness", -92.0); err != nil {
		return err
	}
	if err := set(mistL, fn, "Scale Width", 620.0); err != nil {
		return err
	}
	if err := set(mistL, fn, "Scale Height", 180.0); err != nil {
		return err
	}
	if _, err := aep.AnimateEffectParamVec(mistL, fn, "Offset Turbulence",
		[]aep.VectorKeyframe{{Time: 0, Value: []float64{820, 520}}, {Time: 4, Value: []float64{1280, 710}}}); err != nil {
		return err
	}
	if err := set(mistL, fn, "Opacity", 18.0); err != nil {
		return err
	}
	return mistL.SetBlendingMode(aep.BlendingModeAdd)
}
