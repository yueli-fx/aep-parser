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

func rebuildDefaultShapeLayer(compName string, targetComp *aep.Composition, layer profile.Layer) (*aep.Layer, error) {
	dstLayer, err := aep.NewShapeLayer(targetComp, layer.Name)
	if err != nil {
		return nil, fmt.Errorf("comp %q shape layer %q: %w", compName, layer.Name, err)
	}
	if err := materializeLayerTiming(dstLayer.Layer, layer); err != nil {
		return nil, fmt.Errorf("comp %q shape layer %q timing: %w", compName, layer.Name, err)
	}
	return dstLayer.Layer, nil
}

func rebuildRectGraphicShapeLayer(compName string, targetComp *aep.Composition, layer profile.Layer) (*aep.Layer, error) {
	dstLayer, err := materializeRectGraphicShapeLayer(targetComp, layer)
	if err != nil {
		return nil, fmt.Errorf("comp %q shape layer %q: %w", compName, layer.Name, err)
	}
	if err := materializeLayerTiming(dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q shape layer %q timing: %w", compName, layer.Name, err)
	}
	return dstLayer, nil
}

func rebuildDefaultPrecompLayer(compName string, targetComp *aep.Composition, layer profile.Layer, targetComps convertTargetCompIndex) (*aep.Layer, error) {
	sourceComp, ok := targetComps.sourceComposition(layer)
	if !ok {
		return nil, fmt.Errorf("comp %q precomp layer %q source %q not found", compName, layer.Name, layer.SourceRef.Name)
	}
	dstLayer, err := aep.NewPrecompLayer(targetComp, sourceComp, layer.Name)
	if err != nil {
		return nil, fmt.Errorf("comp %q precomp layer %q: %w", compName, layer.Name, err)
	}
	if err := materializeLayerMetadata(dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q precomp layer %q metadata: %w", compName, layer.Name, err)
	}
	if err := materializeLayerSwitchSurface(dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q precomp layer %q switches: %w", compName, layer.Name, err)
	}
	if err := materializePrecompTransformSurface(targetComp, dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q precomp layer %q transform: %w", compName, layer.Name, err)
	}
	if err := materializeLayerTiming(dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q precomp layer %q timing: %w", compName, layer.Name, err)
	}
	return dstLayer, nil
}
