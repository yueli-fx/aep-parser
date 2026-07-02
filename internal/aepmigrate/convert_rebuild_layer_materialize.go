package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func rebuildDefaultNullLayer(compName string, targetComp *aep.Composition, layer profile.Layer, footage convertFootageIndex) (*aep.Layer, error) {
	dstLayer, err := materializeNullLayer(targetComp, layer, footage)
	if err != nil {
		return nil, fmt.Errorf("comp %q null layer %q: %w", compName, layer.Name, err)
	}
	if err := materializeLayerMetadata(dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q null layer %q metadata: %w", compName, layer.Name, err)
	}
	if err := materializeLayerSwitchSurface(dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q null layer %q switches: %w", compName, layer.Name, err)
	}
	if err := materializeDefaultTransformSurface(dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q null layer %q transform: %w", compName, layer.Name, err)
	}
	if err := materializeLayerTiming(dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q null layer %q timing: %w", compName, layer.Name, err)
	}
	return dstLayer, nil
}

func rebuildDefaultSolidLayer(compName string, targetComp *aep.Composition, layer profile.Layer, footage convertFootageIndex) (*aep.Layer, error) {
	solid, _ := footage.solidDetails(layer)
	dstLayer, err := aep.NewSolidLayer(targetComp, layer.Name, int(solid.Width), int(solid.Height), *solid.SolidColor)
	if err != nil {
		return nil, fmt.Errorf("comp %q solid layer %q: %w", compName, layer.Name, err)
	}
	if err := materializeLayerMetadata(dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q solid layer %q metadata: %w", compName, layer.Name, err)
	}
	if err := materializeLayerSwitchSurface(dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q solid layer %q switches: %w", compName, layer.Name, err)
	}
	if err := materializeCenteredTransformSurface(targetComp, dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q solid layer %q transform: %w", compName, layer.Name, err)
	}
	if err := materializeLayerTiming(dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q solid layer %q timing: %w", compName, layer.Name, err)
	}
	return dstLayer, nil
}

func rebuildDefaultAdjustmentLayer(compName string, targetComp *aep.Composition, layer profile.Layer) (*aep.Layer, error) {
	dstLayer, err := aep.NewAdjustmentLayer(targetComp, layer.Name)
	if err != nil {
		return nil, fmt.Errorf("comp %q adjustment layer %q: %w", compName, layer.Name, err)
	}
	if err := materializeLayerMetadata(dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q adjustment layer %q metadata: %w", compName, layer.Name, err)
	}
	if err := materializeLayerSwitchSurface(dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q adjustment layer %q switches: %w", compName, layer.Name, err)
	}
	if err := materializeDefaultTransformSurface(dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q adjustment layer %q transform: %w", compName, layer.Name, err)
	}
	if err := materializeLayerTiming(dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q adjustment layer %q timing: %w", compName, layer.Name, err)
	}
	return dstLayer, nil
}
