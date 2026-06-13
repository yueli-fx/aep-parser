// flightdeck/showcase/structural-ops/gen.go — from-scratch (no AE) showcase of
// the layer structural-op family (DuplicateLayer / MoveLayer / DeleteLayer /
// SetDimensionsSeparated). Builds one CENTRED CONCENTRIC bullseye on a dark BG
// so a single rendered frame (plus a .done structural dump) makes each op's
// before→after result verifiable.
//
// THE HARD CONSTRAINT THAT SHAPES THIS LAYOUT — solids cannot be repositioned.
//
//	A solid's Position is NOT materialised by this library: the solid template's
//	Transform Group ships with only the per-axis followers `ADBE Position_0` /
//	`ADBE Position_1` in MERGED mode (tdsb byte3 bit1 = 0, i.e. 0x03) and NO
//	merged `ADBE Position` leader at all. In merged mode AE reads the (absent)
//	leader, not the followers, so writing the followers' cdat does nothing — AE
//	falls back to the default position (comp centre) and EVERY solid renders
//	stacked at the centre. (The `layers` showcase documents the same limitation:
//	"Solid 的 Position 未物化（不可 SetPosition），故用居中不同尺寸堆出同心效果".
//	An earlier version of this file wrongly assumed the followers were the
//	authoritative separated axes and tried to `placeSolid` each square to a
//	quadrant — the render came back with all solids stacked at centre, only the
//	front teal visible.)
//
// So every visible square here is a CENTRED solid; spatial separation comes from
// DIFFERENT SIZES (concentric rings) + DIFFERENT COLOURS + STACK ORDER — the
// only solid levers AE actually honours. Each op is demonstrated against that
// concentric bullseye:
//
//	DeleteLayer  (VISIBLE)  — a GREEN ring sits between two PINK rings; deleting
//	             it makes the PINK ring behind it show through, so the render
//	             goes from pink-green-pink to one contiguous PINK band — the green
//	             ring visibly disappears.
//	MoveLayer    (VISIBLE)  — two SAME-SIZE solids (amber in front, violet behind)
//	             form the innermost square; only the front colour shows. Moving
//	             violet in front of amber flips the centre square AMBER→VIOLET.
//	DuplicateLayer (STRUCTURAL) — the clone shares the source's footage (same
//	             colour + size) and is also centred, so it overlaps the source
//	             exactly and adds no pixels; render.jsx's .done dump proves BOTH
//	             A_dup_src and A_dup_clone exist (the verifiable artifact).
//	SetDimensionsSeparated (STRUCTURAL) — operates on a green SHAPE layer's merged
//	             Position leader (shapes DO carry one, solids don't). Separation
//	             has no single-frame pixel signature; render.jsx dumps the leader's
//	             dimensionsSeparated state into the .done file.
//
// API notes RE'd while building this (real constraints, not gold-plating):
//
//   - DeleteLayer / DuplicateLayer / MoveLayer REFUSE non-AV layers ("only AV
//     layers supported"); a shape layer (LayerTypeShape) is rejected. So demos
//     A/B/C use SOLID layers (NewSolidLayer → LayerTypeAV).
//   - A solid cannot be repositioned (see the big comment above) — hence the
//     centred-concentric layout rather than four quadrants.
//   - SetDimensionsSeparated only accepts a MERGED "ADBE Position" leader, which
//     a solid lacks but a shape layer has — hence demo D is a shape layer (and
//     a shape layer CAN be positioned, so D sits as the outermost ring).
//   - All ops need a PARSED layer (itemList/Layr back-refs exist), so every token
//     is built first, then Reopen() upgrades the comp, then the ops run on the
//     reopened comp.
//   - Indices passed to the ops are 0-based positions in Composition.Layers (NOT
//     the 1-based Layer.Index field).
//
// Every op here (DeleteLayer / DuplicateLayer / MoveLayer / SetDimensionsSeparated)
// is double-version ship-gate verified individually; this 4-in-one COMBINATION is
// rendered + eyeballed via render.jsx (delivery contract red line 4 — see
// INDEX.md for the layout the render is checked against).
//
// Run from the repo root: `go run ./flightdeck/showcase/structural-ops`.
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

const outPath = "flightdeck/showcase/structural-ops/structural_ops.aep"

// float32-exact demo colours (NewSolidLayer stores color as float32; eighths/
// quarters round-trip cleanly).
var (
	amber  = [3]float64{1, 0.75, 0.25}
	violet = [3]float64{0.625, 0.5, 1}
	teal   = [3]float64{0.25, 0.875, 0.75}
	pink   = [3]float64{1, 0.5, 0.625}
	green3 = [3]float64{0.5, 0.8125, 0.375}
	dark   = [3]float64{0.0625, 0.0625, 0.125}
)

var green4 = [4]float64{0.5, 0.8125, 0.375, 1}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

// must2 panics on error, discarding the first (layer) return value — for the
// NewSolidLayer calls whose returned layer we re-fetch by name after Reopen.
func must2[T any](_ T, err error) {
	if err != nil {
		panic(err)
	}
}

// indexOf returns the 0-based slice position of the named layer in c.Layers
// (the form DeleteLayer / DuplicateLayer / MoveLayer expect), or -1.
func indexOf(c *aep.Composition, name string) int {
	for i, l := range c.Layers {
		if l.Name == name {
			return i
		}
	}
	return -1
}

func main() {
	p := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(p, "StructuralOpsShowcase", 1920, 1080, 30, 3)
	must(err)

	// --- Build all tokens (centred solids; ops run after Reopen) -------------
	// Build order = stack order (first built = front = rendered on top). The
	// concentric bullseye is assembled front (small) → back (large) so each
	// larger solid peeks out as a ring around the smaller ones in front of it.
	//
	// B (Move): two SAME-SIZE 220px solids form the innermost square. amber is
	// built first (front) so the centre reads AMBER; MoveLayer later puts violet
	// in front → centre flips to VIOLET.
	must2(aep.NewSolidLayer(comp, "B_amber_front", 220, 220, amber))
	must2(aep.NewSolidLayer(comp, "B_violet_back", 220, 220, violet))
	// A (Duplicate): a 380px teal solid → teal ring around the centre square.
	// Cloned after Reopen; the clone is the same footage (colour+size), centred,
	// so it overlaps exactly (no extra pixels) — proven structurally in .done.
	must2(aep.NewSolidLayer(comp, "A_dup_src", 380, 380, teal))
	// C (Delete): three concentric solids — inner PINK, middle GREEN, outer PINK.
	// Deleting the middle green makes the outer pink show through where the green
	// ring was → the green ring visibly disappears.
	must2(aep.NewSolidLayer(comp, "C_inner_pink", 560, 560, pink))
	must2(aep.NewSolidLayer(comp, "C_mid_green", 740, 740, green3))
	must2(aep.NewSolidLayer(comp, "C_outer_pink", 920, 920, pink))

	// D (Separate): a green SHAPE square as the OUTERMOST ring (1100px). A shape
	// layer carries a merged "ADBE Position" leader (separable) — a solid does
	// not. Shapes can also be positioned; this one stays centred to complete the
	// bullseye. Built as a 1100px rect centred at comp middle.
	dsh, err := aep.NewShapeLayer(comp, "D_sep_shape")
	must(err)
	dg := dsh.RootGroup()
	dr, err := dg.AddRect()
	must(err)
	must(dr.SetSize([2]float64{1100, 1100}))
	df, err := dg.AddFill()
	must(err)
	must(df.SetColor(green4))
	must(dsh.Position().SetStaticValue([2]float64{960, 540}))

	// Dark BG (comp-sized solid; centred, full frame; moved to back last).
	must2(aep.NewSolidLayer(comp, "BG", 1920, 1080, dark))

	// Reopen so every layer is parsed → the structural ops have itemList / Layr
	// back-refs and the shape's Position leader resolves.
	rp, err := aep.Reopen(p)
	must(err)
	rc := rp.Compositions[0]

	// --- A: DuplicateLayer (structural; clone overlaps the source) -----------
	aIdx := indexOf(rc, "A_dup_src")
	if aIdx < 0 {
		panic("A_dup_src not found after reopen")
	}
	if _, err := aep.DuplicateLayer(rc, aIdx, "A_dup_clone"); err != nil {
		panic(err)
	}

	// --- D: SetDimensionsSeparated on the green shape's Position leader -------
	dl := rc.LayerByName("D_sep_shape")
	if dl == nil {
		panic("D_sep_shape missing after reopen")
	}
	must(aep.SetDimensionsSeparated(dl.Position(), true))

	// --- B: MoveLayer — flip the centre square AMBER→VIOLET ------------------
	// amber sits in front of violet (lower slice index). Move violet ABOVE amber
	// (to amber's slot) so the centre square renders VIOLET.
	amberIdx := indexOf(rc, "B_amber_front")
	violetIdx := indexOf(rc, "B_violet_back")
	if amberIdx < 0 || violetIdx < 0 {
		panic("demo-B layers not found after reopen")
	}
	must(aep.MoveLayer(rc, violetIdx, amberIdx))

	// --- C: DeleteLayer the middle (green) ring ------------------------------
	cMidIdx := indexOf(rc, "C_mid_green")
	if cMidIdx < 0 {
		panic("C_mid_green not found after reopen")
	}
	must(aep.DeleteLayer(rc, cMidIdx))

	// BG must render behind everything.
	must(aep.MoveToEnd(rc.LayerByName("BG")))

	out, err := os.Create(outPath)
	must(err)
	defer out.Close()
	must(rp.WriteAEP(out))

	// Report final layer roster (order = render stack, top first).
	fmt.Printf("wrote %s (%d layers)\n", outPath, len(rc.Layers))
	for i, l := range rc.Layers {
		fmt.Printf("  [%d] %s\n", i, l.Name)
	}
}
