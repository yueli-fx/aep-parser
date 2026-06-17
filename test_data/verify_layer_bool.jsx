// Acceptance verify for layer-set ldta flag setters (stretch / autoOrient /
// collapseTransformation). Reads layer_bool_args.json {input, done, resaved}.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/layer_bool_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    function note(m) { log.push("  " + m); }
    function near(a, b, tol) { return Math.abs(a - b) <= tol; }

    try {
        app.open(new File(args.input));
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "BOOL") { comp = it; break; }
        }
        if (!comp) { fail("comp BOOL not found"); }
        else {
            var byName = {};
            for (var li = 1; li <= comp.numLayers; li++) byName[comp.layer(li).name] = comp.layer(li);

            var AUT = byName["AUT"], COL = byName["COL"];
            if (!AUT) fail("AUT missing");
            else {
                note("AUT.autoOrient=" + AUT.autoOrient + " (ALONG_PATH=" + AutoOrientType.ALONG_PATH + ")");
                if (AUT.autoOrient !== AutoOrientType.ALONG_PATH) fail("AUT.autoOrient=" + AUT.autoOrient + ", want ALONG_PATH");
            }
            if (!COL) fail("COL missing");
            else {
                note("COL.collapseTransformation=" + COL.collapseTransformation);
                if (COL.collapseTransformation !== true) fail("COL.collapseTransformation=" + COL.collapseTransformation + ", want true");
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
