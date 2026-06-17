// RE template generator — author one animator whose Range Selector has ALL 10
// "ADBE Text Range Advanced" params set non-default, so AE materializes every
// cdat slot. extract_text_animator (name override = "ADBE Text Range Advanced")
// then pulls the verbatim Advanced group body into
// internal/serializer/templates/text_range_advanced_body.bin. SetTextRangeAdvanced
// replaces an animator selector's elided Advanced group with this materialized
// one, resets every slot to its AE default, then overwrites the caller's params.
//
// Output: test_data/re_text_range_advanced.aep + .done
(function () {
    var out = new File("e:/projects/tools/aep-parser/test_data/re_text_range_advanced.aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/gen_text_range_advanced_template.done");
    var log = [];
    try { log.push("ae=" + app.version); } catch (e) {}
    function step(name, fn) { try { fn(); log.push("OK  " + name); } catch (e) { log.push("ERR " + name + " -> " + e.toString()); } }

    var tlayer;
    step("fresh", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        app.project.save(out);
    });
    step("build", function () {
        var comp = app.project.items.addComp("ADVTMPL", 1280, 720, 1, 5, 24);
        tlayer = comp.layers.addText("ABCDEF");
        var tp = tlayer.property("ADBE Text Properties");
        var anim = tp.property("ADBE Text Animators").addProperty("ADBE Text Animator");
        anim.property("ADBE Text Selectors").addProperty("ADBE Text Selector");
        anim.property("ADBE Text Animator Properties").addProperty("ADBE Text Opacity").setValue(0);
    });
    // Re-navigate fully each call — enum Advanced sets invalidate live refs.
    function advProp(mn) {
        return tlayer.property("ADBE Text Properties")
            .property("ADBE Text Animators").property(1)
            .property("ADBE Text Selectors").property(1)
            .property("ADBE Text Range Advanced").property(mn);
    }
    // Each set in its own top-level step() so one host-level throw can't abort
    // the rest (inline try/catch does not catch all ExtendScript host errors).
    var sets = [
        ["ADBE Text Range Units", 2],      ["ADBE Text Range Type2", 2],
        ["ADBE Text Selector Mode", 2],    ["ADBE Text Selector Max Amount", 50],
        ["ADBE Text Range Shape", 2],      ["ADBE Text Selector Smoothness", 50],
        ["ADBE Text Levels Max Ease", 30], ["ADBE Text Levels Min Ease", 20],
        ["ADBE Text Randomize Order", 1],  ["ADBE Text Random Seed", 3]
    ];
    for (var i = 0; i < sets.length; i++) {
        (function (mn, v) {
            step("set " + mn + "=" + v, function () { advProp(mn).setValue(v); });
        })(sets[i][0], sets[i][1]);
    }
    step("save", function () { app.project.save(); });
    try { doneFile.open("w"); doneFile.write(log.join("\n")); doneFile.close(); } catch (e) {}
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
