package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

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
