// Ship-gate verify for PropertyBase structural ops (Remove / MoveTo) on the
// Effect Parade. Reads property_struct_args.json {input, done, resaved, expect}:
//   expect = array of effect match-names in the order AE should read back after
//            Go applied the mutation.
// Opens the Go-written input, finds the layer carrying an Effect Parade, reads
// each effect's matchName, compares to expect, writes PASS/FAIL + diagnostics
// to .done, and resaves so the Go side can confirm AE kept the change.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/property_struct_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  " + m); }

    try {
        app.open(new File(args.input));

        // Find the comp, then the first layer with an Effect Parade.
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem) { comp = it; break; }
        }
        if (!comp) {
            fail("no CompItem in project");
        } else {
            var parade = null;
            for (var li = 1; li <= comp.numLayers; li++) {
                var p = comp.layer(li).property("ADBE Effect Parade");
                if (p && p.numProperties > 0) { parade = p; break; }
            }
            if (!parade) {
                fail("no layer with a non-empty Effect Parade");
            } else {
                var got = [];
                for (var k = 1; k <= parade.numProperties; k++) {
                    got.push(parade.property(k).matchName);
                }
                log.push("  parade=[" + got.join(", ") + "]");
                log.push("  expect=[" + args.expect.join(", ") + "]");
                if (got.length !== args.expect.length) {
                    fail("count " + got.length + " != expected " + args.expect.length);
                } else {
                    for (var j = 0; j < got.length; j++) {
                        if (got[j] !== args.expect[j]) {
                            fail("index " + j + ": " + got[j] + " != " + args.expect[j]);
                        }
                    }
                }
            }
        }

        app.project.save(new File(args.resaved));
    } catch (e) {
        fail("EXC " + e.toString());
    }

    var done = new File(args.done);
    done.open("w");
    done.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n"));
    done.close();

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
