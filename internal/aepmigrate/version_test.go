package aepmigrate

import "testing"

func TestParseVersionLabelAcceptsSupportedTargets(t *testing.T) {
	tests := map[string]VersionLabel{
		"AE2020": VersionAE2020,
		"2020":   VersionAE2020,
		"ae2021": VersionAE2021,
		"2021":   VersionAE2021,
		"ae2022": VersionAE2022,
		"2022":   VersionAE2022,
		"AE2023": VersionAE2023,
		"2023":   VersionAE2023,
		"AE2024": VersionAE2024,
		"2024":   VersionAE2024,
		"AE2025": VersionAE2025,
		"2025":   VersionAE2025,
	}
	for input, want := range tests {
		got, err := ParseVersionLabel(input)
		if err != nil {
			t.Fatalf("ParseVersionLabel(%q): %v", input, err)
		}
		if got != want {
			t.Fatalf("ParseVersionLabel(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestParseVersionLabelRejectsUnsupportedTarget(t *testing.T) {
	if _, err := ParseVersionLabel("AE2019"); err == nil {
		t.Fatal("ParseVersionLabel(AE2019) succeeded, want error")
	}
}

func TestSupportedVersionLabelsAreCopied(t *testing.T) {
	labels := SupportedVersionLabels()
	if len(labels) != 6 || labels[0] != VersionAE2020 || labels[len(labels)-1] != VersionAE2025 {
		t.Fatalf("supported labels = %v, want AE2020..AE2025", labels)
	}
	labels[0] = VersionUnknown
	if got := SupportedVersionLabels()[0]; got != VersionAE2020 {
		t.Fatalf("SupportedVersionLabels returned mutable backing slice, first = %q", got)
	}
}

func TestTargetVersionHelpUsesSupportedRange(t *testing.T) {
	if got := TargetVersionHelp(); got != "target AE version: AE2020 through AE2025" {
		t.Fatalf("TargetVersionHelp = %q", got)
	}
}

func TestNormalizeSourceVersionKeepsUnknownHonest(t *testing.T) {
	if got := NormalizeSourceVersion(""); got.Label != VersionUnknown {
		t.Fatalf("empty source label = %q, want unknown", got.Label)
	}
}
