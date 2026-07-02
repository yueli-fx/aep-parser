package aepmigrate

import (
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeShapeGradientStroke(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	stroke, err := shapeLayer.RootGroup().AddGradientStroke()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Grad Type"); ok {
		if err := stroke.SetGradientType(aep.GradientType(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(source.Properties, "ADBE Vector Grad Start Pt", 2); ok {
		if err := stroke.SetStartPoint([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(source.Properties, "ADBE Vector Grad End Pt", 2); ok {
		if err := stroke.SetEndPoint([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Grad HiLite Length"); ok {
		if err := stroke.SetHighlightLength(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Grad HiLite Angle"); ok {
		if err := stroke.SetHighlightAngle(value); err != nil {
			return err
		}
	}
	if gradient, ok := propertyGradient(source.Properties, "ADBE Vector Grad Colors"); ok {
		if len(gradient.ColorStops) > 0 {
			if err := stroke.SetColorStops(gradient.ColorStops); err != nil {
				return err
			}
		}
		if len(gradient.AlphaStops) > 0 {
			if err := stroke.SetAlphaStops(gradient.AlphaStops); err != nil {
				return err
			}
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Width"); ok {
		if err := stroke.SetStrokeWidth(value); err != nil {
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
	return nil
}
