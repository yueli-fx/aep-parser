package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

func main() {
	p, err := aep.Open("tmp_debug/v2_2_canonical_failing.aep")
	if err != nil {
		fmt.Println("Open:", err)
		os.Exit(1)
	}
	fmt.Println("comps:", len(p.Compositions))
	for _, c := range p.Compositions {
		fmt.Printf("  comp %q layers=%d\n", c.Name, len(c.Layers))
		for i, l := range c.Layers {
			fmt.Printf("    [%d] %q type=%v IsShape=%v\n", i, l.Name, l.Type, l.IsShapeLayer)
			if l.IsShapeLayer {
				s := aep.WrapShapeLayer(l)
				fmt.Printf("        RootGroup children=%d\n", len(s.RootGroup().Children))
			}
		}
	}
}
