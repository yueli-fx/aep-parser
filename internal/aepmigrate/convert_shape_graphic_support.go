package aepmigrate

import "github.com/yueli-fx/aep-parser/internal/profile"

func hasGradientFillGraphic(layer profile.Layer) bool {
	return hasAllProperties(layer, "ADBE Vector Grad Colors", "ADBE Vector Grad Type") &&
		!hasProperty(layer, "ADBE Vector Stroke Width")
}

func hasGradientStrokeGraphic(layer profile.Layer) bool {
	return hasAllProperties(layer, "ADBE Vector Grad Colors", "ADBE Vector Grad Type") &&
		hasProperty(layer, "ADBE Vector Stroke Width")
}

func hasSupportedShapeFilter(layer profile.Layer) bool {
	return hasAnyProperty(layer,
		"ADBE Vector RoundCorner Radius",
		"ADBE Vector Offset Amount",
		"ADBE Vector Trim Start",
		"ADBE Vector Trim End",
		"ADBE Vector Trim Offset",
		"ADBE Vector Zigzag Size",
		"ADBE Vector Zigzag Detail",
		"ADBE Vector Zigzag Points",
		"ADBE Vector PuckerBloat Amount",
		"ADBE Vector Twist Angle",
		"ADBE Vector Twist Center",
		"ADBE Vector Merge Type",
	) ||
		hasWigglePathsFilter(layer) ||
		hasWiggleTransformFilter(layer) ||
		hasRepeaterFilter(layer)
}

func hasWigglePathsFilter(layer profile.Layer) bool {
	return hasAnyProperty(layer,
		"ADBE Vector Roughen Size",
		"ADBE Vector Roughen Detail",
		"ADBE Vector Temporal Freq",
		"ADBE Vector Roughen Points",
	)
}

func hasWiggleTransformFilter(layer profile.Layer) bool {
	return hasAnyProperty(layer,
		"ADBE Vector Wiggler Anchor",
		"ADBE Vector Wiggler Position",
		"ADBE Vector Wiggler Scale",
		"ADBE Vector Wiggler Rotation",
		"ADBE Vector Xform Temporal Freq",
	)
}

func hasRepeaterFilter(layer profile.Layer) bool {
	return hasAnyProperty(layer,
		"ADBE Vector Repeater Copies",
		"ADBE Vector Repeater Offset",
		"ADBE Vector Repeater Order",
		"ADBE Vector Repeater Anchor",
		"ADBE Vector Repeater Position",
		"ADBE Vector Repeater Scale",
		"ADBE Vector Repeater Rotation",
		"ADBE Vector Repeater Opacity 1",
		"ADBE Vector Repeater Opacity 2",
	)
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
