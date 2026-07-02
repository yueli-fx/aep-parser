package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

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
