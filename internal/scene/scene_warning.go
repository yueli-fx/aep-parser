package scene

// ParseWarning is the structured form of Project.Warnings.
//
// Warnings remains the human-readable compatibility view and is still used by
// mutation rollback checks. ParseWarnings carries machine-readable context for
// debugging, diagnostics, and future UI/reporting.
type ParseWarning struct {
	Chunk   string `json:"chunk,omitempty"`
	Offset  int64  `json:"offset,omitempty"`
	Message string `json:"message"`
}

// warningLengths snapshots both warning views for atomic mutation rollback.
func (p *Project) warningLengths() (warningsLen, parseWarningsLen int) {
	if p == nil {
		return 0, 0
	}
	return len(p.Warnings), len(p.ParseWarnings)
}

// recordWarning appends a message-only warning to both warning views.
func (p *Project) recordWarning(message string) {
	p.recordParseWarning(ParseWarning{Message: message})
}

// recordParseWarning appends a structured warning and keeps Warnings in sync.
func (p *Project) recordParseWarning(w ParseWarning) {
	if p == nil {
		return
	}
	p.Warnings = append(p.Warnings, w.Message)
	p.ParseWarnings = append(p.ParseWarnings, w)
}

// rollbackWarnings truncates both warning views to a previously captured pair.
func (p *Project) rollbackWarnings(warningsLen, parseWarningsLen int) {
	if p == nil {
		return
	}
	if warningsLen < 0 {
		warningsLen = 0
	}
	if parseWarningsLen < 0 {
		parseWarningsLen = 0
	}
	if warningsLen < len(p.Warnings) {
		p.Warnings = p.Warnings[:warningsLen]
	}
	if parseWarningsLen < len(p.ParseWarnings) {
		p.ParseWarnings = p.ParseWarnings[:parseWarningsLen]
	}
}
