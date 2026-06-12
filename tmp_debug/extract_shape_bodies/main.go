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
	// V2.2.1 子项⑪: rect/ellipse/fill/stroke bodies all re-extracted from one
	// combined fixture v2_2_shape_all_full.aep where every scalar/enum prop is
	// set non-default (so AE emits each cdat slot incl. the shape enums:
	// Direction / Blend Mode / Composite Order / Fill Rule). Regen via
	// tmp_debug/gen_shape_all_full.jsx. Path stays separate (bezier geometry).
	{"test_data/v2_2_shape_all_full.aep", "ADBE Vector Shape - Rect", "internal/serializer/templates/v2_2_shape_rect_body.bin"},
	{"test_data/v2_2_shape_all_full.aep", "ADBE Vector Graphic - Fill", "internal/serializer/templates/v2_2_shape_fill_body.bin"},
	{"test_data/v2_2_shape_all_full.aep", "ADBE Vector Shape - Ellipse", "internal/serializer/templates/v2_2_shape_ellipse_body.bin"},
	{"test_data/v2_2_shape_path_tolerance.aep", "ADBE Vector Shape - Group", "internal/serializer/templates/v2_2_shape_path_body.bin"},
	{"test_data/v2_2_shape_all_full.aep", "ADBE Vector Graphic - Stroke", "internal/serializer/templates/v2_2_shape_stroke_body.bin"},
	// 子项⑬ Stroke Dashes: dashed variant of the stroke body — same all-full
	// stroke PLUS one Dash 1 + Gap 1 pair (Offset is hidden-until-enabled and
	// script-ungettable, so deferred). Regen via tmp_debug/gen_stroke_dashed.jsx.
	// lowerStrokeNode picks this template when dashes are enabled.
	{"test_data/v2_2_stroke_dashed.aep", "ADBE Vector Graphic - Stroke", "internal/serializer/templates/v2_2_shape_stroke_dashed_body.bin"},
	// Gradient W (SetGradient): G-Fill body carries Grad Type / Start Pt / End
	// Pt scalars + the GCst→GCky→Utf8 prop.map stops XML. Source is an AE 25.6-
	// saved fixture (v2_2_gradient_src.aep ← py-aep gradient.aep) — the only
	// stops-bearing gradient available (AE elides defaults; ExtendScript can't
	// author custom stops). lowerGradientFillNode overwrites the scalars + the
	// Utf8 XML (length-variable; rifx recomputes LIST sizes).
	{"test_data/v2_2_gradient_src.aep", "ADBE Vector Graphic - G-Fill", "internal/serializer/templates/v2_2_shape_gradfill_body.bin"},
	// G-Stroke body: same fixture, same GCst→GCky→Utf8 slot structure as G-Fill
	// but under the ADBE Vector Graphic - G-Stroke tdmn. Used by lowerGradientStrokeNode.
	{"test_data/v2_2_gradient_src.aep", "ADBE Vector Graphic - G-Stroke", "internal/serializer/templates/v2_2_shape_gradstroke_body.bin"},
	// MG roadmap S3 — Trim Paths (`ADBE Vector Filter - Trim`): a vector filter
	// sibling of the shapes inside the Vectors Group, carrying Start / End /
	// Offset / Trim Type cdat slots (all set non-default in the fixture so AE
	// emits every slot). Regen via tmp_debug/gen_shape_trim.jsx. lowerTrimNode
	// overwrites the Start/End/Offset cdats.
	{"test_data/v2_2_trim.aep", "ADBE Vector Filter - Trim", "internal/serializer/templates/v2_2_shape_trim_body.bin"},
	// MG roadmap S5 — Repeater (`ADBE Vector Filter - Repeater`): radial/grid
	// duplication. Top-level Copies/Offset/Order + a nested `ADBE Vector Repeater
	// Transform` group (Anchor/Position/Scale/Rotation/Opacity 1·2). All set
	// non-default in the fixture so AE emits every slot. Regen via
	// tmp_debug/gen_shape_repeater.jsx.
	{"test_data/v2_2_repeater.aep", "ADBE Vector Filter - Repeater", "internal/serializer/templates/v2_2_shape_repeater_body.bin"},
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
