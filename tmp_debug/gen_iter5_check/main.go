// tmp_debug/gen_iter5_check/main.go
//
// iter-5 verification: produce a 1-ShapeLayer .aep matching tolerance.aep's
// content shape (1 Rect[200,100] + 1 Fill[gray,1]) so the resulting Root
// Vectors Group subtree can be byte-diff'd against test_data/v2_2_shape_tolerance.aep.
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

func main() {
	p := aep.NewProject(aep.TargetAE2025)
	comp, err := p.NewComposition("ToleranceMain", 1920, 1080, 30, 5)
	if err != nil {
		fmt.Println("comp:", err)
		os.Exit(1)
	}
	l, err := comp.NewShapeLayer("Nested")
	if err != nil {
		fmt.Println("layer:", err)
		os.Exit(1)
	}
	rect, err := l.RootGroup().AddRect()
	if err != nil {
		fmt.Println("rect:", err)
		os.Exit(1)
	}
	if err := rect.SetSize([2]float64{200, 100}); err != nil {
		fmt.Println("size:", err)
		os.Exit(1)
	}
	fill, err := l.RootGroup().AddFill()
	if err != nil {
		fmt.Println("fill:", err)
		os.Exit(1)
	}
	if err := fill.SetColor([4]float64{0.5, 0.5, 0.5, 1}); err != nil {
		fmt.Println("color:", err)
		os.Exit(1)
	}

	out, err := os.Create("tmp_debug/iter5_check.aep")
	if err != nil {
		fmt.Println("create:", err)
		os.Exit(1)
	}
	defer out.Close()
	if err := p.WriteAEP(out); err != nil {
		fmt.Println("write:", err)
		os.Exit(1)
	}
	fmt.Println("wrote tmp_debug/iter5_check.aep")
}
