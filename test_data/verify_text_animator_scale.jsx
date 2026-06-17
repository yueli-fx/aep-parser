// Acceptance verify for from-scratch Text Scale 3D Animator (kinetic typography
// shrink-in). Reads text_animator_scale_args.json {input, done, resaved, png0,
// png1, png2}. Proves AE accepts the Go-built file, reads back the animator +
// Range Selector + Scale 3D value + the keyframed Range Offset (2 keys 0→100),
// resaves for the Go side, and renders three frames (t=0/1/2) so Go can assert
// on actual pixels that the glyph ink AREA changes as the per-character scale
// sweeps off the characters (red line 4 — render the visible surface).
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/text_animator_scale_args.json");
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
            if (it instanceof CompItem && it.name === "TXSCALEGATE") { comp = it; break; }
        }
        if (!comp) { fail("comp TXSCALEGATE not found"); }
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
                if (!animators) { fail("no Text Animators group"); }
                else {
                    note("animators numProps=" + animators.numProperties);
                    if (animators.numProperties < 1) fail("no animator present");
                    else {
                        var anim = animators.property(1);
                        note("animator name=" + anim.name + " match=" + anim.matchName);
                        var sels = anim.property("ADBE Text Selectors");
                        var sel = sels.property(1);
                        var off = sel.property("ADBE Text Percent Offset");
                        note("offset numKeys=" + off.numKeys);
                        if (off.numKeys !== 2) fail("offset numKeys=" + off.numKeys + ", want 2");
                        else {
                            var k1 = off.keyValue(1), k2 = off.keyValue(2);
                            note("offset key1=" + k1 + " key2=" + k2);
                            if (Math.abs(k1 - 0) > 0.5) fail("offset key1=" + k1 + ", want 0");
                            if (Math.abs(k2 - 100) > 0.5) fail("offset key2=" + k2 + ", want 100");
                        }
                        var props = anim.property("ADBE Text Animator Properties");
                        var sc = props.property("ADBE Text Scale 3D");
                        if (!sc) { fail("no Scale 3D property in animator"); }
                        else {
                            note("scale value=" + sc.value.toString());
                            // built oversize (>150%) so the gate sees an area change
                            if (sc.value[0] < 150) fail("scale x=" + sc.value[0] + ", want oversize");
                        }
                    }
                }
                // Place the text well inside the frame for pixel sampling.
                try {
                    var posProp = txt.property("ADBE Transform Group").property("ADBE Position");
                    note("pos before=" + posProp.value.toString());
                    posProp.setValue([400, 380]);
                    note("pos after=" + posProp.value.toString());
                } catch (e) { note("pos set err: " + e.toString()); }
                try {
                    var r = txt.sourceRectAtTime(2, false);
                    note("textRect@t2 top=" + r.top + " left=" + r.left + " w=" + r.width + " h=" + r.height);
                } catch (e) { note("rect err: " + e.toString()); }
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
