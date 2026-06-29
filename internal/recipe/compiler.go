package recipe

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func CompileToFile(rec Recipe, outPath string, caps CapabilityIndex) (Report, error) {
	report := ValidateWithCapabilities(rec, caps)
	report.OutputPath = outPath
	if !report.Valid {
		return report, nil
	}
	if outPath == "" {
		report.Valid = false
		report.Refusals = append(report.Refusals, Refusal{Code: "missing_output_path", Path: "output_path", Message: "output path is required"})
		return report, nil
	}

	project := aep.NewProject(aep.TargetAE2020)
	compSpec := rec.Comps[0]
	comp, err := aep.NewComposition(project, compSpec.Name, uint16(compSpec.Width), uint16(compSpec.Height), compSpec.FrameRate, compSpec.Duration)
	if err != nil {
		return report, fmt.Errorf("recipe: create comp: %w", err)
	}
	if len(compSpec.BackgroundColor) == 3 {
		if err := comp.SetBGColor(rgb8Color(compSpec.BackgroundColor)); err != nil {
			return report, fmt.Errorf("recipe: comp %q background_color: %w", compSpec.Name, err)
		}
	}
	if compSpec.Renderer != "" {
		if err := aep.SetRenderer(comp, compSpec.Renderer); err != nil {
			return report, fmt.Errorf("recipe: comp %q renderer: %w", compSpec.Name, err)
		}
	}
	if len(compSpec.ResolutionFactor) == 2 {
		if err := comp.SetResolutionFactor(uint16(compSpec.ResolutionFactor[0]), uint16(compSpec.ResolutionFactor[1])); err != nil {
			return report, fmt.Errorf("recipe: comp %q resolution_factor: %w", compSpec.Name, err)
		}
	}
	if compSpec.PixelAspect != nil {
		if err := comp.SetPixelAspect(*compSpec.PixelAspect); err != nil {
			return report, fmt.Errorf("recipe: comp %q pixel_aspect: %w", compSpec.Name, err)
		}
	}
	if compSpec.DisplayStartTime != nil {
		if err := comp.SetDisplayStartTime(*compSpec.DisplayStartTime); err != nil {
			return report, fmt.Errorf("recipe: comp %q display_start_time: %w", compSpec.Name, err)
		}
	}
	if compSpec.FrameBlending != nil {
		if err := comp.SetFrameBlending(*compSpec.FrameBlending); err != nil {
			return report, fmt.Errorf("recipe: comp %q frame_blending: %w", compSpec.Name, err)
		}
	}
	if compSpec.Draft3D != nil {
		if err := comp.SetDraft3D(*compSpec.Draft3D); err != nil {
			return report, fmt.Errorf("recipe: comp %q draft_3d: %w", compSpec.Name, err)
		}
	}
	if compSpec.HideShyLayers != nil {
		if err := comp.SetHideShyLayers(*compSpec.HideShyLayers); err != nil {
			return report, fmt.Errorf("recipe: comp %q hide_shy_layers: %w", compSpec.Name, err)
		}
	}
	if compSpec.PreserveNestedFrameRate != nil {
		if err := comp.SetPreserveNestedFrameRate(*compSpec.PreserveNestedFrameRate); err != nil {
			return report, fmt.Errorf("recipe: comp %q preserve_nested_frame_rate: %w", compSpec.Name, err)
		}
	}
	if compSpec.PreserveNestedResolution != nil {
		if err := comp.SetPreserveNestedResolution(*compSpec.PreserveNestedResolution); err != nil {
			return report, fmt.Errorf("recipe: comp %q preserve_nested_resolution: %w", compSpec.Name, err)
		}
	}
	if compSpec.MotionBlur != nil {
		if err := applyCompMotionBlur(comp, compSpec.MotionBlur); err != nil {
			return report, fmt.Errorf("recipe: comp %q motion_blur: %w", compSpec.Name, err)
		}
	}
	if compSpec.WorkArea != nil {
		if err := comp.SetWorkArea(*compSpec.WorkArea.Start, *compSpec.WorkArea.End); err != nil {
			return report, fmt.Errorf("recipe: comp %q work_area: %w", compSpec.Name, err)
		}
	}
	layersByName := map[string]*aep.Layer{}
	for _, layerSpec := range compSpec.Layers {
		layer, err := compileLayer(comp, layerSpec, compSpec)
		if err != nil {
			return report, err
		}
		if layerSpec.Name != "" {
			layersByName[layerSpec.Name] = layer
		}
	}
	if err := applyLayerParents(compSpec, layersByName); err != nil {
		return report, err
	}
	if err := applyLightSources(compSpec, layersByName); err != nil {
		return report, err
	}
	if hasEffects(compSpec) {
		project, err = materializeEffects(project, compSpec)
		if err != nil {
			return report, err
		}
	}
	if hasTransformExpressions(compSpec) {
		project, err = materializeTransformExpressions(project, compSpec)
		if err != nil {
			return report, err
		}
	}
	project, err = applyCompItemSettings(project, compSpec)
	if err != nil {
		return report, fmt.Errorf("recipe: comp %q item settings: %w", compSpec.Name, err)
	}
	if hasExpectedProfile(rec.ExpectedProfile) {
		prof, err := buildWrittenProfile(project, outPath)
		if err != nil {
			return report, fmt.Errorf("recipe: build profile for expected_profile: %w", err)
		}
		report.ProfileChecks = checkExpectedProfile(rec.ExpectedProfile, prof)
		for _, check := range report.ProfileChecks {
			if !check.Passed {
				report.Valid = false
				report.Refusals = append(report.Refusals, Refusal{
					Code:    "profile_contract_mismatch",
					Path:    check.Path,
					Message: check.Message,
				})
			}
		}
		if !report.Valid {
			return report, nil
		}
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return report, err
	}
	out, err := os.Create(outPath)
	if err != nil {
		return report, err
	}
	defer out.Close()
	if err := project.WriteAEP(out); err != nil {
		return report, err
	}
	return report, nil
}

func buildWrittenProfile(project *aep.Project, outPath string) (*profile.Profile, error) {
	var buf bytes.Buffer
	if err := project.WriteAEP(&buf); err != nil {
		return nil, err
	}
	reopened, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return nil, err
	}
	return profile.Build(reopened, profile.Options{Path: outPath})
}

func hasExpectedProfile(expected ExpectedProfile) bool {
	return expected.CompCount != nil ||
		expected.LayerCount != nil ||
		expected.TextLayerCount != nil ||
		expected.ShapeLayerCount != nil ||
		expected.Label != nil ||
		expected.Comment != "" ||
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
		len(expected.Layers) > 0 ||
		len(expected.Effects) > 0 ||
		len(expected.Properties) > 0 ||
		len(expected.TextStyles) > 0 ||
		len(expected.Keyframes) > 0
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
			if expectedParam.Expression != "" {
				add(paramPath+".expression", expectedParam.Expression, param.Expression, param.Expression == expectedParam.Expression)
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
		if expectedStyle.Tracking != nil {
			path := stylePath + ".tracking"
			actual, ok := profileRunFloat(layer.Text.Runs, expectedStyle.RunIndex, func(run profile.TextStyleRun) float64 {
				return run.Tracking
			})
			add(path, *expectedStyle.Tracking, actual, ok && math.Abs(actual-*expectedStyle.Tracking) < 1e-9)
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

func hasEffects(comp CompSpec) bool {
	for _, layer := range comp.Layers {
		if len(layer.Effects) > 0 {
			return true
		}
	}
	return false
}

func hasTransformExpressions(comp CompSpec) bool {
	for _, layer := range comp.Layers {
		if hasLayerTransformExpressions(layer.Transform.Expressions) {
			return true
		}
	}
	return false
}

func hasLayerTransformExpressions(expressions TransformExpressions) bool {
	return expressions.Position != nil ||
		expressions.AnchorPoint != nil ||
		expressions.Scale != nil ||
		expressions.Rotation != nil ||
		expressions.Opacity != nil
}

func materializeEffects(project *aep.Project, compSpec CompSpec) (*aep.Project, error) {
	reopened, err := aep.Reopen(project)
	if err != nil {
		return nil, fmt.Errorf("recipe: reopen for effects: %w", err)
	}
	if len(reopened.Compositions) == 0 {
		return nil, fmt.Errorf("recipe: reopen for effects: no compositions")
	}
	comp := reopened.Compositions[0]
	for i, layerSpec := range compSpec.Layers {
		if len(layerSpec.Effects) == 0 {
			continue
		}
		if i >= len(comp.Layers) {
			return nil, fmt.Errorf("recipe: reopen for effects: layer index %d missing", i)
		}
		layer := comp.Layers[i]
		for _, effect := range layerSpec.Effects {
			fx, err := aep.AddEffect(layer, effect.MatchName)
			if err != nil {
				return nil, fmt.Errorf("recipe: layer %q add effect %q: %w", layerSpec.Name, effect.MatchName, err)
			}
			for _, param := range effect.Params {
				value, err := normalizeEffectParamValue(param.Value)
				if err != nil {
					return nil, fmt.Errorf("recipe: layer %q effect %q param %q: %w", layerSpec.Name, effect.MatchName, param.MatchName, err)
				}
				property, err := aep.SetEffectParam(layer, fx, param.MatchName, value)
				if err != nil {
					return nil, fmt.Errorf("recipe: layer %q effect %q param %q: %w", layerSpec.Name, effect.MatchName, param.MatchName, err)
				}
				if param.Expression != nil {
					if err := applyPropertyExpression(property, *param.Expression); err != nil {
						return nil, fmt.Errorf("recipe: layer %q effect %q param %q expression: %w", layerSpec.Name, effect.MatchName, param.MatchName, err)
					}
				}
			}
		}
	}
	return reopened, nil
}

func materializeTransformExpressions(project *aep.Project, compSpec CompSpec) (*aep.Project, error) {
	reopened, err := aep.Reopen(project)
	if err != nil {
		return nil, fmt.Errorf("recipe: reopen for expressions: %w", err)
	}
	if len(reopened.Compositions) == 0 {
		return nil, fmt.Errorf("recipe: reopen for expressions: no compositions")
	}
	comp := reopened.Compositions[0]
	for i, layerSpec := range compSpec.Layers {
		if !hasLayerTransformExpressions(layerSpec.Transform.Expressions) {
			continue
		}
		if i >= len(comp.Layers) {
			return nil, fmt.Errorf("recipe: reopen for expressions: layer index %d missing", i)
		}
		if err := applyTransformExpressions(comp.Layers[i], layerSpec.Transform.Expressions); err != nil {
			return nil, fmt.Errorf("recipe: layer %q transform.expressions: %w", layerSpec.Name, err)
		}
	}
	return reopened, nil
}

func applyTransformExpressions(layer *aep.Layer, expressions TransformExpressions) error {
	if err := applyTransformExpression(layer, "position", expressions.Position); err != nil {
		return err
	}
	if err := applyTransformExpression(layer, "anchor_point", expressions.AnchorPoint); err != nil {
		return err
	}
	if err := applyTransformExpression(layer, "scale", expressions.Scale); err != nil {
		return err
	}
	if err := applyTransformExpression(layer, "rotation", expressions.Rotation); err != nil {
		return err
	}
	if err := applyTransformExpression(layer, "opacity", expressions.Opacity); err != nil {
		return err
	}
	return nil
}

func applyPropertyExpression(property *aep.Property, expression ExpressionSpec) error {
	if property == nil {
		return fmt.Errorf("property missing")
	}
	if err := property.SetExpression(expression.Source); err != nil {
		return fmt.Errorf("source: %w", err)
	}
	if expression.Enabled != nil {
		if err := property.SetExpressionEnabled(*expression.Enabled); err != nil {
			return fmt.Errorf("enabled: %w", err)
		}
	}
	return nil
}

func applyTransformExpression(layer *aep.Layer, name string, expression *ExpressionSpec) error {
	if expression == nil {
		return nil
	}
	property, err := transformExpressionProperty(layer, name)
	if err != nil {
		return err
	}
	if err := applyPropertyExpression(property, *expression); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

func transformExpressionProperty(layer *aep.Layer, name string) (*aep.Property, error) {
	switch name {
	case "position":
		return layer.Position(), nil
	case "anchor_point":
		return layer.AnchorPoint(), nil
	case "scale":
		return layer.Scale(), nil
	case "rotation":
		return layer.Rotation(), nil
	case "opacity":
		return layer.Opacity(), nil
	default:
		return nil, fmt.Errorf("unsupported transform expression property %q", name)
	}
}

func applyCompItemSettings(project *aep.Project, compSpec CompSpec) (*aep.Project, error) {
	if compSpec.Label == nil && compSpec.Comment == "" {
		return project, nil
	}
	reopened, err := aep.Reopen(project)
	if err != nil {
		return nil, fmt.Errorf("reopen: %w", err)
	}
	comp := findProjectComp(reopened, compSpec.Name)
	if comp == nil {
		return nil, fmt.Errorf("comp %q not found after reopen", compSpec.Name)
	}
	if compSpec.Label != nil {
		if err := comp.SetLabel(uint8(*compSpec.Label)); err != nil {
			return nil, fmt.Errorf("label: %w", err)
		}
	}
	if compSpec.Comment != "" {
		if err := comp.SetComment(compSpec.Comment); err != nil {
			return nil, fmt.Errorf("comment: %w", err)
		}
	}
	return reopened, nil
}

func findProjectComp(project *aep.Project, name string) *aep.Composition {
	for _, comp := range project.Compositions {
		if comp.Name == name {
			return comp
		}
	}
	return nil
}

func applyCompMotionBlur(comp *aep.Composition, spec *CompMotionBlurSpec) error {
	if spec == nil {
		return nil
	}
	if spec.Enabled != nil {
		if err := comp.SetCompMotionBlur(*spec.Enabled); err != nil {
			return err
		}
	}
	if spec.ShutterAngle != nil {
		if err := comp.SetShutterAngle(uint16(*spec.ShutterAngle)); err != nil {
			return err
		}
	}
	if spec.ShutterPhase != nil {
		if err := comp.SetShutterPhase(int32(*spec.ShutterPhase)); err != nil {
			return err
		}
	}
	if spec.AdaptiveSampleLimit != nil {
		if err := comp.SetMotionBlurAdaptiveSampleLimit(int32(*spec.AdaptiveSampleLimit)); err != nil {
			return err
		}
	}
	if spec.SamplesPerFrame != nil {
		if err := comp.SetMotionBlurSamplesPerFrame(int32(*spec.SamplesPerFrame)); err != nil {
			return err
		}
	}
	return nil
}

func normalizeEffectParamValue(value any) (any, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case bool:
		if v {
			return 1.0, nil
		}
		return 0.0, nil
	case []float64:
		if len(v) == 0 {
			return nil, fmt.Errorf("empty numeric array")
		}
		return v, nil
	case []any:
		if len(v) == 0 {
			return nil, fmt.Errorf("empty numeric array")
		}
		out := make([]float64, 0, len(v))
		for i, item := range v {
			n, ok := item.(float64)
			if !ok {
				return nil, fmt.Errorf("array item %d is %T, want number", i, item)
			}
			out = append(out, n)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported value type %T", value)
	}
}

func compileLayer(comp *aep.Composition, spec Layer, compSpec CompSpec) (*aep.Layer, error) {
	var layer *aep.Layer
	switch spec.Type {
	case "text":
		l, err := aep.NewTextLayer(comp, spec.Name)
		if err != nil {
			return nil, fmt.Errorf("recipe: text layer %q: %w", spec.Name, err)
		}
		if spec.Text != "" {
			if err := l.SetText(spec.Text); err != nil {
				return nil, fmt.Errorf("recipe: text layer %q set text: %w", spec.Name, err)
			}
		}
		if spec.TextStyle != nil {
			if err := applyTextStyle(l, *spec.TextStyle); err != nil {
				return nil, fmt.Errorf("recipe: text layer %q style: %w", spec.Name, err)
			}
		}
		layer = l
	case "shape":
		l, err := aep.NewShapeLayer(comp, spec.Name)
		if err != nil {
			return nil, fmt.Errorf("recipe: shape layer %q: %w", spec.Name, err)
		}
		if spec.Shape != nil {
			if err := compileShape(l.RootGroup(), *spec.Shape); err != nil {
				return nil, fmt.Errorf("recipe: shape layer %q: %w", spec.Name, err)
			}
		}
		layer = l.Layer
	case "solid":
		color := [3]float64{0, 0, 0}
		if spec.Shape != nil && len(spec.Shape.FillColor) >= 3 {
			color = [3]float64{toUnitColor(spec.Shape.FillColor[0]), toUnitColor(spec.Shape.FillColor[1]), toUnitColor(spec.Shape.FillColor[2])}
		}
		l, err := aep.NewSolidLayer(comp, spec.Name, compSpec.Width, compSpec.Height, color)
		if err != nil {
			return nil, fmt.Errorf("recipe: solid layer %q: %w", spec.Name, err)
		}
		layer = l
	case "null":
		l, err := aep.NewNullLayer(comp, spec.Name)
		if err != nil {
			return nil, fmt.Errorf("recipe: null layer %q: %w", spec.Name, err)
		}
		layer = l
	case "adjustment":
		l, err := aep.NewAdjustmentLayer(comp, spec.Name)
		if err != nil {
			return nil, fmt.Errorf("recipe: adjustment layer %q: %w", spec.Name, err)
		}
		layer = l
	case "camera":
		l, err := aep.NewCameraLayer(comp, spec.Name)
		if err != nil {
			return nil, fmt.Errorf("recipe: camera layer %q: %w", spec.Name, err)
		}
		layer = l
	case "light":
		l, err := aep.NewLightLayer(comp, spec.Name)
		if err != nil {
			return nil, fmt.Errorf("recipe: light layer %q: %w", spec.Name, err)
		}
		layer = l
	default:
		return nil, fmt.Errorf("recipe: unsupported layer type %q", spec.Type)
	}
	if layer != nil {
		if spec.Label != nil {
			if err := layer.SetLabel(uint8(*spec.Label)); err != nil {
				return nil, fmt.Errorf("recipe: layer %q label: %w", spec.Name, err)
			}
		}
		if spec.Comment != "" {
			if err := layer.SetComment(spec.Comment); err != nil {
				return nil, fmt.Errorf("recipe: layer %q comment: %w", spec.Name, err)
			}
		}
		if spec.Visible != nil {
			if err := layer.SetVisible(*spec.Visible); err != nil {
				return nil, fmt.Errorf("recipe: layer %q visible: %w", spec.Name, err)
			}
		}
		if spec.Solo != nil {
			if err := layer.SetSolo(*spec.Solo); err != nil {
				return nil, fmt.Errorf("recipe: layer %q solo: %w", spec.Name, err)
			}
		}
		if spec.Locked != nil {
			if err := layer.SetLocked(*spec.Locked); err != nil {
				return nil, fmt.Errorf("recipe: layer %q locked: %w", spec.Name, err)
			}
		}
		if spec.MotionBlur != nil {
			if err := layer.SetMotionBlur(*spec.MotionBlur); err != nil {
				return nil, fmt.Errorf("recipe: layer %q motion_blur: %w", spec.Name, err)
			}
		}
		if spec.Shy != nil {
			if err := layer.SetShy(*spec.Shy); err != nil {
				return nil, fmt.Errorf("recipe: layer %q shy: %w", spec.Name, err)
			}
		}
		if spec.EffectsEnabled != nil {
			if err := layer.SetEffectsEnabled(*spec.EffectsEnabled); err != nil {
				return nil, fmt.Errorf("recipe: layer %q effects_enabled: %w", spec.Name, err)
			}
		}
		if spec.AudioEnabled != nil {
			if err := layer.SetAudioEnabled(*spec.AudioEnabled); err != nil {
				return nil, fmt.Errorf("recipe: layer %q audio_enabled: %w", spec.Name, err)
			}
		}
		if spec.FrameBlendEnabled != nil {
			if err := layer.SetFrameBlendEnabled(*spec.FrameBlendEnabled); err != nil {
				return nil, fmt.Errorf("recipe: layer %q frame_blend_enabled: %w", spec.Name, err)
			}
		}
		if spec.CollapseTransform != nil {
			if err := layer.SetCollapseTransform(*spec.CollapseTransform); err != nil {
				return nil, fmt.Errorf("recipe: layer %q collapse_transform: %w", spec.Name, err)
			}
		}
		if spec.Is3D != nil {
			if err := layer.SetIs3D(*spec.Is3D); err != nil {
				return nil, fmt.Errorf("recipe: layer %q is_3d: %w", spec.Name, err)
			}
		}
		if spec.IsAdjust != nil {
			if err := layer.SetIsAdjust(*spec.IsAdjust); err != nil {
				return nil, fmt.Errorf("recipe: layer %q is_adjust: %w", spec.Name, err)
			}
		}
		if spec.IsNull != nil {
			if err := layer.SetIsNull(*spec.IsNull); err != nil {
				return nil, fmt.Errorf("recipe: layer %q is_null: %w", spec.Name, err)
			}
		}
		if spec.IsGuide != nil {
			if err := layer.SetIsGuide(*spec.IsGuide); err != nil {
				return nil, fmt.Errorf("recipe: layer %q is_guide: %w", spec.Name, err)
			}
		}
		if spec.SamplingBicubic != nil {
			if err := layer.SetSamplingBicubic(*spec.SamplingBicubic); err != nil {
				return nil, fmt.Errorf("recipe: layer %q sampling_bicubic: %w", spec.Name, err)
			}
		}
		if spec.FrameBlendPixelMotion != nil {
			if err := layer.SetFrameBlendPixelMotion(*spec.FrameBlendPixelMotion); err != nil {
				return nil, fmt.Errorf("recipe: layer %q frame_blend_pixel_motion: %w", spec.Name, err)
			}
		}
		if spec.PreserveTransparency != nil {
			if err := layer.SetPreserveTransparency(*spec.PreserveTransparency); err != nil {
				return nil, fmt.Errorf("recipe: layer %q preserve_transparency: %w", spec.Name, err)
			}
		}
		if spec.Quality != "" {
			quality, err := layerQuality(spec.Quality)
			if err != nil {
				return nil, fmt.Errorf("recipe: layer %q quality: %w", spec.Name, err)
			}
			if err := layer.SetQuality(quality); err != nil {
				return nil, fmt.Errorf("recipe: layer %q quality: %w", spec.Name, err)
			}
		}
		if spec.BlendingMode != "" {
			blendingMode, err := layerBlendingMode(spec.BlendingMode)
			if err != nil {
				return nil, fmt.Errorf("recipe: layer %q blending_mode: %w", spec.Name, err)
			}
			if err := layer.SetBlendingMode(blendingMode); err != nil {
				return nil, fmt.Errorf("recipe: layer %q blending_mode: %w", spec.Name, err)
			}
		}
		if spec.TrackMatte != "" {
			trackMatte, err := layerTrackMatte(spec.TrackMatte)
			if err != nil {
				return nil, fmt.Errorf("recipe: layer %q track_matte: %w", spec.Name, err)
			}
			if err := layer.SetTrackMatte(trackMatte); err != nil {
				return nil, fmt.Errorf("recipe: layer %q track_matte: %w", spec.Name, err)
			}
		}
		if spec.AutoOrient != "" {
			autoOrient, err := layerAutoOrient(spec.AutoOrient)
			if err != nil {
				return nil, fmt.Errorf("recipe: layer %q auto_orient: %w", spec.Name, err)
			}
			if err := layer.SetAutoOrient(autoOrient); err != nil {
				return nil, fmt.Errorf("recipe: layer %q auto_orient: %w", spec.Name, err)
			}
		}
		if spec.Camera != nil && spec.Camera.Zoom != nil {
			if err := layer.SetCameraZoom(*spec.Camera.Zoom); err != nil {
				return nil, fmt.Errorf("recipe: layer %q camera.zoom: %w", spec.Name, err)
			}
		}
		if spec.Camera != nil && spec.Camera.DepthOfField != nil {
			if err := layer.SetCameraDepthOfField(*spec.Camera.DepthOfField); err != nil {
				return nil, fmt.Errorf("recipe: layer %q camera.depth_of_field: %w", spec.Name, err)
			}
		}
		if spec.Camera != nil && spec.Camera.FocusDistance != nil {
			if err := layer.SetCameraFocusDistance(*spec.Camera.FocusDistance); err != nil {
				return nil, fmt.Errorf("recipe: layer %q camera.focus_distance: %w", spec.Name, err)
			}
		}
		if spec.Camera != nil && spec.Camera.Aperture != nil {
			if err := layer.SetCameraAperture(*spec.Camera.Aperture); err != nil {
				return nil, fmt.Errorf("recipe: layer %q camera.aperture: %w", spec.Name, err)
			}
		}
		if spec.Camera != nil && spec.Camera.BlurLevel != nil {
			if err := layer.SetCameraBlurLevel(*spec.Camera.BlurLevel); err != nil {
				return nil, fmt.Errorf("recipe: layer %q camera.blur_level: %w", spec.Name, err)
			}
		}
		if spec.Camera != nil && spec.Camera.IrisShape != nil {
			if err := layer.SetIrisShape(*spec.Camera.IrisShape); err != nil {
				return nil, fmt.Errorf("recipe: layer %q camera.iris_shape: %w", spec.Name, err)
			}
		}
		if spec.Camera != nil && spec.Camera.IrisRotation != nil {
			if err := layer.SetIrisRotation(*spec.Camera.IrisRotation); err != nil {
				return nil, fmt.Errorf("recipe: layer %q camera.iris_rotation: %w", spec.Name, err)
			}
		}
		if spec.Camera != nil && spec.Camera.IrisRoundness != nil {
			if err := layer.SetIrisRoundness(*spec.Camera.IrisRoundness); err != nil {
				return nil, fmt.Errorf("recipe: layer %q camera.iris_roundness: %w", spec.Name, err)
			}
		}
		if spec.Camera != nil && spec.Camera.IrisAspectRatio != nil {
			if err := layer.SetIrisAspectRatio(*spec.Camera.IrisAspectRatio); err != nil {
				return nil, fmt.Errorf("recipe: layer %q camera.iris_aspect_ratio: %w", spec.Name, err)
			}
		}
		if spec.Camera != nil && spec.Camera.IrisDiffractionFringe != nil {
			if err := layer.SetIrisDiffractionFringe(*spec.Camera.IrisDiffractionFringe); err != nil {
				return nil, fmt.Errorf("recipe: layer %q camera.iris_diffraction_fringe: %w", spec.Name, err)
			}
		}
		if spec.Camera != nil && spec.Camera.IrisHighlightGain != nil {
			if err := layer.SetIrisHighlightGain(*spec.Camera.IrisHighlightGain); err != nil {
				return nil, fmt.Errorf("recipe: layer %q camera.iris_highlight_gain: %w", spec.Name, err)
			}
		}
		if spec.Camera != nil && spec.Camera.IrisHighlightThreshold != nil {
			if err := layer.SetIrisHighlightThreshold(*spec.Camera.IrisHighlightThreshold); err != nil {
				return nil, fmt.Errorf("recipe: layer %q camera.iris_highlight_threshold: %w", spec.Name, err)
			}
		}
		if spec.Camera != nil && spec.Camera.IrisHighlightSaturation != nil {
			if err := layer.SetIrisHighlightSaturation(*spec.Camera.IrisHighlightSaturation); err != nil {
				return nil, fmt.Errorf("recipe: layer %q camera.iris_highlight_saturation: %w", spec.Name, err)
			}
		}
		if spec.Light != nil && spec.Light.Kind != "" {
			kind, err := lightKind(spec.Light.Kind)
			if err != nil {
				return nil, fmt.Errorf("recipe: layer %q light.kind: %w", spec.Name, err)
			}
			if err := layer.SetLightKind(kind); err != nil {
				return nil, fmt.Errorf("recipe: layer %q light.kind: %w", spec.Name, err)
			}
		}
		if spec.Light != nil && spec.Light.Intensity != nil {
			if err := layer.SetLightIntensity(*spec.Light.Intensity); err != nil {
				return nil, fmt.Errorf("recipe: layer %q light.intensity: %w", spec.Name, err)
			}
		}
		if spec.Light != nil && len(spec.Light.Color) > 0 {
			if err := layer.SetLightColor(spec.Light.Color); err != nil {
				return nil, fmt.Errorf("recipe: layer %q light.color: %w", spec.Name, err)
			}
		}
		if spec.Light != nil && spec.Light.CastsShadows != nil {
			if err := layer.SetLightCastsShadows(*spec.Light.CastsShadows); err != nil {
				return nil, fmt.Errorf("recipe: layer %q light.casts_shadows: %w", spec.Name, err)
			}
		}
		if spec.Light != nil && spec.Light.ShadowDarkness != nil {
			if err := layer.SetLightShadowDarkness(*spec.Light.ShadowDarkness); err != nil {
				return nil, fmt.Errorf("recipe: layer %q light.shadow_darkness: %w", spec.Name, err)
			}
		}
		if spec.Light != nil && spec.Light.ShadowDiffusion != nil {
			if err := layer.SetLightShadowDiffusion(*spec.Light.ShadowDiffusion); err != nil {
				return nil, fmt.Errorf("recipe: layer %q light.shadow_diffusion: %w", spec.Name, err)
			}
		}
		if spec.Light != nil && spec.Light.FalloffType != nil {
			if err := layer.SetLightFalloffType(*spec.Light.FalloffType); err != nil {
				return nil, fmt.Errorf("recipe: layer %q light.falloff_type: %w", spec.Name, err)
			}
		}
		if spec.Light != nil && spec.Light.FalloffStart != nil {
			if err := layer.SetLightFalloffStart(*spec.Light.FalloffStart); err != nil {
				return nil, fmt.Errorf("recipe: layer %q light.falloff_start: %w", spec.Name, err)
			}
		}
		if spec.Light != nil && spec.Light.FalloffDistance != nil {
			if err := layer.SetLightFalloffDistance(*spec.Light.FalloffDistance); err != nil {
				return nil, fmt.Errorf("recipe: layer %q light.falloff_distance: %w", spec.Name, err)
			}
		}
		if spec.Light != nil && spec.Light.ConeAngle != nil {
			if err := layer.SetLightConeAngle(*spec.Light.ConeAngle); err != nil {
				return nil, fmt.Errorf("recipe: layer %q light.cone_angle: %w", spec.Name, err)
			}
		}
		if spec.Light != nil && spec.Light.ConeFeather != nil {
			if err := layer.SetLightConeFeather(*spec.Light.ConeFeather); err != nil {
				return nil, fmt.Errorf("recipe: layer %q light.cone_feather: %w", spec.Name, err)
			}
		}
		if spec.StartTime != nil {
			if err := layer.SetStartTime(*spec.StartTime); err != nil {
				return nil, fmt.Errorf("recipe: layer %q start_time: %w", spec.Name, err)
			}
		}
		if spec.InPoint != nil {
			if err := layer.SetInPoint(*spec.InPoint); err != nil {
				return nil, fmt.Errorf("recipe: layer %q in_point: %w", spec.Name, err)
			}
		}
		if spec.OutPoint != nil {
			if err := layer.SetOutPoint(*spec.OutPoint); err != nil {
				return nil, fmt.Errorf("recipe: layer %q out_point: %w", spec.Name, err)
			}
		}
		if err := applyTransform(layer, spec.Transform); err != nil {
			return nil, fmt.Errorf("recipe: layer %q transform: %w", spec.Name, err)
		}
	}
	return layer, nil
}

func applyLayerParents(compSpec CompSpec, layersByName map[string]*aep.Layer) error {
	for _, layerSpec := range compSpec.Layers {
		if layerSpec.Parent == "" {
			continue
		}
		layer := layersByName[layerSpec.Name]
		parent := layersByName[layerSpec.Parent]
		if layer == nil || parent == nil {
			return fmt.Errorf("recipe: layer %q parent %q not found", layerSpec.Name, layerSpec.Parent)
		}
		if err := layer.SetParent(parent.ID); err != nil {
			return fmt.Errorf("recipe: layer %q parent: %w", layerSpec.Name, err)
		}
	}
	return nil
}

func applyLightSources(compSpec CompSpec, layersByName map[string]*aep.Layer) error {
	for _, layerSpec := range compSpec.Layers {
		if layerSpec.Light == nil || layerSpec.Light.SourceLayer == "" {
			continue
		}
		layer := layersByName[layerSpec.Name]
		target := layersByName[layerSpec.Light.SourceLayer]
		if layer == nil || target == nil {
			return fmt.Errorf("recipe: layer %q light.source_layer %q not found", layerSpec.Name, layerSpec.Light.SourceLayer)
		}
		if err := layer.SetLightSource(target); err != nil {
			return fmt.Errorf("recipe: layer %q light.source_layer: %w", layerSpec.Name, err)
		}
	}
	return nil
}

func applyTextStyle(layer *aep.Layer, spec TextStyleSpec) error {
	if spec.FontSize != nil {
		if err := layer.SetRunFontSize(spec.RunIndex, *spec.FontSize); err != nil {
			return err
		}
	}
	if len(spec.FillColor) >= 3 {
		if err := layer.SetRunFillColor(spec.RunIndex, rgbaColor(spec.FillColor)); err != nil {
			return err
		}
	}
	if spec.Tracking != nil {
		if err := layer.SetRunTracking(spec.RunIndex, *spec.Tracking); err != nil {
			return err
		}
	}
	if spec.FauxBold != nil {
		if err := layer.SetRunFauxBold(spec.RunIndex, *spec.FauxBold); err != nil {
			return err
		}
	}
	if spec.FauxItalic != nil {
		if err := layer.SetRunFauxItalic(spec.RunIndex, *spec.FauxItalic); err != nil {
			return err
		}
	}
	if spec.ApplyStroke != nil {
		if err := layer.SetRunApplyStroke(spec.RunIndex, *spec.ApplyStroke); err != nil {
			return err
		}
	}
	if len(spec.StrokeColor) >= 3 {
		if err := layer.SetRunStrokeColor(spec.RunIndex, rgbaColor(spec.StrokeColor)); err != nil {
			return err
		}
	}
	if spec.StrokeWidth != nil {
		if err := layer.SetRunStrokeWidth(spec.RunIndex, *spec.StrokeWidth); err != nil {
			return err
		}
	}
	if spec.Justification != "" {
		justification, err := textJustification(spec.Justification)
		if err != nil {
			return err
		}
		if err := layer.SetParagraphJustification(spec.ParagraphIndex, justification); err != nil {
			return err
		}
	}
	return nil
}

func textJustification(value string) (aep.TextJustification, error) {
	switch value {
	case "left":
		return aep.TextJustifyLeft, nil
	case "right":
		return aep.TextJustifyRight, nil
	case "center":
		return aep.TextJustifyCenter, nil
	default:
		return 0, fmt.Errorf("unsupported justification %q", value)
	}
}

func layerQuality(value string) (aep.LayerQuality, error) {
	switch value {
	case "wireframe":
		return aep.LayerQualityWireframe, nil
	case "draft":
		return aep.LayerQualityDraft, nil
	case "best":
		return aep.LayerQualityBest, nil
	default:
		return 0, fmt.Errorf("unsupported quality %q", value)
	}
}

func layerAutoOrient(value string) (aep.AutoOrientType, error) {
	switch value {
	case "none":
		return aep.AutoOrientNone, nil
	case "along_path":
		return aep.AutoOrientAlongPath, nil
	case "camera_or_point_of_interest":
		return aep.AutoOrientCameraOrPointOfInterest, nil
	case "characters_toward_camera":
		return aep.AutoOrientCharactersTowardCamera, nil
	default:
		return 0, fmt.Errorf("unsupported auto_orient %q", value)
	}
}

func layerTrackMatte(value string) (aep.TrackMatteType, error) {
	switch value {
	case "none":
		return aep.TrackMatteNone, nil
	case "alpha":
		return aep.TrackMatteAlpha, nil
	case "alpha_inverse":
		return aep.TrackMatteAlphaInverse, nil
	case "luma":
		return aep.TrackMatteLuma, nil
	case "luma_inverse":
		return aep.TrackMatteLumaInverse, nil
	default:
		return 0, fmt.Errorf("unsupported track_matte %q", value)
	}
}

func layerBlendingMode(value string) (aep.BlendingMode, error) {
	switch value {
	case "normal_camera":
		return aep.BlendingModeNormalCamera, nil
	case "normal":
		return aep.BlendingModeNormal, nil
	case "dissolve":
		return aep.BlendingModeDissolve, nil
	case "add":
		return aep.BlendingModeAdd, nil
	case "multiply":
		return aep.BlendingModeMultiply, nil
	case "screen":
		return aep.BlendingModeScreen, nil
	case "overlay":
		return aep.BlendingModeOverlay, nil
	case "soft_light":
		return aep.BlendingModeSoftLight, nil
	case "hard_light":
		return aep.BlendingModeHardLight, nil
	case "darken":
		return aep.BlendingModeDarken, nil
	case "lighten":
		return aep.BlendingModeLighten, nil
	case "classic_difference":
		return aep.BlendingModeClassicDifference, nil
	case "hue":
		return aep.BlendingModeHue, nil
	case "saturation":
		return aep.BlendingModeSaturation, nil
	case "color":
		return aep.BlendingModeColor, nil
	case "luminosity":
		return aep.BlendingModeLuminosity, nil
	case "stencil_alpha":
		return aep.BlendingModeStencilAlpha, nil
	case "stencil_luma":
		return aep.BlendingModeStencilLuma, nil
	case "silhouette_alpha":
		return aep.BlendingModeSilhouetteAlpha, nil
	case "silhouette_luma":
		return aep.BlendingModeSilhouetteLuma, nil
	case "luminescent_premul":
		return aep.BlendingModeLuminescentPremul, nil
	case "alpha_add":
		return aep.BlendingModeAlphaAdd, nil
	case "classic_color_dodge":
		return aep.BlendingModeClassicColorDodge, nil
	case "classic_color_burn":
		return aep.BlendingModeClassicColorBurn, nil
	case "exclusion":
		return aep.BlendingModeExclusion, nil
	case "difference":
		return aep.BlendingModeDifference, nil
	case "color_dodge":
		return aep.BlendingModeColorDodge, nil
	case "color_burn":
		return aep.BlendingModeColorBurn, nil
	case "linear_dodge":
		return aep.BlendingModeLinearDodge, nil
	case "linear_burn":
		return aep.BlendingModeLinearBurn, nil
	case "linear_light":
		return aep.BlendingModeLinearLight, nil
	case "vivid_light":
		return aep.BlendingModeVividLight, nil
	case "pin_light":
		return aep.BlendingModePinLight, nil
	case "hard_mix":
		return aep.BlendingModeHardMix, nil
	case "lighter_color":
		return aep.BlendingModeLighterColor, nil
	case "darker_color":
		return aep.BlendingModeDarkerColor, nil
	case "subtract":
		return aep.BlendingModeSubtract, nil
	case "divide":
		return aep.BlendingModeDivide, nil
	default:
		return 0, fmt.Errorf("unsupported blending_mode %q", value)
	}
}

func lightKind(value string) (aep.LightKind, error) {
	switch strings.ToLower(value) {
	case "parallel":
		return aep.LightKindParallel, nil
	case "spot":
		return aep.LightKindSpot, nil
	case "point":
		return aep.LightKindPoint, nil
	case "ambient":
		return aep.LightKindAmbient, nil
	default:
		return 0, fmt.Errorf("unsupported kind %q", value)
	}
}

func compileShape(group *aep.VectorGroup, shape ShapeSpec) error {
	switch shape.Kind {
	case "rect":
		rect, err := group.AddRect()
		if err != nil {
			return err
		}
		if len(shape.Size) == 2 {
			if err := rect.SetSize([2]float64{shape.Size[0], shape.Size[1]}); err != nil {
				return err
			}
		}
		if len(shape.Position) == 2 {
			if err := rect.SetPosition([2]float64{shape.Position[0], shape.Position[1]}); err != nil {
				return err
			}
		}
		if shape.Roundness != nil {
			if err := rect.SetRoundness(*shape.Roundness); err != nil {
				return err
			}
		}
	case "ellipse":
		ellipse, err := group.AddEllipse()
		if err != nil {
			return err
		}
		if len(shape.Size) == 2 {
			if err := ellipse.SetSize([2]float64{shape.Size[0], shape.Size[1]}); err != nil {
				return err
			}
		}
		if len(shape.Position) == 2 {
			if err := ellipse.SetPosition([2]float64{shape.Position[0], shape.Position[1]}); err != nil {
				return err
			}
		}
	case "star", "polygon":
		star, err := group.AddStar()
		if err != nil {
			return err
		}
		if shape.Kind == "polygon" {
			if err := star.SetStarType(aep.StarTypePolygon); err != nil {
				return err
			}
		}
		if shape.Points != nil {
			if err := star.SetPoints(*shape.Points); err != nil {
				return err
			}
		}
		if len(shape.Position) == 2 {
			if err := star.SetPosition([2]float64{shape.Position[0], shape.Position[1]}); err != nil {
				return err
			}
		}
		if shape.Rotation != nil {
			if err := star.SetRotation(*shape.Rotation); err != nil {
				return err
			}
		}
		if shape.InnerRadius != nil {
			if err := star.SetInnerRadius(*shape.InnerRadius); err != nil {
				return err
			}
		}
		if shape.OuterRadius != nil {
			if err := star.SetOuterRadius(*shape.OuterRadius); err != nil {
				return err
			}
		}
		if shape.InnerRoundness != nil {
			if err := star.SetInnerRoundness(*shape.InnerRoundness); err != nil {
				return err
			}
		}
		if shape.OuterRoundness != nil {
			if err := star.SetOuterRoundness(*shape.OuterRoundness); err != nil {
				return err
			}
		}
	}
	if shape.RoundCorners != nil {
		roundCorners, err := group.AddRoundCorners()
		if err != nil {
			return err
		}
		if shape.RoundCorners.Radius != nil {
			if err := roundCorners.SetRadius(*shape.RoundCorners.Radius); err != nil {
				return err
			}
		}
	}
	if shape.OffsetPaths != nil {
		offsetPaths, err := group.AddOffsetPaths()
		if err != nil {
			return err
		}
		if shape.OffsetPaths.Amount != nil {
			if err := offsetPaths.SetAmount(*shape.OffsetPaths.Amount); err != nil {
				return err
			}
		}
		if shape.OffsetPaths.LineJoin != "" {
			lineJoin, err := offsetLineJoin(shape.OffsetPaths.LineJoin)
			if err != nil {
				return err
			}
			if err := offsetPaths.SetLineJoin(lineJoin); err != nil {
				return err
			}
		}
		if shape.OffsetPaths.MiterLimit != nil {
			if err := offsetPaths.SetMiterLimit(*shape.OffsetPaths.MiterLimit); err != nil {
				return err
			}
		}
		if shape.OffsetPaths.Copies != nil {
			if err := offsetPaths.SetCopies(*shape.OffsetPaths.Copies); err != nil {
				return err
			}
		}
		if shape.OffsetPaths.CopyOffset != nil {
			if err := offsetPaths.SetCopyOffset(*shape.OffsetPaths.CopyOffset); err != nil {
				return err
			}
		}
	}
	if shape.Repeater != nil {
		repeater, err := group.AddRepeater()
		if err != nil {
			return err
		}
		if shape.Repeater.Copies != nil {
			if err := repeater.SetCopies(*shape.Repeater.Copies); err != nil {
				return err
			}
		}
		if shape.Repeater.Offset != nil {
			if err := repeater.SetOffset(*shape.Repeater.Offset); err != nil {
				return err
			}
		}
		if shape.Repeater.Order != "" {
			order, err := repeaterOrder(shape.Repeater.Order)
			if err != nil {
				return err
			}
			if err := repeater.SetOrder(order); err != nil {
				return err
			}
		}
		transform := repeater.Transform()
		if len(shape.Repeater.Anchor) == 2 {
			if err := transform.SetAnchor([2]float64{shape.Repeater.Anchor[0], shape.Repeater.Anchor[1]}); err != nil {
				return err
			}
		}
		if len(shape.Repeater.Position) == 2 {
			if err := transform.SetPosition([2]float64{shape.Repeater.Position[0], shape.Repeater.Position[1]}); err != nil {
				return err
			}
		}
		if len(shape.Repeater.Scale) == 2 {
			if err := transform.SetScale([2]float64{shape.Repeater.Scale[0], shape.Repeater.Scale[1]}); err != nil {
				return err
			}
		}
		if shape.Repeater.Rotation != nil {
			if err := transform.SetRotation(*shape.Repeater.Rotation); err != nil {
				return err
			}
		}
		if shape.Repeater.StartOpacity != nil {
			if err := transform.SetStartOpacity(*shape.Repeater.StartOpacity); err != nil {
				return err
			}
		}
		if shape.Repeater.EndOpacity != nil {
			if err := transform.SetEndOpacity(*shape.Repeater.EndOpacity); err != nil {
				return err
			}
		}
	}
	if shape.MergePaths != nil {
		mergePaths, err := group.AddMergePaths()
		if err != nil {
			return err
		}
		if shape.MergePaths.Type != "" {
			mergeType, err := mergePathsType(shape.MergePaths.Type)
			if err != nil {
				return err
			}
			if err := mergePaths.SetType(mergeType); err != nil {
				return err
			}
		}
	}
	if shape.ZigZag != nil {
		zigZag, err := group.AddZigZag()
		if err != nil {
			return err
		}
		if shape.ZigZag.Size != nil {
			if err := zigZag.SetSize(*shape.ZigZag.Size); err != nil {
				return err
			}
		}
		if shape.ZigZag.Detail != nil {
			if err := zigZag.SetDetail(*shape.ZigZag.Detail); err != nil {
				return err
			}
		}
		if shape.ZigZag.Points != "" {
			points, err := zigZagPoints(shape.ZigZag.Points)
			if err != nil {
				return err
			}
			if err := zigZag.SetPoints(points); err != nil {
				return err
			}
		}
	}
	if shape.PuckerBloat != nil {
		puckerBloat, err := group.AddPuckerBloat()
		if err != nil {
			return err
		}
		if shape.PuckerBloat.Amount != nil {
			if err := puckerBloat.SetAmount(*shape.PuckerBloat.Amount); err != nil {
				return err
			}
		}
	}
	if shape.Twist != nil {
		twist, err := group.AddTwist()
		if err != nil {
			return err
		}
		if shape.Twist.Angle != nil {
			if err := twist.SetAngle(*shape.Twist.Angle); err != nil {
				return err
			}
		}
		if len(shape.Twist.Center) == 2 {
			if err := twist.SetCenter([2]float64{shape.Twist.Center[0], shape.Twist.Center[1]}); err != nil {
				return err
			}
		}
	}
	if shape.WigglePaths != nil {
		wigglePaths, err := group.AddWigglePaths()
		if err != nil {
			return err
		}
		if shape.WigglePaths.Size != nil {
			if err := wigglePaths.SetSize(*shape.WigglePaths.Size); err != nil {
				return err
			}
		}
		if shape.WigglePaths.Detail != nil {
			if err := wigglePaths.SetDetail(*shape.WigglePaths.Detail); err != nil {
				return err
			}
		}
		if shape.WigglePaths.WigglesPerSecond != nil {
			if err := wigglePaths.SetWigglesPerSecond(*shape.WigglePaths.WigglesPerSecond); err != nil {
				return err
			}
		}
		if shape.WigglePaths.RandomSeed != nil {
			if err := wigglePaths.SetRandomSeed(*shape.WigglePaths.RandomSeed); err != nil {
				return err
			}
		}
		if shape.WigglePaths.Points != "" {
			points, err := roughenPoints(shape.WigglePaths.Points)
			if err != nil {
				return err
			}
			if err := wigglePaths.SetPoints(points); err != nil {
				return err
			}
		}
		if shape.WigglePaths.Correlation != nil {
			if err := wigglePaths.SetCorrelation(*shape.WigglePaths.Correlation); err != nil {
				return err
			}
		}
		if shape.WigglePaths.TemporalPhase != nil {
			if err := wigglePaths.SetTemporalPhase(*shape.WigglePaths.TemporalPhase); err != nil {
				return err
			}
		}
		if shape.WigglePaths.SpatialPhase != nil {
			if err := wigglePaths.SetSpatialPhase(*shape.WigglePaths.SpatialPhase); err != nil {
				return err
			}
		}
	}
	if shape.WiggleTransform != nil {
		wiggleTransform, err := group.AddWiggleTransform()
		if err != nil {
			return err
		}
		transform := wiggleTransform.Transform()
		if len(shape.WiggleTransform.Anchor) == 2 {
			if err := transform.SetAnchor([2]float64{shape.WiggleTransform.Anchor[0], shape.WiggleTransform.Anchor[1]}); err != nil {
				return err
			}
		}
		if len(shape.WiggleTransform.Position) == 2 {
			if err := transform.SetPosition([2]float64{shape.WiggleTransform.Position[0], shape.WiggleTransform.Position[1]}); err != nil {
				return err
			}
		}
		if len(shape.WiggleTransform.Scale) == 2 {
			if err := transform.SetScale([2]float64{shape.WiggleTransform.Scale[0], shape.WiggleTransform.Scale[1]}); err != nil {
				return err
			}
		}
		if shape.WiggleTransform.Rotation != nil {
			if err := transform.SetRotation(*shape.WiggleTransform.Rotation); err != nil {
				return err
			}
		}
		if shape.WiggleTransform.WigglesPerSecond != nil {
			if err := wiggleTransform.SetWigglesPerSecond(*shape.WiggleTransform.WigglesPerSecond); err != nil {
				return err
			}
		}
		if shape.WiggleTransform.RandomSeed != nil {
			if err := wiggleTransform.SetRandomSeed(*shape.WiggleTransform.RandomSeed); err != nil {
				return err
			}
		}
		if shape.WiggleTransform.Correlation != nil {
			if err := wiggleTransform.SetCorrelation(*shape.WiggleTransform.Correlation); err != nil {
				return err
			}
		}
		if shape.WiggleTransform.TemporalPhase != nil {
			if err := wiggleTransform.SetTemporalPhase(*shape.WiggleTransform.TemporalPhase); err != nil {
				return err
			}
		}
		if shape.WiggleTransform.SpatialPhase != nil {
			if err := wiggleTransform.SetSpatialPhase(*shape.WiggleTransform.SpatialPhase); err != nil {
				return err
			}
		}
	}
	if shape.Trim != nil {
		trim, err := group.AddTrim()
		if err != nil {
			return err
		}
		if shape.Trim.Start != nil {
			if err := trim.SetStart(*shape.Trim.Start); err != nil {
				return err
			}
		}
		if shape.Trim.End != nil {
			if err := trim.SetEnd(*shape.Trim.End); err != nil {
				return err
			}
		}
		if shape.Trim.Offset != nil {
			if err := trim.SetOffset(*shape.Trim.Offset); err != nil {
				return err
			}
		}
	}
	if len(shape.FillColor) >= 3 || shape.FillOpacity != nil || shape.FillBlendMode != nil || shape.FillCompositeOrder != "" || shape.FillRule != "" {
		fill, err := group.AddFill()
		if err != nil {
			return err
		}
		if len(shape.FillColor) >= 3 {
			if err := fill.SetColor(rgbaColor(shape.FillColor)); err != nil {
				return err
			}
		}
		if shape.FillOpacity != nil {
			if err := fill.SetOpacity(*shape.FillOpacity); err != nil {
				return err
			}
		}
		if shape.FillBlendMode != nil {
			mode, err := shapeBlendMode(*shape.FillBlendMode)
			if err != nil {
				return err
			}
			if err := fill.SetBlendMode(mode); err != nil {
				return err
			}
		}
		if shape.FillCompositeOrder != "" {
			order, err := shapeCompositeOrder(shape.FillCompositeOrder)
			if err != nil {
				return err
			}
			if err := fill.SetCompositeOrder(order); err != nil {
				return err
			}
		}
		if shape.FillRule != "" {
			rule, err := fillRule(shape.FillRule)
			if err != nil {
				return err
			}
			if err := fill.SetFillRule(rule); err != nil {
				return err
			}
		}
	}
	if shape.GradientFill != nil {
		fill, err := group.AddGradientFill()
		if err != nil {
			return err
		}
		if shape.GradientFill.Type != "" {
			typ, err := gradientType(shape.GradientFill.Type)
			if err != nil {
				return err
			}
			if err := fill.SetGradientType(typ); err != nil {
				return err
			}
		}
		if len(shape.GradientFill.StartPoint) == 2 {
			if err := fill.SetStartPoint([2]float64{shape.GradientFill.StartPoint[0], shape.GradientFill.StartPoint[1]}); err != nil {
				return err
			}
		}
		if len(shape.GradientFill.EndPoint) == 2 {
			if err := fill.SetEndPoint([2]float64{shape.GradientFill.EndPoint[0], shape.GradientFill.EndPoint[1]}); err != nil {
				return err
			}
		}
		if shape.GradientFill.HighlightLength != nil {
			if err := fill.SetHighlightLength(*shape.GradientFill.HighlightLength); err != nil {
				return err
			}
		}
		if shape.GradientFill.HighlightAngle != nil {
			if err := fill.SetHighlightAngle(*shape.GradientFill.HighlightAngle); err != nil {
				return err
			}
		}
		if len(shape.GradientFill.ColorStops) > 0 {
			stops, err := gradientColorStops(shape.GradientFill.ColorStops)
			if err != nil {
				return err
			}
			if err := fill.SetColorStops(stops); err != nil {
				return err
			}
		}
		if len(shape.GradientFill.AlphaStops) > 0 {
			stops := gradientAlphaStops(shape.GradientFill.AlphaStops)
			if err := fill.SetAlphaStops(stops); err != nil {
				return err
			}
		}
	}
	if shape.GradientStroke != nil {
		stroke, err := group.AddGradientStroke()
		if err != nil {
			return err
		}
		if shape.GradientStroke.Type != "" {
			typ, err := gradientType(shape.GradientStroke.Type)
			if err != nil {
				return err
			}
			if err := stroke.SetGradientType(typ); err != nil {
				return err
			}
		}
		if len(shape.GradientStroke.StartPoint) == 2 {
			if err := stroke.SetStartPoint([2]float64{shape.GradientStroke.StartPoint[0], shape.GradientStroke.StartPoint[1]}); err != nil {
				return err
			}
		}
		if len(shape.GradientStroke.EndPoint) == 2 {
			if err := stroke.SetEndPoint([2]float64{shape.GradientStroke.EndPoint[0], shape.GradientStroke.EndPoint[1]}); err != nil {
				return err
			}
		}
		if shape.GradientStroke.HighlightLength != nil {
			if err := stroke.SetHighlightLength(*shape.GradientStroke.HighlightLength); err != nil {
				return err
			}
		}
		if shape.GradientStroke.HighlightAngle != nil {
			if err := stroke.SetHighlightAngle(*shape.GradientStroke.HighlightAngle); err != nil {
				return err
			}
		}
		if shape.GradientStroke.Width != nil {
			if err := stroke.SetStrokeWidth(*shape.GradientStroke.Width); err != nil {
				return err
			}
		}
		if shape.GradientStroke.LineCap != "" {
			lineCap, err := strokeLineCap(shape.GradientStroke.LineCap)
			if err != nil {
				return err
			}
			if err := stroke.SetLineCap(lineCap); err != nil {
				return err
			}
		}
		if shape.GradientStroke.LineJoin != "" {
			lineJoin, err := strokeLineJoin(shape.GradientStroke.LineJoin)
			if err != nil {
				return err
			}
			if err := stroke.SetLineJoin(lineJoin); err != nil {
				return err
			}
		}
		if shape.GradientStroke.MiterLimit != nil {
			if err := stroke.SetMiterLimit(*shape.GradientStroke.MiterLimit); err != nil {
				return err
			}
		}
		if len(shape.GradientStroke.ColorStops) > 0 {
			stops, err := gradientColorStops(shape.GradientStroke.ColorStops)
			if err != nil {
				return err
			}
			if err := stroke.SetColorStops(stops); err != nil {
				return err
			}
		}
		if len(shape.GradientStroke.AlphaStops) > 0 {
			stops := gradientAlphaStops(shape.GradientStroke.AlphaStops)
			if err := stroke.SetAlphaStops(stops); err != nil {
				return err
			}
		}
	}
	if shape.Stroke != nil {
		stroke, err := group.AddStroke()
		if err != nil {
			return err
		}
		if len(shape.Stroke.Color) >= 3 {
			color := rgbaColor(shape.Stroke.Color)
			if err := stroke.SetColor(color); err != nil {
				return err
			}
		}
		if shape.Stroke.Width != nil {
			if err := stroke.SetWidth(*shape.Stroke.Width); err != nil {
				return err
			}
		}
		if shape.Stroke.Opacity != nil {
			if err := stroke.SetOpacity(*shape.Stroke.Opacity); err != nil {
				return err
			}
		}
		if shape.Stroke.LineCap != "" {
			lineCap, err := strokeLineCap(shape.Stroke.LineCap)
			if err != nil {
				return err
			}
			if err := stroke.SetLineCap(lineCap); err != nil {
				return err
			}
		}
		if shape.Stroke.LineJoin != "" {
			lineJoin, err := strokeLineJoin(shape.Stroke.LineJoin)
			if err != nil {
				return err
			}
			if err := stroke.SetLineJoin(lineJoin); err != nil {
				return err
			}
		}
		if shape.Stroke.MiterLimit != nil {
			if err := stroke.SetMiterLimit(*shape.Stroke.MiterLimit); err != nil {
				return err
			}
		}
		if shape.Stroke.CompositeOrder != "" {
			order, err := shapeCompositeOrder(shape.Stroke.CompositeOrder)
			if err != nil {
				return err
			}
			if err := stroke.SetCompositeOrder(order); err != nil {
				return err
			}
		}
		if shape.Stroke.Taper != nil {
			taper := stroke.Taper()
			if shape.Stroke.Taper.StartLength != nil {
				if err := taper.SetStartLength(*shape.Stroke.Taper.StartLength); err != nil {
					return err
				}
			}
			if shape.Stroke.Taper.EndLength != nil {
				if err := taper.SetEndLength(*shape.Stroke.Taper.EndLength); err != nil {
					return err
				}
			}
			if shape.Stroke.Taper.StartWidth != nil {
				if err := taper.SetStartWidth(*shape.Stroke.Taper.StartWidth); err != nil {
					return err
				}
			}
			if shape.Stroke.Taper.EndWidth != nil {
				if err := taper.SetEndWidth(*shape.Stroke.Taper.EndWidth); err != nil {
					return err
				}
			}
			if shape.Stroke.Taper.StartEase != nil {
				if err := taper.SetStartEase(*shape.Stroke.Taper.StartEase); err != nil {
					return err
				}
			}
			if shape.Stroke.Taper.EndEase != nil {
				if err := taper.SetEndEase(*shape.Stroke.Taper.EndEase); err != nil {
					return err
				}
			}
		}
		if shape.Stroke.Wave != nil {
			wave := stroke.Wave()
			if shape.Stroke.Wave.Amount != nil {
				if err := wave.SetAmount(*shape.Stroke.Wave.Amount); err != nil {
					return err
				}
			}
			if shape.Stroke.Wave.Wavelength != nil {
				if err := wave.SetWavelength(*shape.Stroke.Wave.Wavelength); err != nil {
					return err
				}
			}
			if shape.Stroke.Wave.Phase != nil {
				if err := wave.SetPhase(*shape.Stroke.Wave.Phase); err != nil {
					return err
				}
			}
		}
		if shape.Stroke.Dashes != nil {
			dashes := stroke.Dashes()
			if shape.Stroke.Dashes.Dash != nil {
				if err := dashes.SetDash(*shape.Stroke.Dashes.Dash); err != nil {
					return err
				}
			}
			if shape.Stroke.Dashes.Gap != nil {
				if err := dashes.SetGap(*shape.Stroke.Dashes.Gap); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func offsetLineJoin(value string) (aep.StrokeLineJoin, error) {
	switch value {
	case "miter":
		return aep.StrokeLineJoinMiter, nil
	case "round":
		return aep.StrokeLineJoinRound, nil
	case "bevel":
		return aep.StrokeLineJoinBevel, nil
	default:
		return 0, fmt.Errorf("unsupported offset line_join %q", value)
	}
}

func strokeLineCap(value string) (aep.StrokeLineCap, error) {
	switch value {
	case "butt":
		return aep.StrokeLineCapButt, nil
	case "round":
		return aep.StrokeLineCapRound, nil
	case "projecting":
		return aep.StrokeLineCapProjecting, nil
	default:
		return 0, fmt.Errorf("unsupported stroke line_cap %q", value)
	}
}

func strokeLineJoin(value string) (aep.StrokeLineJoin, error) {
	switch value {
	case "miter":
		return aep.StrokeLineJoinMiter, nil
	case "round":
		return aep.StrokeLineJoinRound, nil
	case "bevel":
		return aep.StrokeLineJoinBevel, nil
	default:
		return 0, fmt.Errorf("unsupported stroke line_join %q", value)
	}
}

func shapeCompositeOrder(value string) (aep.ShapeCompositeOrder, error) {
	switch value {
	case "above_previous":
		return aep.ShapeCompositeOrderAbovePrevious, nil
	case "below_previous":
		return aep.ShapeCompositeOrderBelowPrevious, nil
	default:
		return 0, fmt.Errorf("unsupported shape composite order %q", value)
	}
}

func shapeBlendMode(value float64) (aep.ShapeBlendMode, error) {
	if value < 1 || !isWholeNumber(value) {
		return 0, fmt.Errorf("shape blend mode must be an integer of at least 1, got %g", value)
	}
	return aep.ShapeBlendMode(value), nil
}

func fillRule(value string) (aep.FillRule, error) {
	switch value {
	case "nonzero_winding":
		return aep.FillRuleNonzeroWinding, nil
	case "even_odd":
		return aep.FillRuleEvenOdd, nil
	default:
		return 0, fmt.Errorf("unsupported fill rule %q", value)
	}
}

func gradientType(value string) (aep.GradientType, error) {
	switch value {
	case "linear":
		return aep.GradientLinear, nil
	case "radial":
		return aep.GradientRadial, nil
	default:
		return 0, fmt.Errorf("unsupported gradient type %q", value)
	}
}

func gradientColorStops(specs []GradientColorStopSpec) ([]aep.GradientColorStop, error) {
	stops := make([]aep.GradientColorStop, 0, len(specs))
	for i, spec := range specs {
		if len(spec.Color) < 3 {
			return nil, fmt.Errorf("gradient color stop %d needs at least 3 color channels", i)
		}
		midpoint := 0.5
		if spec.Midpoint != nil {
			midpoint = *spec.Midpoint
		}
		stops = append(stops, aep.GradientColorStop{
			Offset:   spec.Offset,
			Midpoint: midpoint,
			Color: [3]float64{
				toUnitColor(spec.Color[0]),
				toUnitColor(spec.Color[1]),
				toUnitColor(spec.Color[2]),
			},
		})
	}
	return stops, nil
}

func gradientAlphaStops(specs []GradientAlphaStopSpec) []aep.GradientAlphaStop {
	stops := make([]aep.GradientAlphaStop, 0, len(specs))
	for _, spec := range specs {
		midpoint := 0.5
		if spec.Midpoint != nil {
			midpoint = *spec.Midpoint
		}
		stops = append(stops, aep.GradientAlphaStop{
			Offset:   spec.Offset,
			Midpoint: midpoint,
			Alpha:    spec.Alpha,
		})
	}
	return stops
}

func repeaterOrder(value string) (aep.RepeaterOrder, error) {
	switch value {
	case "below":
		return aep.RepeaterOrderBelow, nil
	case "above":
		return aep.RepeaterOrderAbove, nil
	default:
		return 0, fmt.Errorf("unsupported repeater order %q", value)
	}
}

func mergePathsType(value string) (aep.MergeType, error) {
	switch value {
	case "merge":
		return aep.MergeTypeMerge, nil
	case "add":
		return aep.MergeTypeAdd, nil
	case "subtract":
		return aep.MergeTypeSubtract, nil
	case "intersect":
		return aep.MergeTypeIntersect, nil
	case "exclude":
		return aep.MergeTypeExclude, nil
	default:
		return 0, fmt.Errorf("unsupported merge_paths type %q", value)
	}
}

func zigZagPoints(value string) (aep.ZigZagPoints, error) {
	switch value {
	case "corner":
		return aep.ZigZagPointsCorner, nil
	case "smooth":
		return aep.ZigZagPointsSmooth, nil
	default:
		return 0, fmt.Errorf("unsupported zigzag points %q", value)
	}
}

func roughenPoints(value string) (aep.RoughenPoints, error) {
	switch value {
	case "corner":
		return aep.RoughenPointsCorner, nil
	case "smooth":
		return aep.RoughenPointsSmooth, nil
	default:
		return 0, fmt.Errorf("unsupported wiggle_paths points %q", value)
	}
}

func applyTransform(layer *aep.Layer, spec Transform) error {
	t := aep.NewLayerTransform()
	if len(spec.AnchorPoint) == 2 {
		if err := t.AnchorPoint().SetStaticValue([2]float64{spec.AnchorPoint[0], spec.AnchorPoint[1]}); err != nil {
			return err
		}
	}
	if len(spec.Position) == 2 {
		if err := t.Position().SetStaticValue([2]float64{spec.Position[0], spec.Position[1]}); err != nil {
			return err
		}
	}
	if len(spec.Scale) == 2 {
		if err := t.Scale().SetStaticValue([2]float64{spec.Scale[0], spec.Scale[1]}); err != nil {
			return err
		}
	}
	if spec.Rotation != nil {
		if err := t.Rotation().SetStaticValue(*spec.Rotation); err != nil {
			return err
		}
	}
	if spec.Opacity != nil {
		if err := t.Opacity().SetStaticValue(*spec.Opacity); err != nil {
			return err
		}
	}
	for _, kf := range spec.PositionKeyframes {
		value := [2]float64{kf.Value[0], kf.Value[1]}
		if hasKeyframeEase(kf.InEase, kf.OutEase) {
			if err := t.Position().AddKeyframeWithEase(kf.Time, value, temporalEase(kf.InEase), temporalEase(kf.OutEase)); err != nil {
				return err
			}
		} else if err := t.Position().AddKeyframeLinear(kf.Time, value); err != nil {
			return err
		}
	}
	for _, kf := range spec.AnchorPointKeyframes {
		value := [2]float64{kf.Value[0], kf.Value[1]}
		if hasKeyframeEase(kf.InEase, kf.OutEase) {
			if err := t.AnchorPoint().AddKeyframeWithEase(kf.Time, value, temporalEase(kf.InEase), temporalEase(kf.OutEase)); err != nil {
				return err
			}
		} else if err := t.AnchorPoint().AddKeyframeLinear(kf.Time, value); err != nil {
			return err
		}
	}
	for _, kf := range spec.ScaleKeyframes {
		value := [2]float64{kf.Value[0], kf.Value[1]}
		if hasKeyframeEase(kf.InEase, kf.OutEase) {
			if err := t.Scale().AddKeyframeWithEase(kf.Time, value, temporalEase(kf.InEase), temporalEase(kf.OutEase)); err != nil {
				return err
			}
		} else if err := t.Scale().AddKeyframeLinear(kf.Time, value); err != nil {
			return err
		}
	}
	for _, kf := range spec.RotationKeyframes {
		if hasKeyframeEase(kf.InEase, kf.OutEase) {
			if err := t.Rotation().AddKeyframeWithEase(kf.Time, kf.Value, temporalEase(kf.InEase), temporalEase(kf.OutEase)); err != nil {
				return err
			}
		} else if err := t.Rotation().AddKeyframeLinear(kf.Time, kf.Value); err != nil {
			return err
		}
	}
	for _, kf := range spec.OpacityKeyframes {
		if hasKeyframeEase(kf.InEase, kf.OutEase) {
			if err := t.Opacity().AddKeyframeWithEase(kf.Time, kf.Value, temporalEase(kf.InEase), temporalEase(kf.OutEase)); err != nil {
				return err
			}
		} else if err := t.Opacity().AddKeyframeLinear(kf.Time, kf.Value); err != nil {
			return err
		}
	}
	return aep.SetLayerTransform(layer, t)
}

func hasKeyframeEase(in, out *TemporalEase) bool {
	return in != nil || out != nil
}

func temporalEase(ease *TemporalEase) aep.TemporalEase {
	if ease == nil {
		return aep.TemporalEase{}
	}
	return aep.TemporalEase{
		Speed:     ease.Speed,
		Influence: ease.Influence,
	}
}

func toUnitColor(v float64) float64 {
	if v > 1 {
		return v / 255
	}
	return v
}

func rgbaColor(values []float64) [4]float64 {
	alpha := 1.0
	if len(values) >= 4 {
		alpha = toUnitColor(values[3])
	}
	return [4]float64{
		toUnitColor(values[0]),
		toUnitColor(values[1]),
		toUnitColor(values[2]),
		alpha,
	}
}

func rgb8Color(values []float64) [3]uint8 {
	return [3]uint8{uint8(values[0]), uint8(values[1]), uint8(values[2])}
}
