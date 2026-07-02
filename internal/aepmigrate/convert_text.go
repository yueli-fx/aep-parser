package aepmigrate

import (
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeTextStyle(layer *aep.Layer, source *profile.TextSource) error {
	if source == nil {
		return nil
	}
	if err := materializeTextRuns(layer, source.Runs); err != nil {
		return err
	}
	return materializeTextParagraphs(layer, source.Paragraphs)
}
