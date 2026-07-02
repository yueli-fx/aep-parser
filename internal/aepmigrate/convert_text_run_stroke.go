package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeTextRunStroke(layer *aep.Layer, index int, run profile.TextStyleRun) error {
	if err := layer.SetRunApplyStroke(index, run.ApplyStroke); err != nil {
		return fmt.Errorf("run %d apply_stroke: %w", index, err)
	}
	if err := layer.SetRunStrokeColor(index, run.StrokeColor); err != nil {
		return fmt.Errorf("run %d stroke_color: %w", index, err)
	}
	if err := layer.SetRunStrokeWidth(index, run.StrokeWidth); err != nil {
		return fmt.Errorf("run %d stroke_width: %w", index, err)
	}
	if err := layer.SetRunStrokeOverFill(index, run.StrokeOverFill); err != nil {
		return fmt.Errorf("run %d stroke_over_fill: %w", index, err)
	}
	return nil
}
