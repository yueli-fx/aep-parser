package aep

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/example/aep-parser/internal/rifx"
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
// see test_data/re_text.jsx — and AE 24+ fixtures —
// see test_data/re_text_caps_ae24.jsx):
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

func decodeTextSource(raw []byte) (*TextSource, string) {
	body, bodyOff, err := extractBtdkBody(raw)
	if err != nil {
		return nil, fmt.Sprintf("text source: %v", err)
	}
	root := parsePSDict(body)
	if root == nil {
		return nil, "text source: btdk dict empty"
	}

	ts := &TextSource{}

	// Fonts table — array of font dicts at /0/1/0. Each entry's name is
	// at /0/0/0; variable-font axis values (when present) are a sub-array
	// at /0/0/4, encoded as 16.16 fixed-point integers (AE writes them as
	// large decimal numbers — e.g. 26214400 == 400 << 16 == wght=400).
	if fontArr := psPath(root, "/0/1/0"); fontArr != nil && fontArr.kind == psArr {
		ts.FontAxes = make([][]float64, len(fontArr.arr))
		for i, entry := range fontArr.arr {
			name := psPathStr(entry, "/0/0/0")
			ts.Fonts = append(ts.Fonts, name)
			if axisArr := psPath(entry, "/0/0/4"); axisArr != nil && axisArr.kind == psArr {
				axes := make([]float64, 0, len(axisArr.arr))
				for _, v := range axisArr.arr {
					if v.kind != psNum {
						continue
					}
					axes = append(axes, v.num/65536.0)
				}
				if len(axes) > 0 {
					ts.FontAxes[i] = axes
				}
			}
		}
	}

	// Text content — at /1/1[0]/0/0.
	if t := psPath(root, "/1/1/0/0/0"); t != nil && t.kind == psStr {
		ts.Text = normalizeTextLines(t.str)
		ts.textStringStart = bodyOff + t.srcStart
		ts.textStringEnd = bodyOff + t.srcEnd
	}

	// Paragraphs — array at /1/1[0]/0/5/0. Each entry's style is at /0/0/5.
	if paraArr := psPath(root, "/1/1/0/0/5/0"); paraArr != nil && paraArr.kind == psArr {
		for _, entry := range paraArr.arr {
			ts.Paragraphs = append(ts.Paragraphs, decodeParagraphStyle(entry))
		}
		if len(ts.Paragraphs) > 0 {
			ts.Justification = ts.Paragraphs[0].Justification
		}
	}

	// Style runs — array at /1/1[0]/0/6/0
	if runArr := psPath(root, "/1/1/0/0/6/0"); runArr != nil && runArr.kind == psArr {
		for _, entry := range runArr.arr {
			run := decodeStyleRun(entry, ts.Fonts)
			ts.Runs = append(ts.Runs, run)
		}
	}

	// Box-text bounds — point-text omits /0/8/0[0]/0/1. The /1/0 sub-array
	// holds the bounds polygon as 16 (x,y) vertex pairs (= 32 numbers);
	// the rectangle's xmin/ymin/xmax/ymax come from min/max over those.
	if bounds := psPath(root, "/0/8/0/0/0/1/0"); bounds != nil && bounds.kind == psArr && len(bounds.arr) >= 8 {
		ts.IsBoxText = true
		ts.BoxBounds = computeBounds(bounds)
	}

	// Manual kerning — per-character integer array at /1/1[0]/0/8/0.
	// Only present when some run has AutoKernType == NoAuto and a
	// kerning value was actually set. Last array entry is an
	// end-of-text sentinel (empty /0 dict); skip it. Sibling /1/1[0]/0/7
	// is the AE-script first-char mirror.
	if kernArr := psPath(root, "/1/1/0/0/8/0"); kernArr != nil && kernArr.kind == psArr {
		for i, entry := range kernArr.arr {
			v := psPath(entry, "/0/0")
			if v == nil || v.kind != psNum {
				// Sentinel allowed only as the very last entry.
				if i == len(kernArr.arr)-1 {
					break
				}
				continue
			}
			ts.ManualKerning = append(ts.ManualKerning, int(v.num))
		}
	}
	if v := psPath(root, "/1/1/0/0/7"); v != nil && v.kind == psNum {
		ts.Kerning = int(v.num)
	}

	return ts, ""
}

// decodeStyleRun extracts one entry of /1/1[0]/0/6/0[i] into a TextStyleRun.
func decodeStyleRun(entry *psValue, fonts []string) TextStyleRun {
	r := TextStyleRun{FontIndex: -1}
	// Style block lives at entry/0/0/6 — pull it once and read all
	// per-run keys off it.
	style := psPath(entry, "/0/0/6")
	if style == nil || style.kind != psDict {
		return r
	}
	if v := psStep(style, "0"); v != nil && v.kind == psNum {
		r.FontIndex = int(v.num)
	}
	if v := psStep(style, "1"); v != nil && v.kind == psNum {
		r.FontSize = v.num
	}
	if r.FontIndex >= 0 && r.FontIndex < len(fonts) {
		r.FontName = fonts[r.FontIndex]
	}
	if v := psStep(style, "2"); v != nil && v.kind == psBool {
		r.FauxBold = v.bv
	}
	if v := psStep(style, "3"); v != nil && v.kind == psBool {
		r.FauxItalic = v.bv
	}
	if v := psStep(style, "4"); v != nil && v.kind == psBool {
		r.AutoLeading = v.bv
	}
	if v := psStep(style, "5"); v != nil && v.kind == psNum {
		r.Leading = v.num
	}
	if v := psStep(style, "6"); v != nil && v.kind == psNum {
		r.HorizontalScale = v.num
	}
	if v := psStep(style, "7"); v != nil && v.kind == psNum {
		r.VerticalScale = v.num
	}
	if v := psStep(style, "8"); v != nil && v.kind == psNum {
		r.Tracking = v.num
	}
	if v := psStep(style, "9"); v != nil && v.kind == psNum {
		r.BaselineShift = v.num
	}
	if v := psStep(style, "36"); v != nil && v.kind == psNum {
		r.Tsume = v.num
	}
	r.FillColor = decodePaintColor(psPath(style, "/53"))
	r.StrokeColor = decodePaintColor(psPath(style, "/54"))
	if v := psStep(style, "57"); v != nil && v.kind == psBool {
		r.ApplyStroke = v.bv
	}
	if v := psStep(style, "63"); v != nil && v.kind == psNum {
		r.StrokeWidth = v.num
	}
	// AE 24+ fields. Default values mirror what AE writes for a fresh
	// run: caps/baseline both Normal (0), strokeOverFill true,
	// autoKernType Metric (1), noBreak false, lineJoin Miter (0),
	// digitSet Default (0).
	r.StrokeOverFill = true
	r.AutoKernType = TextAutoKernMetric
	if v := psStep(style, "11"); v != nil && v.kind == psNum {
		r.AutoKernType = TextAutoKernType(int(v.num))
	}
	if v := psStep(style, "12"); v != nil && v.kind == psNum {
		r.CapsOption = TextCapsOption(int(v.num))
	}
	if v := psStep(style, "13"); v != nil && v.kind == psNum {
		r.BaselineOption = TextBaselineOption(int(v.num))
	}
	if v := psStep(style, "52"); v != nil && v.kind == psBool {
		r.NoBreak = v.bv
	}
	if v := psStep(style, "58"); v != nil && v.kind == psBool {
		r.StrokeOverFill = v.bv
	}
	if v := psStep(style, "62"); v != nil && v.kind == psNum {
		r.LineJoinType = TextLineJoinType(int(v.num))
	}
	if v := psStep(style, "70"); v != nil && v.kind == psNum {
		r.DigitSet = TextDigitSet(int(v.num))
	}
	return r
}

// decodeParagraphStyle extracts one entry of /1/1[0]/0/5/0[i] into a
// TextParagraph. Paragraph style block lives at entry/0/0/5.
func decodeParagraphStyle(entry *psValue) TextParagraph {
	p := TextParagraph{AutoHyphenate: true}
	style := psPath(entry, "/0/0/5")
	if style == nil || style.kind != psDict {
		return p
	}
	if v := psStep(style, "0"); v != nil && v.kind == psNum {
		p.Justification = TextJustification(int(v.num))
	}
	if v := psStep(style, "1"); v != nil && v.kind == psNum {
		p.FirstLineIndent = v.num
	}
	if v := psStep(style, "2"); v != nil && v.kind == psNum {
		p.StartIndent = v.num
	}
	if v := psStep(style, "3"); v != nil && v.kind == psNum {
		p.EndIndent = v.num
	}
	if v := psStep(style, "4"); v != nil && v.kind == psNum {
		p.SpaceBefore = v.num
	}
	if v := psStep(style, "5"); v != nil && v.kind == psNum {
		p.SpaceAfter = v.num
	}
	if v := psStep(style, "8"); v != nil && v.kind == psNum {
		p.LeadingType = TextLeadingType(int(v.num))
	}
	if v := psStep(style, "9"); v != nil && v.kind == psBool {
		p.AutoHyphenate = v.bv
	}
	if v := psStep(style, "21"); v != nil && v.kind == psBool {
		p.HangingRoman = v.bv
	}
	if v := psStep(style, "33"); v != nil && v.kind == psNum {
		p.Direction = TextParagraphDirection(int(v.num))
	}
	return p
}

// computeBounds reduces a polygon of (x, y) pairs to its axis-aligned
// bounding box [xmin, ymin, xmax, ymax]. The btdk box-text bounds
// array stores the rectangle as a 16-vertex polygon (4 corners ×
// 4 anchor/tangent slots each), so we don't try to reconstruct the
// individual corners — just snap to the AABB.
func computeBounds(arr *psValue) [4]float64 {
	xmin, ymin := math.Inf(1), math.Inf(1)
	xmax, ymax := math.Inf(-1), math.Inf(-1)
	for i := 0; i+1 < len(arr.arr); i += 2 {
		x := arr.arr[i].asNum()
		y := arr.arr[i+1].asNum()
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
func decodePaintColor(paint *psValue) [4]float64 {
	arr := psPath(paint, "/0/1")
	if arr == nil || arr.kind != psArr || len(arr.arr) != 4 {
		return [4]float64{}
	}
	// Source order is [A, R, G, B]; expose as [R, G, B, A].
	a := arr.arr[0].asNum()
	rr := arr.arr[1].asNum()
	g := arr.arr[2].asNum()
	b := arr.arr[3].asNum()
	return [4]float64{rr, g, b, a}
}

// extractBtdkBody walks the raw btds payload bytes and returns the
// inner btdk LIST's body (PostScript text) plus that body's starting
// offset inside raw (so callers can translate body-local offsets back
// to TextSourceRaw offsets — Layer.SetText needs this). The payload
// begins with a "LIST tdbs" sub-chunk holding the property descriptor;
// the btdk LIST follows. Both follow the regular RIFX LIST header
// layout (LIST + uint32 BE size + 4-byte formType + payload).
func extractBtdkBody(raw []byte) ([]byte, int, error) {
	for off := 0; off+8 <= len(raw); {
		if string(raw[off:off+4]) != "LIST" {
			return nil, 0, fmt.Errorf("expected LIST at %#x, got %q", off, raw[off:off+4])
		}
		size := binary.BigEndian.Uint32(raw[off+4 : off+8])
		if int(size) < 4 || off+8+int(size)-4 > len(raw) {
			return nil, 0, fmt.Errorf("LIST at %#x has bad size %d", off, size)
		}
		formType := rifx.ChunkID{raw[off+8], raw[off+9], raw[off+10], raw[off+11]}
		bodyStart := off + 12
		bodyEnd := off + 8 + int(size)
		if formType == rifx.IDBtdk {
			return raw[bodyStart:bodyEnd], bodyStart, nil
		}
		off = bodyEnd
		// LIST chunks have implicit pad to even length.
		if size%2 != 0 {
			off++
		}
	}
	return nil, 0, fmt.Errorf("no btdk LIST in btds payload")
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
