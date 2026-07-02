package aepmigrate

import (
	"fmt"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeCameraOptions(layer *aep.Layer, source profile.Layer) error {
	scalars := []struct {
		matchName string
		set       func(float64) error
	}{
		{aep.MatchNameCameraZoom, layer.SetCameraZoom},
		{aep.MatchNameCameraDepthOfField, func(v float64) error { return layer.SetCameraDepthOfField(v != 0) }},
		{aep.MatchNameCameraFocusDistance, layer.SetCameraFocusDistance},
		{aep.MatchNameCameraAperture, layer.SetCameraAperture},
		{aep.MatchNameCameraBlurLevel, layer.SetCameraBlurLevel},
		{aep.MatchNameCameraIrisShape, layer.SetIrisShape},
		{aep.MatchNameCameraIrisRotation, layer.SetIrisRotation},
		{aep.MatchNameCameraIrisRoundness, layer.SetIrisRoundness},
		{aep.MatchNameCameraIrisAspectRatio, layer.SetIrisAspectRatio},
		{aep.MatchNameCameraIrisDiffractionFringe, layer.SetIrisDiffractionFringe},
		{aep.MatchNameCameraIrisHighlightGain, layer.SetIrisHighlightGain},
		{aep.MatchNameCameraIrisHighlightThreshold, layer.SetIrisHighlightThreshold},
		{aep.MatchNameCameraIrisHighlightSaturation, layer.SetIrisHighlightSaturation},
	}
	for _, scalar := range scalars {
		value, ok := propertyFloat(source.Properties, scalar.matchName)
		if !ok {
			continue
		}
		if err := scalar.set(value); err != nil {
			return fmt.Errorf("%s: %w", scalar.matchName, err)
		}
	}
	return nil
}

func materializeLightOptions(layer *aep.Layer, source profile.Layer) error {
	if source.LightKind != "" {
		kind, ok := convertLightKind(source.LightKind)
		if !ok {
			return fmt.Errorf("light_kind %q unsupported", source.LightKind)
		}
		if err := layer.SetLightKind(kind); err != nil {
			return fmt.Errorf("light_kind: %w", err)
		}
	}
	if value, ok := propertyVector(source.Properties, aep.MatchNameLightColor, 4); ok {
		if err := layer.SetLightColor(value); err != nil {
			return fmt.Errorf("%s: %w", aep.MatchNameLightColor, err)
		}
	}
	scalars := []struct {
		matchName string
		set       func(float64) error
	}{
		{aep.MatchNameLightIntensity, layer.SetLightIntensity},
		{aep.MatchNameLightConeAngle, layer.SetLightConeAngle},
		{aep.MatchNameLightConeFeather, layer.SetLightConeFeather},
		{aep.MatchNameLightFalloffType, layer.SetLightFalloffType},
		{aep.MatchNameLightFalloffStart, layer.SetLightFalloffStart},
		{aep.MatchNameLightFalloffDistance, layer.SetLightFalloffDistance},
		{aep.MatchNameLightCastsShadows, func(v float64) error { return layer.SetLightCastsShadows(v != 0) }},
		{aep.MatchNameLightShadowDarkness, layer.SetLightShadowDarkness},
		{aep.MatchNameLightShadowDiffusion, layer.SetLightShadowDiffusion},
	}
	for _, scalar := range scalars {
		value, ok := propertyFloat(source.Properties, scalar.matchName)
		if !ok {
			continue
		}
		if err := scalar.set(value); err != nil {
			return fmt.Errorf("%s: %w", scalar.matchName, err)
		}
	}
	return nil
}

func convertLightKind(value string) (aep.LightKind, bool) {
	switch strings.ToLower(value) {
	case "parallel":
		return aep.LightKindParallel, true
	case "spot":
		return aep.LightKindSpot, true
	case "point":
		return aep.LightKindPoint, true
	case "ambient":
		return aep.LightKindAmbient, true
	default:
		return 0, false
	}
}
