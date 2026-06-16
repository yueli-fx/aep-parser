// internal/aep/add_effect_test.go
//
// Go round-trip tests for AddEffect (no AE required). Proves the spliced effect
// pair survives WriteAEP → re-parse with the new effect present, ordered last,
// and with settable parameters. AE acceptance is covered separately by the
// ship-gate (add_effect_shipgate_test.go).
package aep_test

import (
	"bytes"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// layerWithEffects is defined in property_structural_shipgate_test.go.

func TestAddEffect_RoundTrip(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_property_struct_baseline.aep")
	if err != nil {
		t.Skipf("baseline not present: %v", err)
	}
	l := layerWithEffects(proj)
	if l == nil {
		t.Fatal("no layer with effects in baseline")
	}
	before := len(l.Effects)

	fx, err := aep.AddEffect(l, "ADBE Gaussian Blur 2")
	if err != nil {
		t.Fatalf("AddEffect: %v", err)
	}
	if fx == nil {
		t.Fatal("AddEffect returned nil effect")
	}
	if got := len(l.Effects); got != before+1 {
		t.Fatalf("layer.Effects = %d, want %d", got, before+1)
	}
	if l.Effects[len(l.Effects)-1] != fx {
		t.Errorf("added effect is not last in layer.Effects")
	}
	if fx.MatchName != "ADBE Gaussian Blur 2" {
		t.Errorf("effect MatchName = %q, want ADBE Gaussian Blur 2", fx.MatchName)
	}

	// Tune the new effect's first surfaced param before writing (the default
	// GBlur instance surfaces "-0000"; the others are default-elided).
	const blurParam = "ADBE Gaussian Blur 2-0000"
	var blur *aep.Property
	for _, p := range fx.Parameters {
		if p.MatchName == blurParam {
			blur = p
		}
	}
	if blur == nil {
		t.Fatalf("%s param not found among %d params", blurParam, len(fx.Parameters))
	}
	if err := blur.SetStaticValue(33.0); err != nil {
		t.Fatalf("SetStaticValue: %v", err)
	}

	// WriteAEP → re-parse and confirm AE-independent structural survival.
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	rl := layerWithEffects(re)
	if rl == nil {
		t.Fatal("re-parsed: no layer with effects")
	}
	names := paradeChildNames(rl)
	want := []string{"ADBE Gaussian Blur 2", "ADBE Tint", "ADBE Fill", "ADBE Gaussian Blur 2"}
	if !eq(names, want) {
		t.Errorf("re-parsed parade = %v, want %v", names, want)
	}

	// The Blurriness value we set round-trips on the last (added) effect.
	rfx := rl.Effects[len(rl.Effects)-1]
	var rblur *aep.Property
	for _, p := range rfx.Parameters {
		if p.MatchName == blurParam {
			rblur = p
		}
	}
	if rblur == nil {
		t.Fatal("re-parsed param not found")
	}
	if v, ok := rblur.StaticValue.(float64); !ok || v != 33.0 {
		t.Errorf("re-parsed Blurriness = %v (ok=%v), want 33", rblur.StaticValue, ok)
	}
}

// TestAddEffect_AllTemplates_RoundTrip adds every supported effect to a fresh
// copy of the baseline layer and confirms it splices + survives WriteAEP →
// re-parse as the last effect in the parade. Cheap structural coverage for the
// whole template library (AE acceptance for a representative sample is the
// ship-gate's job).
func TestAddEffect_AllTemplates_RoundTrip(t *testing.T) {
	for _, name := range aep.SupportedEffects() {
		t.Run(name, func(t *testing.T) {
			proj, err := aep.Open("../../test_data/re_property_struct_baseline.aep")
			if err != nil {
				t.Skipf("baseline not present: %v", err)
			}
			l := layerWithEffects(proj)
			if l == nil {
				t.Fatal("no layer with effects")
			}
			before := len(l.Effects)
			fx, err := aep.AddEffect(l, name)
			if err != nil {
				t.Fatalf("AddEffect(%q): %v", name, err)
			}
			if fx.MatchName != name {
				t.Errorf("MatchName = %q, want %q", fx.MatchName, name)
			}
			if len(l.Effects) != before+1 {
				t.Fatalf("Effects = %d, want %d", len(l.Effects), before+1)
			}

			var buf bytes.Buffer
			if err := proj.WriteAEP(&buf); err != nil {
				t.Fatalf("WriteAEP: %v", err)
			}
			re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
			if err != nil {
				t.Fatalf("re-parse: %v", err)
			}
			rl := layerWithEffects(re)
			if rl == nil {
				t.Fatal("re-parsed: no layer with effects")
			}
			names := paradeChildNames(rl)
			if len(names) != before+1 {
				t.Fatalf("re-parsed parade len = %d, want %d (%v)", len(names), before+1, names)
			}
			if names[len(names)-1] != name {
				t.Errorf("re-parsed last effect = %q, want %q", names[len(names)-1], name)
			}
		})
	}
}

func TestRemoveEffect_RoundTrip(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_property_struct_baseline.aep")
	if err != nil {
		t.Skipf("baseline not present: %v", err)
	}
	l := layerWithEffects(proj)
	if l == nil {
		t.Fatal("no layer with effects")
	}
	// baseline: [Gaussian Blur, Tint, Fill]; remove middle (index 1, Tint).
	if err := aep.RemoveEffect(l, 1); err != nil {
		t.Fatalf("RemoveEffect: %v", err)
	}
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	rl := layerWithEffects(re)
	if rl == nil {
		t.Fatal("re-parsed: no layer with effects")
	}
	if got, want := paradeChildNames(rl), []string{"ADBE Gaussian Blur 2", "ADBE Fill"}; !eq(got, want) {
		t.Errorf("after RemoveEffect(1): parade = %v, want %v", got, want)
	}
}

func TestRemoveEffect_Refuse(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_property_struct_baseline.aep")
	if err != nil {
		t.Skipf("baseline not present: %v", err)
	}
	l := layerWithEffects(proj)
	if l == nil {
		t.Fatal("no layer with effects")
	}
	if err := aep.RemoveEffect(l, 99); err == nil {
		t.Error("RemoveEffect out-of-range: want error, got nil")
	}
	// Parade-less from-scratch shape layer.
	comp, _ := aep.NewComposition(aep.NewProject(aep.TargetAE2025), "M", 1920, 1080, 30, 5)
	sl, _ := aep.NewShapeLayer(comp, "S")
	if err := aep.RemoveEffect(sl.Layer, 0); err == nil {
		t.Error("RemoveEffect on parade-less layer: want error, got nil")
	}
}

func TestAddEffect_Unsupported(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_property_struct_baseline.aep")
	if err != nil {
		t.Skipf("baseline not present: %v", err)
	}
	l := layerWithEffects(proj)
	if l == nil {
		t.Fatal("no layer with effects")
	}
	if _, err := aep.AddEffect(l, "ADBE No Such Effect"); err == nil {
		t.Error("AddEffect with unsupported name: want error, got nil")
	}
}

func TestAddEffect_NoParade(t *testing.T) {
	// A from-scratch shape layer has no parsed property tree → AddEffect
	// refuses and points at the Reopen upgrade path.
	p := aep.NewProject(aep.TargetAE2025)
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatal(err)
	}
	sl, err := aep.NewShapeLayer(comp, "S")
	if err != nil {
		t.Fatal(err)
	}
	_, err = aep.AddEffect(sl.Layer, "ADBE Gaussian Blur 2")
	if err == nil {
		t.Fatal("AddEffect on from-scratch layer: want error, got nil")
	}
	if !strings.Contains(err.Error(), "Reopen") {
		t.Errorf("refuse error should mention the Reopen upgrade path, got: %v", err)
	}
}

// findLayer returns the layer with the given ID across all comps, or nil.
func findLayer(proj *aep.Project, id uint32) *aep.Layer {
	for _, c := range proj.Compositions {
		for _, l := range c.Layers {
			if l.ID == id {
				return l
			}
		}
	}
	return nil
}

// TestAddEffect_AutoCreateParade_ReopenedFreshLayer is the full from-scratch
// closure: build project + comp + shape layer in Go, Reopen to upgrade the
// built layer into a parsed one, then AddEffect — which must auto-create the
// Effect Parade (the lowered shape layer carries none) — and survive a write →
// re-parse round trip with the parade positioned before the Transform Group.
func TestAddEffect_AutoCreateParade_ReopenedFreshLayer(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := aep.NewShapeLayer(comp, "S"); err != nil {
		t.Fatal(err)
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	var l *aep.Layer
	for _, c := range rp.Compositions {
		for _, cl := range c.Layers {
			if cl.Name == "S" {
				l = cl
			}
		}
	}
	if l == nil {
		t.Fatal("reopened project: layer S not found")
	}
	if l.PropertyTree() == nil {
		t.Fatal("reopened layer has no property tree (Reopen did not upgrade it)")
	}
	if l.EffectsParade() != nil {
		t.Fatal("fresh shape layer should have no Effect Parade before AddEffect")
	}

	fx, err := aep.AddEffect(l, aep.EffectGaussianBlur)
	if err != nil {
		t.Fatalf("AddEffect (auto-create parade): %v", err)
	}
	if fx.MatchName != aep.EffectGaussianBlur {
		t.Errorf("MatchName = %q", fx.MatchName)
	}
	if l.EffectsParade() == nil {
		t.Fatal("parade not visible on scene tree after auto-create")
	}

	// Tree order mirrors chunk order: parade must precede the Transform Group.
	paradeIdx, transformIdx := -1, -1
	for i, c := range l.PropertyTree().Children {
		if g, ok := c.(*aep.AEPropertyGroup); ok {
			switch g.MatchName {
			case "ADBE Effect Parade":
				paradeIdx = i
			case "ADBE Transform Group":
				transformIdx = i
			}
		}
	}
	if paradeIdx < 0 || transformIdx < 0 || paradeIdx >= transformIdx {
		t.Errorf("parade idx %d / transform idx %d: parade must sit before Transform Group", paradeIdx, transformIdx)
	}

	var buf bytes.Buffer
	if err := rp.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	rl := findLayer(re, l.ID)
	if rl == nil {
		t.Fatal("re-parsed: layer not found")
	}
	if got, want := paradeChildNames(rl), []string{aep.EffectGaussianBlur}; !eq(got, want) {
		t.Errorf("re-parsed parade = %v, want %v", got, want)
	}
	if got, want := effectMatchNames(rl), []string{aep.EffectGaussianBlur}; !eq(got, want) {
		t.Errorf("re-parsed Effects = %v, want %v", got, want)
	}
}

// TestAddEffect_AutoCreateParade_ParsedFixtureLayer exercises auto-create on an
// AE-native effect-less layer (real AE sibling-group layout around the splice
// anchor), then proves write → re-parse survival.
func TestAddEffect_AutoCreateParade_ParsedFixtureLayer(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_text.aep")
	if err != nil {
		t.Skipf("re_text.aep not present: %v", err)
	}
	var l *aep.Layer
	for _, c := range proj.Compositions {
		for _, cl := range c.Layers {
			if cl.EffectsParade() == nil && cl.TransformGroup() != nil &&
				cl.Type != aep.LayerTypeCamera && cl.Type != aep.LayerTypeLight &&
				cl.Type != aep.LayerTypeShape && cl.PropertyTree() != nil {
				l = cl
				break
			}
		}
		if l != nil {
			break
		}
	}
	if l == nil {
		t.Skip("no parade-less AV layer in fixture")
	}

	if _, err := aep.AddEffect(l, aep.EffectInvert); err != nil {
		t.Fatalf("AddEffect (auto-create on AE-native layer): %v", err)
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	rl := findLayer(re, l.ID)
	if rl == nil {
		t.Fatal("re-parsed: layer not found")
	}
	if got, want := paradeChildNames(rl), []string{aep.EffectInvert}; !eq(got, want) {
		t.Errorf("re-parsed parade = %v, want %v", got, want)
	}
}

func TestAddEffect_RefuseCameraLight(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := aep.NewShapeLayer(comp, "S"); err != nil {
		t.Fatal(err)
	}
	if _, err := aep.NewCameraLayer(comp, "Cam"); err != nil {
		t.Fatal(err)
	}
	if _, err := aep.NewLightLayer(comp, "Light"); err != nil {
		t.Fatal(err)
	}
	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	for _, c := range rp.Compositions {
		for _, l := range c.Layers {
			if l.Type != aep.LayerTypeCamera && l.Type != aep.LayerTypeLight {
				continue
			}
			if _, err := aep.AddEffect(l, aep.EffectGaussianBlur); err == nil {
				t.Errorf("AddEffect on %s layer %q: want refuse, got nil", l.Type, l.Name)
			}
		}
	}
}

// TestReopen_WriteStable: writing the reopened project must reproduce the same
// bytes the original project wrote (parse → write byte fidelity over the
// freshly built content).
func TestReopen_WriteStable(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := aep.NewShapeLayer(comp, "S"); err != nil {
		t.Fatal(err)
	}
	var orig bytes.Buffer
	if err := p.WriteAEP(&orig); err != nil {
		t.Fatal(err)
	}
	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	var rewrite bytes.Buffer
	if err := rp.WriteAEP(&rewrite); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(orig.Bytes(), rewrite.Bytes()) {
		t.Errorf("reopened write differs from original: %d vs %d bytes", rewrite.Len(), orig.Len())
	}
}

// TestEffectConstants_MatchRegistry guards against the constants and the
// template registry diverging — every exported Effect* constant must be a
// supported (addable) match-name.
func TestEffectConstants_MatchRegistry(t *testing.T) {
	consts := []string{
		aep.EffectGaussianBlur, aep.EffectFill, aep.EffectTint,
		aep.EffectBrightnessContrast, aep.EffectTritone, aep.EffectLevels,
		aep.EffectLevelsIndividual, aep.EffectHueSaturation, aep.EffectBoxBlur,
		aep.EffectGlow, aep.EffectInvert, aep.EffectExposure,
		aep.EffectDropShadow, aep.EffectSharpen, aep.EffectMosaic,
		aep.EffectNoise, aep.EffectTransform, aep.EffectGradientRamp,
		aep.EffectFractalNoise, aep.EffectMotionTile, aep.EffectDirectionalBlur,
		aep.EffectLinearWipe, aep.EffectWaveWarp, aep.EffectCurves,
		aep.EffectSliderControl, aep.EffectPointControl, aep.EffectColorControl,
		aep.EffectAngleControl, aep.EffectCheckboxControl, aep.EffectPoint3DControl,
		aep.EffectSetMatte,
		aep.EffectTurbulentDisplace, aep.EffectRoughenEdges, aep.EffectEcho,
		aep.EffectRadialBlur, aep.EffectFourColorGradient, aep.EffectCheckerboard,
		aep.EffectGrid, aep.EffectStroke, aep.EffectCornerPin, aep.EffectVenetianBlinds,
		aep.EffectTwirl, aep.EffectPolarCoordinates, aep.EffectSpherize,
		aep.EffectMagnify, aep.EffectRipple, aep.EffectOpticsCompensation,
		aep.EffectPosterize, aep.EffectThreshold, aep.EffectFindEdges,
		aep.EffectColorEmboss, aep.EffectEmboss, aep.EffectStrobeLight,
		aep.EffectBrushStrokes, aep.EffectBevelAlpha, aep.EffectBevelEdges,
		aep.EffectPhotoFilter, aep.EffectVibrance, aep.EffectColorBalance,
		aep.EffectColorBalanceHLS, aep.EffectBlackAndWhite, aep.EffectGammaPedestalGain,
		aep.EffectChannelBlur, aep.EffectBilateralBlur, aep.EffectSmartBlur,
		aep.EffectUnsharpMask, aep.EffectShiftChannels, aep.EffectSolidComposite,
		aep.EffectMinimax, aep.EffectArithmetic, aep.EffectCircle,
		aep.EffectLensFlare, aep.EffectCellPattern, aep.EffectAdvancedLightning,
		aep.EffectBeam, aep.EffectPaintBucket, aep.EffectPosterizeTime,
		aep.EffectSimpleChoker, aep.EffectMatteChoker,
		aep.EffectBulge, aep.EffectOffset, aep.EffectMirror,
		aep.EffectFractal, aep.EffectWriteOn, aep.EffectScribble,
		aep.EffectEyedropperFill, aep.EffectAudioSpectrum, aep.EffectAudioWaveform,
		aep.EffectAutoLevels, aep.EffectAutoColor, aep.EffectAutoContrast,
		aep.EffectEqualize, aep.EffectLeaveColor, aep.EffectChangeToColor,
		aep.EffectChangeColor, aep.EffectRadialShadow, aep.EffectRemoveColorMatte,
		aep.EffectDustAndScratches, aep.EffectNoiseAlpha, aep.EffectNoiseHLS,
		aep.EffectRadialWipe, aep.EffectBlockDissolve,
		aep.EffectLumetri, aep.EffectLightning, aep.EffectCCRadialFastBlur,
		aep.EffectCCRadialBlur, aep.EffectCCCrossBlur, aep.EffectCCBendIt,
		aep.EffectCCBender, aep.EffectCCBlobbylize, aep.EffectCCFloMotion,
		aep.EffectCCGriddler, aep.EffectCCLens, aep.EffectCCPageTurn,
		aep.EffectCCPowerPin, aep.EffectCCRipplePulse, aep.EffectCCSlant,
		aep.EffectCCSmear, aep.EffectCCSplit, aep.EffectCCSplit2,
		aep.EffectCCTiler, aep.EffectCCWarpoMatic, aep.EffectCCLightBurst,
		aep.EffectCCLightRays, aep.EffectCCLightSweep, aep.EffectCCThreads,
		aep.EffectCCCylinder, aep.EffectCCSphere, aep.EffectCCSpotlight,
		aep.EffectCCGlass, aep.EffectCCHexTile, aep.EffectCCKaleida,
		aep.EffectCCMrSmoothie, aep.EffectCCPlastic, aep.EffectCCRepeTile,
		aep.EffectCCThreshold, aep.EffectCCThresholdRGB, aep.EffectCCPixelPolly,
		aep.EffectCCScatterize, aep.EffectCCStarBurst, aep.EffectCCForceMotionBlur,
		aep.EffectCCWideTime, aep.EffectCCColorOffset, aep.EffectCCToner,
		aep.EffectCCBurnFilm, aep.EffectCCVignette, aep.EffectCCSimpleWireRemoval,
		aep.EffectDisplacementMap, aep.EffectCompoundBlur, aep.EffectCCVectorBlur,
		aep.EffectBasic3D, aep.EffectBroadcastColors, aep.EffectChannelCombiner,
		aep.EffectCineonConverter, aep.EffectColorKey, aep.EffectColorRange,
		aep.EffectExtract, aep.EffectGeometryLegacy, aep.EffectGradientWipe,
		aep.EffectGrowBounds, aep.EffectKeyCleaner, aep.EffectLayerControl,
		aep.EffectLumaKey, aep.EffectMedian, aep.EffectNoiseHLSAuto,
		aep.EffectColorProfileConverter, aep.EffectTimeDisplacement, aep.EffectTimecode,
		aep.EffectCCBallAction, aep.EffectCCBubbles, aep.EffectCCComposite,
		aep.EffectCCDrizzle, aep.EffectCCEnvironment, aep.EffectCCGlassWipe,
		aep.EffectCCGlueGun, aep.EffectCCGridWipe, aep.EffectCCHair,
		aep.EffectCCImageWipe, aep.EffectCCJaws, aep.EffectCCLightWipe,
		aep.EffectCCMrMercury, aep.EffectCCParticleSystemsII, aep.EffectCCRadialScaleWipe,
		aep.EffectCCRain, aep.EffectCCScaleWipe, aep.EffectCCSnow,
		aep.EffectCCTwister, aep.EffectCCBlockLoad, aep.EffectCCColorNeutralizer,
		aep.EffectCCKernel, aep.EffectCCLineSweep, aep.EffectCCRainfall,
		aep.EffectCCSnowfall,
		aep.EffectAudioBackwards, aep.EffectAudioBassTreble, aep.EffectAudioDelay,
		aep.EffectAudioFlangeChorus, aep.EffectAudioHighLowPass, aep.EffectAudioModulator,
		aep.EffectAudioParametricEQ, aep.EffectAudioReverb, aep.EffectAudioStereoMixer,
		aep.EffectAudioTone,
		aep.EffectWarpStabilizer, aep.Effect3DGlasses, aep.EffectTimewarp,
		aep.EffectCCParticleWorld,
		aep.EffectBezierWarp, aep.EffectMeshWarp, aep.EffectChannelMixer,
		aep.EffectReshape, aep.EffectVectorPaint, aep.EffectTexturize,
		aep.EffectColorLink, aep.EffectCompoundArithmetic, aep.EffectSetChannels,
	}
	supported := map[string]bool{}
	for _, n := range aep.SupportedEffects() {
		supported[n] = true
	}
	if len(consts) != len(aep.SupportedEffects()) {
		t.Errorf("constant count %d != SupportedEffects count %d", len(consts), len(aep.SupportedEffects()))
	}
	for _, c := range consts {
		if !supported[c] {
			t.Errorf("constant %q not in SupportedEffects()", c)
		}
	}
}

func TestSupportedEffects(t *testing.T) {
	got := aep.SupportedEffects()
	if len(got) == 0 {
		t.Fatal("SupportedEffects returned empty")
	}
	found := false
	for _, n := range got {
		if n == "ADBE Gaussian Blur 2" {
			found = true
		}
	}
	if !found {
		t.Errorf("SupportedEffects %v missing ADBE Gaussian Blur 2", got)
	}
}
