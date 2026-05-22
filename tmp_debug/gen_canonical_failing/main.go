// tmp_debug/gen_canonical_failing/main.go
//
// Phase 5 ship-gate debug aid — rebuilds the canonical 3-ShapeLayer aep
// that AE 2020 + 2025 both reject (Phase 5 Task 5.7 FAIL repro).
//
// Output: tmp_debug/v2_2_canonical_failing.aep
//
// Use with tmp_debug/dump_chunks/ to diff against AE-saved fixtures
// (e.g. test_data/v2_2_shape_tolerance.aep) to find structural deviations.
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

func main() {
	p := aep.NewProject(aep.TargetAE2020)
	comp, err := p.NewComposition("Main", 1920, 1080, 30, 5)
	if err != nil {
		fmt.Fprintln(os.Stderr, "NewComposition:", err)
		os.Exit(1)
	}

	a, _ := comp.NewShapeLayer("A_RectFill_Animated")
	rectA, _ := a.RootGroup().AddRect()
	_ = rectA.Size().AddKeyframeLinear(0, [2]float64{50, 50})
	_ = rectA.Size().AddKeyframeLinear(2, [2]float64{300, 200})
	fillA, _ := a.RootGroup().AddFill()
	_ = fillA.Color().AddKeyframeLinear(0, [4]float64{1, 0, 0, 1})
	_ = fillA.Color().AddKeyframeLinear(2, [4]float64{0, 0, 1, 1})
	_ = a.Position().AddKeyframeLinear(0, [2]float64{0, 0})
	_ = a.Position().AddKeyframeLinear(2, [2]float64{500, 300})

	b, _ := comp.NewShapeLayer("B_EllipseStroke_Static")
	ellB, _ := b.RootGroup().AddEllipse()
	_ = ellB.SetSize([2]float64{150, 150})
	strokeB, _ := b.RootGroup().AddStroke()
	_ = strokeB.SetColor([4]float64{0, 0, 1, 1})
	_ = strokeB.SetWidth(5)

	c, _ := comp.NewShapeLayer("C_PathFillStroke_Static")
	pathC, _ := c.RootGroup().AddPath()
	_ = pathC.SetVertices([][2]float64{{0, 0}, {100, 0}, {100, 100}, {0, 100}})
	_ = pathC.SetClosed(true)
	fillC, _ := c.RootGroup().AddFill()
	_ = fillC.SetColor([4]float64{0, 1, 0, 1})
	strokeC, _ := c.RootGroup().AddStroke()
	_ = strokeC.SetWidth(2)

	out, err := os.Create("tmp_debug/v2_2_canonical_failing.aep")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := p.WriteAEP(out); err != nil {
		fmt.Fprintln(os.Stderr, "WriteAEP:", err)
		os.Exit(1)
	}
	out.Close()
	fmt.Println("wrote tmp_debug/v2_2_canonical_failing.aep")
}
