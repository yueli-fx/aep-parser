package aepmigrate

import "testing"

func TestParseVersionLabelAcceptsSupportedTargets(t *testing.T) {
	tests := map[string]VersionLabel{
		"AE2020": VersionAE2020,
		"2020":   VersionAE2020,
		"ae2022": VersionAE2022,
		"2022":   VersionAE2022,
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

func TestNormalizeSourceVersionKeepsUnknownHonest(t *testing.T) {
	if got := NormalizeSourceVersion(""); got.Label != VersionUnknown {
		t.Fatalf("empty source label = %q, want unknown", got.Label)
	}
}
