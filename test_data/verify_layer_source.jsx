// Ship-gate verify for SetSource + ReplaceSource. Reads layer_source_args.json
// {input, done, resaved}. MAIN has L1 (ReplaceSource→B) and L2 (SetSource→B);
// both initially sourced comp A. AE should read layer.source.name == "B".
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/layer_source_args.json");
    argsFile.open("r"); var raw = argsFile.read(); argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL " + m); }

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
            function chk(name) {
                var L = byName[name];
                if (!L) { fail(name + " missing"); return; }
                var sn = (L.source == null) ? "null" : L.source.name;
                log.push("  " + name + ".source=" + sn);
                if (sn !== "B") fail(name + ".source=" + sn + ", want B");
            }
            chk("L1");
            chk("L2");
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
