// RE script for AE 24+ Text fields: fontCapsOption / fontBaselineOption /
// strokeOverFill + paragraph indents / spaceBefore / spaceAfter / autoHyphenate.
//
// Requires AE 24+ (script uses APIs introduced in AE 24.0 ScriptingAPI).
// Tested with AE 2025.
//
// Each layer name doubles as the variation tag. Box-text variants live in
// the second comp so the box-text paragraph fields aren't no-ops.
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/re_text_caps_ae24.aep");
    var log = [];
    function step(name, fn) {
        try {
            fn();
            log.push("OK  " + name);
        } catch (e) {
            log.push("ERR " + name + " -> " + e.toString());
        }
    }

    step("redirect_save", function () {
        app.project.save(outFile);
    });

    var compPoint, compBox;
    step("add_comp_point", function () {
        compPoint = app.project.items.addComp("RE_CAPS_POINT", 1920, 1080, 1, 5, 30);
    });
    step("add_comp_box", function () {
        compBox = app.project.items.addComp("RE_CAPS_BOX", 1920, 1080, 1, 5, 30);
    });

    // ---- per-character / per-run helpers (point text) ----

    function addPointText(name, opts) {
        var tl = compPoint.layers.addText(opts.text == null ? "AaBb" : opts.text);
        tl.name = name;
        var td = tl.sourceText.value;
        var trySet = function (k, v) { try { td[k] = v; } catch (e) { log.push("  skip " + name + "." + k + ": " + e.toString()); } };
        if (opts.fontCapsOption != null) trySet("fontCapsOption", opts.fontCapsOption);
        if (opts.fontBaselineOption != null) trySet("fontBaselineOption", opts.fontBaselineOption);
        if (opts.applyStroke != null) trySet("applyStroke", opts.applyStroke);
        if (opts.strokeColor != null) trySet("strokeColor", opts.strokeColor);
        if (opts.strokeWidth != null) trySet("strokeWidth", opts.strokeWidth);
        if (opts.strokeOverFill != null) trySet("strokeOverFill", opts.strokeOverFill);
        tl.sourceText.setValue(td);
    }

    // baseline keeps everything default — gives us a clean diff target.
    var pointVariants = [
        ["pt_baseline",        { text: "AaBb" }],
        // fontCapsOption (AE 24.0+)
        ["pt_caps_all",        { text: "AaBb", fontCapsOption: FontCapsOption.FONT_ALL_CAPS }],
        ["pt_caps_small",      { text: "AaBb", fontCapsOption: FontCapsOption.FONT_SMALL_CAPS }],
        ["pt_caps_all_small",  { text: "AaBb", fontCapsOption: FontCapsOption.FONT_ALL_SMALL_CAPS }],
        // fontBaselineOption (AE 24.0+)
        ["pt_base_super",      { text: "X2",   fontBaselineOption: FontBaselineOption.FONT_FAUXED_SUPERSCRIPT }],
        ["pt_base_sub",        { text: "H2O",  fontBaselineOption: FontBaselineOption.FONT_FAUXED_SUBSCRIPT }],
        // strokeOverFill (always existed but AE 2020 readonly via JSX)
        ["pt_stroke_over",     { text: "AaBb", applyStroke: true, strokeColor: [1, 0, 0], strokeWidth: 4, strokeOverFill: true }],
        ["pt_stroke_under",    { text: "AaBb", applyStroke: true, strokeColor: [0, 1, 0], strokeWidth: 4, strokeOverFill: false }]
    ];

    for (var i = 0; i < pointVariants.length; i++) {
        (function (name, opts) {
            step("pt_" + name, function () { addPointText(name, opts); });
        })(pointVariants[i][0], pointVariants[i][1]);
    }

    // ---- paragraph fields (box text required) ----

    function addBoxText(name, opts) {
        // addBoxText signature: layers.addBoxText(size, [sourceText])
        var btl = compBox.layers.addBoxText([800, 400], opts.text == null ? "Hello World hyphenation test paragraph" : opts.text);
        btl.name = name;
        var td = btl.sourceText.value;
        var trySet = function (k, v) { try { td[k] = v; } catch (e) { log.push("  skip " + name + "." + k + ": " + e.toString()); } };
        if (opts.startIndent != null) trySet("startIndent", opts.startIndent);
        if (opts.endIndent != null) trySet("endIndent", opts.endIndent);
        if (opts.firstLineIndent != null) trySet("firstLineIndent", opts.firstLineIndent);
        if (opts.spaceBefore != null) trySet("spaceBefore", opts.spaceBefore);
        if (opts.spaceAfter != null) trySet("spaceAfter", opts.spaceAfter);
        if (opts.autoHyphenate != null) trySet("autoHyphenate", opts.autoHyphenate);
        btl.sourceText.setValue(td);
    }

    var boxVariants = [
        ["bx_baseline",        {}],
        // paragraph indents (AE 24.0+)
        ["bx_startIndent_50",  { startIndent: 50 }],
        ["bx_endIndent_60",    { endIndent: 60 }],
        ["bx_firstLine_70",    { firstLineIndent: 70 }],
        ["bx_spaceBefore_30",  { spaceBefore: 30 }],
        ["bx_spaceAfter_40",   { spaceAfter: 40 }],
        // autoHyphenate (AE 24.0+)
        ["bx_hyphenate_on",    { autoHyphenate: true }],
        ["bx_hyphenate_off",   { autoHyphenate: false }]
    ];

    for (var j = 0; j < boxVariants.length; j++) {
        (function (name, opts) {
            step("bx_" + name, function () { addBoxText(name, opts); });
        })(boxVariants[j][0], boxVariants[j][1]);
    }

    step("save", function () {
        app.project.save();
    });

    step("write_done_marker", function () {
        var marker = new File("e:/projects/tools/aep-parser/test_data/re_text_caps_ae24.done");
        marker.open("w");
        marker.write(log.join("\n"));
        marker.close();
    });
})();
