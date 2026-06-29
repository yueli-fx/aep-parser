// Ship-gate verify for ApplyPseudoEffect. Reads pseudo_effect_args.json
// {input, done, matchName, minParams}: opens the Go-written .aep, finds the
// first layer with an Effect Parade, and confirms the pseudo effect is LIVE —
// match-name equals expected, name is not a "Missing:" placeholder, parameter
// count ≥ minParams, enabled — i.e. AE consumed the self-contained pseudo
// definition Go spliced from a .ffx without the preset ever being registered.
// Pure-string single done-write; no resave (close DO_NOT_SAVE + quit).
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/pseudo_effect_args.json");
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
        // Recursive structure dump (free diagnostics — shows how AE indexed any
        // groups/labels so checks can target the right path).
        var dump = function (g, prefix) {
            for (var pi = 1; pi <= g.numProperties; pi++) {
                var pp = g.property(pi);
                var line = prefix + pi + " " + pp.matchName + " '" + pp.name + "'";
                if (pp.numProperties !== undefined && pp.numProperties > 0) {
                    log.push(line + " {");
                    dump(pp, prefix + "  ");
                    log.push(prefix + "}");
                } else {
                    log.push(line);
                }
            }
        };
        try { dump(fx, "  prop "); } catch (de) { log.push("dump err: " + de.toString()); }
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
        // Optional per-control value/range read-back. args.checks = array of
        // {idx, prop, expect}: idx = 1-based property index within the effect,
        // prop ∈ value|min|max, expect = number. Proves AE honored the
        // synthesized pard defaults. (Index access is used over property(mn)
        // because the latter mis-reports minValue for pseudo sliders.)
        if (ok && args.checks) {
            for (var ci = 0; ci < args.checks.length; ci++) {
                var chk = args.checks[ci];
                // idx may be a number (top-level) or an array path into nested
                // groups, e.g. [4,1] = property 1 of group 4.
                var p = fx, path = (chk.idx instanceof Array) ? chk.idx : [chk.idx];
                for (var pj = 0; p && pj < path.length; pj++) p = p.property(path[pj]);
                if (!p) { log.push("check[" + ci + "] missing idx " + chk.idx); ok = false; break; }
                // Touch hasMin/hasMax before minValue/maxValue: on a freshly
                // fetched pseudo-slider property, minValue reads stale (returns
                // maxValue) unless hasMin is queried first.
                var got;
                if (chk.prop === "min") { if (p.hasMin) {} got = p.minValue; }
                else if (chk.prop === "max") { if (p.hasMax) {} got = p.maxValue; }
                else if (chk.prop === "numProperties") { got = p.numProperties; }
                else got = p.value;
                // expect may be a number or an array (point/3D point value).
                if (chk.expect instanceof Array) {
                    var gs = [];
                    for (var gi = 0; gi < chk.expect.length; gi++) gs.push(String(got[gi]));
                    log.push("check[" + ci + "] [" + chk.idx + "]." + chk.prop + "=[" + gs.join(",") + "]");
                    for (var ei = 0; ei < chk.expect.length; ei++) {
                        if (!got || Math.abs(got[ei] - chk.expect[ei]) > 0.001) ok = false;
                    }
                } else {
                    log.push("check[" + ci + "] [" + chk.idx + "]." + chk.prop + "=" + got);
                    if (Math.abs(got - chk.expect) > 0.001) ok = false;
                }
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
