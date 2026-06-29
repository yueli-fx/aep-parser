// Creates a shape layer with gradient fill/stroke to RE the GCst → GCky →
// Utf8 (prop.map XML) chunk path.
//
// Output: test_data/generated/fixtures/re_gradient.aep + re_gradient.done
//
// NOTE: ExtendScript can't directly set gradient color stops via setValue
// — the "ADBE Vector Grad Colors" property is XML and AE doesn't expose
// a typed setter. The default 2-stop gradient IS persisted as a GCky/Utf8
// XML chunk as long as the gradient fill/stroke property has been touched
// (e.g. Type / Start Pt / End Pt set). If you need non-default stops,
// reference py-aep's pre-built fixture at
//   workshop/reference/py-aep/samples/models/property/gradient.aep
// which the Go test suite already uses.
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/generated/fixtures/re_gradient.aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/re_gradient.done");
    var log = [];
    function step(name, fn) {
        try {
            fn();
            log.push("OK  " + name);
        } catch (e) {
            log.push("ERR " + name + " -> " + e.toString());
        }
    }

    step("save_as_new_file", function () {
        app.project.save(outFile);
    });

    var testComp = null;
    step("add_test_comp", function () {
        testComp = app.project.items.addComp("GradientTest", 200, 200, 1, 5, 30);
    });

    step("add_shape_with_gradient_stroke", function () {
        if (testComp === null) throw new Error("testComp not set");
        var shapeLayer = testComp.layers.addShape();
        var contents = shapeLayer.property("ADBE Root Vectors Group");
        var rect = contents.addProperty("ADBE Vector Shape - Rect");
        rect.property("ADBE Vector Rect Size").setValue([150, 150]);
        rect.property("ADBE Vector Rect Position").setValue([100, 100]);
        var gradStroke = contents.addProperty("ADBE Vector Graphic - G-Stroke");
        gradStroke.property("ADBE Vector Stroke Width").setValue(10);
        gradStroke.property("ADBE Vector Grad Type").setValue(1); // 1=Linear
        gradStroke.property("ADBE Vector Grad Start Pt").setValue([0, 0]);
        gradStroke.property("ADBE Vector Grad End Pt").setValue([150, 150]);
    });

    // Gradient fill is intentionally skipped — historically
    // contents.addProperty("ADBE Vector Graphic - G-Fill") returns a node
    // whose "ADBE Vector Grad Type" sub-property is null on some AE
    // versions (caller throws). The stroke variant above gives enough
    // GCst/GCky structure to RE. Re-enable + RE separately when needed.

    step("save_again", function () {
        app.project.save();
    });

    // Write .done marker (replaces alert() — keeps ae_run.ps1 contract).
    try {
        doneFile.open("w");
        doneFile.write(log.join("\n"));
        doneFile.close();
    } catch (e) {
        // best-effort
    }

    // Tear down so AE exits cleanly under ae_run.ps1 supervision.
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
