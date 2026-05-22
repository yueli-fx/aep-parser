// RE script for text-layer btds payload.
//
// Creates a new comp inside the user's currently-open project, adds N
// text layers each varying ONE TextDocument attribute, then redirects
// the project save to test_data/re_text.aep so the user's original
// .aep on disk stays untouched.
//
// Each layer name doubles as the variation tag so the Go-side dumper
// can map "the btds at layer named X" to the variant.
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/re_text.aep");
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

    var comp;
    step("add_comp", function () {
        comp = app.project.items.addComp("RE_TEXT", 1920, 1080, 1, 5, 30);
    });

    function addText(name, opts) {
        var tl = comp.layers.addText(opts.text == null ? "A" : opts.text);
        tl.name = name;
        var td = tl.sourceText.value;
        var trySet = function (k, v) { try { td[k] = v; } catch (e) { log.push("  skip " + name + "." + k + ": " + e.toString()); } };
        if (opts.fontSize != null) trySet("fontSize", opts.fontSize);
        if (opts.fillColor != null) { trySet("applyFill", true); trySet("fillColor", opts.fillColor); }
        if (opts.strokeColor != null) {
            trySet("applyStroke", true);
            trySet("strokeColor", opts.strokeColor);
            trySet("strokeWidth", opts.strokeWidth == null ? 2 : opts.strokeWidth);
        }
        if (opts.fauxBold != null) trySet("fauxBold", opts.fauxBold);
        if (opts.fauxItalic != null) trySet("fauxItalic", opts.fauxItalic);
        if (opts.justification != null) trySet("justification", opts.justification);
        if (opts.tracking != null) trySet("tracking", opts.tracking);
        if (opts.leading != null) trySet("leading", opts.leading);
        if (opts.allCaps != null) trySet("allCaps", opts.allCaps);
        if (opts.smallCaps != null) trySet("smallCaps", opts.smallCaps);
        tl.sourceText.setValue(td);
    }

    // Variants — each isolates one delta from "baseline_A".
    var variants = [
        ["baseline_A",       { text: "A" }],
        ["text_B",           { text: "B" }],
        ["text_AB",          { text: "AB" }],
        ["text_hello",       { text: "Hello" }],
        ["text_unicode",     { text: "你好" }], // 你好
        ["text_two_lines",   { text: "A\rB" }],         // AE uses \r for line break
        ["size_100",         { text: "A", fontSize: 100 }],
        ["size_200",         { text: "A", fontSize: 200 }],
        ["color_red",        { text: "A", fillColor: [1, 0, 0] }],
        ["color_green",      { text: "A", fillColor: [0, 1, 0] }],
        ["color_blue",       { text: "A", fillColor: [0, 0, 1] }],
        ["bold",             { text: "A", fauxBold: true }],
        ["italic",           { text: "A", fauxItalic: true }],
        ["just_center",      { text: "A", justification: ParagraphJustification.CENTER_JUSTIFY }],
        ["just_right",       { text: "A", justification: ParagraphJustification.RIGHT_JUSTIFY }],
        ["all_caps",         { text: "a", allCaps: true }],
        ["tracking_500",     { text: "A", tracking: 500 }],
        ["leading_300",      { text: "A\rB", leading: 300 }],
        ["stroke_yellow",    { text: "A", strokeColor: [1, 1, 0], strokeWidth: 4 }]
    ];

    // Bonus: a box-text variant added separately because the AE API to
    // create one (layers.addBoxText) is different from layers.addText.
    step("add_box_text", function () {
        var btl = comp.layers.addBoxText([400, 200], "BoxText");
        btl.name = "box_text";
    });

    for (var i = 0; i < variants.length; i++) {
        (function (name, opts) {
            step("variant_" + name, function () { addText(name, opts); });
        })(variants[i][0], variants[i][1]);
    }

    step("save", function () {
        app.project.save();
    });

    // Write a sidecar marker so the Go-side runner knows the script
    // finished (instead of polling for AE GUI activity).
    step("write_done_marker", function () {
        var marker = new File("e:/projects/tools/aep-parser/test_data/re_text.done");
        marker.open("w");
        marker.write(log.join("\n"));
        marker.close();
    });
})();
