// flightdeck/showcase/stroke-detail/gen.go — from-scratch (no AE) showcase of the
// StrokeNode detail setters: Dashes, Line Join, Miter Limit (+ Taper / Wave probe).
// Lays out labelled zones on a dark BG, each isolating one stroke detail with a
// deliberately THICK (18-26) stroke on a CLOSED shape so the geometry reads at a
// glance. Writes stroke_detail.aep next to this file.
//
// Run from the repo root: `go run ./flightdeck/showcase/stroke-detail`.
//
// ─────────────────────────────────────────────────────────────────────────────
// WHY CLOSED SHAPES ONLY (root-cause of the previous tiny-render, delivery red-line 4d)
//
//   The first cut built each line/V as an OPEN path (AddPath + SetClosed(false)).
//   Every open-path zone rendered as a tiny ~26px blob: the path geometry that
//   round-trips green in Go COLLAPSES in AE for open shape paths, so a thick stroke
//   ended up painting a near-zero-size silhouette. This is a confirmed false-green
//   boundary: the stroke ship-gates (shape_stroke_shipgate_test.go,
//   shape_taperwave_shipgate_test.go, shape_dashes_shipgate_test.go) all build a
//   CLOSED rect and assert VALUE round-trip only — they never pixel-verify a render
//   and never use an open path. So the only stroke geometry proven to actually
//   render is a closed shape.
//
//   This showcase therefore uses closed primitives exclusively:
//     • Dashes      → closed rect (the one zone that already rendered correctly).
//     • Line Join   → closed 5-point STAR; the sharp outer points are exactly where
//                     Miter (spike) / Round (arc) / Bevel (chamfer) differ.
//     • Miter Limit → very-acute closed star, Miter join, high vs low limit (the
//                     low limit makes AE clip each spike to a bevel). Note: AE only
//                     keeps Miter Limit when Line Join == Miter (RE'd), so these
//                     zones use Miter join.
//     • Wave        → closed star outline wobbled into ripples (two param variants).
//                     Gated for value round-trip only, but it DOES render on a
//                     closed path, so it earns a visual zone.
//
//   DELIBERATELY OMITTED (no honest from-scratch visual — see INDEX.md "Known boundary"):
//     • Line Cap (Butt/Round/Projecting) — caps only appear at the ENDPOINTS of an
//       OPEN path, and open shape paths collapse (above).
//     • Taper — on a CLOSED loop there is no start/end, so AE renders a uniform-width
//       outline: the value round-trips but produces no visual (verified by render).
// ─────────────────────────────────────────────────────────────────────────────
package main

import (
	"fmt"
	"os"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

const outPath = "flightdeck/showcase/stroke-detail/stroke_detail.aep"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

var (
	cyan   = [4]float64{0.30, 0.85, 0.95, 1}
	amber  = [4]float64{1, 0.78, 0.25, 1}
	pink   = [4]float64{1, 0.45, 0.62, 1}
	green  = [4]float64{0.55, 0.86, 0.45, 1}
	violet = [4]float64{0.70, 0.58, 1, 1}
	white  = [4]float64{0.95, 0.96, 1, 1}
)

var comp *aep.Composition

// starStroke builds a standalone shape layer holding ONE closed star and ONE
// stroke (no fill), then drops the layer at comp position pos. The customise
// callback tweaks the stroke (join / miter / taper / wave). One closed star per
// layer keeps each zone's geometry isolated and rendering at full size.
func starStroke(name string, pos [2]float64, points, outerR, innerR float64, col [4]float64, width float64, customise func(s *aep.StrokeNode)) {
	l, err := aep.NewShapeLayer(comp, name)
	must(err)
	g := l.RootGroup()
	star, err := g.AddStar()
	must(err)
	must(star.SetPoints(points))
	must(star.SetOuterRadius(outerR))
	must(star.SetInnerRadius(innerR))
	st, err := g.AddStroke()
	must(err)
	must(st.SetColor(col))
	must(st.SetWidth(width))
	if customise != nil {
		customise(st)
	}
	must(l.Position().SetStaticValue(pos))
}

func main() {
	p := aep.NewProject(aep.TargetAE2020)
	c, err := aep.NewComposition(p, "StrokeDetailShowcase", 1920, 1080, 30, 3)
	must(err)
	comp = c

	// Dark background (built first; moved behind everything at the end).
	bg, err := aep.NewShapeLayer(comp, "BG")
	must(err)
	bgRect, err := bg.RootGroup().AddRect()
	must(err)
	must(bgRect.SetSize([2]float64{2200, 1300}))
	bgFill, err := bg.RootGroup().AddFill()
	must(err)
	must(bgFill.SetColor([4]float64{0.07, 0.08, 0.11, 1}))
	must(bg.Position().SetStaticValue([2]float64{960, 540}))

	// Grid: 4 columns × 2 rows of detail ZONES. Each zone is its own layer.
	colX := []float64{300, 740, 1180, 1620}
	rowY := []float64{330, 770}

	// ── (col0,row0): DASHES — a cyan dashed CLOSED rect outline. ─────────────
	// Dash 36 / Gap 22 at width 18 = obvious marching-ants outline. The one zone
	// that already rendered correctly (closed primitive, no endpoints needed).
	{
		l, err := aep.NewShapeLayer(comp, "01_Dashes")
		must(err)
		g := l.RootGroup()
		r, err := g.AddRect()
		must(err)
		must(r.SetSize([2]float64{300, 220}))
		s, err := g.AddStroke()
		must(err)
		must(s.SetColor(cyan))
		must(s.SetWidth(18))
		d := s.Dashes()
		must(d.SetDash(36)) // SetDash auto-enables dashing
		must(d.SetGap(22))
		must(l.Position().SetStaticValue([2]float64{colX[0], rowY[0]}))
	}

	// ── (col1..3,row0): LINE JOIN — three identical sharp stars, one per join. ─
	// The sharp OUTER points of the star are where the join geometry differs:
	// Miter = pointed spike, Round = rounded arc, Bevel = flat-chamfered corner.
	joins := []struct {
		name string
		join aep.StrokeLineJoin
		col  [4]float64
		x    float64
	}{
		{"02_Join_Miter", aep.StrokeLineJoinMiter, white, colX[1]},
		{"02_Join_Round", aep.StrokeLineJoinRound, green, colX[2]},
		{"02_Join_Bevel", aep.StrokeLineJoinBevel, violet, colX[3]},
	}
	for _, jj := range joins {
		starStroke(jj.name, [2]float64{jj.x, rowY[0]}, 5, 150, 62, jj.col, 24,
			func(s *aep.StrokeNode) { must(s.SetLineJoin(jj.join)) })
	}

	// ── (col0..1,row1): MITER LIMIT — two identical VERY-acute stars, Miter join, ─
	// high vs low limit. High limit keeps the long spikes; low limit (1) exceeds
	// the miter ratio so AE clips every spike down to a bevel. Sharp star (inner
	// 34 / outer 160) makes the outer angle acute enough for the clip to show.
	miters := []struct {
		name  string
		limit float64
		col   [4]float64
		x     float64
	}{
		{"03_Miter_High", 28, amber, colX[0]}, // high limit ⇒ spikes preserved
		{"03_Miter_Low", 1, cyan, colX[1]},    // limit 1 ⇒ AE clips spikes to bevels
	}
	for _, mm := range miters {
		starStroke(mm.name, [2]float64{mm.x, rowY[1]}, 5, 160, 34, mm.col, 20,
			func(s *aep.StrokeNode) {
				must(s.SetLineJoin(aep.StrokeLineJoinMiter))
				must(s.SetMiterLimit(mm.limit))
			})
	}

	// ── (col2..3,row1): WAVE — two stars whose stroke outlines are wobbled into ─
	// ripples, contrasting the Wave params. Fine = many small ripples (short
	// wavelength); Bold = fewer large ripples (long wavelength, big amount). Wave
	// IS ship-gated for value round-trip only, but unlike Taper it DOES render on
	// a closed path — so it earns a visual zone. (Taper was dropped: on a closed
	// loop it has no start/end and renders a uniform-width outline = no visual.)
	{
		starStroke("04_Wave_Fine", [2]float64{colX[2], rowY[1]}, 5, 150, 62, green, 16,
			func(s *aep.StrokeNode) {
				must(s.SetLineJoin(aep.StrokeLineJoinRound))
				wv := s.Wave()
				must(wv.SetAmount(20))
				must(wv.SetWavelength(22)) // short ⇒ many ripples
				must(wv.SetPhase(0))
			})
		starStroke("05_Wave_Bold", [2]float64{colX[3], rowY[1]}, 5, 150, 62, pink, 18,
			func(s *aep.StrokeNode) {
				must(s.SetLineJoin(aep.StrokeLineJoinRound))
				wv := s.Wave()
				must(wv.SetAmount(48))
				must(wv.SetWavelength(60)) // long ⇒ fewer, bigger ripples
				must(wv.SetPhase(0))
			})
	}

	// BG must render behind everything → move it to the bottom of the stack.
	must(aep.MoveToEnd(comp.LayerByName("BG")))

	out, err := os.Create(outPath)
	must(err)
	defer out.Close()
	must(p.WriteAEP(out))
	fmt.Printf("wrote %s (%d layers)\n", outPath, len(comp.Layers))
}
