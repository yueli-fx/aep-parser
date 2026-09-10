// Generate an AE2022 acceptance project using parameter names only.
// go run ./examples/authoring/ae2022-named-effects -out tmp/ae2022-named-effects.aep
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func must[T any](v T, err error) T { check(err); return v }
func check(err error) {
	if err != nil {
		panic(err)
	}
}

type setting struct {
	name  string
	value any
}
type demo struct {
	name, effect string
	values       []setting
	scalar       string
	from, to     float64
	vector       string
	start, end   []float64
}
type expectation struct {
	Comp, Layer, Effect, Parameter string
	property                       *aep.Property
}

var demos = []demo{
	{name: "01 Reference"},
	{name: "02 Blur EN", effect: aep.EffectGaussianBlur, values: []setting{{"Blurriness", 30.0}, {"Blur Dimensions", 1.0}, {"Repeat Edge Pixels", 1.0}}, scalar: "Blurriness", from: 30, to: 0},
	{name: "03 Blur ZH", effect: aep.EffectGaussianBlur, values: []setting{{"模糊度", 30.0}, {"模糊方向", 1.0}, {"重复边缘像素", 1.0}}, scalar: "模糊度", from: 30, to: 0},
	{name: "04 Drop Shadow", effect: aep.EffectDropShadow, values: []setting{{"Shadow Color", []float64{255, 255, 80, 40}}, {"Opacity", 90.0}, {"Direction", 135.0}, {"Distance", 35.0}, {"Softness", 12.0}, {"Shadow Only", 0.0}}, scalar: "Direction", from: 45, to: 315},
	{name: "05 Tint", effect: aep.EffectTint, values: []setting{{"Map Black To", []float64{255, 35, 20, 100}}, {"Map White To", []float64{255, 255, 175, 40}}, {"Amount to Tint", 100.0}}},
	{name: "06 Fill Color", effect: aep.EffectFill, values: []setting{{"Color", []float64{255, 255, 50, 60}}}, vector: "Color", start: []float64{255, 255, 50, 60}, end: []float64{255, 30, 120, 255}},
	{name: "07 Gradient Ramp", effect: aep.EffectGradientRamp, values: []setting{{"Start Color", []float64{255, 255, 170, 20}}, {"End Color", []float64{255, 30, 80, 255}}, {"End of Ramp", []float64{0.75, 0.75}}, {"Ramp Shape", 2.0}}, vector: "Start of Ramp", start: []float64{0.25, 0.25}, end: []float64{0.75, 0.25}},
	{name: "08 Fractal Noise", effect: aep.EffectFractalNoise, values: []setting{{"Contrast", 180.0}, {"Brightness", -20.0}, {"Complexity", 4.0}}, scalar: "Evolution", from: 0, to: 720},
	{name: "09 Directional Blur", effect: aep.EffectDirectionalBlur, values: []setting{{"Direction", 45.0}, {"Blur Length", 30.0}}, scalar: "Blur Length", from: 0, to: 60},
	{name: "10 Brightness Contrast", effect: aep.EffectBrightnessContrast, values: []setting{{"Brightness", 35.0}, {"Contrast", 40.0}}},
	{name: "11 Glow", effect: aep.EffectGlow, values: []setting{{"Glow Threshold", 45.0}, {"Glow Radius", 30.0}, {"Glow Intensity", 1.5}}, scalar: "Glow Intensity", from: 0.5, to: 3},
	{name: "12 Turbulent Displace", effect: aep.EffectTurbulentDisplace, values: []setting{{"Amount", 35.0}, {"Size", 25.0}}, scalar: "Evolution", from: 0, to: 360},
}

func token(comp *aep.Composition, name string) {
	layer := must(aep.NewShapeLayer(comp, name))
	r := must(layer.RootGroup().AddRect())
	check(r.SetSize([2]float64{220, 180}))
	f := must(layer.RootGroup().AddFill())
	check(f.SetColor([4]float64{0.1, 0.8, 0.7, 1}))
	s := must(layer.RootGroup().AddStroke())
	check(s.SetColor([4]float64{1, 0.7, 0.15, 1}))
	check(s.SetWidth(16))
	check(layer.Position().SetStaticValue([2]float64{240, 180}))
}

func compByName(project *aep.Project, name string) *aep.Composition {
	for _, comp := range project.Compositions {
		if comp.Name == name {
			return comp
		}
	}
	panic("missing composition: " + name)
}

func build() (*aep.Project, []expectation) {
	project := aep.NewProject(aep.TargetAE2022)
	must(aep.NewComposition(project, "00 OVERVIEW - AE2022 - 4 seconds", 1920, 1080, 30, 4))
	for _, d := range demos {
		comp := must(aep.NewComposition(project, d.name, 480, 360, 30, 4))
		token(comp, d.name)
	}
	controls := must(aep.NewComposition(project, "13 CONTROLS - inspect Effect Controls", 480, 360, 30, 4))
	token(controls, "Select this layer - five expression controls")
	project = must(aep.Reopen(project))
	var expected []expectation
	record := func(comp *aep.Composition, layer *aep.Layer, fx *aep.Effect, name string, prop *aep.Property) {
		for i, previous := range expected {
			if previous.Comp == comp.Name && previous.Effect == fx.MatchName && previous.property.MatchName == prop.MatchName {
				expected[i] = expectation{comp.Name, layer.Name, fx.MatchName, name, prop}
				return
			}
		}
		expected = append(expected, expectation{comp.Name, layer.Name, fx.MatchName, name, prop})
	}
	for i, d := range demos {
		comp := compByName(project, d.name)
		layer := comp.LayerByName(d.name)
		if d.effect != "" {
			fx := must(aep.AddEffect(layer, d.effect))
			for _, value := range d.values {
				record(comp, layer, fx, value.name, must(aep.SetEffectParam(layer, fx, value.name, value.value)))
			}
			if d.scalar != "" {
				record(comp, layer, fx, d.scalar, must(aep.AnimateEffectParam(layer, fx, d.scalar, []aep.ScalarKeyframe{{Time: 0, Value: d.from}, {Time: 2, Value: d.to}, {Time: 3.9, Value: d.from}})))
			}
			if d.vector != "" {
				record(comp, layer, fx, d.vector, must(aep.AnimateEffectParamVec(layer, fx, d.vector, []aep.VectorKeyframe{{Time: 0, Value: d.start}, {Time: 2, Value: d.end}, {Time: 3.9, Value: d.start}})))
			}
		}
		nested := must(aep.NewPrecompLayer(project.Compositions[0], comp, d.name))
		transform := aep.NewLayerTransform()
		// Precomp anchors use fractions of the source dimensions.
		check(transform.AnchorPoint().SetStaticValue([2]float64{0.5, 0.5}))
		check(transform.Position().SetStaticValue([2]float64{float64(i%4)*480 + 240, float64(i/4)*360 + 180}))
		check(aep.SetLayerTransform(nested, transform))
	}
	controls = compByName(project, "13 CONTROLS - inspect Effect Controls")
	controlLayer := controls.Layers[0]
	for _, d := range []demo{
		{effect: aep.EffectAngleControl, scalar: "Angle", from: 0, to: 180},
		{effect: aep.EffectSliderControl, scalar: "滑块", from: 0, to: 100},
		{effect: aep.EffectColorControl, vector: "Color", start: []float64{255, 255, 0, 0}, end: []float64{255, 0, 0, 255}},
		{effect: aep.EffectPointControl, vector: "Point", start: []float64{0.25, 0.25}, end: []float64{0.75, 0.75}},
		{effect: aep.EffectPoint3DControl, vector: "3D Point", start: []float64{0.25, 0.25, 0.25}, end: []float64{0.75, 0.75, 0.5}},
	} {
		fx := must(aep.AddEffect(controlLayer, d.effect))
		if d.scalar != "" {
			record(controls, controlLayer, fx, d.scalar, must(aep.AnimateEffectParam(controlLayer, fx, d.scalar, []aep.ScalarKeyframe{{Time: 0, Value: d.from}, {Time: 2, Value: d.to}})))
		}
		if d.vector != "" {
			record(controls, controlLayer, fx, d.vector, must(aep.AnimateEffectParamVec(controlLayer, fx, d.vector, []aep.VectorKeyframe{{Time: 0, Value: d.start}, {Time: 2, Value: d.end}})))
		}
	}
	return project, expected
}

func equalValue(a, b any) bool {
	switch x := a.(type) {
	case float64:
		y, ok := b.(float64)
		return ok && math.Abs(x-y) < 0.0001
	case []float64:
		y, ok := b.([]float64)
		if !ok || len(x) != len(y) {
			return false
		}
		for i := range x {
			if math.Abs(x[i]-y[i]) >= 0.0001 {
				return false
			}
		}
		return true
	default:
		return a == nil && b == nil
	}
}

func verify(project *aep.Project, expected []expectation) {
	if len(project.Compositions) != 14 {
		panic("expected 14 compositions")
	}
	for _, want := range expected {
		layer := compByName(project, want.Comp).LayerByName(want.Layer)
		var found *aep.Property
		for _, fx := range layer.Effects {
			if fx.MatchName == want.Effect {
				for _, prop := range fx.Parameters {
					if prop.MatchName == want.property.MatchName {
						found = prop
					}
				}
			}
		}
		if found == nil {
			panic("missing saved parameter: " + want.Comp + " / " + want.Parameter)
		}
		if (len(want.property.Keyframes) == 0 && !equalValue(found.StaticValue, want.property.StaticValue)) || len(found.Keyframes) != len(want.property.Keyframes) {
			panic("saved value mismatch: " + want.Comp + " / " + want.Parameter)
		}
		for i, kf := range found.Keyframes {
			original := want.property.Keyframes[i]
			if math.Abs(kf.Time-original.Time) > 0.0001 || !equalValue(kf.Value, original.Value) {
				panic("saved animation mismatch: " + want.Comp + " / " + want.Parameter)
			}
		}
	}
}

func main() {
	out := flag.String("out", "tmp/ae2022-named-effects.aep", "new output file")
	flag.Parse()
	project, expected := build()
	var data bytes.Buffer
	check(project.WriteAEP(&data))
	verify(must(aep.FromReader(bytes.NewReader(data.Bytes()))), expected)
	check(os.MkdirAll(filepath.Dir(*out), 0o755))
	f := must(os.OpenFile(*out, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644))
	defer f.Close()
	_, err := f.Write(data.Bytes())
	check(err)
	check(f.Close())
	report := must(json.MarshalIndent(expected, "", "  "))
	check(os.WriteFile(*out+".coverage.json", report, 0o644))
	fmt.Printf("Created %s: AE2022, 14 compositions, 16 effect instances, %d verified parameters\n", *out, len(expected))
}
