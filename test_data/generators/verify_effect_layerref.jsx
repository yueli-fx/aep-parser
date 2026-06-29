// Generic ship-gate verify for the LAYER-REFERENCE effect family (Displacement
// Map / Compound Blur / CC Vector Blur). Reads effect_layerref_args.json
// {input, done, resaved, png, comp, host, param}. Opens the Go-written input,
// confirms HOST's first effect carries the layer-ref param resolving to a real
// layer (value>=1), resaves, and renders frame 0 so Go can pixel-assert the
// reference visibly took (red line 4).
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/effect_layerref_args.json");
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
            if (it instanceof CompItem && it.name === args.comp) { comp = it; break; }
        }
        if (!comp) { fail("comp " + args.comp + " not found"); }
        else {
            var host = null;
            for (var li = 1; li <= comp.numLayers; li++) {
                if (comp.layer(li).name === args.host) { host = comp.layer(li); break; }
            }
            if (!host) fail("HOST layer " + args.host + " not found");
            else {
                var fx = host.property("ADBE Effect Parade").property(1);
                if (!fx) fail("HOST has no effect");
                else {
                    note("effect[1] = " + fx.matchName);
                    var p = fx.property(args.param);
                    note("layer-ref param " + args.param + " value=" + (p ? p.value : "nil"));
                    if (!p || Number(p.value) < 1) fail("layer-ref param not set to a layer");
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
