package aepmigrate

import (
	"fmt"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeTextStyle(layer *aep.Layer, source *profile.TextSource) error {
	if source == nil {
		return nil
	}
	for i, run := range source.Runs {
		if err := layer.SetRunFontSize(i, run.FontSize); err != nil {
			return fmt.Errorf("run %d font_size: %w", i, err)
		}
		if err := layer.SetRunFillColor(i, run.FillColor); err != nil {
			return fmt.Errorf("run %d fill_color: %w", i, err)
		}
		if err := layer.SetRunAutoLeading(i, run.AutoLeading); err != nil {
			return fmt.Errorf("run %d auto_leading: %w", i, err)
		}
		if err := layer.SetRunLeading(i, run.Leading); err != nil {
			return fmt.Errorf("run %d leading: %w", i, err)
		}
		if err := layer.SetRunTracking(i, run.Tracking); err != nil {
			return fmt.Errorf("run %d tracking: %w", i, err)
		}
		if err := layer.SetRunBaselineShift(i, run.BaselineShift); err != nil {
			return fmt.Errorf("run %d baseline_shift: %w", i, err)
		}
		if err := layer.SetRunHorizontalScale(i, run.HorizontalScale); err != nil {
			return fmt.Errorf("run %d horizontal_scale: %w", i, err)
		}
		if err := layer.SetRunVerticalScale(i, run.VerticalScale); err != nil {
			return fmt.Errorf("run %d vertical_scale: %w", i, err)
		}
		if err := layer.SetRunTsume(i, run.Tsume); err != nil {
			return fmt.Errorf("run %d tsume: %w", i, err)
		}
		caps, err := textCapsOption(run.CapsOption)
		if err != nil {
			return fmt.Errorf("run %d caps_option: %w", i, err)
		}
		if err := layer.SetRunCapsOption(i, caps); err != nil {
			return fmt.Errorf("run %d caps_option: %w", i, err)
		}
		baseline, err := textBaselineOption(run.BaselineOption)
		if err != nil {
			return fmt.Errorf("run %d baseline_option: %w", i, err)
		}
		if err := layer.SetRunBaselineOption(i, baseline); err != nil {
			return fmt.Errorf("run %d baseline_option: %w", i, err)
		}
		kern, err := textAutoKernType(run.AutoKernType)
		if err != nil {
			return fmt.Errorf("run %d auto_kern_type: %w", i, err)
		}
		if err := layer.SetRunAutoKernType(i, kern); err != nil {
			return fmt.Errorf("run %d auto_kern_type: %w", i, err)
		}
		lineJoin, err := textLineJoinType(run.LineJoinType)
		if err != nil {
			return fmt.Errorf("run %d line_join_type: %w", i, err)
		}
		if err := layer.SetRunLineJoinType(i, lineJoin); err != nil {
			return fmt.Errorf("run %d line_join_type: %w", i, err)
		}
		digitSet, err := textDigitSet(run.DigitSet)
		if err != nil {
			return fmt.Errorf("run %d digit_set: %w", i, err)
		}
		if err := layer.SetRunDigitSet(i, digitSet); err != nil {
			return fmt.Errorf("run %d digit_set: %w", i, err)
		}
		if err := layer.SetRunNoBreak(i, run.NoBreak); err != nil {
			return fmt.Errorf("run %d no_break: %w", i, err)
		}
		if err := layer.SetRunFauxBold(i, run.FauxBold); err != nil {
			return fmt.Errorf("run %d faux_bold: %w", i, err)
		}
		if err := layer.SetRunFauxItalic(i, run.FauxItalic); err != nil {
			return fmt.Errorf("run %d faux_italic: %w", i, err)
		}
		if err := layer.SetRunApplyStroke(i, run.ApplyStroke); err != nil {
			return fmt.Errorf("run %d apply_stroke: %w", i, err)
		}
		if err := layer.SetRunStrokeColor(i, run.StrokeColor); err != nil {
			return fmt.Errorf("run %d stroke_color: %w", i, err)
		}
		if err := layer.SetRunStrokeWidth(i, run.StrokeWidth); err != nil {
			return fmt.Errorf("run %d stroke_width: %w", i, err)
		}
		if err := layer.SetRunStrokeOverFill(i, run.StrokeOverFill); err != nil {
			return fmt.Errorf("run %d stroke_over_fill: %w", i, err)
		}
	}
	for i, paragraph := range source.Paragraphs {
		justification, err := textJustification(paragraph.Justification)
		if err != nil {
			return fmt.Errorf("paragraph %d justification: %w", i, err)
		}
		if err := layer.SetParagraphJustification(i, justification); err != nil {
			return fmt.Errorf("paragraph %d justification: %w", i, err)
		}
		if err := layer.SetParagraphFirstLineIndent(i, paragraph.FirstLineIndent); err != nil {
			return fmt.Errorf("paragraph %d first_line_indent: %w", i, err)
		}
		if err := layer.SetParagraphStartIndent(i, paragraph.StartIndent); err != nil {
			return fmt.Errorf("paragraph %d start_indent: %w", i, err)
		}
		if err := layer.SetParagraphEndIndent(i, paragraph.EndIndent); err != nil {
			return fmt.Errorf("paragraph %d end_indent: %w", i, err)
		}
		if err := layer.SetParagraphSpaceBefore(i, paragraph.SpaceBefore); err != nil {
			return fmt.Errorf("paragraph %d space_before: %w", i, err)
		}
		if err := layer.SetParagraphSpaceAfter(i, paragraph.SpaceAfter); err != nil {
			return fmt.Errorf("paragraph %d space_after: %w", i, err)
		}
		if err := layer.SetParagraphAutoHyphenate(i, paragraph.AutoHyphenate); err != nil {
			return fmt.Errorf("paragraph %d auto_hyphenate: %w", i, err)
		}
		leadingType, err := textLeadingType(paragraph.LeadingType)
		if err != nil {
			return fmt.Errorf("paragraph %d leading_type: %w", i, err)
		}
		if err := layer.SetParagraphLeadingType(i, leadingType); err != nil {
			return fmt.Errorf("paragraph %d leading_type: %w", i, err)
		}
		if err := layer.SetParagraphHangingRoman(i, paragraph.HangingRoman); err != nil {
			return fmt.Errorf("paragraph %d hanging_roman: %w", i, err)
		}
		direction, err := textParagraphDirection(paragraph.Direction)
		if err != nil {
			return fmt.Errorf("paragraph %d direction: %w", i, err)
		}
		if err := layer.SetParagraphDirection(i, direction); err != nil {
			return fmt.Errorf("paragraph %d direction: %w", i, err)
		}
	}
	return nil
}

func normalizeTextEnum(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.ReplaceAll(value, "-", "_")
}

func textJustification(value string) (aep.TextJustification, error) {
	switch normalizeTextEnum(value) {
	case "", "left":
		return aep.TextJustifyLeft, nil
	case "right":
		return aep.TextJustifyRight, nil
	case "center":
		return aep.TextJustifyCenter, nil
	default:
		return 0, fmt.Errorf("unsupported justification %q", value)
	}
}

func textCapsOption(value string) (aep.TextCapsOption, error) {
	switch normalizeTextEnum(value) {
	case "", "normal":
		return aep.TextCapsNormal, nil
	case "small_caps":
		return aep.TextCapsSmall, nil
	case "all_caps":
		return aep.TextCapsAll, nil
	case "all_small_caps":
		return aep.TextCapsAllSmall, nil
	default:
		return 0, fmt.Errorf("unsupported caps_option %q", value)
	}
}

func textBaselineOption(value string) (aep.TextBaselineOption, error) {
	switch normalizeTextEnum(value) {
	case "", "normal":
		return aep.TextBaselineNormal, nil
	case "superscript":
		return aep.TextBaselineSuperscript, nil
	case "subscript":
		return aep.TextBaselineSubscript, nil
	default:
		return 0, fmt.Errorf("unsupported baseline_option %q", value)
	}
}

func textAutoKernType(value string) (aep.TextAutoKernType, error) {
	switch normalizeTextEnum(value) {
	case "no_auto":
		return aep.TextAutoKernNoAuto, nil
	case "", "metric":
		return aep.TextAutoKernMetric, nil
	case "optical":
		return aep.TextAutoKernOptical, nil
	default:
		return 0, fmt.Errorf("unsupported auto_kern_type %q", value)
	}
}

func textLineJoinType(value string) (aep.TextLineJoinType, error) {
	switch normalizeTextEnum(value) {
	case "", "miter":
		return aep.TextLineJoinMiter, nil
	case "round":
		return aep.TextLineJoinRound, nil
	case "bevel":
		return aep.TextLineJoinBevel, nil
	default:
		return 0, fmt.Errorf("unsupported line_join_type %q", value)
	}
}

func textDigitSet(value string) (aep.TextDigitSet, error) {
	switch normalizeTextEnum(value) {
	case "", "default":
		return aep.TextDigitSetDefault, nil
	case "arabic":
		return aep.TextDigitSetArabic, nil
	case "hindi":
		return aep.TextDigitSetHindi, nil
	case "farsi":
		return aep.TextDigitSetFarsi, nil
	case "arabic_rtl":
		return aep.TextDigitSetArabicRTL, nil
	default:
		return 0, fmt.Errorf("unsupported digit_set %q", value)
	}
}

func textLeadingType(value string) (aep.TextLeadingType, error) {
	switch normalizeTextEnum(value) {
	case "", "roman":
		return aep.TextLeadingRoman, nil
	case "japanese":
		return aep.TextLeadingJapanese, nil
	default:
		return 0, fmt.Errorf("unsupported leading_type %q", value)
	}
}

func textParagraphDirection(value string) (aep.TextParagraphDirection, error) {
	switch normalizeTextEnum(value) {
	case "", "ltr":
		return aep.TextDirectionLeftToRight, nil
	case "rtl":
		return aep.TextDirectionRightToLeft, nil
	default:
		return 0, fmt.Errorf("unsupported paragraph_direction %q", value)
	}
}

func materializeProjectTextAnimators(project *aep.Project, prof *profile.Profile) (*aep.Project, error) {
	if !hasProfileTextAnimators(prof) {
		return project, nil
	}
	reopened, err := aep.Reopen(project)
	if err != nil {
		return nil, fmt.Errorf("reopen for text animators: %w", err)
	}
	for compIndex, sourceComp := range prof.Comps {
		targetComp, err := targetCompBySource(reopened, compIndex, sourceComp)
		if err != nil {
			return nil, err
		}
		for layerIndex, sourceLayer := range sourceComp.Layers {
			if !hasLayerTextAnimators(sourceLayer) {
				continue
			}
			if layerIndex >= len(targetComp.Layers) {
				return nil, fmt.Errorf("comp %q text animators: target layer index %d missing", sourceComp.Name, layerIndex)
			}
			if err := materializeLayerTextAnimators(targetComp.Layers[layerIndex], sourceLayer); err != nil {
				return nil, fmt.Errorf("comp %q layer %q text animators: %w", sourceComp.Name, sourceLayer.Name, err)
			}
		}
	}
	return reopened, nil
}

func hasProfileTextAnimators(prof *profile.Profile) bool {
	if prof == nil {
		return false
	}
	for _, comp := range prof.Comps {
		for _, layer := range comp.Layers {
			if hasLayerTextAnimators(layer) {
				return true
			}
		}
	}
	return false
}

func hasLayerTextAnimators(layer profile.Layer) bool {
	if layer.Type != "text" {
		return false
	}
	for _, property := range layer.Properties {
		if isTextAnimatorValueProperty(property.MatchName) {
			return true
		}
	}
	return false
}

func isTextAnimatorValueProperty(matchName string) bool {
	switch matchName {
	case "ADBE Text Opacity",
		"ADBE Text Position 3D",
		"ADBE Text Scale 3D",
		"ADBE Text Rotation",
		"ADBE Text Fill Color",
		"ADBE Text Stroke Color",
		"ADBE Text Tracking Amount",
		"ADBE Text Character Offset",
		"ADBE Text Fill Opacity",
		"ADBE Text Stroke Opacity",
		"ADBE Text Stroke Width",
		"ADBE Text Skew",
		"ADBE Text Rotation X",
		"ADBE Text Rotation Y":
		return true
	default:
		return false
	}
}

func materializeLayerTextAnimators(targetLayer *aep.Layer, sourceLayer profile.Layer) error {
	rangeValues, err := textAnimatorRangeValues(sourceLayer)
	if err != nil {
		return err
	}
	rangeOffsetAnimated := false
	for _, property := range sourceLayer.Properties {
		if !isTextAnimatorValueProperty(property.MatchName) {
			continue
		}
		if isImplicitTextRotationZProperty(sourceLayer, property) {
			continue
		}
		if err := materializeTextAnimator(targetLayer, property, rangeValues); err != nil {
			return fmt.Errorf("property %q: %w", property.MatchName, err)
		}
		if !rangeOffsetAnimated && len(rangeValues.offset.Keyframes) != 0 {
			if err := materializeTextRangeOffsetKeyframes(targetLayer, rangeValues.offset); err != nil {
				return fmt.Errorf("range offset keyframes: %w", err)
			}
			rangeOffsetAnimated = true
		}
		if len(property.Keyframes) != 0 {
			if err := materializeTextAnimatorValueKeyframes(targetLayer, property); err != nil {
				return fmt.Errorf("property %q keyframes: %w", property.MatchName, err)
			}
		}
	}
	return nil
}

func isImplicitTextRotationZProperty(layer profile.Layer, property profile.Property) bool {
	if property.MatchName != "ADBE Text Rotation" || len(property.Keyframes) != 0 || hasPropertyExpression(property) {
		return false
	}
	value, ok := propertyFloatValue(property.StaticValue)
	if !ok || value != 0 {
		return false
	}
	return hasProperty(layer, "ADBE Text Rotation X") || hasProperty(layer, "ADBE Text Rotation Y")
}

type textAnimatorRange struct {
	start  float64
	end    float64
	offset profile.Property
	value  float64
}

func textAnimatorRangeValues(layer profile.Layer) (textAnimatorRange, error) {
	startProperty, ok := propertyByMatchName(layer.Properties, "ADBE Text Percent Start")
	if !ok {
		return textAnimatorRange{}, fmt.Errorf("range start property missing")
	}
	start, ok := textAnimatorInitialFloatValue(startProperty)
	if !ok {
		return textAnimatorRange{}, fmt.Errorf("range start value unsupported (%T)", startProperty.StaticValue)
	}
	endProperty, ok := propertyByMatchName(layer.Properties, "ADBE Text Percent End")
	if !ok {
		return textAnimatorRange{}, fmt.Errorf("range end property missing")
	}
	end, ok := textAnimatorInitialFloatValue(endProperty)
	if !ok {
		return textAnimatorRange{}, fmt.Errorf("range end value unsupported (%T)", endProperty.StaticValue)
	}
	offsetProperty, ok := propertyByMatchName(layer.Properties, "ADBE Text Percent Offset")
	if !ok {
		return textAnimatorRange{}, fmt.Errorf("range offset property missing")
	}
	offset, ok := textAnimatorInitialFloatValue(offsetProperty)
	if !ok {
		return textAnimatorRange{}, fmt.Errorf("range offset value unsupported (%T)", offsetProperty.StaticValue)
	}
	return textAnimatorRange{
		start:  start,
		end:    end,
		offset: offsetProperty,
		value:  offset,
	}, nil
}

func materializeTextAnimator(targetLayer *aep.Layer, property profile.Property, rangeValues textAnimatorRange) error {
	switch property.MatchName {
	case "ADBE Text Opacity":
		value, ok := textAnimatorInitialFloatValue(property)
		if !ok {
			return fmt.Errorf("value unsupported (%T)", property.StaticValue)
		}
		_, err := aep.AddTextOpacityAnimator(targetLayer, value, rangeValues.start, rangeValues.end, rangeValues.value)
		return err
	case "ADBE Text Position 3D":
		value, ok := textAnimatorInitialVectorValue(property, 3)
		if !ok {
			return fmt.Errorf("value unsupported (%T)", property.StaticValue)
		}
		_, err := aep.AddTextPositionAnimator(targetLayer, value[0], value[1], value[2], rangeValues.start, rangeValues.end, rangeValues.value)
		return err
	case "ADBE Text Scale 3D":
		value, ok := textAnimatorInitialVectorValue(property, 3)
		if !ok {
			return fmt.Errorf("value unsupported (%T)", property.StaticValue)
		}
		_, err := aep.AddTextScaleAnimator(targetLayer, value[0], value[1], value[2], rangeValues.start, rangeValues.end, rangeValues.value)
		return err
	case "ADBE Text Rotation":
		value, ok := textAnimatorInitialFloatValue(property)
		if !ok {
			return fmt.Errorf("value unsupported (%T)", property.StaticValue)
		}
		_, err := aep.AddTextRotationAnimator(targetLayer, value, rangeValues.start, rangeValues.end, rangeValues.value)
		return err
	case "ADBE Text Fill Color":
		value, ok := textAnimatorInitialColorValue(property)
		if !ok {
			return fmt.Errorf("value unsupported (%T)", property.StaticValue)
		}
		_, err := aep.AddTextColorAnimator(targetLayer, value[0], value[1], value[2], value[3], rangeValues.start, rangeValues.end, rangeValues.value)
		return err
	case "ADBE Text Stroke Color":
		value, ok := textAnimatorInitialColorValue(property)
		if !ok {
			return fmt.Errorf("value unsupported (%T)", property.StaticValue)
		}
		_, err := aep.AddTextStrokeColorAnimator(targetLayer, value[0], value[1], value[2], value[3], rangeValues.start, rangeValues.end, rangeValues.value)
		return err
	case "ADBE Text Tracking Amount":
		value, ok := textAnimatorInitialFloatValue(property)
		if !ok {
			return fmt.Errorf("value unsupported (%T)", property.StaticValue)
		}
		_, err := aep.AddTextTrackingAnimator(targetLayer, value, rangeValues.start, rangeValues.end, rangeValues.value)
		return err
	case "ADBE Text Character Offset":
		value, ok := textAnimatorInitialFloatValue(property)
		if !ok {
			return fmt.Errorf("value unsupported (%T)", property.StaticValue)
		}
		_, err := aep.AddTextCharacterOffsetAnimator(targetLayer, value, rangeValues.start, rangeValues.end, rangeValues.value)
		return err
	case "ADBE Text Fill Opacity":
		value, ok := textAnimatorInitialFloatValue(property)
		if !ok {
			return fmt.Errorf("value unsupported (%T)", property.StaticValue)
		}
		_, err := aep.AddTextFillOpacityAnimator(targetLayer, value, rangeValues.start, rangeValues.end, rangeValues.value)
		return err
	case "ADBE Text Stroke Opacity":
		value, ok := textAnimatorInitialFloatValue(property)
		if !ok {
			return fmt.Errorf("value unsupported (%T)", property.StaticValue)
		}
		_, err := aep.AddTextStrokeOpacityAnimator(targetLayer, value, rangeValues.start, rangeValues.end, rangeValues.value)
		return err
	case "ADBE Text Stroke Width":
		value, ok := textAnimatorInitialFloatValue(property)
		if !ok {
			return fmt.Errorf("value unsupported (%T)", property.StaticValue)
		}
		_, err := aep.AddTextStrokeWidthAnimator(targetLayer, value, rangeValues.start, rangeValues.end, rangeValues.value)
		return err
	case "ADBE Text Skew":
		value, ok := textAnimatorInitialFloatValue(property)
		if !ok {
			return fmt.Errorf("value unsupported (%T)", property.StaticValue)
		}
		_, err := aep.AddTextSkewAnimator(targetLayer, value, rangeValues.start, rangeValues.end, rangeValues.value)
		return err
	case "ADBE Text Rotation X":
		value, ok := textAnimatorInitialFloatValue(property)
		if !ok {
			return fmt.Errorf("value unsupported (%T)", property.StaticValue)
		}
		_, err := aep.AddTextRotationXAnimator(targetLayer, value, rangeValues.start, rangeValues.end, rangeValues.value)
		return err
	case "ADBE Text Rotation Y":
		value, ok := textAnimatorInitialFloatValue(property)
		if !ok {
			return fmt.Errorf("value unsupported (%T)", property.StaticValue)
		}
		_, err := aep.AddTextRotationYAnimator(targetLayer, value, rangeValues.start, rangeValues.end, rangeValues.value)
		return err
	default:
		return fmt.Errorf("unsupported text animator property")
	}
}

func materializeTextRangeOffsetKeyframes(targetLayer *aep.Layer, property profile.Property) error {
	keyframes, err := textAnimatorScalarKeyframes(property)
	if err != nil {
		return err
	}
	return aep.AnimateTextRangeOffset(targetLayer, 0, keyframes)
}

func materializeTextAnimatorValueKeyframes(targetLayer *aep.Layer, property profile.Property) error {
	switch property.MatchName {
	case "ADBE Text Opacity":
		keyframes, err := textAnimatorScalarKeyframes(property)
		if err != nil {
			return err
		}
		return aep.AnimateTextOpacity(targetLayer, 0, keyframes)
	case "ADBE Text Position 3D":
		keyframes, err := textAnimatorVectorKeyframes(property, 3)
		if err != nil {
			return err
		}
		return aep.AnimateTextPosition(targetLayer, 0, keyframes)
	case "ADBE Text Scale 3D":
		keyframes, err := textAnimatorVectorKeyframes(property, 3)
		if err != nil {
			return err
		}
		return aep.AnimateTextScale(targetLayer, 0, keyframes)
	case "ADBE Text Rotation":
		keyframes, err := textAnimatorScalarKeyframes(property)
		if err != nil {
			return err
		}
		return aep.AnimateTextRotation(targetLayer, 0, keyframes)
	case "ADBE Text Fill Color":
		keyframes, err := textAnimatorColorKeyframes(property)
		if err != nil {
			return err
		}
		return aep.AnimateTextColor(targetLayer, 0, keyframes)
	case "ADBE Text Tracking Amount":
		keyframes, err := textAnimatorScalarKeyframes(property)
		if err != nil {
			return err
		}
		return aep.AnimateTextTracking(targetLayer, 0, keyframes)
	case "ADBE Text Character Offset":
		keyframes, err := textAnimatorScalarKeyframes(property)
		if err != nil {
			return err
		}
		return aep.AnimateTextCharacterOffset(targetLayer, 0, keyframes)
	default:
		return fmt.Errorf("value keyframes are not supported")
	}
}

func textAnimatorInitialFloatValue(property profile.Property) (float64, bool) {
	if len(property.Keyframes) != 0 {
		return propertyFloatValue(property.Keyframes[0].Value)
	}
	return propertyFloatValue(property.StaticValue)
}

func textAnimatorInitialVectorValue(property profile.Property, length int) ([]float64, bool) {
	if len(property.Keyframes) != 0 {
		return staticVectorAtLeast(property.Keyframes[0].Value, length)
	}
	return staticVectorAtLeast(property.StaticValue, length)
}

func textAnimatorInitialColorValue(property profile.Property) ([4]float64, bool) {
	var value any = property.StaticValue
	if len(property.Keyframes) != 0 {
		value = property.Keyframes[0].Value
	}
	vector, ok := staticVectorAtLeast(value, 3)
	if !ok {
		return [4]float64{}, false
	}
	return profileTextColorToRGBA(vector), true
}

func textAnimatorScalarKeyframes(property profile.Property) ([]aep.ScalarKeyframe, error) {
	if len(property.Keyframes) < 2 {
		return nil, fmt.Errorf("need >= 2 keyframes, got %d", len(property.Keyframes))
	}
	out := make([]aep.ScalarKeyframe, 0, len(property.Keyframes))
	for i, keyframe := range property.Keyframes {
		value, ok := propertyFloatValue(keyframe.Value)
		if !ok {
			return nil, fmt.Errorf("keyframes[%d].value must be a number", i)
		}
		out = append(out, aep.ScalarKeyframe{
			Time:    keyframe.Time,
			Value:   value,
			InEase:  effectParamTemporalEase(keyframe.InTemporalEase),
			OutEase: effectParamTemporalEase(keyframe.OutTemporalEase),
		})
	}
	return out, nil
}

func textAnimatorVectorKeyframes(property profile.Property, length int) ([]aep.VectorKeyframe, error) {
	if len(property.Keyframes) < 2 {
		return nil, fmt.Errorf("need >= 2 keyframes, got %d", len(property.Keyframes))
	}
	out := make([]aep.VectorKeyframe, 0, len(property.Keyframes))
	for i, keyframe := range property.Keyframes {
		value, ok := staticVectorAtLeast(keyframe.Value, length)
		if !ok {
			return nil, fmt.Errorf("keyframes[%d].value must be a %d-number array", i, length)
		}
		out = append(out, aep.VectorKeyframe{
			Time:    keyframe.Time,
			Value:   append([]float64(nil), value[:length]...),
			InEase:  effectParamTemporalEase(keyframe.InTemporalEase),
			OutEase: effectParamTemporalEase(keyframe.OutTemporalEase),
		})
	}
	return out, nil
}

func textAnimatorColorKeyframes(property profile.Property) ([]aep.VectorKeyframe, error) {
	if len(property.Keyframes) < 2 {
		return nil, fmt.Errorf("need >= 2 keyframes, got %d", len(property.Keyframes))
	}
	out := make([]aep.VectorKeyframe, 0, len(property.Keyframes))
	for i, keyframe := range property.Keyframes {
		value, ok := staticVectorAtLeast(keyframe.Value, 3)
		if !ok {
			return nil, fmt.Errorf("keyframes[%d].value must be a color array", i)
		}
		color := profileTextColorToRGBA(value)
		out = append(out, aep.VectorKeyframe{
			Time:    keyframe.Time,
			Value:   []float64{color[0], color[1], color[2], color[3]},
			InEase:  effectParamTemporalEase(keyframe.InTemporalEase),
			OutEase: effectParamTemporalEase(keyframe.OutTemporalEase),
		})
	}
	return out, nil
}

func profileTextColorToRGBA(value []float64) [4]float64 {
	if len(value) >= 4 {
		return profileARGBToRGBA(value[:4])
	}
	return [4]float64{
		colorByteToUnit(value[0]),
		colorByteToUnit(value[1]),
		colorByteToUnit(value[2]),
		1,
	}
}
