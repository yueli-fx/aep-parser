package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func rebuildDefaultCameraLayer(compName string, targetComp *aep.Composition, layer profile.Layer) (*aep.Layer, error) {
	dstLayer, err := aep.NewCameraLayer(targetComp, layer.Name)
	if err != nil {
		return nil, fmt.Errorf("comp %q camera layer %q: %w", compName, layer.Name, err)
	}
	if err := materializeLayerMetadata(dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q camera layer %q metadata: %w", compName, layer.Name, err)
	}
	if err := materializeLayerSwitchSurface(dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q camera layer %q switches: %w", compName, layer.Name, err)
	}
	if err := materializeCameraLightTransformSurface(dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q camera layer %q transform: %w", compName, layer.Name, err)
	}
	if err := materializeCameraOptions(dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q camera layer %q options: %w", compName, layer.Name, err)
	}
	if err := materializeLayerTiming(dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q camera layer %q timing: %w", compName, layer.Name, err)
	}
	return dstLayer, nil
}

func rebuildDefaultLightLayer(compName string, targetComp *aep.Composition, layer profile.Layer) (*aep.Layer, error) {
	dstLayer, err := aep.NewLightLayer(targetComp, layer.Name)
	if err != nil {
		return nil, fmt.Errorf("comp %q light layer %q: %w", compName, layer.Name, err)
	}
	if err := materializeLayerMetadata(dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q light layer %q metadata: %w", compName, layer.Name, err)
	}
	if err := materializeLayerSwitchSurface(dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q light layer %q switches: %w", compName, layer.Name, err)
	}
	if err := materializeCameraLightTransformSurface(dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q light layer %q transform: %w", compName, layer.Name, err)
	}
	if err := materializeLightOptions(dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q light layer %q options: %w", compName, layer.Name, err)
	}
	if err := materializeLayerTiming(dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q light layer %q timing: %w", compName, layer.Name, err)
	}
	return dstLayer, nil
}
