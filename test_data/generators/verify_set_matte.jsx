// Ship-gate verify for Set Matte (layer-reference effect). Reads
// set_matte_args.json {input, done, resaved, png}. Opens the Go-written input,
// confirms FX carries Set Matte whose "Take Matte From Layer" resolves to a
// real layer, resaves, and renders frame 0 so Go can assert the matte gated FX
// (red left where MATTE channel is on, black right where it is off) — red line 4.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/set_matte_args.json");
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
            if (it instanceof CompItem && it.name === "SETMATTE") { comp = it; break; }
        }
        if (!comp) { fail("comp SETMATTE not found"); }
        else {
            var fx = null;
            for (var li = 1; li <= comp.numLayers; li++) {
                if (comp.layer(li).name === "FX") { fx = comp.layer(li); break; }
            }
            if (!fx) fail("FX layer not found");
            else {
                var sm = fx.property("ADBE Effect Parade").property(1);
                if (!sm) fail("FX has no effect");
                else {
                    var p = sm.property("ADBE Set Matte3-0001");
                    note("Set Matte -0001 (Take Matte From Layer) value=" + (p ? p.value : "nil"));
                    if (!p || Number(p.value) < 1) fail("Take Matte From Layer not set to a layer");
                }
            }

            app.project.save(new File(args.resaved));

            if (args.png) {
                app.project.bitsPerChannel = 8;
                comp.saveFrameToPng(0, new File(args.png));
                $.sleep(2000);
                if (!(new File(args.png)).exists) fail("saveFrameToPng wrote no file");
                else note("frame png " + (new File(args.png)).length + " bytes");
            }
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
