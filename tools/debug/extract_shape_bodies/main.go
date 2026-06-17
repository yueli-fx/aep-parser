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
	{"test_data/v2_2_shape_all_full.aep", "ADBE Vector Shape - Rect", "internal/serializer/templates/shapes/rect_body.bin"},
	{"test_data/v2_2_shape_all_full.aep", "ADBE Vector Graphic - Fill", "internal/serializer/templates/shapes/fill_body.bin"},
	{"test_data/v2_2_shape_all_full.aep", "ADBE Vector Shape - Ellipse", "internal/serializer/templates/shapes/ellipse_body.bin"},
	{"test_data/v2_2_shape_path_tolerance.aep", "ADBE Vector Shape - Group", "internal/serializer/templates/shapes/path_body.bin"},
	{"test_data/v2_2_shape_all_full.aep", "ADBE Vector Graphic - Stroke", "internal/serializer/templates/shapes/stroke_body.bin"},
	// 子项⑬ Stroke Dashes: dashed variant of the stroke body — same all-full
	// stroke PLUS one Dash 1 + Gap 1 pair (Offset is hidden-until-enabled and
	// script-ungettable, so deferred). Regen via tmp_debug/gen_stroke_dashed.jsx.
	// lowerStrokeNode picks this template when dashes are enabled.
	{"test_data/v2_2_stroke_dashed.aep", "ADBE Vector Graphic - Stroke", "internal/serializer/templates/shapes/stroke_dashed_body.bin"},
	// Gradient W (SetGradient): G-Fill body carries Grad Type / Start Pt / End
	// Pt scalars + the GCst→GCky→Utf8 prop.map stops XML. Source is an AE 25.6-
	// saved fixture (v2_2_gradient_src.aep ← py-aep gradient.aep) — the only
	// stops-bearing gradient available (AE elides defaults; ExtendScript can't
	// author custom stops). lowerGradientFillNode overwrites the scalars + the
	// Utf8 XML (length-variable; rifx recomputes LIST sizes).
	// G-Fill body now sourced from v2_2_gradient_type.aep (Grad Type=2 Radial +
	// Grad Start/End Pt set non-default → AE emits Grad Type AND the ramp-geometry
	// slots alongside Grad Colors). lowerGradientFillNode overwrites Grad Type
	// (default 1=Linear) + Start/End Pt with the node's values + the stops XML.
	// Regen the source via tmp_debug/gen_gradient_type.jsx (which itself opens
	// the v2_2_gradient_dir.aep produced by gen_gradient_dir.jsx).
	{"test_data/v2_2_gradient_type.aep", "ADBE Vector Graphic - G-Fill", "internal/serializer/templates/shapes/gradfill_body.bin"},
	// G-Stroke body now sourced from v2_2_gradstroke_geom.aep (the G-Stroke's
	// Grad Type=2 Radial + Start/End Pt set non-default → AE emits Grad Type +
	// Start/End Pt + the HiLite Length/Angle pair alongside Grad Colors, same
	// five geometry slots as the G-Fill body). lowerGradientStrokeNode overwrites
	// Type (default 1=Linear) + Start/End Pt + HiLite Length/Angle + the stops
	// XML. Regen the source via tmp_debug/gen_gradstroke_geom.jsx (opens the
	// stops-bearing v2_2_gradient_src.aep).
	{"test_data/v2_2_gradstroke_geom.aep", "ADBE Vector Graphic - G-Stroke", "internal/serializer/templates/shapes/gradstroke_body.bin"},
	// MG roadmap S3 — Trim Paths (`ADBE Vector Filter - Trim`): a vector filter
	// sibling of the shapes inside the Vectors Group, carrying Start / End /
	// Offset / Trim Type cdat slots (all set non-default in the fixture so AE
	// emits every slot). Regen via tmp_debug/gen_shape_trim.jsx. lowerTrimNode
	// overwrites the Start/End/Offset cdats.
	{"test_data/v2_2_trim.aep", "ADBE Vector Filter - Trim", "internal/serializer/templates/shapes/trim_body.bin"},
	// MG roadmap S5 — Repeater (`ADBE Vector Filter - Repeater`): radial/grid
	// duplication. Top-level Copies/Offset/Order + a nested `ADBE Vector Repeater
	// Transform` group (Anchor/Position/Scale/Rotation/Opacity 1·2). All set
	// non-default in the fixture so AE emits every slot. Regen via
	// tmp_debug/gen_shape_repeater.jsx.
	{"test_data/v2_2_repeater.aep", "ADBE Vector Filter - Repeater", "internal/serializer/templates/shapes/repeater_body.bin"},
	// MG roadmap S5 — Round Corners (`ADBE Vector Filter - RC`): rounds the
	// corners of the paths below it in the stack. Single sub-stream `ADBE Vector
	// RoundCorner Radius` (1D f64 BE @cdat[0:8]), set non-default in the fixture
	// so AE emits the slot. Regen via tmp_debug/gen_shape_roundcorners.jsx.
	{"test_data/v2_2_roundcorners.aep", "ADBE Vector Filter - RC", "internal/serializer/templates/shapes/roundcorners_body.bin"},
	// MG roadmap S5 — Offset Paths (`ADBE Vector Filter - Offset`): grows/shrinks
	// the paths below it by Amount px. Headline sub-stream `ADBE Vector Offset
	// Amount` (1D f64 BE @cdat[0:8], default 10) set non-default in the fixture so
	// AE emits the slot; Line Join / Miter Limit / Copies / Copy Offset stayed
	// default and are elided (modeled only Amount). Regen via
	// tmp_debug/gen_shape_offset.jsx.
	{"test_data/v2_2_offset.aep", "ADBE Vector Filter - Offset", "internal/serializer/templates/shapes/offset_body.bin"},
	// MG roadmap S5 — Merge Paths (`ADBE Vector Filter - Merge`): boolean-combines
	// the paths below it. Single sub-stream `ADBE Vector Merge Type` (1D f64 BE
	// enum @cdat[0:8]: 1=Merge 2=Add 3=Subtract 4=Intersect 5=Exclude) set
	// non-default (Subtract) in the fixture so AE emits the slot. Regen via
	// tmp_debug/gen_shape_merge.jsx.
	{"test_data/v2_2_merge.aep", "ADBE Vector Filter - Merge", "internal/serializer/templates/shapes/merge_body.bin"},
	// MG roadmap S5 — ZigZag (`ADBE Vector Filter - Zigzag`): distorts the paths
	// below it into a zigzag/wave. Sub-streams `ADBE Vector Zigzag Size` (amplitude,
	// default 5) + `ADBE Vector Zigzag Detail` (ridges per segment, default 10),
	// both set non-default in the fixture so AE emits the slots; `Zigzag Points`
	// (enum, default elided) not modeled. Regen via tmp_debug/gen_shape_zigzag.jsx.
	{"test_data/v2_2_zigzag.aep", "ADBE Vector Filter - Zigzag", "internal/serializer/templates/shapes/zigzag_body.bin"},
	// MG roadmap S5 — PolyStar (`ADBE Vector Shape - Star`): a parametric star
	// shape (like Rect/Ellipse). Sub-streams Points / Position / Rotation / Inner
	// Radius / Outer Radius / Inner·Outer Roundess set non-default so AE emits the
	// slots; Star Type (default Star) + Shape Direction stay default (elided) —
	// Star-type only modeled (Polygon deferred). Regen via tmp_debug/gen_shape_star.jsx.
	{"test_data/v2_2_star.aep", "ADBE Vector Shape - Star", "internal/serializer/templates/shapes/star_body.bin"},
	// PolyStar Polygon type (`ADBE Vector Shape - Star` with Star Type=2): a
	// separate body — a Polygon drops Inner Radius/Roundness (AE hides them), so
	// the slots differ from the Star body. Authored with Type=2 + Points/Position/
	// Rotation/Outer Radius/Outer Roundness non-default. lowerStarNode picks this
	// template when the node's StarType is Polygon. Regen via tmp_debug/gen_shape_polygon.jsx.
	{"test_data/v2_2_polygon.aep", "ADBE Vector Shape - Star", "internal/serializer/templates/shapes/starpolygon_body.bin"},
	// MG roadmap S5 — Pucker & Bloat (`ADBE Vector Filter - PB`): bows the paths'
	// edges outward (bloat, +) or inward (pucker, −). Single sub-stream `ADBE
	// Vector PuckerBloat Amount` (1D f64 BE percent, default 10) set non-default in
	// the fixture so AE emits the slot. Regen via tmp_debug/gen_shape_puckerbloat.jsx.
	{"test_data/v2_2_puckerbloat.aep", "ADBE Vector Filter - PB", "internal/serializer/templates/shapes/puckerbloat_body.bin"},
	// MG roadmap S5 — Twist (`ADBE Vector Filter - Twist`): rotates the paths
	// below it progressively (more rotation farther from the center), bowing
	// straight edges into spirals. Headline sub-stream `ADBE Vector Twist Angle`
	// (1D f64 BE degrees, default 10) set non-default in the fixture so AE emits
	// the slot; `Twist Center` (Vec2, default [0,0]) stays default and is elided
	// (modeled only Angle). Regen via tmp_debug/gen_shape_twist.jsx.
	{"test_data/v2_2_twist.aep", "ADBE Vector Filter - Twist", "internal/serializer/templates/shapes/twist_body.bin"},
	// MG roadmap S5 — Wiggle Paths (`ADBE Vector Filter - Roughen` internally):
	// roughens the paths below it with time-varying random displacement. Headline
	// sub-streams `ADBE Vector Roughen Size` (amplitude) / `Roughen Detail` /
	// `Temporal Freq` (Wiggles/Second) / `Random Seed` set non-default in the
	// fixture so AE emits those four slots; Points (enum) / Correlation /
	// Temporal·Spatial Phase stay default and are elided (modeled only the four
	// headline controls). Regen via tmp_debug/gen_shape_wiggle.jsx.
	{"test_data/v2_2_wiggle.aep", "ADBE Vector Filter - Roughen", "internal/serializer/templates/shapes/wiggle_body.bin"},
	// MG roadmap S5 — Wiggle Transform (`ADBE Vector Filter - Wiggler`): randomly
	// jitters a transform applied to the paths below (usually after a Repeater).
	// Top-level Temporal Freq (Wiggles/Second) + Random Seed, plus a nested `ADBE
	// Vector Wiggler Transform` group (Anchor/Position/Scale/Rotation amplitudes).
	// All modeled slots set non-default in the fixture so AE emits them;
	// Correlation / Temporal·Spatial Phase stay default and are elided. Regen via
	// tmp_debug/gen_shape_wiggletransform.jsx.
	{"test_data/v2_2_wiggletransform.aep", "ADBE Vector Filter - Wiggler", "internal/serializer/templates/shapes/wiggletransform_body.bin"},
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
