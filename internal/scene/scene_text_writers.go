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

// @summary     Set the font size of a text style run
// @description Splices a new font-size value into the btdk PostScript body
//   for the given run.
// @param       runIdx   index of the style run to update
// @param       sizePts  new font size in em points
// @domain      text
// @stability   stable
// @verify      render-pixel
// @gate        TestMGTextStyle_AEShipGate_AE2020,TestMGTextStyle_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk PostScript body splice; the point-size key
//   must be written as REAL (FormatPSReal) — writing a bare integer makes AE
//   read it as a 16.16 fixed-point value, dividing fontSize by 65536
// @alias       font size,字号,字体大小,run font size
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

// @summary     Set the character tracking of a text style run
// @description Splices a new tracking value into the btdk PostScript body
//   for the given run.
// @param       runIdx    index of the style run to update
// @param       tracking  new tracking value in 1/1000 em units
// @domain      text
// @stability   stable
// @verify      render-pixel
// @gate        TestMGTextStyle_AEShipGate_AE2020,TestMGTextStyle_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk PostScript body splice; tracking is a true
//   integer key (FormatPSNumber)
// @alias       tracking,字距,字符间距,character spacing
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

// @summary     Set the baseline shift of a text style run
// @description Splices a new baseline-shift value into the btdk PostScript
//   body for the given run. Positive values shift the run up.
// @param       runIdx  index of the style run to update
// @param       shift   baseline shift in em points (positive = up)
// @domain      text
// @stability   stable
// @verify      ae-accept
// @gate        TestTextRun_AEShipGate_AE2020,TestTextRun_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk PostScript body splice; the point key
//   must be written as REAL (FormatPSReal) or AE reads it as a 16.16
//   fixed-point value; verified against a single-run DOM fixture on both AE
//   versions
// @alias       baseline shift,基线偏移,baseline offset
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

// @summary     Set the leading of a text style run
// @description Splices a new leading value into the btdk PostScript body for
//   the given run. AE auto-leading is gated by the AutoLeading flag — to make
//   this value take effect the caller should also disable auto-leading via
//   SetRunAutoLeading(runIdx, false), otherwise AE overrides it.
// @param       runIdx   index of the style run to update
// @param       leading  new leading value in em points
// @domain      text
// @stability   stable
// @verify      ae-accept
// @gate        TestTextRun_AEShipGate_AE2020,TestTextRun_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk PostScript body splice; the point key
//   must be written as REAL (FormatPSReal); pair with SetRunAutoLeading(false);
//   verified against the DOM value (leading=80) on both AE versions; render
//   still falls back to default leading because the rendered-pixel boundary
//   for from-scratch text is not yet verified
// @alias       leading,行距,line spacing,auto leading
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

// @summary     Toggle auto-leading on a text style run
// @description Splices the auto-leading flag into the btdk PostScript body
//   for the given run. When true, the Leading value is computed by AE
//   (typically FontSize x 1.2); when false, the explicit Leading value
//   applies.
// @param       runIdx  index of the style run to update
// @param       auto    whether AE should compute leading automatically
// @domain      text
// @stability   stable
// @verify      ae-accept
// @gate        TestTextRun_AEShipGate_AE2020,TestTextRun_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk PostScript body splice; verified against
//   a single-run DOM fixture on both AE versions
// @alias       auto leading,自动行距,leading flag
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

// @summary     Repoint a text style run at a different font table entry
// @description Splices a new font index into the btdk PostScript body for
//   the given run, referencing an entry in the layer's Fonts table
//   (TextSource.Fonts). The underlying value is a plain integer; the caller
//   is responsible for ensuring the index is in range.
// @param       runIdx   index of the style run to update
// @param       fontIdx  index into TextSource.Fonts to point the run at
// @domain      text
// @stability   stable
// @verify      ae-accept
// @gate        TestTextFont_AEShipGate_AE2020,TestTextFont_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk PostScript body splice; fontIdx must be
//   within the Fonts table range; verified on both AE versions via
//   AddFont(ArialMT) + SetRunFontIndex producing DOM textDocument.font =
//   ArialMT
// @alias       font index,字体索引,switch font,change font
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

// @summary     Toggle synthetic bold on a text style run
// @description Splices the faux-bold flag into the btdk PostScript body for
//   the given run.
// @param       runIdx  index of the style run to update
// @param       on      whether synthetic bold is enabled
// @domain      text
// @stability   stable
// @verify      ae-accept
// @gate        TestTextRun_AEShipGate_AE2020,TestTextRun_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk PostScript body splice; verified against
//   a single-run DOM fixture on both AE versions
// @alias       faux bold,仿粗体,synthetic bold,fake bold
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

// @summary     Toggle synthetic italic on a text style run
// @description Splices the faux-italic flag into the btdk PostScript body for
//   the given run.
// @param       runIdx  index of the style run to update
// @param       on      whether synthetic italic is enabled
// @domain      text
// @stability   stable
// @verify      ae-accept
// @gate        TestTextRun_AEShipGate_AE2020,TestTextRun_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk PostScript body splice; verified against
//   a single-run DOM fixture on both AE versions
// @alias       faux italic,仿斜体,synthetic italic,fake italic
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

// @summary     Set the horizontal scale of a text style run
// @description Splices a new horizontal-scale value into the btdk
//   PostScript body for the given run. AE default is 1; the scripted setter
//   range is 0..100 (see TextStyleRun for unit notes).
// @param       runIdx  index of the style run to update
// @param       scale   new horizontal scale value
// @domain      text
// @stability   stable
// @verify      ae-accept
// @gate        TestTextRun_AEShipGate_AE2020,TestTextRun_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk PostScript body splice; the point key
//   must be written as REAL (FormatPSReal) or AE reads it as a 16.16
//   fixed-point value; verified by reading back DOM scale=50 on both AE
//   versions
// @alias       horizontal scale,水平缩放,text scale,字体缩放
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

// @summary     Set the vertical scale of a text style run
// @description Splices a new vertical-scale value into the btdk PostScript
//   body for the given run. AE default is 1; the scripted setter range is
//   0..100 (see TextStyleRun for unit notes).
// @param       runIdx  index of the style run to update
// @param       scale   new vertical scale value
// @domain      text
// @stability   stable
// @verify      ae-accept
// @gate        TestTextRun_AEShipGate_AE2020,TestTextRun_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk PostScript body splice; the point key
//   must be written as REAL (FormatPSReal) or AE reads it as a 16.16
//   fixed-point value; verified by reading back DOM scale=150 on both AE
//   versions
// @alias       vertical scale,垂直缩放,text scale,字体缩放
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

// @summary     Set the CJK character-spacing adjustment of a text style run
// @description Splices a new tsume value into the btdk PostScript body for
//   the given run. The setter range is 0..100; the DOM range is 0..1.
// @param       runIdx  index of the style run to update
// @param       tsume   new tsume value in the 0..100 range
// @domain      text
// @stability   stable
// @verify      ae-accept
// @gate        TestTextRun_AEShipGate_AE2020,TestTextRun_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk PostScript body splice; CJK-only
//   attribute; the key must be written as REAL (FormatPSReal) — writing it
//   as a plain number makes AE divide the value by 65536; verified against
//   DOM tsume=0.5 on both AE versions
// @alias       tsume,CJK spacing,字间压缩,CJK 字距
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

// @summary     Set the fill color of a text style run
// @description Splices a new fill color into the btdk PostScript body for
//   the given run, encoded as btdk's [A, R, G, B] array.
// @param       runIdx  index of the style run to update
// @param       rgba    fill color as [R, G, B, A], each component in 0..1
// @domain      text
// @stability   stable
// @verify      ae-accept
// @gate        TestTextRun_AEShipGate_AE2020,TestTextRun_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk PostScript body splice; each RGBA
//   component is in 0..1; verified against a single-run DOM fixture on both
//   AE versions
// @alias       fill color,填充颜色,text color,字体颜色,font color
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

// @summary     Set the stroke color of a text style run
// @description Splices a new stroke color into the btdk PostScript body for
//   the given run.
// @param       runIdx  index of the style run to update
// @param       rgba    stroke color as [R, G, B, A], each component in 0..1
// @domain      text
// @stability   stable
// @verify      ae-accept
// @gate        TestTextRun_AEShipGate_AE2020,TestTextRun_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk PostScript body splice; each RGBA
//   component is in 0..1; verified against a single-run DOM fixture on both
//   AE versions
// @alias       stroke color,描边颜色,text stroke,字体描边
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

// @summary     Toggle whether a text style run renders its stroke
// @description Splices the apply-stroke flag into the btdk PostScript body
//   for the given run.
// @param       runIdx  index of the style run to update
// @param       apply   whether the stroke is rendered
// @domain      text
// @stability   stable
// @verify      ae-accept
// @gate        TestTextRun_AEShipGate_AE2020,TestTextRun_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk PostScript body splice; verified against
//   a single-run DOM fixture on both AE versions
// @alias       apply stroke,启用描边,stroke on off,enable stroke
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

// @summary     Set the stroke width of a text style run
// @description Splices a new stroke-width value into the btdk PostScript
//   body for the given run.
// @param       runIdx  index of the style run to update
// @param       width   new stroke width in em points
// @domain      text
// @stability   stable
// @verify      ae-accept
// @gate        TestTextRun_AEShipGate_AE2020,TestTextRun_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk PostScript body splice; the point key
//   must be written as REAL (FormatPSReal) or AE reads it as a 16.16
//   fixed-point value; verified against a single-run DOM fixture on both AE
//   versions
// @alias       stroke width,描边宽度,stroke size,text outline width
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

// @summary     Set the caps option of a text style run
// @description Splices a new caps-option value into the btdk PostScript
//   body for the given run. The underlying btdk byte exists in AE 2020 files
//   too, but the AE 2020 scripting API marks allCaps / smallCaps read-only,
//   so the option is only writeable from AE 24 onward.
// @param       runIdx  index of the style run to update
// @param       caps    new caps option
// @domain      text
// @stability   stable
// @verify      render-pixel
// @gate        TestMGTextStyle_AEShipGate_AE2020,TestMGTextStyle_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk PostScript body splice; only writeable
//   through the AE 24+ scripting API
// @alias       caps option,大写选项,all caps,small caps,uppercase
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

// @summary     Set the baseline option of a text style run
// @description Splices a new baseline-option value into the btdk PostScript
//   body for the given run. Mirrors the subscript / superscript read-only
//   attributes exposed elsewhere.
// @param       runIdx  index of the style run to update
// @param       base    new baseline option
// @domain      text
// @stability   stable
// @verify      ae-accept
// @gate        TestTextEnum_AEShipGate_AE2020,TestTextEnum_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk PostScript body splice; only writeable
//   through the AE 24+ scripting API; on AE 2020 the enum round-trips as an
//   opaque preserved value, while AE 2025 exposes it through the DOM —
//   verified by re-saving and re-reading on both AE versions
// @alias       baseline option,基线选项,subscript,superscript,上标,下标
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

// @summary     Set the stroke/fill render order of a text style run
// @description Splices the stroke-over-fill flag into the btdk PostScript
//   body for the given run. AE's default is true (stroke renders over fill).
// @param       runIdx  index of the style run to update
// @param       over    true to render the stroke over the fill, false for
//   under
// @domain      text
// @stability   stable
// @verify      ae-accept
// @gate        TestTextRun_AEShipGate_AE2020,TestTextRun_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk PostScript body splice; verified against
//   a single-run DOM fixture on both AE versions
// @alias       stroke over fill,描边覆盖填充,stroke order,描边顺序
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

// @summary     Set the auto-kerning mode of a text style run
// @description Splices a new auto-kern-type value into the btdk PostScript
//   body for the given run (btdk style-run /11). Setting it to
//   TextAutoKernNoAuto without also assigning a manual kerning value reads
//   as zero spacing in AE — manual kerning lives in the separate
//   per-character sub-tree at btdk /1/1[0]/0/8, written via SetManualKerning.
// @param       runIdx  index of the style run to update
// @param       kt      new auto-kern type
// @domain      text
// @stability   stable
// @verify      ae-accept
// @gate        TestTextEnum_AEShipGate_AE2020,TestTextEnum_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk PostScript body splice; only writeable
//   through the AE 24+ scripting API; TextAutoKernNoAuto needs a paired
//   SetManualKerning call to have a visible effect; on AE 2020 the enum
//   round-trips as an opaque preserved value, while AE 2025 exposes it
//   through the DOM — verified by re-saving and re-reading on both AE
//   versions
// @alias       auto kern,自动字距调整,kerning mode,optical kerning,metrics kerning
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

// @summary     Toggle the "do not break" flag on a text style run
// @description Splices the no-break flag into the btdk PostScript body for
//   the given run (btdk style-run /52). When true, AE won't allow line
//   breaks to fall inside the run.
// @param       runIdx  index of the style run to update
// @param       on      whether line breaks inside the run are disallowed
// @domain      text
// @stability   stable
// @verify      ae-accept
// @gate        TestTextEnum_AEShipGate_AE2020,TestTextEnum_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk PostScript body splice; only writeable
//   through the AE 24+ scripting API; on AE 2020 the enum round-trips as an
//   opaque preserved value, while AE 2025 exposes it through the DOM —
//   verified by re-saving and re-reading on both AE versions
// @alias       no break,禁止换行,do not break,word wrap
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

// @summary     Set the stroke corner join style of a text style run
// @description Splices a new line-join-type value into the btdk PostScript
//   body for the given run (btdk style-run /62).
// @param       runIdx  index of the style run to update
// @param       j       new line join type
// @domain      text
// @stability   stable
// @verify      ae-accept
// @gate        TestTextEnum_AEShipGate_AE2020,TestTextEnum_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk PostScript body splice; only writeable
//   through the AE 24+ scripting API; on AE 2020 the enum round-trips as an
//   opaque preserved value, while AE 2025 exposes it through the DOM —
//   verified by re-saving and re-reading on both AE versions
// @alias       line join,stroke join,线段连接,描边角点,miter join,bevel join,round join
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

// @summary     Set the digit set of a text style run
// @description Splices a new digit-set value into the btdk PostScript body
//   for the given run (btdk style-run /70).
// @param       runIdx  index of the style run to update
// @param       d       new digit set
// @domain      text
// @stability   stable
// @verify      ae-accept
// @gate        TestTextEnum_AEShipGate_AE2020,TestTextEnum_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk PostScript body splice; only writeable
//   through the AE 24+ scripting API; on AE 2020 the enum round-trips as an
//   opaque preserved value, while AE 2025 exposes it through the DOM —
//   verified by re-saving and re-reading on both AE versions
// @alias       digit set,数字集,Arabic digits,Hindi digits,阿拉伯数字,印地数字
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

// @summary     Append a new entry to the layer's font table
// @description Extends the btdk Fonts array (btdk path /0/1/0) by one entry
//   plus its dict wrapper. The new entry is serialized in the same shape AE
//   writes, with the second dict slot marking the font as user-resolved
//   (matching the slot AE writes for non-default fonts). Use the returned
//   index with SetRunFontIndex to point a style run at the new font.
// @param       fontName  PostScript name of an installed system font (for
//   example ArialMT)
// @returns     the index of the newly appended font entry
// @domain      text
// @stability   stable
// @verify      ae-accept
// @gate        TestTextFont_AEShipGate_AE2020,TestTextFont_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable; extends the btdk font table array by one
//   record; pair with SetRunFontIndex to reference the new index; fontName
//   must be the PostScript name of a font actually installed on the
//   rendering system; verified on both AE versions
// @alias       add font,添加字体,font table,字体表,register font
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

// @summary     Set the alignment of a text paragraph
// @description Splices a new justification value into the btdk PostScript
//   paragraph body for the given paragraph.
// @param       paraIdx  index of the paragraph to update
// @param       j        new justification value
// @domain      text
// @stability   stable
// @verify      render-pixel
// @gate        TestMGTextStyle_AEShipGate_AE2020,TestMGTextStyle_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk PostScript paragraph body splice
// @alias       justification,对齐,alignment,text align,left align,right align,center,全对齐
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

// @summary     Set the first-line indent of a text paragraph
// @description Splices a new firstLineIndent value into the btdk paragraph
//   body for the given paragraph. The AE 2020 scripting API is a no-op for
//   this attribute on point text; it is writeable from AE 24 onward.
// @param       paraIdx  index of the paragraph to update
// @param       v        first-line indent in em points
// @domain      text
// @stability   stable
// @verify      ae-accept
// @gate        TestTextPara_AEShipGate_AE2020,TestTextPara_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk paragraph splice; the key must be
//   written as REAL (FormatPSReal) or AE divides the value by 65536;
//   verified against the DOM firstLineIndent value plus a resave-preservation
//   check on both AE versions
// @alias       first line indent,首行缩进,paragraph indent,段落缩进
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

// @summary     Set the start indent of a text paragraph
// @description Splices a new startIndent value into the btdk paragraph body
//   for the given paragraph.
// @param       paraIdx  index of the paragraph to update
// @param       v        start indent in em points
// @domain      text
// @stability   stable
// @verify      ae-accept
// @gate        TestTextPara_AEShipGate_AE2020,TestTextPara_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk paragraph splice; the key must be
//   written as REAL (FormatPSReal) or AE divides the value by 65536; the AE
//   textDocument has no leftMargin DOM property, so this is verified by
//   resave-preservation — confirming the value survives a round trip
//   through the AE engine — on both AE versions
// @alias       start indent,左缩进,paragraph start indent,left margin
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

// @summary     Set the end indent of a text paragraph
// @description Splices a new endIndent value into the btdk paragraph body
//   for the given paragraph.
// @param       paraIdx  index of the paragraph to update
// @param       v        end indent in em points
// @domain      text
// @stability   stable
// @verify      ae-accept
// @gate        TestTextPara_AEShipGate_AE2020,TestTextPara_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk paragraph splice; the key must be
//   written as REAL (FormatPSReal) or AE divides the value by 65536; the AE
//   textDocument has no rightMargin DOM property, so this is verified by
//   resave-preservation — confirming the value survives a round trip
//   through the AE engine — on both AE versions
// @alias       end indent,右缩进,paragraph end indent,right margin
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

// @summary     Set the space-before of a text paragraph
// @description Splices a new spaceBefore value into the btdk paragraph body
//   for the given paragraph.
// @param       paraIdx  index of the paragraph to update
// @param       v        space before in em points
// @domain      text
// @stability   stable
// @verify      ae-accept
// @gate        TestTextPara_AEShipGate_AE2020,TestTextPara_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk paragraph splice; the key must be
//   written as REAL (FormatPSReal) or AE divides the value by 65536;
//   verified against the DOM spaceBefore value plus a resave-preservation
//   check on both AE versions
// @alias       space before,段前间距,paragraph space before,paragraph spacing
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

// @summary     Set the space-after of a text paragraph
// @description Splices a new spaceAfter value into the btdk paragraph body
//   for the given paragraph.
// @param       paraIdx  index of the paragraph to update
// @param       v        space after in em points
// @domain      text
// @stability   stable
// @verify      ae-accept
// @gate        TestTextPara_AEShipGate_AE2020,TestTextPara_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk paragraph splice; the key must be
//   written as REAL (FormatPSReal) or AE divides the value by 65536;
//   verified against the DOM spaceAfter value plus a resave-preservation
//   check on both AE versions
// @alias       space after,段后间距,paragraph space after,paragraph spacing
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

// @summary     Toggle auto-hyphenation on a text paragraph
// @description Splices the auto-hyphenate flag into the btdk paragraph body
//   for the given paragraph. AE's default is true.
// @param       paraIdx  index of the paragraph to update
// @param       on       whether auto-hyphenation is enabled
// @domain      text
// @stability   stable
// @verify      ae-accept
// @gate        TestTextEnum_AEShipGate_AE2020,TestTextEnum_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk PostScript paragraph body splice; only
//   writeable through the AE 24+ scripting API; verified against the DOM
//   autoHyphenate value plus a resave-preservation check on both AE versions
// @alias       auto hyphenate,自动连字,hyphenation,断字
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

// @summary     Set the leading type of a text paragraph
// @description Splices a new leading-type value into the btdk paragraph
//   body for the given paragraph (btdk paragraph /8).
// @param       paraIdx  index of the paragraph to update
// @param       lt       new leading type
// @domain      text
// @stability   stable
// @verify      ae-accept
// @gate        TestTextEnum_AEShipGate_AE2020,TestTextEnum_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk PostScript paragraph body splice; only
//   writeable through the AE 24+ scripting API; on AE 2020 the enum
//   round-trips as an opaque preserved value, while AE 2025 exposes it
//   through the DOM — verified by re-saving and re-reading on both AE
//   versions
// @alias       leading type,行距类型,roman leading,Japanese leading,段落行距
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

// @summary     Toggle Roman Hanging Punctuation on a text paragraph
// @description Splices the hanging-roman flag into the btdk paragraph body
//   for the given paragraph (btdk paragraph /21). Only meaningful for
//   box-text.
// @param       paraIdx  index of the paragraph to update
// @param       on       whether Roman Hanging Punctuation is enabled
// @domain      text
// @stability   stable
// @verify      ae-accept
// @gate        TestTextEnum_AEShipGate_AE2020,TestTextEnum_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk PostScript paragraph body splice; only
//   writeable through the AE 24+ scripting API; only meaningful for
//   box-text; verified against the DOM hangingRoman value plus a
//   resave-preservation check on both AE versions
// @alias       hanging roman,悬挂标点,Roman Hanging Punctuation,段落悬挂
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

// @summary     Set the reading direction of a text paragraph
// @description Splices a new paragraph-direction value into the btdk
//   paragraph body for the given paragraph (btdk paragraph /33). Affects how
//   mixed left-to-right and right-to-left text composes.
// @param       paraIdx  index of the paragraph to update
// @param       d        new paragraph reading direction
// @domain      text
// @stability   stable
// @verify      ae-accept
// @gate        TestTextEnum_AEShipGate_AE2020,TestTextEnum_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable btdk PostScript paragraph body splice; only
//   writeable through the AE 24+ scripting API; on AE 2020 the enum
//   round-trips as an opaque preserved value, while AE 2025 exposes it
//   through the DOM — verified by re-saving and re-reading on both AE
//   versions
// @alias       paragraph direction,段落方向,LTR,RTL,reading direction,阅读方向,从右到左
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

// @summary     Set per-character manual kerning values on a text layer
// @description Writes per-character manual kerning values into the btdk
//   dict. The values slice length must equal the current per-character count
//   (len(TextSource.ManualKerning)). For the values to actually render, the
//   matching style run's AutoKernType must be TextAutoKernNoAuto — toggle it
//   independently via SetRunAutoKernType. Mirrors the AE scripting
//   TextDocument.kerning behavior: the first-character value is also
//   written to the sibling /1/1[0]/0/7 slot, the first-char scalar reflected
//   by the scripting API.
// @param       values  one manual kerning value (1/1000 em units) per
//   existing character
// @domain      text
// @stability   stable
// @verify      roundtrip
// @since       AE2020
// @boundary    length-variable; rewrites the btdk /1/1[0]/0/8 per-character
//   array in place; the layer must already have a manual-kerning slot — AE
//   only emits that sub-tree once some run has been set to
//   AutoKernType=NoAuto with a non-zero kerning value, so first-time
//   enablement on a layer that lacks the slot is unsupported here and must
//   be created from AE first; the values slice length must equal the
//   existing character count
// @alias       manual kerning,手动字距,kerning values,per-character kerning,字距调整
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
