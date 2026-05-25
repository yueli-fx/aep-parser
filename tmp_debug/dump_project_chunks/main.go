// tmp_debug/dump_project_chunks/main.go
//
// Dump every root-level chunk (children of RIFX/Egg!) printing ID/FormType,
// recursing into LISTs. Helps RE project-level registries (LSIf / LRdr / etc)
// that may register layers separately from the Item LIST.
package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/example/aep-parser/internal/rifx"
)

func main() {
	f, _ := os.Open(os.Args[1])
	defer f.Close()
	root, _ := rifx.Parse(f)
	for i, ch := range root.Children {
		printChunk(ch, fmt.Sprintf("#%d  ", i), 0, 5)
	}
}

func printChunk(c *rifx.Chunk, prefix string, depth, maxDepth int) {
	ind := strings.Repeat("  ", depth)
	if c.IsList() {
		fmt.Printf("%s%sLIST %s (%d children)\n", ind, prefix, string(c.FormType[:]), len(c.Children))
		if depth >= maxDepth {
			return
		}
		for _, gg := range c.Children {
			printChunk(gg, "", depth+1, maxDepth)
		}
		return
	}
	snip := ""
	if len(c.Data) > 0 {
		n := len(c.Data)
		if n > 48 {
			n = 48
		}
		snip = " hex=" + hex.EncodeToString(c.Data[:n])
		if len(c.Data) > 48 {
			snip += "..."
		}
	}
	fmt.Printf("%s%s%s (%dB)%s\n", ind, prefix, string(c.ID[:]), len(c.Data), snip)
}
