// Dump idta (item metadata) hex for all comp Items in an aep.
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
	idtaID := rifx.ChunkID{'i', 'd', 't', 'a'}
	var walk func(c *rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		if c.ID == idtaID {
			fmt.Printf("idta (%d B):\n", len(c.Data))
			for i := 0; i < len(c.Data); i += 16 {
				end := i + 16
				if end > len(c.Data) {
					end = len(c.Data)
				}
				fmt.Printf("  @0x%02x: %s\n", i, hex.EncodeToString(c.Data[i:end]))
			}
			fmt.Println()
		}
		for _, ch := range c.Children {
			walk(ch)
		}
	}
	walk(root)
}
