package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

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
