// tmp_debug/gen_v5_only/main.go — variant #5 = +AddRect +SetSize +AddFill.
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

func main() {
	p := aep.NewProject(aep.TargetAE2025)
	comp, _ := p.NewComposition("Main", 1920, 1080, 30, 5)
	l, _ := comp.NewShapeLayer("L")
	rect, _ := l.RootGroup().AddRect()
	_ = rect.SetSize([2]float64{200, 200})
	fill, _ := l.RootGroup().AddFill()
	_ = fill.SetColor([4]float64{1, 0, 0, 1})
	out, _ := os.Create("tmp_debug/minfail_v5.aep")
	defer out.Close()
	if err := p.WriteAEP(out); err != nil {
		fmt.Println("write:", err)
		os.Exit(1)
	}
	fmt.Println("wrote tmp_debug/minfail_v5.aep")
}
