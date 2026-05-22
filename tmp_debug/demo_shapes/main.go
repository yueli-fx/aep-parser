// Verify Shape primitive structuring against re_shapes.aep.
// Usage: go run ./tmp_debug/demo_shapes test_data/re_shapes.aep
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

func dump(label string, p *aep.Property) {
	if p == nil {
		fmt.Printf("    %s: <nil>\n", label)
		return
	}
	fmt.Printf("    %s: static=%v keyframes=%d\n", label, p.StaticValue, len(p.Keyframes))
}

func main() {
	pr, err := aep.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, c := range pr.Compositions {
		if c.Name != "RE_SHAPES" {
			continue
		}
		for _, l := range c.Layers {
			if !l.IsShapeLayer {
				continue
			}
			fmt.Printf("Shape layer: %s — %d primitives\n", l.Name, len(l.ShapePrimitives))
			for i, p := range l.ShapePrimitives {
				fmt.Printf("  [%d] kind=%s group=%q\n", i, p.Kind, p.GroupName)
				dump("Size", p.Size)
				dump("Position", p.Position)
				dump("Roundness", p.Roundness)
				dump("StarType", p.StarType)
				dump("Points", p.Points)
				dump("Rotation", p.Rotation)
				dump("InnerRadius", p.InnerRadius)
				dump("OuterRadius", p.OuterRadius)
				dump("InnerRoundness", p.InnerRoundness)
				dump("OuterRoundness", p.OuterRoundness)
			}
		}
	}
}
