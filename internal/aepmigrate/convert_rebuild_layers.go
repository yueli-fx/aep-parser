package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func rebuildCompLayers(target VersionLabel, comp profile.Composition, targetComp *aep.Composition, footage convertFootageIndex, comps convertCompIndex, targetComps convertTargetCompIndex) error {
	var createdLayers []convertLayerPair
	recordLayer := func(source profile.Layer, target *aep.Layer) {
		createdLayers = append(createdLayers, convertLayerPair{source: source, target: target})
	}
	for _, layer := range comp.Layers {
		var dstLayer *aep.Layer
		var err error
		switch {
		case isSupportedDefaultNullLayer(layer, footage):
			dstLayer, err = rebuildDefaultNullLayer(comp.Name, targetComp, layer, footage)
		case isSupportedDefaultSolidLayer(layer, footage):
			dstLayer, err = rebuildDefaultSolidLayer(comp.Name, targetComp, layer, footage)
		case isSupportedDefaultAdjustmentLayer(layer, footage):
			dstLayer, err = rebuildDefaultAdjustmentLayer(comp.Name, targetComp, layer)
		case isSupportedDefaultCameraLayer(layer):
			dstLayer, err = rebuildDefaultCameraLayer(comp.Name, targetComp, layer)
		case isSupportedDefaultLightLayer(layer):
			dstLayer, err = rebuildDefaultLightLayer(comp.Name, targetComp, layer)
		case isSupportedDefaultTextLayer(layer):
			dstLayer, err = rebuildDefaultTextLayer(comp.Name, targetComp, layer)
		case isSupportedDefaultShapeLayer(layer):
			dstLayer, err = rebuildDefaultShapeLayer(comp.Name, targetComp, layer)
		case isSupportedRectGraphicShapeLayer(layer):
			dstLayer, err = rebuildRectGraphicShapeLayer(comp.Name, targetComp, layer)
		case isSupportedDefaultPrecompLayer(layer, comps):
			dstLayer, err = rebuildDefaultPrecompLayer(comp.Name, targetComp, layer, targetComps)
		default:
			return fmt.Errorf("unsupported layer %q in comp %q", layer.Name, comp.Name)
		}
		if err != nil {
			return err
		}
		recordLayer(layer, dstLayer)
	}
	if err := materializeLayerRefs(target, createdLayers); err != nil {
		return fmt.Errorf("comp %q layer refs: %w", comp.Name, err)
	}
	return nil
}
