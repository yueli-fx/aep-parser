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
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Zigzag Size", zigZag.SetSize); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Zigzag Detail", zigZag.SetDetail); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Zigzag Points", func(value float64) error {
		return zigZag.SetPoints(aep.ZigZagPoints(int(value)))
	}); err != nil {
		return err
	}
	return nil
}

func materializeShapePuckerBloat(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	puckerBloat, err := shapeLayer.RootGroup().AddPuckerBloat()
	if err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector PuckerBloat Amount", puckerBloat.SetAmount); err != nil {
		return err
	}
	return nil
}

func materializeShapeTwist(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	twist, err := shapeLayer.RootGroup().AddTwist()
	if err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Twist Angle", twist.SetAngle); err != nil {
		return err
	}
	if err := applyShapeVector2Property(source.Properties, "ADBE Vector Twist Center", twist.SetCenter); err != nil {
		return err
	}
	return nil
}

func materializeShapeMergePaths(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	mergePaths, err := shapeLayer.RootGroup().AddMergePaths()
	if err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Merge Type", func(value float64) error {
		return mergePaths.SetType(aep.MergeType(int(value)))
	}); err != nil {
		return err
	}
	return nil
}
