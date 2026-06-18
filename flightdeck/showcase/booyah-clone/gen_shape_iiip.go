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

func buildShapeIiip(p *aep.Project, orc *oracle) *aep.Composition {
	const compName = "シェイイイイプ！！！"
	comp, err := aep.NewComposition(p, compName, 1920, 1080, 30, 6)
	must(err)
	sl, err := aep.NewShapeLayer(comp, "シェイプレイヤー 1")
	must(err)
	root := sl.RootGroup()

	orig := orc.mustComp(compName).Layers[0]
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
