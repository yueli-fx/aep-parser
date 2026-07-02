package aepmigrate

import (
	"math"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func assertConvertedLayerMaskProfile(t *testing.T, outPath string) {
	t.Helper()
	converted, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open converted: %v", err)
	}
	prof, err := profile.Build(converted, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("Build converted profile: %v", err)
	}
	layer := profileLayerByName(prof, "Plate")
	if layer == nil || len(layer.Masks) != 1 {
		t.Fatalf("converted masks = %+v, want one mask on Plate", layer)
	}
	mask := layer.Masks[0]
	if mask.Name != "Window" || mask.Mode != "subtract" || !mask.Inverted || !mask.Locked {
		t.Fatalf("mask identity/options = %+v", mask)
	}
	if !profileFloatSlicesEqual(mask.Color, []float64{255, 128, 0}) ||
		mask.MotionBlur != "off" ||
		mask.FeatherFalloff != "linear" ||
		math.Abs(mask.Opacity-0.5) > 1e-9 ||
		!profileFloatSlicesEqual(mask.Feather, []float64{12, 8}) ||
		math.Abs(mask.Expansion-(-4)) > 1e-9 ||
		!mask.Closed ||
		len(mask.Vertices) != 4 ||
		len(mask.PathKeyframes) != 2 {
		t.Fatalf("mask profile = %+v, want preserved options and path keyframes", mask)
	}
}

func profileLayerByName(prof *profile.Profile, name string) *profile.Layer {
	if prof == nil {
		return nil
	}
	for ci := range prof.Comps {
		for li := range prof.Comps[ci].Layers {
			if prof.Comps[ci].Layers[li].Name == name {
				return &prof.Comps[ci].Layers[li]
			}
		}
	}
	return nil
}

func profileFloatSlicesEqual(got, want []float64) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if math.Abs(got[i]-want[i]) > 1e-9 {
			return false
		}
	}
	return true
}
