// tmp_debug/swap_rvg/main.go — iter-8 transplant test.
//
// Base = minfail_v3.aep (ours with Rect, currently drop). Swap in
// tolerance's Root Vectors Group body. Run AE verify_baseline.
//
//   - layers.length=1 → shape-content silent drop is INSIDE our Root Vectors
//     Group / Rect body construction. iter-8 follows iter-7 strategy: embed
//     tolerance Root Vectors Group bytes as boilerplate.
//   - layers.length=0 → trigger is OUTSIDE Root Vectors Group (5-level nesting
//     wrapper / Item-level chunks shifted by Rect addition / etc).
package main

import (
	"fmt"
	"os"

	"github.com/example/aep-parser/internal/rifx"
)

func main() {
	base := load("tmp_debug/minfail_v3.aep")
	donor := load("test_data/v2_2_shape_tolerance.aep")

	baseLayr := findFirstUserLayr(base)
	donorLayr := findFirstUserLayr(donor)
	baseTdgp := findOuterTdgp(baseLayr)
	donorTdgp := findOuterTdgp(donorLayr)

	// Find tolerance's Root Vectors Group body.
	var donorRvgBody *rifx.Chunk
	for i, ch := range donorTdgp.Children {
		if ch.ID == rifx.IDTdmn && trimNUL(string(ch.Data)) == "ADBE Root Vectors Group" && i+1 < len(donorTdgp.Children) {
			donorRvgBody = donorTdgp.Children[i+1]
			break
		}
	}
	if donorRvgBody == nil {
		fmt.Println("tolerance missing Root Vectors Group")
		os.Exit(1)
	}

	// Replace ours Root Vectors Group body.
	for i, ch := range baseTdgp.Children {
		if ch.ID == rifx.IDTdmn && trimNUL(string(ch.Data)) == "ADBE Root Vectors Group" && i+1 < len(baseTdgp.Children) {
			baseTdgp.Children[i+1] = donorRvgBody
			break
		}
	}

	out, _ := os.Create("tmp_debug/swap_rvg.aep")
	defer out.Close()
	_ = base.Write(out)
	fmt.Println("wrote tmp_debug/swap_rvg.aep — base=ours v3 (with Rect), Root Vectors Group from tolerance")
}

func load(path string) *rifx.Chunk {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	r, err := rifx.Parse(f)
	if err != nil {
		panic(err)
	}
	return r
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

func findOuterTdgp(layr *rifx.Chunk) *rifx.Chunk {
	for _, ch := range layr.Children {
		if ch.IsList() && ch.FormType == rifx.IDTdgp {
			return ch
		}
	}
	return nil
}

func trimNUL(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			return s[:i]
		}
	}
	return s
}
