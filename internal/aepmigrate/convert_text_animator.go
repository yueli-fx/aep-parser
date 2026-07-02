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
