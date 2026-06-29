// Ship-gate verify for length-variable SetText. Reads set_text_args.json
// {input, done, resaved, expect:{layerName:text,...}}. Opens the Go-written
// file, confirms AE accepts it and each named TextLayer reads back the
// expected (length-changed) string, resaves so the Go side can confirm AE
// kept the texts.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/set_text_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  " + m); }

    try {
        app.open(new File(args.input));

        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem) { comp = it; break; }
        }
        if (!comp) {
            fail("no CompItem in project");
        } else {
            var found = {};
            var names = [];
            for (var li = 1; li <= comp.numLayers; li++) {
                var L = comp.layer(li);
                var kind = (L instanceof TextLayer) ? "text" : "av";
                names.push(L.name + "(" + kind + ")");
                if (L instanceof TextLayer) found[L.name] = L;
            }
            log.push("  layers=[" + names.join(", ") + "]");

            for (var nm in args.expect) {
                var tl = found[nm];
                if (!tl) { fail("no TextLayer named " + nm); continue; }
                var txt = "";
                try { txt = tl.sourceText.value.text; }
                catch (e) { fail(nm + " sourceText read EXC " + e.toString()); continue; }
                log.push("  " + nm + ".text=" + txt + " (len " + txt.length + ")");
                if (txt !== args.expect[nm]) fail(nm + " text mismatch (want len " + args.expect[nm].length + ")");
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
