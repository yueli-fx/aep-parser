// Ship-gate verify for animated COLOR / POINT effect params
// (AnimateEffectParamVec). Reads anim_effect_vec_args.json
// {input, done, resaved, fill_t0, fill_t2, ramp_t0, ramp_t2}.
// Opens the Go-written input, sanity-checks each animated param's numKeys via
// match-name property access, resaves, and renders FILLANIM + RAMPANIM at t=0
// and t=2 so Go can assert the color/point animation at the render surface
// (red line 4). Writes PASS/FAIL to .done.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/anim_effect_vec_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    function note(m) { log.push("  " + m); }

    function findComp(name) {
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === name) return it;
        }
        return null;
    }
    function render(comp, t, file) {
        app.project.bitsPerChannel = 8;
        comp.saveFrameToPng(t, new File(file));
        $.sleep(1500);
        if (!(new File(file)).exists) fail("saveFrameToPng wrote no file: " + file);
        else note("png " + file + " " + (new File(file)).length + "b");
    }

    try {
        app.open(new File(args.input));

        var fc = findComp("FILLANIM");
        var rc = findComp("RAMPANIM");
        if (!fc) fail("FILLANIM not found");
        if (!rc) fail("RAMPANIM not found");

        if (fc) {
            var fp = fc.layer(1).property("ADBE Effect Parade").property(1).property("ADBE Fill-0002");
            note("Fill Color numKeys=" + (fp ? fp.numKeys : "nil"));
            if (!fp || fp.numKeys < 2) fail("Fill Color not animated (numKeys<2)");
        }
        if (rc) {
            var rpp = rc.layer(1).property("ADBE Effect Parade").property(1).property("ADBE Ramp-0001");
            note("Ramp Start numKeys=" + (rpp ? rpp.numKeys : "nil"));
            if (!rpp || rpp.numKeys < 2) fail("Ramp Start not animated (numKeys<2)");
        }

        app.project.save(new File(args.resaved));

        if (fc) {
            render(fc, 0, args.fill_t0);
            render(fc, 2, args.fill_t2);
        }
        if (rc) {
            render(rc, 0, args.ramp_t0);
            render(rc, 2, args.ramp_t2);
        }
    } catch (e2) {
        fail("EXC " + e2.toString() + " line=" + e2.line);
    }

    var done = new File(args.done);
    done.open("w");
    done.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n"));
    done.close();

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e3) {}
    try { app.quit(); } catch (e4) {}
})();
