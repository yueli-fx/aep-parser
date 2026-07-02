package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeTextRunOptions(layer *aep.Layer, index int, run profile.TextStyleRun) error {
	caps, err := textCapsOption(run.CapsOption)
	if err != nil {
		return fmt.Errorf("run %d caps_option: %w", index, err)
	}
	if err := layer.SetRunCapsOption(index, caps); err != nil {
		return fmt.Errorf("run %d caps_option: %w", index, err)
	}
	baseline, err := textBaselineOption(run.BaselineOption)
	if err != nil {
		return fmt.Errorf("run %d baseline_option: %w", index, err)
	}
	if err := layer.SetRunBaselineOption(index, baseline); err != nil {
		return fmt.Errorf("run %d baseline_option: %w", index, err)
	}
	kern, err := textAutoKernType(run.AutoKernType)
	if err != nil {
		return fmt.Errorf("run %d auto_kern_type: %w", index, err)
	}
	if err := layer.SetRunAutoKernType(index, kern); err != nil {
		return fmt.Errorf("run %d auto_kern_type: %w", index, err)
	}
	lineJoin, err := textLineJoinType(run.LineJoinType)
	if err != nil {
		return fmt.Errorf("run %d line_join_type: %w", index, err)
	}
	if err := layer.SetRunLineJoinType(index, lineJoin); err != nil {
		return fmt.Errorf("run %d line_join_type: %w", index, err)
	}
	digitSet, err := textDigitSet(run.DigitSet)
	if err != nil {
		return fmt.Errorf("run %d digit_set: %w", index, err)
	}
	if err := layer.SetRunDigitSet(index, digitSet); err != nil {
		return fmt.Errorf("run %d digit_set: %w", index, err)
	}
	if err := layer.SetRunNoBreak(index, run.NoBreak); err != nil {
		return fmt.Errorf("run %d no_break: %w", index, err)
	}
	if err := layer.SetRunFauxBold(index, run.FauxBold); err != nil {
		return fmt.Errorf("run %d faux_bold: %w", index, err)
	}
	if err := layer.SetRunFauxItalic(index, run.FauxItalic); err != nil {
		return fmt.Errorf("run %d faux_italic: %w", index, err)
	}
	return nil
}
