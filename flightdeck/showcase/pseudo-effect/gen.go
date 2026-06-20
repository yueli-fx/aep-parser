// Showcase: BuildPseudoEffect from-scratch pseudo effect exercising EVERY
// control kind, emitted as two .aep — pseudo_demo_en.aep (English labels) and
// pseudo_demo_zh.aep (Chinese labels, GBK). Open either in AE and inspect the
// "Demo" effect's Effect Controls panel.
//
// Pseudo-effect controls are parameter containers (no rendered pixels of their
// own), so this is a read-value showcase: the proof is the live control tree in
// AE, not a frame. Run: go run ./flightdeck/showcase/pseudo-effect
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/example/aep-parser/internal/aep"
)

// labels is one localized label set; the two files differ only in these strings.
type labels struct {
	effect                                         string
	basic, strength, rotation, tint, enabled, mode string
	modeOpts                                       []string
	spatial, center, pos3d, source                 string
	advanced, subAmount, invert                    string
}

func controls(L labels) []aep.PseudoControl {
	return []aep.PseudoControl{
		{Kind: aep.PseudoLabel, Name: L.basic},
		{Kind: aep.PseudoSlider, Name: L.strength, Min: 0, Max: 100, Default: 75},
		{Kind: aep.PseudoAngle, Name: L.rotation, Default: 45},
		{Kind: aep.PseudoColor, Name: L.tint, Color: []float64{1, 0.5, 0, 1}}, // orange
		{Kind: aep.PseudoCheckbox, Name: L.enabled, Checked: true},
		{Kind: aep.PseudoDropdown, Name: L.mode, Options: L.modeOpts, Default: 2},
		{Kind: aep.PseudoLabel, Name: L.spatial},
		{Kind: aep.PseudoPoint, Name: L.center, PointDefault: []float64{0.5, 0.5}},       // comp center
		{Kind: aep.PseudoPoint3D, Name: L.pos3d, PointDefault: []float64{0.25, 0.75, 0}}, // off-center
		// Layer picker with no explicit target → binds the host layer (AE hides
		// a picker whose tdpi is 0/None; a fresh PEM picker defaults to host).
		{Kind: aep.PseudoLayer, Name: L.source, LayerID: 0},
		{Kind: aep.PseudoGroupStart, Name: L.advanced},
		{Kind: aep.PseudoSlider, Name: L.subAmount, Min: -50, Max: 50, Default: 10},
		{Kind: aep.PseudoCheckbox, Name: L.invert, Checked: false},
		{Kind: aep.PseudoGroupEnd},
	}
}

// Pure ASCII — codepage-independent, displays correctly on any system.
var english = labels{
	effect: "Demo", basic: "-- Basic --", strength: "Strength", rotation: "Rotation",
	tint: "Tint", enabled: "Enabled", mode: "Blend Mode",
	modeOpts: []string{"Normal", "Add", "Screen"},
	spatial:  "-- Spatial --", center: "Center", pos3d: "Position 3D", source: "Source Layer",
	advanced: "Advanced", subAmount: "Sub Amount", invert: "Invert",
}

var chinese = labels{
	effect: "演示", basic: "— 基础 —", strength: "强度", rotation: "旋转",
	tint: "色调", enabled: "启用", mode: "混合模式",
	modeOpts: []string{"普通", "相加", "屏幕"},
	spatial:  "— 空间 —", center: "中心点", pos3d: "三维位置", source: "源图层",
	advanced: "高级", subAmount: "子数值", invert: "反转",
}

func build(L labels, uid string) error {
	p := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(p, "Main", 400, 400, 30, 5)
	if err != nil {
		return err
	}
	// Host carries the effect; Source is a second layer the None picker could pick.
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
	// matchName segment stays ASCII ("Demo"); the localized effect name is the
	// displayName (value-group tdsn / UTF-8 — CJK round-trips, codepage-free).
	if _, err := aep.BuildPseudoEffect(host, uid, "Demo", L.effect, controls(L)); err != nil {
		return err
	}
	out := filepath.Join("flightdeck", "showcase", "pseudo-effect", "pseudo_demo_"+uid+".aep")
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := rp.WriteAEP(f); err != nil {
		return err
	}
	fmt.Println("wrote", out, "—", len(controls(L)), "controls (all kinds)")
	return nil
}

func main() {
	if err := build(english, "en"); err != nil {
		panic(err)
	}
	if err := build(chinese, "zh"); err != nil {
		panic(err)
	}
}
