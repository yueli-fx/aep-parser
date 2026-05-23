// Dump cdta and diff side-by-side.
package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/example/aep-parser/internal/rifx"
)

func main() {
	a := readCdta(os.Args[1])
	b := readCdta(os.Args[2])
	fmt.Fprintf(os.Stderr, "a=%d b=%d bytes\n", len(a), len(b))
	for i := 0; i < len(a) && i < len(b); i += 16 {
		end := i + 16
		if end > len(a) { end = len(a) }
		if end > len(b) { end = len(b) }
		aHex := hex.EncodeToString(a[i:end])
		bHex := hex.EncodeToString(b[i:end])
		marker := "  "
		if aHex != bHex {
			marker = "**"
		}
		fmt.Printf("%s @0x%02x  A: %-32s   B: %-32s\n", marker, i, aHex, bHex)
	}
}

func readCdta(path string) []byte {
	f, _ := os.Open(path)
	defer f.Close()
	root, _ := rifx.Parse(f)
	cdtaID := rifx.ChunkID{'c', 'd', 't', 'a'}
	var found *rifx.Chunk
	var walk func(c *rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		if found != nil { return }
		if c.ID == cdtaID { found = c; return }
		for _, ch := range c.Children { walk(ch) }
	}
	walk(root)
	if found == nil { return nil }
	return found.Data
}
