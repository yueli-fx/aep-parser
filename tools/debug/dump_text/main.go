// Dump the btds payload for each text layer in an .aep file.
// Usage: go run ./tmp_debug/dump_text path/to/file.aep [layer_name_filter]
package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	aep "github.com/example/aep-parser/internal/aep"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: dump_text file.aep [layer_filter]")
		os.Exit(2)
	}
	filter := ""
	if len(os.Args) >= 3 {
		filter = os.Args[2]
	}
	p, err := aep.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "open:", err)
		os.Exit(1)
	}
	for _, c := range p.Compositions {
		fmt.Fprintf(os.Stderr, "comp=%q layers=%d\n", c.Name, len(c.Layers))
		for _, l := range c.Layers {
			if l.TextSourceRaw == nil {
				continue
			}
			if filter != "" && !strings.Contains(l.Name, filter) {
				continue
			}
			fmt.Printf("=== comp=%q layer=%q (%d bytes) ===\n", c.Name, l.Name, len(l.TextSourceRaw))
			fmt.Println(hex.Dump(l.TextSourceRaw))
		}
	}
}
