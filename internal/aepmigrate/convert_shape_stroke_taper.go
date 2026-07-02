package aepmigrate

import (
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeShapeStrokeTaper(stroke *aep.StrokeNode, source profile.Layer) error {
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper Start Length"); ok {
		if err := stroke.Taper().SetStartLength(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper End Length"); ok {
		if err := stroke.Taper().SetEndLength(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper Start Width"); ok {
		if err := stroke.Taper().SetStartWidth(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper End Width"); ok {
		if err := stroke.Taper().SetEndWidth(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper Start Ease"); ok {
		if err := stroke.Taper().SetStartEase(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper End Ease"); ok {
		if err := stroke.Taper().SetEndEase(value); err != nil {
			return err
		}
	}
	return nil
}
