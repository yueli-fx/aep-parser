// Ship-gate verify for CROSS-EFFECT enum materialization (Invert Channel).
// Reads effect_enum_args.json {input, done, resaved, png, expect}:
//   expect = array of {layer: "<name>", param: "<full param match-name>",
//            value: <enum int>} — each named layer's Invert effect should read
//            back the materialized Channel enum.
// Opens the Go-written input, reads each layer's Invert-0001 off its Effect
// Parade (by match-name property access — the robust path; .effect() DOM
// accessor is flaky), resaves, and renders frame 0 so Go can assert each enum
// value's distinct rendered fill (red line 4). Writes PASS/FAIL to .done.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/effect_enum_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    function note(m) { log.push("  " + m); }

    try {
        app.open(new File(args.input));

        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "EFFECTENUM") { comp = it; break; }
        }
        if (!comp) { fail("comp EFFECTENUM not found"); }
        else {
            for (var j = 0; j < args.expect.length; j++) {
                var e = args.expect[j];
                var lay = null;
                for (var li = 1; li <= comp.numLayers; li++) {
                    if (comp.layer(li).name === e.layer) { lay = comp.layer(li); break; }
                }
                if (!lay) { fail(e.layer + " layer not found"); continue; }
                var parade = lay.property("ADBE Effect Parade");
                if (!parade || parade.numProperties < 1) { fail(e.layer + " has no Effect Parade"); continue; }
                var fx = parade.property(1); // single Invert effect
                var prop = fx.property(e.param);
                if (!prop) { fail(e.layer + ": param " + e.param + " not found"); continue; }
                var v = Number(prop.value);
                note(e.layer + " " + e.param + "=" + v);
                if (Math.abs(v - e.value) > 0.01) fail(e.layer + ": Channel=" + v + ", want " + e.value);
            }
        }

        app.project.save(new File(args.resaved));

        if (comp && args.png) {
            app.project.bitsPerChannel = 8;
            comp.saveFrameToPng(0, new File(args.png));
            $.sleep(2000);
            var pngFile = new File(args.png);
            if (pngFile.exists) { note("frame png " + pngFile.length + " bytes"); }
            else { fail("saveFrameToPng wrote no file"); }
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
