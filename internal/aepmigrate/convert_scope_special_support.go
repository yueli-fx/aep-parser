package aepmigrate

import "github.com/yueli-fx/aep-parser/internal/profile"

func isSupportedDefaultCameraLayer(layer profile.Layer) bool {
	return isSupportedDefaultCameraOrLightLayer(layer, "camera")
}

func isSupportedDefaultLightLayer(layer profile.Layer) bool {
	return isSupportedDefaultCameraOrLightLayer(layer, "light")
}

func isSupportedDefaultCameraOrLightLayer(layer profile.Layer, typ string) bool {
	if layer.Type != typ || layer.SourceRef != nil {
		return false
	}
	if layer.ParentRef != nil || layer.MatteRef != nil {
		return false
	}
	if typ != "light" && layer.LightSourceRef != nil {
		return false
	}
	if layer.Text != nil || len(layer.Effects) != 0 || len(layer.Masks) != 0 || len(layer.Shapes) != 0 || len(layer.Markers) != 0 {
		return false
	}
	flags := layer.Flags
	return flags.Visible &&
		flags.TrackMatte == 0 &&
		!flags.IsNull &&
		!flags.IsAdjustment &&
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

func isSupportedDefaultTextLayer(layer profile.Layer) bool {
	if layer.Type != "text" || layer.Text == nil || layer.SourceRef != nil {
		return false
	}
	if layer.MatteRef != nil || layer.LightSourceRef != nil {
		return false
	}
	if len(layer.Masks) != 0 || len(layer.Shapes) != 0 || len(layer.Markers) != 0 {
		return false
	}
	if !isSupportedStaticEffectSurface(layer) {
		return false
	}
	if layer.Text.IsBoxText {
		return false
	}
	flags := layer.Flags
	return flags.Blend != 0 &&
		flags.TrackMatte == 0 &&
		!flags.IsNull &&
		flags.CollapseTransform
}

func isSupportedDefaultPrecompLayer(layer profile.Layer, comps convertCompIndex) bool {
	if layer.Type != "av" || layer.SourceRef == nil || layer.SourceRef.Kind != "composition" {
		return false
	}
	if _, ok := comps.sourceComposition(layer); !ok {
		return false
	}
	if layer.ParentRef != nil || layer.MatteRef != nil || layer.LightSourceRef != nil {
		return false
	}
	if layer.Text != nil || len(layer.Effects) != 0 || len(layer.Masks) != 0 || len(layer.Shapes) != 0 || len(layer.Markers) != 0 {
		return false
	}
	flags := layer.Flags
	return flags.Visible &&
		flags.Blend == 2 &&
		flags.TrackMatte == 0 &&
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
