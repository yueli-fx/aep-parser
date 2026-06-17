// RE: how AE stores a Set Matte effect's "Take Matte From Layer" param — a
// LAYER-reference param pointing at a DIFFERENT layer than the host (the
// deferred tdpi-to-non-host case, coverage.md). Builds a comp with a MATTE
// source layer + an FX layer carrying Set Matte that references MATTE, saves so
// extract_effect_lib + dump can locate the layer-ref bytes (tdpi? cdat? a
// separate chunk?).
(function () {
    var dir = "e:/projects/tools/aep-parser/test_data/";
    var outFile = new File(dir + "re_set_matte.aep");
    var log = [];
    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString() + " line=" + e.line); }
    }

    var comp, matte, fx, sm;

    step("fresh_project", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
    });

    step("build", function () {
        comp = app.project.items.addComp("setmatte", 1920, 1080, 1, 5, 30);
        matte = comp.layers.addSolid([0, 1, 0], "MATTE", 1920, 1080, 1); // green
        fx = comp.layers.addSolid([1, 0, 0], "FX", 1920, 1080, 1);       // red, added on top
        log.push("  FX index=" + fx.index + " MATTE index=" + matte.index + " id=" + matte.id);
    });

    step("add_set_matte", function () {
        sm = fx.property("ADBE Effect Parade").addProperty("ADBE Set Matte3");
        log.push("  Set Matte numProps=" + sm.numProperties);
        for (var i = 1; i <= sm.numProperties; i++) {
            var pr = sm.property(i);
            log.push("    [" + i + "] " + pr.matchName + " name=" + pr.name);
        }
    });

    // Try several ways to point the layer-ref param at MATTE; log which works.
    step("set_matte_ref", function () {
        var p = sm.property("ADBE Set Matte3-0001"); // Take Matte From Layer
        if (!p) { log.push("  no -0001 param"); return; }
        log.push("  -0001 matchName=" + p.matchName + " name=" + p.name + " pre.value=" + p.value);
        var ok = false;
        try { p.setValue(matte.index); ok = true; log.push("  setValue(index) OK value=" + p.value); } catch (e) { log.push("  setValue(index) ERR " + e); }
        if (!ok) {
            try { p.setValue(matte); ok = true; log.push("  setValue(layer) OK value=" + p.value); } catch (e2) { log.push("  setValue(layer) ERR " + e2); }
        }
    });

    step("save", function () { app.project.save(outFile); });

    step("write_done", function () {
        var marker = new File(dir + "re_set_matte.done");
        marker.open("w");
        marker.write(log.join("\n"));
        marker.close();
    });

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
