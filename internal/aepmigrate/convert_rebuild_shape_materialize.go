package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

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
