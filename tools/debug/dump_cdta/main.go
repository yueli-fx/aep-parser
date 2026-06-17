// Dump cdta hex for every named composition; useful for diffing
// fixture variants that differ only by a flag.
// Usage: go run ./tmp_debug/dump_cdta test_data/re_batch.aep [name_prefix]
package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/example/aep-parser/internal/rifx"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: dump_cdta file.aep [name_prefix]")
		os.Exit(2)
	}
	prefix := ""
	if len(os.Args) >= 3 {
		prefix = os.Args[2]
	}
	f, _ := os.Open(os.Args[1])
	defer f.Close()
	root, err := rifx.Parse(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var walk func(c *rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		if c.IsList() && c.FormType == (rifx.ChunkID{'I', 't', 'e', 'm'}) {
			var name string
			for _, ch := range c.Children {
				if ch.ID == rifx.IDUtf8 {
					name = ch.Text()
					break
				}
			}
			if name != "" && (prefix == "" || strings.HasPrefix(name, prefix)) {
				if cdta := c.FindFirst(rifx.IDCdta); cdta != nil {
					fmt.Printf("=== %q (%d bytes cdta) ===\n", name, len(cdta.Data))
					fmt.Println(hex.Dump(cdta.Data))
				}
			}
		}
		for _, ch := range c.Children {
			walk(ch)
		}
	}
	walk(root)
}
