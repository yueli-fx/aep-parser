// Ship-gate verify for SetEffectParam (effect-parameter materialization,
// synthesis-lite). Reads effect_param_args.json {input, done, resaved, expect}:
//   expect = array of {param: "<full match-name>", value: <number>} the named
//            Gaussian Blur instance should read back after Go materialized +
//            set the params.
// Opens the Go-written input, finds the first layer with an Effect Parade,
// reads each expected param's value off the Gaussian Blur effect, compares
// (small float tolerance), writes PASS/FAIL + diagnostics to .done, and
// resaves so the Go side can confirm AE kept the values.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/effect_param_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  " + m); }

    function checkValues(label) {
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem) { comp = it; break; }
        }
        var fx = null;
        if (!comp) {
            fail(label + ": no CompItem in project");
            return;
        }
        for (var li = 1; li <= comp.numLayers; li++) {
            var p = comp.layer(li).property("ADBE Effect Parade");
            if (p && p.numProperties > 0) {
                fx = p.property("ADBE Gaussian Blur 2");
                if (fx) break;
            }
        }
        if (!fx) {
            fail(label + ": no Gaussian Blur instance found");
            return;
        }
        for (var j = 0; j < args.expect.length; j++) {
            var e = args.expect[j];
            var prop = fx.property(e.param);
            if (!prop) {
                fail(label + ": " + e.param + " not found on effect");
                continue;
            }
            var v = prop.value;
            var num = (v === true) ? 1 : (v === false) ? 0 : Number(v);
            log.push("  " + label + ": " + e.param + " = " + v + " (want " + e.value + ")");
            if (!(Math.abs(num - e.value) < 0.0001)) {
                fail(label + ": " + e.param + ": " + v + " != " + e.value);
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
