package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

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
