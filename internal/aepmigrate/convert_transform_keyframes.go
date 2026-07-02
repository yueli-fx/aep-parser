package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/codec"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func addTransformVectorKeyframes(stream *codec.PropertyStream[[2]float64], property profile.Property, convert func(any) ([2]float64, bool)) error {
	if len(property.Keyframes) < 2 {
		return fmt.Errorf("need >= 2 keyframes, got %d", len(property.Keyframes))
	}
	for i, keyframe := range property.Keyframes {
		value, ok := convert(keyframe.Value)
		if !ok {
			return fmt.Errorf("keyframes[%d].value unsupported (%T)", i, keyframe.Value)
		}
		inEase := transformTemporalEase(keyframe.InTemporalEase)
		outEase := transformTemporalEase(keyframe.OutTemporalEase)
		if inEase != (aep.TemporalEase{}) || outEase != (aep.TemporalEase{}) {
			if err := stream.AddKeyframeWithEase(keyframe.Time, value, inEase, outEase); err != nil {
				return err
			}
			continue
		}
		if err := stream.AddKeyframeLinear(keyframe.Time, value); err != nil {
			return err
		}
	}
	return nil
}

func addTransformScalarKeyframes(stream *codec.PropertyStream[float64], property profile.Property, convert func(any) (float64, bool)) error {
	if len(property.Keyframes) < 2 {
		return fmt.Errorf("need >= 2 keyframes, got %d", len(property.Keyframes))
	}
	for i, keyframe := range property.Keyframes {
		value, ok := convert(keyframe.Value)
		if !ok {
			return fmt.Errorf("keyframes[%d].value unsupported (%T)", i, keyframe.Value)
		}
		inEase := transformTemporalEase(keyframe.InTemporalEase)
		outEase := transformTemporalEase(keyframe.OutTemporalEase)
		if inEase != (aep.TemporalEase{}) || outEase != (aep.TemporalEase{}) {
			if err := stream.AddKeyframeWithEase(keyframe.Time, value, inEase, outEase); err != nil {
				return err
			}
			continue
		}
		if err := stream.AddKeyframeLinear(keyframe.Time, value); err != nil {
			return err
		}
	}
	return nil
}

func profileVector2(value any) ([2]float64, bool) {
	vector, ok := staticVectorAtLeast(value, 2)
	if !ok {
		return [2]float64{}, false
	}
	return [2]float64{vector[0], vector[1]}, true
}

func profileScaleVector2(value any) ([2]float64, bool) {
	vector, ok := staticVectorAtLeast(value, 2)
	if !ok {
		return [2]float64{}, false
	}
	return [2]float64{profileScaleToWriter(vector[0]), profileScaleToWriter(vector[1])}, true
}

func profileScalar(value any) (float64, bool) {
	return propertyFloatValue(value)
}

func profileOpacityScalar(value any) (float64, bool) {
	valueFloat, ok := propertyFloatValue(value)
	if !ok {
		return 0, false
	}
	return profileOpacityToWriter(valueFloat), true
}

func transformTemporalEase(in []profile.TemporalEase) aep.TemporalEase {
	if len(in) == 0 {
		return aep.TemporalEase{}
	}
	return aep.TemporalEase{Speed: in[0].Speed, Influence: in[0].Influence}
}
