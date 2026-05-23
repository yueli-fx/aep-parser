// Dump head chunk hex.
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
	headID := rifx.ChunkID{'h', 'e', 'a', 'd'}
	h := root.FindFirst(headID)
	if h == nil {
		fmt.Println("no head chunk")
		return
	}
	fmt.Printf("head (%d B):\n", len(h.Data))
	for i := 0; i < len(h.Data); i += 4 {
		end := i + 4
		if end > len(h.Data) {
			end = len(h.Data)
		}
		fmt.Printf("  @0x%02x: %s\n", i, hex.EncodeToString(h.Data[i:end]))
	}
}
