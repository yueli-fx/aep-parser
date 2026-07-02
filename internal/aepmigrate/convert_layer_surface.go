package aepmigrate

import (
	"fmt"
	"math"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

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
