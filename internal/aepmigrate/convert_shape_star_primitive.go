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
	if err := applyShapeFloatProperty(shape.Properties, "ADBE Vector Star Type", func(value float64) error {
		return star.SetStarType(aep.StarType(int(value)))
	}); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(shape.Properties, "ADBE Vector Star Points", star.SetPoints); err != nil {
		return err
	}
	if err := applyShapeVector2Property(shape.Properties, "ADBE Vector Star Position", star.SetPosition); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(shape.Properties, "ADBE Vector Star Rotation", star.SetRotation); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(shape.Properties, "ADBE Vector Star Inner Radius", star.SetInnerRadius); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(shape.Properties, "ADBE Vector Star Outer Radius", star.SetOuterRadius); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(shape.Properties, "ADBE Vector Star Inner Roundess", star.SetInnerRoundness); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(shape.Properties, "ADBE Vector Star Outer Roundess", star.SetOuterRoundness); err != nil {
		return err
	}
	return nil
}
