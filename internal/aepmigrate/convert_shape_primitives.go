package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeShapePrimitive(shapeLayer *aep.ShapeLayer, shape profile.Shape) error {
	switch shape.Kind {
	case "rect":
		return materializeRectPrimitive(shapeLayer, shape)
	case "ellipse":
		return materializeEllipsePrimitive(shapeLayer, shape)
	case "star":
		return materializeStarPrimitive(shapeLayer, shape)
	default:
		return fmt.Errorf("unsupported shape primitive %q", shape.Kind)
	}
}

func materializeRectPrimitive(shapeLayer *aep.ShapeLayer, shape profile.Shape) error {
	rect, err := shapeLayer.RootGroup().AddRect()
	if err != nil {
		return err
	}
	if err := applyShapeVector2Property(shape.Properties, "ADBE Vector Rect Size", rect.SetSize); err != nil {
		return err
	}
	if err := applyShapeVector2Property(shape.Properties, "ADBE Vector Rect Position", rect.SetPosition); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(shape.Properties, "ADBE Vector Rect Roundness", rect.SetRoundness); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(shape.Properties, "ADBE Vector Shape Direction", func(value float64) error {
		return rect.SetDirection(aep.ShapeDirection(int(value)))
	}); err != nil {
		return err
	}
	return nil
}

func materializeEllipsePrimitive(shapeLayer *aep.ShapeLayer, shape profile.Shape) error {
	ellipse, err := shapeLayer.RootGroup().AddEllipse()
	if err != nil {
		return err
	}
	if err := applyShapeVector2Property(shape.Properties, "ADBE Vector Ellipse Size", ellipse.SetSize); err != nil {
		return err
	}
	if err := applyShapeVector2Property(shape.Properties, "ADBE Vector Ellipse Position", ellipse.SetPosition); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(shape.Properties, "ADBE Vector Shape Direction", func(value float64) error {
		return ellipse.SetDirection(aep.ShapeDirection(int(value)))
	}); err != nil {
		return err
	}
	return nil
}
