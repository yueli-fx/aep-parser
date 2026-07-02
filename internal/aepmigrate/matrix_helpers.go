package aepmigrate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/recipe"
)

func matrixAEOpenLabels(labels []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, label := range labels {
		if strings.EqualFold(strings.TrimSpace(label), "all") {
			for _, expanded := range matrixAEHostLabels {
				if !seen[expanded] {
					seen[expanded] = true
					out = append(out, expanded)
				}
			}
			continue
		}
		normalized := normalizeAEHostLabel(label)
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		out = append(out, normalized)
	}
	return out
}

func expandMatrixSourceLabels(labels []string) []string {
	return expandMatrixWriterLabels(labels)
}

func expandMatrixTargetLabels(labels []string) []string {
	return expandMatrixWriterLabels(labels)
}

func expandMatrixWriterLabels(labels []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, label := range labels {
		if strings.EqualFold(strings.TrimSpace(label), "all") {
			for _, expanded := range matrixWriterLabels {
				if !seen[expanded] {
					seen[expanded] = true
					out = append(out, expanded)
				}
			}
			continue
		}
		normalized := normalizeAEHostLabel(label)
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		out = append(out, normalized)
	}
	return out
}

func normalizeAEHostLabel(label string) string {
	label = strings.ToUpper(strings.TrimSpace(label))
	if label == "" {
		return ""
	}
	if strings.HasPrefix(label, "AE") {
		return label
	}
	if len(label) == 4 {
		for _, r := range label {
			if r < '0' || r > '9' {
				return label
			}
		}
		return "AE" + label
	}
	return label
}

func matrixSourceWriterLabel(label string) (string, bool) {
	if strings.EqualFold(strings.TrimSpace(label), "recipe") {
		return "", true
	}
	target, ok := matrixWriterTarget(label)
	if !ok {
		return "", false
	}
	return string(target), true
}

func matrixWriterTarget(label string) (VersionLabel, bool) {
	target, err := ParseVersionLabel(label)
	if err != nil {
		return "", false
	}
	switch target {
	case VersionAE2020, VersionAE2021, VersionAE2022, VersionAE2023, VersionAE2024, VersionAE2025:
		return target, true
	default:
		return "", false
	}
}

func readMatrixRecipe(path string) (recipe.Recipe, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return recipe.Recipe{}, err
	}
	var rec recipe.Recipe
	if err := json.Unmarshal(data, &rec); err != nil {
		return recipe.Recipe{}, err
	}
	return rec, nil
}

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

func writeMatrixReport(path string, report MatrixReport) error {
	return writeMatrixJSON(path, report)
}

func writeMatrixJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
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
