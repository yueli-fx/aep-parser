package aepmigrate

import "github.com/yueli-fx/aep-parser/internal/profile"

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
