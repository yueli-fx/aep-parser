// RE script for the two un-RE'd SetText cases: empty text + 3-paragraph text.
//
// Produces re_text_multipara.aep with layers:
//   empty_addText   — comp.layers.addText("")           (if AE allows it)
//   empty_setValue  — addText("A") then sourceText="" via setValue
//   three_para      — "L1\rL2\rL3"   (3 paragraphs, uniform style)
//   one_para        — "X"            (1-para control, same project/fonts)
//
// Go-side dumps each layer's btdk to compare paragraph-array / counter /
// string structure against the single-paragraph baseline.
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/fixtures/re_text_multipara.aep");
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
        comp = app.project.items.addComp("RE_MULTIPARA", 1920, 1080, 1, 5, 30);
    });

    step("empty_addText", function () {
        var tl = comp.layers.addText("");
        tl.name = "empty_addText";
        // probe what the runtime reports the value as
        try { log.push("  empty_addText value=" + JSON.stringify(tl.sourceText.value.text)); } catch (e) {}
    });

    step("empty_setValue", function () {
        var tl = comp.layers.addText("A");
        tl.name = "empty_setValue";
        var td = tl.sourceText.value;
        td.text = "";
        tl.sourceText.setValue(td);
        try { log.push("  empty_setValue value=" + JSON.stringify(tl.sourceText.value.text)); } catch (e) {}
    });

    step("three_para", function () {
        var tl = comp.layers.addText("L1\rL2\rL3");
        tl.name = "three_para";
    });

    step("one_para", function () {
        var tl = comp.layers.addText("X");
        tl.name = "one_para";
    });

    step("save", function () { app.project.save(); });

    step("write_done_marker", function () {
        var marker = new File("e:/projects/tools/aep-parser/test_data/re_text_multipara.done");
        marker.open("w");
        marker.write(log.join("\n"));
        marker.close();
    });

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
