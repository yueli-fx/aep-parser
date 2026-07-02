package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeEffectParams(targetComp *aep.Composition, targetLayer *aep.Layer, sourceLayer profile.Layer, targetEffect *aep.Effect, sourceEffect profile.Effect) error {
	for _, param := range sourceEffect.Params {
		if param.LayerRef != nil && !effectParamLayerRefIsSelf(param.LayerRef, sourceLayer) {
			if err := materializeEffectLayerRefParam(targetComp, targetLayer, targetEffect, param); err != nil {
				return err
			}
			continue
		}
		property, err := materializeEffectValueParam(targetLayer, targetEffect, param)
		if err != nil {
			return err
		}
		if err := materializeEffectParamExpression(property, param); err != nil {
			return err
		}
	}
	return nil
}

func materializeEffectLayerRefParam(targetComp *aep.Composition, targetLayer *aep.Layer, targetEffect *aep.Effect, param profile.Property) error {
	target := targetLayerBySourceRef(targetComp, param.LayerRef)
	if target == nil {
		return fmt.Errorf("param %q target layer %q not found", param.MatchName, param.LayerRef.Name)
	}
	if err := aep.SetEffectLayerParam(targetLayer, targetEffect, param.MatchName, target); err != nil {
		return fmt.Errorf("param %q layer_ref: %w", param.MatchName, err)
	}
	return nil
}

func materializeEffectValueParam(targetLayer *aep.Layer, targetEffect *aep.Effect, param profile.Property) (*aep.Property, error) {
	if len(param.Keyframes) != 0 {
		property, err := materializeEffectParamKeyframes(targetLayer, targetEffect, param)
		if err != nil {
			return nil, fmt.Errorf("param %q keyframes: %w", param.MatchName, err)
		}
		return property, nil
	}
	value, ok := effectParamStaticValue(param.StaticValue)
	if !ok {
		if param.Changed {
			return nil, fmt.Errorf("param %q static value unsupported (%T)", param.MatchName, param.StaticValue)
		}
		return nil, nil
	}
	property, err := aep.SetEffectParam(targetLayer, targetEffect, param.MatchName, value)
	if err != nil {
		return nil, fmt.Errorf("param %q static value: %w", param.MatchName, err)
	}
	return property, nil
}

func materializeEffectParamExpression(property *aep.Property, param profile.Property) error {
	if !effectParamHasExpression(param) {
		return nil
	}
	if property == nil {
		return fmt.Errorf("param %q expression requires a materialized value/keyframes param", param.MatchName)
	}
	if param.Expression != "" {
		if err := property.SetExpression(param.Expression); err != nil {
			return fmt.Errorf("param %q expression source: %w", param.MatchName, err)
		}
	}
	if param.ExpressionEnabled != nil {
		if err := property.SetExpressionEnabled(*param.ExpressionEnabled); err != nil {
			return fmt.Errorf("param %q expression enabled: %w", param.MatchName, err)
		}
	}
	return nil
}
