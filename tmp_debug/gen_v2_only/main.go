// tmp_debug/gen_v2_only/main.go
//
// iter-6a focused builder: variant #2 (empty ShapeLayer, no shape kids) only.
// Writes tmp_debug/minfail_v2.aep for single-variant AE verification.
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
	if _, err := comp.NewShapeLayer("L"); err != nil {
		fmt.Println("layer:", err)
		os.Exit(1)
	}

	out, err := os.Create("tmp_debug/minfail_v2.aep")
	if err != nil {
		fmt.Println("create:", err)
		os.Exit(1)
	}
	defer out.Close()
	if err := p.WriteAEP(out); err != nil {
		fmt.Println("write:", err)
		os.Exit(1)
	}
	fmt.Println("wrote tmp_debug/minfail_v2.aep")
}
