// Ship-gate verify for ApplyPseudoEffect. Reads pseudo_effect_args.json
// {input, done, matchName, minParams}: opens the Go-written .aep, finds the
// first layer with an Effect Parade, and confirms the pseudo effect is LIVE —
// match-name equals expected, name is not a "Missing:" placeholder, parameter
// count ≥ minParams, enabled — i.e. AE consumed the self-contained pseudo
// definition Go spliced from a .ffx without the preset ever being registered.
// Pure-string single done-write; no resave (close DO_NOT_SAVE + quit).
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/pseudo_effect_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = false;
    try {
        app.open(new File(args.input));
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem) { comp = it; break; }
        }
        if (!comp) throw new Error("no CompItem in project");
        var layer = null;
        for (var li = 1; li <= comp.numLayers; li++) {
            var pr = comp.layer(li).property("ADBE Effect Parade");
            if (pr && pr.numProperties > 0) { layer = comp.layer(li); break; }
        }
        if (!layer) throw new Error("no layer with a non-empty Effect Parade");
        var fx = layer.property("ADBE Effect Parade").property(1);
        log.push("matchName=" + fx.matchName);
        log.push("numProperties=" + fx.numProperties);
        log.push("enabled=" + fx.enabled);
        var codes = [];
        for (var c = 0; c < fx.name.length; c++) codes.push(fx.name.charCodeAt(c));
        log.push("name.charCodes=" + codes.join(","));
        ok = (fx.matchName === args.matchName) &&
             (fx.name.indexOf("Missing") < 0) &&
             (fx.numProperties >= args.minParams) &&
             (fx.enabled === true);
        // Optional CJK / custom display-name check by char-code (keeps non-ASCII
        // out of this source). args.nameCodes = array of expected charCodeAt values.
        if (ok && args.nameCodes) {
            ok = (fx.name.length === args.nameCodes.length);
            for (var k = 0; ok && k < args.nameCodes.length; k++) {
                if (fx.name.charCodeAt(k) !== args.nameCodes[k]) ok = false;
            }
        }
    } catch (e) { log.push("ERROR: " + e.toString() + " line=" + e.line); }

    var done = new File(args.done);
    done.open("w");
    done.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
    done.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
