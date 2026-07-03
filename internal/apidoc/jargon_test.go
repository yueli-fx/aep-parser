package apidoc

import "testing"

func TestLintJargon_Hits(t *testing.T) {
	cases := []string{
		"This was the V2.2 builder transplant",
		"see the V2.2 internal plan for the rule",
		"derived in tmp_debug/foo",
		"RE'd from tolerance.aep",
		"the wave 2 ldta pass",
		"per the flightdeck cockpit",
		"py-aep golden disagrees",
	}
	for _, c := range cases {
		if v := LintJargon(c); len(v) == 0 {
			t.Errorf("LintJargon(%q) = clean, want a hit", c)
		}
	}
}

func TestLintJargon_Clean(t *testing.T) {
	cases := []string{
		"Appends a closed Bezier mask to the Mask Parade.",
		"AE 2020 and AE 2025 both accept this.",          // real user-facing versions stay
		"the tdb4 keyframe layout differs by 3 offsets", // chunk IDs are domain, not jargon
	}
	for _, c := range cases {
		if v := LintJargon(c); len(v) != 0 {
			t.Errorf("LintJargon(%q) = %+v, want clean", c, v)
		}
	}
}

func TestLintJargon_NolintEscape(t *testing.T) {
	if v := LintJargon("compatible back to V2.2 //nolint:jargon"); len(v) != 0 {
		t.Errorf("nolint line should be exempt, got %+v", v)
	}
}
