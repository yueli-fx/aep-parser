package aepmigrate

import (
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeRectGraphicShapeLayer(comp *aep.Composition, source profile.Layer) (*aep.Layer, error) {
	shapeLayer, err := aep.NewShapeLayer(comp, source.Name)
	if err != nil {
		return nil, err
	}
	if err := materializeShapePrimitive(shapeLayer, source.Shapes[0]); err != nil {
		return nil, err
	}
	if hasProperty(source, "ADBE Vector RoundCorner Radius") {
		if err := materializeShapeRoundCorners(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasProperty(source, "ADBE Vector Offset Amount") {
		if err := materializeShapeOffsetPaths(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasTrimFilter(source) {
		if err := materializeShapeTrim(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasZigZagFilter(source) {
		if err := materializeShapeZigZag(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasProperty(source, "ADBE Vector PuckerBloat Amount") {
		if err := materializeShapePuckerBloat(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasTwistFilter(source) {
		if err := materializeShapeTwist(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasWigglePathsFilter(source) {
		if err := materializeShapeWigglePaths(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasWiggleTransformFilter(source) {
		if err := materializeShapeWiggleTransform(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasRepeaterFilter(source) {
		if err := materializeShapeRepeater(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasProperty(source, "ADBE Vector Merge Type") {
		if err := materializeShapeMergePaths(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasGradientFillGraphic(source) {
		if err := materializeShapeGradientFill(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasGradientStrokeGraphic(source) {
		if err := materializeShapeGradientStroke(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasProperty(source, "ADBE Vector Fill Color") {
		if err := materializeShapeFill(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasProperty(source, "ADBE Vector Stroke Color") {
		if err := materializeShapeStroke(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	return shapeLayer.Layer, nil
}
