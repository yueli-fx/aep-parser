package aeversion

import (
	"fmt"
	"strings"
)

const (
	AE2020 = "AE2020"
	AE2021 = "AE2021"
	AE2022 = "AE2022"
	AE2023 = "AE2023"
	AE2024 = "AE2024"
	AE2025 = "AE2025"
)

var supportedLabels = []string{
	AE2020,
	AE2021,
	AE2022,
	AE2023,
	AE2024,
	AE2025,
}

func SupportedLabels() []string {
	return append([]string(nil), supportedLabels...)
}

func IsSupported(label string) bool {
	for _, supported := range supportedLabels {
		if label == supported {
			return true
		}
	}
	return false
}

func ParseLabel(value string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	for _, label := range supportedLabels {
		if normalized == label || normalized == strings.TrimPrefix(label, "AE") {
			return label, nil
		}
	}
	return "", fmt.Errorf("target version must be %s", SupportedListForError())
}

func SupportedRange() string {
	labels := SupportedLabels()
	if len(labels) == 0 {
		return "supported AE versions"
	}
	if len(labels) == 1 {
		return labels[0]
	}
	return labels[0] + "-" + labels[len(labels)-1]
}

func TargetHelp() string {
	labels := SupportedLabels()
	if len(labels) == 0 {
		return "target AE version"
	}
	return fmt.Sprintf("target AE version: %s through %s", labels[0], labels[len(labels)-1])
}

func SupportedListForError() string {
	labels := SupportedLabels()
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
