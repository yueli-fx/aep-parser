// Ship-gate verify for batch-7c text enum setters (AE24+ run/paragraph enums).
// Reads text_enum_args.json {input, done, resaved}. Confirms AE accepts the file
// and a text layer exists, logs whatever textDocument enum properties this AE
// version exposes (best-effort — names vary / AE2020 may lack them), and resaves
// so the Go side can do the authoritative re-parse preservation check.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/text_enum_args.json");
    argsFile.open("r"); var raw = argsFile.read(); argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL " + m); }
    function probe(label, getter) {
        try {
            var v = getter();
            if (v === undefined) { log.push("  " + label + " undefined"); return; }
            log.push("  " + label + "=" + v);
        } catch (e) { log.push("  " + label + " N/A: " + e); }
    }

    try {
        app.open(new File(args.input));
        log.push("  AE " + app.version);
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem) { comp = it; break; }
        }
        if (!comp) { fail("no CompItem"); }
        else {
            var L = null;
            for (var li = 1; li <= comp.numLayers; li++) {
                var cand = comp.layer(li);
                if (cand.property("Source Text") !== null) { L = cand; break; }
            }
            if (!L) { fail("no text layer"); }
            else {
                var td = L.property("Source Text").value;
                probe("digitSet", function () { return td.digitSet; });
                probe("autoKernType", function () { return td.autoKernType; });
                probe("baselineDirection", function () { return td.baselineDirection; });
                probe("lineJoinType", function () { return td.lineJoinType; });
                probe("noBreakAttr", function () { return td.noBreakAttr; });
                probe("autoHyphenate", function () { return td.autoHyphenate; });
                probe("leadingType", function () { return td.leadingType; });
                probe("hangingRoman", function () { return td.hangingRoman; });
                probe("direction", function () { return td.direction; });
            }
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
