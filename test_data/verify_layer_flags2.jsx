// Ship-gate verify for SetEffectsEnabled / SetTimeRemapEnabled / SetIsNull.
// Reads layer_flags2_args.json {input, done, resaved}.
//   EON  effectsActive=true   EOFF effectsActive=false   NUL nullLayer=true
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/layer_flags2_args.json");
    argsFile.open("r"); var raw = argsFile.read(); argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL " + m); }
    function eqBool(name, got, want) {
        log.push("  " + name + "=" + got);
        if (!!got !== !!want) fail(name + "=" + got + ", want " + want);
    }

    try {
        app.open(new File(args.input));
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "MAIN") { comp = it; break; }
        }
        if (!comp) { fail("comp MAIN not found"); }
        else {
            var byName = {};
            for (var li = 1; li <= comp.numLayers; li++) byName[comp.layer(li).name] = comp.layer(li);
            function get(n) { var L = byName[n]; if (!L) fail(n + " missing"); return L; }

            var EON = get("EON"), EOFF = get("EOFF"), NUL = get("NUL");
            if (EON) eqBool("EON.effectsActive", EON.effectsActive, true);
            if (EOFF) eqBool("EOFF.effectsActive", EOFF.effectsActive, false);
            if (NUL) eqBool("NUL.nullLayer", NUL.nullLayer, true);
        }
        app.project.save(new File(args.resaved));
    } catch (e) {
        fail("EXC " + e.toString());
    }

    var done = new File(args.done);
    done.open("w"); done.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n")); done.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
