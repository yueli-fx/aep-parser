// flightdeck/showcase/keyframe-channels/gen.go — from-scratch (no AE) showcase of
// keyframing ALL FOUR transform channels (Position / Scale / Rotation / Opacity),
// one channel per row. Each channel animates start→end over 4s with two linear
// keyframes; rendered at the MID time (t=2s) every shape sits at its INTERPOLATED
// value (not an endpoint), which is the single-still proof that per-channel
// keyframe interpolation renders:
//   - POSITION → x 300→1620, at t=2s the dot sits at the centre x≈960
//   - SCALE    → 25%→150%, at t=2s the square is ~87.5% (mid size)
//   - ROTATION → 0°→180°, at t=2s the tall bar has spun to 90° (now horizontal)
//   - OPACITY  → 100%→10%, at t=2s the square is ~55% opaque (visibly dimmed)
//
// This complements keyframes-ease (which varies the EASE CURVE on position);
// here the interpolation is plain linear but spans every transform channel.
// Writes keyframe_channels.aep next to this file. render.jsx renders t=2s.
// Run from repo root: `go run ./flightdeck/showcase/keyframe-channels`.
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

const outPath = "flightdeck/showcase/keyframe-channels/keyframe_channels.aep"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

var comp *aep.Composition

// rect builds a filled rectangle shape layer (no stroke) at static position pos.
func rect(name string, size [2]float64, col [4]float64, pos [2]float64) *aep.ShapeLayer {
	l, err := aep.NewShapeLayer(comp, name)
	must(err)
	r, err := l.RootGroup().AddRect()
	must(err)
	must(r.SetSize(size))
	f, err := l.RootGroup().AddFill()
	must(err)
	must(f.SetColor(col))
	must(l.Position().SetStaticValue(pos))
	return l
}

func main() {
	p := aep.NewProject(aep.TargetAE2020)
	c, err := aep.NewComposition(p, "KeyframeChannels", 1920, 1080, 30, 4)
	must(err)
	comp = c

	bg, err := aep.NewShapeLayer(comp, "BG")
	must(err)
	bgr, err := bg.RootGroup().AddRect()
	must(err)
	must(bgr.SetSize([2]float64{2200, 1300}))
	bgf, err := bg.RootGroup().AddFill()
	must(err)
	must(bgf.SetColor([4]float64{0.07, 0.08, 0.11, 1}))
	must(bg.Position().SetStaticValue([2]float64{960, 540}))

	// ── ROW 1 (y=210): POSITION — orange dot travels x 300→1620; mid x≈960. ──
	{
		l, err := aep.NewShapeLayer(comp, "1_Position")
		must(err)
		e, err := l.RootGroup().AddEllipse()
		must(err)
		must(e.SetSize([2]float64{96, 96}))
		f, err := l.RootGroup().AddFill()
		must(err)
		must(f.SetColor([4]float64{1.0, 0.55, 0.1, 1}))
		must(l.Position().AddKeyframeLinear(0, [2]float64{300, 210}))
		must(l.Position().AddKeyframeLinear(4, [2]float64{1620, 210}))
	}

	// ── ROW 2 (y=460): SCALE — teal square grows 25%→150%; mid ~87.5%. ───────
	{
		l := rect("2_Scale", [2]float64{170, 170}, [4]float64{0.25, 0.85, 0.95, 1}, [2]float64{960, 460})
		must(l.Scale().AddKeyframeLinear(0, [2]float64{25, 25}))
		must(l.Scale().AddKeyframeLinear(4, [2]float64{150, 150}))
	}

	// ── ROW 3 (y=690): ROTATION — amber tall bar spins 0°→180°; mid 90° (flat). ─
	{
		l := rect("3_Rotation", [2]float64{56, 200}, [4]float64{1.0, 0.78, 0.25, 1}, [2]float64{960, 690})
		must(l.Rotation().AddKeyframeLinear(0, 0))
		must(l.Rotation().AddKeyframeLinear(4, 180))
	}

	// ── ROW 4 (y=920): OPACITY — pink square fades 100%→10%; mid ~55% (dimmed). ─
	{
		l := rect("4_Opacity", [2]float64{170, 170}, [4]float64{1.0, 0.2, 0.55, 1}, [2]float64{960, 920})
		must(l.Opacity().AddKeyframeLinear(0, 100))
		must(l.Opacity().AddKeyframeLinear(4, 10))
	}

	rp, err := aep.Reopen(p)
	must(err)
	must(aep.MoveToEnd(rp.Compositions[0].LayerByName("BG")))

	out, err := os.Create(outPath)
	must(err)
	defer out.Close()
	must(rp.WriteAEP(out))
	fmt.Printf("wrote %s (%d layers)\n", outPath, len(rp.Compositions[0].Layers))
}
