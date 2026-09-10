// showcase/transform-values/gen.go — from-scratch (no AE) showcase of
// the Layer Transform value setters. Builds a 3×2 grid of the SAME source token
// (a teal square with an amber stroke) on a dark BG, then sets a DIFFERENT Layr
// transform channel on each cell to its own static value. Cell 1 is the
// untouched reference so the before/after of every channel is one glance apart.
//
// The five persisted Layr transform channels (Anchor / Position / Scale /
// Rotation / Opacity) are double-version ship-gate verified — coverage.md
// 子项⑥ `TestV2_2_XfKf_*` reads them back in AE at user units (scale 150/200,
// rot 90, opacity 50). This grid is the COMBINATION, rendered + eyeballed via
// render.jsx (delivery contract red line 4).
//
// UNITS (the load-bearing honesty point — the shape-layer transform lowering
// stores these as PERCENT/DEGREES, not the "normalized 1.0" the Layer.Set*
// convenience-wrapper doc comments imply; we drive the raw property streams,
// which is exactly what the ship-gate exercises):
//   - Position / Anchor : layer pixels ([x,y]).
//   - Scale             : PERCENT ([200,200] = 200%; on-disk ÷100 → 2.0).
//   - Rotation          : DEGREES (45 = 45°).
//   - Opacity           : PERCENT (40 = 40%; on-disk ÷100 → 0.4).
//
// Run from the repo root: `go run ./showcase/transform-values`.
package main

import (
	"fmt"
	"os"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

const outPath = "showcase/transform-values/transform_values.aep"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

// 3 columns × 2 rows in a 1920×1080 frame.
var colX = []float64{420, 960, 1500}
var rowY = []float64{330, 760}

func at(col, row int) [2]float64 { return [2]float64{colX[col], rowY[row]} }

var (
	teal  = [4]float64{0.30, 0.85, 0.80, 1}
	amber = [4]float64{1, 0.78, 0.25, 1}
)

// xform is the transform tweak for one cell. Exactly one channel per cell is
// non-default so the rendered grid reads as a single-variable before/after.
// A nil pointer means "leave at default" (default scale 100%, rot 0, opacity
// 100%, position = cell center).
type cell struct {
	name     string
	col, row int
	// non-default channel (nil = default):
	scale    *[2]float64 // percent
	rotation *float64    // degrees
	opacity  *float64    // percent
	// posOffset shifts the layer off its cell center (pixels). Used for the
	// MOVE cell to make a position change visible against the grid.
	posOffset [2]float64
}

func f64(v float64) *float64       { return &v }
func vec(v [2]float64) *[2]float64 { return &v }

func main() {
	p := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(p, "TransformValuesShowcase", 1920, 1080, 30, 3)
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
		// (0,0) untouched reference.
		{name: "1_Source", col: 0, row: 0},
		// (1,0) Scale 200% — token twice the size.
		{name: "2_Scale200", col: 1, row: 0, scale: vec([2]float64{200, 200})},
		// (2,0) Scale 50% — token half the size.
		{name: "3_Scale50", col: 2, row: 0, scale: vec([2]float64{50, 50})},
		// (0,1) Rotation 45° — token tilted to a diamond.
		{name: "4_Rotate45", col: 0, row: 1, rotation: f64(45)},
		// (1,1) Opacity 40% — token faded toward the dark BG.
		{name: "5_Opacity40", col: 1, row: 1, opacity: f64(40)},
		// (2,1) Position shifted +160px right / -120px up from cell center.
		{name: "6_MovePos", col: 2, row: 1, posOffset: [2]float64{160, -120}},
	}

	for _, c := range cells {
		l, err := aep.NewShapeLayer(comp, c.name)
		must(err)
		g := l.RootGroup()
		r, err := g.AddRect()
		must(err)
		must(r.SetSize([2]float64{180, 180}))
		f, err := g.AddFill()
		must(err)
		must(f.SetColor(teal))
		s, err := g.AddStroke()
		must(err)
		must(s.SetColor(amber))
		must(s.SetWidth(16))

		// Position (cell center + optional offset for the MOVE cell).
		center := at(c.col, c.row)
		pos := [2]float64{center[0] + c.posOffset[0], center[1] + c.posOffset[1]}
		must(l.Position().SetStaticValue(pos))

		// Apply the one non-default transform channel for this cell.
		if c.scale != nil {
			must(l.Scale().SetStaticValue(*c.scale))
		}
		if c.rotation != nil {
			must(l.Rotation().SetStaticValue(*c.rotation))
		}
		if c.opacity != nil {
			must(l.Opacity().SetStaticValue(*c.opacity))
		}
	}

	// BG was created first → top layer; sink it behind everything. MoveToEnd
	// needs a parsed layer, hence Reopen.
	rp, err := aep.Reopen(p)
	must(err)
	rc := rp.Compositions[0]
	must(aep.MoveToEnd(rc.LayerByName("BG")))

	out, err := os.Create(outPath)
	must(err)
	defer out.Close()
	must(rp.WriteAEP(out))
	fmt.Printf("wrote %s (%d layers)\n", outPath, len(rc.Layers))
}
