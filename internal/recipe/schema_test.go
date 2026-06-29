package recipe_test

import (
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/recipe"
)

func TestValidateAcceptsMinimalTextShapeRecipe(t *testing.T) {
	rec := minimalRecipe()

	report := recipe.Validate(rec)

	if report.Valid != true {
		t.Fatalf("Valid = false, report=%+v", report)
	}
	if len(report.Refusals) != 0 {
		t.Fatalf("Refusals = %+v, want none", report.Refusals)
	}
}

func TestValidateRejectsUnsupportedLayerType(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers = append(rec.Comps[0].Layers, recipe.Layer{
		Type: "camera",
		Name: "Camera 1",
	})

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "unsupported_layer_type")
}

func TestValidateRejectsOutOfRangeKeyframes(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Transform.PositionKeyframes = []recipe.VectorKeyframe{
		{Time: 5, Value: []float64{960, 540}},
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "keyframe_time_out_of_range")
}

func TestValidateRejectsUnsortedKeyframes(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Transform.PositionKeyframes = []recipe.VectorKeyframe{
		{Time: 1, Value: []float64{960, 540}},
		{Time: 0, Value: []float64{900, 540}},
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "keyframes_not_sorted")
}

func TestValidateRejectsOutOfRangeOpacityKeyframes(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Transform.OpacityKeyframes = []recipe.ScalarKeyframe{
		{Time: 5, Value: 50},
		{Time: 1, Value: 101},
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "keyframe_time_out_of_range")
	assertRefusal(t, report, "invalid_opacity_keyframe_value")
}

func TestValidateRejectsUnsortedOpacityKeyframes(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Transform.OpacityKeyframes = []recipe.ScalarKeyframe{
		{Time: 1, Value: 100},
		{Time: 0, Value: 50},
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "keyframes_not_sorted")
}

func TestValidateRejectsInvalidScaleKeyframes(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Transform.ScaleKeyframes = []recipe.VectorKeyframe{
		{Time: 5, Value: []float64{100, 100}},
		{Time: 1, Value: []float64{80}},
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "keyframe_time_out_of_range")
	assertRefusal(t, report, "invalid_vector_size")
}

func TestValidateRejectsUnsortedScaleKeyframes(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Transform.ScaleKeyframes = []recipe.VectorKeyframe{
		{Time: 1, Value: []float64{100, 100}},
		{Time: 0, Value: []float64{80, 80}},
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "keyframes_not_sorted")
}

func TestValidateRejectsOutOfRangeRotationKeyframes(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Transform.RotationKeyframes = []recipe.ScalarKeyframe{
		{Time: 5, Value: 45},
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "keyframe_time_out_of_range")
}

func TestValidateRejectsUnsortedRotationKeyframes(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Transform.RotationKeyframes = []recipe.ScalarKeyframe{
		{Time: 1, Value: 45},
		{Time: 0, Value: 0},
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "keyframes_not_sorted")
}

func TestValidateRejectsInvalidAnchorPointKeyframes(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Transform.AnchorPointKeyframes = []recipe.VectorKeyframe{
		{Time: 5, Value: []float64{0, 0}},
		{Time: 1, Value: []float64{10}},
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "keyframe_time_out_of_range")
	assertRefusal(t, report, "invalid_vector_size")
}

func TestValidateRejectsUnsortedAnchorPointKeyframes(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Transform.AnchorPointKeyframes = []recipe.VectorKeyframe{
		{Time: 1, Value: []float64{20, 10}},
		{Time: 0, Value: []float64{0, 0}},
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "keyframes_not_sorted")
}

func TestValidateAcceptsExpectedKeyframes(t *testing.T) {
	rec := minimalRecipe()
	rec.ExpectedProfile = recipe.ExpectedProfile{
		Keyframes: []recipe.ExpectedKeyframedProperty{{
			LayerName: "Title",
			MatchName: "ADBE Position",
			Keyframes: []recipe.ExpectedKeyframe{
				{Time: 0, Value: []float64{900, 540}},
				{Time: 1, Value: []float64{1020, 540}},
			},
		}},
	}

	report := recipe.Validate(rec)

	if !report.Valid {
		t.Fatalf("Valid = false, report=%+v", report)
	}
}

func TestValidateRejectsInvalidExpectedKeyframes(t *testing.T) {
	rec := minimalRecipe()
	rec.ExpectedProfile = recipe.ExpectedProfile{
		Keyframes: []recipe.ExpectedKeyframedProperty{{
			LayerName: "",
			MatchName: "",
			Keyframes: []recipe.ExpectedKeyframe{
				{Time: -1, Value: []float64{}},
			},
		}},
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "invalid_expected_profile")
}

func TestValidateRefusesUnsupportedEffects(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Effects = []recipe.Effect{{MatchName: "Third Party Magic"}}

	report := recipe.ValidateWithCapabilities(rec, recipe.StaticCapabilities{
		"Third Party Magic": recipe.CapabilityUnsupported,
	})

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "unsupported_effect")
}

func TestValidateReportsUsedCapabilities(t *testing.T) {
	report := recipe.ValidateWithCapabilities(minimalRecipe(), stableCapabilityIndex{})

	if !report.Valid {
		t.Fatalf("Valid = false, report=%+v", report)
	}
	assertCapability(t, report, "NewComposition")
	assertCapability(t, report, "NewTextLayer")
	assertCapability(t, report, "Layer.SetText")
	assertCapability(t, report, "NewShapeLayer")
	assertCapability(t, report, "RectNode.SetSize")
	assertCapability(t, report, "FillNode.SetColor")
	assertCapability(t, report, "SetLayerTransform")
	if len(report.Downgrades) != 0 {
		t.Fatalf("Downgrades = %+v, want none", report.Downgrades)
	}
}

func TestValidateReportsShapeStrokeCapabilities(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.Stroke = &recipe.StrokeSpec{
		Color:      []float64{255, 0, 0, 255},
		Width:      ptr(6),
		Opacity:    ptr(80),
		LineCap:    "projecting",
		LineJoin:   "bevel",
		MiterLimit: ptr(10),
		Taper: &recipe.StrokeTaperSpec{
			StartLength: ptr(30),
			EndLength:   ptr(40),
			StartWidth:  ptr(55),
			EndWidth:    ptr(65),
			StartEase:   ptr(35),
			EndEase:     ptr(50),
		},
		Wave: &recipe.StrokeWaveSpec{
			Amount:     ptr(20),
			Wavelength: ptr(70),
			Phase:      ptr(45),
		},
		Dashes: &recipe.StrokeDashesSpec{
			Dash: ptr(18),
			Gap:  ptr(7),
		},
	}

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if !report.Valid {
		t.Fatalf("Valid = false, report=%+v", report)
	}
	assertCapability(t, report, "VectorGroup.AddStroke")
	assertCapability(t, report, "StrokeNode.SetColor")
	assertCapability(t, report, "StrokeNode.SetWidth")
	assertCapability(t, report, "StrokeNode.SetOpacity")
	assertCapability(t, report, "StrokeNode.SetLineCap")
	assertCapability(t, report, "StrokeNode.SetLineJoin")
	assertCapability(t, report, "StrokeNode.SetMiterLimit")
	assertCapability(t, report, "StrokeTaper.SetStartLength")
	assertCapability(t, report, "StrokeTaper.SetEndLength")
	assertCapability(t, report, "StrokeTaper.SetStartWidth")
	assertCapability(t, report, "StrokeTaper.SetEndWidth")
	assertCapability(t, report, "StrokeTaper.SetStartEase")
	assertCapability(t, report, "StrokeTaper.SetEndEase")
	assertCapability(t, report, "StrokeWave.SetAmount")
	assertCapability(t, report, "StrokeWave.SetWavelength")
	assertCapability(t, report, "StrokeWave.SetPhase")
	assertCapability(t, report, "StrokeDashes.SetDash")
	assertCapability(t, report, "StrokeDashes.SetGap")
}

func TestValidateReportsShapeDetailCapabilities(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.Position = []float64{12, -6}
	rec.Comps[0].Layers[1].Shape.Roundness = ptr(18)
	rec.Comps[0].Layers[1].Shape.FillOpacity = ptr(45)

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if !report.Valid {
		t.Fatalf("Valid = false, report=%+v", report)
	}
	assertCapability(t, report, "RectNode.SetPosition")
	assertCapability(t, report, "RectNode.SetRoundness")
	assertCapability(t, report, "FillNode.SetOpacity")
}

func TestValidateReportsShapeStarCapabilities(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.Kind = "polygon"
	rec.Comps[0].Layers[1].Shape.Points = ptr(6)
	rec.Comps[0].Layers[1].Shape.Position = []float64{20, -10}
	rec.Comps[0].Layers[1].Shape.Rotation = ptr(30)
	rec.Comps[0].Layers[1].Shape.InnerRadius = ptr(45)
	rec.Comps[0].Layers[1].Shape.OuterRadius = ptr(120)
	rec.Comps[0].Layers[1].Shape.InnerRoundness = ptr(10)
	rec.Comps[0].Layers[1].Shape.OuterRoundness = ptr(20)

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if !report.Valid {
		t.Fatalf("Valid = false, report=%+v", report)
	}
	assertCapability(t, report, "VectorGroup.AddStar")
	assertCapability(t, report, "StarNode.SetStarType")
	assertCapability(t, report, "StarNode.SetPoints")
	assertCapability(t, report, "StarNode.SetPosition")
	assertCapability(t, report, "StarNode.SetRotation")
	assertCapability(t, report, "StarNode.SetInnerRadius")
	assertCapability(t, report, "StarNode.SetOuterRadius")
	assertCapability(t, report, "StarNode.SetInnerRoundness")
	assertCapability(t, report, "StarNode.SetOuterRoundness")
}

func TestValidateReportsCompBackgroundColorCapability(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].BackgroundColor = []float64{12, 34, 56}

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	assertCapability(t, report, "SetBGColor")
}

func TestValidateReportsCompLabelCapability(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Label = ptr(12)

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	assertCapability(t, report, "SetLabel")
}

func TestValidateReportsCompCommentCapability(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Comment = "reviewed recipe comp"

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	assertCapability(t, report, "SetComment")
}

func TestValidateReportsLayerLabelCapability(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Label = ptr(10)

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	assertCapability(t, report, "Layer.SetLabel")
}

func TestValidateReportsLayerCommentCapability(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Comment = "reviewed recipe layer"

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	assertCapability(t, report, "Layer.SetComment")
}

func TestValidateReportsLayerMotionBlurCapability(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].MotionBlur = boolPtr(true)

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	assertCapability(t, report, "Layer.SetMotionBlur")
}

func TestValidateReportsLayerShyCapability(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Shy = boolPtr(true)

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	assertCapability(t, report, "Layer.SetShy")
}

func TestValidateReportsLayerCommonSwitchCapabilities(t *testing.T) {
	rec := minimalRecipe()
	layer := &rec.Comps[0].Layers[0]
	layer.Visible = boolPtr(false)
	layer.Solo = boolPtr(true)
	layer.Locked = boolPtr(true)
	layer.EffectsEnabled = boolPtr(false)
	layer.AudioEnabled = boolPtr(false)
	layer.FrameBlendEnabled = boolPtr(true)

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	assertCapability(t, report, "Layer.SetVisible")
	assertCapability(t, report, "Layer.SetSolo")
	assertCapability(t, report, "Layer.SetLocked")
	assertCapability(t, report, "Layer.SetEffectsEnabled")
	assertCapability(t, report, "Layer.SetAudioEnabled")
	assertCapability(t, report, "Layer.SetFrameBlendEnabled")
}

func TestValidateReportsLayerAdvancedSwitchCapabilities(t *testing.T) {
	rec := minimalRecipe()
	layer := &rec.Comps[0].Layers[0]
	layer.CollapseTransform = boolPtr(true)
	layer.Is3D = boolPtr(true)
	layer.IsAdjust = boolPtr(true)
	layer.IsGuide = boolPtr(true)
	layer.SamplingBicubic = boolPtr(true)
	layer.FrameBlendPixelMotion = boolPtr(true)
	layer.PreserveTransparency = boolPtr(true)

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	assertCapability(t, report, "Layer.SetCollapseTransform")
	assertCapability(t, report, "Layer.SetIs3D")
	assertCapability(t, report, "Layer.SetIsAdjust")
	assertCapability(t, report, "Layer.SetIsGuide")
	assertCapability(t, report, "Layer.SetSamplingBicubic")
	assertCapability(t, report, "Layer.SetFrameBlendPixelMotion")
	assertCapability(t, report, "Layer.SetPreserveTransparency")
}

func TestValidateReportsLayerQualityAndBlendingModeCapabilities(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Quality = "draft"
	rec.Comps[0].Layers[0].BlendingMode = "multiply"

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	assertCapability(t, report, "Layer.SetQuality")
	assertCapability(t, report, "Layer.SetBlendingMode")
}

func TestValidateRejectsInvalidLayerQualityAndBlendingMode(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Quality = "preview"
	rec.Comps[0].Layers[0].BlendingMode = "sparkle"

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "invalid_layer_quality")
	assertRefusal(t, report, "invalid_layer_blending_mode")
}

func TestValidateReportsLayerTimingCapabilities(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].StartTime = ptr(0.25)
	rec.Comps[0].Layers[0].InPoint = ptr(0.1)
	rec.Comps[0].Layers[0].OutPoint = ptr(0.9)

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	assertCapability(t, report, "Layer.SetStartTime")
	assertCapability(t, report, "Layer.SetInPoint")
	assertCapability(t, report, "Layer.SetOutPoint")
}

func TestValidateRejectsInvalidLayerTiming(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].InPoint = ptr(-0.1)
	rec.Comps[0].Layers[0].OutPoint = ptr(10)

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "invalid_layer_in_point")
	assertRefusal(t, report, "invalid_layer_out_point")
}

func TestValidateReportsCompMotionBlurCapabilities(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].MotionBlur = &recipe.CompMotionBlurSpec{
		Enabled:             boolPtr(true),
		ShutterAngle:        ptr(360),
		ShutterPhase:        ptr(-90),
		AdaptiveSampleLimit: ptr(256),
		SamplesPerFrame:     ptr(32),
	}

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	assertCapability(t, report, "SetCompMotionBlur")
	assertCapability(t, report, "SetShutterAngle")
	assertCapability(t, report, "SetShutterPhase")
	assertCapability(t, report, "SetMotionBlurAdaptiveSampleLimit")
	assertCapability(t, report, "SetMotionBlurSamplesPerFrame")
}

func TestValidateReportsCompWorkAreaCapability(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].WorkArea = &recipe.CompWorkAreaSpec{
		Start: ptr(1.5),
		End:   ptr(3.5),
	}

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	assertCapability(t, report, "SetWorkArea")
}

func TestValidateReportsCompRendererCapability(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Renderer = "ADBE Escher"

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	assertCapability(t, report, "SetRenderer")
}

func TestValidateReportsCompResolutionFactorCapability(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].ResolutionFactor = []float64{2, 2}

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	assertCapability(t, report, "SetResolutionFactor")
}

func TestValidateReportsCompPixelAspectCapability(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].PixelAspect = ptr(2.0)

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	assertCapability(t, report, "SetPixelAspect")
}

func TestValidateReportsCompDisplayStartTimeCapability(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].DisplayStartTime = ptr(0.5)

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	assertCapability(t, report, "SetDisplayStartTime")
}

func TestValidateReportsCompFrameBlendingCapability(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].FrameBlending = boolPtr(true)

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	assertCapability(t, report, "SetFrameBlending")
}

func TestValidateReportsCompHideShyLayersCapability(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].HideShyLayers = boolPtr(true)

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	assertCapability(t, report, "SetHideShyLayers")
}

func TestValidateReportsCompPreserveNestedFrameRateCapability(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].PreserveNestedFrameRate = boolPtr(true)

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	assertCapability(t, report, "SetPreserveNestedFrameRate")
}

func TestValidateReportsCompPreserveNestedResolutionCapability(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].PreserveNestedResolution = boolPtr(true)

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	assertCapability(t, report, "SetPreserveNestedResolution")
}

func TestValidateRejectsInvalidCompBackgroundColor(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].BackgroundColor = []float64{12, 34, 256}

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "invalid_comp_background_color")
}

func TestValidateRejectsInvalidCompLabel(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Label = ptr(17)

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "invalid_comp_label")
}

func TestValidateRejectsInvalidLayerLabel(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Label = ptr(17)

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "invalid_layer_label")
}

func TestValidateRejectsInvalidCompWorkArea(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].WorkArea = &recipe.CompWorkAreaSpec{
		Start: ptr(3.5),
		End:   ptr(1.5),
	}

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "invalid_comp_work_area")
}

func TestValidateRejectsInvalidCompResolutionFactor(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].ResolutionFactor = []float64{2, 0}

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "invalid_comp_resolution_factor")
}

func TestValidateRejectsInvalidCompPixelAspect(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].PixelAspect = ptr(0)

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "invalid_comp_pixel_aspect")
}

func TestValidateRejectsInvalidCompDisplayStartTime(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].DisplayStartTime = ptr(-0.5)

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "invalid_comp_display_start_time")
}

func TestValidateRejectsInvalidCompMotionBlur(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].MotionBlur = &recipe.CompMotionBlurSpec{
		ShutterAngle:        ptr(721),
		AdaptiveSampleLimit: ptr(-1),
		SamplesPerFrame:     ptr(-1),
	}

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "invalid_comp_motion_blur_shutter_angle")
	assertRefusal(t, report, "invalid_comp_motion_blur_adaptive_sample_limit")
	assertRefusal(t, report, "invalid_comp_motion_blur_samples_per_frame")
}

func TestValidateReportsShapeTrimCapabilities(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.Trim = &recipe.TrimSpec{
		Start:  ptr(10),
		End:    ptr(85),
		Offset: ptr(15),
	}

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if !report.Valid {
		t.Fatalf("Valid = false, report=%+v", report)
	}
	assertCapability(t, report, "VectorGroup.AddTrim")
}

func TestValidateReportsShapeRoundCornersCapabilities(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.RoundCorners = &recipe.RoundCornersSpec{
		Radius: ptr(18),
	}

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if !report.Valid {
		t.Fatalf("Valid = false, report=%+v", report)
	}
	assertCapability(t, report, "VectorGroup.AddRoundCorners")
	assertCapability(t, report, "RoundCornersNode.SetRadius")
}

func TestValidateReportsShapeOffsetPathsCapabilities(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.OffsetPaths = &recipe.OffsetPathsSpec{
		Amount:     ptr(24),
		LineJoin:   "bevel",
		MiterLimit: ptr(2),
		Copies:     ptr(3),
		CopyOffset: ptr(1.5),
	}

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if !report.Valid {
		t.Fatalf("Valid = false, report=%+v", report)
	}
	assertCapability(t, report, "VectorGroup.AddOffsetPaths")
	assertCapability(t, report, "OffsetPathsNode.SetAmount")
	assertCapability(t, report, "OffsetPathsNode.SetLineJoin")
	assertCapability(t, report, "OffsetPathsNode.SetMiterLimit")
	assertCapability(t, report, "OffsetPathsNode.SetCopies")
	assertCapability(t, report, "OffsetPathsNode.SetCopyOffset")
}

func TestValidateReportsShapeRepeaterCapabilities(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.Repeater = &recipe.RepeaterSpec{
		Copies:       ptr(5),
		Offset:       ptr(2),
		Order:        "above",
		Anchor:       []float64{15, 25},
		Position:     []float64{120, 0},
		Scale:        []float64{80, 90},
		Rotation:     ptr(30),
		StartOpacity: ptr(100),
		EndOpacity:   ptr(25),
	}

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if !report.Valid {
		t.Fatalf("Valid = false, report=%+v", report)
	}
	assertCapability(t, report, "VectorGroup.AddRepeater")
	assertCapability(t, report, "RepeaterNode.SetCopies")
	assertCapability(t, report, "RepeaterNode.SetOffset")
	assertCapability(t, report, "RepeaterNode.SetOrder")
	assertCapability(t, report, "RepeaterTransform.SetAnchor")
	assertCapability(t, report, "RepeaterTransform.SetPosition")
	assertCapability(t, report, "RepeaterTransform.SetScale")
	assertCapability(t, report, "RepeaterTransform.SetRotation")
	assertCapability(t, report, "RepeaterTransform.SetStartOpacity")
	assertCapability(t, report, "RepeaterTransform.SetEndOpacity")
}

func TestValidateReportsShapeMergePathsCapabilities(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.MergePaths = &recipe.MergePathsSpec{
		Type: "subtract",
	}

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if !report.Valid {
		t.Fatalf("Valid = false, report=%+v", report)
	}
	assertCapability(t, report, "VectorGroup.AddMergePaths")
	assertCapability(t, report, "MergePathsNode.SetType")
}

func TestValidateReportsShapeZigZagCapabilities(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.ZigZag = &recipe.ZigZagSpec{
		Size:   ptr(40),
		Detail: ptr(8),
		Points: "smooth",
	}

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if !report.Valid {
		t.Fatalf("Valid = false, report=%+v", report)
	}
	assertCapability(t, report, "VectorGroup.AddZigZag")
	assertCapability(t, report, "ZigZagNode.SetSize")
	assertCapability(t, report, "ZigZagNode.SetDetail")
	assertCapability(t, report, "ZigZagNode.SetPoints")
}

func TestValidateReportsShapePuckerBloatCapabilities(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.PuckerBloat = &recipe.PuckerBloatSpec{
		Amount: ptr(100),
	}

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if !report.Valid {
		t.Fatalf("Valid = false, report=%+v", report)
	}
	assertCapability(t, report, "VectorGroup.AddPuckerBloat")
	assertCapability(t, report, "PuckerBloatNode.SetAmount")
}

func TestValidateReportsShapeTwistCapabilities(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.Twist = &recipe.TwistSpec{
		Angle:  ptr(150),
		Center: []float64{24, -12},
	}

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if !report.Valid {
		t.Fatalf("Valid = false, report=%+v", report)
	}
	assertCapability(t, report, "VectorGroup.AddTwist")
	assertCapability(t, report, "TwistNode.SetAngle")
	assertCapability(t, report, "TwistNode.SetCenter")
}

func TestValidateReportsShapeWigglePathsCapabilities(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.WigglePaths = &recipe.WigglePathsSpec{
		Size:             ptr(60),
		Detail:           ptr(30),
		WigglesPerSecond: ptr(4),
		RandomSeed:       ptr(9),
		Points:           "smooth",
		Correlation:      ptr(80),
		TemporalPhase:    ptr(45),
		SpatialPhase:     ptr(20),
	}

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if !report.Valid {
		t.Fatalf("Valid = false, report=%+v", report)
	}
	assertCapability(t, report, "VectorGroup.AddWigglePaths")
	assertCapability(t, report, "WigglePathsNode.SetSize")
	assertCapability(t, report, "WigglePathsNode.SetDetail")
	assertCapability(t, report, "WigglePathsNode.SetWigglesPerSecond")
	assertCapability(t, report, "WigglePathsNode.SetRandomSeed")
	assertCapability(t, report, "WigglePathsNode.SetPoints")
	assertCapability(t, report, "WigglePathsNode.SetCorrelation")
	assertCapability(t, report, "WigglePathsNode.SetTemporalPhase")
	assertCapability(t, report, "WigglePathsNode.SetSpatialPhase")
}

func TestValidateReportsShapeWiggleTransformCapabilities(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.WiggleTransform = &recipe.WiggleTransformSpec{
		Anchor:           []float64{10, 12},
		Position:         []float64{80, 60},
		Scale:            []float64{20, 30},
		Rotation:         ptr(25),
		WigglesPerSecond: ptr(4),
		RandomSeed:       ptr(9),
		Correlation:      ptr(80),
		TemporalPhase:    ptr(45),
		SpatialPhase:     ptr(20),
	}

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if !report.Valid {
		t.Fatalf("Valid = false, report=%+v", report)
	}
	assertCapability(t, report, "VectorGroup.AddWiggleTransform")
	assertCapability(t, report, "WigglerTransform.SetAnchor")
	assertCapability(t, report, "WigglerTransform.SetPosition")
	assertCapability(t, report, "WigglerTransform.SetScale")
	assertCapability(t, report, "WigglerTransform.SetRotation")
	assertCapability(t, report, "WiggleTransformNode.SetWigglesPerSecond")
	assertCapability(t, report, "WiggleTransformNode.SetRandomSeed")
	assertCapability(t, report, "WiggleTransformNode.SetCorrelation")
	assertCapability(t, report, "WiggleTransformNode.SetTemporalPhase")
	assertCapability(t, report, "WiggleTransformNode.SetSpatialPhase")
}

func TestValidateReportsTextStyleCapabilities(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].TextStyle = &recipe.TextStyleSpec{
		FontSize:      ptr(96),
		FillColor:     []float64{64, 128, 255, 255},
		Tracking:      ptr(120),
		FauxBold:      boolPtr(true),
		FauxItalic:    boolPtr(true),
		ApplyStroke:   boolPtr(true),
		StrokeColor:   []float64{255, 32, 64, 255},
		StrokeWidth:   ptr(8),
		Justification: "center",
	}

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if !report.Valid {
		t.Fatalf("Valid = false, report=%+v", report)
	}
	assertCapability(t, report, "Layer.SetRunFontSize")
	assertCapability(t, report, "Layer.SetRunFillColor")
	assertCapability(t, report, "Layer.SetRunTracking")
	assertCapability(t, report, "Layer.SetRunFauxBold")
	assertCapability(t, report, "Layer.SetRunFauxItalic")
	assertCapability(t, report, "Layer.SetRunApplyStroke")
	assertCapability(t, report, "Layer.SetRunStrokeColor")
	assertCapability(t, report, "Layer.SetRunStrokeWidth")
	assertCapability(t, report, "Layer.SetParagraphJustification")
}

func TestValidateRejectsInvalidTextStyle(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].TextStyle = &recipe.TextStyleSpec{
		RunIndex:       -1,
		ParagraphIndex: -1,
		FontSize:       ptr(0),
		FillColor:      []float64{255, 0},
		StrokeColor:    []float64{255, 0, 300},
		StrokeWidth:    ptr(-1),
		Justification:  "middle",
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "invalid_text_style_run_index")
	assertRefusal(t, report, "invalid_text_style_paragraph_index")
	assertRefusal(t, report, "invalid_text_font_size")
	assertRefusal(t, report, "invalid_text_fill_color")
	assertRefusal(t, report, "invalid_text_stroke_color")
	assertRefusal(t, report, "invalid_text_stroke_width")
	assertRefusal(t, report, "invalid_text_justification")
}

func TestValidateRejectsInvalidShapeDetail(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.Position = []float64{12}
	rec.Comps[0].Layers[1].Shape.Roundness = ptr(-1)
	rec.Comps[0].Layers[1].Shape.FillOpacity = ptr(101)

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "invalid_vector_size")
	assertRefusal(t, report, "invalid_shape_roundness")
	assertRefusal(t, report, "invalid_shape_fill_opacity")
}

func TestValidateRejectsInvalidShapeStar(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.Kind = "star"
	rec.Comps[0].Layers[1].Shape.Points = ptr(2)
	rec.Comps[0].Layers[1].Shape.Position = []float64{12}
	rec.Comps[0].Layers[1].Shape.InnerRadius = ptr(-1)
	rec.Comps[0].Layers[1].Shape.OuterRadius = ptr(-1)

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "invalid_shape_star_points")
	assertRefusal(t, report, "invalid_vector_size")
	assertRefusal(t, report, "invalid_shape_star_inner_radius")
	assertRefusal(t, report, "invalid_shape_star_outer_radius")
}

func TestValidateRejectsInvalidShapeTrim(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.Trim = &recipe.TrimSpec{
		Start: ptr(-1),
		End:   ptr(101),
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "invalid_shape_trim_start")
	assertRefusal(t, report, "invalid_shape_trim_end")
}

func TestValidateRejectsInvalidShapeRoundCorners(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.RoundCorners = &recipe.RoundCornersSpec{
		Radius: ptr(-1),
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "invalid_shape_round_corners_radius")
}

func TestValidateRejectsInvalidShapeOffsetPaths(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.OffsetPaths = &recipe.OffsetPathsSpec{
		LineJoin:   "square",
		MiterLimit: ptr(0),
		Copies:     ptr(0),
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "invalid_shape_offset_line_join")
	assertRefusal(t, report, "invalid_shape_offset_miter_limit")
	assertRefusal(t, report, "invalid_shape_offset_copies")
}

func TestValidateRejectsInvalidShapeRepeater(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.Repeater = &recipe.RepeaterSpec{
		Copies:       ptr(0),
		Order:        "front",
		Anchor:       []float64{15},
		Position:     []float64{120},
		Scale:        []float64{80},
		StartOpacity: ptr(-1),
		EndOpacity:   ptr(101),
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "invalid_shape_repeater_copies")
	assertRefusal(t, report, "invalid_shape_repeater_order")
	assertRefusal(t, report, "invalid_vector_size")
	assertRefusal(t, report, "invalid_shape_repeater_start_opacity")
	assertRefusal(t, report, "invalid_shape_repeater_end_opacity")
}

func TestValidateRejectsInvalidShapeMergePaths(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.MergePaths = &recipe.MergePathsSpec{
		Type: "mask",
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "invalid_shape_merge_paths_type")
}

func TestValidateRejectsInvalidShapeZigZag(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.ZigZag = &recipe.ZigZagSpec{
		Size:   ptr(-1),
		Detail: ptr(-1),
		Points: "sharp",
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "invalid_shape_zigzag_size")
	assertRefusal(t, report, "invalid_shape_zigzag_detail")
	assertRefusal(t, report, "invalid_shape_zigzag_points")
}

func TestValidateRejectsInvalidShapeTwist(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.Twist = &recipe.TwistSpec{
		Center: []float64{24},
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "invalid_vector_size")
}

func TestValidateRejectsInvalidShapeWigglePaths(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.WigglePaths = &recipe.WigglePathsSpec{
		Points:      "rounded",
		Correlation: ptr(101),
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "invalid_shape_wiggle_paths_points")
	assertRefusal(t, report, "invalid_shape_wiggle_paths_correlation")
}

func TestValidateRejectsInvalidShapeWiggleTransform(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.WiggleTransform = &recipe.WiggleTransformSpec{
		Anchor:      []float64{10},
		Position:    []float64{80},
		Scale:       []float64{20},
		Correlation: ptr(101),
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "invalid_vector_size")
	assertRefusal(t, report, "invalid_shape_wiggle_transform_correlation")
}

func TestValidateRejectsInvalidShapeStroke(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.Stroke = &recipe.StrokeSpec{
		Color:      []float64{255, 0},
		Width:      ptr(-1),
		Opacity:    ptr(101),
		LineCap:    "square",
		LineJoin:   "corner",
		MiterLimit: ptr(0),
		Dashes: &recipe.StrokeDashesSpec{
			Dash: ptr(-1),
			Gap:  ptr(-1),
		},
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "invalid_shape_stroke_color")
	assertRefusal(t, report, "invalid_shape_stroke_width")
	assertRefusal(t, report, "invalid_shape_stroke_opacity")
	assertRefusal(t, report, "invalid_shape_stroke_line_cap")
	assertRefusal(t, report, "invalid_shape_stroke_line_join")
	assertRefusal(t, report, "invalid_shape_stroke_miter_limit")
	assertRefusal(t, report, "invalid_shape_stroke_dash")
	assertRefusal(t, report, "invalid_shape_stroke_gap")
}

func TestValidateAcceptsSupportedEffects(t *testing.T) {
	effects := aep.SupportedEffects()
	if len(effects) == 0 {
		t.Fatal("SupportedEffects is empty")
	}
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Effects = []recipe.Effect{{MatchName: effects[0]}}

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if !report.Valid {
		t.Fatalf("Valid = false, report=%+v", report)
	}
	assertCapability(t, report, "AddEffect")
}

func TestValidateAcceptsSupportedEffectParams(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Effects = []recipe.Effect{{
		MatchName: "ADBE Gaussian Blur 2",
		Params: []recipe.EffectParam{
			{MatchName: "ADBE Gaussian Blur 2-0001", Value: 25.0},
			{MatchName: "ADBE Gaussian Blur 2-0003", Value: true},
		},
	}}

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if !report.Valid {
		t.Fatalf("Valid = false, report=%+v", report)
	}
	assertCapability(t, report, "SetEffectParam")
}

func TestValidateRejectsUnsupportedEffectParamValue(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Effects = []recipe.Effect{{
		MatchName: "ADBE Gaussian Blur 2",
		Params: []recipe.EffectParam{
			{MatchName: "ADBE Gaussian Blur 2-0001", Value: "high"},
		},
	}}

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "unsupported_effect_param_value")
}

func minimalRecipe() recipe.Recipe {
	return recipe.Recipe{
		SchemaVersion: recipe.SchemaVersion,
		Project:       recipe.ProjectSpec{Name: "Minimal title card"},
		Comps: []recipe.CompSpec{{
			Name:            "Main",
			Width:           1920,
			Height:          1080,
			FrameRate:       30,
			Duration:        4,
			BackgroundColor: []float64{0, 0, 0},
			Layers: []recipe.Layer{
				{
					Type: "text",
					Name: "Title",
					Text: "BOOYAH",
					Transform: recipe.Transform{
						Position: []float64{960, 540},
						Scale:    []float64{100, 100},
						Opacity:  ptr(100),
					},
				},
				{
					Type: "shape",
					Name: "Underline",
					Shape: &recipe.ShapeSpec{
						Kind:      "rect",
						Size:      []float64{640, 12},
						FillColor: []float64{255, 255, 255},
					},
					Transform: recipe.Transform{
						Position: []float64{960, 650},
					},
				},
			},
		}},
	}
}

func assertRefusal(t *testing.T, report recipe.Report, code string) {
	t.Helper()
	for _, refusal := range report.Refusals {
		if refusal.Code == code {
			return
		}
	}
	t.Fatalf("refusal %q not found in %+v", code, report.Refusals)
}

func assertCapability(t *testing.T, report recipe.Report, query string) {
	t.Helper()
	for _, use := range report.Capabilities {
		if use.Query == query {
			return
		}
	}
	t.Fatalf("capability %q not found in %+v", query, report.Capabilities)
}

func ptr(v float64) *float64 {
	return &v
}

func boolPtr(v bool) *bool {
	return &v
}

type stableCapabilityIndex struct{}

func (stableCapabilityIndex) Lookup(query string) recipe.CapabilityLookup {
	return recipe.CapabilityLookup{
		Query:  query,
		Status: recipe.CapabilitySupported,
		Symbol: query,
		Domain: "test",
		Tier:   "stable",
		Verify: "ae-accept",
	}
}
