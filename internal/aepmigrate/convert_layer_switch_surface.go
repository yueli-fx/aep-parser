package aepmigrate

import (
	"fmt"
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
