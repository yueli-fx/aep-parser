// tmp_debug/transplant_tdgp/main.go
//
// Finer-grain transplant: tolerance.aep as base, only the user Layr's
// LIST(tdgp) (the outer property tree, 3rd of 4 Layr children) replaced
// with ours. Keeps tolerance's Utf8 layer name "Nested".
//
//   - layers.length=1 → problem isolated to Utf8 layer name (unlikely but testable)
//   - layers.length=0 → problem is in our LIST(tdgp) content
package main

import (
	"fmt"
	"os"

	"github.com/example/aep-parser/internal/rifx"
)

func main() {
	baseFile, _ := os.Open("test_data/v2_2_shape_tolerance.aep")
	baseRoot, _ := rifx.Parse(baseFile)
	baseFile.Close()

	donorFile, _ := os.Open("tmp_debug/minfail_v2.aep")
	donorRoot, _ := rifx.Parse(donorFile)
	donorFile.Close()

	donorLayr := findFirstUserLayr(donorRoot)
	baseLayr := findFirstUserLayr(baseRoot)
	if donorLayr == nil || baseLayr == nil {
		fmt.Println("missing Layr")
		os.Exit(1)
	}

	// Find LIST(tdgp) (3rd child) in donor + base.
	var donorTdgp *rifx.Chunk
	for _, ch := range donorLayr.Children {
		if ch.IsList() && ch.FormType == rifx.IDTdgp {
			donorTdgp = ch
			break
		}
	}
	if donorTdgp == nil {
		fmt.Println("donor has no LIST(tdgp)")
		os.Exit(1)
	}
	for i, ch := range baseLayr.Children {
		if ch.IsList() && ch.FormType == rifx.IDTdgp {
			baseLayr.Children[i] = donorTdgp
			break
		}
	}

	out, _ := os.Create("tmp_debug/transplant_tdgp.aep")
	defer out.Close()
	if err := baseRoot.Write(out); err != nil {
		panic(err)
	}
	fmt.Println("wrote tmp_debug/transplant_tdgp.aep — base=tolerance, Layr's LIST(tdgp) replaced with ours")
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
