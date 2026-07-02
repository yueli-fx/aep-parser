package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func rebuildDefaultTextLayer(compName string, targetComp *aep.Composition, layer profile.Layer) (*aep.Layer, error) {
	dstLayer, err := aep.NewTextLayer(targetComp, layer.Name)
	if err != nil {
		return nil, fmt.Errorf("comp %q text layer %q: %w", compName, layer.Name, err)
	}
	if err := materializeLayerMetadata(dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q text layer %q metadata: %w", compName, layer.Name, err)
	}
	if err := materializeLayerSwitchSurface(dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q text layer %q switches: %w", compName, layer.Name, err)
	}
	if layer.Text != nil {
		if err := dstLayer.SetText(layer.Text.Text); err != nil {
			return nil, fmt.Errorf("comp %q text layer %q text: %w", compName, layer.Name, err)
		}
		if err := materializeTextStyle(dstLayer, layer.Text); err != nil {
			return nil, fmt.Errorf("comp %q text layer %q text style: %w", compName, layer.Name, err)
		}
	}
	if err := materializeDefaultTransformSurface(dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q text layer %q transform: %w", compName, layer.Name, err)
	}
	if err := materializeLayerTiming(dstLayer, layer); err != nil {
		return nil, fmt.Errorf("comp %q text layer %q timing: %w", compName, layer.Name, err)
	}
	return dstLayer, nil
}
