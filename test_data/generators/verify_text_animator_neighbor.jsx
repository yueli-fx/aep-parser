// Generic acceptance verify for the from-scratch "free-neighbor" Text Animator
// leaves (Fill/Stroke Opacity, Stroke Width/Color, Skew, Rotation X/Y). Reads
// text_animator_neighbor_args.json:
//   { input, done, resaved, png0, png1, png2, comp,
//     fontSize?, posX?, posY?,
//     applyFill?, applyStroke?, strokeWidth?, strokeWhite? }
// Proves AE accepts the Go-built file, reads back that the animator + Range
// Selector survived (numKeys 2 on the swept Range Offset, 0→100), applies the
// requested cosmetic fixture settings (font size / position / fill+stroke setup
// — sampling concerns, not the tested capability), resaves for the Go side, and
// renders three frames (t=0/1/2) for the Go pixel-signature assertion (red line
// 4). PASS/FAIL + notes go to the .done marker.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/text_animator_neighbor_args.json");
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
                if (!animators || animators.numProperties < 1) { fail("no Text Animator present"); }
                else {
                    var anim = animators.property(1);
                    note("animator=" + anim.name);
                    var sel = anim.property("ADBE Text Selectors").property(1);
                    var off = sel.property("ADBE Text Percent Offset");
                    note("offset numKeys=" + off.numKeys);
                    if (off.numKeys !== 2) fail("offset numKeys=" + off.numKeys + ", want 2");
                    else {
                        var k1 = off.keyValue(1), k2 = off.keyValue(2);
                        note("offset key1=" + k1 + " key2=" + k2);
                        if (Math.abs(k1 - 0) > 0.5) fail("offset key1=" + k1 + ", want 0");
                        if (Math.abs(k2 - 100) > 0.5) fail("offset key2=" + k2 + ", want 100");
                    }
                    var aprops = anim.property("ADBE Text Animator Properties");
                    note("animator props=" + aprops.numProperties);
                }

                // Cosmetic fixture setup (sampling concern, not the tested write).
                try {
                    var st = tp.property("ADBE Text Document");
                    var td = st.value;
                    if (args.fontSize) td.fontSize = args.fontSize;
                    if (args.applyFill === false) td.applyFill = false;
                    if (args.applyFill === true) td.applyFill = true;
                    if (args.applyStroke === true) {
                        td.applyStroke = true;
                        if (args.strokeWidth) td.strokeWidth = args.strokeWidth;
                        if (args.strokeWhite) td.strokeColor = [1, 1, 1];
                        td.strokeOverFill = true;
                    }
                    st.setValue(td);
                    note("doc set: fontSize=" + td.fontSize + " applyFill=" + td.applyFill +
                         " applyStroke=" + td.applyStroke + " strokeWidth=" + td.strokeWidth);
                } catch (e) { note("doc set err: " + e.toString()); }

                if (args.posX !== undefined && args.posY !== undefined) {
                    try {
                        txt.property("ADBE Transform Group").property("ADBE Position").setValue([args.posX, args.posY]);
                        note("pos set [" + args.posX + "," + args.posY + "]");
                    } catch (e) { note("pos err: " + e.toString()); }
                }
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
