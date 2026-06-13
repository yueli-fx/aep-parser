// flightdeck/showcase/text/gen.go — from-scratch (no AE) showcase of text layers.
// Creates text layers and sets their strings via SetText (after Reopen, the
// chunk-backed length-variable path). Writes text.aep next to this file.
// Run from repo root: `go run ./flightdeck/showcase/text`.
//
// NOTE: from-scratch text exposes NewTextLayer + SetText (the string). Font /
// size / colour setters are not part of the gated public API — the layers render
// with AE's default text style. render.jsx renders frame 0; eyeball appearance.
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

const outPath = "flightdeck/showcase/text/text.aep"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	p := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(p, "TextShowcase", 1920, 1080, 30, 3)
	must(err)

	// Dark BG (AE's default text colour is light, so it reads on dark).
	bg, err := aep.NewShapeLayer(comp, "BG")
	must(err)
	bgr, err := bg.RootGroup().AddRect()
	must(err)
	must(bgr.SetSize([2]float64{2200, 1300}))
	bgf, err := bg.RootGroup().AddFill()
	must(err)
	must(bgf.SetColor([4]float64{0.07, 0.08, 0.11, 1}))
	must(bg.Position().SetStaticValue([2]float64{960, 540}))

	// One text layer carrying a multi-line string. From-scratch text exposes
	// NewTextLayer + SetText only — there is no public position / colour / size
	// setter (the layer's Position property isn't materialized), so a single
	// layer with '\r'-separated paragraphs is the clean way to show several lines
	// without overlapping layers.
	if _, err := aep.NewTextLayer(comp, "T"); err != nil {
		panic(err)
	}

	rp, err := aep.Reopen(p)
	must(err)
	cc := rp.Compositions[0]
	must(cc.LayerByName("T").SetText("AEP-PARSER\rfrom-scratch text\rno AE needed"))
	must(aep.MoveToEnd(cc.LayerByName("BG")))

	out, err := os.Create(outPath)
	must(err)
	defer out.Close()
	must(rp.WriteAEP(out))
	fmt.Printf("wrote %s (%d layers)\n", outPath, len(cc.Layers))
}
