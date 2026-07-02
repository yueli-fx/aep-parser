package aepmigrate

import "strings"

func mergeHostMaps(a, b map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v != "" {
			out[k] = v
		}
	}
	return out
}

func summarizeMatrix(cases []MatrixCase) MatrixSummary {
	s := MatrixSummary{Total: len(cases)}
	for _, c := range cases {
		switch c.Status {
		case MatrixStatusPass:
			s.Passed++
		case MatrixStatusBlocked:
			s.Blocked++
		case MatrixStatusFailed:
			s.Failed++
		case MatrixStatusSkipped:
			s.Skipped++
		}
	}
	return s
}

func sanitizeMatrixName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unnamed"
	}
	var b strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_' || r == '.':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return b.String()
}
