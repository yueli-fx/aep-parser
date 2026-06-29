// Fixture B for the Range Advanced template merge: Shape left at Square (default)
// so Smoothness is enabled, Smoothness set non-default so its cdat materializes.
// (AE hides Smoothness unless Shape=Square — it can't coexist with a non-default
// Shape in one fixture, hence the two-fixture synthesis merge.)
(function () {
    var out = new File("e:/projects/tools/aep-parser/test_data/fixtures/re_text_range_adv_sm.aep");
    var done = new File("e:/projects/tools/aep-parser/test_data/gen_text_range_adv_smoothness.done");
    var log = [];
    function step(n, f) { try { f(); log.push("OK " + n); } catch (e) { log.push("ERR " + n + " -> " + e); } }
    var tlayer;
    step("fresh", function () { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); app.newProject(); app.project.save(out); });
    step("build", function () {
        var comp = app.project.items.addComp("ADVSM", 1280, 720, 1, 5, 24);
        tlayer = comp.layers.addText("ABCDEF");
        var tp = tlayer.property("ADBE Text Properties");
        var anim = tp.property("ADBE Text Animators").addProperty("ADBE Text Animator");
        anim.property("ADBE Text Selectors").addProperty("ADBE Text Selector");
        anim.property("ADBE Text Animator Properties").addProperty("ADBE Text Opacity").setValue(0);
    });
    step("set Smoothness=50", function () {
        tlayer.property("ADBE Text Properties").property("ADBE Text Animators").property(1)
            .property("ADBE Text Selectors").property(1)
            .property("ADBE Text Range Advanced").property("ADBE Text Selector Smoothness").setValue(50);
    });
    step("save", function () { app.project.save(); });
    try { done.open("w"); done.write(log.join("\n")); done.close(); } catch (e) {}
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
