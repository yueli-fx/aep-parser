// showcase/camera-light/gen.go — from-scratch (no AE) showcase of
// NewCameraLayer / NewLightLayer + their option setters. Camera/light only act on
// 3D geometry (none here) → 📋 readback showcase: gen.go creates one camera + one
// light and sets distinctive option values; verify.jsx dumps the layer type +
// option DOM values to .done for the user to read-check (not a png).
//
// Option setters are applied NON-FATALLY (cone angle/feather only exist on a Spot
// light; the embedded light template's type decides which apply).
//
// Writes camera_light.aep next to this file.
// Run from repo root: `go run ./showcase/camera-light`.
package main

import (
	"fmt"
	"os"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

const outPath = "showcase/camera-light/camera_light.aep"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func try(name string, err error) {
	if err != nil {
		fmt.Printf("  SKIP %s: %v\n", name, err)
		return
	}
	fmt.Printf("  ok   %s\n", name)
}

func main() {
	p := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(p, "CameraLight", 1920, 1080, 30, 5)
	must(err)

	// A trivial visible layer so the comp is non-empty.
	bg, err := aep.NewShapeLayer(comp, "BG")
	must(err)
	r, err := bg.RootGroup().AddRect()
	must(err)
	must(r.SetSize([2]float64{400, 400}))
	fl, err := bg.RootGroup().AddFill()
	must(err)
	must(fl.SetColor([4]float64{0.25, 0.6, 0.95, 1}))
	must(bg.Position().SetStaticValue([2]float64{960, 540}))

	// Camera layer + distinctive option values.
	cam, err := aep.NewCameraLayer(comp, "Cam01")
	must(err)
	fmt.Println("camera options:")
	try("SetCameraZoom(1500)", cam.SetCameraZoom(1500))
	try("SetCameraDepthOfField(true)", cam.SetCameraDepthOfField(true))
	try("SetCameraFocusDistance(2000)", cam.SetCameraFocusDistance(2000))
	try("SetCameraAperture(50)", cam.SetCameraAperture(50))
	try("SetCameraBlurLevel(120)", cam.SetCameraBlurLevel(120))

	// Light layer + distinctive option values.
	light, err := aep.NewLightLayer(comp, "Light01")
	must(err)
	fmt.Println("light options:")
	try("SetLightIntensity(140)", light.SetLightIntensity(140))
	try("SetLightColor([1,0.8,0.4])", light.SetLightColor([]float64{1, 0.8, 0.4, 1}))
	try("SetLightConeAngle(75)", light.SetLightConeAngle(75))
	try("SetLightConeFeather(40)", light.SetLightConeFeather(40))

	out, err := os.Create(outPath)
	must(err)
	defer out.Close()
	must(p.WriteAEP(out))
	fmt.Printf("wrote %s\n", outPath)
}
