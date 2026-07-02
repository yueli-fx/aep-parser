package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeTextParagraphs(layer *aep.Layer, paragraphs []profile.TextParagraph) error {
	for i, paragraph := range paragraphs {
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
