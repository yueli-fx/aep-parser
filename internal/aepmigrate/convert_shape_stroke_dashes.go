package aepmigrate

import (
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeShapeStrokeDashes(stroke *aep.StrokeNode, source profile.Layer) error {
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Dash 1"); ok {
		if err := stroke.Dashes().SetDash(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Gap 1"); ok {
		if err := stroke.Dashes().SetGap(value); err != nil {
			return err
		}
	}
	return nil
}
