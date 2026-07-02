package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeTextRunMetrics(layer *aep.Layer, index int, run profile.TextStyleRun) error {
	if err := layer.SetRunFontSize(index, run.FontSize); err != nil {
		return fmt.Errorf("run %d font_size: %w", index, err)
	}
	if err := layer.SetRunFillColor(index, run.FillColor); err != nil {
		return fmt.Errorf("run %d fill_color: %w", index, err)
	}
	if err := layer.SetRunAutoLeading(index, run.AutoLeading); err != nil {
		return fmt.Errorf("run %d auto_leading: %w", index, err)
	}
	if err := layer.SetRunLeading(index, run.Leading); err != nil {
		return fmt.Errorf("run %d leading: %w", index, err)
	}
	if err := layer.SetRunTracking(index, run.Tracking); err != nil {
		return fmt.Errorf("run %d tracking: %w", index, err)
	}
	if err := layer.SetRunBaselineShift(index, run.BaselineShift); err != nil {
		return fmt.Errorf("run %d baseline_shift: %w", index, err)
	}
	if err := layer.SetRunHorizontalScale(index, run.HorizontalScale); err != nil {
		return fmt.Errorf("run %d horizontal_scale: %w", index, err)
	}
	if err := layer.SetRunVerticalScale(index, run.VerticalScale); err != nil {
		return fmt.Errorf("run %d vertical_scale: %w", index, err)
	}
	if err := layer.SetRunTsume(index, run.Tsume); err != nil {
		return fmt.Errorf("run %d tsume: %w", index, err)
	}
	return nil
}
