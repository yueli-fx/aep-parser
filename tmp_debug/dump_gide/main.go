// tmp_debug/dump_gide/main.go — dump full Gide content (gdta + lhd3) per Layr.
package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/example/aep-parser/internal/rifx"
)

func main() {
	f, _ := os.Open(os.Args[1])
	defer f.Close()
	root, err := rifx.Parse(f)
	if err != nil {
		panic(err)
	}
	count := 0
	var walk func(c *rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		for _, ch := range c.Children {
			if ch.IsList() && string(ch.FormType[:]) == "Gide" {
				count++
				fmt.Printf("--- Gide #%d (%d children) ---\n", count, len(ch.Children))
				for _, g := range ch.Children {
					if g.IsList() {
						fmt.Printf("  LIST %s (%d children)\n", string(g.FormType[:]), len(g.Children))
						for _, gg := range g.Children {
							fmt.Printf("    chunk %s (%dB) hex=%s\n", string(gg.ID[:]), len(gg.Data), hex.EncodeToString(gg.Data))
						}
					} else {
						fmt.Printf("  chunk %s (%dB) hex=%s\n", string(g.ID[:]), len(g.Data), hex.EncodeToString(g.Data))
					}
				}
			}
			if ch.IsList() {
				walk(ch)
			}
		}
	}
	walk(root)
}
