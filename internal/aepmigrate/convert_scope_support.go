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

func isSupportedDefaultShapeLayer(layer profile.Layer) bool {
	if layer.Type != "shape" || layer.SourceRef != nil || layer.Text != nil {
		return false
	}
	if layer.ParentRef != nil || layer.MatteRef != nil || layer.LightSourceRef != nil {
		return false
	}
	if len(layer.Effects) != 0 || len(layer.Masks) != 0 || len(layer.Shapes) != 0 || len(layer.Markers) != 0 {
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
		flags.CollapseTransform &&
		!flags.SamplingBicubic &&
		!flags.PreserveTransparency
}

func isSupportedRectGraphicShapeLayer(layer profile.Layer) bool {
	if !isSupportedShapeLayerBase(layer) {
		return false
	}
	if len(layer.Shapes) != 1 || !isSupportedParametricGraphicShape(layer.Shapes[0]) {
		return false
	}
	hasFill := hasProperty(layer, "ADBE Vector Fill Color")
	hasStroke := hasProperty(layer, "ADBE Vector Stroke Color")
	hasGradientFill := hasGradientFillGraphic(layer)
	hasGradientStroke := hasGradientStrokeGraphic(layer)
	if !hasFill && !hasStroke && !hasGradientFill && !hasGradientStroke && !hasSupportedShapeFilter(layer) {
		return false
	}
	if (hasProperty(layer, "ADBE Vector Grad Colors") && !hasGradientFill && !hasGradientStroke) ||
		hasProperty(layer, "ADBE Vector Stroke Dash 2") ||
		hasProperty(layer, "ADBE Vector Stroke Gap 2") ||
		hasProperty(layer, "ADBE Vector Stroke Offset") ||
		hasProperty(layer, "ADBE Vector Taper Wave Units") ||
		hasProperty(layer, "ADBE Vector Taper Wave Cycles") {
		return false
	}
	return true
}

func hasGradientFillGraphic(layer profile.Layer) bool {
	return hasProperty(layer, "ADBE Vector Grad Colors") &&
		hasProperty(layer, "ADBE Vector Grad Type") &&
		!hasProperty(layer, "ADBE Vector Stroke Width")
}

func hasGradientStrokeGraphic(layer profile.Layer) bool {
	return hasProperty(layer, "ADBE Vector Grad Colors") &&
		hasProperty(layer, "ADBE Vector Grad Type") &&
		hasProperty(layer, "ADBE Vector Stroke Width")
}

func hasSupportedShapeFilter(layer profile.Layer) bool {
	return hasProperty(layer, "ADBE Vector RoundCorner Radius") ||
		hasProperty(layer, "ADBE Vector Offset Amount") ||
		hasProperty(layer, "ADBE Vector Trim Start") ||
		hasProperty(layer, "ADBE Vector Trim End") ||
		hasProperty(layer, "ADBE Vector Trim Offset") ||
		hasProperty(layer, "ADBE Vector Zigzag Size") ||
		hasProperty(layer, "ADBE Vector Zigzag Detail") ||
		hasProperty(layer, "ADBE Vector Zigzag Points") ||
		hasProperty(layer, "ADBE Vector PuckerBloat Amount") ||
		hasProperty(layer, "ADBE Vector Twist Angle") ||
		hasProperty(layer, "ADBE Vector Twist Center") ||
		hasWigglePathsFilter(layer) ||
		hasWiggleTransformFilter(layer) ||
		hasRepeaterFilter(layer) ||
		hasProperty(layer, "ADBE Vector Merge Type")
}

func hasWigglePathsFilter(layer profile.Layer) bool {
	return hasProperty(layer, "ADBE Vector Roughen Size") ||
		hasProperty(layer, "ADBE Vector Roughen Detail") ||
		hasProperty(layer, "ADBE Vector Temporal Freq") ||
		hasProperty(layer, "ADBE Vector Roughen Points")
}

func hasWiggleTransformFilter(layer profile.Layer) bool {
	return hasProperty(layer, "ADBE Vector Wiggler Anchor") ||
		hasProperty(layer, "ADBE Vector Wiggler Position") ||
		hasProperty(layer, "ADBE Vector Wiggler Scale") ||
		hasProperty(layer, "ADBE Vector Wiggler Rotation") ||
		hasProperty(layer, "ADBE Vector Xform Temporal Freq")
}

func hasRepeaterFilter(layer profile.Layer) bool {
	return hasProperty(layer, "ADBE Vector Repeater Copies") ||
		hasProperty(layer, "ADBE Vector Repeater Offset") ||
		hasProperty(layer, "ADBE Vector Repeater Order") ||
		hasProperty(layer, "ADBE Vector Repeater Anchor") ||
		hasProperty(layer, "ADBE Vector Repeater Position") ||
		hasProperty(layer, "ADBE Vector Repeater Scale") ||
		hasProperty(layer, "ADBE Vector Repeater Rotation") ||
		hasProperty(layer, "ADBE Vector Repeater Opacity 1") ||
		hasProperty(layer, "ADBE Vector Repeater Opacity 2")
}

func isSupportedParametricGraphicShape(shape profile.Shape) bool {
	switch shape.Kind {
	case "rect":
		_, ok := propertyVector(shape.Properties, "ADBE Vector Rect Size", 2)
		return ok
	case "ellipse":
		_, ok := propertyVector(shape.Properties, "ADBE Vector Ellipse Size", 2)
		return ok
	case "star":
		_, hasPoints := propertyFloat(shape.Properties, "ADBE Vector Star Points")
		_, hasOuterRadius := propertyFloat(shape.Properties, "ADBE Vector Star Outer Radius")
		return hasPoints || hasOuterRadius
	default:
		return false
	}
}

func isSupportedShapeLayerBase(layer profile.Layer) bool {
	if layer.Type != "shape" || layer.SourceRef != nil || layer.Text != nil {
		return false
	}
	if layer.ParentRef != nil || layer.MatteRef != nil || layer.LightSourceRef != nil {
		return false
	}
	if len(layer.Effects) != 0 || len(layer.Masks) != 0 || len(layer.Markers) != 0 {
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
		flags.CollapseTransform &&
		!flags.SamplingBicubic &&
		!flags.PreserveTransparency
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
