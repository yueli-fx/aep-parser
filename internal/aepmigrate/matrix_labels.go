package aepmigrate

import "strings"

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
