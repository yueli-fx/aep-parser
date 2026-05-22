// Print every composition's comp-level markers.
// Usage: go run ./tmp_debug/demo_compmark file.aep
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: demo_compmark file.aep")
		os.Exit(2)
	}
	p, err := aep.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, c := range p.Compositions {
		if len(c.Markers) == 0 {
			continue
		}
		fmt.Printf("Comp %q — %d comp markers:\n", c.Name, len(c.Markers))
		for i, m := range c.Markers {
			fmt.Printf("  [%d] t=%.3fs duration=%.3fs label=%d comment=%q chapter=%q\n",
				i, m.Time, m.Duration, m.Label, m.Comment, m.Chapter)
		}
	}
}
