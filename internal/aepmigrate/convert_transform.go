package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func layerTransformFromStaticProfile(layer profile.Layer) (*aep.LayerTransform, error) {
	transform := aep.NewLayerTransform()
	if err := populateLayerTransformFromProfile(transform, layer); err != nil {
		return nil, err
	}
	return transform, nil
}

func populateLayerTransformFromProfile(transform *aep.LayerTransform, layer profile.Layer) error {
	if property, ok := propertyByMatchName(layer.Properties, "ADBE Anchor Point"); ok {
		if len(property.Keyframes) != 0 {
			if err := addTransformVectorKeyframes(transform.AnchorPoint(), property, profileVector2); err != nil {
				return fmt.Errorf("anchor point keyframes: %w", err)
			}
		} else if value, ok := staticVectorAtLeast(property.StaticValue, 2); ok {
			if err := transform.AnchorPoint().SetStaticValue([2]float64{value[0], value[1]}); err != nil {
				return err
			}
		}
	}
	if property, ok := propertyByMatchName(layer.Properties, "ADBE Position"); ok {
		if len(property.Keyframes) != 0 {
			if err := addTransformVectorKeyframes(transform.Position(), property, profileVector2); err != nil {
				return fmt.Errorf("position keyframes: %w", err)
			}
		} else if value, ok := staticVectorAtLeast(property.StaticValue, 2); ok {
			if err := transform.Position().SetStaticValue([2]float64{value[0], value[1]}); err != nil {
				return err
			}
		}
	}
	if property, ok := propertyByMatchName(layer.Properties, "ADBE Scale"); ok {
		if len(property.Keyframes) != 0 {
			if err := addTransformVectorKeyframes(transform.Scale(), property, profileScaleVector2); err != nil {
				return fmt.Errorf("scale keyframes: %w", err)
			}
		} else if value, ok := staticVectorAtLeast(property.StaticValue, 2); ok {
			if err := transform.Scale().SetStaticValue([2]float64{profileScaleToWriter(value[0]), profileScaleToWriter(value[1])}); err != nil {
				return err
			}
		}
	}
	if property, ok := propertyByMatchName(layer.Properties, "ADBE Rotate Z"); ok {
		if len(property.Keyframes) != 0 {
			if err := addTransformScalarKeyframes(transform.Rotation(), property, profileScalar); err != nil {
				return fmt.Errorf("rotation keyframes: %w", err)
			}
		} else if value, ok := propertyFloatValue(property.StaticValue); ok {
			if err := transform.Rotation().SetStaticValue(value); err != nil {
				return err
			}
		}
	}
	if property, ok := propertyByMatchName(layer.Properties, "ADBE Opacity"); ok {
		if len(property.Keyframes) != 0 {
			if err := addTransformScalarKeyframes(transform.Opacity(), property, profileOpacityScalar); err != nil {
				return fmt.Errorf("opacity keyframes: %w", err)
			}
		} else if value, ok := propertyFloatValue(property.StaticValue); ok {
			if err := transform.Opacity().SetStaticValue(profileOpacityToWriter(value)); err != nil {
				return err
			}
		}
	}
	return nil
}
