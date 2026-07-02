package aeversion

import "testing"

func TestSupportedLabelsAreCopied(t *testing.T) {
	labels := SupportedLabels()
	if len(labels) != 6 || labels[0] != AE2020 || labels[len(labels)-1] != AE2025 {
		t.Fatalf("SupportedLabels = %v, want AE2020..AE2025", labels)
	}
	labels[0] = "AE1999"
	if got := SupportedLabels()[0]; got != AE2020 {
		t.Fatalf("SupportedLabels returned mutable backing slice, first = %q", got)
	}
}

func TestParseLabelAcceptsSupportedTargets(t *testing.T) {
	tests := map[string]string{
		"AE2020": AE2020,
		"2020":   AE2020,
		"ae2021": AE2021,
		"2021":   AE2021,
		"ae2022": AE2022,
		"2022":   AE2022,
		"AE2023": AE2023,
		"2023":   AE2023,
		"AE2024": AE2024,
		"2024":   AE2024,
		"AE2025": AE2025,
		"2025":   AE2025,
	}
	for input, want := range tests {
		got, err := ParseLabel(input)
		if err != nil {
			t.Fatalf("ParseLabel(%q): %v", input, err)
		}
		if got != want {
			t.Fatalf("ParseLabel(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestParseLabelRejectsUnsupportedTarget(t *testing.T) {
	if _, err := ParseLabel("AE2019"); err == nil {
		t.Fatal("ParseLabel(AE2019) succeeded, want error")
	}
}

func TestHelpTextUsesSupportedRange(t *testing.T) {
	if got := SupportedRange(); got != "AE2020-AE2025" {
		t.Fatalf("SupportedRange = %q", got)
	}
	if got := TargetHelp(); got != "target AE version: AE2020 through AE2025" {
		t.Fatalf("TargetHelp = %q", got)
	}
}
