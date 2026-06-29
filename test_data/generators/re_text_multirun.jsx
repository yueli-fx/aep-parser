// RE: how AE represents multiple style runs in btdk, and what changing the
// text on a multi-run doc does to them — the open question blocking multi-run
// SetText.
//
// Produces re_text_multirun.aep (AE 2025; characterRange is AE 24+):
//   multirun_src       — "Hello": chars 0-2 red, 2-5 blue (two style runs)
//   multirun_setvalue  — two-run doc, then td.text="World!!" + setValue
//   multirun_charrange — two-run doc, then characterRange(0,5).text="ABCDEFGH"
//
// Go-side dumps each layer's /1/1/0/0/6/0 run array (entry count + per-entry /1
// + fill color) to learn AE's run-redistribution rule.
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/fixtures/re_text_multirun.aep");
    var log = [];
    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }

    step("fresh_project", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        app.project.save(outFile);
    });

    var comp;
    step("add_comp", function () {
        comp = app.project.items.addComp("RE_MULTIRUN", 1920, 1080, 1, 5, 30);
    });

    // Build a two-run layer: apply per-range fill on ONE held document, then
    // set it back so the style runs persist into btdk.
    function twoRun(name) {
        var tl = comp.layers.addText("Hello");
        tl.name = name;
        var td = tl.sourceText.value;            // one snapshot we keep
        td.characterRange(0, 2).fillColor = [1, 0, 0];
        td.characterRange(2, 5).fillColor = [0, 0, 1];
        tl.sourceText.setValue(td);
        return tl;
    }

    step("multirun_src", function () { twoRun("multirun_src"); });

    step("multirun_setvalue", function () {
        var tl = twoRun("multirun_setvalue");
        var td = tl.sourceText.value;            // current (two-run) document
        td.text = "World!!";                     // change just the text
        tl.sourceText.setValue(td);
        log.push("  setvalue readback len=" + tl.sourceText.value.text.length);
    });

    step("multirun_charrange", function () {
        var tl = twoRun("multirun_charrange");
        var td = tl.sourceText.value;
        td.characterRange(0, 5).text = "ABCDEFGH";   // replace via range
        tl.sourceText.setValue(td);
        log.push("  charrange readback len=" + tl.sourceText.value.text.length);
    });

    step("save", function () { app.project.save(); });

    step("write_done_marker", function () {
        var marker = new File("e:/projects/tools/aep-parser/test_data/re_text_multirun.done");
        marker.open("w");
        marker.write(log.join("\n"));
        marker.close();
    });

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
