// flightdeck/showcase/animated-path/gen.go — from-scratch (no AE) showcase of an
// ANIMATED shape path: one closed Path stream carrying two linear keyframes that
// morph a WIDE HORIZONTAL BAR (t=0) into a TALL VERTICAL BAR (t=4). Both shapes
// share 4 vertices so AE interpolates them vertex-by-vertex; at the MID time
// (t=2s) every vertex is exactly halfway, so the path renders as a SQUARE — the
// single-still proof that path keyframes interpolate the geometry itself (not just
// transform). Scrub the timeline in AE to watch the full bar→square→bar morph.
//
// Closed path + solid fill (the silhouette is unambiguous); 2 keyframes only,
// well under the lhd3 >4-keyframe capacity-paging boundary. Writes
// animated_path.aep next to this file. render.jsx renders t=2s.
// Run from repo root: `go run ./flightdeck/showcase/animated-path`.
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

const outPath = "flightdeck/showcase/animated-path/animated_path.aep"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

// closedBar builds a 4-vertex closed rectangle path (straight edges, zero tangents)
// centred on the layer origin, spanning ±halfW × ±halfH.
func closedBar(halfW, halfH float64) aep.BezierPath {
	v := [][2]float64{{-halfW, -halfH}, {halfW, -halfH}, {halfW, halfH}, {-halfW, halfH}}
	return aep.BezierPath{
		Vertices:    v,
		InTangents:  make([][2]float64, len(v)),
		OutTangents: make([][2]float64, len(v)),
		Closed:      true,
	}
}

func main() {
	p := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(p, "AnimatedPath", 1920, 1080, 30, 4)
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

	// The morphing shape: a closed path with two keyframes.
	//   t=0  → wide horizontal bar (440×120)
	//   t=4  → tall vertical bar  (120×440)
	//   t=2  → square 280×280 (AE interpolates each vertex halfway)
	l, err := aep.NewShapeLayer(comp, "Morph")
	must(err)
	pth, err := l.RootGroup().AddPath()
	must(err)
	must(pth.Path().AddKeyframeLinear(0, closedBar(220, 60)))  // horizontal bar
	must(pth.Path().AddKeyframeLinear(4, closedBar(60, 220)))  // vertical bar
	f, err := l.RootGroup().AddFill()
	must(err)
	must(f.SetColor([4]float64{0.25, 0.85, 0.95, 1})) // teal
	must(l.Position().SetStaticValue([2]float64{960, 540}))

	rp, err := aep.Reopen(p)
	must(err)
	must(aep.MoveToEnd(rp.Compositions[0].LayerByName("BG")))

	out, err := os.Create(outPath)
	must(err)
	defer out.Close()
	must(rp.WriteAEP(out))
	fmt.Printf("wrote %s (%d layers)\n", outPath, len(rp.Compositions[0].Layers))
}
