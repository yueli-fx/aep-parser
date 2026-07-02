package aepmigrate

import "github.com/yueli-fx/aep-parser/internal/profile"

func hasUnsupportedAdvancedLayerSwitches(flags profile.LayerFlags) bool {
	return flags.Is3D ||
		hasUnsupportedAdvancedLayerSwitchesAllowing3D(flags)
}

func hasUnsupportedAdvancedLayerSwitchesAllowing3D(flags profile.LayerFlags) bool {
	return flags.Solo ||
		flags.Shy ||
		flags.Locked ||
		flags.IsGuide ||
		flags.MotionBlur ||
		flags.FrameBlendEnabled ||
		flags.MarkersLocked ||
		flags.FrameBlendPixelMotion ||
		flags.SamplingBicubic ||
		flags.PreserveTransparency
}
