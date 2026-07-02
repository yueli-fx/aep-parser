package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeTextStyle(layer *aep.Layer, source *profile.TextSource) error {
	if source == nil {
		return nil
	}
	for i, run := range source.Runs {
		if err := layer.SetRunFontSize(i, run.FontSize); err != nil {
			return fmt.Errorf("run %d font_size: %w", i, err)
		}
		if err := layer.SetRunFillColor(i, run.FillColor); err != nil {
			return fmt.Errorf("run %d fill_color: %w", i, err)
		}
		if err := layer.SetRunAutoLeading(i, run.AutoLeading); err != nil {
			return fmt.Errorf("run %d auto_leading: %w", i, err)
		}
		if err := layer.SetRunLeading(i, run.Leading); err != nil {
			return fmt.Errorf("run %d leading: %w", i, err)
		}
		if err := layer.SetRunTracking(i, run.Tracking); err != nil {
			return fmt.Errorf("run %d tracking: %w", i, err)
		}
		if err := layer.SetRunBaselineShift(i, run.BaselineShift); err != nil {
			return fmt.Errorf("run %d baseline_shift: %w", i, err)
		}
		if err := layer.SetRunHorizontalScale(i, run.HorizontalScale); err != nil {
			return fmt.Errorf("run %d horizontal_scale: %w", i, err)
		}
		if err := layer.SetRunVerticalScale(i, run.VerticalScale); err != nil {
			return fmt.Errorf("run %d vertical_scale: %w", i, err)
		}
		if err := layer.SetRunTsume(i, run.Tsume); err != nil {
			return fmt.Errorf("run %d tsume: %w", i, err)
		}
		caps, err := textCapsOption(run.CapsOption)
		if err != nil {
			return fmt.Errorf("run %d caps_option: %w", i, err)
		}
		if err := layer.SetRunCapsOption(i, caps); err != nil {
			return fmt.Errorf("run %d caps_option: %w", i, err)
		}
		baseline, err := textBaselineOption(run.BaselineOption)
		if err != nil {
			return fmt.Errorf("run %d baseline_option: %w", i, err)
		}
		if err := layer.SetRunBaselineOption(i, baseline); err != nil {
			return fmt.Errorf("run %d baseline_option: %w", i, err)
		}
		kern, err := textAutoKernType(run.AutoKernType)
		if err != nil {
			return fmt.Errorf("run %d auto_kern_type: %w", i, err)
		}
		if err := layer.SetRunAutoKernType(i, kern); err != nil {
			return fmt.Errorf("run %d auto_kern_type: %w", i, err)
		}
		lineJoin, err := textLineJoinType(run.LineJoinType)
		if err != nil {
			return fmt.Errorf("run %d line_join_type: %w", i, err)
		}
		if err := layer.SetRunLineJoinType(i, lineJoin); err != nil {
			return fmt.Errorf("run %d line_join_type: %w", i, err)
		}
		digitSet, err := textDigitSet(run.DigitSet)
		if err != nil {
			return fmt.Errorf("run %d digit_set: %w", i, err)
		}
		if err := layer.SetRunDigitSet(i, digitSet); err != nil {
			return fmt.Errorf("run %d digit_set: %w", i, err)
		}
		if err := layer.SetRunNoBreak(i, run.NoBreak); err != nil {
			return fmt.Errorf("run %d no_break: %w", i, err)
		}
		if err := layer.SetRunFauxBold(i, run.FauxBold); err != nil {
			return fmt.Errorf("run %d faux_bold: %w", i, err)
		}
		if err := layer.SetRunFauxItalic(i, run.FauxItalic); err != nil {
			return fmt.Errorf("run %d faux_italic: %w", i, err)
		}
		if err := layer.SetRunApplyStroke(i, run.ApplyStroke); err != nil {
			return fmt.Errorf("run %d apply_stroke: %w", i, err)
		}
		if err := layer.SetRunStrokeColor(i, run.StrokeColor); err != nil {
			return fmt.Errorf("run %d stroke_color: %w", i, err)
		}
		if err := layer.SetRunStrokeWidth(i, run.StrokeWidth); err != nil {
			return fmt.Errorf("run %d stroke_width: %w", i, err)
		}
		if err := layer.SetRunStrokeOverFill(i, run.StrokeOverFill); err != nil {
			return fmt.Errorf("run %d stroke_over_fill: %w", i, err)
		}
	}
	for i, paragraph := range source.Paragraphs {
		justification, err := textJustification(paragraph.Justification)
		if err != nil {
			return fmt.Errorf("paragraph %d justification: %w", i, err)
		}
		if err := layer.SetParagraphJustification(i, justification); err != nil {
			return fmt.Errorf("paragraph %d justification: %w", i, err)
		}
		if err := layer.SetParagraphFirstLineIndent(i, paragraph.FirstLineIndent); err != nil {
			return fmt.Errorf("paragraph %d first_line_indent: %w", i, err)
		}
		if err := layer.SetParagraphStartIndent(i, paragraph.StartIndent); err != nil {
			return fmt.Errorf("paragraph %d start_indent: %w", i, err)
		}
		if err := layer.SetParagraphEndIndent(i, paragraph.EndIndent); err != nil {
			return fmt.Errorf("paragraph %d end_indent: %w", i, err)
		}
		if err := layer.SetParagraphSpaceBefore(i, paragraph.SpaceBefore); err != nil {
			return fmt.Errorf("paragraph %d space_before: %w", i, err)
		}
		if err := layer.SetParagraphSpaceAfter(i, paragraph.SpaceAfter); err != nil {
			return fmt.Errorf("paragraph %d space_after: %w", i, err)
		}
		if err := layer.SetParagraphAutoHyphenate(i, paragraph.AutoHyphenate); err != nil {
			return fmt.Errorf("paragraph %d auto_hyphenate: %w", i, err)
		}
		leadingType, err := textLeadingType(paragraph.LeadingType)
		if err != nil {
			return fmt.Errorf("paragraph %d leading_type: %w", i, err)
		}
		if err := layer.SetParagraphLeadingType(i, leadingType); err != nil {
			return fmt.Errorf("paragraph %d leading_type: %w", i, err)
		}
		if err := layer.SetParagraphHangingRoman(i, paragraph.HangingRoman); err != nil {
			return fmt.Errorf("paragraph %d hanging_roman: %w", i, err)
		}
		direction, err := textParagraphDirection(paragraph.Direction)
		if err != nil {
			return fmt.Errorf("paragraph %d direction: %w", i, err)
		}
		if err := layer.SetParagraphDirection(i, direction); err != nil {
			return fmt.Errorf("paragraph %d direction: %w", i, err)
		}
	}
	return nil
}
