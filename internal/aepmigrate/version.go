package aepmigrate

import (
	"fmt"
	"strings"
)

var supportedVersionLabels = []VersionLabel{
	VersionAE2020,
	VersionAE2021,
	VersionAE2022,
	VersionAE2023,
	VersionAE2024,
	VersionAE2025,
}

func SupportedVersionLabels() []VersionLabel {
	return append([]VersionLabel(nil), supportedVersionLabels...)
}

func SupportedVersionStrings() []string {
	labels := SupportedVersionLabels()
	out := make([]string, 0, len(labels))
	for _, label := range labels {
		out = append(out, string(label))
	}
	return out
}

func IsSupportedVersionLabel(label VersionLabel) bool {
	for _, supported := range supportedVersionLabels {
		if label == supported {
			return true
		}
	}
	return false
}

func TargetVersionHelp() string {
	labels := SupportedVersionStrings()
	if len(labels) == 0 {
		return "target AE version"
	}
	return fmt.Sprintf("target AE version: %s through %s", labels[0], labels[len(labels)-1])
}

type SourceVersion struct {
	Label VersionLabel
	Raw   string
}

func ParseVersionLabel(value string) (VersionLabel, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	for _, label := range supportedVersionLabels {
		labelText := string(label)
		if normalized == labelText || normalized == strings.TrimPrefix(labelText, "AE") {
			return label, nil
		}
	}
	return "", fmt.Errorf("target version must be %s", supportedVersionListForError())
}

func supportedVersionListForError() string {
	labels := SupportedVersionStrings()
	switch len(labels) {
	case 0:
		return "a supported AE version"
	case 1:
		return labels[0]
	case 2:
		return labels[0] + " or " + labels[1]
	default:
		return strings.Join(labels[:len(labels)-1], ", ") + ", or " + labels[len(labels)-1]
	}
}

func NormalizeSourceVersion(raw string) SourceVersion {
	trimmed := strings.TrimSpace(raw)
	upper := strings.ToUpper(trimmed)
	switch {
	case strings.Contains(upper, "2025") || strings.HasPrefix(upper, "25."):
		return SourceVersion{Label: VersionAE2025, Raw: trimmed}
	case strings.Contains(upper, "2024") || strings.HasPrefix(upper, "24."):
		return SourceVersion{Label: VersionAE2024, Raw: trimmed}
	case strings.Contains(upper, "2023") || strings.HasPrefix(upper, "23."):
		return SourceVersion{Label: VersionAE2023, Raw: trimmed}
	case strings.Contains(upper, "2022") || strings.HasPrefix(upper, "22."):
		return SourceVersion{Label: VersionAE2022, Raw: trimmed}
	case strings.Contains(upper, "2021") || strings.HasPrefix(upper, "18."):
		return SourceVersion{Label: VersionAE2021, Raw: trimmed}
	case strings.Contains(upper, "2020") || strings.HasPrefix(upper, "17."):
		return SourceVersion{Label: VersionAE2020, Raw: trimmed}
	default:
		return SourceVersion{Label: VersionUnknown, Raw: trimmed}
	}
}
