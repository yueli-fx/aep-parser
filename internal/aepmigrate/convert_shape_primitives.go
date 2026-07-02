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
	if value, ok := propertyVector(shape.Properties, "ADBE Vector Rect Size", 2); ok {
		if err := rect.SetSize([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(shape.Properties, "ADBE Vector Rect Position", 2); ok {
		if err := rect.SetPosition([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(shape.Properties, "ADBE Vector Rect Roundness"); ok {
		if err := rect.SetRoundness(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(shape.Properties, "ADBE Vector Shape Direction"); ok {
		if err := rect.SetDirection(aep.ShapeDirection(int(value))); err != nil {
			return err
		}
	}
	return nil
}

func materializeEllipsePrimitive(shapeLayer *aep.ShapeLayer, shape profile.Shape) error {
	ellipse, err := shapeLayer.RootGroup().AddEllipse()
	if err != nil {
		return err
	}
	if value, ok := propertyVector(shape.Properties, "ADBE Vector Ellipse Size", 2); ok {
		if err := ellipse.SetSize([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(shape.Properties, "ADBE Vector Ellipse Position", 2); ok {
		if err := ellipse.SetPosition([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(shape.Properties, "ADBE Vector Shape Direction"); ok {
		if err := ellipse.SetDirection(aep.ShapeDirection(int(value))); err != nil {
			return err
		}
	}
	return nil
}
