// AE 24+ TextDocument 扩展字段 RE 工程：
//
//   kerning            (per-char int)              AE 24+
//   ligature           (bool)                       AE 24+
//   direction          (RTL/LTR enum)               AE 24+
//   lineOrientation    (Horiz / VertRTL / VertLTR)  AE 24.2+
//   lineJoinType       (Miter / Round / Bevel)      AE 24+
//   noBreak            (bool)                       AE 24+
//   composerEngine     (LatinCJK / Universal)       AE 24+
//   digitSet           (Default/Arabic/Hindi/...)    AE 24+
//   autoKernType       (NoAuto / Metric / Optical)  AE 24+
//   leadingType        (Roman / Japanese)           AE 24+
//   hangingRoman       (bool)                        AE 24+
//
// Box-text only fields (RE_TEXT24_BX comp):
//   boxAutoFitPolicy / boxFirstBaselineAlignment / boxInsetSpacing /
//   boxVerticalAlignment   (all AE 24.6+)
//
// Each variant isolates one delta from "baseline" so Go-side diff per
// btdk PostScript dict reveals the field offset.

(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/fixtures/re_text_ae24_more.aep");
    var log = [];
    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }

    step("redirect_save", function () { app.project.save(outFile); });

    // ── per-character fields (point text comp) ──
    var compPt;
    step("add_comp_pt", function () {
        compPt = app.project.items.addComp("RE_TEXT24_PT", 1920, 1080, 1, 5, 30);
    });

    function addPoint(name, opts) {
        var tl = compPt.layers.addText(opts.text == null ? "AaBb" : opts.text);
        tl.name = name;
        var td = tl.sourceText.value;
        var trySet = function (k, v) { try { td[k] = v; } catch (e) { log.push("  skip " + name + "." + k + ": " + e.toString()); } };
        if (opts.kerning != null)        trySet("kerning", opts.kerning);
        if (opts.ligature != null)       trySet("ligature", opts.ligature);
        if (opts.lineJoinType != null)   trySet("lineJoinType", opts.lineJoinType);
        if (opts.noBreak != null)        trySet("noBreak", opts.noBreak);
        if (opts.autoKernType != null)   trySet("autoKernType", opts.autoKernType);
        if (opts.digitSet != null)       trySet("digitSet", opts.digitSet);
        if (opts.baselineDirection != null) trySet("baselineDirection", opts.baselineDirection);
        tl.sourceText.setValue(td);
    }

    var pointVariants = [
        ["pt_baseline",         {}],
        ["pt_kerning_100",      { kerning: 100 }],
        ["pt_kerning_neg50",    { kerning: -50 }],
        ["pt_ligature_false",   { ligature: false }],
        ["pt_linejoin_round",   { lineJoinType: LineJoinType.LINE_JOIN_ROUND }],
        ["pt_linejoin_bevel",   { lineJoinType: LineJoinType.LINE_JOIN_BEVEL }],
        ["pt_nobreak_true",     { noBreak: true }],
        ["pt_autokern_metric",  { autoKernType: AutoKernType.METRIC_KERN }],
        ["pt_autokern_optical", { autoKernType: AutoKernType.OPTICAL_KERN }],
        ["pt_digitset_arabic",  { digitSet: DigitSet.ARABIC_DIGITS }],
        ["pt_digitset_hindi",   { digitSet: DigitSet.HINDI_DIGITS }],
        ["pt_baseline_vert",    { baselineDirection: BaselineDirection.BASELINE_VERTICAL_ROTATED }]
    ];
    for (var i = 0; i < pointVariants.length; i++) {
        (function (name, opts) {
            step("pt_" + name, function () { addPoint(name, opts); });
        })(pointVariants[i][0], pointVariants[i][1]);
    }

    // ── paragraph-level fields (need box text) ──
    var compBx;
    step("add_comp_bx", function () {
        compBx = app.project.items.addComp("RE_TEXT24_BX", 1920, 1080, 1, 5, 30);
    });

    function addBox(name, opts) {
        var btl = compBx.layers.addBoxText([800, 400], opts.text == null ? "Hello World" : opts.text);
        btl.name = name;
        var td = btl.sourceText.value;
        var trySet = function (k, v) { try { td[k] = v; } catch (e) { log.push("  skip " + name + "." + k + ": " + e.toString()); } };
        if (opts.direction != null)        trySet("direction", opts.direction);
        if (opts.lineOrientation != null)  trySet("lineOrientation", opts.lineOrientation);
        if (opts.composerEngine != null)   trySet("composerEngine", opts.composerEngine);
        if (opts.everyLineComposer != null) trySet("everyLineComposer", opts.everyLineComposer);
        if (opts.hangingRoman != null)     trySet("hangingRoman", opts.hangingRoman);
        if (opts.leadingType != null)      trySet("leadingType", opts.leadingType);
        btl.sourceText.setValue(td);
    }

    var boxVariants = [
        ["bx_baseline",                  {}],
        ["bx_direction_rtl",             { direction: ParagraphDirection.DIRECTION_RIGHT_TO_LEFT }],
        ["bx_lineorient_vert_rtl",       { lineOrientation: LineOrientation.VERTICAL_RIGHT_TO_LEFT }],
        ["bx_lineorient_vert_ltr",       { lineOrientation: LineOrientation.VERTICAL_LEFT_TO_RIGHT }],
        ["bx_composer_latin_cjk",        { composerEngine: ComposerEngine.LATIN_CJK_ENGINE }],
        ["bx_everylinecomposer_false",   { everyLineComposer: false }],
        ["bx_hangingroman_true",         { hangingRoman: true }],
        ["bx_leadingtype_japanese",      { leadingType: LeadingType.JAPANESE_LEADING_TYPE }]
    ];
    for (var j = 0; j < boxVariants.length; j++) {
        (function (name, opts) {
            step("bx_" + name, function () { addBox(name, opts); });
        })(boxVariants[j][0], boxVariants[j][1]);
    }

    step("save", function () { app.project.save(); });

    step("write_done_marker", function () {
        var marker = new File("e:/projects/tools/aep-parser/test_data/re_text_ae24_more.done");
        marker.open("w");
        marker.write(log.join("\n"));
        marker.close();
    });
})();
