// RE: where does AE place ADBE Effect Parade in a SHAPE layer's property tree?
// Create a shape layer with a rectangle + a Gaussian Blur effect, save. The Go
// side then dumps the shape Layr's outer tdgp group order to locate the parade
// relative to Transform / Layer Styles / Material Options etc. (AddEffect Phase 2).
(function () {
    var outFile  = new File("e:/projects/tools/aep-parser/test_data/generated/fixtures/re_shape_effect.aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/re_shape_effect.done");
    var log = [];
    try { log.push("ae=" + app.version); } catch (e) {}
    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        var comp = app.project.items.addComp("ShapeFx", 1920, 1080, 1, 5, 24);
        var layer = comp.layers.addShape();
        layer.name = "ShapeFx";
        // add a rectangle so the shape is non-trivial
        var contents = layer.property("ADBE Root Vectors Group");
        var grp = contents.addProperty("ADBE Vector Group");
        grp.property("ADBE Vectors Group").addProperty("ADBE Vector Shape - Rect");
        grp.property("ADBE Vectors Group").addProperty("ADBE Vector Graphic - Fill");
        // add a Gaussian Blur effect
        layer.property("ADBE Effect Parade").addProperty("ADBE Gaussian Blur 2");
        app.project.save(outFile);
        log.push("OK saved; numEffects=" + layer.property("ADBE Effect Parade").numProperties);
    } catch (e) {
        log.push("EXC " + e.toString());
    }
    doneFile.open("w");
    doneFile.write((log.join("\n").indexOf("EXC") === -1 ? "PASS" : "FAIL") + "\n" + log.join("\n"));
    doneFile.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
