// Ship-gate verify for batch-7b paragraph SetParagraph* setters. Reads
// text_para_args.json {input, done, resaved}. Single-paragraph text surfaces the
// paragraph attributes as whole-document textDocument properties (AE2022+ for the
// indent/spacing DOM; AE2020 falls back to Go resave-preservation). Discovery run
// also exposes the FormatPSReal bug (a bare-int point value reads as value/65536).
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/text_para_args.json");
    argsFile.open("r"); var raw = argsFile.read(); argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL " + m); }
    function near(label, getter, want, tol) {
        if (tol === undefined) tol = 0.1;
        var got;
        try { got = getter(); } catch (e) { log.push("  " + label + " N/A: " + e); return; }
        if (got === undefined || got === null) { log.push("  " + label + " undefined (DOM N/A this ver)"); return; }
        if (Math.abs(got - want) > tol) fail(label + "=" + got + " want " + want);
        else log.push("  " + label + "=" + got);
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
                near("firstLineIndent", function () { return td.firstLineIndent; }, 20);
                near("leftMargin", function () { return td.leftMargin; }, 15);
                near("rightMargin", function () { return td.rightMargin; }, 10);
                near("spaceBefore", function () { return td.spaceBefore; }, 25);
                near("spaceAfter", function () { return td.spaceAfter; }, 30);
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
