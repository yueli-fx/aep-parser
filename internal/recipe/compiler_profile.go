package recipe

import (
	"fmt"
	"math"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/profile"
)

func hasExpectedProfile(expected ExpectedProfile) bool {
	return expected.CompCount != nil ||
		expected.LayerCount != nil ||
		expected.TextLayerCount != nil ||
		expected.ShapeLayerCount != nil ||
		expected.Name != "" ||
		expected.Width != nil ||
		expected.Height != nil ||
		expected.FrameRate != nil ||
		expected.Duration != nil ||
		expected.Label != nil ||
		expected.Comment != "" ||
		expected.MotionGraphicsTemplateName != "" ||
		len(expected.BackgroundColor) > 0 ||
		len(expected.ResolutionFactor) > 0 ||
		expected.PixelAspect != nil ||
		expected.DisplayStartTime != nil ||
		expected.Renderer != "" ||
		expected.Draft3D != nil ||
		expected.FrameBlending != nil ||
		expected.HideShyLayers != nil ||
		expected.PreserveNestedFrameRate != nil ||
		expected.PreserveNestedResolution != nil ||
		expected.MotionBlur != nil ||
		expected.WorkArea != nil ||
		len(expected.EssentialGraphics) > 0 ||
		len(expected.Layers) > 0 ||
		len(expected.Effects) > 0 ||
		len(expected.Properties) > 0 ||
		len(expected.TextStyles) > 0 ||
		len(expected.Keyframes) > 0 ||
		len(expected.Masks) > 0
}

func checkExpectedProfile(expected ExpectedProfile, prof *profile.Profile) []ProfileCheck {
	var checks []ProfileCheck
	add := func(path string, expected, actual any, passed bool) {
		msg := ""
		if !passed {
			msg = fmt.Sprintf("%s expected %v, got %v", path, expected, actual)
		}
		checks = append(checks, ProfileCheck{
			Path:     path,
			Passed:   passed,
			Expected: expected,
			Actual:   actual,
			Message:  msg,
		})
	}
	if expected.CompCount != nil {
		actual := prof.Fingerprint.CompCount
		add("expected_profile.comp_count", *expected.CompCount, actual, actual == *expected.CompCount)
	}
	if expected.LayerCount != nil {
		actual := prof.Fingerprint.LayerCount
		add("expected_profile.layer_count", *expected.LayerCount, actual, actual == *expected.LayerCount)
	}
	if expected.TextLayerCount != nil {
		actual := countProfileLayers(prof, func(layer profile.Layer) bool { return layer.Text != nil })
		add("expected_profile.text_layer_count", *expected.TextLayerCount, actual, actual == *expected.TextLayerCount)
	}
	if expected.ShapeLayerCount != nil {
		actual := countProfileLayers(prof, func(layer profile.Layer) bool { return len(layer.Shapes) > 0 })
		add("expected_profile.shape_layer_count", *expected.ShapeLayerCount, actual, actual == *expected.ShapeLayerCount)
	}
	if expected.Name != "" {
		actual := ""
		if len(prof.Comps) > 0 {
			actual = prof.Comps[0].Name
		}
		add("expected_profile.name", expected.Name, actual, actual == expected.Name)
	}
	if expected.Width != nil {
		actual := 0.0
		if len(prof.Comps) > 0 {
			actual = float64(prof.Comps[0].Width)
		}
		add("expected_profile.width", *expected.Width, actual, actual == *expected.Width)
	}
	if expected.Height != nil {
		actual := 0.0
		if len(prof.Comps) > 0 {
			actual = float64(prof.Comps[0].Height)
		}
		add("expected_profile.height", *expected.Height, actual, actual == *expected.Height)
	}
	if expected.FrameRate != nil {
		actual := 0.0
		if len(prof.Comps) > 0 {
			actual = prof.Comps[0].FrameRate
		}
		add("expected_profile.frame_rate", *expected.FrameRate, actual, math.Abs(actual-*expected.FrameRate) < 1e-6)
	}
	if expected.Duration != nil {
		actual := 0.0
		if len(prof.Comps) > 0 {
			actual = prof.Comps[0].Duration
		}
		add("expected_profile.duration", *expected.Duration, actual, math.Abs(actual-*expected.Duration) < 1e-6)
	}
	if expected.Label != nil {
		actual := 0.0
		if len(prof.Comps) > 0 {
			actual = float64(prof.Comps[0].Label)
		}
		add("expected_profile.label", *expected.Label, actual, actual == *expected.Label)
	}
	if expected.Comment != "" {
		actual := ""
		if len(prof.Comps) > 0 {
			actual = prof.Comps[0].Comment
		}
		add("expected_profile.comment", expected.Comment, actual, actual == expected.Comment)
	}
	if expected.MotionGraphicsTemplateName != "" {
		actual := ""
		if len(prof.Comps) > 0 {
			actual = prof.Comps[0].MotionGraphicsTemplateName
		}
		add("expected_profile.motion_graphics_template_name", expected.MotionGraphicsTemplateName, actual, actual == expected.MotionGraphicsTemplateName)
	}
	if len(expected.BackgroundColor) > 0 {
		actual := []float64(nil)
		if len(prof.Comps) > 0 {
			actual = rgbToFloatSlice(prof.Comps[0].BackgroundColor)
		}
		add("expected_profile.background_color", expected.BackgroundColor, actual, profileValueEqual(expected.BackgroundColor, actual))
	}
	if len(expected.ResolutionFactor) > 0 {
		actual := []float64(nil)
		if len(prof.Comps) > 0 {
			actual = uint16PairToFloatSlice(prof.Comps[0].ResolutionFactor)
		}
		add("expected_profile.resolution_factor", expected.ResolutionFactor, actual, profileValueEqual(expected.ResolutionFactor, actual))
	}
	if expected.PixelAspect != nil {
		actual := 0.0
		if len(prof.Comps) > 0 {
			actual = prof.Comps[0].PixelAspect
		}
		add("expected_profile.pixel_aspect", *expected.PixelAspect, actual, math.Abs(actual-*expected.PixelAspect) < 1e-6)
	}
	if expected.DisplayStartTime != nil {
		actual := 0.0
		if len(prof.Comps) > 0 {
			actual = prof.Comps[0].DisplayStartTime
		}
		add("expected_profile.display_start_time", *expected.DisplayStartTime, actual, math.Abs(actual-*expected.DisplayStartTime) < 1e-6)
	}
	if expected.Renderer != "" {
		actual := ""
		if len(prof.Comps) > 0 {
			actual = prof.Comps[0].Renderer
		}
		add("expected_profile.renderer", expected.Renderer, actual, actual == expected.Renderer)
	}
	if expected.Draft3D != nil {
		actual := false
		if len(prof.Comps) > 0 {
			actual = prof.Comps[0].Draft3D
		}
		add("expected_profile.draft_3d", *expected.Draft3D, actual, actual == *expected.Draft3D)
	}
	if expected.FrameBlending != nil {
		actual := false
		if len(prof.Comps) > 0 {
			actual = prof.Comps[0].FrameBlending
		}
		add("expected_profile.frame_blending", *expected.FrameBlending, actual, actual == *expected.FrameBlending)
	}
	if expected.HideShyLayers != nil {
		actual := false
		if len(prof.Comps) > 0 {
			actual = prof.Comps[0].HideShyLayers
		}
		add("expected_profile.hide_shy_layers", *expected.HideShyLayers, actual, actual == *expected.HideShyLayers)
	}
	if expected.PreserveNestedFrameRate != nil {
		actual := false
		if len(prof.Comps) > 0 {
			actual = prof.Comps[0].PreserveNestedFrameRate
		}
		add("expected_profile.preserve_nested_frame_rate", *expected.PreserveNestedFrameRate, actual, actual == *expected.PreserveNestedFrameRate)
	}
	if expected.PreserveNestedResolution != nil {
		actual := false
		if len(prof.Comps) > 0 {
			actual = prof.Comps[0].PreserveNestedResolution
		}
		add("expected_profile.preserve_nested_resolution", *expected.PreserveNestedResolution, actual, actual == *expected.PreserveNestedResolution)
	}
	if expected.MotionBlur != nil {
		checkExpectedMotionBlur(expected.MotionBlur, prof, add)
	}
	if expected.WorkArea != nil {
		checkExpectedWorkArea(expected.WorkArea, prof, add)
	}
	for i, controller := range expected.EssentialGraphics {
		controllerPath := fmt.Sprintf("expected_profile.essential_graphics[%d]", i)
		var actualName any
		var actualType any
		if len(prof.Comps) > 0 && i < len(prof.Comps[0].EssentialGraphics) {
			actualName = prof.Comps[0].EssentialGraphics[i].Name
			actualType = prof.Comps[0].EssentialGraphics[i].Type
		}
		add(controllerPath+".name", controller.Name, actualName, actualName == controller.Name)
		if controller.Type != "" {
			add(controllerPath+".type", controller.Type, actualType, actualType == controller.Type)
		}
	}
	for i, expectedLayer := range expected.Layers {
		checkExpectedLayer(i, expectedLayer, prof, add)
	}
	for i, expectedEffect := range expected.Effects {
		effectPath := fmt.Sprintf("expected_profile.effects[%d]", i)
		effect := findProfileEffect(prof, expectedEffect.LayerName, expectedEffect.MatchName)
		add(effectPath, expectedEffect.MatchName, effectMatchName(effect), effect != nil)
		if effect == nil {
			continue
		}
		for pi, expectedParam := range expectedEffect.Params {
			paramPath := fmt.Sprintf("%s.params[%d]", effectPath, pi)
			param := findProfileParam(effect.Params, expectedParam.MatchName)
			if param == nil {
				add(paramPath, expectedParam.Value, nil, false)
				continue
			}
			if expectedParam.Value != nil {
				passed := profileValueEqual(expectedParam.Value, param.StaticValue)
				add(paramPath, expectedParam.Value, param.StaticValue, passed)
			}
			if expectedParam.TargetLayer != "" {
				var actual any
				passed := false
				if param.LayerRef != nil {
					actual = param.LayerRef.Name
					passed = param.LayerRef.Name == expectedParam.TargetLayer
				}
				add(paramPath+".target_layer", expectedParam.TargetLayer, actual, passed)
			}
			if expectedParam.Expression != "" {
				add(paramPath+".expression", expectedParam.Expression, param.Expression, param.Expression == expectedParam.Expression)
			}
			if expectedParam.ExpressionEnabled != nil {
				var actual any
				passed := false
				if param.ExpressionEnabled != nil {
					actual = *param.ExpressionEnabled
					passed = *param.ExpressionEnabled == *expectedParam.ExpressionEnabled
				}
				add(paramPath+".expression_enabled", *expectedParam.ExpressionEnabled, actual, passed)
			}
		}
	}
	for i, expectedProp := range expected.Properties {
		propPath := fmt.Sprintf("expected_profile.properties[%d]", i)
		prop := findProfileLayerProperty(prof, expectedProp.LayerName, expectedProp.MatchName)
		if prop == nil {
			add(propPath, expectedProp.Value, nil, false)
			continue
		}
		if expectedProp.Value != nil {
			passed := profileValueEqual(expectedProp.Value, prop.StaticValue)
			add(propPath, expectedProp.Value, prop.StaticValue, passed)
		}
		if expectedProp.Expression != "" {
			add(propPath+".expression", expectedProp.Expression, prop.Expression, prop.Expression == expectedProp.Expression)
		}
		if expectedProp.ExpressionEnabled != nil {
			var actual any
			passed := false
			if prop.ExpressionEnabled != nil {
				actual = *prop.ExpressionEnabled
				passed = *prop.ExpressionEnabled == *expectedProp.ExpressionEnabled
			}
			add(propPath+".expression_enabled", *expectedProp.ExpressionEnabled, actual, passed)
		}
	}
	for i, expectedStyle := range expected.TextStyles {
		stylePath := fmt.Sprintf("expected_profile.text_styles[%d]", i)
		layer := findProfileLayer(prof, expectedStyle.LayerName)
		if layer == nil || layer.Text == nil {
			add(stylePath, "text layer", nil, false)
			continue
		}
		if expectedStyle.FontSize != nil {
			path := stylePath + ".font_size"
			actual, ok := profileRunFloat(layer.Text.Runs, expectedStyle.RunIndex, func(run profile.TextStyleRun) float64 {
				return run.FontSize
			})
			add(path, *expectedStyle.FontSize, actual, ok && math.Abs(actual-*expectedStyle.FontSize) < 1e-9)
		}
		if len(expectedStyle.FillColor) > 0 {
			path := stylePath + ".fill_color"
			actual, ok := profileRunColor(layer.Text.Runs, expectedStyle.RunIndex, func(run profile.TextStyleRun) [4]float64 {
				return run.FillColor
			})
			add(path, expectedStyle.FillColor, actual, ok && profileValueEqual(expectedStyle.FillColor, actual))
		}
		if expectedStyle.AutoLeading != nil {
			path := stylePath + ".auto_leading"
			actual, ok := profileRunBool(layer.Text.Runs, expectedStyle.RunIndex, func(run profile.TextStyleRun) bool {
				return run.AutoLeading
			})
			add(path, *expectedStyle.AutoLeading, actual, ok && actual == *expectedStyle.AutoLeading)
		}
		if expectedStyle.Leading != nil {
			path := stylePath + ".leading"
			actual, ok := profileRunFloat(layer.Text.Runs, expectedStyle.RunIndex, func(run profile.TextStyleRun) float64 {
				return run.Leading
			})
			add(path, *expectedStyle.Leading, actual, ok && math.Abs(actual-*expectedStyle.Leading) < 1e-9)
		}
		if expectedStyle.Tracking != nil {
			path := stylePath + ".tracking"
			actual, ok := profileRunFloat(layer.Text.Runs, expectedStyle.RunIndex, func(run profile.TextStyleRun) float64 {
				return run.Tracking
			})
			add(path, *expectedStyle.Tracking, actual, ok && math.Abs(actual-*expectedStyle.Tracking) < 1e-9)
		}
		if expectedStyle.BaselineShift != nil {
			path := stylePath + ".baseline_shift"
			actual, ok := profileRunFloat(layer.Text.Runs, expectedStyle.RunIndex, func(run profile.TextStyleRun) float64 {
				return run.BaselineShift
			})
			add(path, *expectedStyle.BaselineShift, actual, ok && math.Abs(actual-*expectedStyle.BaselineShift) < 1e-9)
		}
		if expectedStyle.HorizontalScale != nil {
			path := stylePath + ".horizontal_scale"
			actual, ok := profileRunFloat(layer.Text.Runs, expectedStyle.RunIndex, func(run profile.TextStyleRun) float64 {
				return run.HorizontalScale
			})
			add(path, *expectedStyle.HorizontalScale, actual, ok && math.Abs(actual-*expectedStyle.HorizontalScale) < 1e-9)
		}
		if expectedStyle.VerticalScale != nil {
			path := stylePath + ".vertical_scale"
			actual, ok := profileRunFloat(layer.Text.Runs, expectedStyle.RunIndex, func(run profile.TextStyleRun) float64 {
				return run.VerticalScale
			})
			add(path, *expectedStyle.VerticalScale, actual, ok && math.Abs(actual-*expectedStyle.VerticalScale) < 1e-9)
		}
		if expectedStyle.Tsume != nil {
			path := stylePath + ".tsume"
			actual, ok := profileRunFloat(layer.Text.Runs, expectedStyle.RunIndex, func(run profile.TextStyleRun) float64 {
				return run.Tsume
			})
			add(path, *expectedStyle.Tsume, actual, ok && math.Abs(actual-*expectedStyle.Tsume) < 1e-9)
		}
		if expectedStyle.FauxBold != nil {
			path := stylePath + ".faux_bold"
			actual, ok := profileRunBool(layer.Text.Runs, expectedStyle.RunIndex, func(run profile.TextStyleRun) bool {
				return run.FauxBold
			})
			add(path, *expectedStyle.FauxBold, actual, ok && actual == *expectedStyle.FauxBold)
		}
		if expectedStyle.FauxItalic != nil {
			path := stylePath + ".faux_italic"
			actual, ok := profileRunBool(layer.Text.Runs, expectedStyle.RunIndex, func(run profile.TextStyleRun) bool {
				return run.FauxItalic
			})
			add(path, *expectedStyle.FauxItalic, actual, ok && actual == *expectedStyle.FauxItalic)
		}
		if expectedStyle.ApplyStroke != nil {
			path := stylePath + ".apply_stroke"
			actual, ok := profileRunBool(layer.Text.Runs, expectedStyle.RunIndex, func(run profile.TextStyleRun) bool {
				return run.ApplyStroke
			})
			add(path, *expectedStyle.ApplyStroke, actual, ok && actual == *expectedStyle.ApplyStroke)
		}
		if len(expectedStyle.StrokeColor) > 0 {
			path := stylePath + ".stroke_color"
			actual, ok := profileRunColor(layer.Text.Runs, expectedStyle.RunIndex, func(run profile.TextStyleRun) [4]float64 {
				return run.StrokeColor
			})
			add(path, expectedStyle.StrokeColor, actual, ok && profileValueEqual(expectedStyle.StrokeColor, actual))
		}
		if expectedStyle.StrokeWidth != nil {
			path := stylePath + ".stroke_width"
			actual, ok := profileRunFloat(layer.Text.Runs, expectedStyle.RunIndex, func(run profile.TextStyleRun) float64 {
				return run.StrokeWidth
			})
			add(path, *expectedStyle.StrokeWidth, actual, ok && math.Abs(actual-*expectedStyle.StrokeWidth) < 1e-9)
		}
		if expectedStyle.Justification != "" {
			path := stylePath + ".justification"
			actual, ok := profileParagraphJustification(layer.Text.Paragraphs, expectedStyle.ParagraphIndex)
			add(path, expectedStyle.Justification, actual, ok && strings.EqualFold(actual, expectedStyle.Justification))
		}
	}
	for i, expectedKeyframes := range expected.Keyframes {
		kfPropPath := fmt.Sprintf("expected_profile.keyframes[%d]", i)
		prop := findProfileLayerProperty(prof, expectedKeyframes.LayerName, expectedKeyframes.MatchName)
		if prop == nil {
			add(kfPropPath, expectedKeyframes.MatchName, nil, false)
			continue
		}
		add(kfPropPath+".count", len(expectedKeyframes.Keyframes), len(prop.Keyframes), len(prop.Keyframes) == len(expectedKeyframes.Keyframes))
		for ki, expectedKF := range expectedKeyframes.Keyframes {
			kfPath := fmt.Sprintf("%s.keyframes[%d]", kfPropPath, ki)
			if ki >= len(prop.Keyframes) {
				add(kfPath, expectedKF.Value, nil, false)
				continue
			}
			actualKF := prop.Keyframes[ki]
			timeOK := math.Abs(actualKF.Time-expectedKF.Time) < 1e-6
			valueOK := profileValueEqual(expectedKF.Value, actualKF.Value)
			add(kfPath, expectedKF.Value, actualKF.Value, timeOK && valueOK)
			if !timeOK {
				add(kfPath+".time", expectedKF.Time, actualKF.Time, false)
			}
		}
	}
	for i, expectedMask := range expected.Masks {
		maskPath := fmt.Sprintf("expected_profile.masks[%d]", i)
		mask := findProfileMask(prof, expectedMask.LayerName, expectedMask.Name)
		add(maskPath, expectedMask.Name, profileMaskName(mask), mask != nil)
		if mask == nil {
			continue
		}
		if expectedMask.Name != "" {
			add(maskPath+".name", expectedMask.Name, mask.Name, mask.Name == expectedMask.Name)
		}
		if expectedMask.Mode != "" {
			add(maskPath+".mode", expectedMask.Mode, mask.Mode, mask.Mode == expectedMask.Mode)
		}
		if expectedMask.Inverted != nil {
			add(maskPath+".inverted", *expectedMask.Inverted, mask.Inverted, mask.Inverted == *expectedMask.Inverted)
		}
		if expectedMask.Locked != nil {
			add(maskPath+".locked", *expectedMask.Locked, mask.Locked, mask.Locked == *expectedMask.Locked)
		}
		if len(expectedMask.Color) > 0 {
			add(maskPath+".color", expectedMask.Color, mask.Color, profileValueEqual(expectedMask.Color, mask.Color))
		}
		if expectedMask.MotionBlur != "" {
			add(maskPath+".motion_blur", expectedMask.MotionBlur, mask.MotionBlur, mask.MotionBlur == expectedMask.MotionBlur)
		}
		if expectedMask.FeatherFalloff != "" {
			add(maskPath+".feather_falloff", expectedMask.FeatherFalloff, mask.FeatherFalloff, mask.FeatherFalloff == expectedMask.FeatherFalloff)
		}
		if expectedMask.Opacity != nil {
			add(maskPath+".opacity", *expectedMask.Opacity, mask.Opacity, math.Abs(mask.Opacity-*expectedMask.Opacity) < 1e-9)
		}
		if len(expectedMask.Feather) > 0 {
			add(maskPath+".feather", expectedMask.Feather, mask.Feather, profileValueEqual(expectedMask.Feather, mask.Feather))
		}
		if expectedMask.Expansion != nil {
			add(maskPath+".expansion", *expectedMask.Expansion, mask.Expansion, math.Abs(mask.Expansion-*expectedMask.Expansion) < 1e-9)
		}
		if expectedMask.Closed != nil {
			add(maskPath+".closed", *expectedMask.Closed, mask.Closed, mask.Closed == *expectedMask.Closed)
		}
		if expectedMask.VertexCount != nil {
			add(maskPath+".vertex_count", *expectedMask.VertexCount, len(mask.Vertices), len(mask.Vertices) == *expectedMask.VertexCount)
		}
		if len(expectedMask.PathKeyframes) > 0 {
			add(maskPath+".path_keyframes.count", len(expectedMask.PathKeyframes), len(mask.PathKeyframes), len(mask.PathKeyframes) == len(expectedMask.PathKeyframes))
		}
		for ki, expectedKF := range expectedMask.PathKeyframes {
			kfPath := fmt.Sprintf("%s.path_keyframes[%d]", maskPath, ki)
			if ki >= len(mask.PathKeyframes) {
				add(kfPath, expectedKF.Time, nil, false)
				continue
			}
			actualKF := mask.PathKeyframes[ki]
			timeOK := math.Abs(actualKF.Time-expectedKF.Time) < 1e-6
			vertexCountOK := expectedKF.VertexCount == nil || len(actualKF.Vertices) == *expectedKF.VertexCount
			add(kfPath, expectedKF.Time, actualKF.Time, timeOK && vertexCountOK)
			if expectedKF.VertexCount != nil {
				add(kfPath+".vertex_count", *expectedKF.VertexCount, len(actualKF.Vertices), len(actualKF.Vertices) == *expectedKF.VertexCount)
			}
		}
	}
	return checks
}

func checkExpectedMotionBlur(expected *ExpectedMotionBlurSpec, prof *profile.Profile, add func(string, any, any, bool)) {
	if expected == nil {
		return
	}
	if len(prof.Comps) == 0 {
		add("expected_profile.motion_blur", "composition", nil, false)
		return
	}
	actual := prof.Comps[0].MotionBlur
	if expected.Enabled != nil {
		add("expected_profile.motion_blur.enabled", *expected.Enabled, actual.Enabled, actual.Enabled == *expected.Enabled)
	}
	if expected.ShutterAngle != nil {
		got := float64(actual.ShutterAngle)
		add("expected_profile.motion_blur.shutter_angle", *expected.ShutterAngle, got, got == *expected.ShutterAngle)
	}
	if expected.ShutterPhase != nil {
		got := float64(actual.ShutterPhase)
		add("expected_profile.motion_blur.shutter_phase", *expected.ShutterPhase, got, got == *expected.ShutterPhase)
	}
	if expected.AdaptiveSampleLimit != nil {
		got := float64(actual.AdaptiveSampleLimit)
		add("expected_profile.motion_blur.adaptive_sample_limit", *expected.AdaptiveSampleLimit, got, got == *expected.AdaptiveSampleLimit)
	}
	if expected.SamplesPerFrame != nil {
		got := float64(actual.SamplesPerFrame)
		add("expected_profile.motion_blur.samples_per_frame", *expected.SamplesPerFrame, got, got == *expected.SamplesPerFrame)
	}
}

func checkExpectedWorkArea(expected *ExpectedWorkAreaSpec, prof *profile.Profile, add func(string, any, any, bool)) {
	if expected == nil {
		return
	}
	if len(prof.Comps) == 0 {
		add("expected_profile.work_area", "composition", nil, false)
		return
	}
	actual := prof.Comps[0].WorkArea
	if expected.Start != nil {
		add("expected_profile.work_area.start", *expected.Start, actual.Start, math.Abs(actual.Start-*expected.Start) < 1e-6)
	}
	if expected.End != nil {
		add("expected_profile.work_area.end", *expected.End, actual.End, math.Abs(actual.End-*expected.End) < 1e-6)
	}
}

func countProfileLayers(prof *profile.Profile, include func(profile.Layer) bool) int {
	var count int
	for _, comp := range prof.Comps {
		for _, layer := range comp.Layers {
			if include(layer) {
				count++
			}
		}
	}
	return count
}

func findProfileEffect(prof *profile.Profile, layerName, matchName string) *profile.Effect {
	layer := findProfileLayer(prof, layerName)
	if layer == nil {
		return nil
	}
	for i := range layer.Effects {
		if layer.Effects[i].MatchName == matchName {
			return &layer.Effects[i]
		}
	}
	return nil
}

func findProfileLayer(prof *profile.Profile, layerName string) *profile.Layer {
	for _, comp := range prof.Comps {
		for i := range comp.Layers {
			if comp.Layers[i].Name == layerName {
				return &comp.Layers[i]
			}
		}
	}
	return nil
}

func findProfileMask(prof *profile.Profile, layerName, maskName string) *profile.Mask {
	layer := findProfileLayer(prof, layerName)
	if layer == nil {
		return nil
	}
	for i := range layer.Masks {
		if maskName == "" || layer.Masks[i].Name == maskName {
			return &layer.Masks[i]
		}
	}
	return nil
}

func checkExpectedLayer(index int, expected ExpectedLayer, prof *profile.Profile, add func(string, any, any, bool)) {
	layerPath := fmt.Sprintf("expected_profile.layers[%d]", index)
	layer := findProfileLayer(prof, expected.Name)
	add(layerPath, expected.Name, profileLayerName(layer), layer != nil)
	if layer == nil {
		return
	}
	if expected.Name != "" {
		add(layerPath+".name", expected.Name, layer.Name, layer.Name == expected.Name)
	}
	if expected.Type != "" {
		add(layerPath+".type", expected.Type, layer.Type, layer.Type == expected.Type)
	}
	if expected.Quality != "" {
		add(layerPath+".quality", expected.Quality, layer.Quality, layer.Quality == expected.Quality)
	}
	if expected.BlendingMode != "" {
		add(layerPath+".blending_mode", expected.BlendingMode, layer.BlendingMode, layer.BlendingMode == expected.BlendingMode)
	}
	if expected.AutoOrient != "" {
		add(layerPath+".auto_orient", expected.AutoOrient, layer.AutoOrient, layer.AutoOrient == expected.AutoOrient)
	}
	if expected.LightKind != "" {
		add(layerPath+".light_kind", expected.LightKind, layer.LightKind, layer.LightKind == expected.LightKind)
	}
	if expected.Source != "" {
		actual := ""
		if layer.SourceRef != nil {
			actual = layer.SourceRef.Name
		}
		add(layerPath+".source", expected.Source, actual, actual == expected.Source)
	}
	if expected.SourceKind != "" {
		actual := ""
		if layer.SourceRef != nil {
			actual = layer.SourceRef.Kind
		}
		add(layerPath+".source_kind", expected.SourceKind, actual, actual == expected.SourceKind)
	}
	if expected.LightSource != "" {
		actual := ""
		if layer.LightSourceRef != nil {
			actual = layer.LightSourceRef.Name
		}
		add(layerPath+".light_source", expected.LightSource, actual, actual == expected.LightSource)
	}
	if expected.Parent != "" {
		actual := ""
		if layer.ParentRef != nil {
			actual = layer.ParentRef.Name
		}
		add(layerPath+".parent", expected.Parent, actual, actual == expected.Parent)
	}
	if expected.TrackMatte != "" {
		actual := trackMatteProfileName(layer.Flags.TrackMatteName)
		add(layerPath+".track_matte", expected.TrackMatte, actual, actual == expected.TrackMatte)
	}
	if expected.Matte != "" {
		actual := ""
		if layer.MatteRef != nil {
			actual = layer.MatteRef.Name
		}
		add(layerPath+".matte", expected.Matte, actual, actual == expected.Matte)
	}
	if expected.Label != nil {
		actual := float64(layer.Label)
		add(layerPath+".label", *expected.Label, actual, actual == *expected.Label)
	}
	if expected.Comment != "" {
		add(layerPath+".comment", expected.Comment, layer.Comment, layer.Comment == expected.Comment)
	}
	if expected.Timing != nil {
		checkExpectedLayerTiming(layerPath+".timing", expected.Timing, layer.Timing, add)
	}
	if expected.Flags != nil {
		checkExpectedLayerFlags(layerPath+".flags", expected.Flags, layer.Flags, add)
	}
}

func checkExpectedLayerTiming(path string, expected *ExpectedLayerTiming, actual profile.LayerTiming, add func(string, any, any, bool)) {
	if expected.StartTime != nil {
		add(path+".start_time", *expected.StartTime, actual.StartTime, math.Abs(actual.StartTime-*expected.StartTime) < 1e-6)
	}
	if expected.InPoint != nil {
		add(path+".in_point", *expected.InPoint, actual.InPoint, math.Abs(actual.InPoint-*expected.InPoint) < 1e-6)
	}
	if expected.OutPoint != nil {
		add(path+".out_point", *expected.OutPoint, actual.OutPoint, math.Abs(actual.OutPoint-*expected.OutPoint) < 1e-6)
	}
	if expected.Duration != nil {
		add(path+".duration", *expected.Duration, actual.Duration, math.Abs(actual.Duration-*expected.Duration) < 1e-6)
	}
	if expected.Stretch != nil {
		add(path+".stretch", *expected.Stretch, actual.Stretch, math.Abs(actual.Stretch-*expected.Stretch) < 1e-6)
	}
}

func checkExpectedLayerFlags(path string, expected *ExpectedLayerFlags, actual profile.LayerFlags, add func(string, any, any, bool)) {
	addBool := func(name string, expected *bool, actual bool) {
		if expected != nil {
			add(path+"."+name, *expected, actual, actual == *expected)
		}
	}
	addBool("visible", expected.Visible, actual.Visible)
	addBool("solo", expected.Solo, actual.Solo)
	addBool("shy", expected.Shy, actual.Shy)
	addBool("locked", expected.Locked, actual.Locked)
	addBool("is_3d", expected.Is3D, actual.Is3D)
	addBool("is_adjustment", expected.IsAdjustment, actual.IsAdjustment)
	addBool("is_null", expected.IsNull, actual.IsNull)
	addBool("is_guide", expected.IsGuide, actual.IsGuide)
	addBool("motion_blur", expected.MotionBlur, actual.MotionBlur)
	addBool("effects_enabled", expected.EffectsEnabled, actual.EffectsEnabled)
	addBool("audio_enabled", expected.AudioEnabled, actual.AudioEnabled)
	addBool("frame_blend_enabled", expected.FrameBlendEnabled, actual.FrameBlendEnabled)
	addBool("markers_locked", expected.MarkersLocked, actual.MarkersLocked)
	addBool("frame_blend_pixel_motion", expected.FrameBlendPixelMotion, actual.FrameBlendPixelMotion)
	addBool("collapse_transform", expected.CollapseTransform, actual.CollapseTransform)
	addBool("sampling_bicubic", expected.SamplingBicubic, actual.SamplingBicubic)
	addBool("preserve_transparency", expected.PreserveTransparency, actual.PreserveTransparency)
}

func effectMatchName(effect *profile.Effect) any {
	if effect == nil {
		return nil
	}
	return effect.MatchName
}

func profileLayerName(layer *profile.Layer) any {
	if layer == nil {
		return nil
	}
	return layer.Name
}

func profileMaskName(mask *profile.Mask) any {
	if mask == nil {
		return nil
	}
	return mask.Name
}

func trackMatteProfileName(value string) string {
	switch value {
	case "Alpha":
		return "alpha"
	case "AlphaInverse":
		return "alpha_inverse"
	case "Luma":
		return "luma"
	case "LumaInverse":
		return "luma_inverse"
	default:
		return value
	}
}

func findProfileParam(params []profile.Property, matchName string) *profile.Property {
	for i := range params {
		if params[i].MatchName == matchName {
			return &params[i]
		}
	}
	return nil
}

func findProfileLayerProperty(prof *profile.Profile, layerName, matchName string) *profile.Property {
	layer := findProfileLayer(prof, layerName)
	if layer == nil {
		return nil
	}
	if prop := findProfileParam(layer.Properties, matchName); prop != nil {
		return prop
	}
	for i := range layer.Effects {
		if prop := findProfileParam(layer.Effects[i].Params, matchName); prop != nil {
			return prop
		}
	}
	for i := range layer.Shapes {
		if prop := findProfileParam(layer.Shapes[i].Properties, matchName); prop != nil {
			return prop
		}
	}
	return nil
}

func profileRunFloat(runs []profile.TextStyleRun, index int, value func(profile.TextStyleRun) float64) (float64, bool) {
	if index < 0 || index >= len(runs) {
		return 0, false
	}
	return value(runs[index]), true
}

func profileRunBool(runs []profile.TextStyleRun, index int, value func(profile.TextStyleRun) bool) (bool, bool) {
	if index < 0 || index >= len(runs) {
		return false, false
	}
	return value(runs[index]), true
}

func profileRunColor(runs []profile.TextStyleRun, index int, value func(profile.TextStyleRun) [4]float64) ([]float64, bool) {
	if index < 0 || index >= len(runs) {
		return nil, false
	}
	color := value(runs[index])
	return []float64{color[0], color[1], color[2], color[3]}, true
}

func profileParagraphJustification(paragraphs []profile.TextParagraph, index int) (string, bool) {
	if index < 0 || index >= len(paragraphs) {
		return "", false
	}
	return paragraphs[index].Justification, true
}

func profileValueEqual(expected, actual any) bool {
	expected, err := normalizeEffectParamValue(expected)
	if err != nil {
		return false
	}
	actual, err = normalizeEffectParamValue(actual)
	if err != nil {
		return false
	}
	switch e := expected.(type) {
	case float64:
		a, ok := actual.(float64)
		return ok && math.Abs(e-a) < 1e-9
	case []float64:
		a, ok := actual.([]float64)
		if !ok || len(e) != len(a) {
			return false
		}
		for i := range e {
			if math.Abs(e[i]-a[i]) >= 1e-9 {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func rgbToFloatSlice(rgb [3]uint8) []float64 {
	return []float64{float64(rgb[0]), float64(rgb[1]), float64(rgb[2])}
}

func uint16PairToFloatSlice(pair [2]uint16) []float64 {
	return []float64{float64(pair[0]), float64(pair[1])}
}
