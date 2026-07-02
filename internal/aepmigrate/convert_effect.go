package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeProjectEffects(project *aep.Project, prof *profile.Profile) (*aep.Project, error) {
	if !hasProfileEffects(prof) {
		return project, nil
	}
	reopened, err := aep.Reopen(project)
	if err != nil {
		return nil, fmt.Errorf("reopen for effects: %w", err)
	}
	for compIndex, sourceComp := range prof.Comps {
		targetComp, err := targetCompBySource(reopened, compIndex, sourceComp)
		if err != nil {
			return nil, err
		}
		for layerIndex, sourceLayer := range sourceComp.Layers {
			if len(sourceLayer.Effects) == 0 {
				continue
			}
			if layerIndex >= len(targetComp.Layers) {
				return nil, fmt.Errorf("comp %q effects: target layer index %d missing", sourceComp.Name, layerIndex)
			}
			targetLayer := targetComp.Layers[layerIndex]
			for _, sourceEffect := range sourceLayer.Effects {
				targetEffect, err := aep.AddEffect(targetLayer, sourceEffect.MatchName)
				if err != nil {
					return nil, fmt.Errorf("comp %q layer %q add effect %q: %w", sourceComp.Name, sourceLayer.Name, sourceEffect.MatchName, err)
				}
				if err := materializeEffectParams(targetComp, targetLayer, sourceLayer, targetEffect, sourceEffect); err != nil {
					return nil, fmt.Errorf("comp %q layer %q effect %q params: %w", sourceComp.Name, sourceLayer.Name, sourceEffect.MatchName, err)
				}
			}
		}
	}
	return reopened, nil
}

func materializeEffectParams(targetComp *aep.Composition, targetLayer *aep.Layer, sourceLayer profile.Layer, targetEffect *aep.Effect, sourceEffect profile.Effect) error {
	for _, param := range sourceEffect.Params {
		var property *aep.Property
		if param.LayerRef != nil && !effectParamLayerRefIsSelf(param.LayerRef, sourceLayer) {
			target := targetLayerBySourceRef(targetComp, param.LayerRef)
			if target == nil {
				return fmt.Errorf("param %q target layer %q not found", param.MatchName, param.LayerRef.Name)
			}
			if err := aep.SetEffectLayerParam(targetLayer, targetEffect, param.MatchName, target); err != nil {
				return fmt.Errorf("param %q layer_ref: %w", param.MatchName, err)
			}
			continue
		}
		if len(param.Keyframes) != 0 {
			var err error
			property, err = materializeEffectParamKeyframes(targetLayer, targetEffect, param)
			if err != nil {
				return fmt.Errorf("param %q keyframes: %w", param.MatchName, err)
			}
		} else {
			value, ok := effectParamStaticValue(param.StaticValue)
			if !ok {
				if param.Changed {
					return fmt.Errorf("param %q static value unsupported (%T)", param.MatchName, param.StaticValue)
				}
			} else {
				var err error
				property, err = aep.SetEffectParam(targetLayer, targetEffect, param.MatchName, value)
				if err != nil {
					return fmt.Errorf("param %q static value: %w", param.MatchName, err)
				}
			}
		}
		if effectParamHasExpression(param) {
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
		}
	}
	return nil
}
