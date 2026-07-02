package aepmigrate

import (
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeShapeStrokeTaper(stroke *aep.StrokeNode, source profile.Layer) error {
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Taper Start Length", stroke.Taper().SetStartLength); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Taper End Length", stroke.Taper().SetEndLength); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Taper Start Width", stroke.Taper().SetStartWidth); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Taper End Width", stroke.Taper().SetEndWidth); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Taper Start Ease", stroke.Taper().SetStartEase); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Taper End Ease", stroke.Taper().SetEndEase); err != nil {
		return err
	}
	return nil
}
