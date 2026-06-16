package main

import "testing"

func TestValidateCap(t *testing.T) {
	cases := []struct {
		name string
		cap  Cap
		ok   bool
	}{
		{"stable+render ok", Cap{Domain: "effect", Tier: "stable", Verify: "render-pixel", Gate: []string{"T"}}, true},
		{"stable+ae-accept ok", Cap{Domain: "layer-create", Tier: "stable", Verify: "ae-accept", Gate: []string{"T"}}, true},
		{"stable+roundtrip rejected", Cap{Domain: "x", Tier: "stable", Verify: "roundtrip", Gate: []string{"T"}}, false},
		{"alpha+roundtrip ok", Cap{Domain: "x", Tier: "alpha", Verify: "roundtrip"}, true},
		{"alpha+none rejected", Cap{Domain: "x", Tier: "alpha", Verify: "none"}, false},
		{"planned+gate rejected", Cap{Domain: "x", Tier: "planned", Verify: "none", Gate: []string{"T"}}, false},
		{"planned clean ok", Cap{Domain: "x", Tier: "planned", Verify: "none"}, true},
		{"ae-accept no gate rejected", Cap{Domain: "x", Tier: "stable", Verify: "ae-accept"}, false},
		{"missing required field", Cap{Tier: "stable", Verify: "ae-accept", Gate: []string{"T"}}, false},
		{"bad tier", Cap{Domain: "x", Tier: "bogus", Verify: "none"}, false},
	}
	for _, tc := range cases {
		c := tc.cap
		err := validateCap(&c)
		if tc.ok && err != nil {
			t.Errorf("%s: expected ok, got %v", tc.name, err)
		}
		if !tc.ok && err == nil {
			t.Errorf("%s: expected error, got nil", tc.name)
		}
	}
}

func TestValidateCap_MinVerDefault(t *testing.T) {
	c := Cap{Domain: "x", Tier: "alpha", Verify: "roundtrip"}
	if err := validateCap(&c); err != nil {
		t.Fatal(err)
	}
	if c.MinVer != "2020" {
		t.Errorf("minver default: %q", c.MinVer)
	}
}
