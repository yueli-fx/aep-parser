// tmp_debug/extract_transform_group/main.go
//
// Extract tolerance.aep's user Layr "Nested" Transform Group body chunk
// (the LIST(tdgp) with 15 children: tdsb + tdsn + 6 stream tdmn-LIST pairs +
// Group End) as a stand-alone binary blob. Embedded into the V2.2 builder
// (iter-7) as the AE-acceptable transform schema; lower_layer.go deep-clones
// and overwrites Position_0/_1 cdat values with runtime user settings.
package main

import (
	"fmt"
	"os"

	"github.com/example/aep-parser/internal/rifx"
)

const outputPath = "internal/aep/templates/v2_2_transform_group_body.bin"

func main() {
	f, err := os.Open("test_data/v2_2_shape_tolerance.aep")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	root, err := rifx.Parse(f)
	if err != nil {
		panic(err)
	}

	// Find first user Layr's Transform Group body LIST(tdgp).
	var found *rifx.Chunk
	var walk func(*rifx.Chunk, bool)
	walk = func(c *rifx.Chunk, inLayr bool) {
		if found != nil {
			return
		}
		isLayr := c.IsList() && c.FormType == rifx.IDLayr
		isTransformChild := inLayr
		for i, ch := range c.Children {
			if ch.ID == rifx.IDTdmn && trimNUL(string(ch.Data)) == "ADBE Transform Group" {
				if i+1 < len(c.Children) && c.Children[i+1].IsList() && c.Children[i+1].FormType == rifx.IDTdgp {
					found = c.Children[i+1]
					return
				}
			}
			if ch.IsList() {
				walk(ch, isLayr || isTransformChild)
				if found != nil {
					return
				}
			}
		}
	}
	walk(root, false)
	if found == nil {
		panic("Transform Group body not found")
	}
	fmt.Printf("Transform Group LIST(tdgp): %d children\n", len(found.Children))

	out, err := os.Create(outputPath)
	if err != nil {
		panic(err)
	}
	defer out.Close()
	if err := found.Write(out); err != nil {
		panic(err)
	}
	stat, _ := out.Stat()
	fmt.Printf("wrote %s (%d bytes)\n", outputPath, stat.Size())
}

func trimNUL(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			return s[:i]
		}
	}
	return s
}
