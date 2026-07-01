package aepmigrate

import (
	"fmt"
	"math"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

type convertLayerPair struct {
	source profile.Layer
	target *aep.Layer
}

func materializeLayerRefs(target VersionLabel, layers []convertLayerPair) error {
	bySourceID := map[uint32]*aep.Layer{}
	byName := map[string]*aep.Layer{}
	for _, pair := range layers {
		if pair.source.ID != 0 {
			bySourceID[pair.source.ID] = pair.target
		}
		if _, exists := byName[pair.source.Name]; !exists {
			byName[pair.source.Name] = pair.target
		}
	}
	for _, pair := range layers {
		if pair.source.ParentRef == nil {
			continue
		}
		parent := bySourceID[pair.source.ParentRef.ID]
		if parent == nil && pair.source.ParentRef.Name != "" {
			parent = byName[pair.source.ParentRef.Name]
		}
		if parent == nil {
			return fmt.Errorf("layer %q parent %q not found", pair.source.Name, pair.source.ParentRef.Name)
		}
		if err := pair.target.SetParent(parent.ID); err != nil {
			return fmt.Errorf("layer %q parent %q: %w", pair.source.Name, pair.source.ParentRef.Name, err)
		}
	}
	for _, pair := range layers {
		if pair.source.LightSourceRef == nil {
			continue
		}
		source := bySourceID[pair.source.LightSourceRef.ID]
		if source == nil && pair.source.LightSourceRef.Name != "" {
			source = byName[pair.source.LightSourceRef.Name]
		}
		if source == nil {
			return fmt.Errorf("layer %q light source %q not found", pair.source.Name, pair.source.LightSourceRef.Name)
		}
		if err := pair.target.SetLightSource(source); err != nil {
			return fmt.Errorf("layer %q light source %q: %w", pair.source.Name, pair.source.LightSourceRef.Name, err)
		}
	}
	for _, pair := range layers {
		mode, ok := convertTrackMatteMode(pair.source.Flags.TrackMatte)
		if !ok {
			return fmt.Errorf("layer %q track matte mode %d unsupported", pair.source.Name, pair.source.Flags.TrackMatte)
		}
		if mode == aep.TrackMatteNone {
			continue
		}
		if pair.source.MatteRef != nil && pair.source.MatteRefKind == "explicit" && target == VersionAE2025 {
			matte := bySourceID[pair.source.MatteRef.ID]
			if matte == nil && pair.source.MatteRef.Name != "" {
				matte = byName[pair.source.MatteRef.Name]
			}
			if matte == nil {
				return fmt.Errorf("layer %q matte source %q not found", pair.source.Name, pair.source.MatteRef.Name)
			}
			if err := pair.target.SetTrackMatteSource(matte, mode); err != nil {
				return fmt.Errorf("layer %q matte source %q: %w", pair.source.Name, pair.source.MatteRef.Name, err)
			}
			continue
		}
		if err := pair.target.SetTrackMatte(mode); err != nil {
			return fmt.Errorf("layer %q track matte: %w", pair.source.Name, err)
		}
	}
	return nil
}

func convertTrackMatteMode(mode uint8) (aep.TrackMatteType, bool) {
	switch mode {
	case uint8(aep.TrackMatteNone):
		return aep.TrackMatteNone, true
	case uint8(aep.TrackMatteAlpha):
		return aep.TrackMatteAlpha, true
	case uint8(aep.TrackMatteAlphaInverse):
		return aep.TrackMatteAlphaInverse, true
	case uint8(aep.TrackMatteLuma):
		return aep.TrackMatteLuma, true
	case uint8(aep.TrackMatteLumaInverse):
		return aep.TrackMatteLumaInverse, true
	default:
		return aep.TrackMatteNone, false
	}
}

func materializeLayerMetadata(layer *aep.Layer, source profile.Layer) error {
	if source.Label != 0 && source.Label != layer.Label {
		if err := layer.SetLabel(source.Label); err != nil {
			return err
		}
	}
	if source.Comment != "" {
		if err := layer.SetComment(source.Comment); err != nil {
			return err
		}
	}
	return nil
}

func materializeNullLayer(comp *aep.Composition, source profile.Layer, footage convertFootageIndex) (*aep.Layer, error) {
	if solid, ok := footage.solidDetails(source); ok && solid.Width != 0 && solid.Height != 0 && solid.SolidColor != nil {
		layer, err := aep.NewSolidLayer(comp, source.Name, int(solid.Width), int(solid.Height), *solid.SolidColor)
		if err != nil {
			return nil, err
		}
		if err := layer.SetIsNull(true); err != nil {
			return nil, err
		}
		return layer, nil
	}
	return aep.NewNullLayer(comp, source.Name)
}

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

func convertAutoOrient(value string) (aep.AutoOrientType, bool) {
	switch strings.ToLower(value) {
	case "none":
		return aep.AutoOrientNone, true
	case "along_path", "along-path":
		return aep.AutoOrientAlongPath, true
	case "camera_or_point_of_interest", "camera-or-point-of-interest":
		return aep.AutoOrientCameraOrPointOfInterest, true
	case "characters_toward_camera", "characters-toward-camera":
		return aep.AutoOrientCharactersTowardCamera, true
	default:
		return 0, false
	}
}

func materializeLayerSwitchSurface(layer *aep.Layer, source profile.Layer) error {
	flags := source.Flags
	if err := layer.SetVisible(flags.Visible); err != nil {
		return err
	}
	if err := layer.SetSolo(flags.Solo); err != nil {
		return err
	}
	if err := layer.SetShy(flags.Shy); err != nil {
		return err
	}
	if err := layer.SetLocked(flags.Locked); err != nil {
		return err
	}
	if err := layer.SetEffectsEnabled(flags.EffectsEnabled); err != nil {
		return err
	}
	if err := layer.SetAudioEnabled(flags.AudioEnabled); err != nil {
		return err
	}
	if err := layer.SetMotionBlur(flags.MotionBlur); err != nil {
		return err
	}
	if err := layer.SetFrameBlendEnabled(flags.FrameBlendEnabled); err != nil {
		return err
	}
	if err := layer.SetMarkersLocked(flags.MarkersLocked); err != nil {
		return err
	}
	if err := layer.SetCollapseTransform(flags.CollapseTransform); err != nil {
		return err
	}
	if err := layer.SetIs3D(flags.Is3D); err != nil {
		return err
	}
	if err := layer.SetIsAdjust(flags.IsAdjustment); err != nil {
		return err
	}
	if err := layer.SetIsGuide(flags.IsGuide); err != nil {
		return err
	}
	if err := layer.SetSamplingBicubic(flags.SamplingBicubic); err != nil {
		return err
	}
	if err := layer.SetFrameBlendPixelMotion(flags.FrameBlendPixelMotion); err != nil {
		return err
	}
	if err := layer.SetPreserveTransparency(flags.PreserveTransparency); err != nil {
		return err
	}
	if source.AutoOrient != "" {
		autoOrient, ok := convertAutoOrient(source.AutoOrient)
		if !ok {
			return fmt.Errorf("auto_orient %q unsupported", source.AutoOrient)
		}
		if err := layer.SetAutoOrient(autoOrient); err != nil {
			return err
		}
	}
	if flags.Blend != 0 {
		if err := layer.SetBlendingMode(aep.BlendingMode(flags.Blend)); err != nil {
			return err
		}
	}
	quality, ok := convertLayerQuality(source.Quality)
	if ok {
		if err := layer.SetQuality(quality); err != nil {
			return err
		}
	}
	return nil
}

func convertLayerQuality(value string) (aep.LayerQuality, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "wireframe":
		return aep.LayerQualityWireframe, true
	case "draft":
		return aep.LayerQualityDraft, true
	case "best":
		return aep.LayerQualityBest, true
	default:
		return 0, false
	}
}

func materializeDefaultTransformSurface(layer *aep.Layer, source profile.Layer) error {
	if !hasTransformProperties(source) {
		return nil
	}
	transform, err := layerTransformFromStaticProfile(source)
	if err != nil {
		return err
	}
	return aep.SetLayerTransform(layer, transform)
}

func materializeCenteredTransformSurface(comp *aep.Composition, layer *aep.Layer, source profile.Layer) error {
	if !hasTransformProperties(source) {
		return nil
	}
	if layer == nil {
		return fmt.Errorf("layer not found after creation")
	}
	if _, ok := propertyVectorAtLeast(source.Properties, "ADBE Position", 2); ok {
		return materializeDefaultTransformSurface(layer, source)
	}
	transform := aep.NewLayerTransform()
	if err := transform.Position().SetStaticValue([2]float64{float64(comp.Width) / 2, float64(comp.Height) / 2}); err != nil {
		return err
	}
	return aep.SetLayerTransform(layer, transform)
}

func materializeCameraLightTransformSurface(layer *aep.Layer, source profile.Layer) error {
	if hasProperty(source, "ADBE Position_2") {
		return nil
	}
	return materializeDefaultTransformSurface(layer, source)
}

func materializePrecompTransformSurface(comp *aep.Composition, layer *aep.Layer, source profile.Layer) error {
	if hasStaticPropertyVector(source, "ADBE Position", []float64{0, 0, 0}) {
		return materializeDefaultTransformSurface(layer, source)
	}
	return materializeCenteredTransformSurface(comp, layer, source)
}

func materializeLayerTiming(layer *aep.Layer, source profile.Layer) error {
	if source.Timing.StartTime != 0 {
		if err := layer.SetStartTime(source.Timing.StartTime); err != nil {
			return err
		}
	}
	if source.Timing.InPoint != 0 {
		if err := layer.SetInPoint(source.Timing.InPoint); err != nil {
			return err
		}
	}
	if source.Timing.OutPoint != 0 && !isDefaultLayerOutPoint(source.Timing) {
		if err := layer.SetOutPoint(source.Timing.OutPoint); err != nil {
			return err
		}
	}
	if source.Timing.Stretch != 0 && source.Timing.Stretch != 1 {
		if err := layer.SetStretch(source.Timing.Stretch); err != nil {
			return err
		}
	}
	return nil
}

func isDefaultLayerOutPoint(timing profile.LayerTiming) bool {
	return timing.InPoint == 0 && math.Abs(timing.OutPoint-timing.Duration) < 1e-6
}
