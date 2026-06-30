package serializer

import "github.com/yueli-fx/aep-parser/internal/scene"

type parseWarningSink struct {
	warnings      *[]string
	parseWarnings *[]scene.ParseWarning
}

func stringWarningSink(warnings *[]string) *parseWarningSink {
	if warnings == nil {
		return nil
	}
	return &parseWarningSink{warnings: warnings}
}

func projectWarningSink(p *Project) *parseWarningSink {
	if p == nil {
		return nil
	}
	return &parseWarningSink{warnings: &p.Warnings, parseWarnings: &p.ParseWarnings}
}

func (s *parseWarningSink) record(w scene.ParseWarning) {
	if s == nil {
		return
	}
	if s.warnings != nil {
		*s.warnings = append(*s.warnings, w.Message)
	}
	if s.parseWarnings != nil {
		*s.parseWarnings = append(*s.parseWarnings, w)
	}
}
