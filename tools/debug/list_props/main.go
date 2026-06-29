// Dump every property on every layer matching the comp filter.
package main

import (
	"fmt"
	"os"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func main() {
	p, _ := aep.Open(os.Args[1])
	filter := ""
	if len(os.Args) >= 3 {
		filter = os.Args[2]
	}
	for _, c := range p.Compositions {
		if filter != "" && c.Name != filter {
			continue
		}
		for _, l := range c.Layers {
			fmt.Printf("=== [%s] %s (%s) ===\n", c.Name, l.Name, l.Type)
			for _, prop := range l.Properties {
				fmt.Printf("  %s  comp=%d static=%v expr=%q\n",
					prop.MatchName, prop.Components, prop.StaticValue, prop.Expression)
			}
		}
	}
}
