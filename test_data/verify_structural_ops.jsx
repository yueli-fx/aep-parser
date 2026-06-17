// Acceptance verify for the layer-list structural ops. Reads
// structural_ops_args.json {input, done, resaved}. Opens the Go-built file and
// checks each comp's layer order (top-to-bottom = layer(1)..layer(n)) against
// the expected post-op order, then resaves for the Go side. No render — the ops'
// surface is layer count + order, read from the DOM.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/structural_ops_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    function note(m) { log.push("  " + m); }

    var expected = {
        "DEL":  ["A", "C"],
        "DUP":  ["A", "Bd", "B", "C"],
        "INS":  ["A", "S", "B"],
        "MOVL": ["B", "C", "A"],
        "MTOB": ["C", "A", "B"],
        "MTOE": ["B", "C", "A"],
        "MAFT": ["B", "C", "A"],
        "MBEF": ["C", "A", "B"],
        "SRC":  ["S"]
    };

    try {
        app.open(new File(args.input));

        var comps = {};
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem) comps[it.name] = it;
        }

        for (var name in expected) {
            if (!expected.hasOwnProperty(name)) continue;
            var comp = comps[name];
            if (!comp) { fail(name + " comp missing"); continue; }
            var want = expected[name];
            var got = [];
            for (var li = 1; li <= comp.numLayers; li++) got.push(comp.layer(li).name);
            note(name + " order=[" + got.join(",") + "]");
            if (got.length !== want.length) {
                fail(name + " numLayers=" + got.length + ", want " + want.length + " ([" + want.join(",") + "])");
                continue;
            }
            for (var k = 0; k < want.length; k++) {
                if (got[k] !== want[k]) {
                    fail(name + " layer[" + (k + 1) + "]=" + got[k] + ", want " + want[k] + " (full [" + got.join(",") + "] vs [" + want.join(",") + "])");
                    break;
                }
            }
        }

        app.project.save(new File(args.resaved));
    } catch (e) {
        fail("EXC " + e.toString() + " line=" + e.line);
    }

    var done = new File(args.done);
    done.open("w");
    done.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n"));
    done.close();

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
