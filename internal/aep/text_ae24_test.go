package aep_test

import (
	"bytes"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestTextAE24MoreRunFieldsReal(t *testing.T) {
	proj := openTextAE24More(t)
	cases := []struct {
		layer string
		check func(t *testing.T, r aep.TextStyleRun)
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
