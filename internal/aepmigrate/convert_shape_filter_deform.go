package aepmigrate

import (
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeShapeZigZag(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	zigZag, err := shapeLayer.RootGroup().AddZigZag()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Zigzag Size"); ok {
		if err := zigZag.SetSize(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Zigzag Detail"); ok {
		if err := zigZag.SetDetail(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Zigzag Points"); ok {
		if err := zigZag.SetPoints(aep.ZigZagPoints(int(value))); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapePuckerBloat(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	puckerBloat, err := shapeLayer.RootGroup().AddPuckerBloat()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector PuckerBloat Amount"); ok {
		if err := puckerBloat.SetAmount(value); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapeTwist(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	twist, err := shapeLayer.RootGroup().AddTwist()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Twist Angle"); ok {
		if err := twist.SetAngle(value); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(source.Properties, "ADBE Vector Twist Center", 2); ok {
		if err := twist.SetCenter([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapeMergePaths(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	mergePaths, err := shapeLayer.RootGroup().AddMergePaths()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Merge Type"); ok {
		if err := mergePaths.SetType(aep.MergeType(int(value))); err != nil {
			return err
		}
	}
	return nil
}
