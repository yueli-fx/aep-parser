package aepmigrate

import (
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeShapeRoundCorners(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	roundCorners, err := shapeLayer.RootGroup().AddRoundCorners()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector RoundCorner Radius"); ok {
		if err := roundCorners.SetRadius(value); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapeOffsetPaths(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	offsetPaths, err := shapeLayer.RootGroup().AddOffsetPaths()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Offset Amount"); ok {
		if err := offsetPaths.SetAmount(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Offset Line Join"); ok {
		if err := offsetPaths.SetLineJoin(aep.StrokeLineJoin(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Offset Miter Limit"); ok {
		if err := offsetPaths.SetMiterLimit(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Offset Copies"); ok {
		if err := offsetPaths.SetCopies(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Offset Copy Offset"); ok {
		if err := offsetPaths.SetCopyOffset(value); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapeTrim(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	trim, err := shapeLayer.RootGroup().AddTrim()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Trim Start"); ok {
		if err := trim.SetStart(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Trim End"); ok {
		if err := trim.SetEnd(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Trim Offset"); ok {
		if err := trim.SetOffset(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Trim Type"); ok {
		if err := trim.SetType(aep.TrimType(int(value))); err != nil {
			return err
		}
	}
	return nil
}
