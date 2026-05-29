// tmp_debug/extract_shape_bodies/main.go — iter-8 extraction.
//
// Pulls the LIST(tdgp) body for each user-visible shape kind out of
// tolerance.aep into per-kind binary blobs under internal/aep/templates/:
//
//   v2_2_shape_rect_body.bin   — body following `ADBE Vector Shape - Rect` tdmn
//   v2_2_shape_fill_body.bin   — body following `ADBE Vector Graphic - Fill` tdmn
//
// V2.2 alpha (iter-8) only supports the two shape kinds tolerance.aep
// happens to contain. Ellipse / Path / Stroke embed bytes need a separate
// AE-saved fixture, deferred to V2.2.1.
package main

import (
	"fmt"
	"os"

	"github.com/example/aep-parser/internal/rifx"
)

type extraction struct {
	srcPath   string
	matchName string
	outPath   string
}

var extractions = []extraction{
	{"test_data/v2_2_shape_tolerance.aep", "ADBE Vector Shape - Rect", "internal/aep/templates/v2_2_shape_rect_body.bin"},
	{"test_data/v2_2_shape_tolerance.aep", "ADBE Vector Graphic - Fill", "internal/aep/templates/v2_2_shape_fill_body.bin"},
	{"test_data/v2_2_shape_ellipse_tolerance.aep", "ADBE Vector Shape - Ellipse", "internal/aep/templates/v2_2_shape_ellipse_body.bin"},
}

func main() {
	for _, ex := range extractions {
		f, err := os.Open(ex.srcPath)
		if err != nil {
			panic(err)
		}
		root, err := rifx.Parse(f)
		f.Close()
		if err != nil {
			panic(err)
		}
		body := findShapeBody(root, ex.matchName)
		if body == nil {
			fmt.Printf("MISS: %s not found in %s\n", ex.matchName, ex.srcPath)
			continue
		}
		out, err := os.Create(ex.outPath)
		if err != nil {
			panic(err)
		}
		if err := body.Write(out); err != nil {
			panic(err)
		}
		stat, _ := out.Stat()
		out.Close()
		fmt.Printf("wrote %s (%d bytes, %d children)\n", ex.outPath, stat.Size(), len(body.Children))
	}
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

func trimNUL(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			return s[:i]
		}
	}
	return s
}
