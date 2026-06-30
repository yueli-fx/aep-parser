package recipe

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func validateTextStyle(style TextStyleSpec, stylePath string, recordCapability func(string, string) CapabilityLookup, addRefusal func(string, string, string)) {
	if style.RunIndex < 0 {
		addRefusal("invalid_text_style_run_index", stylePath+".run_index", "run_index must be non-negative")
	}
	if style.ParagraphIndex < 0 {
		addRefusal("invalid_text_style_paragraph_index", stylePath+".paragraph_index", "paragraph_index must be non-negative")
	}
	if style.FontSize != nil {
		recordCapability("Layer.SetRunFontSize", stylePath+".font_size")
		if *style.FontSize <= 0 {
			addRefusal("invalid_text_font_size", stylePath+".font_size", "font_size must be positive")
		}
	}
	if len(style.FillColor) > 0 {
		recordCapability("Layer.SetRunFillColor", stylePath+".fill_color")
		validateColor(style.FillColor, stylePath+".fill_color", "invalid_text_fill_color", addRefusal)
	}
	if style.AutoLeading != nil {
		recordCapability("Layer.SetRunAutoLeading", stylePath+".auto_leading")
	}
	if style.Leading != nil {
		recordCapability("Layer.SetRunLeading", stylePath+".leading")
	}
	if style.Tracking != nil {
		recordCapability("Layer.SetRunTracking", stylePath+".tracking")
	}
	if style.BaselineShift != nil {
		recordCapability("Layer.SetRunBaselineShift", stylePath+".baseline_shift")
	}
	if style.HorizontalScale != nil {
		recordCapability("Layer.SetRunHorizontalScale", stylePath+".horizontal_scale")
		if *style.HorizontalScale < 0 {
			addRefusal("invalid_text_horizontal_scale", stylePath+".horizontal_scale", "horizontal_scale must be non-negative")
		}
	}
	if style.VerticalScale != nil {
		recordCapability("Layer.SetRunVerticalScale", stylePath+".vertical_scale")
		if *style.VerticalScale < 0 {
			addRefusal("invalid_text_vertical_scale", stylePath+".vertical_scale", "vertical_scale must be non-negative")
		}
	}
	if style.Tsume != nil {
		recordCapability("Layer.SetRunTsume", stylePath+".tsume")
		if *style.Tsume < 0 || *style.Tsume > 100 {
			addRefusal("invalid_text_tsume", stylePath+".tsume", "tsume must be between 0 and 100")
		}
	}
	if style.CapsOption != "" {
		recordCapability("Layer.SetRunCapsOption", stylePath+".caps_option")
		if _, err := textCapsOption(style.CapsOption); err != nil {
			addRefusal("invalid_text_caps_option", stylePath+".caps_option", "caps_option must be normal, small_caps, all_caps, or all_small_caps")
		}
	}
	if style.BaselineOption != "" {
		recordCapability("Layer.SetRunBaselineOption", stylePath+".baseline_option")
		if _, err := textBaselineOption(style.BaselineOption); err != nil {
			addRefusal("invalid_text_baseline_option", stylePath+".baseline_option", "baseline_option must be normal, superscript, or subscript")
		}
	}
	if style.AutoKernType != "" {
		recordCapability("Layer.SetRunAutoKernType", stylePath+".auto_kern_type")
		if _, err := textAutoKernType(style.AutoKernType); err != nil {
			addRefusal("invalid_text_auto_kern_type", stylePath+".auto_kern_type", "auto_kern_type must be no_auto, metric, or optical")
		}
	}
	if style.LineJoinType != "" {
		recordCapability("Layer.SetRunLineJoinType", stylePath+".line_join_type")
		if _, err := textLineJoinType(style.LineJoinType); err != nil {
			addRefusal("invalid_text_line_join_type", stylePath+".line_join_type", "line_join_type must be miter, round, or bevel")
		}
	}
	if style.DigitSet != "" {
		recordCapability("Layer.SetRunDigitSet", stylePath+".digit_set")
		if _, err := textDigitSet(style.DigitSet); err != nil {
			addRefusal("invalid_text_digit_set", stylePath+".digit_set", "digit_set must be default, arabic, hindi, farsi, or arabic_rtl")
		}
	}
	if style.NoBreak != nil {
		recordCapability("Layer.SetRunNoBreak", stylePath+".no_break")
	}
	if style.FauxBold != nil {
		recordCapability("Layer.SetRunFauxBold", stylePath+".faux_bold")
	}
	if style.FauxItalic != nil {
		recordCapability("Layer.SetRunFauxItalic", stylePath+".faux_italic")
	}
	if style.ApplyStroke != nil {
		recordCapability("Layer.SetRunApplyStroke", stylePath+".apply_stroke")
	}
	if len(style.StrokeColor) > 0 {
		recordCapability("Layer.SetRunStrokeColor", stylePath+".stroke_color")
		validateColor(style.StrokeColor, stylePath+".stroke_color", "invalid_text_stroke_color", addRefusal)
	}
	if style.StrokeWidth != nil {
		recordCapability("Layer.SetRunStrokeWidth", stylePath+".stroke_width")
		if *style.StrokeWidth < 0 {
			addRefusal("invalid_text_stroke_width", stylePath+".stroke_width", "stroke_width must be non-negative")
		}
	}
	if style.StrokeOverFill != nil {
		recordCapability("Layer.SetRunStrokeOverFill", stylePath+".stroke_over_fill")
	}
	if style.Justification != "" {
		recordCapability("Layer.SetParagraphJustification", stylePath+".justification")
		if !validTextJustification(style.Justification) {
			addRefusal("invalid_text_justification", stylePath+".justification", "justification must be left, right, or center")
		}
	}
	if style.FirstLineIndent != nil {
		recordCapability("Layer.SetParagraphFirstLineIndent", stylePath+".first_line_indent")
	}
	if style.StartIndent != nil {
		recordCapability("Layer.SetParagraphStartIndent", stylePath+".start_indent")
	}
	if style.EndIndent != nil {
		recordCapability("Layer.SetParagraphEndIndent", stylePath+".end_indent")
	}
	if style.SpaceBefore != nil {
		recordCapability("Layer.SetParagraphSpaceBefore", stylePath+".space_before")
	}
	if style.SpaceAfter != nil {
		recordCapability("Layer.SetParagraphSpaceAfter", stylePath+".space_after")
	}
	if style.AutoHyphenate != nil {
		recordCapability("Layer.SetParagraphAutoHyphenate", stylePath+".auto_hyphenate")
	}
	if style.LeadingType != "" {
		recordCapability("Layer.SetParagraphLeadingType", stylePath+".leading_type")
		if _, err := textLeadingType(style.LeadingType); err != nil {
			addRefusal("invalid_text_leading_type", stylePath+".leading_type", "leading_type must be roman or japanese")
		}
	}
	if style.HangingRoman != nil {
		recordCapability("Layer.SetParagraphHangingRoman", stylePath+".hanging_roman")
	}
	if style.ParagraphDirection != "" {
		recordCapability("Layer.SetParagraphDirection", stylePath+".paragraph_direction")
		if _, err := textParagraphDirection(style.ParagraphDirection); err != nil {
			addRefusal("invalid_text_paragraph_direction", stylePath+".paragraph_direction", "paragraph_direction must be ltr or rtl")
		}
	}
}

func validTextJustification(value string) bool {
	switch value {
	case "left", "right", "center":
		return true
	default:
		return false
	}
}

func validOffsetLineJoin(value string) bool {
	switch value {
	case "miter", "round", "bevel":
		return true
	default:
		return false
	}
}

func validStrokeLineCap(value string) bool {
	switch value {
	case "butt", "round", "projecting":
		return true
	default:
		return false
	}
}

func validStrokeLineJoin(value string) bool {
	switch value {
	case "miter", "round", "bevel":
		return true
	default:
		return false
	}
}

func validShapeCompositeOrder(value string) bool {
	switch value {
	case "above_previous", "below_previous":
		return true
	default:
		return false
	}
}

func validShapeBlendMode(value float64) bool {
	return value >= 1 && isWholeNumber(value)
}

func validFillRule(value string) bool {
	switch value {
	case "nonzero_winding", "even_odd":
		return true
	default:
		return false
	}
}

func validGradientType(value string) bool {
	switch value {
	case "linear", "radial":
		return true
	default:
		return false
	}
}

func validRepeaterOrder(value string) bool {
	switch value {
	case "below", "above":
		return true
	default:
		return false
	}
}

func validMergePathsType(value string) bool {
	switch value {
	case "merge", "add", "subtract", "intersect", "exclude":
		return true
	default:
		return false
	}
}

func validZigZagPoints(value string) bool {
	switch value {
	case "corner", "smooth":
		return true
	default:
		return false
	}
}

func validRoughenPoints(value string) bool {
	switch value {
	case "corner", "smooth":
		return true
	default:
		return false
	}
}

func usesTransform(t Transform) bool {
	return len(t.Position) > 0 ||
		len(t.Scale) > 0 ||
		len(t.AnchorPoint) > 0 ||
		t.Rotation != nil ||
		t.Opacity != nil ||
		len(t.PositionKeyframes) > 0 ||
		len(t.AnchorPointKeyframes) > 0 ||
		len(t.ScaleKeyframes) > 0 ||
		len(t.RotationKeyframes) > 0 ||
		len(t.OpacityKeyframes) > 0
}

func validateVec(values []float64, want int, path string, addRefusal func(string, string, string)) {
	if len(values) == 0 {
		return
	}
	if len(values) != want {
		addRefusal("invalid_vector_size", path, fmt.Sprintf("expected %d values", want))
	}
}

func validateKeyframeEase(ease *TemporalEase, path string, addRefusal func(string, string, string)) {
	if ease == nil {
		return
	}
	if ease.Influence <= 0 || ease.Influence > 1 {
		addRefusal("invalid_keyframe_ease_influence", path+".influence", "keyframe ease influence must be greater than 0 and at most 1")
	}
}

func validMaskMode(value string) bool {
	switch value {
	case "none", "add", "subtract", "intersect", "lighten", "darken", "difference":
		return true
	default:
		return false
	}
}

func validMaskMotionBlur(value string) bool {
	switch value {
	case "same_as_layer", "on", "off":
		return true
	default:
		return false
	}
}

func validMaskFeatherFalloff(value string) bool {
	switch value {
	case "smooth", "linear":
		return true
	default:
		return false
	}
}

func validateMaskFeather(values []float64, path, code string, addRefusal func(string, string, string)) {
	if len(values) == 0 {
		return
	}
	if len(values) != 2 {
		addRefusal(code, path, "mask feather must have 2 values")
		return
	}
	for i, value := range values {
		if value < 0 {
			addRefusal(code, fmt.Sprintf("%s[%d]", path, i), "mask feather values must be non-negative")
		}
	}
}

func validateTransformExpressions(expressions TransformExpressions, path string, recordCapability func(string, string) CapabilityLookup, addRefusal func(string, string, string)) {
	validateTransformExpression(expressions.Position, path+".position", recordCapability, addRefusal)
	validateTransformExpression(expressions.AnchorPoint, path+".anchor_point", recordCapability, addRefusal)
	validateTransformExpression(expressions.Scale, path+".scale", recordCapability, addRefusal)
	validateTransformExpression(expressions.Rotation, path+".rotation", recordCapability, addRefusal)
	validateTransformExpression(expressions.Opacity, path+".opacity", recordCapability, addRefusal)
}

func validateTransformExpression(expression *ExpressionSpec, path string, recordCapability func(string, string) CapabilityLookup, addRefusal func(string, string, string)) {
	if expression == nil {
		return
	}
	recordCapability("Property.SetExpression", path+".source")
	if expression.Source == "" {
		addRefusal("missing_transform_expression_source", path+".source", "transform expression source is required")
	}
	if expression.Enabled != nil {
		recordCapability("Property.SetExpressionEnabled", path+".enabled")
	}
}

func validateTextAnimator(animator TextAnimatorSpec, path string, recordCapability func(string, string) CapabilityLookup, addRefusal func(string, string, string)) {
	switch animator.Property {
	case "opacity":
		recordCapability("AddTextOpacityAnimator", path)
	case "position":
		recordCapability("AddTextPositionAnimator", path)
	case "scale":
		recordCapability("AddTextScaleAnimator", path)
	case "rotation":
		recordCapability("AddTextRotationAnimator", path)
	case "color":
		recordCapability("AddTextColorAnimator", path)
	case "tracking":
		recordCapability("AddTextTrackingAnimator", path)
	case "character_offset":
		recordCapability("AddTextCharacterOffsetAnimator", path)
	case "fill_opacity":
		recordCapability("AddTextFillOpacityAnimator", path)
	case "stroke_opacity":
		recordCapability("AddTextStrokeOpacityAnimator", path)
	case "stroke_width":
		recordCapability("AddTextStrokeWidthAnimator", path)
	case "skew":
		recordCapability("AddTextSkewAnimator", path)
	case "rotation_x":
		recordCapability("AddTextRotationXAnimator", path)
	case "rotation_y":
		recordCapability("AddTextRotationYAnimator", path)
	case "stroke_color":
		recordCapability("AddTextStrokeColorAnimator", path)
	default:
		addRefusal("unsupported_text_animator_property", path+".property", "text animator property must be opacity, position, scale, rotation, color, tracking, character_offset, fill_opacity, stroke_opacity, stroke_width, skew, rotation_x, rotation_y, or stroke_color")
	}
	if animator.Value == nil {
		addRefusal("missing_text_animator_value", path+".value", "text animator value is required")
	} else {
		switch animator.Property {
		case "opacity":
			if _, ok := animator.Value.(float64); !ok {
				addRefusal("invalid_text_animator_value", path+".value", "text animator opacity value must be a number")
			}
		case "rotation", "tracking", "character_offset", "fill_opacity", "stroke_opacity", "stroke_width", "skew", "rotation_x", "rotation_y":
			if _, ok := animator.Value.(float64); !ok {
				addRefusal("invalid_text_animator_value", path+".value", "text animator "+animator.Property+" value must be a number")
			}
		case "position":
			values, ok := numericSliceValue(animator.Value)
			if !ok || len(values) != 3 {
				addRefusal("invalid_text_animator_value", path+".value", "text animator position value must be a 3-number array")
			}
		case "scale":
			values, ok := numericSliceValue(animator.Value)
			if !ok || len(values) != 3 {
				addRefusal("invalid_text_animator_value", path+".value", "text animator scale value must be a 3-number array")
			}
		case "color", "stroke_color":
			values, ok := numericSliceValue(animator.Value)
			if !ok {
				addRefusal("invalid_text_animator_value", path+".value", "text animator "+animator.Property+" value must be a 3- or 4-number array")
			} else {
				validateColor(values, path+".value", "invalid_text_animator_value", addRefusal)
			}
		}
	}
	if animator.RangeStart == nil {
		addRefusal("missing_text_animator_range_start", path+".range_start", "range_start is required")
	}
	if animator.RangeEnd == nil {
		addRefusal("missing_text_animator_range_end", path+".range_end", "range_end is required")
	}
	if animator.RangeOffset == nil {
		addRefusal("missing_text_animator_range_offset", path+".range_offset", "range_offset is required")
	}
	if len(animator.RangeOffsetKeyframes) > 0 {
		recordCapability("AnimateTextRangeOffset", path+".range_offset_keyframes")
		validateScalarKeyframes(animator.RangeOffsetKeyframes, path+".range_offset_keyframes", "range_offset_keyframes", addRefusal)
	}
	if len(animator.ValueKeyframes) > 0 {
		switch animator.Property {
		case "opacity":
			recordCapability("AnimateTextOpacity", path+".value_keyframes")
			validateValueKeyframes(animator.ValueKeyframes, path+".value_keyframes", "scalar", 0, addRefusal)
		case "position":
			recordCapability("AnimateTextPosition", path+".value_keyframes")
			validateValueKeyframes(animator.ValueKeyframes, path+".value_keyframes", "vector", 3, addRefusal)
		case "scale":
			recordCapability("AnimateTextScale", path+".value_keyframes")
			validateValueKeyframes(animator.ValueKeyframes, path+".value_keyframes", "vector", 3, addRefusal)
		case "rotation":
			recordCapability("AnimateTextRotation", path+".value_keyframes")
			validateValueKeyframes(animator.ValueKeyframes, path+".value_keyframes", "scalar", 0, addRefusal)
		case "color":
			recordCapability("AnimateTextColor", path+".value_keyframes")
			validateValueKeyframes(animator.ValueKeyframes, path+".value_keyframes", "color", 0, addRefusal)
		case "tracking":
			recordCapability("AnimateTextTracking", path+".value_keyframes")
			validateValueKeyframes(animator.ValueKeyframes, path+".value_keyframes", "scalar", 0, addRefusal)
		case "character_offset":
			recordCapability("AnimateTextCharacterOffset", path+".value_keyframes")
			validateValueKeyframes(animator.ValueKeyframes, path+".value_keyframes", "scalar", 0, addRefusal)
		default:
			addRefusal("unsupported_text_animator_value_keyframes", path+".value_keyframes", "value_keyframes currently support opacity, position, scale, rotation, color, tracking, and character_offset text animators")
			validateValueKeyframes(animator.ValueKeyframes, path+".value_keyframes", "any", 0, addRefusal)
		}
	}
}

func validateScalarKeyframes(keyframes []ScalarKeyframe, path, label string, addRefusal func(string, string, string)) {
	if len(keyframes) < 2 {
		addRefusal("invalid_text_animator_keyframes", path, label+" must include at least 2 keyframes")
	}
	for i, kf := range keyframes {
		kfPath := fmt.Sprintf("%s[%d]", path, i)
		if kf.Time < 0 {
			addRefusal("keyframe_time_out_of_range", kfPath+".time", "keyframe time must be non-negative")
		}
		if i > 0 && kf.Time < keyframes[i-1].Time {
			addRefusal("keyframes_not_sorted", kfPath+".time", "keyframes must be sorted by time")
		}
		validateKeyframeEase(kf.InEase, kfPath+".in_ease", addRefusal)
		validateKeyframeEase(kf.OutEase, kfPath+".out_ease", addRefusal)
	}
}

func validateValueKeyframes(keyframes []ValueKeyframe, path, kind string, vectorSize int, addRefusal func(string, string, string)) {
	if len(keyframes) < 2 {
		addRefusal("invalid_text_animator_keyframes", path, "value_keyframes must include at least 2 keyframes")
	}
	for i, kf := range keyframes {
		kfPath := fmt.Sprintf("%s[%d]", path, i)
		if kf.Time < 0 {
			addRefusal("keyframe_time_out_of_range", kfPath+".time", "keyframe time must be non-negative")
		}
		if i > 0 && kf.Time < keyframes[i-1].Time {
			addRefusal("keyframes_not_sorted", kfPath+".time", "keyframes must be sorted by time")
		}
		switch kind {
		case "scalar":
			if _, ok := numericValue(kf.Value); !ok {
				addRefusal("invalid_text_animator_keyframe_value", kfPath+".value", "value_keyframes value must be a number for this text animator property")
			}
		case "vector":
			values, ok := numericSliceValue(kf.Value)
			if !ok || len(values) != vectorSize {
				addRefusal("invalid_text_animator_keyframe_value", kfPath+".value", fmt.Sprintf("value_keyframes value must be a %d-number array for this text animator property", vectorSize))
			}
		case "color":
			values, ok := numericSliceValue(kf.Value)
			if !ok || (len(values) != 3 && len(values) != 4) {
				addRefusal("invalid_text_animator_keyframe_value", kfPath+".value", "value_keyframes value must be a 3- or 4-number color array for this text animator property")
			}
		}
		validateKeyframeEase(kf.InEase, kfPath+".in_ease", addRefusal)
		validateKeyframeEase(kf.OutEase, kfPath+".out_ease", addRefusal)
	}
}

func numericValue(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint64:
		return float64(v), true
	default:
		return 0, false
	}
}

func numericSliceValue(value any) ([]float64, bool) {
	switch v := value.(type) {
	case []float64:
		return v, true
	case []any:
		out := make([]float64, 0, len(v))
		for _, item := range v {
			n, ok := item.(float64)
			if !ok {
				return nil, false
			}
			out = append(out, n)
		}
		return out, true
	default:
		return nil, false
	}
}

func validateColor(values []float64, path, code string, addRefusal func(string, string, string)) {
	if len(values) != 3 && len(values) != 4 {
		addRefusal(code, path, "color must have 3 or 4 channels")
		return
	}
	for i, value := range values {
		if value < 0 || value > 255 {
			addRefusal(code, fmt.Sprintf("%s[%d]", path, i), "color channels must be between 0 and 255")
		}
	}
}

func validateGradientColorStops(stops []GradientColorStopSpec, path, code, label string, addRefusal func(string, string, string)) {
	if len(stops) < 2 {
		addRefusal(code, path, label+" color_stops must include at least 2 stops")
	}
	for i, stop := range stops {
		stopPath := fmt.Sprintf("%s[%d]", path, i)
		if stop.Offset < 0 || stop.Offset > 1 {
			addRefusal(code, stopPath+".offset", label+" stop offset must be between 0 and 1")
		}
		if stop.Midpoint != nil && (*stop.Midpoint < 0 || *stop.Midpoint > 1) {
			addRefusal(code, stopPath+".midpoint", label+" stop midpoint must be between 0 and 1")
		}
		validateColor(stop.Color, stopPath+".color", code, addRefusal)
	}
}

func validateGradientAlphaStops(stops []GradientAlphaStopSpec, path, code, label string, addRefusal func(string, string, string)) {
	if len(stops) < 2 {
		addRefusal(code, path, label+" alpha_stops must include at least 2 stops")
	}
	for i, stop := range stops {
		stopPath := fmt.Sprintf("%s[%d]", path, i)
		if stop.Offset < 0 || stop.Offset > 1 {
			addRefusal(code, stopPath+".offset", label+" alpha stop offset must be between 0 and 1")
		}
		if stop.Midpoint != nil && (*stop.Midpoint < 0 || *stop.Midpoint > 1) {
			addRefusal(code, stopPath+".midpoint", label+" alpha stop midpoint must be between 0 and 1")
		}
		if stop.Alpha < 0 || stop.Alpha > 1 {
			addRefusal(code, stopPath+".alpha", label+" alpha stop alpha must be between 0 and 1")
		}
	}
}

func validateRGBColor(values []float64, path, code string, addRefusal func(string, string, string)) {
	if len(values) != 3 {
		addRefusal(code, path, "color must have 3 channels")
		return
	}
	for i, value := range values {
		if value < 0 || value > 255 {
			addRefusal(code, fmt.Sprintf("%s[%d]", path, i), "color channels must be between 0 and 255")
		}
	}
}

func isWholeNumber(value float64) bool {
	return value == float64(int64(value))
}

func validEffectParamValue(value any) bool {
	switch v := value.(type) {
	case float64, int, bool:
		return true
	case []float64:
		return len(v) > 0
	case []any:
		if len(v) == 0 {
			return false
		}
		for _, item := range v {
			if _, ok := item.(float64); !ok {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func supportedEffect(matchName string) bool {
	for _, effect := range aep.SupportedEffects() {
		if effect == matchName {
			return true
		}
	}
	return false
}
