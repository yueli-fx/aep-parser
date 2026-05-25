// tmp_debug/swap_fill_body/main.go — iter-8 second isolate test.
//
// Base = minfail_v5.aep (ours with Rect + Fill, currently drop). Keep ours
// wrappers + ours Rect body. Only swap the LIST(tdgp) following the
// `ADBE Vector Graphic - Fill` tdmn with tolerance's Fill body.
//
// Combined with swap_rect_body result (Rect body alone triggers drop), this
// tells us: when our Rect body is broken AND our Fill body is broken, does
// swapping ONLY the Fill body unblock layer? Tests whether the silent-drop
// validator gates on first-bad-shape or accumulates.
//
//   - layers.length=1 → only-Fill-swap clears drop ⇒ AE's validator only
//     gates on Fill body; our Rect body is actually OK (contradicts
//     swap_rect_body result — would mean shape-body validation is OR'd:
//     ANY swap to tolerance clears drop).
//   - layers.length=0 → our Rect body still triggers drop. Each broken shape
//     body independently triggers. iter-8 strategy needs to fix every
//     shape kind.
package main

import (
	"fmt"
	"os"

	"github.com/example/aep-parser/internal/rifx"
)

func main() {
	base := load("tmp_debug/minfail_v5.aep")
	donor := load("test_data/v2_2_shape_tolerance.aep")

	donorFillBody := findShapeBody(donor, "ADBE Vector Graphic - Fill")
	if donorFillBody == nil {
		fmt.Println("tolerance missing Fill body")
		os.Exit(1)
	}
	if !replaceShapeBody(base, "ADBE Vector Graphic - Fill", donorFillBody) {
		fmt.Println("failed to splice in donor Fill body (ours v5 missing?)")
		os.Exit(1)
	}

	out, _ := os.Create("tmp_debug/swap_fill_body.aep")
	defer out.Close()
	_ = base.Write(out)
	fmt.Println("wrote tmp_debug/swap_fill_body.aep — ours wrappers+Rect retained, Fill body replaced from tolerance")
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
