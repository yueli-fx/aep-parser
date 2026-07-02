package aepmigrate

import "github.com/yueli-fx/aep-parser/internal/profile"

func isSupportedDefaultNullLayer(layer profile.Layer, footage convertFootageIndex) bool {
	if layer.Type != "null" {
		return false
	}
	if layer.SourceRef != nil {
		solid, ok := footage.solidDetails(layer)
		if !ok || solid.Width == 0 || solid.Height == 0 || solid.SolidColor == nil {
			return false
		}
	}
	if layer.ParentRef != nil || layer.MatteRef != nil || layer.LightSourceRef != nil {
		return false
	}
	if layer.Text != nil || len(layer.Effects) != 0 || len(layer.Masks) != 0 || len(layer.Shapes) != 0 || len(layer.Markers) != 0 {
		return false
	}
	flags := layer.Flags
	return flags.Blend == 2 &&
		flags.TrackMatte == 0 &&
		flags.IsNull &&
		flags.EffectsEnabled &&
		flags.AudioEnabled &&
		!flags.Is3D &&
		!flags.Solo &&
		!flags.Shy &&
		!flags.Locked &&
		!flags.IsAdjustment &&
		!flags.IsGuide &&
		!flags.MotionBlur &&
		!flags.FrameBlendEnabled &&
		!flags.MarkersLocked &&
		!flags.FrameBlendPixelMotion &&
		!flags.CollapseTransform &&
		!flags.SamplingBicubic &&
		!flags.PreserveTransparency
}

func isSupportedTrackMatteMode(mode uint8) bool {
	return mode <= 4
}

func isSupportedDefaultSolidLayer(layer profile.Layer, footage convertFootageIndex) bool {
	if layer.Type != "av" || layer.SourceRef == nil || layer.SourceRef.Kind != "footage" {
		return false
	}
	solid, ok := footage.solidDetails(layer)
	if !ok || solid.Width == 0 || solid.Height == 0 || solid.SolidColor == nil {
		return false
	}
	if layer.ParentRef != nil || layer.LightSourceRef != nil {
		return false
	}
	if layer.Text != nil || !isSupportedMaskSurface(layer.Masks) || len(layer.Shapes) != 0 || len(layer.Markers) != 0 {
		return false
	}
	if !isSupportedStaticEffectSurface(layer) {
		return false
	}
	flags := layer.Flags
	return flags.Visible &&
		flags.Blend == 2 &&
		isSupportedTrackMatteMode(flags.TrackMatte) &&
		!flags.IsNull &&
		flags.EffectsEnabled &&
		flags.AudioEnabled &&
		!flags.Is3D &&
		!flags.Solo &&
		!flags.Shy &&
		!flags.Locked &&
		!flags.IsAdjustment &&
		!flags.IsGuide &&
		!flags.MotionBlur &&
		!flags.FrameBlendEnabled &&
		!flags.MarkersLocked &&
		!flags.FrameBlendPixelMotion &&
		!flags.CollapseTransform &&
		!flags.SamplingBicubic &&
		!flags.PreserveTransparency
}

func isSupportedDefaultAdjustmentLayer(layer profile.Layer, footage convertFootageIndex) bool {
	if layer.Type != "adjustment" || layer.SourceRef == nil || layer.SourceRef.Kind != "footage" {
		return false
	}
	source, ok := footage.solidDetails(layer)
	if !ok || source.Width == 0 || source.Height == 0 || source.SolidColor == nil {
		return false
	}
	if layer.MatteRef != nil || layer.LightSourceRef != nil {
		return false
	}
	if layer.Text != nil || len(layer.Masks) != 0 || len(layer.Shapes) != 0 || len(layer.Markers) != 0 {
		return false
	}
	if !isSupportedStaticEffectSurface(layer) {
		return false
	}
	flags := layer.Flags
	return flags.Visible &&
		flags.Blend == 2 &&
		flags.TrackMatte == 0 &&
		!flags.IsNull &&
		flags.IsAdjustment &&
		flags.EffectsEnabled &&
		flags.AudioEnabled &&
		!flags.Is3D &&
		!flags.Solo &&
		!flags.Shy &&
		!flags.Locked &&
		!flags.IsGuide &&
		!flags.MotionBlur &&
		!flags.FrameBlendEnabled &&
		!flags.MarkersLocked &&
		!flags.FrameBlendPixelMotion &&
		!flags.CollapseTransform &&
		!flags.SamplingBicubic &&
		!flags.PreserveTransparency
}
