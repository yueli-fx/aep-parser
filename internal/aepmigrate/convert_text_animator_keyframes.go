package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

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
	case "ADBE Text Fill Opacity":
		keyframes, err := textAnimatorScalarKeyframes(property)
		if err != nil {
			return err
		}
		return aep.AnimateTextFillOpacity(targetLayer, 0, keyframes)
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
