package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeProjectTransformExpressions(project *aep.Project, prof *profile.Profile) (*aep.Project, error) {
	if !hasProfileTransformExpressions(prof) {
		return project, nil
	}
	reopened, err := aep.Reopen(project)
	if err != nil {
		return nil, fmt.Errorf("reopen for transform expressions: %w", err)
	}
	for compIndex, sourceComp := range prof.Comps {
		targetComp, err := targetCompBySource(reopened, compIndex, sourceComp)
		if err != nil {
			return nil, err
		}
		for layerIndex, sourceLayer := range sourceComp.Layers {
			if !hasLayerTransformExpressions(sourceLayer) {
				continue
			}
			if layerIndex >= len(targetComp.Layers) {
				return nil, fmt.Errorf("comp %q transform expressions: target layer index %d missing", sourceComp.Name, layerIndex)
			}
			if err := materializeLayerTransformExpressions(targetComp.Layers[layerIndex], sourceLayer); err != nil {
				return nil, fmt.Errorf("comp %q layer %q transform expressions: %w", sourceComp.Name, sourceLayer.Name, err)
			}
		}
	}
	return reopened, nil
}

func hasProfileTransformExpressions(prof *profile.Profile) bool {
	if prof == nil {
		return false
	}
	for _, comp := range prof.Comps {
		for _, layer := range comp.Layers {
			if hasLayerTransformExpressions(layer) {
				return true
			}
		}
	}
	return false
}

func hasLayerTransformExpressions(layer profile.Layer) bool {
	for _, property := range layer.Properties {
		if isTransformProperty(property.MatchName) && property.Expression != "" {
			return true
		}
		if isTransformProperty(property.MatchName) && property.ExpressionEnabled != nil {
			return true
		}
	}
	return false
}

func materializeLayerTransformExpressions(targetLayer *aep.Layer, sourceLayer profile.Layer) error {
	for _, property := range sourceLayer.Properties {
		if !isTransformProperty(property.MatchName) || !hasPropertyExpression(property) {
			continue
		}
		targetProperty := targetTransformProperty(targetLayer, property.MatchName)
		if targetProperty == nil {
			return fmt.Errorf("property %q not found", property.MatchName)
		}
		if property.Expression != "" {
			if err := targetProperty.SetExpression(property.Expression); err != nil {
				return fmt.Errorf("property %q expression source: %w", property.MatchName, err)
			}
		}
		if property.ExpressionEnabled != nil {
			if err := targetProperty.SetExpressionEnabled(*property.ExpressionEnabled); err != nil {
				return fmt.Errorf("property %q expression enabled: %w", property.MatchName, err)
			}
		}
	}
	return nil
}

func hasPropertyExpression(property profile.Property) bool {
	return property.Expression != "" || property.ExpressionEnabled != nil
}

func isTransformProperty(matchName string) bool {
	switch matchName {
	case "ADBE Anchor Point", "ADBE Position", "ADBE Scale", "ADBE Rotate Z", "ADBE Opacity":
		return true
	default:
		return false
	}
}

func targetTransformProperty(layer *aep.Layer, matchName string) *aep.Property {
	switch matchName {
	case "ADBE Anchor Point":
		return layer.AnchorPoint()
	case "ADBE Position":
		return layer.Position()
	case "ADBE Scale":
		return layer.Scale()
	case "ADBE Rotate Z":
		return layer.Rotation()
	case "ADBE Opacity":
		return layer.Opacity()
	default:
		return nil
	}
}
