// RE discovery — enumerate the "ADBE Text Range Advanced" sub-group's fixed
// child properties (Units / Based On / Mode / Amount / Shape / Smoothness /
// Ease High·Low / Randomize Order / Random Seed) so we can add Add*/Set*
// facades for the render-gateable ones (same 抽模板 + overwrite cdat vein as the
// animator leaves). Advanced is a NAMED group, so property(k) enumerates its
// REAL fixed children (unlike Animator Properties, which exposes the full
// schema). For each: report matchName / name / propertyType / propertyValueType
// / value; set a non-default value where settable to materialize the cdat.
//
// Output: test_data/fixtures/probe_text_range_advanced.aep + .done
(function () {
    var base = "e:/projects/tools/aep-parser/test_data/probe_text_range_advanced";
    var outFile  = new File(base + ".aep");
    var doneFile = new File(base + ".done");
    var log = [];
    try { log.push("ae=" + app.version); } catch (e) {}
    function step(name, fn) { try { fn(); log.push("OK  " + name); } catch (e) { log.push("ERR " + name + " -> " + e.toString()); } }

    var comp, tlayer, sel, adv;
    step("fresh", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        app.project.save(outFile);
    });
    step("build", function () {
        comp = app.project.items.addComp("ADV", 1280, 720, 1, 5, 24);
        tlayer = comp.layers.addText("ABCDEF");
        var tp = tlayer.property("ADBE Text Properties");
        var animators = tp.property("ADBE Text Animators");
        var anim = animators.addProperty("ADBE Text Animator");
        var sels = anim.property("ADBE Text Selectors");
        sel = sels.addProperty("ADBE Text Selector");
        // give the animator an Opacity leaf so Amount/Shape have something to act on
        anim.property("ADBE Text Animator Properties").addProperty("ADBE Text Opacity").setValue(0);
        adv = sel.property("ADBE Text Range Advanced");
    });
    step("enumerate", function () {
        log.push("advanced numProps=" + adv.numProperties);
        for (var k = 1; k <= adv.numProperties; k++) {
            var line = "  [" + k + "]";
            try {
                var p = adv.property(k);
                try { line += " mn=" + p.matchName; } catch (e) {}
                try { line += " name=" + p.name; } catch (e) {}
                try { line += " ptype=" + p.propertyType; } catch (e) {}
                try { line += " vtype=" + p.propertyValueType; } catch (e) {}
                try { line += " value=" + String(p.value); } catch (e) { line += " value=<" + e + ">"; }
            } catch (eo) { line += " <prop err " + eo + ">"; }
            log.push(line);
        }
    });
    // Try common enum sets explicitly (Mode/Shape) by matchName guesses + by index 1/3.
    step("save", function () { app.project.save(); });
    try { doneFile.open("w"); doneFile.write(log.join("\n")); doneFile.close(); } catch (e) {}
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
