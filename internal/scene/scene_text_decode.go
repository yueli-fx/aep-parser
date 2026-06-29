package scene

import (
	"fmt"
	"math"

	"github.com/yueli-fx/aep-parser/internal/codec"
)

// Text-layer "btds" payload is two stacked sub-LISTs:
//
//	LIST tdbs (property descriptor: tdsb / tdsn / tdb4 / cdat)
//	LIST btdk (CoolType-style PostScript text-engine document)
//
// The btdk body is a stream of `/key value` pairs at the top level
// (implicit dict), with `<<…>>` for nested dicts, `[…]` for arrays,
// `(…)` for strings, plus PostScript names (/Foo), numbers, and bools.
// Strings prefixed with the BOM `FE FF` are UTF-16BE.
//
// Reverse-engineered key paths (verified against AE 2020 fixtures —
// see test_data/generators/re_text.jsx — and AE 24+ fixtures —
// see test_data/generators/re_text_caps_ae24.jsx):
//
//	/0/1/0[i]/0/0/0      — font name (UTF-16BE) for font resource i
//	/1/1[0]/0/0          — paragraph text string (UTF-16BE, trailing CR per line)
//	/1/1[0]/0/5/0[i]/0/0/5/0   — paragraph justification (0=L, 1=R, 2=C)
//	/1/1[0]/0/5/0[i]/0/0/5/1   — paragraph firstLineIndent (AE 24+; float)
//	/1/1[0]/0/5/0[i]/0/0/5/2   — paragraph startIndent       (AE 24+; float)
//	/1/1[0]/0/5/0[i]/0/0/5/3   — paragraph endIndent         (AE 24+; float)
//	/1/1[0]/0/5/0[i]/0/0/5/4   — paragraph spaceBefore       (AE 24+; float)
//	/1/1[0]/0/5/0[i]/0/0/5/5   — paragraph spaceAfter        (AE 24+; float)
//	/1/1[0]/0/5/0[i]/0/0/5/9   — paragraph autoHyphenate     (AE 24+; bool, default true)
//	/1/1[0]/0/6/0[i]/0/0/6/0   — style-run font index (→ /0/1/0[idx])
//	/1/1[0]/0/6/0[i]/0/0/6/1   — style-run font size (em points)
//	/1/1[0]/0/6/0[i]/0/0/6/12  — style-run fontCapsOption     (AE 24+; 0=Normal,1=Small,2=All,3=AllSmall)
//	/1/1[0]/0/6/0[i]/0/0/6/13  — style-run fontBaselineOption (AE 24+; 0=Normal,1=Super,2=Sub)
//	/1/1[0]/0/6/0[i]/0/0/6/53/0/1   — style-run fill paint [A,R,G,B]
//	/1/1[0]/0/6/0[i]/0/0/6/58  — style-run strokeOverFill (bool, default true; AE 2020 readonly via JSX)

// DecodeTextSource parses a Layer.TextSourceRaw payload into a TextSource.
// Returns nil + a warning string if anything in the PostScript tree is
// missing or malformed; the raw bytes stay accessible via Layer.TextSourceRaw
// for downstream tools.
func DecodeTextSource(raw []byte) (*TextSource, string) {
	body, bodyOff, err := codec.ExtractBtdkBody(raw)
	if err != nil {
		return nil, fmt.Sprintf("text source: %v", err)
	}
	root := codec.ParsePSDict(body)
	if root == nil {
		return nil, "text source: btdk dict empty"
	}

	ts := &TextSource{}

	// Fonts table — array of font dicts at /0/1/0. Each entry's name is
	// at /0/0/0; variable-font axis values (when present) are a sub-array
	// at /0/0/4, encoded as 16.16 fixed-point integers (AE writes them as
	// large decimal numbers — e.g. 26214400 == 400 << 16 == wght=400).
	if fontArr := codec.PsPath(root, "/0/1/0"); fontArr != nil && fontArr.Kind == codec.PsArr {
		ts.FontAxes = make([][]float64, len(fontArr.Arr))
		for i, entry := range fontArr.Arr {
			name := codec.PsPathStr(entry, "/0/0/0")
			ts.Fonts = append(ts.Fonts, name)
			if axisArr := codec.PsPath(entry, "/0/0/4"); axisArr != nil && axisArr.Kind == codec.PsArr {
				axes := make([]float64, 0, len(axisArr.Arr))
				for _, v := range axisArr.Arr {
					if v.Kind != codec.PsNum {
						continue
					}
					axes = append(axes, v.Num/65536.0)
				}
				if len(axes) > 0 {
					ts.FontAxes[i] = axes
				}
			}
		}
	}

	// Text content — at /1/1[0]/0/0.
	if t := codec.PsPath(root, "/1/1/0/0/0"); t != nil && t.Kind == codec.PsStr {
		ts.Text = normalizeTextLines(t.Str)
		ts.textStringStart = bodyOff + t.SrcStart
		ts.textStringEnd = bodyOff + t.SrcEnd
	}

	// Paragraphs — array at /1/1[0]/0/5/0. Each entry's style is at /0/0/5.
	if paraArr := codec.PsPath(root, "/1/1/0/0/5/0"); paraArr != nil && paraArr.Kind == codec.PsArr {
		for _, entry := range paraArr.Arr {
			ts.Paragraphs = append(ts.Paragraphs, decodeParagraphStyle(entry))
		}
		if len(ts.Paragraphs) > 0 {
			ts.Justification = ts.Paragraphs[0].Justification
		}
	}

	// Style runs — array at /1/1[0]/0/6/0
	if runArr := codec.PsPath(root, "/1/1/0/0/6/0"); runArr != nil && runArr.Kind == codec.PsArr {
		for _, entry := range runArr.Arr {
			run := decodeStyleRun(entry, ts.Fonts)
			ts.Runs = append(ts.Runs, run)
		}
	}

	// Box-text bounds — point-text omits /0/8/0[0]/0/1. The /1/0 sub-array
	// holds the bounds polygon as 16 (x,y) vertex pairs (= 32 numbers);
	// the rectangle's xmin/ymin/xmax/ymax come from min/max over those.
	if bounds := codec.PsPath(root, "/0/8/0/0/0/1/0"); bounds != nil && bounds.Kind == codec.PsArr && len(bounds.Arr) >= 8 {
		ts.IsBoxText = true
		ts.BoxBounds = computeBounds(bounds)
	}

	// Manual kerning — per-character integer array at /1/1[0]/0/8/0.
	// Only present when some run has AutoKernType == NoAuto and a
	// kerning value was actually set. Last array entry is an
	// end-of-text sentinel (empty /0 dict); skip it. Sibling /1/1[0]/0/7
	// is the AE-script first-char mirror.
	if kernArr := codec.PsPath(root, "/1/1/0/0/8/0"); kernArr != nil && kernArr.Kind == codec.PsArr {
		for i, entry := range kernArr.Arr {
			v := codec.PsPath(entry, "/0/0")
			if v == nil || v.Kind != codec.PsNum {
				// Sentinel allowed only as the very last entry.
				if i == len(kernArr.Arr)-1 {
					break
				}
				continue
			}
			ts.ManualKerning = append(ts.ManualKerning, int(v.Num))
		}
	}
	if v := codec.PsPath(root, "/1/1/0/0/7"); v != nil && v.Kind == codec.PsNum {
		ts.Kerning = int(v.Num)
	}

	return ts, ""
}

// decodeStyleRun extracts one entry of /1/1[0]/0/6/0[i] into a TextStyleRun.
func decodeStyleRun(entry *codec.PsValue, fonts []string) TextStyleRun {
	r := TextStyleRun{FontIndex: -1}
	// Style block lives at entry/0/0/6 — pull it once and read all
	// per-run keys off it.
	style := codec.PsPath(entry, "/0/0/6")
	if style == nil || style.Kind != codec.PsDict {
		return r
	}
	if v := codec.PsStep(style, "0"); v != nil && v.Kind == codec.PsNum {
		r.FontIndex = int(v.Num)
	}
	if v := codec.PsStep(style, "1"); v != nil && v.Kind == codec.PsNum {
		r.FontSize = v.Num
	}
	if r.FontIndex >= 0 && r.FontIndex < len(fonts) {
		r.FontName = fonts[r.FontIndex]
	}
	if v := codec.PsStep(style, "2"); v != nil && v.Kind == codec.PsBool {
		r.FauxBold = v.Bv
	}
	if v := codec.PsStep(style, "3"); v != nil && v.Kind == codec.PsBool {
		r.FauxItalic = v.Bv
	}
	if v := codec.PsStep(style, "4"); v != nil && v.Kind == codec.PsBool {
		r.AutoLeading = v.Bv
	}
	if v := codec.PsStep(style, "5"); v != nil && v.Kind == codec.PsNum {
		r.Leading = v.Num
	}
	if v := codec.PsStep(style, "6"); v != nil && v.Kind == codec.PsNum {
		r.HorizontalScale = v.Num
	}
	if v := codec.PsStep(style, "7"); v != nil && v.Kind == codec.PsNum {
		r.VerticalScale = v.Num
	}
	if v := codec.PsStep(style, "8"); v != nil && v.Kind == codec.PsNum {
		r.Tracking = v.Num
	}
	if v := codec.PsStep(style, "9"); v != nil && v.Kind == codec.PsNum {
		r.BaselineShift = v.Num
	}
	if v := codec.PsStep(style, "36"); v != nil && v.Kind == codec.PsNum {
		r.Tsume = v.Num
	}
	r.FillColor = decodePaintColor(codec.PsPath(style, "/53"))
	r.StrokeColor = decodePaintColor(codec.PsPath(style, "/54"))
	if v := codec.PsStep(style, "57"); v != nil && v.Kind == codec.PsBool {
		r.ApplyStroke = v.Bv
	}
	if v := codec.PsStep(style, "63"); v != nil && v.Kind == codec.PsNum {
		r.StrokeWidth = v.Num
	}
	// AE 24+ fields. Default values mirror what AE writes for a fresh
	// run: caps/baseline both Normal (0), strokeOverFill true,
	// autoKernType Metric (1), noBreak false, lineJoin Miter (0),
	// digitSet Default (0).
	r.StrokeOverFill = true
	r.AutoKernType = TextAutoKernMetric
	if v := codec.PsStep(style, "11"); v != nil && v.Kind == codec.PsNum {
		r.AutoKernType = TextAutoKernType(int(v.Num))
	}
	if v := codec.PsStep(style, "12"); v != nil && v.Kind == codec.PsNum {
		r.CapsOption = TextCapsOption(int(v.Num))
	}
	if v := codec.PsStep(style, "13"); v != nil && v.Kind == codec.PsNum {
		r.BaselineOption = TextBaselineOption(int(v.Num))
	}
	if v := codec.PsStep(style, "52"); v != nil && v.Kind == codec.PsBool {
		r.NoBreak = v.Bv
	}
	if v := codec.PsStep(style, "58"); v != nil && v.Kind == codec.PsBool {
		r.StrokeOverFill = v.Bv
	}
	if v := codec.PsStep(style, "62"); v != nil && v.Kind == codec.PsNum {
		r.LineJoinType = TextLineJoinType(int(v.Num))
	}
	if v := codec.PsStep(style, "70"); v != nil && v.Kind == codec.PsNum {
		r.DigitSet = TextDigitSet(int(v.Num))
	}
	return r
}

// decodeParagraphStyle extracts one entry of /1/1[0]/0/5/0[i] into a
// TextParagraph. Paragraph style block lives at entry/0/0/5.
func decodeParagraphStyle(entry *codec.PsValue) TextParagraph {
	p := TextParagraph{AutoHyphenate: true}
	style := codec.PsPath(entry, "/0/0/5")
	if style == nil || style.Kind != codec.PsDict {
		return p
	}
	if v := codec.PsStep(style, "0"); v != nil && v.Kind == codec.PsNum {
		p.Justification = TextJustification(int(v.Num))
	}
	if v := codec.PsStep(style, "1"); v != nil && v.Kind == codec.PsNum {
		p.FirstLineIndent = v.Num
	}
	if v := codec.PsStep(style, "2"); v != nil && v.Kind == codec.PsNum {
		p.StartIndent = v.Num
	}
	if v := codec.PsStep(style, "3"); v != nil && v.Kind == codec.PsNum {
		p.EndIndent = v.Num
	}
	if v := codec.PsStep(style, "4"); v != nil && v.Kind == codec.PsNum {
		p.SpaceBefore = v.Num
	}
	if v := codec.PsStep(style, "5"); v != nil && v.Kind == codec.PsNum {
		p.SpaceAfter = v.Num
	}
	if v := codec.PsStep(style, "8"); v != nil && v.Kind == codec.PsNum {
		p.LeadingType = TextLeadingType(int(v.Num))
	}
	if v := codec.PsStep(style, "9"); v != nil && v.Kind == codec.PsBool {
		p.AutoHyphenate = v.Bv
	}
	if v := codec.PsStep(style, "21"); v != nil && v.Kind == codec.PsBool {
		p.HangingRoman = v.Bv
	}
	if v := codec.PsStep(style, "33"); v != nil && v.Kind == codec.PsNum {
		p.Direction = TextParagraphDirection(int(v.Num))
	}
	return p
}

// computeBounds reduces a polygon of (x, y) pairs to its axis-aligned
// bounding box [xmin, ymin, xmax, ymax]. The btdk box-text bounds
// array stores the rectangle as a 16-vertex polygon (4 corners ×
// 4 anchor/tangent slots each), so we don't try to reconstruct the
// individual corners — just snap to the AABB.
func computeBounds(arr *codec.PsValue) [4]float64 {
	xmin, ymin := math.Inf(1), math.Inf(1)
	xmax, ymax := math.Inf(-1), math.Inf(-1)
	for i := 0; i+1 < len(arr.Arr); i += 2 {
		x := arr.Arr[i].AsNum()
		y := arr.Arr[i+1].AsNum()
		if x < xmin {
			xmin = x
		}
		if x > xmax {
			xmax = x
		}
		if y < ymin {
			ymin = y
		}
		if y > ymax {
			ymax = y
		}
	}
	if math.IsInf(xmin, 1) {
		return [4]float64{}
	}
	return [4]float64{xmin, ymin, xmax, ymax}
}

// decodePaintColor extracts the [R, G, B, A] color from a paint dict
// shaped `<< /99 /SimplePaint /0 << /0 1 /1 [A R G B] >> >>`. Returns
// the zero array if the dict is missing or shaped differently (e.g.
// when AE uses a non-SimplePaint entry for gradients — not yet seen).
func decodePaintColor(paint *codec.PsValue) [4]float64 {
	arr := codec.PsPath(paint, "/0/1")
	if arr == nil || arr.Kind != codec.PsArr || len(arr.Arr) != 4 {
		return [4]float64{}
	}
	// Source order is [A, R, G, B]; expose as [R, G, B, A].
	a := arr.Arr[0].AsNum()
	rr := arr.Arr[1].AsNum()
	g := arr.Arr[2].AsNum()
	b := arr.Arr[3].AsNum()
	return [4]float64{rr, g, b, a}
}

// normalizeTextLines strips AE's trailing CR-terminator from each
// paragraph and joins lines with LF for Go-friendly multi-line text.
func normalizeTextLines(s string) string {
	// AE stores each paragraph terminated with \r. Trim one trailing \r
	// if present and convert remaining \r to \n.
	if len(s) > 0 && s[len(s)-1] == '\r' {
		s = s[:len(s)-1]
	}
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if r == '\r' {
			out = append(out, '\n')
		} else {
			out = append(out, r)
		}
	}
	return string(out)
}
