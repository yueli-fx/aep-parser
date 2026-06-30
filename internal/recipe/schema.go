package recipe

import (
	"fmt"
)

func Validate(rec Recipe) Report {
	return ValidateWithCapabilities(rec, unknownCapabilities{})
}

func ValidateWithCapabilities(rec Recipe, caps CapabilityIndex) Report {
	if caps == nil {
		caps = unknownCapabilities{}
	}
	report := Report{SchemaVersion: SchemaVersion, Valid: true}
	addRefusal := func(code, path, message string) {
		report.Valid = false
		report.Refusals = append(report.Refusals, Refusal{Code: code, Path: path, Message: message})
	}
	recordCapability := func(query, path string) CapabilityLookup {
		lookup := caps.Lookup(query)
		if lookup.Query == "" {
			lookup.Query = query
		}
		report.Capabilities = append(report.Capabilities, CapabilityUse{Path: path, CapabilityLookup: lookup})
		switch {
		case lookup.Status == CapabilityUnknown:
			report.Downgrades = append(report.Downgrades, Downgrade{
				Code:    "unknown_capability",
				Path:    path,
				Query:   query,
				Message: fmt.Sprintf("capability %q is not present in the loaded index", query),
			})
		case lookup.Status == CapabilitySupported && lookup.Tier != "" && lookup.Tier != "stable":
			report.Downgrades = append(report.Downgrades, Downgrade{
				Code:    "non_stable_capability",
				Path:    path,
				Query:   query,
				Message: fmt.Sprintf("capability %q is %s", query, lookup.Tier),
			})
		}
		return lookup
	}

	if rec.SchemaVersion != SchemaVersion {
		addRefusal("unsupported_schema_version", "schema_version", fmt.Sprintf("schema_version must be %d", SchemaVersion))
	}
	if len(rec.Comps) == 0 {
		addRefusal("missing_comp", "comps", "at least one comp is required")
		return report
	}
	projectTarget, err := parseRecipeProjectTarget(rec.Project.TargetVersion)
	if err != nil {
		addRefusal("invalid_project_target_version", "project.target_version", err.Error())
	}
	if rec.Project.TargetVersion != "" {
		recordCapability("NewProject", "project.target_version")
	}
	if len(rec.Comps) > 1 {
		addRefusal("too_many_comps", "comps", "first recipe slice supports exactly one comp")
	}
	validateExpectedProfile(rec.ExpectedProfile, addRefusal)
	for ci, comp := range rec.Comps {
		compPath := fmt.Sprintf("comps[%d]", ci)
		recordCapability("NewComposition", compPath)
		if comp.Name == "" {
			addRefusal("missing_comp_name", compPath+".name", "comp name is required")
		}
		if comp.Width <= 0 || comp.Height <= 0 || comp.FrameRate <= 0 || comp.Duration <= 0 {
			addRefusal("invalid_comp_timing_or_size", compPath, "width, height, frame_rate, and duration must be positive")
		}
		if len(comp.BackgroundColor) > 0 {
			recordCapability("SetBGColor", compPath+".background_color")
			validateRGBColor(comp.BackgroundColor, compPath+".background_color", "invalid_comp_background_color", addRefusal)
		}
		if comp.Label != nil {
			recordCapability("SetLabel", compPath+".label")
			validateCompLabel(*comp.Label, compPath+".label", addRefusal)
		}
		if comp.Comment != "" {
			recordCapability("SetComment", compPath+".comment")
		}
		if comp.MotionGraphicsTemplateName != "" {
			recordCapability("Composition.SetMotionGraphicsTemplateName", compPath+".motion_graphics_template_name")
		}
		if comp.Renderer != "" {
			recordCapability("SetRenderer", compPath+".renderer")
		}
		if len(comp.ResolutionFactor) > 0 {
			recordCapability("SetResolutionFactor", compPath+".resolution_factor")
			validateResolutionFactor(comp.ResolutionFactor, compPath+".resolution_factor", addRefusal)
		}
		if comp.PixelAspect != nil {
			recordCapability("SetPixelAspect", compPath+".pixel_aspect")
			validatePixelAspect(*comp.PixelAspect, compPath+".pixel_aspect", addRefusal)
		}
		if comp.DisplayStartTime != nil {
			recordCapability("SetDisplayStartTime", compPath+".display_start_time")
			validateDisplayStartTime(*comp.DisplayStartTime, compPath+".display_start_time", addRefusal)
		}
		if comp.FrameBlending != nil {
			recordCapability("SetFrameBlending", compPath+".frame_blending")
		}
		if comp.Draft3D != nil {
			recordCapability("SetDraft3D", compPath+".draft_3d")
		}
		if comp.HideShyLayers != nil {
			recordCapability("SetHideShyLayers", compPath+".hide_shy_layers")
		}
		if comp.PreserveNestedFrameRate != nil {
			recordCapability("SetPreserveNestedFrameRate", compPath+".preserve_nested_frame_rate")
		}
		if comp.PreserveNestedResolution != nil {
			recordCapability("SetPreserveNestedResolution", compPath+".preserve_nested_resolution")
		}
		if comp.MotionBlur != nil {
			validateCompMotionBlur(comp.MotionBlur, compPath+".motion_blur", recordCapability, addRefusal)
		}
		if comp.WorkArea != nil {
			validateCompWorkArea(comp.WorkArea, compPath+".work_area", comp.Duration, recordCapability, addRefusal)
		}
		layerNames := map[string]bool{}
		for _, layer := range comp.Layers {
			if layer.Name != "" {
				layerNames[layer.Name] = true
			}
		}
		for li, layer := range comp.Layers {
			layerPath := fmt.Sprintf("%s.layers[%d]", compPath, li)
			validateLayer(layer, layerPath, comp.Duration, recordCapability, addRefusal)
			if layer.Parent != "" {
				recordCapability("Layer.SetParent", layerPath+".parent")
				if !layerNames[layer.Parent] {
					addRefusal("unknown_layer_parent", layerPath+".parent", fmt.Sprintf("parent layer %q was not found in the comp", layer.Parent))
				}
			}
			if layer.Matte != "" {
				recordCapability("Layer.SetTrackMatteSource", layerPath+".matte")
				if projectTarget != aepTargetAE2025 {
					addRefusal("explicit_matte_requires_ae2025", layerPath+".matte", "explicit matte source requires project.target_version AE2025")
				}
				if layer.TrackMatte == "" || layer.TrackMatte == "none" {
					addRefusal("missing_explicit_matte_mode", layerPath+".track_matte", "explicit matte source requires a non-none track_matte mode")
				}
				if layer.Matte == layer.Name {
					addRefusal("self_explicit_matte", layerPath+".matte", "layer cannot use itself as an explicit matte source")
				}
				if !layerNames[layer.Matte] {
					addRefusal("unknown_explicit_matte", layerPath+".matte", fmt.Sprintf("matte layer %q was not found in the comp", layer.Matte))
				}
			}
			if layer.Light != nil && layer.Light.SourceLayer != "" && !layerNames[layer.Light.SourceLayer] {
				addRefusal("unknown_light_source", layerPath+".light.source_layer", fmt.Sprintf("source layer %q was not found in the comp", layer.Light.SourceLayer))
			}
			for ei, effect := range layer.Effects {
				for pi, param := range effect.Params {
					if param.TargetLayer != "" && !layerNames[param.TargetLayer] {
						addRefusal("unknown_effect_target_layer", fmt.Sprintf("%s.effects[%d].params[%d].target_layer", layerPath, ei, pi), fmt.Sprintf("target layer %q was not found in the comp", param.TargetLayer))
					}
				}
			}
		}
	}
	return report
}

func validateExpectedProfile(expected ExpectedProfile, addRefusal func(string, string, string)) {
	if expected.CompCount != nil && *expected.CompCount < 0 {
		addRefusal("invalid_expected_profile", "expected_profile.comp_count", "expected count must be non-negative")
	}
	if expected.LayerCount != nil && *expected.LayerCount < 0 {
		addRefusal("invalid_expected_profile", "expected_profile.layer_count", "expected count must be non-negative")
	}
	if expected.TextLayerCount != nil && *expected.TextLayerCount < 0 {
		addRefusal("invalid_expected_profile", "expected_profile.text_layer_count", "expected count must be non-negative")
	}
	if expected.ShapeLayerCount != nil && *expected.ShapeLayerCount < 0 {
		addRefusal("invalid_expected_profile", "expected_profile.shape_layer_count", "expected count must be non-negative")
	}
	if expected.Label != nil {
		validateCompLabel(*expected.Label, "expected_profile.label", addRefusal)
	}
	if len(expected.BackgroundColor) > 0 {
		validateRGBColor(expected.BackgroundColor, "expected_profile.background_color", "invalid_expected_profile", addRefusal)
	}
	if len(expected.ResolutionFactor) > 0 {
		validateResolutionFactor(expected.ResolutionFactor, "expected_profile.resolution_factor", addRefusal)
	}
	if expected.PixelAspect != nil {
		validatePixelAspect(*expected.PixelAspect, "expected_profile.pixel_aspect", addRefusal)
	}
	if expected.DisplayStartTime != nil {
		validateDisplayStartTime(*expected.DisplayStartTime, "expected_profile.display_start_time", addRefusal)
	}
	if expected.Width != nil && (*expected.Width <= 0 || !isWholeNumber(*expected.Width)) {
		addRefusal("invalid_expected_profile", "expected_profile.width", "width must be a positive integer")
	}
	if expected.Height != nil && (*expected.Height <= 0 || !isWholeNumber(*expected.Height)) {
		addRefusal("invalid_expected_profile", "expected_profile.height", "height must be a positive integer")
	}
	if expected.FrameRate != nil && *expected.FrameRate <= 0 {
		addRefusal("invalid_expected_profile", "expected_profile.frame_rate", "frame_rate must be positive")
	}
	if expected.Duration != nil && *expected.Duration <= 0 {
		addRefusal("invalid_expected_profile", "expected_profile.duration", "duration must be positive")
	}
	for i, layer := range expected.Layers {
		layerPath := fmt.Sprintf("expected_profile.layers[%d]", i)
		if layer.Name == "" {
			addRefusal("invalid_expected_profile", layerPath+".name", "layer name is required")
		}
		if layer.Label != nil {
			validateLayerLabel(*layer.Label, layerPath+".label", addRefusal)
		}
		if layer.Quality != "" {
			if _, err := layerQuality(layer.Quality); err != nil {
				addRefusal("invalid_expected_profile", layerPath+".quality", "quality must be wireframe, draft, or best")
			}
		}
		if layer.BlendingMode != "" {
			if _, err := layerBlendingMode(layer.BlendingMode); err != nil {
				addRefusal("invalid_expected_profile", layerPath+".blending_mode", "blending_mode is not supported")
			}
		}
		if layer.TrackMatte != "" {
			if _, err := layerTrackMatte(layer.TrackMatte); err != nil {
				addRefusal("invalid_expected_profile", layerPath+".track_matte", "track_matte must be none, alpha, alpha_inverse, luma, or luma_inverse")
			}
		}
		if layer.AutoOrient != "" {
			if _, err := layerAutoOrient(layer.AutoOrient); err != nil {
				addRefusal("invalid_expected_profile", layerPath+".auto_orient", "auto_orient must be none, along_path, camera_or_point_of_interest, or characters_toward_camera")
			}
		}
		if layer.Timing != nil {
			validateExpectedLayerTiming(layer.Timing, layerPath+".timing", addRefusal)
		}
	}
	for i, effect := range expected.Effects {
		effectPath := fmt.Sprintf("expected_profile.effects[%d]", i)
		if effect.LayerName == "" {
			addRefusal("invalid_expected_profile", effectPath+".layer_name", "layer_name is required")
		}
		if effect.MatchName == "" {
			addRefusal("invalid_expected_profile", effectPath+".match_name", "effect match_name is required")
		}
		for pi, param := range effect.Params {
			paramPath := fmt.Sprintf("%s.params[%d]", effectPath, pi)
			if param.MatchName == "" {
				addRefusal("invalid_expected_profile", paramPath+".match_name", "param match_name is required")
			}
			if param.Expression == "" && param.ExpressionEnabled == nil && param.Value == nil && param.TargetLayer == "" {
				addRefusal("invalid_expected_profile", paramPath, "expected param value, target_layer, expression, or expression_enabled is required")
			}
			if param.Value != nil && !validEffectParamValue(param.Value) {
				addRefusal("invalid_expected_profile", paramPath+".value", "expected param value must be a number, boolean, or numeric array")
			}
		}
	}
	for i, controller := range expected.EssentialGraphics {
		controllerPath := fmt.Sprintf("expected_profile.essential_graphics[%d]", i)
		if controller.Name == "" {
			addRefusal("invalid_expected_profile", controllerPath+".name", "controller name is required")
		}
		if controller.Type != "" {
			switch controller.Type {
			case "checkbox", "slider", "color", "point", "text", "comment", "multidimensional", "group", "dropdown", "unknown":
			default:
				addRefusal("invalid_expected_profile", controllerPath+".type", "controller type must be checkbox, slider, color, point, text, comment, multidimensional, group, dropdown, or unknown")
			}
		}
	}
	for i, prop := range expected.Properties {
		propPath := fmt.Sprintf("expected_profile.properties[%d]", i)
		if prop.LayerName == "" {
			addRefusal("invalid_expected_profile", propPath+".layer_name", "layer_name is required")
		}
		if prop.MatchName == "" {
			addRefusal("invalid_expected_profile", propPath+".match_name", "property match_name is required")
		}
		if prop.Expression == "" && prop.ExpressionEnabled == nil && prop.Value == nil {
			addRefusal("invalid_expected_profile", propPath, "expected property value, expression, or expression_enabled is required")
		}
		if prop.Value != nil && !validEffectParamValue(prop.Value) {
			addRefusal("invalid_expected_profile", propPath+".value", "expected property value must be a number, boolean, or numeric array")
		}
	}
	for i, style := range expected.TextStyles {
		stylePath := fmt.Sprintf("expected_profile.text_styles[%d]", i)
		if style.LayerName == "" {
			addRefusal("invalid_expected_profile", stylePath+".layer_name", "layer_name is required")
		}
		if style.RunIndex < 0 {
			addRefusal("invalid_expected_profile", stylePath+".run_index", "run_index must be non-negative")
		}
		if style.ParagraphIndex < 0 {
			addRefusal("invalid_expected_profile", stylePath+".paragraph_index", "paragraph_index must be non-negative")
		}
		if style.FontSize != nil && *style.FontSize <= 0 {
			addRefusal("invalid_expected_profile", stylePath+".font_size", "font_size must be positive")
		}
		if len(style.FillColor) > 0 {
			validateColor(style.FillColor, stylePath+".fill_color", "invalid_expected_profile", addRefusal)
		}
		if style.Tsume != nil && (*style.Tsume < 0 || *style.Tsume > 100) {
			addRefusal("invalid_expected_profile", stylePath+".tsume", "tsume must be between 0 and 100")
		}
		if style.CapsOption != "" {
			if _, err := textCapsOption(style.CapsOption); err != nil {
				addRefusal("invalid_expected_profile", stylePath+".caps_option", "caps_option must be normal, small_caps, all_caps, or all_small_caps")
			}
		}
		if style.BaselineOption != "" {
			if _, err := textBaselineOption(style.BaselineOption); err != nil {
				addRefusal("invalid_expected_profile", stylePath+".baseline_option", "baseline_option must be normal, superscript, or subscript")
			}
		}
		if style.AutoKernType != "" {
			if _, err := textAutoKernType(style.AutoKernType); err != nil {
				addRefusal("invalid_expected_profile", stylePath+".auto_kern_type", "auto_kern_type must be no_auto, metric, or optical")
			}
		}
		if style.LineJoinType != "" {
			if _, err := textLineJoinType(style.LineJoinType); err != nil {
				addRefusal("invalid_expected_profile", stylePath+".line_join_type", "line_join_type must be miter, round, or bevel")
			}
		}
		if style.DigitSet != "" {
			if _, err := textDigitSet(style.DigitSet); err != nil {
				addRefusal("invalid_expected_profile", stylePath+".digit_set", "digit_set must be default, arabic, hindi, farsi, or arabic_rtl")
			}
		}
		if len(style.StrokeColor) > 0 {
			validateColor(style.StrokeColor, stylePath+".stroke_color", "invalid_expected_profile", addRefusal)
		}
		if style.StrokeWidth != nil && *style.StrokeWidth < 0 {
			addRefusal("invalid_expected_profile", stylePath+".stroke_width", "stroke_width must be non-negative")
		}
		if style.Justification != "" && !validTextJustification(style.Justification) {
			addRefusal("invalid_expected_profile", stylePath+".justification", "justification must be left, right, or center")
		}
	}
	for i, keyframed := range expected.Keyframes {
		kfPropPath := fmt.Sprintf("expected_profile.keyframes[%d]", i)
		if keyframed.LayerName == "" {
			addRefusal("invalid_expected_profile", kfPropPath+".layer_name", "layer_name is required")
		}
		if keyframed.MatchName == "" {
			addRefusal("invalid_expected_profile", kfPropPath+".match_name", "property match_name is required")
		}
		if len(keyframed.Keyframes) == 0 {
			addRefusal("invalid_expected_profile", kfPropPath+".keyframes", "at least one keyframe is required")
		}
		for ki, kf := range keyframed.Keyframes {
			kfPath := fmt.Sprintf("%s.keyframes[%d]", kfPropPath, ki)
			if kf.Time < 0 {
				addRefusal("invalid_expected_profile", kfPath+".time", "keyframe time must be non-negative")
			}
			if ki > 0 && kf.Time < keyframed.Keyframes[ki-1].Time {
				addRefusal("invalid_expected_profile", kfPath+".time", "keyframes must be sorted by time")
			}
			if !validEffectParamValue(kf.Value) {
				addRefusal("invalid_expected_profile", kfPath+".value", "expected keyframe value must be a number, boolean, or numeric array")
			}
		}
	}
	for i, mask := range expected.Masks {
		maskPath := fmt.Sprintf("expected_profile.masks[%d]", i)
		if mask.LayerName == "" {
			addRefusal("invalid_expected_profile", maskPath+".layer_name", "layer_name is required")
		}
		if mask.Mode != "" && !validMaskMode(mask.Mode) {
			addRefusal("invalid_expected_profile", maskPath+".mode", "mask mode is not supported")
		}
		if len(mask.Color) > 0 {
			validateRGBColor(mask.Color, maskPath+".color", "invalid_expected_profile", addRefusal)
		}
		if mask.MotionBlur != "" && !validMaskMotionBlur(mask.MotionBlur) {
			addRefusal("invalid_expected_profile", maskPath+".motion_blur", "mask motion_blur is not supported")
		}
		if mask.FeatherFalloff != "" && !validMaskFeatherFalloff(mask.FeatherFalloff) {
			addRefusal("invalid_expected_profile", maskPath+".feather_falloff", "mask feather_falloff is not supported")
		}
		if mask.Opacity != nil && (*mask.Opacity < 0 || *mask.Opacity > 1) {
			addRefusal("invalid_expected_profile", maskPath+".opacity", "mask opacity must be between 0 and 1")
		}
		validateMaskFeather(mask.Feather, maskPath+".feather", "invalid_expected_profile", addRefusal)
		if mask.VertexCount != nil && *mask.VertexCount < 0 {
			addRefusal("invalid_expected_profile", maskPath+".vertex_count", "vertex_count must be non-negative")
		}
		for ki, kf := range mask.PathKeyframes {
			kfPath := fmt.Sprintf("%s.path_keyframes[%d]", maskPath, ki)
			if kf.Time < 0 {
				addRefusal("invalid_expected_profile", kfPath+".time", "mask path keyframe time must be non-negative")
			}
			if ki > 0 && kf.Time < mask.PathKeyframes[ki-1].Time {
				addRefusal("invalid_expected_profile", kfPath+".time", "mask path keyframes must be sorted by time")
			}
			if kf.VertexCount != nil && *kf.VertexCount < 0 {
				addRefusal("invalid_expected_profile", kfPath+".vertex_count", "vertex_count must be non-negative")
			}
		}
	}
}
