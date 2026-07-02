package aepmigrate

import (
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeShapeStroke(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	stroke, err := shapeLayer.RootGroup().AddStroke()
	if err != nil {
		return err
	}
	if err := materializeShapeStrokeBase(stroke, source); err != nil {
		return err
	}
	if err := materializeShapeStrokeDashes(stroke, source); err != nil {
		return err
	}
	if err := materializeShapeStrokeTaper(stroke, source); err != nil {
		return err
	}
	return materializeShapeStrokeWave(stroke, source)
}

func materializeShapeStrokeBase(stroke *aep.StrokeNode, source profile.Layer) error {
	if value, ok := propertyVector(source.Properties, "ADBE Vector Stroke Color", 4); ok {
		if err := stroke.SetColor(profileARGBToRGBA(value)); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Opacity"); ok {
		if err := stroke.SetOpacity(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Width"); ok {
		if err := stroke.SetWidth(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Line Cap"); ok {
		if err := stroke.SetLineCap(aep.StrokeLineCap(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Line Join"); ok {
		if err := stroke.SetLineJoin(aep.StrokeLineJoin(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Miter Limit"); ok {
		if err := stroke.SetMiterLimit(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Blend Mode"); ok {
		if err := stroke.SetBlendMode(aep.ShapeBlendMode(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Composite Order"); ok {
		if err := stroke.SetCompositeOrder(aep.ShapeCompositeOrder(int(value))); err != nil {
			return err
		}
	}
	return nil
}
