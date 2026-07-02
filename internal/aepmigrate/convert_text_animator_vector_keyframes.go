package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func textAnimatorVectorKeyframes(property profile.Property, length int) ([]aep.VectorKeyframe, error) {
	if len(property.Keyframes) < 2 {
		return nil, fmt.Errorf("need >= 2 keyframes, got %d", len(property.Keyframes))
	}
	out := make([]aep.VectorKeyframe, 0, len(property.Keyframes))
	for i, keyframe := range property.Keyframes {
		value, ok := staticVectorAtLeast(keyframe.Value, length)
		if !ok {
			return nil, fmt.Errorf("keyframes[%d].value must be a %d-number array", i, length)
		}
		out = append(out, aep.VectorKeyframe{
			Time:    keyframe.Time,
			Value:   append([]float64(nil), value[:length]...),
			InEase:  effectParamTemporalEase(keyframe.InTemporalEase),
			OutEase: effectParamTemporalEase(keyframe.OutTemporalEase),
		})
	}
	return out, nil
}
