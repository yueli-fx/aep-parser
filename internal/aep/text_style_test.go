package aep_test

import (
	"bytes"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

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
