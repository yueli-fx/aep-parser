package serializer

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"testing"

	"github.com/example/aep-parser/internal/rifx"
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

func be32(d []byte, off int) uint32 { return binary.BigEndian.Uint32(d[off:]) }

// TestSynthControlEntries_PardLayout asserts the synthesized pard bytes for the
// newly-supported control kinds match the field values RE'd from AE's own
// Pseudo Effect Maker output (test_data/pseudo_rich_demo.aep): dropdown 0009,
// group-start 0011, label 0004 + group-end 0005.
func TestSynthControlEntries_PardLayout(t *testing.T) {
	// Dropdown: 2 options, default selection 2 — golden 0009 had @0x38=2,
	// @0x3C=0x00020002 (hi16 count=2, lo16 sel=2) + trailing pdnm "a|b".
	t.Run("dropdown", func(t *testing.T) {
		es, err := synthControlEntries(PseudoControl{Kind: PseudoDropdown, Name: "Menu", Options: []string{"a", "b"}, Default: 2})
		if err != nil {
			t.Fatal(err)
		}
		if len(es) != 1 {
			t.Fatalf("dropdown: got %d entries, want 1", len(es))
		}
		d := es[0].pard.Data
		if d[0x0F] != 0x07 {
			t.Errorf("control_type @0x0F = %#x, want 0x07", d[0x0F])
		}
		if got := be32(d, 0x38); got != 2 {
			t.Errorf("@0x38 sel = %#x, want 0x2", got)
		}
		if got := be32(d, 0x3C); got != 0x00020002 {
			t.Errorf("@0x3C = %#x, want 0x00020002", got)
		}
		if len(es[0].trailing) != 1 || es[0].trailing[0].ID != rifx.IDPdnm {
			t.Fatalf("dropdown: want one trailing pdnm")
		}
		if want := utf8StringData("a|b"); !bytes.Equal(es[0].trailing[0].Data, want) {
			t.Errorf("dropdown pdnm = %x, want %x (Utf8 a|b)", es[0].trailing[0].Data, want)
		}
	})

	// Group start: 0x0d, label flag (@0x04) clear, @0x30=2 — golden 0011.
	t.Run("group-start", func(t *testing.T) {
		es, _ := synthControlEntries(PseudoControl{Kind: PseudoGroupStart, Name: "Grp"})
		d := es[0].pard.Data
		if d[0x0F] != 0x0d || be32(d, 0x04) != 0 || be32(d, 0x30) != 2 {
			t.Errorf("group-start pard @0x0F/@0x04/@0x30 = %#x/%#x/%#x, want 0x0d/0x0/0x2", d[0x0F], be32(d, 0x04), be32(d, 0x30))
		}
	})

	// Group end: 0x0e, @0x04=0x08, @0x30=2 — golden 0013/0005.
	t.Run("group-end", func(t *testing.T) {
		es, _ := synthControlEntries(PseudoControl{Kind: PseudoGroupEnd})
		d := es[0].pard.Data
		if d[0x0F] != 0x0e || be32(d, 0x04) != 0x08 || be32(d, 0x30) != 2 {
			t.Errorf("group-end pard @0x0F/@0x04/@0x30 = %#x/%#x/%#x, want 0x0e/0x08/0x2", d[0x0F], be32(d, 0x04), be32(d, 0x30))
		}
	})

	// Label: a group-start with the label flag (@0x04=0x20) immediately closed
	// by a generated group-end — golden 0004 (label "标签") + 0005 (group-end).
	t.Run("label", func(t *testing.T) {
		es, _ := synthControlEntries(PseudoControl{Kind: PseudoLabel, Name: "Note"})
		if len(es) != 2 {
			t.Fatalf("label: got %d entries, want 2 (start+end)", len(es))
		}
		start := es[0].pard.Data
		if start[0x0F] != 0x0d || be32(start, 0x04) != 0x20 || be32(start, 0x30) != 2 {
			t.Errorf("label start @0x0F/@0x04/@0x30 = %#x/%#x/%#x, want 0x0d/0x20/0x2", start[0x0F], be32(start, 0x04), be32(start, 0x30))
		}
		end := es[1].pard.Data
		if end[0x0F] != 0x0e || be32(end, 0x04) != 0x08 {
			t.Errorf("label end @0x0F/@0x04 = %#x/%#x, want 0x0e/0x08", end[0x0F], be32(end, 0x04))
		}
	})
}
