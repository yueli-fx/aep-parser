package aep

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
)

// Text per-run write API. Each setter targets one psValue node inside
// the btdk PostScript dict and splices new bytes in. The btdk chunk
// (and ancestor LIST sizes) are length-variable — WriteAEP recomputes
// the parent sizes from chunk Data lengths.
//
// Implementation pattern:
//   1. Locate the target value's byte range via fresh extract+parse of
//      the btdk body (offsets recorded by parsePSValue).
//   2. Splice new bytes into Layer.btdsChunk.Data.
//   3. Re-decode the entire btds payload so Layer.TextSource reflects
//      the new state.
//
// Setters are placed on Layer (not TextSource or TextStyleRun) to keep
// the call site discoverable alongside SetText.

// splicePSValue locates the psValue at `path` (relative to the btdk
// root dict, e.g. "/1/1/0/0/6/0/0/0/0/6/1") and replaces its on-disk
// bytes with newSrc. Returns an error if the path doesn't resolve or
// the layer has no decoded text source.
func (l *Layer) splicePSValue(path string, newSrc []byte) error {
	if l.btdsChunk == nil {
		return fmt.Errorf("layer %q: not a text layer", l.Name)
	}
	body, bodyOff, err := extractBtdkBody(l.btdsChunk.Data)
	if err != nil {
		return fmt.Errorf("layer %q: %w", l.Name, err)
	}
	root := parsePSDict(body)
	if root == nil {
		return fmt.Errorf("layer %q: btdk dict empty", l.Name)
	}
	target := psPath(root, path)
	if target == nil {
		return fmt.Errorf("layer %q: no psValue at %q", l.Name, path)
	}
	if target.srcEnd <= target.srcStart {
		return fmt.Errorf("layer %q: target at %q has zero-width src range", l.Name, path)
	}
	absStart := bodyOff + target.srcStart
	absEnd := bodyOff + target.srcEnd
	delta := len(newSrc) - (absEnd - absStart)

	old := l.btdsChunk.Data
	newRaw := make([]byte, 0, len(old)+delta)
	newRaw = append(newRaw, old[:absStart]...)
	newRaw = append(newRaw, newSrc...)
	newRaw = append(newRaw, old[absEnd:]...)

	// The inner LIST btdk header (immediately before bodyOff) carries
	// its own uint32 BE size — bump it by delta so next time we extract
	// the body we read the correct length. Without this, a length-
	// changing splice misaligns extractBtdkBody on the next call.
	// btdk LIST layout (8 bytes header + 4 bytes formType, body starts at bodyOff):
	//   bodyOff-12: "LIST" (4 bytes)
	//   bodyOff-8:  size uint32 BE (4 bytes) — includes the formType
	//   bodyOff-4:  "btdk" (4 bytes formType)
	//   bodyOff:    body
	if delta != 0 && bodyOff >= 8 {
		sizeOff := bodyOff - 8
		oldSize := binary.BigEndian.Uint32(newRaw[sizeOff : sizeOff+4])
		binary.BigEndian.PutUint32(newRaw[sizeOff:sizeOff+4], uint32(int(oldSize)+delta))
	}

	l.btdsChunk.Data = newRaw
	l.TextSourceRaw = newRaw
	ts, _ := decodeTextSource(newRaw)
	if ts != nil {
		l.TextSource = ts
	}
	return nil
}

// runStylePath builds the PostScript path to the style-run dict at
// the given run index: /1/1/0/0/6/0/{runIdx}/0/0/6. From that base,
// per-key sub-paths (/0, /1, /5, /8, /9, /53, /57, /63, etc.) reach
// each field documented in parse_text.go's path table.
func runStylePath(runIdx int) string {
	return fmt.Sprintf("/1/1/0/0/6/0/%d/0/0/6", runIdx)
}

// paragraphStylePath builds the PostScript path to the paragraph
// style dict at the given paragraph index: /1/1/0/0/5/0/{paraIdx}/0/0/5.
func paragraphStylePath(paraIdx int) string {
	return fmt.Sprintf("/1/1/0/0/5/0/%d/0/0/5", paraIdx)
}

// formatPSNumber writes a float64 in the same shape AE uses: trailing
// zeros trimmed but a single trailing ".0" kept (so 88 → "88.0", but
// 1.5 stays "1.5"). Integer-valued whole numbers stay int-looking
// (matches AE's mix of "88" and "30.0" — we pick the integer form
// when fractional == 0).
func formatPSNumber(v float64) string {
	// Use 'g' for compactness, but ensure ints look like ints.
	if v == float64(int64(v)) {
		return strconv.FormatInt(int64(v), 10)
	}
	return strconv.FormatFloat(v, 'g', -1, 64)
}

// formatPSColorArray formats a [R, G, B, A] color (each 0..1) as the
// btdk source order [A, R, G, B] inside square brackets. Used by
// SetRunFillColor / SetRunStrokeColor.
func formatPSColorArray(rgba [4]float64) []byte {
	var b bytes.Buffer
	b.WriteByte('[')
	b.WriteByte(' ')
	for _, v := range [4]float64{rgba[3], rgba[0], rgba[1], rgba[2]} {
		b.WriteString(formatPSNumber(v))
		b.WriteByte(' ')
	}
	b.WriteByte(']')
	return b.Bytes()
}

// ──────────────────────────────────────────────────────────────────
// Per-run setters
// ──────────────────────────────────────────────────────────────────

// validateRunIdx ensures runIdx is within the current TextSource.Runs
// slice and that the layer has a text source.
func (l *Layer) validateRunIdx(runIdx int) error {
	if l.TextSource == nil {
		return fmt.Errorf("layer %q: not a text layer", l.Name)
	}
	if runIdx < 0 || runIdx >= len(l.TextSource.Runs) {
		return fmt.Errorf("layer %q: run index %d out of range [0,%d)", l.Name, runIdx, len(l.TextSource.Runs))
	}
	return nil
}

// SetRunFontSize writes a new font size (em points) to style run #runIdx.
// length-variable splice in the btdk PostScript body.
func (l *Layer) SetRunFontSize(runIdx int, sizePts float64) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	return l.splicePSValue(runStylePath(runIdx)+"/1", []byte(formatPSNumber(sizePts)))
}

// SetRunTracking writes character tracking (1/1000 em) on style run #runIdx.
func (l *Layer) SetRunTracking(runIdx int, tracking float64) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	return l.splicePSValue(runStylePath(runIdx)+"/8", []byte(formatPSNumber(tracking)))
}

// SetRunBaselineShift writes baseline shift (em points; positive = up)
// on style run #runIdx.
func (l *Layer) SetRunBaselineShift(runIdx int, shift float64) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	return l.splicePSValue(runStylePath(runIdx)+"/9", []byte(formatPSNumber(shift)))
}

// SetRunLeading writes leading (em points) on style run #runIdx. Note:
// AE auto-leading is gated by the AutoLeading flag — to use this
// value the caller should also disable auto-leading via
// SetRunAutoLeading(runIdx, false), otherwise AE overrides it.
func (l *Layer) SetRunLeading(runIdx int, leading float64) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	return l.splicePSValue(runStylePath(runIdx)+"/5", []byte(formatPSNumber(leading)))
}

// SetRunAutoLeading toggles AE's "auto leading" flag on style run #runIdx.
// When true, the Leading value is computed by AE (typically FontSize × 1.2);
// when false, the explicit Leading value applies.
func (l *Layer) SetRunAutoLeading(runIdx int, auto bool) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	v := "false"
	if auto {
		v = "true"
	}
	return l.splicePSValue(runStylePath(runIdx)+"/4", []byte(v))
}

// SetRunFontIndex repoints style run #runIdx at a different entry in
// the Fonts table (TextSource.Fonts). Caller is responsible for ensuring
// the index is in range; the underlying psValue is just an integer.
func (l *Layer) SetRunFontIndex(runIdx, fontIdx int) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	if l.TextSource != nil && (fontIdx < 0 || fontIdx >= len(l.TextSource.Fonts)) {
		return fmt.Errorf("layer %q: font index %d out of range [0,%d)", l.Name, fontIdx, len(l.TextSource.Fonts))
	}
	return l.splicePSValue(runStylePath(runIdx)+"/0", []byte(strconv.Itoa(fontIdx)))
}

// SetRunFauxBold toggles synthetic bold on style run #runIdx.
func (l *Layer) SetRunFauxBold(runIdx int, on bool) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	v := "false"
	if on {
		v = "true"
	}
	return l.splicePSValue(runStylePath(runIdx)+"/2", []byte(v))
}

// SetRunFauxItalic toggles synthetic italic on style run #runIdx.
func (l *Layer) SetRunFauxItalic(runIdx int, on bool) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	v := "false"
	if on {
		v = "true"
	}
	return l.splicePSValue(runStylePath(runIdx)+"/3", []byte(v))
}

// SetRunHorizontalScale / SetRunVerticalScale write the raw scale
// values used by AE on style run #runIdx. See TextStyleRun docs for
// the unit notes (AE default is 1; scripted setter range 0..100).
func (l *Layer) SetRunHorizontalScale(runIdx int, scale float64) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	return l.splicePSValue(runStylePath(runIdx)+"/6", []byte(formatPSNumber(scale)))
}

func (l *Layer) SetRunVerticalScale(runIdx int, scale float64) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	return l.splicePSValue(runStylePath(runIdx)+"/7", []byte(formatPSNumber(scale)))
}

// SetRunTsume writes the CJK character-spacing adjustment (0..100)
// on style run #runIdx.
func (l *Layer) SetRunTsume(runIdx int, tsume float64) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	return l.splicePSValue(runStylePath(runIdx)+"/36", []byte(formatPSNumber(tsume)))
}

// SetRunFillColor writes the fill paint color [R, G, B, A] (each 0..1)
// on style run #runIdx. Encoded as btdk's [A, R, G, B] array.
func (l *Layer) SetRunFillColor(runIdx int, rgba [4]float64) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	return l.splicePSValue(runStylePath(runIdx)+"/53/0/1", formatPSColorArray(rgba))
}

// SetRunStrokeColor writes the stroke paint color [R, G, B, A] on
// style run #runIdx.
func (l *Layer) SetRunStrokeColor(runIdx int, rgba [4]float64) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	return l.splicePSValue(runStylePath(runIdx)+"/54/0/1", formatPSColorArray(rgba))
}

// SetRunApplyStroke toggles whether the stroke is rendered on
// style run #runIdx.
func (l *Layer) SetRunApplyStroke(runIdx int, apply bool) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	v := "false"
	if apply {
		v = "true"
	}
	return l.splicePSValue(runStylePath(runIdx)+"/57", []byte(v))
}

// SetRunStrokeWidth writes the stroke width (em points) on style run #runIdx.
func (l *Layer) SetRunStrokeWidth(runIdx int, width float64) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	return l.splicePSValue(runStylePath(runIdx)+"/63", []byte(formatPSNumber(width)))
}

// SetRunCapsOption writes the font caps option on style run #runIdx
// (AE 24+ writeable; underlying btdk byte exists in AE 2020 files too,
// but AE 2020 ScriptingAPI marks allCaps / smallCaps readonly so
// fixture generation requires AE 24).
func (l *Layer) SetRunCapsOption(runIdx int, caps TextCapsOption) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	if caps < TextCapsNormal || caps > TextCapsAllSmall {
		return fmt.Errorf("layer %q: invalid TextCapsOption %d", l.Name, int(caps))
	}
	return l.splicePSValue(runStylePath(runIdx)+"/12", []byte(strconv.Itoa(int(caps))))
}

// SetRunBaselineOption writes the font baseline option on style run #runIdx.
// AE 24+ writeable; mirrors subscript / superscript readonly attrs.
func (l *Layer) SetRunBaselineOption(runIdx int, base TextBaselineOption) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	if base < TextBaselineNormal || base > TextBaselineSubscript {
		return fmt.Errorf("layer %q: invalid TextBaselineOption %d", l.Name, int(base))
	}
	return l.splicePSValue(runStylePath(runIdx)+"/13", []byte(strconv.Itoa(int(base))))
}

// SetRunStrokeOverFill toggles whether the stroke renders over the fill
// (true) or under it (false) on style run #runIdx. Default in AE is true.
func (l *Layer) SetRunStrokeOverFill(runIdx int, over bool) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	v := "false"
	if over {
		v = "true"
	}
	return l.splicePSValue(runStylePath(runIdx)+"/58", []byte(v))
}

// SetRunAutoKernType writes the auto-kerning mode on style run #runIdx.
// AE 24+ writeable (btdk style-run /11). Note: setting to TextAutoKernNoAuto
// without also assigning a manual kerning value will read as "0 spacing"
// in AE — manual kerning lives in a separate per-character sub-tree at
// btdk /1/1[0]/0/8 which this library does not yet expose for writing.
func (l *Layer) SetRunAutoKernType(runIdx int, kt TextAutoKernType) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	if kt < TextAutoKernNoAuto || kt > TextAutoKernOptical {
		return fmt.Errorf("layer %q: invalid TextAutoKernType %d", l.Name, int(kt))
	}
	return l.splicePSValue(runStylePath(runIdx)+"/11", []byte(strconv.Itoa(int(kt))))
}

// SetRunNoBreak toggles the "do not break" character flag on style run
// #runIdx (AE 24+ writeable, btdk style-run /52). When true, AE won't
// allow line breaks to fall inside the run.
func (l *Layer) SetRunNoBreak(runIdx int, on bool) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	v := "false"
	if on {
		v = "true"
	}
	return l.splicePSValue(runStylePath(runIdx)+"/52", []byte(v))
}

// SetRunLineJoinType writes the stroke corner join style on style run
// #runIdx (AE 24+ writeable, btdk style-run /62).
func (l *Layer) SetRunLineJoinType(runIdx int, j TextLineJoinType) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	if j < TextLineJoinMiter || j > TextLineJoinBevel {
		return fmt.Errorf("layer %q: invalid TextLineJoinType %d", l.Name, int(j))
	}
	return l.splicePSValue(runStylePath(runIdx)+"/62", []byte(strconv.Itoa(int(j))))
}

// SetRunDigitSet writes the digit set on style run #runIdx (AE 24+
// writeable, btdk style-run /70).
func (l *Layer) SetRunDigitSet(runIdx int, d TextDigitSet) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	if d < TextDigitSetDefault || d > TextDigitSetArabicRTL {
		return fmt.Errorf("layer %q: invalid TextDigitSet %d", l.Name, int(d))
	}
	return l.splicePSValue(runStylePath(runIdx)+"/70", []byte(strconv.Itoa(int(d))))
}

// ──────────────────────────────────────────────────────────────────
// Font table extension
// ──────────────────────────────────────────────────────────────────

// AddFont appends a new entry to the layer's Fonts table (btdk path
// /0/1/0) and returns the new font index. Subsequent
// SetRunFontIndex(runIdx, returnedIndex) calls can reference it.
//
// length-variable: extends the PostScript font array by one entry
// + its dict wrapper. The new entry is serialized in the same shape
// AE writes:
//
//	<< /0 << /99 /CoolTypeFont /0 << /0 (<FE FF utf16be name>) /2 0 >> >> >>
//
// where `/2 0` marks the font as user-resolved (matches the second
// real-world font slot AE writes for non-default fonts).
//
// Returns an error if the layer isn't a text layer or the btdk Fonts
// array can't be located.
func (l *Layer) AddFont(fontName string) (int, error) {
	if l.btdsChunk == nil || l.TextSource == nil {
		return -1, fmt.Errorf("layer %q: not a text layer", l.Name)
	}
	if fontName == "" {
		return -1, fmt.Errorf("layer %q: fontName must be non-empty", l.Name)
	}
	body, bodyOff, err := extractBtdkBody(l.btdsChunk.Data)
	if err != nil {
		return -1, fmt.Errorf("layer %q: %w", l.Name, err)
	}
	root := parsePSDict(body)
	if root == nil {
		return -1, fmt.Errorf("layer %q: btdk dict empty", l.Name)
	}
	arr := psPath(root, "/0/1/0")
	if arr == nil || arr.kind != psArr {
		return -1, fmt.Errorf("layer %q: Fonts array at /0/1/0 not found", l.Name)
	}
	// We splice just inside the closing `]`. Find the array's last
	// child srcEnd and insert " <new>" between it and `]`. When the
	// array is empty (rare), insert directly after `[`.
	var insertAt int
	if len(arr.arr) == 0 {
		// arr.srcStart points at the `[`; insert one byte after.
		insertAt = arr.srcStart + 1
	} else {
		last := arr.arr[len(arr.arr)-1]
		insertAt = last.srcEnd
	}
	newEntry := serializeFontEntry(fontName)

	// Build the byte sequence to inject: leading space + serialized entry.
	var injected []byte
	injected = append(injected, ' ')
	injected = append(injected, newEntry...)

	old := l.btdsChunk.Data
	abs := bodyOff + insertAt
	newRaw := make([]byte, 0, len(old)+len(injected))
	newRaw = append(newRaw, old[:abs]...)
	newRaw = append(newRaw, injected...)
	newRaw = append(newRaw, old[abs:]...)

	// Update inner LIST btdk size header (same invariant as splicePSValue).
	if bodyOff >= 8 {
		sizeOff := bodyOff - 8
		oldSize := binary.BigEndian.Uint32(newRaw[sizeOff : sizeOff+4])
		binary.BigEndian.PutUint32(newRaw[sizeOff:sizeOff+4], uint32(int(oldSize)+len(injected)))
	}

	l.btdsChunk.Data = newRaw
	l.TextSourceRaw = newRaw
	ts, _ := decodeTextSource(newRaw)
	if ts != nil {
		l.TextSource = ts
	}
	return len(arr.arr), nil // index of newly-added font (was len-1 + 1 → len of old)
}

// serializeFontEntry renders a CoolTypeFont entry in the shape AE
// writes for non-default fonts:
//
//	<< /0 << /99 /CoolTypeFont /0 << /0 (FE FF utf16be name) /2 0 >> >> >>
func serializeFontEntry(name string) []byte {
	encoded := encodeAEPSStringNoCR(name)
	var b bytes.Buffer
	b.WriteString("<< /0 << /99 /CoolTypeFont /0 << /0 ")
	b.Write(encoded)
	b.WriteString(" /2 0 >> >> >>")
	return b.Bytes()
}

// encodeAEPSStringNoCR is encodeAEPSText without the auto-appended
// trailing \r — used for non-paragraph strings like font names that
// should be encoded literally.
func encodeAEPSStringNoCR(s string) []byte {
	var buf bytes.Buffer
	buf.WriteByte('(')
	buf.WriteByte(0xfe)
	buf.WriteByte(0xff)
	for _, r := range s {
		writeUTF16BEEscapedForBuf(&buf, r)
	}
	buf.WriteByte(')')
	return buf.Bytes()
}

// writeUTF16BEEscapedForBuf is a local wrapper around the existing
// writeUTF16BEEscaped helper from write.go — kept private here to
// avoid the import cycle hazard of cross-file helper sharing.
func writeUTF16BEEscapedForBuf(buf *bytes.Buffer, r rune) {
	writeUTF16BEEscaped(buf, r)
}

// ──────────────────────────────────────────────────────────────────
// Per-paragraph setter
// ──────────────────────────────────────────────────────────────────

// validateParaIdx ensures paraIdx is within the current TextSource.Paragraphs.
func (l *Layer) validateParaIdx(paraIdx int) error {
	if l.TextSource == nil {
		return fmt.Errorf("layer %q: not a text layer", l.Name)
	}
	if paraIdx < 0 || paraIdx >= len(l.TextSource.Paragraphs) {
		return fmt.Errorf("layer %q: paragraph index %d out of range [0,%d)",
			l.Name, paraIdx, len(l.TextSource.Paragraphs))
	}
	return nil
}

// SetParagraphJustification writes a new alignment enum on
// paragraph #paraIdx.
func (l *Layer) SetParagraphJustification(paraIdx int, j TextJustification) error {
	if err := l.validateParaIdx(paraIdx); err != nil {
		return err
	}
	return l.splicePSValue(paragraphStylePath(paraIdx)+"/0", []byte(strconv.Itoa(int(j))))
}

// SetParagraphFirstLineIndent writes firstLineIndent (em points) on
// paragraph #paraIdx. AE 24+ writeable (AE 2020 ScriptingAPI no-op on point text).
func (l *Layer) SetParagraphFirstLineIndent(paraIdx int, v float64) error {
	if err := l.validateParaIdx(paraIdx); err != nil {
		return err
	}
	return l.splicePSValue(paragraphStylePath(paraIdx)+"/1", []byte(formatPSNumber(v)))
}

// SetParagraphStartIndent writes startIndent (em points) on paragraph #paraIdx.
func (l *Layer) SetParagraphStartIndent(paraIdx int, v float64) error {
	if err := l.validateParaIdx(paraIdx); err != nil {
		return err
	}
	return l.splicePSValue(paragraphStylePath(paraIdx)+"/2", []byte(formatPSNumber(v)))
}

// SetParagraphEndIndent writes endIndent (em points) on paragraph #paraIdx.
func (l *Layer) SetParagraphEndIndent(paraIdx int, v float64) error {
	if err := l.validateParaIdx(paraIdx); err != nil {
		return err
	}
	return l.splicePSValue(paragraphStylePath(paraIdx)+"/3", []byte(formatPSNumber(v)))
}

// SetParagraphSpaceBefore writes spaceBefore (em points) on paragraph #paraIdx.
func (l *Layer) SetParagraphSpaceBefore(paraIdx int, v float64) error {
	if err := l.validateParaIdx(paraIdx); err != nil {
		return err
	}
	return l.splicePSValue(paragraphStylePath(paraIdx)+"/4", []byte(formatPSNumber(v)))
}

// SetParagraphSpaceAfter writes spaceAfter (em points) on paragraph #paraIdx.
func (l *Layer) SetParagraphSpaceAfter(paraIdx int, v float64) error {
	if err := l.validateParaIdx(paraIdx); err != nil {
		return err
	}
	return l.splicePSValue(paragraphStylePath(paraIdx)+"/5", []byte(formatPSNumber(v)))
}

// SetParagraphAutoHyphenate toggles auto-hyphenation on paragraph #paraIdx.
// Default in AE is true. AE 24+ writeable.
func (l *Layer) SetParagraphAutoHyphenate(paraIdx int, on bool) error {
	if err := l.validateParaIdx(paraIdx); err != nil {
		return err
	}
	v := "false"
	if on {
		v = "true"
	}
	return l.splicePSValue(paragraphStylePath(paraIdx)+"/9", []byte(v))
}

// SetParagraphLeadingType writes the leading-type enum on paragraph #paraIdx
// (AE 24+ writeable, btdk paragraph /8).
func (l *Layer) SetParagraphLeadingType(paraIdx int, lt TextLeadingType) error {
	if err := l.validateParaIdx(paraIdx); err != nil {
		return err
	}
	if lt < TextLeadingRoman || lt > TextLeadingJapanese {
		return fmt.Errorf("layer %q: invalid TextLeadingType %d", l.Name, int(lt))
	}
	return l.splicePSValue(paragraphStylePath(paraIdx)+"/8", []byte(strconv.Itoa(int(lt))))
}

// SetParagraphHangingRoman toggles Roman Hanging Punctuation on
// paragraph #paraIdx (AE 24+ writeable, btdk paragraph /21). Only
// meaningful for box-text.
func (l *Layer) SetParagraphHangingRoman(paraIdx int, on bool) error {
	if err := l.validateParaIdx(paraIdx); err != nil {
		return err
	}
	v := "false"
	if on {
		v = "true"
	}
	return l.splicePSValue(paragraphStylePath(paraIdx)+"/21", []byte(v))
}

// SetParagraphDirection writes the paragraph reading direction on
// paragraph #paraIdx (AE 24+ writeable, btdk paragraph /33). Affects
// how mixed LTR/RTL text composes.
func (l *Layer) SetParagraphDirection(paraIdx int, d TextParagraphDirection) error {
	if err := l.validateParaIdx(paraIdx); err != nil {
		return err
	}
	if d < TextDirectionLeftToRight || d > TextDirectionRightToLeft {
		return fmt.Errorf("layer %q: invalid TextParagraphDirection %d", l.Name, int(d))
	}
	return l.splicePSValue(paragraphStylePath(paraIdx)+"/33", []byte(strconv.Itoa(int(d))))
}

// ──────────────────────────────────────────────────────────────────
// Manual kerning (per-character)
// ──────────────────────────────────────────────────────────────────

// SetManualKerning writes per-character manual kerning values (in
// 1/1000 em units) into the btdk dict. The values slice length must
// equal the current per-character count (`len(TextSource.ManualKerning)`).
//
// Pre-condition: the layer must already have a manual-kerning slot —
// AE only emits the /1/1[0]/0/8 sub-tree once some run has been set
// to AutoKernType=NoAuto with a non-zero kerning value. First-time
// enablement on a layer that lacks the slot is a structural change
// and is not supported; create the slot from AE first (or set kerning
// via the AE UI / scripting) and then mutate it here.
//
// For the values to actually render, the matching style run's
// AutoKernType must be TextAutoKernNoAuto — toggle it independently
// via SetRunAutoKernType.
//
// Mirrors AE-script TextDocument.kerning: the first-character value
// is also written to sibling /1/1[0]/0/7 (the first-char scalar that
// the script API reflects).
func (l *Layer) SetManualKerning(values []int) error {
	if l.TextSource == nil {
		return fmt.Errorf("layer %q: not a text layer", l.Name)
	}
	n := len(l.TextSource.ManualKerning)
	if n == 0 {
		return fmt.Errorf("layer %q: no manual-kerning slot present "+
			"(set a non-zero kerning value in AE first to materialize /1/1[0]/0/8)",
			l.Name)
	}
	if len(values) != n {
		return fmt.Errorf("layer %q: SetManualKerning: got %d values, want %d (one per existing char)",
			l.Name, len(values), n)
	}
	// Write per-char values first. splicePSValue re-parses the body
	// each call, so length-changing splices stay consistent across the
	// loop. Path: /1/1/0/0/8/0/{i}/0/0.
	for i, v := range values {
		path := fmt.Sprintf("/1/1/0/0/8/0/%d/0/0", i)
		if err := l.splicePSValue(path, []byte(strconv.Itoa(v))); err != nil {
			return err
		}
	}
	// Mirror values[0] to the sibling first-char scalar /1/1[0]/0/7.
	return l.splicePSValue("/1/1/0/0/7", []byte(strconv.Itoa(values[0])))
}
