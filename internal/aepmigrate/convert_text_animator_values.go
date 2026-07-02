package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func textAnimatorInitialFloatValue(property profile.Property) (float64, bool) {
	if len(property.Keyframes) != 0 {
		return propertyFloatValue(property.Keyframes[0].Value)
	}
	return propertyFloatValue(property.StaticValue)
}

func textAnimatorInitialVectorValue(property profile.Property, length int) ([]float64, bool) {
	if len(property.Keyframes) != 0 {
		return staticVectorAtLeast(property.Keyframes[0].Value, length)
	}
	return staticVectorAtLeast(property.StaticValue, length)
}

func textAnimatorInitialColorValue(property profile.Property) ([4]float64, bool) {
	var value any = property.StaticValue
	if len(property.Keyframes) != 0 {
		value = property.Keyframes[0].Value
	}
	vector, ok := staticVectorAtLeast(value, 3)
	if !ok {
		return [4]float64{}, false
	}
	return profileTextColorToRGBA(vector), true
}

func textAnimatorScalarKeyframes(property profile.Property) ([]aep.ScalarKeyframe, error) {
	if len(property.Keyframes) < 2 {
		return nil, fmt.Errorf("need >= 2 keyframes, got %d", len(property.Keyframes))
	}
	out := make([]aep.ScalarKeyframe, 0, len(property.Keyframes))
	for i, keyframe := range property.Keyframes {
		value, ok := propertyFloatValue(keyframe.Value)
		if !ok {
			return nil, fmt.Errorf("keyframes[%d].value must be a number", i)
		}
		out = append(out, aep.ScalarKeyframe{
			Time:    keyframe.Time,
			Value:   value,
			InEase:  effectParamTemporalEase(keyframe.InTemporalEase),
			OutEase: effectParamTemporalEase(keyframe.OutTemporalEase),
		})
	}
	return out, nil
}

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

func textAnimatorColorKeyframes(property profile.Property) ([]aep.VectorKeyframe, error) {
	if len(property.Keyframes) < 2 {
		return nil, fmt.Errorf("need >= 2 keyframes, got %d", len(property.Keyframes))
	}
	out := make([]aep.VectorKeyframe, 0, len(property.Keyframes))
	for i, keyframe := range property.Keyframes {
		value, ok := staticVectorAtLeast(keyframe.Value, 3)
		if !ok {
			return nil, fmt.Errorf("keyframes[%d].value must be a color array", i)
		}
		color := profileTextColorToRGBA(value)
		out = append(out, aep.VectorKeyframe{
			Time:    keyframe.Time,
			Value:   []float64{color[0], color[1], color[2], color[3]},
			InEase:  effectParamTemporalEase(keyframe.InTemporalEase),
			OutEase: effectParamTemporalEase(keyframe.OutTemporalEase),
		})
	}
	return out, nil
}

func profileTextColorToRGBA(value []float64) [4]float64 {
	if len(value) >= 4 {
		return profileARGBToRGBA(value[:4])
	}
	return [4]float64{
		colorByteToUnit(value[0]),
		colorByteToUnit(value[1]),
		colorByteToUnit(value[2]),
		1,
	}
}
