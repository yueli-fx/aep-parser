// Recursively dump every chunk in both files' first Layr, side-by-side, with byte-level diff markers.
package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/example/aep-parser/internal/rifx"
)

func main() {
	a := findFirstLayr(os.Args[1])
	b := findFirstLayr(os.Args[2])
	if a == nil || b == nil {
		fmt.Println("no Layr found in one of the files")
		return
	}
	dump("A", a, "")
	fmt.Println("---")
	dump("B", b, "")
}

func findFirstLayr(path string) *rifx.Chunk {
	f, _ := os.Open(path)
	defer f.Close()
	root, _ := rifx.Parse(f)
	layrFT := rifx.ChunkID{'L', 'a', 'y', 'r'}
	var found *rifx.Chunk
	var walk func(c *rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		if found != nil { return }
		if c.IsList() && c.FormType == layrFT { found = c; return }
		for _, ch := range c.Children { walk(ch) }
	}
	walk(root)
	return found
}

func dump(prefix string, c *rifx.Chunk, indent string) {
	id := strings.TrimRight(string(c.ID[:]), "\x00")
	if c.IsList() {
		ft := strings.TrimRight(string(c.FormType[:]), "\x00")
		fmt.Printf("%s [LIST %s] children=%d\n", indent, ft, len(c.Children))
		for _, ch := range c.Children {
			dump(prefix, ch, indent+"  ")
		}
		return
	}
	preview := ""
	if len(c.Data) <= 8 {
		preview = hex.EncodeToString(c.Data)
	} else {
		preview = hex.EncodeToString(c.Data[:8]) + "..."
	}
	fmt.Printf("%s %-4s (%d B) %s\n", indent, id, len(c.Data), preview)
}
