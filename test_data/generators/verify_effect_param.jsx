// Ship-gate verify for SetEffectParam (effect-parameter materialization,
// synthesis-lite). Reads effect_param_args.json {input, done, resaved, expect}:
//   expect = array of {effect: "<effect match-name>", param: "<full param
//            match-name>", value: <number or array>} the layer's effects
//            should read back (in JSX UI units — degrees / rgba 0-1 / layer
//            pixels) after Go materialized + set the params.
// Opens the Go-written input, finds the first layer with an Effect Parade,
// reads each expected param's value off its effect, compares (small float
// tolerance), resaves, then REOPENS the resave and reads again — a param set
// to a value equal to this AE version's default is legitimately re-elided
// from the file, but its effective value must survive via the default.
// Writes PASS/FAIL + diagnostics to .done.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/effect_param_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  " + m); }
    // String-concatenating an AE color value ("" + v) throws "Invalid numeric
    // result (divide by zero?)" — ExtendScript routes the array through
    // valueOf. Stringify arrays explicitly via join.
    function str(x) {
        try { return (x instanceof Array) ? x.join(",") : ("" + x); }
        catch (e) { return "<unprintable>"; }
    }

    function checkValues(label) {
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem) { comp = it; break; }
        }
        var parade = null;
        if (!comp) {
            fail(label + ": no CompItem in project");
            return;
        }
        for (var li = 1; li <= comp.numLayers; li++) {
            var p = comp.layer(li).property("ADBE Effect Parade");
            if (p && p.numProperties > 0) { parade = p; break; }
        }
        if (!parade) {
            fail(label + ": no layer with effects found");
            return;
        }
        for (var j = 0; j < args.expect.length; j++) {
            var e = args.expect[j];
            var fx = parade.property(e.effect);
            if (!fx) {
                fail(label + ": effect " + e.effect + " not found");
                continue;
            }
            var prop = fx.property(e.param);
            if (!prop) {
                fail(label + ": " + e.param + " not found on effect");
                continue;
            }
            var v = prop.value;
            log.push("  " + label + ": " + e.param + " = " + str(v) + " (want " + str(e.value) + ")");
            if (e.value instanceof Array) {
                var bad = !(v instanceof Array);
                if (!bad) {
                    for (var k = 0; k < e.value.length; k++) {
                        if (!(Math.abs(Number(v[k]) - e.value[k]) < 0.0001)) bad = true;
                    }
                }
                if (bad) fail(label + ": " + e.param + ": " + str(v) + " != " + str(e.value));
            } else {
                var num = (v === true) ? 1 : (v === false) ? 0 : Number(v);
                if (!(Math.abs(num - e.value) < 0.0001)) {
                    fail(label + ": " + e.param + ": " + str(v) + " != " + str(e.value));
                }
            }
        }
    }

    try {
        app.open(new File(args.input));
        checkValues("input");
        app.project.save(new File(args.resaved));

        // Reopen the resave: the effective values must survive in AE's own
        // canonical form — a param may legitimately drop out of the file when
        // the set value equals THIS AE version's default (value-keyed
        // elision), but reading it back must still yield the set value.
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.open(new File(args.resaved));
        checkValues("resaved");
    } catch (e2) {
        fail("EXC " + e2.toString());
    }

    var doneFile = new File(args.done);
    doneFile.open("w");
    doneFile.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n"));
    doneFile.close();

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e3) {}
    try { app.quit(); } catch (e4) {}
})();
