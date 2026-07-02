package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/codec"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func layerTransformFromStaticProfile(layer profile.Layer) (*aep.LayerTransform, error) {
	transform := aep.NewLayerTransform()
	if property, ok := propertyByMatchName(layer.Properties, "ADBE Anchor Point"); ok {
		if len(property.Keyframes) != 0 {
			if err := addTransformVectorKeyframes(transform.AnchorPoint(), property, profileVector2); err != nil {
				return nil, fmt.Errorf("anchor point keyframes: %w", err)
			}
		} else if value, ok := staticVectorAtLeast(property.StaticValue, 2); ok {
			if err := transform.AnchorPoint().SetStaticValue([2]float64{value[0], value[1]}); err != nil {
				return nil, err
			}
		}
	}
	if property, ok := propertyByMatchName(layer.Properties, "ADBE Position"); ok {
		if len(property.Keyframes) != 0 {
			if err := addTransformVectorKeyframes(transform.Position(), property, profileVector2); err != nil {
				return nil, fmt.Errorf("position keyframes: %w", err)
			}
		} else if value, ok := staticVectorAtLeast(property.StaticValue, 2); ok {
			if err := transform.Position().SetStaticValue([2]float64{value[0], value[1]}); err != nil {
				return nil, err
			}
		}
	}
	if property, ok := propertyByMatchName(layer.Properties, "ADBE Scale"); ok {
		if len(property.Keyframes) != 0 {
			if err := addTransformVectorKeyframes(transform.Scale(), property, profileScaleVector2); err != nil {
				return nil, fmt.Errorf("scale keyframes: %w", err)
			}
		} else if value, ok := staticVectorAtLeast(property.StaticValue, 2); ok {
			if err := transform.Scale().SetStaticValue([2]float64{profileScaleToWriter(value[0]), profileScaleToWriter(value[1])}); err != nil {
				return nil, err
			}
		}
	}
	if property, ok := propertyByMatchName(layer.Properties, "ADBE Rotate Z"); ok {
		if len(property.Keyframes) != 0 {
			if err := addTransformScalarKeyframes(transform.Rotation(), property, profileScalar); err != nil {
				return nil, fmt.Errorf("rotation keyframes: %w", err)
			}
		} else if value, ok := propertyFloatValue(property.StaticValue); ok {
			if err := transform.Rotation().SetStaticValue(value); err != nil {
				return nil, err
			}
		}
	}
	if property, ok := propertyByMatchName(layer.Properties, "ADBE Opacity"); ok {
		if len(property.Keyframes) != 0 {
			if err := addTransformScalarKeyframes(transform.Opacity(), property, profileOpacityScalar); err != nil {
				return nil, fmt.Errorf("opacity keyframes: %w", err)
			}
		} else if value, ok := propertyFloatValue(property.StaticValue); ok {
			if err := transform.Opacity().SetStaticValue(profileOpacityToWriter(value)); err != nil {
				return nil, err
			}
		}
	}
	return transform, nil
}

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
