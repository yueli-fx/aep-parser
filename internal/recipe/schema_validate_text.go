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
	if style.Tracking != nil {
		recordCapability("Layer.SetRunTracking", stylePath+".tracking")
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
	if style.Justification != "" {
		recordCapability("Layer.SetParagraphJustification", stylePath+".justification")
		if !validTextJustification(style.Justification) {
			addRefusal("invalid_text_justification", stylePath+".justification", "justification must be left, right, or center")
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
	default:
		addRefusal("unsupported_text_animator_property", path+".property", "text animator property must be opacity, position, scale, rotation, color, tracking, or character_offset")
	}
	if animator.Value == nil {
		addRefusal("missing_text_animator_value", path+".value", "text animator value is required")
	} else {
		switch animator.Property {
		case "opacity":
			if _, ok := animator.Value.(float64); !ok {
				addRefusal("invalid_text_animator_value", path+".value", "text animator opacity value must be a number")
			}
		case "rotation", "tracking", "character_offset":
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
		case "color":
			values, ok := numericSliceValue(animator.Value)
			if !ok {
				addRefusal("invalid_text_animator_value", path+".value", "text animator color value must be a 3- or 4-number array")
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
		if len(animator.RangeOffsetKeyframes) < 2 {
			addRefusal("invalid_text_animator_range_offset_keyframes", path+".range_offset_keyframes", "range_offset_keyframes must include at least 2 keyframes")
		}
		for i, kf := range animator.RangeOffsetKeyframes {
			kfPath := fmt.Sprintf("%s.range_offset_keyframes[%d]", path, i)
			if kf.Time < 0 {
				addRefusal("keyframe_time_out_of_range", kfPath+".time", "keyframe time must be non-negative")
			}
			if i > 0 && kf.Time < animator.RangeOffsetKeyframes[i-1].Time {
				addRefusal("keyframes_not_sorted", kfPath+".time", "keyframes must be sorted by time")
			}
			validateKeyframeEase(kf.InEase, kfPath+".in_ease", addRefusal)
			validateKeyframeEase(kf.OutEase, kfPath+".out_ease", addRefusal)
		}
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
