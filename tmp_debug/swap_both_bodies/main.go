// tmp_debug/swap_both_bodies/main.go — iter-8 third isolate test.
//
// Base = minfail_v5.aep (ours with Rect + Fill, drop). Swap BOTH Rect body
// AND Fill body with tolerance's versions. Keeps ours wrappers.
//
//   - layers.length=1 → confirmed: shape bodies are the only silent-drop
//     trigger across shape kinds. iter-8 strategy = embed-tolerance-body
//     skeleton for each shape kind (Rect, Ellipse, Path, Fill, Stroke).
//   - layers.length=0 → there's another trigger beyond individual shape bodies
//     (e.g. inter-shape ordering / Vectors Group child count constraint /
//     accumulated effect not fixed by per-body swap).
package main

import (
	"fmt"
	"os"

	"github.com/example/aep-parser/internal/rifx"
)

func main() {
	base := load("tmp_debug/minfail_v5.aep")
	donor := load("test_data/v2_2_shape_tolerance.aep")

	donorRect := findShapeBody(donor, "ADBE Vector Shape - Rect")
	donorFill := findShapeBody(donor, "ADBE Vector Graphic - Fill")
	if donorRect == nil || donorFill == nil {
		fmt.Println("tolerance missing Rect or Fill body")
		os.Exit(1)
	}
	if !replaceShapeBody(base, "ADBE Vector Shape - Rect", donorRect) {
		fmt.Println("failed to splice Rect body")
		os.Exit(1)
	}
	if !replaceShapeBody(base, "ADBE Vector Graphic - Fill", donorFill) {
		fmt.Println("failed to splice Fill body")
		os.Exit(1)
	}

	out, _ := os.Create("tmp_debug/swap_both_bodies.aep")
	defer out.Close()
	_ = base.Write(out)
	fmt.Println("wrote tmp_debug/swap_both_bodies.aep — ours wrappers retained, Rect + Fill bodies replaced from tolerance")
}

func findShapeBody(root *rifx.Chunk, matchName string) *rifx.Chunk {
	var found *rifx.Chunk
	var walk func(*rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		if found != nil {
			return
		}
		for i := 0; i < len(c.Children); i++ {
			ch := c.Children[i]
			if ch.ID == rifx.IDTdmn && trimNUL(string(ch.Data)) == matchName && i+1 < len(c.Children) {
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

func replaceShapeBody(root *rifx.Chunk, matchName string, donor *rifx.Chunk) bool {
	var done bool
	var walk func(*rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		if done {
			return
		}
		for i := 0; i < len(c.Children); i++ {
			ch := c.Children[i]
			if ch.ID == rifx.IDTdmn && trimNUL(string(ch.Data)) == matchName && i+1 < len(c.Children) {
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
