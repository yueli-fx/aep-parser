// tmp_debug/find_layerid_refs/main.go
//
// Walk the entire chunk tree and report every chunk whose Data contains the
// bytes `00 00 00 0d` (= layer ID 13 as uint32 BE). Helps RE the "comp owns
// layer X" link — if tolerance.aep has a chunk we don't emit that references
// the user layer ID, that's the silent-drop missing piece.
//
// Skips ldta chunks (every Layr's ldta starts with its own LayerID so they'd
// all match) — we want to find OTHER chunks that reference 13.
package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/example/aep-parser/internal/rifx"
)

var target = []byte{0x00, 0x00, 0x00, 0x0d}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: find_layerid_refs <path.aep> [needle-hex]")
		os.Exit(1)
	}
	if len(os.Args) >= 3 {
		dec, err := hex.DecodeString(os.Args[2])
		if err != nil {
			panic(err)
		}
		target = dec
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer f.Close()
	root, err := rifx.Parse(f)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Searching for bytes %s in %s\n", hex.EncodeToString(target), os.Args[1])
	fmt.Println()

	path := []string{}
	var walk func(c *rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		var label string
		if c.IsList() {
			label = "LIST " + string(c.FormType[:])
		} else {
			label = string(c.ID[:])
		}
		path = append(path, label)
		defer func() { path = path[:len(path)-1] }()

		if !c.IsList() && len(c.Data) >= 4 {
			// Skip ldta chunks: every Layr's ldta starts with its own LayerID,
			// so they'd all match and drown out the signal.
			if string(c.ID[:]) == "ldta" {
				return
			}
			if idx := bytes.Index(c.Data, target); idx >= 0 {
				fmt.Printf("hit: %s @offset 0x%x in %dB\n", joinPath(path), idx, len(c.Data))
				// Print 16 bytes around the hit for context
				lo := idx - 8
				if lo < 0 {
					lo = 0
				}
				hi := idx + 12
				if hi > len(c.Data) {
					hi = len(c.Data)
				}
				fmt.Printf("  context: ...%s [match at offset %d]\n",
					hex.EncodeToString(c.Data[lo:hi]), idx-lo)
			}
		}
		for _, ch := range c.Children {
			walk(ch)
		}
	}
	walk(root)
}

func joinPath(p []string) string {
	s := ""
	for i, x := range p {
		if i > 0 {
			s += " > "
		}
		s += x
	}
	return s
}
