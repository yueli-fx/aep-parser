package aepmigrate

import "github.com/yueli-fx/aep-parser/internal/profile"

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
		!flags.IsAdjustment &&
		flags.CollapseTransform &&
		!hasUnsupportedAdvancedLayerSwitches(flags)
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
		hasAnyProperty(layer,
			"ADBE Vector Stroke Dash 2",
			"ADBE Vector Stroke Gap 2",
			"ADBE Vector Stroke Offset",
			"ADBE Vector Taper Wave Units",
			"ADBE Vector Taper Wave Cycles",
		) {
		return false
	}
	return true
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
		!flags.IsAdjustment &&
		flags.CollapseTransform &&
		!hasUnsupportedAdvancedLayerSwitches(flags)
}
