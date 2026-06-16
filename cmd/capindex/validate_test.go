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
		{"stable+roundtrip ok (decoupled: stable API, no AE gate)", Cap{Domain: "layer-set", Tier: "stable", Verify: "roundtrip"}, true},
		{"stable+none rejected", Cap{Domain: "layer-set", Tier: "stable", Verify: "none"}, false},
		{"alpha+roundtrip ok", Cap{Domain: "layer-set", Tier: "alpha", Verify: "roundtrip"}, true},
		{"alpha+none rejected", Cap{Domain: "layer-set", Tier: "alpha", Verify: "none"}, false},
		{"planned+gate rejected", Cap{Domain: "shape", Tier: "planned", Verify: "none", Gate: []string{"T"}}, false},
		{"planned clean ok", Cap{Domain: "shape", Tier: "planned", Verify: "none"}, true},
		{"ae-accept no gate rejected", Cap{Domain: "effect", Tier: "stable", Verify: "ae-accept"}, false},
		{"missing required field", Cap{Tier: "stable", Verify: "ae-accept", Gate: []string{"T"}}, false},
		{"bad tier", Cap{Domain: "shape", Tier: "bogus", Verify: "none"}, false},
		{"bad domain", Cap{Domain: "bogus", Tier: "alpha", Verify: "roundtrip"}, false},
		{"meta+none ok", Cap{Domain: "meta", Tier: "stable", Verify: "none"}, true},
		{"meta+roundtrip ok", Cap{Domain: "meta", Tier: "alpha", Verify: "roundtrip"}, true},
		{"meta+ae-accept rejected", Cap{Domain: "meta", Tier: "stable", Verify: "ae-accept", Gate: []string{"T"}}, false},
		{"meta+gate rejected", Cap{Domain: "meta", Tier: "stable", Verify: "none", Gate: []string{"T"}}, false},
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
	c := Cap{Domain: "layer-set", Tier: "alpha", Verify: "roundtrip"}
	if err := validateCap(&c); err != nil {
		t.Fatal(err)
	}
	if c.MinVer != "2020" {
		t.Errorf("minver default: %q", c.MinVer)
	}
}
