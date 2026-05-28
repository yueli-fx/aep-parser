// tmp_debug/probe_ldta/main.go
//
// Dump first 32 bytes of each Layr's ldta + the Utf8 name. Quick way
// to confirm clone vs source ldta bytes differ only at @0x00 (layer
// ID).
//
// Usage: go run ./tmp_debug/probe_ldta <file.aep>
package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/example/aep-parser/internal/rifx"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: probe_ldta <file.aep>")
		os.Exit(2)
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
	var item *rifx.Chunk
	var walk func(c *rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		if item != nil {
			return
		}
		if c.IsList() && c.FormType == rifx.IDItem {
			for _, k := range c.Children {
				if k.IsList() && k.FormType == rifx.IDLayr {
					item = c
					return
				}
			}
		}
		for _, k := range c.Children {
			walk(k)
		}
	}
	walk(root)
	if item == nil {
		fmt.Fprintln(os.Stderr, "no Item with Layr children")
		os.Exit(1)
	}
	idx := 0
	for _, c := range item.Children {
		if !c.IsList() || c.FormType != rifx.IDLayr {
			continue
		}
		var ldta, utf8 *rifx.Chunk
		for _, k := range c.Children {
			if k.ID == rifx.IDLdta {
				ldta = k
			}
			if k.ID == rifx.IDUtf8 {
				utf8 = k
			}
		}
		name := ""
		if utf8 != nil {
			name = utf8.Text()
		}
		fmt.Printf("Layr[%d] name=%q\n", idx, name)
		if ldta != nil {
			fmt.Printf("  ldta len=%d  first 32B = %s\n", len(ldta.Data), hex.EncodeToString(ldta.Data[:32]))
			fmt.Printf("              @0x60..0x7F = %s\n", hex.EncodeToString(ldta.Data[0x60:0x80]))
			fmt.Printf("              @0x80..0x9F = %s\n", hex.EncodeToString(ldta.Data[0x80:0xa0]))
		}
		idx++
	}
}
