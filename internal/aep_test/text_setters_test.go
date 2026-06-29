package aep_test

import (
	"bytes"
	"math"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// TestSetTextLengthPreserving covers the length-preserving text-write
// API: replace text bytes in-place when the encoded byte count matches,
// reject when it doesn't, then byte-roundtrip the file to confirm AE
// would still open it.
func TestSetTextLengthPreserving(t *testing.T) {
	// Build an AEP whose text layer carries "Hello" — synthetic btds
	// payload using the same helpers as TestTextSourceDecodeSynthetic.
	payload := buildBtdsWrapped(buildBtdkPSText("Hello", "Myriad-Roman", 50, 0, [4]float64{1, 1, 1, 1}))
	data := buildExtendedAEP(payload, nil, nil)

	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	layer := proj.Compositions[0].Layers[0]
	if layer.TextSource == nil {
		t.Fatal("decoded TextSource is nil")
	}
	origRawLen := len(layer.TextSourceRaw)

	// Same-length swap — should succeed and update the decoded text.
	if err := layer.SetText("World"); err != nil {
		t.Fatalf("SetText same-length: %v", err)
	}
	if layer.TextSource.Text != "World" {
		t.Errorf("after SetText: Text = %q, want %q", layer.TextSource.Text, "World")
	}
	if got := len(layer.TextSourceRaw); got != origRawLen {
		t.Errorf("TextSourceRaw length changed: got %d, want %d", got, origRawLen)
	}

	// Length-mismatch — should error and leave the layer untouched.
	wantBefore := layer.TextSource.Text
	if err := layer.SetText("Hi"); err == nil {
		t.Error("SetText with shorter text: expected error, got nil")
	}
	if layer.TextSource.Text != wantBefore {
		t.Errorf("after failed SetText: Text = %q, want unchanged %q", layer.TextSource.Text, wantBefore)
	}

	// Round-trip the project bytes and reopen — the new text should
	// survive a full WriteAEP → FromReader cycle.
	var out bytes.Buffer
	if err := proj.WriteAEP(&out); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(out.Bytes()))
	if err != nil {
		t.Fatalf("FromReader (round-trip): %v", err)
	}
	got := proj2.Compositions[0].Layers[0].TextSource.Text
	if got != "World" {
		t.Errorf("after round-trip: Text = %q, want %q", got, "World")
	}

	// Length predictor must agree with the actual write.
	if got, want := aep.TextEncodedByteLen("Hello"), aep.TextEncodedByteLen("World"); got != want {
		t.Errorf("TextEncodedByteLen disagreement: Hello=%d World=%d (ASCII same-length should match)", got, want)
	}
}

// TestTextPerRunSetters covers length-variable per-run writes against
// the AE 2020 fixture re_text.aep — change font size / fill color /
// tracking / leading / baseline shift / justification, round-trip via
// WriteAEP, and confirm the decoded TextSource reflects the new values.
func TestTextPerRunSetters(t *testing.T) {
	proj, err := aep.Open("../../test_data/fixtures/re_text.aep")
	if err != nil {
		t.Skipf("re_text.aep not present; rerun re_text.jsx")
	}
	var comp *aep.Composition
	for _, c := range proj.Compositions {
		if c.Name == "RE_TEXT" {
			if comp == nil || len(c.Layers) > len(comp.Layers) {
				comp = c
			}
		}
	}
	if comp == nil {
		t.Fatal("RE_TEXT comp missing")
	}
	var layer *aep.Layer
	for _, l := range comp.Layers {
		if l.Name == "baseline_A" {
			layer = l
			break
		}
	}
	if layer == nil || layer.TextSource == nil {
		t.Fatal("baseline_A text layer not found")
	}

	// Drive every per-run setter.
	mustNoErr := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	mustNoErr("SetRunFontSize", layer.SetRunFontSize(0, 144))
	mustNoErr("SetRunTracking", layer.SetRunTracking(0, 200))
	mustNoErr("SetRunBaselineShift", layer.SetRunBaselineShift(0, 12))
	mustNoErr("SetRunAutoLeading", layer.SetRunAutoLeading(0, false))
	mustNoErr("SetRunLeading", layer.SetRunLeading(0, 180))
	mustNoErr("SetRunFillColor", layer.SetRunFillColor(0, [4]float64{0.25, 0.5, 0.75, 1}))
	mustNoErr("SetRunApplyStroke", layer.SetRunApplyStroke(0, true))
	mustNoErr("SetRunStrokeColor", layer.SetRunStrokeColor(0, [4]float64{1, 0, 0, 1}))
	mustNoErr("SetRunStrokeWidth", layer.SetRunStrokeWidth(0, 8))

	// In-memory check (TextSource was re-decoded by each setter).
	r := layer.TextSource.Runs[0]
	if r.FontSize != 144 {
		t.Errorf("after Set: FontSize = %g, want 144", r.FontSize)
	}
	if r.Tracking != 200 {
		t.Errorf("after Set: Tracking = %g, want 200", r.Tracking)
	}
	if r.BaselineShift != 12 {
		t.Errorf("after Set: BaselineShift = %g, want 12", r.BaselineShift)
	}
	if r.AutoLeading {
		t.Errorf("after Set: AutoLeading = true, want false")
	}
	if r.Leading != 180 {
		t.Errorf("after Set: Leading = %g, want 180", r.Leading)
	}
	wantFill := [4]float64{0.25, 0.5, 0.75, 1}
	for i := 0; i < 4; i++ {
		if math.Abs(r.FillColor[i]-wantFill[i]) > 1e-6 {
			t.Errorf("after Set: FillColor[%d] = %g, want %g (full=%v)", i, r.FillColor[i], wantFill[i], r.FillColor)
		}
	}
	if !r.ApplyStroke || r.StrokeWidth != 8 {
		t.Errorf("after Set: ApplyStroke=%v StrokeWidth=%g", r.ApplyStroke, r.StrokeWidth)
	}
	wantStroke := [4]float64{1, 0, 0, 1}
	for i := 0; i < 4; i++ {
		if math.Abs(r.StrokeColor[i]-wantStroke[i]) > 1e-6 {
			t.Errorf("after Set: StrokeColor[%d] = %g, want %g", i, r.StrokeColor[i], wantStroke[i])
		}
	}

	// Paragraph justification setter (single para since text="A").
	if len(layer.TextSource.Paragraphs) > 0 {
		mustNoErr("SetParagraphJustification", layer.SetParagraphJustification(0, aep.TextJustifyCenter))
		if layer.TextSource.Justification != aep.TextJustifyCenter {
			t.Errorf("after Set: Justification = %s", layer.TextSource.Justification)
		}
	}

	// Round-trip via WriteAEP — confirm the modified btds parses back
	// into the same Go values.
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	var layer2 *aep.Layer
	for _, c := range proj2.Compositions {
		if c.Name != "RE_TEXT" {
			continue
		}
		for _, l := range c.Layers {
			if l.Name == "baseline_A" {
				layer2 = l
				break
			}
		}
	}
	if layer2 == nil || layer2.TextSource == nil || len(layer2.TextSource.Runs) == 0 {
		t.Fatal("after roundtrip: baseline_A or text source missing")
	}
	rt := layer2.TextSource.Runs[0]
	if rt.FontSize != 144 {
		t.Errorf("roundtrip: FontSize = %g", rt.FontSize)
	}
	if rt.Tracking != 200 {
		t.Errorf("roundtrip: Tracking = %g", rt.Tracking)
	}
	if rt.BaselineShift != 12 {
		t.Errorf("roundtrip: BaselineShift = %g", rt.BaselineShift)
	}
	if math.Abs(rt.FillColor[0]-0.25) > 1e-6 || math.Abs(rt.FillColor[1]-0.5) > 1e-6 {
		t.Errorf("roundtrip: FillColor = %v", rt.FillColor)
	}
	if !rt.ApplyStroke || rt.StrokeWidth != 8 {
		t.Errorf("roundtrip: ApplyStroke=%v StrokeWidth=%g", rt.ApplyStroke, rt.StrokeWidth)
	}
	if layer2.TextSource.Justification != aep.TextJustifyCenter {
		t.Errorf("roundtrip: Justification = %s", layer2.TextSource.Justification)
	}
}

func TestTextPerRunSettersRejectOutOfRange(t *testing.T) {
	l := &aep.Layer{Name: "stub"}
	if err := l.SetRunFontSize(0, 50); err == nil {
		t.Error("SetRunFontSize on non-text layer: expected error")
	}
}
