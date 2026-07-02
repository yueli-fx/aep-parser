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
		switch {
		case isSupportedDefaultNullLayer(layer, footage):
			dstLayer, err := materializeNullLayer(targetComp, layer, footage)
			if err != nil {
				return fmt.Errorf("comp %q null layer %q: %w", comp.Name, layer.Name, err)
			}
			if err := materializeLayerMetadata(dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q null layer %q metadata: %w", comp.Name, layer.Name, err)
			}
			if err := materializeLayerSwitchSurface(dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q null layer %q switches: %w", comp.Name, layer.Name, err)
			}
			if err := materializeDefaultTransformSurface(dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q null layer %q transform: %w", comp.Name, layer.Name, err)
			}
			if err := materializeLayerTiming(dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q null layer %q timing: %w", comp.Name, layer.Name, err)
			}
			recordLayer(layer, dstLayer)
		case isSupportedDefaultSolidLayer(layer, footage):
			solid, _ := footage.solidDetails(layer)
			dstLayer, err := aep.NewSolidLayer(targetComp, layer.Name, int(solid.Width), int(solid.Height), *solid.SolidColor)
			if err != nil {
				return fmt.Errorf("comp %q solid layer %q: %w", comp.Name, layer.Name, err)
			}
			if err := materializeLayerMetadata(dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q solid layer %q metadata: %w", comp.Name, layer.Name, err)
			}
			if err := materializeLayerSwitchSurface(dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q solid layer %q switches: %w", comp.Name, layer.Name, err)
			}
			if err := materializeCenteredTransformSurface(targetComp, dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q solid layer %q transform: %w", comp.Name, layer.Name, err)
			}
			if err := materializeLayerTiming(dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q solid layer %q timing: %w", comp.Name, layer.Name, err)
			}
			recordLayer(layer, dstLayer)
		case isSupportedDefaultAdjustmentLayer(layer, footage):
			dstLayer, err := aep.NewAdjustmentLayer(targetComp, layer.Name)
			if err != nil {
				return fmt.Errorf("comp %q adjustment layer %q: %w", comp.Name, layer.Name, err)
			}
			if err := materializeLayerMetadata(dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q adjustment layer %q metadata: %w", comp.Name, layer.Name, err)
			}
			if err := materializeLayerSwitchSurface(dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q adjustment layer %q switches: %w", comp.Name, layer.Name, err)
			}
			if err := materializeDefaultTransformSurface(dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q adjustment layer %q transform: %w", comp.Name, layer.Name, err)
			}
			if err := materializeLayerTiming(dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q adjustment layer %q timing: %w", comp.Name, layer.Name, err)
			}
			recordLayer(layer, dstLayer)
		case isSupportedDefaultCameraLayer(layer):
			dstLayer, err := aep.NewCameraLayer(targetComp, layer.Name)
			if err != nil {
				return fmt.Errorf("comp %q camera layer %q: %w", comp.Name, layer.Name, err)
			}
			if err := materializeLayerMetadata(dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q camera layer %q metadata: %w", comp.Name, layer.Name, err)
			}
			if err := materializeLayerSwitchSurface(dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q camera layer %q switches: %w", comp.Name, layer.Name, err)
			}
			if err := materializeCameraLightTransformSurface(dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q camera layer %q transform: %w", comp.Name, layer.Name, err)
			}
			if err := materializeCameraOptions(dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q camera layer %q options: %w", comp.Name, layer.Name, err)
			}
			if err := materializeLayerTiming(dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q camera layer %q timing: %w", comp.Name, layer.Name, err)
			}
			recordLayer(layer, dstLayer)
		case isSupportedDefaultLightLayer(layer):
			dstLayer, err := aep.NewLightLayer(targetComp, layer.Name)
			if err != nil {
				return fmt.Errorf("comp %q light layer %q: %w", comp.Name, layer.Name, err)
			}
			if err := materializeLayerMetadata(dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q light layer %q metadata: %w", comp.Name, layer.Name, err)
			}
			if err := materializeLayerSwitchSurface(dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q light layer %q switches: %w", comp.Name, layer.Name, err)
			}
			if err := materializeCameraLightTransformSurface(dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q light layer %q transform: %w", comp.Name, layer.Name, err)
			}
			if err := materializeLightOptions(dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q light layer %q options: %w", comp.Name, layer.Name, err)
			}
			if err := materializeLayerTiming(dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q light layer %q timing: %w", comp.Name, layer.Name, err)
			}
			recordLayer(layer, dstLayer)
		case isSupportedDefaultTextLayer(layer):
			dstLayer, err := aep.NewTextLayer(targetComp, layer.Name)
			if err != nil {
				return fmt.Errorf("comp %q text layer %q: %w", comp.Name, layer.Name, err)
			}
			if err := materializeLayerMetadata(dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q text layer %q metadata: %w", comp.Name, layer.Name, err)
			}
			if err := materializeLayerSwitchSurface(dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q text layer %q switches: %w", comp.Name, layer.Name, err)
			}
			if layer.Text != nil {
				if err := dstLayer.SetText(layer.Text.Text); err != nil {
					return fmt.Errorf("comp %q text layer %q text: %w", comp.Name, layer.Name, err)
				}
				if err := materializeTextStyle(dstLayer, layer.Text); err != nil {
					return fmt.Errorf("comp %q text layer %q text style: %w", comp.Name, layer.Name, err)
				}
			}
			if err := materializeDefaultTransformSurface(dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q text layer %q transform: %w", comp.Name, layer.Name, err)
			}
			if err := materializeLayerTiming(dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q text layer %q timing: %w", comp.Name, layer.Name, err)
			}
			recordLayer(layer, dstLayer)
		case isSupportedDefaultShapeLayer(layer):
			dstLayer, err := aep.NewShapeLayer(targetComp, layer.Name)
			if err != nil {
				return fmt.Errorf("comp %q shape layer %q: %w", comp.Name, layer.Name, err)
			}
			if err := materializeLayerTiming(dstLayer.Layer, layer); err != nil {
				return fmt.Errorf("comp %q shape layer %q timing: %w", comp.Name, layer.Name, err)
			}
			recordLayer(layer, dstLayer.Layer)
		case isSupportedRectGraphicShapeLayer(layer):
			dstLayer, err := materializeRectGraphicShapeLayer(targetComp, layer)
			if err != nil {
				return fmt.Errorf("comp %q shape layer %q: %w", comp.Name, layer.Name, err)
			}
			if err := materializeLayerTiming(dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q shape layer %q timing: %w", comp.Name, layer.Name, err)
			}
			recordLayer(layer, dstLayer)
		case isSupportedDefaultPrecompLayer(layer, comps):
			sourceComp, ok := targetComps.sourceComposition(layer)
			if !ok {
				return fmt.Errorf("comp %q precomp layer %q source %q not found", comp.Name, layer.Name, layer.SourceRef.Name)
			}
			dstLayer, err := aep.NewPrecompLayer(targetComp, sourceComp, layer.Name)
			if err != nil {
				return fmt.Errorf("comp %q precomp layer %q: %w", comp.Name, layer.Name, err)
			}
			if err := materializeLayerMetadata(dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q precomp layer %q metadata: %w", comp.Name, layer.Name, err)
			}
			if err := materializeLayerSwitchSurface(dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q precomp layer %q switches: %w", comp.Name, layer.Name, err)
			}
			if err := materializePrecompTransformSurface(targetComp, dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q precomp layer %q transform: %w", comp.Name, layer.Name, err)
			}
			if err := materializeLayerTiming(dstLayer, layer); err != nil {
				return fmt.Errorf("comp %q precomp layer %q timing: %w", comp.Name, layer.Name, err)
			}
			recordLayer(layer, dstLayer)
		default:
			return fmt.Errorf("unsupported layer %q in comp %q", layer.Name, comp.Name)
		}
	}
	if err := materializeLayerRefs(target, createdLayers); err != nil {
		return fmt.Errorf("comp %q layer refs: %w", comp.Name, err)
	}
	return nil
}
