// RE — does "ADBE Text Selectors" accept a 2nd "ADBE Text Selector" as a clean
// indexed-group append? Build animator, add 2 selectors, dump the Selectors group
// child structure + each selector's range, so we can splice a 2nd selector
// template the same vein as adding an animator.
(function () {
    var out = new File("e:/projects/tools/aep-parser/test_data/re_text_multi_selector.aep");
    var done = new File("e:/projects/tools/aep-parser/test_data/probe_text_multi_selector.done");
    var log = [];
    function step(n, f) { try { f(); log.push("OK " + n); } catch (e) { log.push("ERR " + n + " -> " + e); } }
    var anim;
    step("fresh", function () { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); app.newProject(); app.project.save(out); });
    step("build", function () {
        var comp = app.project.items.addComp("MULTISEL", 1280, 720, 1, 5, 24);
        var t = comp.layers.addText("ABCDEF");
        var tp = t.property("ADBE Text Properties");
        anim = tp.property("ADBE Text Animators").addProperty("ADBE Text Animator");
        anim.property("ADBE Text Animator Properties").addProperty("ADBE Text Opacity").setValue(0);
    });
    step("selector1", function () {
        var s1 = anim.property("ADBE Text Selectors").addProperty("ADBE Text Selector");
        s1.property("ADBE Text Percent Start").setValue(0);
        s1.property("ADBE Text Percent End").setValue(50);
    });
    step("selector2", function () {
        var s2 = anim.property("ADBE Text Selectors").addProperty("ADBE Text Selector");
        s2.property("ADBE Text Percent Start").setValue(50);
        s2.property("ADBE Text Percent End").setValue(100);
    });
    step("dump", function () {
        var sels = anim.property("ADBE Text Selectors");
        log.push("  selectors numProps=" + sels.numProperties + " ptype=" + sels.propertyType);
        for (var k = 1; k <= sels.numProperties; k++) {
            var s = sels.property(k);
            log.push("  [" + k + "] " + s.matchName + " name=" + s.name +
                " start=" + s.property("ADBE Text Percent Start").value +
                " end=" + s.property("ADBE Text Percent End").value);
        }
    });
    step("save", function () { app.project.save(); });
    try { done.open("w"); done.write(log.join("\n")); done.close(); } catch (e) {}
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
