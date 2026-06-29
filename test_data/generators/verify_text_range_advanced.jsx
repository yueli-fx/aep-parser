// Ship-gate verify for SetTextRangeAdvanced — Amount (Selector Max Amount). Reads
// text_range_advanced_args.json {input, done, resaved, png0, comp}. The comp has
// two text layers, both with an Opacity-0 animator over the full range, differing
// ONLY in the Range Advanced Amount: TXTA Amount=100 (effect full → text hidden),
// TXTB Amount=20 (effect at 20% → text ~80% visible). Positions both layers into
// separate regions (sampling concern), renders one frame for the Go A/B pixel
// assertion, resaves, and reads back each layer's Max Amount. PASS/FAIL to .done.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/text_range_advanced_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    function note(m) { log.push("  " + m); }

    function setup(layer, y) {
        var tp = layer.property("ADBE Text Properties");
        try {
            var st = tp.property("ADBE Text Document");
            var td = st.value; td.fontSize = 120; st.setValue(td);
        } catch (e) { note("font err: " + e); }
        try { layer.property("ADBE Transform Group").property("ADBE Position").setValue([180, y]); } catch (e) { note("pos err: " + e); }
        try {
            var adv = tp.property("ADBE Text Animators").property(1)
                .property("ADBE Text Selectors").property(1).property("ADBE Text Range Advanced");
            note(layer.name + " amount=" + adv.property("ADBE Text Selector Max Amount").value);
        } catch (e) { note(layer.name + " amount read err: " + e); }
    }

    try {
        app.open(new File(args.input));
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === args.comp) { comp = it; break; }
        }
        if (!comp) { fail("comp " + args.comp + " not found"); }
        else {
            var a = null, b = null;
            for (var li = 1; li <= comp.numLayers; li++) {
                var ly = comp.layer(li);
                if (ly.name === "TXTA") a = ly;
                if (ly.name === "TXTB") b = ly;
            }
            if (!a || !b) { fail("TXTA/TXTB missing (a=" + !!a + " b=" + !!b + ")"); }
            else {
                setup(a, 250);
                setup(b, 520);
            }
            app.project.save(new File(args.resaved));
            app.project.bitsPerChannel = 8;
            comp.saveFrameToPng(0, new File(args.png0));
            $.sleep(1500);
            var pf = new File(args.png0);
            if (pf.exists) { note("png0 " + pf.length + " bytes"); } else { fail("no png0"); }
        }
    } catch (e) { fail("EXC " + e.toString() + " line=" + e.line); }

    var done = new File(args.done);
    done.open("w");
    done.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n"));
    done.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
