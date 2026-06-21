// flightdeck/showcase/booyah-clone/gen_shape_iiip.go — comp ① "シェイイイイプ！！！".
// One shape layer whose Root Vectors Group holds 4 animated rects (Size + Position
// keyframes, all linear) sharing ONE Fill on top. Original structure (probe): the
// 4 rects + 1 fill are flat siblings in the root group (no per-rect sub-groups).
// Values pulled from the original via the oracle; chunks built via our shape API.
package main

import (
	"fmt"

	aep "github.com/example/aep-parser/internal/aep"
)

// shapeIiipCompName is comp ①'s name, referenced by precomp layers that nest it (⑤+).
const shapeIiipCompName = "シェイイイイプ！！！"

func buildShapeIiip(p *aep.Project, orc *oracle) *aep.Composition {
	const compName = shapeIiipCompName
	// Use 30 fps (not the original's 29.97): NewComposition's cdta time-base
	// writer is inconsistent for fractional fps (writes a modern scale=1 marker
	// but legacy tpf/rate bytes), so AE would read a 30720 kf tickrate while we
	// encode 23976 — compressing every keyframe by 23976/30720. At a clean 30
	// fps the encoding is self-consistent (30720) and, because keyframe SECONDS
	// are preserved through a matching tickrate, AE evaluates them at the exact
	// original comp-times (the ~0.1% 29.97-vs-30 frame-snap is sub-pixel). The
	// NewComposition fractional-fps cdta bug is tracked separately.
	comp, err := aep.NewComposition(p, compName, 1920, 1080, 30, 6)
	must(err)
	sl, err := aep.NewShapeLayer(comp, "シェイプレイヤー 1")
	must(err)
	root := sl.RootGroup()

	orig := orc.mustComp(compName).Layers[0]

	// AE centers a new shape layer at comp-center; NewShapeLayer leaves it at (0,0).
	// The rects carry negative shape-space positions, so an un-centered layer puts
	// every rect off-screen (top-left) → blank render despite a correct DOM. Copy
	// the original layer's Transform Position onto the shape layer's own transform
	// (the embedded-template path lowerShapeLayer writes — NOT Layer.SetPosition,
	// whose materialized Position property a from-scratch shape layer lacks).
	tg := findGroup(orig.PropertyTree(), "ADBE Transform Group")
	px, py := float64(comp.Width)/2, float64(comp.Height)/2
	if pos := findProp(tg, "ADBE Position"); pos != nil && pos.StaticValue != nil {
		p := toFloats(pos.StaticValue)
		px, py = p[0], p[1]
	}
	must(sl.Transform().Position().SetStaticValue([2]float64{px, py}))

	// Match the original layer's timeline span. The rects' keyframes run to ~2.7s,
	// but the original shape layer is visible only [0, 0.901] (a glitch flash);
	// NewShapeLayer spans the whole comp, so without trimming the out-point the
	// clone renders bars past the moment the original layer has already ended.
	// Set the scene fields (buildLdtaBytes honors them) — Layer.SetOutPoint is an
	// in-place ldta edit that a not-yet-written from-scratch layer has no chunk for.
	sl.StartTime = orig.StartTime
	sl.Duration = orig.Duration

	rvg := findGroup(orig.PropertyTree(), "ADBE Root Vectors Group")
	if rvg == nil {
		panic("comp①: original has no Root Vectors Group")
	}

	nRect := 0
	for _, child := range rvg.Children {
		g, ok := child.(*aep.AEPropertyGroup)
		if !ok {
			continue
		}
		switch g.MatchName {
		case "ADBE Vector Shape - Rect":
			r, err := root.AddRect()
			must(err)
			for _, kf := range findProp(g, "ADBE Vector Rect Size").Keyframes {
				must(r.Size().AddKeyframeLinear(kf.Time, to2(kf.Value)))
			}
			for _, kf := range findProp(g, "ADBE Vector Rect Position").Keyframes {
				must(r.Position().AddKeyframeLinear(kf.Time, to2(kf.Value)))
			}
			nRect++
		case "ADBE Vector Graphic - Fill":
			f, err := root.AddFill()
			must(err)
			// raw "ADBE Vector Fill Color" = [A,R,G,B] 0-255; FillNode.SetColor
			// wants [R,G,B,A] 0-1 (pinned via tmp_debug/fill_color probe).
			c := to4(findProp(g, "ADBE Vector Fill Color").StaticValue)
			must(f.SetColor([4]float64{c[1] / 255, c[2] / 255, c[3] / 255, c[0] / 255}))
		}
	}
	fmt.Printf("  ① シェイイイイプ: %d rects + fill\n", nRect)
	return comp
}
