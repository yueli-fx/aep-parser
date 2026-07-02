package aepmigrate

import (
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeShapeStrokeDashes(stroke *aep.StrokeNode, source profile.Layer) error {
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Stroke Dash 1", stroke.Dashes().SetDash); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Stroke Gap 1", stroke.Dashes().SetGap); err != nil {
		return err
	}
	return nil
}
