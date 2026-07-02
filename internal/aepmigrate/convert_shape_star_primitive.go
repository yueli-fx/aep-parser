package aepmigrate

import (
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeStarPrimitive(shapeLayer *aep.ShapeLayer, shape profile.Shape) error {
	star, err := shapeLayer.RootGroup().AddStar()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(shape.Properties, "ADBE Vector Star Type"); ok {
		if err := star.SetStarType(aep.StarType(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(shape.Properties, "ADBE Vector Star Points"); ok {
		if err := star.SetPoints(value); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(shape.Properties, "ADBE Vector Star Position", 2); ok {
		if err := star.SetPosition([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(shape.Properties, "ADBE Vector Star Rotation"); ok {
		if err := star.SetRotation(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(shape.Properties, "ADBE Vector Star Inner Radius"); ok {
		if err := star.SetInnerRadius(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(shape.Properties, "ADBE Vector Star Outer Radius"); ok {
		if err := star.SetOuterRadius(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(shape.Properties, "ADBE Vector Star Inner Roundess"); ok {
		if err := star.SetInnerRoundness(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(shape.Properties, "ADBE Vector Star Outer Roundess"); ok {
		if err := star.SetOuterRoundness(value); err != nil {
			return err
		}
	}
	return nil
}
