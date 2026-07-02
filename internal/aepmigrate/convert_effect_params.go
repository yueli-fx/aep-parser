package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

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
