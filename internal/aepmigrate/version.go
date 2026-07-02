package aepmigrate

import (
	"strings"

	"github.com/yueli-fx/aep-parser/internal/aeversion"
)

func SupportedVersionLabels() []VersionLabel {
	labels := aeversion.SupportedLabels()
	out := make([]VersionLabel, 0, len(labels))
	for _, label := range labels {
		out = append(out, VersionLabel(label))
	}
	return out
}

func SupportedVersionStrings() []string {
	return aeversion.SupportedLabels()
}

func IsSupportedVersionLabel(label VersionLabel) bool {
	return aeversion.IsSupported(string(label))
}

func TargetVersionHelp() string {
	return aeversion.TargetHelp()
}

type SourceVersion struct {
	Label VersionLabel
	Raw   string
}

func ParseVersionLabel(value string) (VersionLabel, error) {
	label, err := aeversion.ParseLabel(value)
	if err != nil {
		return "", err
	}
	return VersionLabel(label), nil
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
