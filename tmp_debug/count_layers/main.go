// tmp_debug/count_layers/main.go
//
// Open file via OUR Go parser; print layer count + type breakdown +
// first few layer IDs/names. If our parser sees N layers but AE sees N-1
// (silent drop), the drop is in AE's load-time validation, not in chunk
// structure. Otherwise our writer is producing fewer layer LISTs than we think.
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

func main() {
	for _, p := range os.Args[1:] {
		fmt.Printf("=== %s ===\n", p)
		proj, err := aep.Open(p)
		if err != nil {
			fmt.Println("  open error:", err)
			continue
		}
		fmt.Printf("  comps=%d\n", len(proj.Compositions))
		for _, c := range proj.Compositions {
			fmt.Printf("  comp %q (id=%d): layers=%d\n", c.Name, c.ID, len(c.Layers))
			for i, l := range c.Layers {
				if i > 15 {
					fmt.Printf("    ... and %d more\n", len(c.Layers)-i)
					break
				}
				fmt.Printf("    layer[%d] id=%d type=%v name=%q\n", i, l.ID, l.Type, l.Name)
			}
		}
	}
}
