// flightdeck/showcase/3d-camera/gen.go — from-scratch (no AE) showcase of the
// 3D layer + camera capabilities shipped 2026-06-15: Is3D enable, per-layer Z
// (parallax), Rotate Y (perspective tumble), all driven by a Go-created 3D
// camera. The whole scene is built from scratch in Go — no AE authoring.
//
// One frame demonstrates, side by side:
//   - PARALLAX: three same-size (300px) cards at increasing depth (z = -500 /
//     +400 / +1300) render at decreasing apparent size under the camera.
//   - TUMBLE:  a card at Rotate Y = 50° renders as a foreshortened trapezoid.
//
// The 3D switch + transform channels need no special builder calls — SetIs3D
// plus the ordinary SetPosition([x,y,z]) / SetRotateY on the reopened layers,
// because the from-scratch transform body is already a full 6-axis 3D schema
// (incidents/layer-3d-enable-bit-materializes.md).
//
// Writes 3d_camera.aep next to this file.
// Run from repo root: `go run ./flightdeck/showcase/3d-camera`.
package main

import (
	"fmt"
	"os"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

const outPath = "flightdeck/showcase/3d-camera/3d_camera.aep"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

// card adds a 300px square shape layer of the given colour at screen (x,y).
func card(comp *aep.Composition, name string, x, y float64, color [4]float64) {
	l, err := aep.NewShapeLayer(comp, name)
	must(err)
	r, err := l.RootGroup().AddRect()
	must(err)
	must(r.SetSize([2]float64{300, 300}))
	f, err := l.RootGroup().AddFill()
	must(err)
	must(f.SetColor(color))
	must(l.Position().SetStaticValue([2]float64{x, y}))
}

func main() {
	// AE 2020 target — the read floor; opens in 2020 + 2025. The 3D layer /
	// camera capabilities here all pass the double-version gate, so nothing is
	// 2025-specific (only use TargetAE2025 for genuinely 2025-only content).
	p := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(p, "Showcase3DCamera", 1920, 1080, 30, 5)
	must(err)

	// Dark backdrop (2D — ignores the camera, stays a full-frame stage).
	bg, err := aep.NewShapeLayer(comp, "BG")
	must(err)
	br, err := bg.RootGroup().AddRect()
	must(err)
	must(br.SetSize([2]float64{2400, 1400}))
	bf, err := bg.RootGroup().AddFill()
	must(err)
	must(bf.SetColor([4]float64{0.06, 0.07, 0.10, 1}))
	must(bg.Position().SetStaticValue([2]float64{960, 540}))

	// PARALLAX row — three identical 300px cards, placed near→far in depth.
	card(comp, "Near", 560, 420, [4]float64{0.95, 0.30, 0.35, 1})  // red, nearest → biggest
	card(comp, "Mid", 960, 420, [4]float64{0.30, 0.85, 0.45, 1})   // green, middle
	card(comp, "Far", 1360, 420, [4]float64{0.35, 0.55, 0.95, 1})  // blue, farthest → smallest
	// TUMBLE — a card rotated about Y for perspective foreshortening.
	card(comp, "Tumble", 960, 800, [4]float64{0.98, 0.80, 0.25, 1}) // yellow

	cam, err := aep.NewCameraLayer(comp, "Camera")
	must(err)
	_ = cam

	// Reopen (parse-the-clone) so the layers carry ldta/transform backrefs, then
	// flip the cards to 3D + set their depth / rotation, and place the camera.
	rp, err := aep.Reopen(p)
	must(err)
	c := rp.Compositions[0]

	type spec struct {
		name    string
		z       float64
		rotateY float64
	}
	for _, s := range []spec{
		{"Near", -500, 0},
		{"Mid", 400, 0},
		{"Far", 1300, 0},
		{"Tumble", 0, 50},
	} {
		l := c.LayerByName(s.name)
		must(l.SetIs3D(true))
		pos := l.Position().StaticValue
		xy, _ := pos.([]float64)
		x, y := 960.0, 540.0
		if len(xy) >= 2 {
			x, y = xy[0], xy[1]
		}
		must(l.SetPosition([]float64{x, y, s.z}))
		if s.rotateY != 0 {
			must(l.SetRotateY(s.rotateY))
		}
	}
	must(c.LayerByName("Camera").SetPosition([]float64{960, 540, -1700}))
	must(aep.MoveToEnd(c.LayerByName("BG")))

	out, err := os.Create(outPath)
	must(err)
	defer out.Close()
	must(rp.WriteAEP(out))
	fmt.Printf("wrote %s\n", outPath)
}
