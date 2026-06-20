package serializer

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"math"
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
		got := pardNameBytes(tc.name, PseudoLabelGBK)
		want, _ := hex.DecodeString(tc.wantHex)
		if !bytes.Equal(got, want) {
			t.Errorf("pardNameBytes(%q) = %x, want %s (AE-native bytes)", tc.name, got, tc.wantHex)
		}
	}
}

// TestPardNameBytes_ShiftJIS proves Japanese labels are Shift-JIS encoded (the
// codepage a Japanese Windows decodes the pard name in). Can't be ship-gated on
// a Western-codepage machine — byte-equivalence to the standard Shift-JIS
// encoding is the evidence (色 = 0x9046, ア = 0x8341 in Shift-JIS/cp932).
func TestPardNameBytes_ShiftJIS(t *testing.T) {
	cases := []struct {
		name    string
		wantHex string
	}{
		{"色", "9046"},             // kanji
		{"ア", "8341"},             // katakana
		{"Color", "436f6c6f72"}, // ASCII passthrough (unchanged across codepages)
	}
	for _, tc := range cases {
		got := pardNameBytes(tc.name, PseudoLabelShiftJIS)
		want, _ := hex.DecodeString(tc.wantHex)
		if !bytes.Equal(got, want) {
			t.Errorf("pardNameBytes(%q, ShiftJIS) = %x, want %s", tc.name, got, tc.wantHex)
		}
	}
	// The same label under GBK differs — proving the setting actually switches
	// encoders (not a no-op).
	if bytes.Equal(pardNameBytes("色", PseudoLabelGBK), pardNameBytes("色", PseudoLabelShiftJIS)) {
		t.Error("色 should encode to different bytes under GBK vs Shift-JIS")
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
		es, err := synthControlEntries(PseudoControl{Kind: PseudoDropdown, Name: "Menu", Options: []string{"a", "b"}, Default: 2}, PseudoLabelGBK)
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
		es, _ := synthControlEntries(PseudoControl{Kind: PseudoGroupStart, Name: "Grp"}, PseudoLabelGBK)
		d := es[0].pard.Data
		if d[0x0F] != 0x0d || be32(d, 0x04) != 0 || be32(d, 0x30) != 2 {
			t.Errorf("group-start pard @0x0F/@0x04/@0x30 = %#x/%#x/%#x, want 0x0d/0x0/0x2", d[0x0F], be32(d, 0x04), be32(d, 0x30))
		}
	})

	// Group end: 0x0e, @0x04=0x08, @0x30=2 — golden 0013/0005.
	t.Run("group-end", func(t *testing.T) {
		es, _ := synthControlEntries(PseudoControl{Kind: PseudoGroupEnd}, PseudoLabelGBK)
		d := es[0].pard.Data
		if d[0x0F] != 0x0e || be32(d, 0x04) != 0x08 || be32(d, 0x30) != 2 {
			t.Errorf("group-end pard @0x0F/@0x04/@0x30 = %#x/%#x/%#x, want 0x0e/0x08/0x2", d[0x0F], be32(d, 0x04), be32(d, 0x30))
		}
	})

	// Label: a group-start with the label flag (@0x04=0x20) immediately closed
	// by a generated group-end — golden 0004 (label "标签") + 0005 (group-end).
	t.Run("label", func(t *testing.T) {
		es, _ := synthControlEntries(PseudoControl{Kind: PseudoLabel, Name: "Note"}, PseudoLabelGBK)
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

func f64at(d []byte, off int) float64 {
	return math.Float64frombits(binary.BigEndian.Uint64(d[off:]))
}

// TestSynthControlValueEntry verifies the value-entry synthesis for the controls
// whose value can't be elided into the pard: Point/3DPoint custom coords (cdat =
// fraction of coord space, AE-verified [0.25,0.125]→[100,50] in a 400 comp) and
// the Layer picker (binding in tdpi). Bytes mirror test_data/pseudo_rich_demo.aep.
func TestSynthControlValueEntry(t *testing.T) {
	tdbsOf := func(chunks []*rifx.Chunk) *rifx.Chunk {
		if len(chunks) != 2 || chunks[0].ID != rifx.IDTdmn || chunks[1].FormType != rifx.IDTdbs {
			t.Fatalf("want [tdmn, LIST tdbs], got %d chunks", len(chunks))
		}
		return chunks[1]
	}
	child := func(c *rifx.Chunk, id rifx.ChunkID) *rifx.Chunk {
		for _, ch := range c.Children {
			if ch.ID == id {
				return ch
			}
		}
		return nil
	}

	t.Run("point", func(t *testing.T) {
		ve := synthControlValueEntry(PseudoControl{Kind: PseudoPoint, Name: "C", PointDefault: []float64{0.25, 0.125}}, "mn")
		tdbs := tdbsOf(ve)
		tdb4 := child(tdbs, rifx.IDtdb4)
		if tdb4 == nil || len(tdb4.Data) != 124 || binary.BigEndian.Uint16(tdb4.Data[2:]) != 2 {
			t.Fatalf("point tdb4: want 124B dim=2")
		}
		cdat := child(tdbs, rifx.IDCdat)
		if cdat == nil || len(cdat.Data) != 48 {
			t.Fatalf("point cdat: want 48B, got %v", cdat)
		}
		if f64at(cdat.Data, 0) != 0.25 || f64at(cdat.Data, 8) != 0.125 {
			t.Errorf("point cdat = [%v,%v], want [0.25,0.125]", f64at(cdat.Data, 0), f64at(cdat.Data, 8))
		}
	})

	t.Run("point3d", func(t *testing.T) {
		ve := synthControlValueEntry(PseudoControl{Kind: PseudoPoint3D, PointDefault: []float64{0.25, 0.125, 0.0625}}, "mn")
		cdat := child(tdbsOf(ve), rifx.IDCdat)
		if cdat == nil || len(cdat.Data) != 72 {
			t.Fatalf("3dpoint cdat: want 72B")
		}
		if f64at(cdat.Data, 0) != 0.25 || f64at(cdat.Data, 8) != 0.125 || f64at(cdat.Data, 16) != 0.0625 {
			t.Errorf("3dpoint cdat wrong: %v %v %v", f64at(cdat.Data, 0), f64at(cdat.Data, 8), f64at(cdat.Data, 16))
		}
	})

	t.Run("point-origin-elided", func(t *testing.T) {
		if ve := synthControlValueEntry(PseudoControl{Kind: PseudoPoint}, "mn"); ve != nil {
			t.Errorf("point with no PointDefault should elide its value entry, got %d chunks", len(ve))
		}
	})

	t.Run("layer", func(t *testing.T) {
		ve := synthControlValueEntry(PseudoControl{Kind: PseudoLayer, Name: "L", LayerID: 15}, "mn")
		tdbs := tdbsOf(ve)
		tdpi := child(tdbs, rifx.IDTdpi)
		if tdpi == nil || be32(tdpi.Data, 0) != 15 {
			t.Errorf("layer tdpi = %v, want 15", tdpi)
		}
		if child(tdbs, rifx.IDTdps) == nil {
			t.Errorf("layer value entry missing tdps")
		}
	})
}
