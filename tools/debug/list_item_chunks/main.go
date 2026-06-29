// Recursively list all chunks under each comp Item.
package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/rifx"
)

func main() {
	if len(os.Args) < 2 { fmt.Fprintln(os.Stderr, "usage: list_item_chunks file.aep [name_prefix]"); os.Exit(2) }
	prefix := ""
	if len(os.Args) >= 3 { prefix = os.Args[2] }
	f, _ := os.Open(os.Args[1])
	defer f.Close()
	root, err := rifx.Parse(f)
	if err != nil { panic(err) }

	var dump func(c *rifx.Chunk, depth int)
	dump = func(c *rifx.Chunk, depth int) {
		indent := strings.Repeat("  ", depth)
		tag := string(c.ID[:])
		if c.IsList() {
			form := string(c.FormType[:])
			fmt.Printf("%sLIST %s (formType=%s, %d children)\n", indent, tag, form, len(c.Children))
			for _, ch := range c.Children { dump(ch, depth+1) }
		} else {
			snippet := ""
			if len(c.Data) > 0 && len(c.Data) <= 64 {
				snippet = " hex=" + hex.EncodeToString(c.Data)
			} else if len(c.Data) > 0 {
				snippet = " head=" + hex.EncodeToString(c.Data[:32])
			}
			fmt.Printf("%schunk %s (%d B)%s\n", indent, tag, len(c.Data), snippet)
		}
	}

	var walk func(c *rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		if c.IsList() && c.FormType == (rifx.ChunkID{'I','t','e','m'}) {
			var name string
			for _, ch := range c.Children {
				if ch.ID == rifx.IDUtf8 { name = ch.Text(); break }
			}
			if name != "" && (prefix == "" || strings.HasPrefix(name, prefix)) {
				fmt.Printf("\n=== Item LIST: %q ===\n", name)
				for _, ch := range c.Children { dump(ch, 1) }
			}
		}
		for _, ch := range c.Children { walk(ch) }
	}
	walk(root)
}
