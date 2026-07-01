package aepmigrate

import (
	"fmt"
	"strings"
)

type SourceVersion struct {
	Label VersionLabel
	Raw   string
}

func ParseVersionLabel(value string) (VersionLabel, error) {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "AE2020", "2020":
		return VersionAE2020, nil
	case "AE2021", "2021":
		return VersionAE2021, nil
	case "AE2022", "2022":
		return VersionAE2022, nil
	case "AE2023", "2023":
		return VersionAE2023, nil
	case "AE2024", "2024":
		return VersionAE2024, nil
	case "AE2025", "2025":
		return VersionAE2025, nil
	default:
		return "", fmt.Errorf("target version must be AE2020, AE2021, AE2022, AE2023, AE2024, or AE2025")
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
