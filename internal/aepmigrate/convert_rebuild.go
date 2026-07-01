package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func rebuildProject(target VersionLabel, prof *profile.Profile) (*aep.Project, error) {
	project := aep.NewProject(aepTarget(target))
	footage := newConvertFootageIndex(prof)
	comps := newConvertCompIndex(prof)
	targetComps := newConvertTargetCompIndex()
	for _, comp := range prof.Comps {
		next, err := aep.NewComposition(project, comp.Name, comp.Width, comp.Height, comp.FrameRate, comp.Duration)
		if err != nil {
			return nil, err
		}
		if err := applyStableCompSettings(next, comp); err != nil {
			return nil, err
		}
		targetComps.add(comp, next)
	}
	for _, comp := range prof.Comps {
		next, ok := targetComps.bySourceID[comp.ID]
		if !ok {
			next = targetComps.byName[comp.Name]
		}
		if next == nil {
			return nil, fmt.Errorf("comp %q: target comp missing after creation", comp.Name)
		}
		var createdLayers []convertLayerPair
		recordLayer := func(source profile.Layer, target *aep.Layer) {
			createdLayers = append(createdLayers, convertLayerPair{source: source, target: target})
		}
		for _, layer := range comp.Layers {
			switch {
			case isSupportedDefaultNullLayer(layer, footage):
				dstLayer, err := materializeNullLayer(next, layer, footage)
				if err != nil {
					return nil, fmt.Errorf("comp %q null layer %q: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerMetadata(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q null layer %q metadata: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerSwitchSurface(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q null layer %q switches: %w", comp.Name, layer.Name, err)
				}
				if err := materializeDefaultTransformSurface(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q null layer %q transform: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerTiming(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q null layer %q timing: %w", comp.Name, layer.Name, err)
				}
				recordLayer(layer, dstLayer)
			case isSupportedDefaultSolidLayer(layer, footage):
				solid, _ := footage.solidDetails(layer)
				dstLayer, err := aep.NewSolidLayer(next, layer.Name, int(solid.Width), int(solid.Height), *solid.SolidColor)
				if err != nil {
					return nil, fmt.Errorf("comp %q solid layer %q: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerMetadata(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q solid layer %q metadata: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerSwitchSurface(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q solid layer %q switches: %w", comp.Name, layer.Name, err)
				}
				if err := materializeCenteredTransformSurface(next, dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q solid layer %q transform: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerTiming(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q solid layer %q timing: %w", comp.Name, layer.Name, err)
				}
				recordLayer(layer, dstLayer)
			case isSupportedDefaultAdjustmentLayer(layer, footage):
				dstLayer, err := aep.NewAdjustmentLayer(next, layer.Name)
				if err != nil {
					return nil, fmt.Errorf("comp %q adjustment layer %q: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerMetadata(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q adjustment layer %q metadata: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerSwitchSurface(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q adjustment layer %q switches: %w", comp.Name, layer.Name, err)
				}
				if err := materializeDefaultTransformSurface(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q adjustment layer %q transform: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerTiming(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q adjustment layer %q timing: %w", comp.Name, layer.Name, err)
				}
				recordLayer(layer, dstLayer)
			case isSupportedDefaultCameraLayer(layer):
				dstLayer, err := aep.NewCameraLayer(next, layer.Name)
				if err != nil {
					return nil, fmt.Errorf("comp %q camera layer %q: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerMetadata(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q camera layer %q metadata: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerSwitchSurface(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q camera layer %q switches: %w", comp.Name, layer.Name, err)
				}
				if err := materializeCameraLightTransformSurface(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q camera layer %q transform: %w", comp.Name, layer.Name, err)
				}
				if err := materializeCameraOptions(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q camera layer %q options: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerTiming(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q camera layer %q timing: %w", comp.Name, layer.Name, err)
				}
				recordLayer(layer, dstLayer)
			case isSupportedDefaultLightLayer(layer):
				dstLayer, err := aep.NewLightLayer(next, layer.Name)
				if err != nil {
					return nil, fmt.Errorf("comp %q light layer %q: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerMetadata(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q light layer %q metadata: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerSwitchSurface(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q light layer %q switches: %w", comp.Name, layer.Name, err)
				}
				if err := materializeCameraLightTransformSurface(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q light layer %q transform: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLightOptions(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q light layer %q options: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerTiming(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q light layer %q timing: %w", comp.Name, layer.Name, err)
				}
				recordLayer(layer, dstLayer)
			case isSupportedDefaultTextLayer(layer):
				dstLayer, err := aep.NewTextLayer(next, layer.Name)
				if err != nil {
					return nil, fmt.Errorf("comp %q text layer %q: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerMetadata(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q text layer %q metadata: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerSwitchSurface(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q text layer %q switches: %w", comp.Name, layer.Name, err)
				}
				if layer.Text != nil {
					if err := dstLayer.SetText(layer.Text.Text); err != nil {
						return nil, fmt.Errorf("comp %q text layer %q text: %w", comp.Name, layer.Name, err)
					}
					if err := materializeTextStyle(dstLayer, layer.Text); err != nil {
						return nil, fmt.Errorf("comp %q text layer %q text style: %w", comp.Name, layer.Name, err)
					}
				}
				if err := materializeDefaultTransformSurface(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q text layer %q transform: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerTiming(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q text layer %q timing: %w", comp.Name, layer.Name, err)
				}
				recordLayer(layer, dstLayer)
			case isSupportedDefaultShapeLayer(layer):
				dstLayer, err := aep.NewShapeLayer(next, layer.Name)
				if err != nil {
					return nil, fmt.Errorf("comp %q shape layer %q: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerTiming(dstLayer.Layer, layer); err != nil {
					return nil, fmt.Errorf("comp %q shape layer %q timing: %w", comp.Name, layer.Name, err)
				}
				recordLayer(layer, dstLayer.Layer)
			case isSupportedRectGraphicShapeLayer(layer):
				dstLayer, err := materializeRectGraphicShapeLayer(next, layer)
				if err != nil {
					return nil, fmt.Errorf("comp %q shape layer %q: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerTiming(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q shape layer %q timing: %w", comp.Name, layer.Name, err)
				}
				recordLayer(layer, dstLayer)
			case isSupportedDefaultPrecompLayer(layer, comps):
				sourceComp, ok := targetComps.sourceComposition(layer)
				if !ok {
					return nil, fmt.Errorf("comp %q precomp layer %q source %q not found", comp.Name, layer.Name, layer.SourceRef.Name)
				}
				dstLayer, err := aep.NewPrecompLayer(next, sourceComp, layer.Name)
				if err != nil {
					return nil, fmt.Errorf("comp %q precomp layer %q: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerMetadata(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q precomp layer %q metadata: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerSwitchSurface(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q precomp layer %q switches: %w", comp.Name, layer.Name, err)
				}
				if err := materializePrecompTransformSurface(next, dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q precomp layer %q transform: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerTiming(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q precomp layer %q timing: %w", comp.Name, layer.Name, err)
				}
				recordLayer(layer, dstLayer)
			default:
				return nil, fmt.Errorf("unsupported layer %q in comp %q", layer.Name, comp.Name)
			}
		}
		if err := materializeLayerRefs(target, createdLayers); err != nil {
			return nil, fmt.Errorf("comp %q layer refs: %w", comp.Name, err)
		}
	}
	return project, nil
}

func aepTarget(target VersionLabel) aep.AETarget {
	switch target {
	case VersionAE2022:
		return aep.TargetAE2022
	case VersionAE2025:
		return aep.TargetAE2025
	default:
		return aep.TargetAE2020
	}
}
