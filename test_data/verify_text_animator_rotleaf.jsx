// Acceptance verify for animating a text animator LEAF itself (Rotation),
// distinct from sweeping the Range Offset. Reads
// text_animator_rotleaf_args.json {input, done, resaved, png0, png1, png2}.
// Proves AE accepts the Go-built file, reads back that the Rotation leaf carries
// 2 keyframes (0→90) while the Range Offset stays STATIC (numKeys 0) — i.e. the
// driven value is animated, not the selector — resaves for Go, and renders three
// frames so Go can assert on pixels that the single "L" glyph's ink bbox flips
// from tall to wide as the leaf rotation sweeps (red line 4). Single asymmetric
// "L" so the bbox aspect ratio is an unambiguous orientation signal.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/text_animator_rotleaf_args.json");
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
            if (it instanceof CompItem && it.name === "TXROTLEAF") { comp = it; break; }
        }
        if (!comp) { fail("comp TXROTLEAF not found"); }
        else {
            note("comp=" + comp.name + " layers=" + comp.numLayers);

            var txt = null;
            for (var li = 1; li <= comp.numLayers; li++) {
                if (comp.layer(li).name === "TXT") { txt = comp.layer(li); break; }
            }
            if (!txt) { fail("TXT layer missing"); }
            else {
                var tp = txt.property("ADBE Text Properties");
                var animators = tp.property("ADBE Text Animators");
                if (!animators || animators.numProperties < 1) { fail("no Text Animators"); }
                else {
                    var anim = animators.property(1);
                    var sel = anim.property("ADBE Text Selectors").property(1);
                    var off = sel.property("ADBE Text Percent Offset");
                    note("offset numKeys=" + off.numKeys);
                    if (off.numKeys !== 0) fail("offset numKeys=" + off.numKeys + ", want 0 (selector must stay static)");
                    var rot = anim.property("ADBE Text Animator Properties").property("ADBE Text Rotation");
                    if (!rot) { fail("no Rotation leaf"); }
                    else {
                        note("rotation numKeys=" + rot.numKeys);
                        if (rot.numKeys !== 2) fail("rotation numKeys=" + rot.numKeys + ", want 2");
                        else {
                            var k1 = rot.keyValue(1), k2 = rot.keyValue(2);
                            note("rotation key1=" + k1 + " key2=" + k2);
                            if (Math.abs(k1 - 0) > 0.5) fail("rotation key1=" + k1 + ", want 0");
                            if (Math.abs(k2 - 90) > 0.5) fail("rotation key2=" + k2 + ", want 90");
                        }
                    }
                }
                try {
                    var st = tp.property("ADBE Text Document");
                    var td = st.value;
                    td.fontSize = 240;
                    st.setValue(td);
                    note("fontSize set 240");
                } catch (e) { note("fontSize err: " + e.toString()); }
                try {
                    var posProp = txt.property("ADBE Transform Group").property("ADBE Position");
                    posProp.setValue([560, 460]);
                    note("pos after=" + posProp.value.toString());
                } catch (e) { note("pos set err: " + e.toString()); }
            }
        }

        app.project.save(new File(args.resaved));

        if (comp) {
            app.project.bitsPerChannel = 8;
            comp.saveFrameToPng(0, new File(args.png0));
            comp.saveFrameToPng(1, new File(args.png1));
            comp.saveFrameToPng(2, new File(args.png2));
            $.sleep(2000);
            var names = ["png0", "png1", "png2"];
            for (var n = 0; n < names.length; n++) {
                var pf = new File(args[names[n]]);
                if (pf.exists) { note(names[n] + " " + pf.length + " bytes"); }
                else { fail("saveFrameToPng wrote no file: " + names[n]); }
            }
        }
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
