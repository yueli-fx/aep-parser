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
