// Ship-gate verify for AddFont + SetRunFontIndex. Reads text_font_args.json
// {input, done, resaved}. Reads textDocument.font on the switched layer "SW"
// (expect "ArialMT") and logs the default layer "DEF" font for contrast.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/text_font_args.json");
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
            if (it instanceof CompItem && it.name === "TXTF") { comp = it; break; }
        }
        if (!comp) { fail("comp TXTF not found"); }
        else {
            var byName = {};
            for (var li = 1; li <= comp.numLayers; li++) byName[comp.layer(li).name] = comp.layer(li);

            var DEF = byName["DEF"], SW = byName["SW"];
            if (DEF && DEF.property("Source Text")) {
                log.push("  DEF.font=" + DEF.property("Source Text").value.font + " (default, context)");
            }
            if (!SW || !SW.property("Source Text")) { fail("SW text layer missing"); }
            else {
                var f = SW.property("Source Text").value.font;
                log.push("  SW.font=" + f);
                if (f !== "ArialMT") fail("SW.font=" + f + ", want ArialMT");
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
