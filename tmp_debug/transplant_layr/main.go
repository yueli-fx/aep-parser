// tmp_debug/transplant_layr/main.go
//
// Build a "transplant" .aep: tolerance.aep as base, but its first user Layr
// (the "Nested" ShapeLayer) replaced with our minfail_v2.aep's user Layr.
// AE runs the result through verify_baseline; the layers.length result
// isolates the silent-drop cause:
//
//   - layers.length=1  → problem is OUTSIDE the Layr chunk (cdta / head / item-level / ...)
//   - layers.length=0  → problem is INSIDE the Layr chunk (ldta / tdgp / Gide / ...)
//
// One AE run, decisive.
package main

import (
	"fmt"
	"os"

	"github.com/example/aep-parser/internal/rifx"
)

const (
	basePath   = "test_data/v2_2_shape_tolerance.aep"
	donorPath  = "tmp_debug/minfail_v2.aep"
	outputPath = "tmp_debug/transplant.aep"
)

func main() {
	// Load both
	baseFile, err := os.Open(basePath)
	if err != nil {
		panic(err)
	}
	baseRoot, err := rifx.Parse(baseFile)
	baseFile.Close()
	if err != nil {
		panic(err)
	}

	donorFile, err := os.Open(donorPath)
	if err != nil {
		panic(err)
	}
	donorRoot, err := rifx.Parse(donorFile)
	donorFile.Close()
	if err != nil {
		panic(err)
	}

	donorLayr := findFirstUserLayr(donorRoot)
	if donorLayr == nil {
		fmt.Println("donor has no user Layr")
		os.Exit(1)
	}

	// Replace base's first user Layr with donor's.
	if !replaceFirstUserLayr(baseRoot, donorLayr) {
		fmt.Println("base has no user Layr to replace")
		os.Exit(1)
	}

	// Save.
	out, err := os.Create(outputPath)
	if err != nil {
		panic(err)
	}
	defer out.Close()
	if err := baseRoot.Write(out); err != nil {
		panic(err)
	}
	fmt.Printf("wrote %s — base=tolerance.aep with first user Layr replaced by minfail_v2's Layr\n", outputPath)
}

func findFirstUserLayr(root *rifx.Chunk) *rifx.Chunk {
	var found *rifx.Chunk
	var walk func(*rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		if found != nil {
			return
		}
		for _, ch := range c.Children {
			if ch.IsList() && ch.FormType == rifx.IDLayr {
				found = ch
				return
			}
			if ch.IsList() {
				walk(ch)
				if found != nil {
					return
				}
			}
		}
	}
	walk(root)
	return found
}

func replaceFirstUserLayr(root *rifx.Chunk, donor *rifx.Chunk) bool {
	var replaced bool
	var walk func(*rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		if replaced {
			return
		}
		for i, ch := range c.Children {
			if ch.IsList() && ch.FormType == rifx.IDLayr {
				c.Children[i] = donor
				replaced = true
				return
			}
			if ch.IsList() {
				walk(ch)
				if replaced {
					return
				}
			}
		}
	}
	walk(root)
	return replaced
}
