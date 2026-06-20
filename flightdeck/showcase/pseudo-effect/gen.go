// Showcase: BuildPseudoEffect from-scratch pseudo effect exercising EVERY control
// kind, emitted as two .aep so the configurable values can be compared in AE:
//   pseudo_default.aep — every control at its bare type default ({Kind, Name})
//   pseudo_max.aep      — every configurable field set to a distinctive value
// Open either in AE and inspect the "Demo" effect's Effect Controls panel.
//
// Pseudo-effect controls are parameter containers (no rendered pixels of their
// own), so this is a read-value showcase: the proof is the live control tree in
// AE, not a frame. Run: go run ./flightdeck/showcase/pseudo-effect
//
// Configurable today: slider Min/Max/Default · angle Default · color Color ·
// checkbox Checked · dropdown Options/Default · point/3d PointDefault · layer
// LayerID (host default; binding a non-host layer does not render yet) · label/
// group Name. NOT yet configurable: slider visible-range (= valid range) and the
// percent display mode; see incidents/pseudo-control-render-field-map.md.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/example/aep-parser/internal/aep"
)

// defaultControls: every control kind at its bare type default (no custom value).
func defaultControls() []aep.PseudoControl {
	return []aep.PseudoControl{
		{Kind: aep.PseudoLabel, Name: "-- Values --"},
		{Kind: aep.PseudoSlider, Name: "Slider"},                                  // 0..100, default 0
		{Kind: aep.PseudoAngle, Name: "Angle"},                                    // 0 deg
		{Kind: aep.PseudoColor, Name: "Color"},                                    // white
		{Kind: aep.PseudoCheckbox, Name: "Checkbox"},                              // off
		{Kind: aep.PseudoDropdown, Name: "Dropdown", Options: []string{"A", "B"}}, // first selected
		{Kind: aep.PseudoLabel, Name: "-- Spatial --"},
		{Kind: aep.PseudoPoint, Name: "Point"},      // 0,0
		{Kind: aep.PseudoPoint3D, Name: "Point 3D"}, // 0,0,0
		{Kind: aep.PseudoLayer, Name: "Layer"},      // → host
		{Kind: aep.PseudoGroupStart, Name: "Group"},
		{Kind: aep.PseudoSlider, Name: "Nested Slider"},
		{Kind: aep.PseudoCheckbox, Name: "Nested Check"},
		{Kind: aep.PseudoGroupEnd},
	}
}

// maxControls: every configurable field set to a distinctive non-default value.
func maxControls() []aep.PseudoControl {
	return []aep.PseudoControl{
		{Kind: aep.PseudoLabel, Name: "-- Values --"},
		{Kind: aep.PseudoSlider, Name: "Slider", Min: -50, Max: 150, Default: 75},
		{Kind: aep.PseudoAngle, Name: "Angle", Default: 120},
		{Kind: aep.PseudoColor, Name: "Color", Color: []float64{0.2, 0.8, 1.0, 1.0}}, // cyan
		{Kind: aep.PseudoCheckbox, Name: "Checkbox", Checked: true},
		{Kind: aep.PseudoDropdown, Name: "Dropdown", Options: []string{"One", "Two", "Three", "Four"}, Default: 3},
		{Kind: aep.PseudoLabel, Name: "-- Spatial (dimmed) --", Dimmed: true}, // gray label

		{Kind: aep.PseudoPoint, Name: "Point", PointDefault: []float64{0.75, 0.25}},      // 300,100 in a 400 comp
		{Kind: aep.PseudoPoint3D, Name: "Point 3D", PointDefault: []float64{0.1, 0.2, 0.3}}, // 40,80,120
		{Kind: aep.PseudoLayer, Name: "Layer"}, // → host (non-host binding not renderable yet)
		{Kind: aep.PseudoGroupStart, Name: "Group"},
		{Kind: aep.PseudoSlider, Name: "Nested Slider", Min: 0, Max: 1000, Default: 250},
		{Kind: aep.PseudoCheckbox, Name: "Nested Check", Checked: true},
		{Kind: aep.PseudoGroupEnd},
	}
}

func build(variant string, controls []aep.PseudoControl) error {
	p := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(p, "Main", 400, 400, 30, 5)
	if err != nil {
		return err
	}
	// Host carries the effect; Source is a second layer the picker could select.
	if _, err := aep.NewShapeLayer(comp, "Host"); err != nil {
		return err
	}
	if _, err := aep.NewShapeLayer(comp, "Source"); err != nil {
		return err
	}
	rp, err := aep.Reopen(p)
	if err != nil {
		return err
	}
	var host *aep.Layer
	for _, c := range rp.Compositions {
		for _, l := range c.Layers {
			if l.Name == "Host" {
				host = l
			}
		}
	}
	if _, err := aep.BuildPseudoEffect(host, variant, "Demo", "Demo", controls); err != nil {
		return err
	}
	out := filepath.Join("flightdeck", "showcase", "pseudo-effect", "pseudo_"+variant+".aep")
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := rp.WriteAEP(f); err != nil {
		return err
	}
	fmt.Println("wrote", out, "—", len(controls), "controls")
	return nil
}

func main() {
	if err := build("default", defaultControls()); err != nil {
		panic(err)
	}
	if err := build("max", maxControls()); err != nil {
		panic(err)
	}
}
