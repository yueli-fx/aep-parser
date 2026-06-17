// Generic acceptance verify for animating a 3D/4D text animator LEAF itself
// (Position 3D or Fill Color), vs sweeping the Range Offset. Reads
// text_animator_vecleaf_args.json {input, done, resaved, png0, png1, png2, comp,
// leaf, basewhite}. Proves AE accepts the Go-built file, reads back that the
// named leaf carries 2 keyframes while the Range Offset stays STATIC (numKeys 0)
// — i.e. the driven value animates, not the selector — resaves for Go, and
// renders three frames so Go can assert on pixels. basewhite forces the base text
// colour white (used by the Fill Color gate so the override is unambiguous).
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/text_animator_vecleaf_args.json");
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
                    var leaf = anim.property("ADBE Text Animator Properties").property(args.leaf);
                    if (!leaf) { fail("no leaf " + args.leaf); }
                    else {
                        note("leaf " + args.leaf + " numKeys=" + leaf.numKeys);
                        if (leaf.numKeys !== 2) fail("leaf numKeys=" + leaf.numKeys + ", want 2");
                    }
                }
                if (args.basewhite) {
                    try {
                        var st = tp.property("ADBE Text Document");
                        var td = st.value;
                        td.fontSize = 150;
                        td.fillColor = [1, 1, 1];
                        td.applyFill = true;
                        st.setValue(td);
                        note("base white + fontSize 150 set");
                    } catch (e) { note("doc set err: " + e.toString()); }
                } else {
                    try {
                        var st2 = tp.property("ADBE Text Document");
                        var td2 = st2.value;
                        td2.fontSize = 150;
                        st2.setValue(td2);
                        note("fontSize 150 set");
                    } catch (e) { note("doc set err: " + e.toString()); }
                }
                try {
                    var posProp = txt.property("ADBE Transform Group").property("ADBE Position");
                    posProp.setValue([180, 420]);
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
