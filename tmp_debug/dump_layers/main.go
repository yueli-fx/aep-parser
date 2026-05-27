// tmp_debug/dump_layers/main.go
//
// V3 Phase 2 DeleteLayer RE helper. For each comp in the .aep:
//   - prints the parser's *Layer view (ID / Type / Name / ParentID /
//     TrackMatteLayerID / SourceID), in parse order
//   - prints the raw Item LIST children IDs in chunk order, so we can see
//     whether AE splices (back-shifts subsequent Layr/Ewst pairs) or
//     leaves a gap when a layer is deleted
//   - notes Ewst sibling presence per Layr (AE convention: every Layr has
//     a 0-child Ewst LIST immediately after — Q4 expects this to also be
//     deleted when its Layr is deleted)
//
// Usage:
//   go run ./tmp_debug/dump_layers <file.aep> [<file2.aep> ...]
//
// Designed to be diffed across the 4 re_delete_layer_*.aep fixtures (Q1-Q4).
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/rifx"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: dump_layers file.aep [file2.aep ...]")
		os.Exit(2)
	}
	for _, p := range os.Args[1:] {
		fmt.Printf("=== %s ===\n", p)
		if err := dumpOne(p); err != nil {
			fmt.Fprintf(os.Stderr, "  ERROR: %v\n", err)
		}
		fmt.Println()
	}
}

func dumpOne(path string) error {
	// First pass: parsed view via aep.Open (gets Layer fields).
	proj, err := aep.Open(path)
	if err != nil {
		return fmt.Errorf("aep.Open: %w", err)
	}

	// Second pass: raw chunk tree via rifx.Parse — for Item LIST child
	// order + Ewst siblings.
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("os.Open: %w", err)
	}
	defer f.Close()
	root, err := rifx.Parse(f)
	if err != nil {
		return fmt.Errorf("rifx.Parse: %w", err)
	}

	fmt.Printf("  comps=%d\n", len(proj.Compositions))
	for _, c := range proj.Compositions {
		fmt.Printf("  --- comp %q (id=%d, layers=%d) ---\n", c.Name, c.ID, len(c.Layers))

		// Parsed view.
		fmt.Println("    parsed layers (in parse order):")
		for i, l := range c.Layers {
			fmt.Printf("      [%d] id=%d type=%v name=%q parent=%d matteSrc=%d source=%d\n",
				i, l.ID, l.Type, l.Name, l.ParentID, l.TrackMatteLayerID, l.SourceID)
		}

		// Raw chunk view — find this comp's Item LIST.
		item := findCompItem(root, c.ID)
		if item == nil {
			fmt.Println("    raw Item LIST: NOT FOUND (parser found comp but Item LIST missing in chunk tree)")
			continue
		}
		fmt.Printf("    raw Item LIST children (%d chunks):\n", len(item.Children))
		for i, ch := range item.Children {
			label := string(ch.ID[:])
			if ch.IsList() {
				ft := string(ch.FormType[:])
				fmt.Printf("      [%d] LIST %s (form=%s, %d children)\n", i, label, ft, len(ch.Children))
			} else {
				fmt.Printf("      [%d] %s (%d bytes)\n", i, label, len(ch.Data))
			}
		}

		// Ewst sibling presence — pair each Layr with the chunk that
		// follows it; flag missing/non-Ewst neighbors.
		fmt.Println("    Layr / Ewst pairing (Q4):")
		ewstID := rifx.ChunkID{'E', 'w', 's', 't'}
		for i, ch := range item.Children {
			if !ch.IsList() || ch.FormType != rifx.IDLayr {
				continue
			}
			// Look at the next child.
			next := "<end of Item>"
			isEwst := false
			if i+1 < len(item.Children) {
				n := item.Children[i+1]
				if n.IsList() {
					next = "LIST " + string(n.FormType[:])
					if n.FormType == ewstID {
						isEwst = true
					}
				} else {
					next = string(n.ID[:])
				}
			}
			marker := ""
			if !isEwst {
				marker = "  <-- WARN: no Ewst after this Layr"
			}
			fmt.Printf("      Layr[child %d] next=%q%s\n", i, next, marker)
		}
	}
	return nil
}

// findCompItem walks the chunk tree for an Item LIST whose first idta
// chunk encodes id == compID. Mirrors parseProject's identification path
// just enough to locate the comp's owning Item.
func findCompItem(root *rifx.Chunk, compID uint32) *rifx.Chunk {
	var found *rifx.Chunk
	var walk func(c *rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		if found != nil {
			return
		}
		if c.IsList() && c.FormType == rifx.IDItem {
			if id := readItemID(c); id == compID {
				found = c
				return
			}
		}
		for _, ch := range c.Children {
			walk(ch)
		}
	}
	walk(root)
	return found
}

// readItemID extracts the item ID from an Item LIST's idta chunk.
// idta layout per cdta_layout.go: id is uint32 BE @0x10. Returns 0 on miss.
func readItemID(item *rifx.Chunk) uint32 {
	idta := item.FindFirst(rifx.ChunkID{'i', 'd', 't', 'a'})
	if idta == nil || len(idta.Data) < 0x14 {
		return 0
	}
	return uint32(idta.Data[0x10])<<24 | uint32(idta.Data[0x11])<<16 |
		uint32(idta.Data[0x12])<<8 | uint32(idta.Data[0x13])
}
