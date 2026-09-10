package aepmigrate

import (
	"fmt"
	"github.com/yueli-fx/aep-parser/internal/serializer"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeEffectParamKeyframes(layer *aep.Layer, effect *aep.Effect, param profile.Property) (*aep.Property, error) {
	if _, ok := effectParamNumber(param.Keyframes[0].Value); ok {
		keyframes, err := effectParamScalarKeyframes(param.Keyframes)
		if err != nil {
			return nil, err
		}
		return serializer.AnimateEffectParam(layer, effect, param.MatchName, keyframes)
	}
	keyframes, err := effectParamVectorKeyframes(param.Keyframes)
	if err != nil {
		return nil, err
	}
	return serializer.AnimateEffectParamVec(layer, effect, param.MatchName, keyframes)
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

func effectParamTemporalEase(in []profile.TemporalEase) aep.TemporalEase {
	if len(in) == 0 {
		return aep.TemporalEase{}
	}
	return aep.TemporalEase{Speed: in[0].Speed, Influence: in[0].Influence}
}
