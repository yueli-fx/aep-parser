// tmp_debug/gen_v3_only/main.go — iter-8 starting baseline.
// variant #3 = empty ShapeLayer + AddRect (default size). Smallest case
// that triggers shape-content silent drop (variant #2 passes; #3 drops).
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

func main() {
	p := aep.NewProject(aep.TargetAE2025)
	comp, err := p.NewComposition("Main", 1920, 1080, 30, 5)
	if err != nil {
		fmt.Println("comp:", err)
		os.Exit(1)
	}
	l, err := comp.NewShapeLayer("L")
	if err != nil {
		fmt.Println("layer:", err)
		os.Exit(1)
	}
	if _, err := l.RootGroup().AddRect(); err != nil {
		fmt.Println("rect:", err)
		os.Exit(1)
	}
	out, _ := os.Create("tmp_debug/minfail_v3.aep")
	defer out.Close()
	if err := p.WriteAEP(out); err != nil {
		fmt.Println("write:", err)
		os.Exit(1)
	}
	fmt.Println("wrote tmp_debug/minfail_v3.aep")
}
