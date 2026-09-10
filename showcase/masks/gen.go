// showcase/masks/gen.go — from-scratch (no AE) showcase of the
// AddMask capability (mask clipping). Builds a 2×2 grid of solid-colour shape
// squares on a dark BG, then drops one mask of a different outline on each
// square so AE clips the square down to the mask shape: a circle, a triangle,
// a 5-point star, and an INVERTED square (punched hole). The user should see
// each coloured square reduced to its mask silhouette — flat fill inside the
// mask outline, transparent (BG showing) outside.
//
// AddMask needs a *parsed* layer, so all source squares are built first, then
// Reopen() upgrades them, then masks are spliced on the reopened comp.
//
// Coordinate units (RE'd, see internal/serializer/mutate_mask_add.go): for
// SOURCE-LESS layers (shape / text) the mask path is read back 1:1 as
// LAYER-LOCAL pixels — origin at the layer anchor (= the shape group origin,
// [0,0]), which Position then places at the cell centre. So a mask vertex at
// layer-local [0,0] lands exactly on the square's centre. Each square is a
// Rect drawn at the group origin (spanning ±size/2), and every mask path below
// is authored in that same centred layer-local space.
//
// AddMask + Mask.SetInverted are double-version ship-gate verified
// (add_mask_shipgate_test.go); this 4-in-one COMBINATION is rendered +
// eyeballed via render.jsx (delivery contract red line 4 — see INDEX.md for
// the per-cell table the render is checked against).
//
// Run from the repo root: `go run ./showcase/masks`.
package main

import (
	"fmt"
	"math"
	"os"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

const outPath = "showcase/masks/masks.aep"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

// 2×2 grid cell centres in a 1920×1080 frame.
var colX = []float64{620, 1300}
var rowY = []float64{330, 750}

func at(col, row int) [2]float64 { return [2]float64{colX[col], rowY[row]} }

var (
	teal  = [4]float64{0.30, 0.85, 0.80, 1}
	amber = [4]float64{1, 0.78, 0.25, 1}
	pink  = [4]float64{1, 0.45, 0.62, 1}
	green = [4]float64{0.55, 0.86, 0.45, 1}
)

// The square each mask clips. side = full edge length; mask paths are authored
// in layer-local space centred at [0,0], so vertices range over ±side/2.
const side = 320

// circlePath — 4-vertex bezier circle of radius r centred at [0,0]. The
// magic-number tangent length (kappa·r, kappa≈0.5523) makes 4 cubic segments
// approximate a circle. Tangents are per-vertex offsets from the anchor.
func circlePath(r float64) aep.BezierPath {
	const k = 0.5522847498 // 4/3 · tan(π/8)
	t := k * r
	// vertices clockwise from the top: top, right, bottom, left.
	v := [][2]float64{{0, -r}, {r, 0}, {0, r}, {-r, 0}}
	// Tangents tangent to the circle: out points along the clockwise travel
	// direction, in is its mirror (in = -out) for a smooth closed circle.
	// (The previous values had out/in reversed → cusps instead of a circle.)
	out := [][2]float64{{t, 0}, {0, t}, {-t, 0}, {0, -t}}
	in := [][2]float64{{-t, 0}, {0, -t}, {t, 0}, {0, t}}
	return aep.BezierPath{Vertices: v, InTangents: in, OutTangents: out, Closed: true}
}

// trianglePath — upright equilateral-ish triangle inscribed in ±r, zero tangents.
func trianglePath(r float64) aep.BezierPath {
	v := [][2]float64{
		{0, -r},               // apex (top)
		{r * 0.866, r * 0.5},  // bottom-right
		{-r * 0.866, r * 0.5}, // bottom-left
	}
	return aep.BezierPath{Vertices: v, Closed: true}
}

// starPath — n-point star alternating outer/inner radius, zero tangents.
func starPath(points int, outer, inner float64) aep.BezierPath {
	var v [][2]float64
	for i := 0; i < points*2; i++ {
		r := outer
		if i%2 == 1 {
			r = inner
		}
		// start at the top (−90°), step half a point each vertex.
		ang := -math.Pi/2 + float64(i)*math.Pi/float64(points)
		v = append(v, [2]float64{r * math.Cos(ang), r * math.Sin(ang)})
	}
	return aep.BezierPath{Vertices: v, Closed: true}
}

// squarePath — axis-aligned square of half-extent h centred at [0,0].
func squarePath(h float64) aep.BezierPath {
	return aep.BezierPath{
		Vertices: [][2]float64{{-h, -h}, {h, -h}, {h, h}, {-h, h}},
		Closed:   true,
	}
}

type cell struct {
	name     string
	col, row int
	color    [4]float64
	path     aep.BezierPath
	inverted bool
}

func main() {
	p := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(p, "MasksShowcase", 1920, 1080, 30, 3)
	must(err)

	// Dark background.
	bg, err := aep.NewShapeLayer(comp, "BG")
	must(err)
	br, err := bg.RootGroup().AddRect()
	must(err)
	must(br.SetSize([2]float64{2200, 1300}))
	bf, err := bg.RootGroup().AddFill()
	must(err)
	must(bf.SetColor([4]float64{0.07, 0.08, 0.11, 1}))
	must(bg.Position().SetStaticValue([2]float64{960, 540}))

	cells := []cell{
		// (0,0) circle mask — teal square clipped to a disc.
		{name: "01_Circle", col: 0, row: 0, color: teal, path: circlePath(side / 2)},
		// (1,0) triangle mask — amber square clipped to a triangle.
		{name: "02_Triangle", col: 1, row: 0, color: amber, path: trianglePath(side / 2)},
		// (0,1) star mask — pink square clipped to a 5-point star.
		{name: "03_Star", col: 0, row: 1, color: pink, path: starPath(5, side/2, side*0.22)},
		// (1,1) inverted square mask — green square with a square hole punched
		// out of its centre (inverted: keep OUTSIDE the mask, drop inside).
		{name: "04_InvertedHole", col: 1, row: 1, color: green, path: squarePath(side * 0.28), inverted: true},
	}

	// Build all coloured squares first (built layers, not yet parsed).
	for _, c := range cells {
		l, err := aep.NewShapeLayer(comp, c.name)
		must(err)
		g := l.RootGroup()
		r, err := g.AddRect()
		must(err)
		must(r.SetSize([2]float64{side, side}))
		f, err := g.AddFill()
		must(err)
		must(f.SetColor(c.color))
		must(l.Position().SetStaticValue(at(c.col, c.row)))
	}

	// Reopen so every layer is parsed → AddMask works.
	rp, err := aep.Reopen(p)
	must(err)
	rc := rp.Compositions[0]

	for _, c := range cells {
		l := rc.LayerByName(c.name)
		if l == nil {
			panic("layer not found: " + c.name)
		}
		m, err := aep.AddMask(l, c.name+"_mask", c.path)
		must(err)
		if c.inverted {
			must(m.SetInverted(true))
		}
	}

	// BG must render behind everything.
	must(aep.MoveToEnd(rc.LayerByName("BG")))

	out, err := os.Create(outPath)
	must(err)
	defer out.Close()
	must(rp.WriteAEP(out))
	fmt.Printf("wrote %s (%d layers)\n", outPath, len(rc.Layers))
}
