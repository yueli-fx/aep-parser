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

func materializeEffectParamKeyframes(layer *aep.Layer, effect *aep.Effect, param profile.Property) (*aep.Property, error) {
	if _, ok := effectParamNumber(param.Keyframes[0].Value); ok {
		keyframes, err := effectParamScalarKeyframes(param.Keyframes)
		if err != nil {
			return nil, err
		}
		return aep.AnimateEffectParam(layer, effect, param.MatchName, keyframes)
	}
	keyframes, err := effectParamVectorKeyframes(param.Keyframes)
	if err != nil {
		return nil, err
	}
	return aep.AnimateEffectParamVec(layer, effect, param.MatchName, keyframes)
}

func hasProfileEffects(prof *profile.Profile) bool {
	if prof == nil {
		return false
	}
	for _, comp := range prof.Comps {
		for _, layer := range comp.Layers {
			if len(layer.Effects) != 0 {
				return true
			}
		}
	}
	return false
}

func isSupportedStaticEffectSurface(layer profile.Layer) bool {
	for _, effect := range layer.Effects {
		if !isSupportedEffectMatchName(effect.MatchName) {
			return false
		}
		for _, param := range effect.Params {
			if param.LayerRef != nil {
				if len(param.Keyframes) != 0 || effectParamHasExpression(param) {
					return false
				}
				continue
			}
			if len(param.Keyframes) != 0 {
				if !isSupportedEffectParamKeyframes(param.Keyframes) {
					return false
				}
				continue
			}
			if param.StaticValue != nil {
				_, ok := effectParamStaticValue(param.StaticValue)
				if !ok {
					return false
				}
				continue
			}
			if effectParamHasExpression(param) || param.Changed {
				return false
			}
		}
	}
	return true
}

func effectParamHasExpression(param profile.Property) bool {
	return param.Expression != "" || param.ExpressionEnabled != nil
}

func isSupportedEffectParamKeyframes(keyframes []profile.Keyframe) bool {
	if len(keyframes) < 2 {
		return false
	}
	if _, ok := effectParamNumber(keyframes[0].Value); ok {
		_, err := effectParamScalarKeyframes(keyframes)
		return err == nil
	}
	_, err := effectParamVectorKeyframes(keyframes)
	return err == nil
}

func isSupportedEffectMatchName(matchName string) bool {
	for _, supported := range aep.SupportedEffects() {
		if supported == matchName {
			return true
		}
	}
	return false
}

func effectParamStaticValue(value any) (any, bool) {
	switch v := value.(type) {
	case nil:
		return nil, false
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case bool:
		if v {
			return 1.0, true
		}
		return 0.0, true
	case []float64:
		return append([]float64(nil), v...), true
	case []any:
		out := make([]float64, 0, len(v))
		for _, item := range v {
			n, ok := effectParamNumber(item)
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

func effectParamNumber(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

func effectParamScalarKeyframes(in []profile.Keyframe) ([]aep.ScalarKeyframe, error) {
	out := make([]aep.ScalarKeyframe, 0, len(in))
	for i, kf := range in {
		value, ok := effectParamNumber(kf.Value)
		if !ok {
			return nil, fmt.Errorf("keyframes[%d].value must be a number", i)
		}
		out = append(out, aep.ScalarKeyframe{
			Time:    kf.Time,
			Value:   value,
			InEase:  effectParamTemporalEase(kf.InTemporalEase),
			OutEase: effectParamTemporalEase(kf.OutTemporalEase),
		})
	}
	return out, nil
}

func effectParamVectorKeyframes(in []profile.Keyframe) ([]aep.VectorKeyframe, error) {
	out := make([]aep.VectorKeyframe, 0, len(in))
	for i, kf := range in {
		value, ok := effectParamVectorValue(kf.Value)
		if !ok || len(value) < 2 || len(value) > 4 {
			return nil, fmt.Errorf("keyframes[%d].value must be a 2-, 3-, or 4-number array", i)
		}
		out = append(out, aep.VectorKeyframe{
			Time:    kf.Time,
			Value:   value,
			InEase:  effectParamTemporalEase(kf.InTemporalEase),
			OutEase: effectParamTemporalEase(kf.OutTemporalEase),
		})
	}
	return out, nil
}

func effectParamVectorValue(value any) ([]float64, bool) {
	switch v := value.(type) {
	case []float64:
		return append([]float64(nil), v...), true
	case []any:
		out := make([]float64, 0, len(v))
		for _, item := range v {
			n, ok := effectParamNumber(item)
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

func effectParamTemporalEase(in []profile.TemporalEase) aep.TemporalEase {
	if len(in) == 0 {
		return aep.TemporalEase{}
	}
	return aep.TemporalEase{Speed: in[0].Speed, Influence: in[0].Influence}
}

func effectParamLayerRefIsSelf(ref *profile.LayerRef, layer profile.Layer) bool {
	if ref == nil {
		return false
	}
	if ref.Name != "" && ref.Name == layer.Name {
		return true
	}
	if ref.ID != 0 && ref.ID == layer.ID {
		return true
	}
	if ref.Index != 0 && ref.Index == layer.Index {
		return true
	}
	return false
}
