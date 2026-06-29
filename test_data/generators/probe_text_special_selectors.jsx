// RE — can "ADBE Text Selectors" take Wiggly + Expression selectors? Minimal,
// bulletproof: each addProperty in its own step(); dump child match-names only.
(function () {
    var out = new File("e:/projects/tools/aep-parser/test_data/fixtures/re_text_special_selectors.aep");
    var done = new File("e:/projects/tools/aep-parser/test_data/probe_text_special_selectors.done");
    var log = [];
    function step(n, f) { try { f(); log.push("OK " + n); } catch (e) { log.push("ERR " + n + " -> " + e); } }
    var sels;
    step("fresh", function () { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); app.newProject(); app.project.save(out); });
    step("build", function () {
        var comp = app.project.items.addComp("SPECSEL", 1280, 720, 1, 5, 24);
        var t = comp.layers.addText("ABCDEF");
        var anim = t.property("ADBE Text Properties").property("ADBE Text Animators").addProperty("ADBE Text Animator");
        anim.property("ADBE Text Animator Properties").addProperty("ADBE Text Opacity").setValue(0);
        sels = anim.property("ADBE Text Selectors");
    });
    function dumpKids(p) {
        var n = 0; try { n = p.numProperties; } catch (e) {}
        log.push("  mn=" + p.matchName + " name=" + p.name + " ptype=" + p.propertyType + " nprops=" + n);
        for (var k = 1; k <= n; k++) {
            try { var c = p.property(k); log.push("    [" + k + "] " + c.matchName + " (" + c.name + ") pt=" + c.propertyType); } catch (e) {}
        }
    }
    step("add Wiggly", function () { dumpKids(sels.addProperty("ADBE Text Wiggly Selector")); });
    step("add Expressible", function () { dumpKids(sels.addProperty("ADBE Text Expressible Selector")); });
    step("save", function () { app.project.save(); });
    try { done.open("w"); done.write(log.join("\n")); done.close(); } catch (e) {}
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
