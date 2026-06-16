package scene

import "fmt"

// Text per-run write API. Each setter targets one codec.PsValue node inside
// the btdk PostScript dict and splices new bytes in. The btdk chunk
// (and ancestor LIST sizes) are length-variable — WriteAEP recomputes
// the parent sizes from chunk Data lengths.
//
// Implementation pattern:
//   1. Validate the run / paragraph index against the decoded TextSource.
//   2. Delegate the byte splice to layerBackrefs (LayerWriter).
//   3. Re-sync Layer.TextSourceRaw / TextSource from the spliced btds
//      bytes so subsequent reads reflect the new state.
//
// Setters are placed on Layer (not TextSource or TextStyleRun) to keep
// the call site discoverable alongside SetText.

// resyncTextSource refreshes Layer.TextSourceRaw and re-decodes
// Layer.TextSource from the (possibly re-spliced) btds chunk bytes.
// Called after a back-side splice so scene reads stay accurate.
func (l *Layer) resyncTextSource() {
	if l.back == nil {
		return
	}
	newRaw := l.back.BtdsData()
	if newRaw == nil {
		return
	}
	l.TextSourceRaw = newRaw
	ts, _ := DecodeTextSource(newRaw)
	if ts != nil {
		l.TextSource = ts
	}
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
//
//aep:cap domain=text tier=stable verify=render-pixel gate=TestMGTextStyle_AEShipGate_AE2020,TestMGTextStyle_AEShipGate_AE2025 boundary="length-variable;btdk PostScript body splice;point-size key 必须写 REAL(FormatPSReal),写裸整数 AE 读成 16.16 定点→fontSize/65536" alias="font size,字号,字体大小,run font size"
func (l *Layer) SetRunFontSize(runIdx int, sizePts float64) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	if err := l.back.SetRunFontSize(runIdx, sizePts); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}

// SetRunTracking writes character tracking (1/1000 em) on style run #runIdx.
//
//aep:cap domain=text tier=stable verify=render-pixel gate=TestMGTextStyle_AEShipGate_AE2020,TestMGTextStyle_AEShipGate_AE2025 boundary="length-variable;btdk PostScript body splice;tracking 是真整数键(FormatPSNumber)" alias="tracking,字距,字符间距,character spacing"
func (l *Layer) SetRunTracking(runIdx int, tracking float64) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	if err := l.back.SetRunTracking(runIdx, tracking); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}

// SetRunBaselineShift writes baseline shift (em points; positive = up)
// on style run #runIdx.
//
//aep:cap domain=text tier=stable verify=roundtrip boundary="length-variable;btdk PostScript body splice;point key 写 REAL(FormatPSReal)否则 AE 读成 16.16 定点;无专门 AE gate→round-trip" alias="baseline shift,基线偏移,baseline offset"
func (l *Layer) SetRunBaselineShift(runIdx int, shift float64) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	if err := l.back.SetRunBaselineShift(runIdx, shift); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}

// SetRunLeading writes leading (em points) on style run #runIdx. Note:
// AE auto-leading is gated by the AutoLeading flag — to use this
// value the caller should also disable auto-leading via
// SetRunAutoLeading(runIdx, false), otherwise AE overrides it.
//
//aep:cap domain=text tier=stable verify=roundtrip boundary="length-variable;btdk PostScript body splice;point key 写 REAL(FormatPSReal);需搭配 SetRunAutoLeading(false) 才生效;evidence-defer:从零层 AE 渲染默认行距,值 round-trip 对但像素不变(见 incident)" alias="leading,行距,line spacing,auto leading"
func (l *Layer) SetRunLeading(runIdx int, leading float64) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	if err := l.back.SetRunLeading(runIdx, leading); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}

// SetRunAutoLeading toggles AE's "auto leading" flag on style run #runIdx.
// When true, the Leading value is computed by AE (typically FontSize × 1.2);
// when false, the explicit Leading value applies.
//
//aep:cap domain=text tier=stable verify=roundtrip boundary="length-variable;btdk PostScript body splice;无专门 AE gate→round-trip" alias="auto leading,自动行距,leading flag"
func (l *Layer) SetRunAutoLeading(runIdx int, auto bool) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	if err := l.back.SetRunAutoLeading(runIdx, auto); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}

// SetRunFontIndex repoints style run #runIdx at a different entry in
// the Fonts table (TextSource.Fonts). Caller is responsible for ensuring
// the index is in range; the underlying psValue is just an integer.
//
//aep:cap domain=text tier=stable verify=roundtrip boundary="length-variable;btdk PostScript body splice;fontIdx 须在 Fonts 表范围内;无专门 AE gate→round-trip" alias="font index,字体索引,switch font,change font"
func (l *Layer) SetRunFontIndex(runIdx, fontIdx int) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	if l.TextSource != nil && (fontIdx < 0 || fontIdx >= len(l.TextSource.Fonts)) {
		return fmt.Errorf("layer %q: font index %d out of range [0,%d)", l.Name, fontIdx, len(l.TextSource.Fonts))
	}
	if err := l.back.SetRunFontIndex(runIdx, fontIdx); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}

// SetRunFauxBold toggles synthetic bold on style run #runIdx.
//
//aep:cap domain=text tier=stable verify=roundtrip boundary="length-variable;btdk PostScript body splice;无专门 AE gate→round-trip" alias="faux bold,仿粗体,synthetic bold,fake bold"
func (l *Layer) SetRunFauxBold(runIdx int, on bool) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	if err := l.back.SetRunFauxBold(runIdx, on); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}

// SetRunFauxItalic toggles synthetic italic on style run #runIdx.
//
//aep:cap domain=text tier=stable verify=roundtrip boundary="length-variable;btdk PostScript body splice;无专门 AE gate→round-trip" alias="faux italic,仿斜体,synthetic italic,fake italic"
func (l *Layer) SetRunFauxItalic(runIdx int, on bool) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	if err := l.back.SetRunFauxItalic(runIdx, on); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}

// SetRunHorizontalScale / SetRunVerticalScale write the raw scale
// values used by AE on style run #runIdx. See TextStyleRun docs for
// the unit notes (AE default is 1; scripted setter range 0..100).
//
//aep:cap domain=text tier=stable verify=roundtrip boundary="length-variable;btdk PostScript body splice;point key 写 REAL(FormatPSReal)否则 AE 读成 16.16 定点;无专门 AE gate→round-trip" alias="horizontal scale,水平缩放,text scale,字体缩放"
func (l *Layer) SetRunHorizontalScale(runIdx int, scale float64) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	if err := l.back.SetRunHorizontalScale(runIdx, scale); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}

// SetRunVerticalScale writes the vertical scale value used by AE on style
// run #runIdx. See TextStyleRun docs for unit notes (AE default is 1;
// scripted setter range 0..100).
//
//aep:cap domain=text tier=stable verify=roundtrip boundary="length-variable;btdk PostScript body splice;point key 写 REAL(FormatPSReal)否则 AE 读成 16.16 定点;无专门 AE gate→round-trip" alias="vertical scale,垂直缩放,text scale,字体缩放"
func (l *Layer) SetRunVerticalScale(runIdx int, scale float64) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	if err := l.back.SetRunVerticalScale(runIdx, scale); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}

// SetRunTsume writes the CJK character-spacing adjustment (0..100)
// on style run #runIdx.
//
//aep:cap domain=text tier=stable verify=roundtrip boundary="length-variable;btdk PostScript body splice;CJK 专用;无专门 AE gate→round-trip" alias="tsume,CJK spacing,字间压缩,CJK 字距"
func (l *Layer) SetRunTsume(runIdx int, tsume float64) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	if err := l.back.SetRunTsume(runIdx, tsume); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}

// SetRunFillColor writes the fill paint color [R, G, B, A] (each 0..1)
// on style run #runIdx. Encoded as btdk's [A, R, G, B] array.
//
//aep:cap domain=text tier=stable verify=roundtrip boundary="length-variable;btdk PostScript body splice;RGBA 各分量 0..1;无专门 AE gate→round-trip" alias="fill color,填充颜色,text color,字体颜色,font color"
func (l *Layer) SetRunFillColor(runIdx int, rgba [4]float64) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	if err := l.back.SetRunFillColor(runIdx, rgba); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}

// SetRunStrokeColor writes the stroke paint color [R, G, B, A] on
// style run #runIdx.
//
//aep:cap domain=text tier=stable verify=roundtrip boundary="length-variable;btdk PostScript body splice;RGBA 各分量 0..1;无专门 AE gate→round-trip" alias="stroke color,描边颜色,text stroke,字体描边"
func (l *Layer) SetRunStrokeColor(runIdx int, rgba [4]float64) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	if err := l.back.SetRunStrokeColor(runIdx, rgba); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}

// SetRunApplyStroke toggles whether the stroke is rendered on
// style run #runIdx.
//
//aep:cap domain=text tier=stable verify=roundtrip boundary="length-variable;btdk PostScript body splice;无专门 AE gate→round-trip" alias="apply stroke,启用描边,stroke on off,enable stroke"
func (l *Layer) SetRunApplyStroke(runIdx int, apply bool) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	if err := l.back.SetRunApplyStroke(runIdx, apply); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}

// SetRunStrokeWidth writes the stroke width (em points) on style run #runIdx.
//
//aep:cap domain=text tier=stable verify=roundtrip boundary="length-variable;btdk PostScript body splice;point key 写 REAL(FormatPSReal)否则 AE 读成 16.16 定点;无专门 AE gate→round-trip" alias="stroke width,描边宽度,stroke size,text outline width"
func (l *Layer) SetRunStrokeWidth(runIdx int, width float64) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	if err := l.back.SetRunStrokeWidth(runIdx, width); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}

// SetRunCapsOption writes the font caps option on style run #runIdx
// (AE 24+ writeable; underlying btdk byte exists in AE 2020 files too,
// but AE 2020 ScriptingAPI marks allCaps / smallCaps readonly so
// fixture generation requires AE 24).
//
//aep:cap domain=text tier=stable verify=render-pixel gate=TestMGTextStyle_AEShipGate_AE2020,TestMGTextStyle_AEShipGate_AE2025 boundary="length-variable;btdk PostScript body splice;AE 24+ ScriptingAPI 才可写" alias="caps option,大写选项,all caps,small caps,uppercase"
func (l *Layer) SetRunCapsOption(runIdx int, caps TextCapsOption) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	if caps < TextCapsNormal || caps > TextCapsAllSmall {
		return fmt.Errorf("layer %q: invalid TextCapsOption %d", l.Name, int(caps))
	}
	if err := l.back.SetRunCapsOption(runIdx, caps); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}

// SetRunBaselineOption writes the font baseline option on style run #runIdx.
// AE 24+ writeable; mirrors subscript / superscript readonly attrs.
//
//aep:cap domain=text tier=stable verify=roundtrip boundary="length-variable;btdk PostScript body splice;AE 24+ ScriptingAPI 才可写;无专门 AE gate→round-trip" alias="baseline option,基线选项,subscript,superscript,上标,下标"
func (l *Layer) SetRunBaselineOption(runIdx int, base TextBaselineOption) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	if base < TextBaselineNormal || base > TextBaselineSubscript {
		return fmt.Errorf("layer %q: invalid TextBaselineOption %d", l.Name, int(base))
	}
	if err := l.back.SetRunBaselineOption(runIdx, base); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}

// SetRunStrokeOverFill toggles whether the stroke renders over the fill
// (true) or under it (false) on style run #runIdx. Default in AE is true.
//
//aep:cap domain=text tier=stable verify=roundtrip boundary="length-variable;btdk PostScript body splice;无专门 AE gate→round-trip" alias="stroke over fill,描边覆盖填充,stroke order,描边顺序"
func (l *Layer) SetRunStrokeOverFill(runIdx int, over bool) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	if err := l.back.SetRunStrokeOverFill(runIdx, over); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}

// SetRunAutoKernType writes the auto-kerning mode on style run #runIdx.
// AE 24+ writeable (btdk style-run /11). Note: setting to TextAutoKernNoAuto
// without also assigning a manual kerning value will read as "0 spacing"
// in AE — manual kerning lives in a separate per-character sub-tree at
// btdk /1/1[0]/0/8 which this library does not yet expose for writing.
//
//aep:cap domain=text tier=stable verify=roundtrip boundary="length-variable;btdk PostScript body splice;AE 24+ ScriptingAPI 才可写;NoAuto 需配合 SetManualKerning 才有效;无专门 AE gate→round-trip" alias="auto kern,自动字距调整,kerning mode,optical kerning,metrics kerning"
func (l *Layer) SetRunAutoKernType(runIdx int, kt TextAutoKernType) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	if kt < TextAutoKernNoAuto || kt > TextAutoKernOptical {
		return fmt.Errorf("layer %q: invalid TextAutoKernType %d", l.Name, int(kt))
	}
	if err := l.back.SetRunAutoKernType(runIdx, kt); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}

// SetRunNoBreak toggles the "do not break" character flag on style run
// #runIdx (AE 24+ writeable, btdk style-run /52). When true, AE won't
// allow line breaks to fall inside the run.
//
//aep:cap domain=text tier=stable verify=roundtrip boundary="length-variable;btdk PostScript body splice;AE 24+ ScriptingAPI 才可写;无专门 AE gate→round-trip" alias="no break,禁止换行,do not break,word wrap"
func (l *Layer) SetRunNoBreak(runIdx int, on bool) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	if err := l.back.SetRunNoBreak(runIdx, on); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}

// SetRunLineJoinType writes the stroke corner join style on style run
// #runIdx (AE 24+ writeable, btdk style-run /62).
//
//aep:cap domain=text tier=stable verify=roundtrip boundary="length-variable;btdk PostScript body splice;AE 24+ ScriptingAPI 才可写;无专门 AE gate→round-trip" alias="line join,stroke join,线段连接,描边角点,miter join,bevel join,round join"
func (l *Layer) SetRunLineJoinType(runIdx int, j TextLineJoinType) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	if j < TextLineJoinMiter || j > TextLineJoinBevel {
		return fmt.Errorf("layer %q: invalid TextLineJoinType %d", l.Name, int(j))
	}
	if err := l.back.SetRunLineJoinType(runIdx, j); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}

// SetRunDigitSet writes the digit set on style run #runIdx (AE 24+
// writeable, btdk style-run /70).
//
//aep:cap domain=text tier=stable verify=roundtrip boundary="length-variable;btdk PostScript body splice;AE 24+ ScriptingAPI 才可写;无专门 AE gate→round-trip" alias="digit set,数字集,Arabic digits,Hindi digits,阿拉伯数字,印地数字"
func (l *Layer) SetRunDigitSet(runIdx int, d TextDigitSet) error {
	if err := l.validateRunIdx(runIdx); err != nil {
		return err
	}
	if d < TextDigitSetDefault || d > TextDigitSetArabicRTL {
		return fmt.Errorf("layer %q: invalid TextDigitSet %d", l.Name, int(d))
	}
	if err := l.back.SetRunDigitSet(runIdx, d); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
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
//
//aep:cap domain=text tier=stable verify=roundtrip boundary="length-variable;btdk 字体表数组扩展一条记录;需搭配 SetRunFontIndex 引用新索引;无专门 AE gate→round-trip" alias="add font,添加字体,font table,字体表,register font"
func (l *Layer) AddFont(fontName string) (int, error) {
	if l.back == nil || l.TextSource == nil {
		return -1, fmt.Errorf("layer %q: not a text layer", l.Name)
	}
	idx, err := l.back.AddFont(fontName)
	if err != nil {
		return -1, err
	}
	l.resyncTextSource()
	return idx, nil
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
//
//aep:cap domain=text tier=stable verify=render-pixel gate=TestMGTextStyle_AEShipGate_AE2020,TestMGTextStyle_AEShipGate_AE2025 boundary="length-variable;btdk PostScript paragraph body splice" alias="justification,对齐,alignment,text align,left align,right align,center,全对齐"
func (l *Layer) SetParagraphJustification(paraIdx int, j TextJustification) error {
	if err := l.validateParaIdx(paraIdx); err != nil {
		return err
	}
	if err := l.back.SetParagraphJustification(paraIdx, j); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}

// SetParagraphFirstLineIndent writes firstLineIndent (em points) on
// paragraph #paraIdx. AE 24+ writeable (AE 2020 ScriptingAPI no-op on point text).
//
//aep:cap domain=text tier=stable verify=roundtrip boundary="length-variable;btdk PostScript paragraph body splice;AE 2020 ScriptingAPI 对 point text 无效;无专门 AE gate→round-trip" alias="first line indent,首行缩进,paragraph indent,段落缩进"
func (l *Layer) SetParagraphFirstLineIndent(paraIdx int, v float64) error {
	if err := l.validateParaIdx(paraIdx); err != nil {
		return err
	}
	if err := l.back.SetParagraphFirstLineIndent(paraIdx, v); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}

// SetParagraphStartIndent writes startIndent (em points) on paragraph #paraIdx.
//
//aep:cap domain=text tier=stable verify=roundtrip boundary="length-variable;btdk PostScript paragraph body splice;无专门 AE gate→round-trip" alias="start indent,左缩进,paragraph start indent,left margin"
func (l *Layer) SetParagraphStartIndent(paraIdx int, v float64) error {
	if err := l.validateParaIdx(paraIdx); err != nil {
		return err
	}
	if err := l.back.SetParagraphStartIndent(paraIdx, v); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}

// SetParagraphEndIndent writes endIndent (em points) on paragraph #paraIdx.
//
//aep:cap domain=text tier=stable verify=roundtrip boundary="length-variable;btdk PostScript paragraph body splice;无专门 AE gate→round-trip" alias="end indent,右缩进,paragraph end indent,right margin"
func (l *Layer) SetParagraphEndIndent(paraIdx int, v float64) error {
	if err := l.validateParaIdx(paraIdx); err != nil {
		return err
	}
	if err := l.back.SetParagraphEndIndent(paraIdx, v); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}

// SetParagraphSpaceBefore writes spaceBefore (em points) on paragraph #paraIdx.
//
//aep:cap domain=text tier=stable verify=roundtrip boundary="length-variable;btdk PostScript paragraph body splice;无专门 AE gate→round-trip" alias="space before,段前间距,paragraph space before,paragraph spacing"
func (l *Layer) SetParagraphSpaceBefore(paraIdx int, v float64) error {
	if err := l.validateParaIdx(paraIdx); err != nil {
		return err
	}
	if err := l.back.SetParagraphSpaceBefore(paraIdx, v); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}

// SetParagraphSpaceAfter writes spaceAfter (em points) on paragraph #paraIdx.
//
//aep:cap domain=text tier=stable verify=roundtrip boundary="length-variable;btdk PostScript paragraph body splice;无专门 AE gate→round-trip" alias="space after,段后间距,paragraph space after,paragraph spacing"
func (l *Layer) SetParagraphSpaceAfter(paraIdx int, v float64) error {
	if err := l.validateParaIdx(paraIdx); err != nil {
		return err
	}
	if err := l.back.SetParagraphSpaceAfter(paraIdx, v); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}

// SetParagraphAutoHyphenate toggles auto-hyphenation on paragraph #paraIdx.
// Default in AE is true. AE 24+ writeable.
//
//aep:cap domain=text tier=stable verify=roundtrip boundary="length-variable;btdk PostScript paragraph body splice;AE 24+ ScriptingAPI 才可写;无专门 AE gate→round-trip" alias="auto hyphenate,自动连字,hyphenation,断字"
func (l *Layer) SetParagraphAutoHyphenate(paraIdx int, on bool) error {
	if err := l.validateParaIdx(paraIdx); err != nil {
		return err
	}
	if err := l.back.SetParagraphAutoHyphenate(paraIdx, on); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}

// SetParagraphLeadingType writes the leading-type enum on paragraph #paraIdx
// (AE 24+ writeable, btdk paragraph /8).
//
//aep:cap domain=text tier=stable verify=roundtrip boundary="length-variable;btdk PostScript paragraph body splice;AE 24+ ScriptingAPI 才可写;无专门 AE gate→round-trip" alias="leading type,行距类型,roman leading,Japanese leading,段落行距"
func (l *Layer) SetParagraphLeadingType(paraIdx int, lt TextLeadingType) error {
	if err := l.validateParaIdx(paraIdx); err != nil {
		return err
	}
	if lt < TextLeadingRoman || lt > TextLeadingJapanese {
		return fmt.Errorf("layer %q: invalid TextLeadingType %d", l.Name, int(lt))
	}
	if err := l.back.SetParagraphLeadingType(paraIdx, lt); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}

// SetParagraphHangingRoman toggles Roman Hanging Punctuation on
// paragraph #paraIdx (AE 24+ writeable, btdk paragraph /21). Only
// meaningful for box-text.
//
//aep:cap domain=text tier=stable verify=roundtrip boundary="length-variable;btdk PostScript paragraph body splice;AE 24+ ScriptingAPI 才可写;仅对 box-text 有意义;无专门 AE gate→round-trip" alias="hanging roman,悬挂标点,Roman Hanging Punctuation,段落悬挂"
func (l *Layer) SetParagraphHangingRoman(paraIdx int, on bool) error {
	if err := l.validateParaIdx(paraIdx); err != nil {
		return err
	}
	if err := l.back.SetParagraphHangingRoman(paraIdx, on); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}

// SetParagraphDirection writes the paragraph reading direction on
// paragraph #paraIdx (AE 24+ writeable, btdk paragraph /33). Affects
// how mixed LTR/RTL text composes.
//
//aep:cap domain=text tier=stable verify=roundtrip boundary="length-variable;btdk PostScript paragraph body splice;AE 24+ ScriptingAPI 才可写;无专门 AE gate→round-trip" alias="paragraph direction,段落方向,LTR,RTL,reading direction,阅读方向,从右到左"
func (l *Layer) SetParagraphDirection(paraIdx int, d TextParagraphDirection) error {
	if err := l.validateParaIdx(paraIdx); err != nil {
		return err
	}
	if d < TextDirectionLeftToRight || d > TextDirectionRightToLeft {
		return fmt.Errorf("layer %q: invalid TextParagraphDirection %d", l.Name, int(d))
	}
	if err := l.back.SetParagraphDirection(paraIdx, d); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
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
//
//aep:cap domain=text tier=stable verify=roundtrip boundary="length-variable;btdk /1/1[0]/0/8 per-char 数组原位重写;需 btdk 已有 manual-kerning slot(首次须先在 AE 中建立);值个数必须等于现有字符数;无专门 AE gate→round-trip" alias="manual kerning,手动字距,kerning values,per-character kerning,字距调整"
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
	if err := l.back.SetManualKerning(values); err != nil {
		return err
	}
	l.resyncTextSource()
	return nil
}
