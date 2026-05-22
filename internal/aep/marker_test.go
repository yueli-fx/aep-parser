package aep_test

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// nmhdMeta describes the per-marker NmHd values a synthetic marker should
// carry. Zero values reproduce AE's defaults (no duration, no label color).
type nmhdMeta struct {
	DurationTicks uint32 // NmHd @0x08; raw 600ths-of-a-second
	Label         uint8  // NmHd @0x10; 0..16 timeline label color index
}

// buildMarkers wraps N markers (time-only, all five Utf8 fields per marker)
// into an "ADBE Marker" mrst LIST. Pass per-marker nmhdMeta to set the NmHd
// duration / label fields; omitting metas (the common case) writes 20 zero
// bytes per NmHd, matching AE's defaults.
func (b *rifxBuilder) buildMarkers(times []float64, comments [][5]string, metas ...nmhdMeta) []byte {
	// tdbs branch: kfl with 16-byte keyframes (time at 0x00, rest zero).
	var ldatData []byte
	for _, t := range times {
		blk := make([]byte, 16)
		binary.BigEndian.PutUint32(blk[0:], uint32(math.Round(t*8000)))
		ldatData = append(ldatData, blk...)
	}
	lhd3 := b.chunk("lhd3", buildLhd3(uint32(len(times)), 16))
	ldat := b.chunk("ldat", ldatData)
	var kflInner []byte
	kflInner = append(kflInner, lhd3...)
	kflInner = append(kflInner, ldat...)
	kflList := b.listChunk("LIST", "list", kflInner)
	tdb4 := b.chunk("tdb4", buildTdb4(0x01))
	var tdbsInner []byte
	tdbsInner = append(tdbsInner, tdb4...)
	tdbsInner = append(tdbsInner, kflList...)
	tdbsList := b.listChunk("LIST", "tdbs", tdbsInner)

	// mrky branch: one Nmrd per marker, each with NmHd + 5 Utf8.
	var mrkyBody []byte
	for i := range times {
		nmHdData := make([]byte, 20)
		if i < len(metas) {
			binary.BigEndian.PutUint32(nmHdData[0x08:], metas[i].DurationTicks)
			nmHdData[0x10] = metas[i].Label
		}
		nmHd := b.chunk("NmHd", nmHdData)
		var c [5]string
		if i < len(comments) {
			c = comments[i]
		}
		var nmrdBody []byte
		nmrdBody = append(nmrdBody, nmHd...)
		for _, s := range c {
			nmrdBody = append(nmrdBody, b.chunk("Utf8", []byte(s))...)
		}
		nmrdList := b.listChunk("LIST", "Nmrd", nmrdBody)
		mrkyBody = append(mrkyBody, nmrdList...)
	}
	mrkyList := b.listChunk("LIST", "mrky", mrkyBody)

	var mrstInner []byte
	mrstInner = append(mrstInner, tdbsList...)
	mrstInner = append(mrstInner, mrkyList...)
	mrstList := b.listChunk("LIST", "mrst", mrstInner)

	var out []byte
	out = append(out, b.chunk("tdmn", []byte("ADBE Marker"))...)
	out = append(out, mrstList...)
	return out
}

// TestMarkerNmHdDecode exercises duration + label decode through synthetic
// NmHd payloads. Duration is stored in NmHd @0x08 as a uint32 of 600ths-
// of-a-second; label color at NmHd @0x10 as a raw uint8 index (0..16).
func TestMarkerNmHdDecode(t *testing.T) {
	cases := []struct {
		name         string
		nmhd         nmhdMeta
		wantDuration float64
		wantLabel    uint8
	}{
		{"point marker (defaults)", nmhdMeta{}, 0.0, 0},
		{"1-second duration", nmhdMeta{DurationTicks: 600}, 1.0, 0},
		{"2.5-second duration", nmhdMeta{DurationTicks: 1500}, 2.5, 0},
		{"label only", nmhdMeta{Label: 7}, 0.0, 7},
		{"duration + label", nmhdMeta{DurationTicks: 300, Label: 3}, 0.5, 3},
		{"max label index", nmhdMeta{Label: 16}, 0.0, 16},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rb := &rifxBuilder{}
			markers := rb.buildMarkers([]float64{1.0}, [][5]string{{"m"}}, c.nmhd)
			var tdgpBody []byte
			tdgpBody = append(tdgpBody, markers...)
			tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
			data := wrapAsLayer(tdgpBody)

			proj, err := aep.FromReader(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("FromReader: %v", err)
			}
			ms := proj.Compositions[0].Layers[0].Markers
			if len(ms) != 1 {
				t.Fatalf("Markers = %d, want 1", len(ms))
			}
			if math.Abs(ms[0].Duration-c.wantDuration) > 1e-9 {
				t.Errorf("Duration = %v, want %v", ms[0].Duration, c.wantDuration)
			}
			if ms[0].Label != c.wantLabel {
				t.Errorf("Label = %d, want %d", ms[0].Label, c.wantLabel)
			}
		})
	}
}

// TestMarkerNmHdRealFixture is a smoke test: parses test_data/re_tickrate.aep,
// walks every layer for markers, and asserts NmHd fields surface
// consistently with the fixture's actual bytes.
func TestMarkerNmHdRealFixture(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_tickrate.aep")
	if err != nil {
		t.Fatalf("Open re_tickrate.aep: %v", err)
	}
	var all []*aep.Marker
	for _, comp := range proj.Compositions {
		for _, layer := range comp.Layers {
			all = append(all, layer.Markers...)
		}
	}
	if len(all) == 0 {
		t.Skip("no layer markers in fixture (may be comp-level markers; out of scope)")
	}
	for i, m := range all {
		if m.Duration != 0.0 {
			t.Errorf("marker[%d].Duration = %v, want 0.0 (fixture has point markers with NmHd[0x08:0x0B]=0)",
				i, m.Duration)
		}
		if m.Label != 0 {
			t.Errorf("marker[%d].Label = %d, want 0 (fixture has NmHd[0x10]=0)", i, m.Label)
		}
	}
}

// TestCompositionMarkersReal exercises composition-level marker decoding
// against test_data/re_compmarker.aep (built by /tmp/re_compmarker.jsx
// in AE 2020). Skips silently when the fixture is absent.
func TestCompositionMarkersReal(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_compmarker.aep")
	if err != nil {
		t.Skipf("re_compmarker.aep not present; rerun the JSX in AE to regenerate")
	}
	var comp *aep.Composition
	for _, c := range proj.Compositions {
		if c.Name == "RE_CM" {
			comp = c
			break
		}
	}
	if comp == nil {
		t.Fatal("RE_CM comp missing from fixture")
	}
	if len(comp.Markers) != 2 {
		t.Fatalf("Composition.Markers = %d, want 2", len(comp.Markers))
	}
	m0 := comp.Markers[0]
	if math.Abs(m0.Time-1.0) > 1e-3 {
		t.Errorf("Markers[0].Time = %g, want 1.0", m0.Time)
	}
	if math.Abs(m0.Duration-1.5) > 1e-3 {
		t.Errorf("Markers[0].Duration = %g, want 1.5", m0.Duration)
	}
	if m0.Label != 3 {
		t.Errorf("Markers[0].Label = %d, want 3", m0.Label)
	}
	if m0.Comment != "comp marker A" {
		t.Errorf("Markers[0].Comment = %q, want %q", m0.Comment, "comp marker A")
	}
	m1 := comp.Markers[1]
	if math.Abs(m1.Time-2.5) > 1e-3 {
		t.Errorf("Markers[1].Time = %g, want 2.5", m1.Time)
	}
	if m1.Duration != 0 {
		t.Errorf("Markers[1].Duration = %g, want 0 (point marker)", m1.Duration)
	}
	if m1.Comment != "second marker" {
		t.Errorf("Markers[1].Comment = %q", m1.Comment)
	}
	if m1.Chapter != "chap-X" {
		t.Errorf("Markers[1].Chapter = %q", m1.Chapter)
	}
}

// TestMarkerSetters covers Marker.SetTime / SetDuration / SetLabel
// (length-preserving) + SetComment / SetChapter / SetURL (length-
// variable Utf8 chunk replacement). Drives them through the real
// test_data/re_compmarker.aep fixture and round-trips via WriteAEP.
func TestMarkerSetters(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_compmarker.aep")
	if err != nil {
		t.Skipf("re_compmarker.aep not present; rerun the JSX in AE")
	}
	var comp *aep.Composition
	for _, c := range proj.Compositions {
		if c.Name == "RE_CM" {
			comp = c
			break
		}
	}
	if comp == nil || len(comp.Markers) != 2 {
		t.Fatalf("expected 2 comp markers; got %d", len(comp.Markers))
	}

	m0, m1 := comp.Markers[0], comp.Markers[1]

	mustNoErr := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	mustNoErr("m0.SetTime", m0.SetTime(0.5))
	mustNoErr("m0.SetDuration", m0.SetDuration(2.0))
	mustNoErr("m0.SetLabel", m0.SetLabel(9))
	mustNoErr("m0.SetComment", m0.SetComment("rewritten primary"))
	mustNoErr("m1.SetURL", m1.SetURL("https://example.org/new"))

	if math.Abs(m0.Time-0.5) > 1e-3 {
		t.Errorf("m0.Time = %g, want 0.5", m0.Time)
	}
	if math.Abs(m0.Duration-2.0) > 1e-3 {
		t.Errorf("m0.Duration = %g, want 2.0", m0.Duration)
	}
	if m0.Label != 9 {
		t.Errorf("m0.Label = %d, want 9", m0.Label)
	}
	if m0.Comment != "rewritten primary" {
		t.Errorf("m0.Comment = %q", m0.Comment)
	}
	if m1.URL != "https://example.org/new" {
		t.Errorf("m1.URL = %q", m1.URL)
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	var comp2 *aep.Composition
	for _, c := range proj2.Compositions {
		if c.Name == "RE_CM" {
			comp2 = c
			break
		}
	}
	if comp2 == nil || len(comp2.Markers) != 2 {
		t.Fatalf("after roundtrip: comp/markers missing")
	}
	r0, r1 := comp2.Markers[0], comp2.Markers[1]
	if math.Abs(r0.Time-0.5) > 1e-3 {
		t.Errorf("roundtrip m0.Time = %g", r0.Time)
	}
	if math.Abs(r0.Duration-2.0) > 1e-3 {
		t.Errorf("roundtrip m0.Duration = %g", r0.Duration)
	}
	if r0.Label != 9 {
		t.Errorf("roundtrip m0.Label = %d", r0.Label)
	}
	if r0.Comment != "rewritten primary" {
		t.Errorf("roundtrip m0.Comment = %q", r0.Comment)
	}
	if r1.URL != "https://example.org/new" {
		t.Errorf("roundtrip m1.URL = %q", r1.URL)
	}
}

func TestMarkerSettersRejectMissingChunks(t *testing.T) {
	m := &aep.Marker{}
	if err := m.SetTime(1); err == nil {
		t.Error("SetTime on standalone marker: expected error")
	}
	if err := m.SetDuration(1); err == nil {
		t.Error("SetDuration on standalone marker: expected error")
	}
	if err := m.SetLabel(1); err == nil {
		t.Error("SetLabel on standalone marker: expected error")
	}
	if err := m.SetComment("x"); err == nil {
		t.Error("SetComment on standalone marker: expected error")
	}
}
