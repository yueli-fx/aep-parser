// tmp_debug/swap_rect_body/main.go — iter-8 isolate test.
//
// Base = minfail_v3.aep (ours with Rect, drop). Keep ours Root Vectors Group
// + Vector Group + Vectors Group + Vector Transform/Materials Group wrappers.
// Only swap the LIST(tdgp) body that follows the `ADBE Vector Shape - Rect`
// tdmn inside ours Vectors Group, replacing it with tolerance's Rect body.
//
//   - layers.length=1 → ours wrappers byte-OK; Rect body interior has hidden
//     semantic requirement. Option B (dynamic shape emit) needs Rect body
//     bytes to be byte-exact tolerance — likely embed Rect body skeleton +
//     overwrite Size/Position/Color cdat scalars (iter-7 pattern, finer
//     granularity).
//   - layers.length=0 → ours Vector Group / Vectors Group / Vector Transform/
//     Materials wrappers also have issues. Need to isolate further before
//     committing to an embed strategy.
package main

import (
	"fmt"
	"os"

	"github.com/example/aep-parser/internal/rifx"
)

func main() {
	base := load("tmp_debug/minfail_v3.aep")
	donor := load("test_data/v2_2_shape_tolerance.aep")

	baseRectBody := findRectBody(base)
	donorRectBody := findRectBody(donor)
	if baseRectBody == nil {
		fmt.Println("ours v3 has no Rect body")
		os.Exit(1)
	}
	if donorRectBody == nil {
		fmt.Println("tolerance has no Rect body")
		os.Exit(1)
	}

	// Find the parent that holds baseRectBody and replace it.
	replaced := replaceRectBody(base, donorRectBody)
	if !replaced {
		fmt.Println("failed to splice in donor rect body")
		os.Exit(1)
	}

	out, _ := os.Create("tmp_debug/swap_rect_body.aep")
	defer out.Close()
	_ = base.Write(out)
	fmt.Println("wrote tmp_debug/swap_rect_body.aep — ours wrappers retained, Rect body LIST(tdgp) replaced from tolerance")
}

// findRectBody walks the entire chunk tree and returns the LIST(tdgp) that
// IMMEDIATELY follows a `ADBE Vector Shape - Rect` tdmn at any depth.
func findRectBody(root *rifx.Chunk) *rifx.Chunk {
	var found *rifx.Chunk
	var walk func(*rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		if found != nil {
			return
		}
		for i := 0; i < len(c.Children); i++ {
			ch := c.Children[i]
			if ch.ID == rifx.IDTdmn && trimNUL(string(ch.Data)) == "ADBE Vector Shape - Rect" && i+1 < len(c.Children) {
				next := c.Children[i+1]
				if next.IsList() && next.FormType == rifx.IDTdgp {
					found = next
					return
				}
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

func replaceRectBody(root *rifx.Chunk, donor *rifx.Chunk) bool {
	var done bool
	var walk func(*rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		if done {
			return
		}
		for i := 0; i < len(c.Children); i++ {
			ch := c.Children[i]
			if ch.ID == rifx.IDTdmn && trimNUL(string(ch.Data)) == "ADBE Vector Shape - Rect" && i+1 < len(c.Children) {
				next := c.Children[i+1]
				if next.IsList() && next.FormType == rifx.IDTdgp {
					c.Children[i+1] = donor
					done = true
					return
				}
			}
			if ch.IsList() {
				walk(ch)
				if done {
					return
				}
			}
		}
	}
	walk(root)
	return done
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

func trimNUL(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			return s[:i]
		}
	}
	return s
}
