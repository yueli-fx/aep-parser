// Dump the first tdsb chunk inside the first Layr's outer LIST(tdgp).
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
	root, _ := rifx.Parse(f)
	layrFT := rifx.ChunkID{'L', 'a', 'y', 'r'}
	tdgpFT := rifx.IDTdgp
	tdsbID := rifx.ChunkID{'t', 'd', 's', 'b'}

	var layr *rifx.Chunk
	var walk func(c *rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		if layr != nil { return }
		if c.IsList() && c.FormType == layrFT { layr = c; return }
		for _, ch := range c.Children { walk(ch) }
	}
	walk(root)
	if layr == nil { fmt.Println("no Layr"); return }

	var outer *rifx.Chunk
	for _, ch := range layr.Children {
		if ch.IsList() && ch.FormType == tdgpFT {
			outer = ch
			break
		}
	}
	if outer == nil { fmt.Println("no outer tdgp"); return }

	for _, ch := range outer.Children {
		if ch.ID == tdsbID {
			fmt.Printf("outer tdgp tdsb (%d B): %s\n", len(ch.Data), hex.EncodeToString(ch.Data))
			return
		}
	}
	fmt.Println("no tdsb in outer tdgp")
}
