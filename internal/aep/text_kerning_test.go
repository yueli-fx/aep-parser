package aep_test

import (
	"bytes"
	"math"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestTextManualKerningReal(t *testing.T) {
	proj := openTextAE24More(t)
	cases := []struct {
		layer       string
		wantKerning int
		wantManual  []int
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
