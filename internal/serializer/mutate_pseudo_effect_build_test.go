package serializer

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// TestPardNameBytes_GBKMatchesAENative is the byte-equivalence proof for CJK
// control labels. AE reads a pseudo control's pard @0x10 name in the system ANSI
// codepage, and AE's own Pseudo Effect Maker writes it GBK-encoded. We assert
// pardNameBytes produces the exact bytes AE wrote in test_data/pseudo_rich_demo.aep
// (color pard @0x10 = d1d5c9ab = "颜色"; angle pard @0x10 = bdc7b6c8 = "角度"),
// so a CJK label displays correctly on a simplified-Chinese system. This can't
// be ship-gated on a Western-codepage machine — byte-equivalence to AE-authored
// output is the available evidence. See incidents/pseudo-control-label-ansi-codepage.md.
func TestPardNameBytes_GBKMatchesAENative(t *testing.T) {
	cases := []struct {
		name    string
		wantHex string // verbatim from pseudo_rich_demo.aep pard @0x10
	}{
		{"颜色", "d1d5c9ab"},     // AE color pard 0003
		{"角度", "bdc7b6c8"},     // AE angle pard 0001
		{"Strength", "537472656e677468"}, // ASCII passthrough
	}
	for _, tc := range cases {
		got := pardNameBytes(tc.name)
		want, _ := hex.DecodeString(tc.wantHex)
		if !bytes.Equal(got, want) {
			t.Errorf("pardNameBytes(%q) = %x, want %s (AE-native bytes)", tc.name, got, tc.wantHex)
		}
	}
}
