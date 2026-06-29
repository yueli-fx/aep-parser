// Generic 3-frame render verify (no animator-keyframe assertion) — for selector
// types whose motion is intrinsic (Wiggly Selector) rather than offset-keyframed.
// Reads text_render3_args.json {input, done, resaved, png0, png1, png2, comp,
// fontSize?, posX?, posY?}. Opens the Go-built file, confirms a TXT layer with a
// text animator exists, applies cosmetic font/position (sampling concern),
// renders three frames for the Go pixel assertion, resaves, writes PASS/FAIL.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/text_render3_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    function note(m) { log.push("  " + m); }

    try {
        app.open(new File(args.input));
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === args.comp) { comp = it; break; }
        }
        if (!comp) { fail("comp " + args.comp + " not found"); }
        else {
            var txt = null;
            for (var li = 1; li <= comp.numLayers; li++) {
                if (comp.layer(li).name === "TXT") { txt = comp.layer(li); break; }
            }
            if (!txt) { fail("TXT layer missing"); }
            else {
                var tp = txt.property("ADBE Text Properties");
                var animators = tp.property("ADBE Text Animators");
                if (!animators || animators.numProperties < 1) fail("no text animator present");
                else {
                    var sels = animators.property(1).property("ADBE Text Selectors");
                    note("animator selectors=" + sels.numProperties);
                }
                try {
                    var st = tp.property("ADBE Text Document");
                    var td = st.value; if (args.fontSize) td.fontSize = args.fontSize; st.setValue(td);
                } catch (e) { note("font err: " + e); }
                if (args.posX !== undefined) {
                    try { txt.property("ADBE Transform Group").property("ADBE Position").setValue([args.posX, args.posY]); } catch (e) { note("pos err: " + e); }
                }
            }
            app.project.save(new File(args.resaved));
            app.project.bitsPerChannel = 8;
            comp.saveFrameToPng(0, new File(args.png0));
            comp.saveFrameToPng(1, new File(args.png1));
            comp.saveFrameToPng(2, new File(args.png2));
            $.sleep(2000);
            var names = ["png0", "png1", "png2"];
            for (var n = 0; n < names.length; n++) {
                var pf = new File(args[names[n]]);
                if (!pf.exists) fail("no " + names[n]);
            }
        }
    } catch (e) { fail("EXC " + e.toString() + " line=" + e.line); }

    var done = new File(args.done);
    done.open("w");
    done.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n"));
    done.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
