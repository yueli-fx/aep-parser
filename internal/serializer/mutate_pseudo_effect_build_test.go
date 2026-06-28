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
		{"颜色", "d1d5c9ab"},               // AE color pard 0003
		{"角度", "bdc7b6c8"},               // AE angle pard 0001
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
		{"色", "9046"},           // kanji
		{"ア", "8341"},           // katakana
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
		es, err := synthControlEntries(PseudoControl{Kind: PseudoDropdown, Name: "Menu", Options: []string{"a", "b"}, Default: 2}, PseudoLabelGBK, 0)
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
		es, _ := synthControlEntries(PseudoControl{Kind: PseudoGroupStart, Name: "Grp"}, PseudoLabelGBK, 0)
		d := es[0].pard.Data
		if d[0x0F] != 0x0d || be32(d, 0x04) != 0 || be32(d, 0x30) != 2 {
			t.Errorf("group-start pard @0x0F/@0x04/@0x30 = %#x/%#x/%#x, want 0x0d/0x0/0x2", d[0x0F], be32(d, 0x04), be32(d, 0x30))
		}
	})

	// Group end: 0x0e, @0x04=0x08, @0x30=2 — golden 0013/0005.
	t.Run("group-end", func(t *testing.T) {
		es, _ := synthControlEntries(PseudoControl{Kind: PseudoGroupEnd}, PseudoLabelGBK, 0)
		d := es[0].pard.Data
		if d[0x0F] != 0x0e || be32(d, 0x04) != 0x08 || be32(d, 0x30) != 2 {
			t.Errorf("group-end pard @0x0F/@0x04/@0x30 = %#x/%#x/%#x, want 0x0e/0x08/0x2", d[0x0F], be32(d, 0x04), be32(d, 0x30))
		}
	})

	// Label: a group-start immediately closed
	// by a generated group-end — golden 0004 (label "标签") + 0005 (group-end).
	t.Run("label", func(t *testing.T) {
		es, _ := synthControlEntries(PseudoControl{Kind: PseudoLabel, Name: "Note"}, PseudoLabelGBK, 0)
		if len(es) != 2 {
			t.Fatalf("label: got %d entries, want 2 (start+end)", len(es))
		}
		start := es[0].pard.Data
		if start[0x0F] != 0x0d || be32(start, 0x04) != 0 || be32(start, 0x30) != 2 {
			t.Errorf("label start @0x0F/@0x04/@0x30 = %#x/%#x/%#x, want 0x0d/0x0/0x2", start[0x0F], be32(start, 0x04), be32(start, 0x30))
		}
		end := es[1].pard.Data
		if end[0x0F] != 0x0e || be32(end, 0x04) != 0x08 {
			t.Errorf("label end @0x0F/@0x04 = %#x/%#x, want 0x0e/0x08", end[0x0F], be32(end, 0x04))
		}
	})

	t.Run("dimmed-label", func(t *testing.T) {
		es, _ := synthControlEntries(PseudoControl{Kind: PseudoLabel, Name: "Note", Dimmed: true}, PseudoLabelGBK, 0)
		start := es[0].pard.Data
		if start[0x0F] != 0x0d || be32(start, 0x04) != 0x20 || be32(start, 0x30) != 2 {
			t.Errorf("dimmed label start @0x0F/@0x04/@0x30 = %#x/%#x/%#x, want 0x0d/0x20/0x2", start[0x0F], be32(start, 0x04), be32(start, 0x30))
		}
	})
}

func f64at(d []byte, off int) float64 {
	return math.Float64frombits(binary.BigEndian.Uint64(d[off:]))
}

// TestPseudoControlEntries_ValueRules verifies the RE'd elision rules: a
// point/3D default lives IN the pard and emits NO value entry (AE builds the
// control from the pard); only the layer picker carries a value entry, and it
// must be flagged tdsb=1 (a plain property, NOT 3=the effect-header anchor — AE
// hides a picker flagged 3) with tdpi = the bound layer (an unset LayerID binds
// the host). RE'd from pseudo2.aep (Pseudo/148432) + pseudo_rich_demo.aep.
func TestPseudoControlEntries_ValueRules(t *testing.T) {
	const hostID = 99
	fx1616 := func(d []byte, off int) float64 { return float64(be32(d, off)) / 65536 }
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

	t.Run("point-coords-in-pard-no-value-entry", func(t *testing.T) {
		es, err := synthControlEntries(PseudoControl{Kind: PseudoPoint, Name: "C", PointDefault: []float64{0.25, 0.125}}, PseudoLabelGBK, hostID)
		if err != nil {
			t.Fatal(err)
		}
		if len(es) != 1 || es[0].valueEntry != nil {
			t.Fatalf("point must build from pard with NO value entry")
		}
		d := es[0].pard.Data
		if fx1616(d, 0x38) != 0.25 || fx1616(d, 0x3C) != 0.125 {
			t.Errorf("point pard @0x38/@0x3C (16.16) = %v/%v, want 0.25/0.125", fx1616(d, 0x38), fx1616(d, 0x3C))
		}
	})

	t.Run("3d-coords-in-pard-no-value-entry", func(t *testing.T) {
		es, _ := synthControlEntries(PseudoControl{Kind: PseudoPoint3D, PointDefault: []float64{0.25, 0.125, 0.0625}}, PseudoLabelGBK, hostID)
		if es[0].valueEntry != nil {
			t.Fatalf("3d must build from pard with NO value entry")
		}
		d := es[0].pard.Data
		if f64at(d, 0x38) != 0.25 || f64at(d, 0x40) != 0.125 || f64at(d, 0x48) != 0.0625 {
			t.Errorf("3d pard coords (f64) = %v/%v/%v, want 0.25/0.125/0.0625", f64at(d, 0x38), f64at(d, 0x40), f64at(d, 0x48))
		}
	})

	t.Run("layer-value-entry-tdsb1-tdpi", func(t *testing.T) {
		es, _ := synthControlEntries(PseudoControl{Kind: PseudoLayer, Name: "L", LayerID: 15}, PseudoLabelGBK, hostID)
		if es[0].valueEntry == nil {
			t.Fatal("layer must carry a value entry (tdpi binding)")
		}
		tdbs := tdbsOf(es[0].valueEntry("mn"))
		if be32(tdbs.Children[0].Data, 0) != 1 {
			t.Errorf("layer tdsb flag = %d, want 1 (plain property, not 3=anchor)", be32(tdbs.Children[0].Data, 0))
		}
		if tdpi := child(tdbs, rifx.IDTdpi); tdpi == nil || be32(tdpi.Data, 0) != 15 {
			t.Errorf("layer tdpi = %v, want 15", tdpi)
		}
	})

	t.Run("layer-unset-binds-host", func(t *testing.T) {
		es, _ := synthControlEntries(PseudoControl{Kind: PseudoLayer, Name: "L"}, PseudoLabelGBK, hostID)
		tdbs := tdbsOf(es[0].valueEntry("mn"))
		if tdpi := child(tdbs, rifx.IDTdpi); tdpi == nil || be32(tdpi.Data, 0) != hostID {
			t.Errorf("unset LayerID should bind host %d, got %v", hostID, tdpi)
		}
	})
}

// TestBuildPseudoEffect_LayerPickerBindsChosenLayer is the end-to-end regression
// for the layer-picker host-binding bug. A pseudo effect carries two kinds of
// tdpi: the effect header binds to its HOST layer (the layer the effect sits on),
// while a Layer-picker control binds to the user-CHOSEN layer. The shared
// effect-splice core (addEffectFromChunks) once blanket-retargeted every tdpi to
// the host — correct for AddEffect's foreign template bindings, but it silently
// clobbered the picker's chosen layer to the host. The unit test above checks
// synthControlValueEntry in isolation and so never caught it; this exercises the
// full BuildPseudoEffect splice. AE 2020 + AE 2025 accept and preserve both
// bindings on resave (header=Host, picker=Target).
func TestBuildPseudoEffect_LayerPickerBindsChosenLayer(t *testing.T) {
	p := NewProject()
	comp, err := NewComposition(p, "Main", 400, 400, 30, 5)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewShapeLayer(comp, "Host"); err != nil {
		t.Fatal(err)
	}
	if _, err := NewShapeLayer(comp, "Target"); err != nil {
		t.Fatal(err)
	}
	rp, err := Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	var host, target *Layer
	for _, c := range rp.Compositions {
		for _, cl := range c.Layers {
			switch cl.Name {
			case "Host":
				host = cl
			case "Target":
				target = cl
			}
		}
	}
	if host == nil || target == nil {
		t.Fatal("Host/Target not found after Reopen")
	}
	if host.ID == target.ID {
		t.Fatalf("Host and Target share ID %d — test cannot distinguish bindings", host.ID)
	}

	ctrls := []PseudoControl{
		{Kind: PseudoSlider, Name: "Amt", Min: 0, Max: 100, Default: 50},
		{Kind: PseudoLayer, Name: "Src", LayerID: target.ID},
	}
	if _, err := BuildPseudoEffect(host, "t", "Demo", "Demo", ctrls, PseudoLabelGBK); err != nil {
		t.Fatalf("BuildPseudoEffect: %v", err)
	}

	lb := layerBack(host)
	if lb == nil || lb.layrList == nil {
		t.Fatal("host layer has no Layr back-ref")
	}
	var tdpis []uint32
	var walk func(c *rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		if c.ID == rifx.IDTdpi && len(c.Data) >= 4 {
			tdpis = append(tdpis, binary.BigEndian.Uint32(c.Data[0:4]))
		}
		for _, ch := range c.Children {
			walk(ch)
		}
	}
	walk(lb.layrList)

	var sawHost, sawTarget bool
	for _, v := range tdpis {
		switch v {
		case host.ID:
			sawHost = true
		case target.ID:
			sawTarget = true
		}
	}
	if !sawHost {
		t.Errorf("no tdpi == host ID %d (effect header must bind to host); tdpis=%v", host.ID, tdpis)
	}
	if !sawTarget {
		t.Errorf("layer-picker tdpi not bound to chosen layer %d (clobbered to host?); tdpis=%v", target.ID, tdpis)
	}
}
