package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

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
