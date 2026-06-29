// Print the raw chunk tree of every composition in an .aep file —
// used to find where unknown composition-level data (e.g. comp
// markers) live.
// Usage: go run ./tmp_debug/dump_comp file.aep comp_name_filter
package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/rifx"
)

func walk(c *rifx.Chunk, depth int, breadcrumb string) {
	indent := strings.Repeat("  ", depth)
	if c.IsList() {
		fmt.Printf("%sLIST %s (path=%s, %d children)\n", indent, c.FormType, breadcrumb, len(c.Children))
		for i, ch := range c.Children {
			walk(ch, depth+1, fmt.Sprintf("%s/%s[%d]", breadcrumb, ch.ID, i))
		}
		return
	}
	// Leaf chunk — show id, size, hex preview, ASCII preview
	preview := c.Data
	if len(preview) > 32 {
		preview = preview[:32]
	}
	asciiPrev := make([]byte, len(preview))
	for i, b := range preview {
		if b >= 0x20 && b < 0x7f {
			asciiPrev[i] = b
		} else {
			asciiPrev[i] = '.'
		}
	}
	fmt.Printf("%s%s (%d bytes) %s |%s|\n",
		indent, c.ID, len(c.Data),
		strings.TrimSpace(hex.EncodeToString(preview)),
		string(asciiPrev))
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: dump_comp file.aep [name_filter]")
		os.Exit(2)
	}
	filter := ""
	if len(os.Args) >= 3 {
		filter = os.Args[2]
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()
	root, err := rifx.Parse(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// Find comp Items by walking Fold LISTs.
	var find func(c *rifx.Chunk, path string)
	find = func(c *rifx.Chunk, path string) {
		if c.IsList() && c.FormType == (rifx.ChunkID{'I', 't', 'e', 'm'}) {
			// Check name
			var name string
			for _, ch := range c.Children {
				if ch.ID == (rifx.ChunkID{'U', 't', 'f', '8'}) {
					name = ch.Text()
					break
				}
			}
			if filter == "" || strings.Contains(name, filter) {
				fmt.Printf("=== Item name=%q (path=%s) ===\n", name, path)
				walk(c, 0, path)
			}
		}
		for i, ch := range c.Children {
			find(ch, fmt.Sprintf("%s/%s[%d]", path, ch.ID, i))
		}
	}
	find(root, "")
}
