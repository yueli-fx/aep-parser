// RE: what AE does to the per-character manual-kerning table (/1/1/0/0/8/0)
// when the text length changes — the last guard in length-variable SetText.
//
// td.kerning (AE 24+) sets a uniform manual-kerning value across all chars,
// materializing the /8 slot (postdates incidents/kerning-first-enable.md).
//
// Produces re_text_kern_resize.aep (AE 2024):
//   kern_src    — "AaBb" kerning -50           (4-char control: slot [-50×4])
//   kern_grow   — "AaBb" kerning -50, then text="Hello"  (5 chars)
//   kern_shrink — "AaBb" kerning -50, then text="Xy"     (2 chars)
//
// Go-side dumps /1/1/0/0/8/0 to learn the resize/drop rule.
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/fixtures/re_text_kern_resize.aep");
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
        comp = app.project.items.addComp("RE_KERN", 1920, 1080, 1, 5, 30);
    });

    function kernLayer(name, newText) {
        var tl = comp.layers.addText("AaBb");
        tl.name = name;
        var td = tl.sourceText.value;
        td.kerning = -50;            // uniform manual kerning → materializes /8
        if (newText != null) td.text = newText;
        tl.sourceText.setValue(td);
        return tl;
    }

    step("kern_src",    function () { kernLayer("kern_src", null); });
    step("kern_grow",   function () { kernLayer("kern_grow", "Hello"); });
    step("kern_shrink", function () { kernLayer("kern_shrink", "Xy"); });

    step("save", function () { app.project.save(); });

    step("write_done_marker", function () {
        var marker = new File("e:/projects/tools/aep-parser/test_data/re_text_kern_resize.done");
        marker.open("w");
        marker.write(log.join("\n"));
        marker.close();
    });

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
