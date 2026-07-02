package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeDefaultTransformSurface(layer *aep.Layer, source profile.Layer) error {
	if !hasTransformProperties(source) {
		return nil
	}
	transform, err := layerTransformFromStaticProfile(source)
	if err != nil {
		return err
	}
	return aep.SetLayerTransform(layer, transform)
}

func materializeCenteredTransformSurface(comp *aep.Composition, layer *aep.Layer, source profile.Layer) error {
	if !hasTransformProperties(source) {
		return nil
	}
	if layer == nil {
		return fmt.Errorf("layer not found after creation")
	}
	if _, ok := propertyVectorAtLeast(source.Properties, "ADBE Position", 2); ok {
		return materializeDefaultTransformSurface(layer, source)
	}
	transform := aep.NewLayerTransform()
	if err := transform.Position().SetStaticValue([2]float64{float64(comp.Width) / 2, float64(comp.Height) / 2}); err != nil {
		return err
	}
	return aep.SetLayerTransform(layer, transform)
}

func materializeCameraLightTransformSurface(layer *aep.Layer, source profile.Layer) error {
	if hasProperty(source, "ADBE Position_2") {
		return nil
	}
	return materializeDefaultTransformSurface(layer, source)
}

func materializePrecompTransformSurface(comp *aep.Composition, layer *aep.Layer, source profile.Layer) error {
	if hasStaticPropertyVector(source, "ADBE Position", []float64{0, 0, 0}) {
		return materializeDefaultTransformSurface(layer, source)
	}
	return materializeCenteredTransformSurface(comp, layer, source)
}
