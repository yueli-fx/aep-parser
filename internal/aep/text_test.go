package aep_test

import (
	"bytes"
	"math"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestTextSourceDecodeSynthetic(t *testing.T) {
	cases := []struct {
		name     string
		text     string
		font     string
		size     float64
		just     int
		color    [4]float64 // [R, G, B, A]
		wantJust aep.TextJustification
	}{
		{"ascii_left", "Hello", "Myriad-Roman", 88, 0, [4]float64{1, 0, 0, 1}, aep.TextJustifyLeft},
		{"unicode_center", "你好", "YouYuan", 100, 2, [4]float64{0, 1, 0, 1}, aep.TextJustifyCenter},
		{"multiline_right", "A\rB", "Arial", 50, 1, [4]float64{0, 0, 1, 1}, aep.TextJustifyRight},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			payload := buildBtdsWrapped(buildBtdkPSText(tc.text, tc.font, tc.size, tc.just, tc.color))
			data := buildExtendedAEP(payload, nil, nil)
			proj, err := aep.FromReader(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("FromReader: %v", err)
			}
			if len(proj.Warnings) != 0 {
				t.Errorf("unexpected parse warnings: %v", proj.Warnings)
			}
			layer := proj.Compositions[0].Layers[0]
			if layer.TextSource == nil {
				t.Fatal("Layer.TextSource = nil (decode failed)")
			}
			ts := layer.TextSource

			// AE's CR-per-paragraph convention: source "A\rB" becomes "A\nB".
			want := strings.ReplaceAll(tc.text, "\r", "\n")
			if ts.Text != want {
				t.Errorf("Text = %q, want %q", ts.Text, want)
			}
			if ts.Justification != tc.wantJust {
				t.Errorf("Justification = %s, want %s", ts.Justification, tc.wantJust)
			}
			if len(ts.Fonts) != 1 || ts.Fonts[0] != tc.font {
				t.Errorf("Fonts = %v, want [%q]", ts.Fonts, tc.font)
			}
			if len(ts.Runs) != 1 {
				t.Fatalf("Runs = %d, want 1", len(ts.Runs))
			}
			run := ts.Runs[0]
			if run.FontIndex != 0 {
				t.Errorf("Run.FontIndex = %d, want 0", run.FontIndex)
			}
			if run.FontName != tc.font {
				t.Errorf("Run.FontName = %q, want %q", run.FontName, tc.font)
			}
			if run.FontSize != tc.size {
				t.Errorf("Run.FontSize = %g, want %g", run.FontSize, tc.size)
			}
			for i := 0; i < 4; i++ {
				if math.Abs(run.FillColor[i]-tc.color[i]) > 1e-9 {
					t.Errorf("Run.FillColor[%d] = %g, want %g (full=%v)",
						i, run.FillColor[i], tc.color[i], run.FillColor)
				}
			}
		})
	}
}

// TestTextSourceDecodeReal exercises the decoder against the real AE
// 2020 fixture built by test_data/re_text.jsx. Skips silently if the
// file isn't present (CI / fresh clones won't have it).
func TestTextSourceDecodeReal(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_text.aep")
	if err != nil {
		t.Skipf("re_text.aep not present; run test_data/re_text.jsx in AE to generate")
	}
	// Find the RE_TEXT comp with the largest layer count (newest run wins
	// when multiple sessions left stale comps behind).
	var comp *aep.Composition
	for _, c := range proj.Compositions {
		if c.Name != "RE_TEXT" {
			continue
		}
		if comp == nil || len(c.Layers) > len(comp.Layers) {
			comp = c
		}
	}
	if comp == nil {
		t.Fatal("no RE_TEXT comp in fixture")
	}
	byName := map[string]*aep.Layer{}
	for _, l := range comp.Layers {
		byName[l.Name] = l
	}

	check := func(name, wantText string, wantJust aep.TextJustification, wantSize float64) {
		t.Helper()
		l, ok := byName[name]
		if !ok {
			t.Errorf("layer %q missing from RE_TEXT", name)
			return
		}
		if l.TextSource == nil {
			t.Errorf("layer %q TextSource = nil", name)
			return
		}
		if l.TextSource.Text != wantText {
			t.Errorf("layer %q Text = %q, want %q", name, l.TextSource.Text, wantText)
		}
		if l.TextSource.Justification != wantJust {
			t.Errorf("layer %q Just = %s, want %s", name, l.TextSource.Justification, wantJust)
		}
		if wantSize > 0 && len(l.TextSource.Runs) > 0 && l.TextSource.Runs[0].FontSize != wantSize {
			t.Errorf("layer %q FontSize = %g, want %g", name, l.TextSource.Runs[0].FontSize, wantSize)
		}
	}

	check("baseline_A", "A", aep.TextJustifyLeft, 0)
	check("text_AB", "AB", aep.TextJustifyLeft, 0)
	check("text_hello", "Hello", aep.TextJustifyLeft, 0)
	check("text_unicode", "你好", aep.TextJustifyLeft, 0)
	check("text_two_lines", "A\nB", aep.TextJustifyLeft, 0)
	check("size_100", "A", aep.TextJustifyLeft, 100)
	check("size_200", "A", aep.TextJustifyLeft, 200)
	check("just_center", "A", aep.TextJustifyCenter, 0)
	check("just_right", "A", aep.TextJustifyRight, 0)

	// Color checks: the helper above is text/just/size only. Spot-check
	// fill color on the dedicated color variants.
	colors := map[string][4]float64{
		"color_red":   {1, 0, 0, 1},
		"color_green": {0, 1, 0, 1},
		"color_blue":  {0, 0, 1, 1},
	}
	for name, want := range colors {
		l, ok := byName[name]
		if !ok || l.TextSource == nil || len(l.TextSource.Runs) == 0 {
			t.Errorf("color variant %q missing or undecoded", name)
			continue
		}
		got := l.TextSource.Runs[0].FillColor
		for i := 0; i < 4; i++ {
			if math.Abs(got[i]-want[i]) > 1e-6 {
				t.Errorf("layer %q FillColor[%d] = %g, want %g (full=%v)",
					name, i, got[i], want[i], got)
				break
			}
		}
	}

	// Extended per-run fields.
	if l := byName["tracking_500"]; l != nil && l.TextSource != nil && len(l.TextSource.Runs) > 0 {
		if got := l.TextSource.Runs[0].Tracking; got != 500 {
			t.Errorf("tracking_500.Tracking = %g, want 500", got)
		}
	} else {
		t.Error("tracking_500 missing or undecoded")
	}
	if l := byName["leading_300"]; l != nil && l.TextSource != nil && len(l.TextSource.Runs) > 0 {
		r := l.TextSource.Runs[0]
		if r.AutoLeading {
			t.Errorf("leading_300.AutoLeading = true, want false (explicit leading)")
		}
		if r.Leading != 300 {
			t.Errorf("leading_300.Leading = %g, want 300", r.Leading)
		}
		if len(l.TextSource.Paragraphs) != 2 {
			t.Errorf("leading_300.Paragraphs = %d, want 2 (two-line text)", len(l.TextSource.Paragraphs))
		}
	} else {
		t.Error("leading_300 missing or undecoded")
	}
	if l := byName["baseline_A"]; l != nil && l.TextSource != nil && len(l.TextSource.Runs) > 0 {
		r := l.TextSource.Runs[0]
		if !r.AutoLeading {
			t.Errorf("baseline_A.AutoLeading = false, want true (default)")
		}
		// AE auto-leading is FontSize × 1.2.
		if want := r.FontSize * 1.2; math.Abs(r.Leading-want) > 0.1 {
			t.Errorf("baseline_A.Leading = %g, want ~%g (auto = size × 1.2)", r.Leading, want)
		}
		if r.Tracking != 0 {
			t.Errorf("baseline_A.Tracking = %g, want 0", r.Tracking)
		}
		if r.ApplyStroke {
			t.Errorf("baseline_A.ApplyStroke = true, want false (default)")
		}
	}
	if l := byName["stroke_yellow"]; l != nil && l.TextSource != nil && len(l.TextSource.Runs) > 0 {
		r := l.TextSource.Runs[0]
		if !r.ApplyStroke {
			t.Errorf("stroke_yellow.ApplyStroke = false, want true")
		}
		if r.StrokeWidth != 4 {
			t.Errorf("stroke_yellow.StrokeWidth = %g, want 4", r.StrokeWidth)
		}
		want := [4]float64{1, 1, 0, 1}
		for i := 0; i < 4; i++ {
			if math.Abs(r.StrokeColor[i]-want[i]) > 1e-6 {
				t.Errorf("stroke_yellow.StrokeColor[%d] = %g, want %g (full=%v)",
					i, r.StrokeColor[i], want[i], r.StrokeColor)
				break
			}
		}
	} else {
		t.Error("stroke_yellow missing or undecoded")
	}

	// Box-text detection: point variants should be IsBoxText=false, the
	// dedicated box_text variant should report the rectangle's AABB.
	if l := byName["baseline_A"]; l != nil && l.TextSource != nil {
		if l.TextSource.IsBoxText {
			t.Error("baseline_A.IsBoxText = true, want false (point text)")
		}
	}
	if l := byName["box_text"]; l != nil && l.TextSource != nil {
		if !l.TextSource.IsBoxText {
			t.Error("box_text.IsBoxText = false, want true")
		}
		want := [4]float64{-200, -100, 200, 100} // 400×200 box centered at origin
		for i := 0; i < 4; i++ {
			if math.Abs(l.TextSource.BoxBounds[i]-want[i]) > 1e-6 {
				t.Errorf("box_text.BoxBounds[%d] = %g, want %g (full=%v)",
					i, l.TextSource.BoxBounds[i], want[i], l.TextSource.BoxBounds)
				break
			}
		}
	} else {
		t.Error("box_text variant missing — regenerate test_data/re_text.aep")
	}
}

// TestTextSourceExtendedFieldsReal verifies the new TextStyleRun
// fields (BaselineShift / VerticalScale / Tsume) against the AE 2020
// fixture built by /tmp/re_text2.jsx. Skips if absent.
func TestTextSourceExtendedFieldsReal(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_text2.aep")
	if err != nil {
		t.Skipf("re_text2.aep not present")
	}
	var comp *aep.Composition
	for _, c := range proj.Compositions {
		if c.Name == "RE_TEXT2" {
			comp = c
			break
		}
	}
	if comp == nil {
		t.Fatal("RE_TEXT2 comp missing")
	}
	byName := map[string]*aep.Layer{}
	for _, l := range comp.Layers {
		byName[l.Name] = l
	}

	checkRunField := func(layerName string, getter func(r aep.TextStyleRun) float64, want float64, label string) {
		t.Helper()
		l, ok := byName[layerName]
		if !ok {
			t.Errorf("layer %q missing", layerName)
			return
		}
		if l.TextSource == nil || len(l.TextSource.Runs) == 0 {
			t.Errorf("layer %q has no decoded text run", layerName)
			return
		}
		if got := getter(l.TextSource.Runs[0]); math.Abs(got-want) > 1e-3 {
			t.Errorf("layer %q %s = %g, want %g", layerName, label, got, want)
		}
	}

	checkRunField("baseShift_30", func(r aep.TextStyleRun) float64 { return r.BaselineShift }, 30, "BaselineShift")
	checkRunField("baseShift_neg", func(r aep.TextStyleRun) float64 { return r.BaselineShift }, -20, "BaselineShift")
	checkRunField("vScale_80", func(r aep.TextStyleRun) float64 { return r.VerticalScale }, 80, "VerticalScale")
	checkRunField("tsume_50", func(r aep.TextStyleRun) float64 { return r.Tsume }, 50, "Tsume")
	checkRunField("baseline", func(r aep.TextStyleRun) float64 { return r.HorizontalScale }, 1, "HorizontalScale default")
	checkRunField("baseline", func(r aep.TextStyleRun) float64 { return r.VerticalScale }, 1, "VerticalScale default")
	checkRunField("baseline", func(r aep.TextStyleRun) float64 { return r.BaselineShift }, 0, "BaselineShift default")
	checkRunField("baseline", func(r aep.TextStyleRun) float64 { return r.Tsume }, 0, "Tsume default")
}

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
	proj, err := aep.Open("../../test_data/re_text.aep")
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

func TestTextCapsBaselineStrokeReal(t *testing.T) {
	_, compPt, _ := openCapsAE24(t)
	byName := map[string]*aep.Layer{}
	for _, l := range compPt.Layers {
		byName[l.Name] = l
	}

	checkRun := func(layerName string, check func(t *testing.T, r aep.TextStyleRun)) {
		t.Helper()
		l, ok := byName[layerName]
		if !ok {
			t.Errorf("layer %q missing", layerName)
			return
		}
		if l.TextSource == nil || len(l.TextSource.Runs) == 0 {
			t.Errorf("layer %q has no decoded text run", layerName)
			return
		}
		check(t, l.TextSource.Runs[0])
	}

	// baseline: all defaults
	checkRun("pt_baseline", func(t *testing.T, r aep.TextStyleRun) {
		if r.CapsOption != aep.TextCapsNormal {
			t.Errorf("pt_baseline CapsOption = %s, want Normal", r.CapsOption)
		}
		if r.BaselineOption != aep.TextBaselineNormal {
			t.Errorf("pt_baseline BaselineOption = %s, want Normal", r.BaselineOption)
		}
		if !r.StrokeOverFill {
			t.Errorf("pt_baseline StrokeOverFill = false, want true (AE default)")
		}
		if r.AllCaps() || r.SmallCaps() || r.Superscript() || r.Subscript() {
			t.Errorf("pt_baseline mirror flags should all be false")
		}
	})

	// caps variants
	checkRun("pt_caps_all", func(t *testing.T, r aep.TextStyleRun) {
		if r.CapsOption != aep.TextCapsAll {
			t.Errorf("CapsOption = %s, want AllCaps", r.CapsOption)
		}
		if !r.AllCaps() || r.SmallCaps() {
			t.Errorf("AllCaps()/SmallCaps() = %v/%v, want true/false", r.AllCaps(), r.SmallCaps())
		}
	})
	checkRun("pt_caps_small", func(t *testing.T, r aep.TextStyleRun) {
		if r.CapsOption != aep.TextCapsSmall {
			t.Errorf("CapsOption = %s, want SmallCaps", r.CapsOption)
		}
		if r.AllCaps() || !r.SmallCaps() {
			t.Errorf("AllCaps()/SmallCaps() = %v/%v, want false/true", r.AllCaps(), r.SmallCaps())
		}
	})
	checkRun("pt_caps_all_small", func(t *testing.T, r aep.TextStyleRun) {
		if r.CapsOption != aep.TextCapsAllSmall {
			t.Errorf("CapsOption = %s, want AllSmallCaps", r.CapsOption)
		}
		if !r.AllCaps() || !r.SmallCaps() {
			t.Errorf("AllCaps()/SmallCaps() = %v/%v, want true/true", r.AllCaps(), r.SmallCaps())
		}
	})

	// baseline option variants
	checkRun("pt_base_super", func(t *testing.T, r aep.TextStyleRun) {
		if r.BaselineOption != aep.TextBaselineSuperscript {
			t.Errorf("BaselineOption = %s, want Superscript", r.BaselineOption)
		}
		if !r.Superscript() || r.Subscript() {
			t.Errorf("Superscript()/Subscript() = %v/%v, want true/false", r.Superscript(), r.Subscript())
		}
	})
	checkRun("pt_base_sub", func(t *testing.T, r aep.TextStyleRun) {
		if r.BaselineOption != aep.TextBaselineSubscript {
			t.Errorf("BaselineOption = %s, want Subscript", r.BaselineOption)
		}
		if r.Superscript() || !r.Subscript() {
			t.Errorf("Superscript()/Subscript() = %v/%v, want false/true", r.Superscript(), r.Subscript())
		}
	})

	// strokeOverFill variants
	checkRun("pt_stroke_over", func(t *testing.T, r aep.TextStyleRun) {
		if !r.StrokeOverFill {
			t.Errorf("pt_stroke_over StrokeOverFill = false, want true")
		}
	})
	checkRun("pt_stroke_under", func(t *testing.T, r aep.TextStyleRun) {
		if r.StrokeOverFill {
			t.Errorf("pt_stroke_under StrokeOverFill = true, want false")
		}
	})
}

func TestTextParagraphFieldsReal(t *testing.T) {
	_, _, compBx := openCapsAE24(t)
	byName := map[string]*aep.Layer{}
	for _, l := range compBx.Layers {
		byName[l.Name] = l
	}

	checkPara := func(layerName string, check func(t *testing.T, p aep.TextParagraph)) {
		t.Helper()
		l, ok := byName[layerName]
		if !ok {
			t.Errorf("layer %q missing", layerName)
			return
		}
		if l.TextSource == nil || len(l.TextSource.Paragraphs) == 0 {
			t.Errorf("layer %q has no decoded paragraphs", layerName)
			return
		}
		check(t, l.TextSource.Paragraphs[0])
	}

	checkPara("bx_baseline", func(t *testing.T, p aep.TextParagraph) {
		if p.FirstLineIndent != 0 || p.StartIndent != 0 || p.EndIndent != 0 || p.SpaceBefore != 0 || p.SpaceAfter != 0 {
			t.Errorf("bx_baseline indents/spaces should all be 0; got %+v", p)
		}
		if !p.AutoHyphenate {
			t.Errorf("bx_baseline AutoHyphenate = false, want true (AE default)")
		}
	})
	checkPara("bx_startIndent_50", func(t *testing.T, p aep.TextParagraph) {
		if p.StartIndent != 50 {
			t.Errorf("StartIndent = %g, want 50", p.StartIndent)
		}
	})
	checkPara("bx_endIndent_60", func(t *testing.T, p aep.TextParagraph) {
		if p.EndIndent != 60 {
			t.Errorf("EndIndent = %g, want 60", p.EndIndent)
		}
	})
	checkPara("bx_firstLine_70", func(t *testing.T, p aep.TextParagraph) {
		if p.FirstLineIndent != 70 {
			t.Errorf("FirstLineIndent = %g, want 70", p.FirstLineIndent)
		}
	})
	checkPara("bx_spaceBefore_30", func(t *testing.T, p aep.TextParagraph) {
		if p.SpaceBefore != 30 {
			t.Errorf("SpaceBefore = %g, want 30", p.SpaceBefore)
		}
	})
	checkPara("bx_spaceAfter_40", func(t *testing.T, p aep.TextParagraph) {
		if p.SpaceAfter != 40 {
			t.Errorf("SpaceAfter = %g, want 40", p.SpaceAfter)
		}
	})
	checkPara("bx_hyphenate_off", func(t *testing.T, p aep.TextParagraph) {
		if p.AutoHyphenate {
			t.Errorf("bx_hyphenate_off AutoHyphenate = true, want false")
		}
	})
}

// TestTextWave1Roundtrip drives every AE 24+ setter against the
// baseline layers and verifies values persist through WriteAEP.
func TestTextWave1Roundtrip(t *testing.T) {
	proj, compPt, compBx := openCapsAE24(t)
	var ptBase, bxBase *aep.Layer
	for _, l := range compPt.Layers {
		if l.Name == "pt_baseline" {
			ptBase = l
		}
	}
	for _, l := range compBx.Layers {
		if l.Name == "bx_baseline" {
			bxBase = l
		}
	}
	if ptBase == nil || bxBase == nil {
		t.Fatalf("baseline layers missing (ptBase=%v bxBase=%v)", ptBase, bxBase)
	}

	mustNoErr := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}

	// Per-run AE 24+ setters on pt_baseline.
	mustNoErr("SetRunCapsOption", ptBase.SetRunCapsOption(0, aep.TextCapsAllSmall))
	mustNoErr("SetRunBaselineOption", ptBase.SetRunBaselineOption(0, aep.TextBaselineSuperscript))
	mustNoErr("SetRunStrokeOverFill(false)", ptBase.SetRunStrokeOverFill(0, false))

	// Paragraph AE 24+ setters on bx_baseline.
	mustNoErr("SetParagraphFirstLineIndent", bxBase.SetParagraphFirstLineIndent(0, 25))
	mustNoErr("SetParagraphStartIndent", bxBase.SetParagraphStartIndent(0, 35))
	mustNoErr("SetParagraphEndIndent", bxBase.SetParagraphEndIndent(0, 45))
	mustNoErr("SetParagraphSpaceBefore", bxBase.SetParagraphSpaceBefore(0, 12))
	mustNoErr("SetParagraphSpaceAfter", bxBase.SetParagraphSpaceAfter(0, 18))
	mustNoErr("SetParagraphAutoHyphenate(false)", bxBase.SetParagraphAutoHyphenate(0, false))

	// In-memory check (each setter re-decoded the TextSource).
	pr := ptBase.TextSource.Runs[0]
	if pr.CapsOption != aep.TextCapsAllSmall {
		t.Errorf("in-mem CapsOption = %s, want AllSmall", pr.CapsOption)
	}
	if pr.BaselineOption != aep.TextBaselineSuperscript {
		t.Errorf("in-mem BaselineOption = %s, want Super", pr.BaselineOption)
	}
	if pr.StrokeOverFill {
		t.Errorf("in-mem StrokeOverFill = true, want false")
	}
	bp := bxBase.TextSource.Paragraphs[0]
	if bp.FirstLineIndent != 25 || bp.StartIndent != 35 || bp.EndIndent != 45 || bp.SpaceBefore != 12 || bp.SpaceAfter != 18 {
		t.Errorf("in-mem paragraph fields = %+v", bp)
	}
	if bp.AutoHyphenate {
		t.Errorf("in-mem AutoHyphenate = true, want false")
	}

	// Roundtrip.
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	var ptBase2, bxBase2 *aep.Layer
	for _, c := range proj2.Compositions {
		for _, l := range c.Layers {
			if l.Name == "pt_baseline" {
				ptBase2 = l
			}
			if l.Name == "bx_baseline" {
				bxBase2 = l
			}
		}
	}
	if ptBase2 == nil || bxBase2 == nil {
		t.Fatalf("roundtrip baseline layers missing")
	}
	rt := ptBase2.TextSource.Runs[0]
	if rt.CapsOption != aep.TextCapsAllSmall || rt.BaselineOption != aep.TextBaselineSuperscript || rt.StrokeOverFill {
		t.Errorf("roundtrip pt fields wrong: caps=%s base=%s strokeOver=%v", rt.CapsOption, rt.BaselineOption, rt.StrokeOverFill)
	}
	bp2 := bxBase2.TextSource.Paragraphs[0]
	if bp2.FirstLineIndent != 25 || bp2.StartIndent != 35 || bp2.EndIndent != 45 || bp2.SpaceBefore != 12 || bp2.SpaceAfter != 18 || bp2.AutoHyphenate {
		t.Errorf("roundtrip paragraph fields = %+v", bp2)
	}
}

func TestTextWave1RejectInvalid(t *testing.T) {
	l := &aep.Layer{Name: "stub"}
	if err := l.SetRunCapsOption(0, aep.TextCapsAll); err == nil {
		t.Error("SetRunCapsOption on non-text layer: expected error")
	}
	if err := l.SetParagraphStartIndent(0, 10); err == nil {
		t.Error("SetParagraphStartIndent on non-text layer: expected error")
	}
}

func TestTextAE24MoreRunFieldsReal(t *testing.T) {
	proj := openTextAE24More(t)
	cases := []struct {
		layer       string
		check       func(t *testing.T, r aep.TextStyleRun)
	}{
		{"pt_baseline", func(t *testing.T, r aep.TextStyleRun) {
			if r.AutoKernType != aep.TextAutoKernMetric {
				t.Errorf("baseline AutoKernType = %v, want Metric", r.AutoKernType)
			}
			if r.NoBreak {
				t.Errorf("baseline NoBreak = true, want false")
			}
			if r.LineJoinType != aep.TextLineJoinMiter {
				t.Errorf("baseline LineJoinType = %v, want Miter", r.LineJoinType)
			}
			if r.DigitSet != aep.TextDigitSetDefault {
				t.Errorf("baseline DigitSet = %v, want Default", r.DigitSet)
			}
		}},
		{"pt_autokern_optical", func(t *testing.T, r aep.TextStyleRun) {
			if r.AutoKernType != aep.TextAutoKernOptical {
				t.Errorf("AutoKernType = %v, want Optical", r.AutoKernType)
			}
		}},
		{"pt_nobreak_true", func(t *testing.T, r aep.TextStyleRun) {
			if !r.NoBreak {
				t.Errorf("NoBreak = false, want true")
			}
		}},
		{"pt_linejoin_round", func(t *testing.T, r aep.TextStyleRun) {
			if r.LineJoinType != aep.TextLineJoinRound {
				t.Errorf("LineJoinType = %v, want Round", r.LineJoinType)
			}
		}},
		{"pt_linejoin_bevel", func(t *testing.T, r aep.TextStyleRun) {
			if r.LineJoinType != aep.TextLineJoinBevel {
				t.Errorf("LineJoinType = %v, want Bevel", r.LineJoinType)
			}
		}},
		{"pt_digitset_arabic", func(t *testing.T, r aep.TextStyleRun) {
			if r.DigitSet != aep.TextDigitSetArabic {
				t.Errorf("DigitSet = %v, want Arabic", r.DigitSet)
			}
		}},
		{"pt_digitset_hindi", func(t *testing.T, r aep.TextStyleRun) {
			if r.DigitSet != aep.TextDigitSetHindi {
				t.Errorf("DigitSet = %v, want Hindi", r.DigitSet)
			}
		}},
	}
	for _, c := range cases {
		l := findLayerInComp(proj, "RE_TEXT24_PT", c.layer)
		if l == nil || l.TextSource == nil || len(l.TextSource.Runs) == 0 {
			t.Errorf("layer %q missing/empty", c.layer)
			continue
		}
		c.check(t, l.TextSource.Runs[0])
	}
}

func TestTextAE24MoreParagraphFieldsReal(t *testing.T) {
	proj := openTextAE24More(t)
	cases := []struct {
		layer string
		check func(t *testing.T, p aep.TextParagraph)
	}{
		{"bx_baseline", func(t *testing.T, p aep.TextParagraph) {
			if p.LeadingType != aep.TextLeadingRoman {
				t.Errorf("baseline LeadingType = %v, want Roman", p.LeadingType)
			}
			if p.HangingRoman {
				t.Errorf("baseline HangingRoman = true, want false")
			}
			if p.Direction != aep.TextDirectionLeftToRight {
				t.Errorf("baseline Direction = %v, want LTR", p.Direction)
			}
		}},
		{"bx_direction_rtl", func(t *testing.T, p aep.TextParagraph) {
			if p.Direction != aep.TextDirectionRightToLeft {
				t.Errorf("Direction = %v, want RTL", p.Direction)
			}
		}},
		{"bx_hangingroman_true", func(t *testing.T, p aep.TextParagraph) {
			if !p.HangingRoman {
				t.Errorf("HangingRoman = false, want true")
			}
		}},
		{"bx_leadingtype_japanese", func(t *testing.T, p aep.TextParagraph) {
			if p.LeadingType != aep.TextLeadingJapanese {
				t.Errorf("LeadingType = %v, want Japanese", p.LeadingType)
			}
		}},
	}
	for _, c := range cases {
		l := findLayerInComp(proj, "RE_TEXT24_BX", c.layer)
		if l == nil || l.TextSource == nil || len(l.TextSource.Paragraphs) == 0 {
			t.Errorf("layer %q missing/empty", c.layer)
			continue
		}
		c.check(t, l.TextSource.Paragraphs[0])
	}
}

func TestTextAE24MoreSettersRoundtrip(t *testing.T) {
	proj := openTextAE24More(t)
	pt := findLayerInComp(proj, "RE_TEXT24_PT", "pt_baseline")
	bx := findLayerInComp(proj, "RE_TEXT24_BX", "bx_baseline")
	if pt == nil || bx == nil {
		t.Fatalf("baseline layers missing (pt=%v bx=%v)", pt, bx)
	}
	mustNoErr := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	// Per-run
	mustNoErr("SetRunAutoKernType", pt.SetRunAutoKernType(0, aep.TextAutoKernOptical))
	mustNoErr("SetRunNoBreak", pt.SetRunNoBreak(0, true))
	mustNoErr("SetRunLineJoinType", pt.SetRunLineJoinType(0, aep.TextLineJoinBevel))
	mustNoErr("SetRunDigitSet", pt.SetRunDigitSet(0, aep.TextDigitSetArabic))
	// Per-paragraph
	mustNoErr("SetParagraphLeadingType", bx.SetParagraphLeadingType(0, aep.TextLeadingJapanese))
	mustNoErr("SetParagraphHangingRoman", bx.SetParagraphHangingRoman(0, true))
	mustNoErr("SetParagraphDirection", bx.SetParagraphDirection(0, aep.TextDirectionRightToLeft))

	// In-mem check
	r := pt.TextSource.Runs[0]
	if r.AutoKernType != aep.TextAutoKernOptical || !r.NoBreak || r.LineJoinType != aep.TextLineJoinBevel || r.DigitSet != aep.TextDigitSetArabic {
		t.Errorf("in-mem run fields wrong: %+v", r)
	}
	p := bx.TextSource.Paragraphs[0]
	if p.LeadingType != aep.TextLeadingJapanese || !p.HangingRoman || p.Direction != aep.TextDirectionRightToLeft {
		t.Errorf("in-mem para fields wrong: %+v", p)
	}

	// Roundtrip
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	pt2 := findLayerInComp(proj2, "RE_TEXT24_PT", "pt_baseline")
	bx2 := findLayerInComp(proj2, "RE_TEXT24_BX", "bx_baseline")
	if pt2 == nil || bx2 == nil {
		t.Fatalf("roundtrip baseline missing")
	}
	r2 := pt2.TextSource.Runs[0]
	if r2.AutoKernType != aep.TextAutoKernOptical || !r2.NoBreak || r2.LineJoinType != aep.TextLineJoinBevel || r2.DigitSet != aep.TextDigitSetArabic {
		t.Errorf("roundtrip run fields wrong: %+v", r2)
	}
	p2 := bx2.TextSource.Paragraphs[0]
	if p2.LeadingType != aep.TextLeadingJapanese || !p2.HangingRoman || p2.Direction != aep.TextDirectionRightToLeft {
		t.Errorf("roundtrip para fields wrong: %+v", p2)
	}
}

func TestTextManualKerningReal(t *testing.T) {
	proj := openTextAE24More(t)
	cases := []struct {
		layer        string
		wantKerning  int
		wantManual   []int
	}{
		// fixture text is "AaBb" (4 chars) for the pt_kerning_* variants.
		{"pt_kerning_neg50", -50, []int{-50, -50, -50, -50}},
		{"pt_kerning_100", 100, []int{100, 100, 100, 100}},
		// autokern Metric/Optical layers carry no /1/1[0]/0/8 sub-tree
		// and report Kerning=0 / ManualKerning=nil.
		{"pt_autokern_optical", 0, nil},
		{"pt_autokern_metric", 0, nil},
		// A baseline (also Metric default) — same as autokern_metric.
		{"pt_baseline_vert", 0, nil},
	}
	for _, c := range cases {
		l := findLayerInComp(proj, "RE_TEXT24_PT", c.layer)
		if l == nil {
			t.Errorf("%s missing", c.layer)
			continue
		}
		if l.TextSource == nil {
			t.Errorf("%s: TextSource nil", c.layer)
			continue
		}
		if l.TextSource.Kerning != c.wantKerning {
			t.Errorf("%s Kerning = %d, want %d", c.layer, l.TextSource.Kerning, c.wantKerning)
		}
		got := l.TextSource.ManualKerning
		if len(got) != len(c.wantManual) {
			t.Errorf("%s ManualKerning len=%d, want %d (%v)", c.layer, len(got), len(c.wantManual), got)
			continue
		}
		for i, v := range c.wantManual {
			if got[i] != v {
				t.Errorf("%s ManualKerning[%d] = %d, want %d", c.layer, i, got[i], v)
			}
		}
	}
}

func TestSetManualKerningRoundtrip(t *testing.T) {
	proj := openTextAE24More(t)
	l := findLayerInComp(proj, "RE_TEXT24_PT", "pt_kerning_neg50")
	if l == nil {
		t.Fatalf("pt_kerning_neg50 missing")
	}
	if l.TextSource == nil || len(l.TextSource.ManualKerning) != 4 {
		t.Fatalf("pre: want 4 manual-kerning entries, got %v", l.TextSource.ManualKerning)
	}
	// Pick values with varying byte widths so the splice path exercises
	// length-changing rewrites for both narrower (1-byte) and wider
	// (4-byte) values than the original -50 (3 bytes).
	newVals := []int{7, -200, 0, 999}
	if err := l.SetManualKerning(newVals); err != nil {
		t.Fatalf("SetManualKerning: %v", err)
	}

	// In-mem check.
	for i, v := range newVals {
		if l.TextSource.ManualKerning[i] != v {
			t.Errorf("in-mem ManualKerning[%d] = %d, want %d", i, l.TextSource.ManualKerning[i], v)
		}
	}
	if l.TextSource.Kerning != newVals[0] {
		t.Errorf("in-mem Kerning = %d, want %d", l.TextSource.Kerning, newVals[0])
	}

	// Roundtrip.
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	l2 := findLayerInComp(proj2, "RE_TEXT24_PT", "pt_kerning_neg50")
	if l2 == nil {
		t.Fatalf("roundtrip pt_kerning_neg50 missing")
	}
	if got := l2.TextSource.ManualKerning; len(got) != len(newVals) {
		t.Fatalf("roundtrip ManualKerning len = %d, want %d (%v)", len(got), len(newVals), got)
	}
	for i, v := range newVals {
		if l2.TextSource.ManualKerning[i] != v {
			t.Errorf("roundtrip ManualKerning[%d] = %d, want %d", i, l2.TextSource.ManualKerning[i], v)
		}
	}
	if l2.TextSource.Kerning != newVals[0] {
		t.Errorf("roundtrip Kerning = %d, want %d", l2.TextSource.Kerning, newVals[0])
	}
}

func TestSetManualKerningRejectsInvalid(t *testing.T) {
	proj := openTextAE24More(t)

	// No-slot layer (AE didn't emit /1/1[0]/0/8 because AutoKernType != NoAuto).
	noSlot := findLayerInComp(proj, "RE_TEXT24_PT", "pt_autokern_optical")
	if noSlot == nil {
		t.Fatalf("pt_autokern_optical missing")
	}
	if err := noSlot.SetManualKerning([]int{0, 0, 0, 0}); err == nil {
		t.Error("no-slot layer: expected error, got nil")
	}

	// Wrong-length values on a layer with an existing slot.
	withSlot := findLayerInComp(proj, "RE_TEXT24_PT", "pt_kerning_neg50")
	if withSlot == nil {
		t.Fatalf("pt_kerning_neg50 missing")
	}
	if err := withSlot.SetManualKerning([]int{1, 2}); err == nil {
		t.Error("wrong length: expected error, got nil")
	}

	// Non-text layer: pick any non-text layer in the project.
	var nonText *aep.Layer
	for _, c := range proj.Compositions {
		for _, lyr := range c.Layers {
			if lyr.TextSource == nil {
				nonText = lyr
				break
			}
		}
		if nonText != nil {
			break
		}
	}
	if nonText == nil {
		t.Skip("no non-text layer found in fixture")
	}
	if err := nonText.SetManualKerning([]int{0}); err == nil {
		t.Error("non-text layer: expected error, got nil")
	}
}

func TestFontAxesReal(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_varfont_ae24.aep")
	if err != nil {
		t.Skipf("re_varfont_ae24.aep not present; run test_data/re_varfont_ae24.jsx in AE: %v", err)
	}
	cases := []struct {
		layer       string
		fontPS      string
		wantAxis0   float64 // wght for Bahnschrift instances
		hasAxisData bool
	}{
		// Non-variable font: no /4 axes sub-array in btdk.
		{"vf_baseline", "TimesNewRomanPSMT", 0, false},
		// Bahnschrift default (wght=400, wdth=100); we only assert wght
		// because some Windows builds ship axis-list-of-1 (wght only).
		{"vf_default", "Bahnschrift", 400, true},
		// Bahnschrift-Bold static instance — AE writes wght=700 into /4.
		{"vf_bold", "Bahnschrift-Bold", 700, true},
	}
	for _, c := range cases {
		var l *aep.Layer
		for _, comp := range proj.Compositions {
			if found := layerByName(comp, c.layer); found != nil {
				l = found
				break
			}
		}
		if l == nil {
			t.Errorf("%s: layer missing", c.layer)
			continue
		}
		if l.TextSource == nil {
			t.Errorf("%s: TextSource nil", c.layer)
			continue
		}
		if len(l.TextSource.Fonts) == 0 || l.TextSource.Fonts[0] != c.fontPS {
			t.Errorf("%s Fonts[0] = %v, want [%s, ...]", c.layer, l.TextSource.Fonts, c.fontPS)
			continue
		}
		axes := l.TextSource.FontAxes
		if len(axes) != len(l.TextSource.Fonts) {
			t.Errorf("%s FontAxes len = %d, want %d (parallel to Fonts)", c.layer, len(axes), len(l.TextSource.Fonts))
			continue
		}
		got := axes[0]
		if c.hasAxisData {
			if len(got) == 0 {
				t.Errorf("%s FontAxes[0] empty, want axis data ([%g, ...])", c.layer, c.wantAxis0)
				continue
			}
			if math.Abs(got[0]-c.wantAxis0) > 0.01 {
				t.Errorf("%s FontAxes[0][0] = %g, want %g", c.layer, got[0], c.wantAxis0)
			}
		} else if len(got) != 0 {
			t.Errorf("%s FontAxes[0] = %v, want nil (non-variable font)", c.layer, got)
		}
	}
}
