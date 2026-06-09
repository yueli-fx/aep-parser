package scene

import (
	"fmt"
)

// TextJustification matches AE's paragraph alignment enum.
type TextJustification int

const (
	TextJustifyLeft   TextJustification = 0
	TextJustifyRight  TextJustification = 1
	TextJustifyCenter TextJustification = 2
)

// String returns the human-readable name.
func (j TextJustification) String() string {
	switch j {
	case TextJustifyLeft:
		return "Left"
	case TextJustifyRight:
		return "Right"
	case TextJustifyCenter:
		return "Center"
	}
	return fmt.Sprintf("Justify(%d)", int(j))
}

// TextCapsOption mirrors AE's FontCapsOption enum (AE 24+).
// In AE 2020 scripting this was readonly via allCaps / smallCaps; AE 24
// exposed write access through fontCapsOption. Stored at btdk
// /1/1[0]/0/6/0[i]/0/0/6/12.
type TextCapsOption int

const (
	TextCapsNormal   TextCapsOption = 0
	TextCapsSmall    TextCapsOption = 1
	TextCapsAll      TextCapsOption = 2
	TextCapsAllSmall TextCapsOption = 3
)

// String returns the human-readable name.
func (c TextCapsOption) String() string {
	switch c {
	case TextCapsNormal:
		return "Normal"
	case TextCapsSmall:
		return "SmallCaps"
	case TextCapsAll:
		return "AllCaps"
	case TextCapsAllSmall:
		return "AllSmallCaps"
	}
	return fmt.Sprintf("Caps(%d)", int(c))
}

// TextAutoKernType mirrors AE's AutoKernType enum (AE 24+ exposed,
// underlying btdk field present in AE 2020+). Stored at btdk
// /1/1[0]/0/6/0[i]/0/0/6/11. When set to NoAutoKern, the manual kerning
// value lives in a separate per-character sub-tree at /1/1[0]/0/8 which
// this library does not yet decode.
type TextAutoKernType int

const (
	TextAutoKernNoAuto  TextAutoKernType = 0
	TextAutoKernMetric  TextAutoKernType = 1
	TextAutoKernOptical TextAutoKernType = 2
)

// String returns the human-readable name.
func (k TextAutoKernType) String() string {
	switch k {
	case TextAutoKernNoAuto:
		return "NoAuto"
	case TextAutoKernMetric:
		return "Metric"
	case TextAutoKernOptical:
		return "Optical"
	}
	return fmt.Sprintf("AutoKern(%d)", int(k))
}

// TextLineJoinType mirrors AE's LineJoinType enum (AE 24+ writeable).
// Controls stroke corner joining for text outlines. btdk style-run /62.
type TextLineJoinType int

const (
	TextLineJoinMiter TextLineJoinType = 0
	TextLineJoinRound TextLineJoinType = 1
	TextLineJoinBevel TextLineJoinType = 2
)

// String returns the human-readable name.
func (j TextLineJoinType) String() string {
	switch j {
	case TextLineJoinMiter:
		return "Miter"
	case TextLineJoinRound:
		return "Round"
	case TextLineJoinBevel:
		return "Bevel"
	}
	return fmt.Sprintf("LineJoin(%d)", int(j))
}

// TextDigitSet mirrors AE's DigitSet enum (AE 24+ writeable). btdk
// style-run /70. Only values 0..2 confirmed via fixture; Farsi/ArabicRTL
// follow AE's enum order per documentation.
type TextDigitSet int

const (
	TextDigitSetDefault   TextDigitSet = 0
	TextDigitSetArabic    TextDigitSet = 1
	TextDigitSetHindi     TextDigitSet = 2
	TextDigitSetFarsi     TextDigitSet = 3 // per AE docs; not fixture-verified
	TextDigitSetArabicRTL TextDigitSet = 4 // per AE docs; not fixture-verified
)

// String returns the human-readable name.
func (d TextDigitSet) String() string {
	switch d {
	case TextDigitSetDefault:
		return "Default"
	case TextDigitSetArabic:
		return "Arabic"
	case TextDigitSetHindi:
		return "Hindi"
	case TextDigitSetFarsi:
		return "Farsi"
	case TextDigitSetArabicRTL:
		return "ArabicRTL"
	}
	return fmt.Sprintf("DigitSet(%d)", int(d))
}

// TextLeadingType mirrors AE's LeadingType enum (AE 24+ writeable).
// Paragraph-level field at btdk /1/1[0]/0/5/0[i]/0/0/5/8.
type TextLeadingType int

const (
	TextLeadingRoman    TextLeadingType = 0
	TextLeadingJapanese TextLeadingType = 1
)

// String returns the human-readable name.
func (t TextLeadingType) String() string {
	switch t {
	case TextLeadingRoman:
		return "Roman"
	case TextLeadingJapanese:
		return "Japanese"
	}
	return fmt.Sprintf("LeadingType(%d)", int(t))
}

// TextParagraphDirection mirrors AE's ParagraphDirection enum (AE 24+
// writeable). Paragraph-level field at btdk /1/1[0]/0/5/0[i]/0/0/5/33.
type TextParagraphDirection int

const (
	TextDirectionLeftToRight TextParagraphDirection = 0
	TextDirectionRightToLeft TextParagraphDirection = 1
)

// String returns the human-readable name.
func (d TextParagraphDirection) String() string {
	switch d {
	case TextDirectionLeftToRight:
		return "LTR"
	case TextDirectionRightToLeft:
		return "RTL"
	}
	return fmt.Sprintf("Direction(%d)", int(d))
}

// TextBaselineOption mirrors AE's FontBaselineOption enum (AE 24+).
// In AE 2020 scripting this was readonly via subscript / superscript;
// AE 24 exposed write access through fontBaselineOption. Stored at
// btdk /1/1[0]/0/6/0[i]/0/0/6/13.
type TextBaselineOption int

const (
	TextBaselineNormal      TextBaselineOption = 0
	TextBaselineSuperscript TextBaselineOption = 1
	TextBaselineSubscript   TextBaselineOption = 2
)

// String returns the human-readable name.
func (b TextBaselineOption) String() string {
	switch b {
	case TextBaselineNormal:
		return "Normal"
	case TextBaselineSuperscript:
		return "Superscript"
	case TextBaselineSubscript:
		return "Subscript"
	}
	return fmt.Sprintf("Baseline(%d)", int(b))
}

// TextStyleRun is one character-level style span inside a TextSource.
// Multi-style text (different size/color per word) has multiple runs;
// uniform text has exactly one. Character index spans are not yet RE'd —
// runs appear in document order matching AE's paragraph palette.
type TextStyleRun struct {
	FontIndex       int                // index into TextSource.Fonts
	FontName        string             // resolved Fonts[FontIndex] for convenience; "" if out of range
	FontSize        float64            // em points (AE's "Font Size" field)
	FillColor       [4]float64         // [R, G, B, A], each 0..1
	FauxBold        bool               // run-level synthetic bold (true when font lacks a bold cut)
	FauxItalic      bool               // run-level synthetic italic
	AutoLeading     bool               // true → Leading is AE's auto value (FontSize × 1.2 typical)
	Leading         float64            // em points; only meaningful when AutoLeading is false
	Tracking        float64            // 1/1000 em (AE's tracking field; 0 = normal)
	BaselineShift   float64            // em points; positive = up, negative = down
	HorizontalScale float64            // raw value (AE default 1, scripted range 0..100; unit not fully RE'd — pass through)
	VerticalScale   float64            // raw value (AE default 1, scripted range 0..100)
	Tsume           float64            // CJK character-spacing adjustment (0..100, AE default 0)
	ApplyStroke     bool               // whether StrokeColor/StrokeWidth are active
	StrokeColor     [4]float64         // [R, G, B, A]
	StrokeWidth     float64            // em points
	CapsOption      TextCapsOption     // AE 24+ writeable; mirrors AE's allCaps / smallCaps readonly attrs
	BaselineOption  TextBaselineOption // AE 24+ writeable; mirrors AE's subscript / superscript readonly attrs
	StrokeOverFill  bool               // stroke renders over fill (default true). AE 2020 readonly via JSX
	AutoKernType    TextAutoKernType   // AE 24+ writeable; auto-kerning mode (NoAuto / Metric / Optical)
	NoBreak         bool               // AE 24+ writeable; "do not break" character flag
	LineJoinType    TextLineJoinType   // AE 24+ writeable; stroke corner join style
	DigitSet        TextDigitSet       // AE 24+ writeable; digit set (Default / Arabic / Hindi / ...)
}

// AllCaps mirrors AE's readonly TextDocument.allCaps attribute. True
// when CapsOption is TextCapsAll or TextCapsAllSmall.
func (r TextStyleRun) AllCaps() bool {
	return r.CapsOption == TextCapsAll || r.CapsOption == TextCapsAllSmall
}

// SmallCaps mirrors AE's readonly TextDocument.smallCaps attribute.
// True when CapsOption is TextCapsSmall or TextCapsAllSmall.
func (r TextStyleRun) SmallCaps() bool {
	return r.CapsOption == TextCapsSmall || r.CapsOption == TextCapsAllSmall
}

// Superscript mirrors AE's readonly TextDocument.superscript attribute.
func (r TextStyleRun) Superscript() bool {
	return r.BaselineOption == TextBaselineSuperscript
}

// Subscript mirrors AE's readonly TextDocument.subscript attribute.
func (r TextStyleRun) Subscript() bool {
	return r.BaselineOption == TextBaselineSubscript
}

// TextParagraph is one paragraph's style (split by '\r' inside the
// btdk text string). Indent / space / autoHyphenate fields are AE 24+
// writeable (the BTDK chunk has always stored them; AE 2020 ScriptingAPI
// just couldn't toggle them on point text).
type TextParagraph struct {
	Justification   TextJustification      // /0
	FirstLineIndent float64                // /1 — em points
	StartIndent     float64                // /2 — em points
	EndIndent       float64                // /3 — em points
	SpaceBefore     float64                // /4 — em points
	SpaceAfter      float64                // /5 — em points
	LeadingType     TextLeadingType        // /8 — AE 24+ (0=Roman, 1=Japanese)
	AutoHyphenate   bool                   // /9 — default true
	HangingRoman    bool                   // /21 — AE 24+ (box-text-only meaningful)
	Direction       TextParagraphDirection // /33 — AE 24+ (0=LTR, 1=RTL)
}

// TextSource is the decoded text-layer document. Populated when the
// btds payload parses cleanly; nil otherwise (raw bytes remain on
// Layer.TextSourceRaw).
type TextSource struct {
	Text  string   // concatenated paragraph text, \r → \n, trailing newline trimmed
	Fonts []string // PostScript font names from the resource table
	// FontAxes is parallel to Fonts: FontAxes[i] holds the OpenType design-axis
	// values for Fonts[i] when it's a variable font. The slice is in the same
	// order AE wrote it (matches the font's named axes — Bahnschrift =
	// [wght, wdth], Inter Variable = [wght, slnt], etc.). Nil for non-variable
	// fonts (most fonts; the /4 sub-array is absent in the btdk font dict).
	// Values are decoded from 16.16 fixed-point into floats — e.g. wght=700
	// reads as 700.0. Read-only; axis values are tied to font-instance
	// identity (AE rewrites them when you swap to a different instance via
	// SetRunFontIndex / AddFont), so changing axes independently is not
	// exposed.
	FontAxes      [][]float64
	Runs          []TextStyleRun    // character style runs in document order
	Paragraphs    []TextParagraph   // per-paragraph styling in document order
	Justification TextJustification // first paragraph's justification (== Paragraphs[0].Justification when present); kept for convenience
	IsBoxText     bool              // true → box (paragraph) text with explicit bounds; false → point text
	BoxBounds     [4]float64        // [xmin, ymin, xmax, ymax] in layer-local coords; zero when !IsBoxText

	// ManualKerning is the per-character manual kerning values in 1/1000
	// em units (AE's "Kerning" panel field; only meaningful when the
	// matching run's AutoKernType == TextAutoKernNoAuto). Length matches
	// the character count of Text — one entry per character. Nil / empty
	// when no manual kerning was ever applied (autoKernType is
	// Metric/Optical for all chars). Decoded from the btdk per-char
	// run-length array at /1/1[0]/0/8/0 (each entry has /0/0 = value,
	// /1 = run length, AE always writes 1; the trailing entry is an
	// end-of-text sentinel with an empty /0 dict and is excluded here).
	//
	// AE script's TextDocument.kerning getter is the equivalent of
	// ManualKerning[0] — see the Kerning field for the mirrored
	// first-char value.
	ManualKerning []int

	// Kerning mirrors AE script's TextDocument.kerning ("only reflects
	// the first character"). Sourced from /1/1[0]/0/7 (a scalar sibling
	// of /8 — AE keeps it in sync with /8/0[0] for convenience). Zero
	// when the kerning sub-tree is absent (autoKernType is Metric or
	// Optical across the document).
	Kerning int

	// textStringStart / textStringEnd bracket the encoded PostScript
	// text string (including parens, BOM, terminator) inside the parent
	// Layer.TextSourceRaw bytes. Used by Layer.SetText to splice a
	// replacement in-place. Both zero if decoding couldn't locate the
	// string (SetText then refuses with an error).
	textStringStart, textStringEnd int
}
