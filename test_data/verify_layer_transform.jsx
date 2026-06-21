// Ship-gate verify for SetLayerTransform. Reads layer_transform_args.json
// {input, done, resaved}. Opens the Go-written file, confirms AE accepts it and
// that the text layer's Transform Group carries the MATERIALIZED + ANIMATED
// channels SetLayerTransform wrote onto a templated (default-elided) layer:
// Anchor static, Position 3 kf, Opacity 3 kf — read back via DOM valueAtTime.
// Resaves so the Go side can confirm preservation.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/layer_transform_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  " + m); }
    function near(a, b, tol) { return Math.abs(a - b) <= tol; }
    function vecNear(v, ex, tol) {
        for (var i = 0; i < ex.length; i++) { if (!near(v[i], ex[i], tol)) return false; }
        return true;
    }

    try {
        app.open(new File(args.input));
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem) { comp = it; break; }
        }
        if (!comp) { fail("no CompItem"); }
        else {
            var L = comp.layer(1);
            if (!(L instanceof TextLayer)) fail("layer 1 not a TextLayer");
            var tg = L.property("ADBE Transform Group");
            var anc = tg.property("ADBE Anchor Point");
            var pos = tg.property("ADBE Position");
            var op = tg.property("ADBE Opacity");

            log.push("  anchor=" + anc.value.toString());
            if (!vecNear(anc.value, [9, -436], 0.5)) fail("anchor != [9,-436]");

            log.push("  pos numKeys=" + pos.numKeys);
            if (pos.numKeys !== 3) fail("pos numKeys " + pos.numKeys + " != 3");
            else {
                if (!vecNear(pos.valueAtTime(0, false), [600, 540], 0.5)) fail("pos@0 != [600,540]");
                if (!vecNear(pos.valueAtTime(1, false), [1320, 540], 0.5)) fail("pos@1 != [1320,540]");
                if (!vecNear(pos.valueAtTime(2, false), [960, 300], 0.5)) fail("pos@2 != [960,300]");
            }

            log.push("  op numKeys=" + op.numKeys + " @0=" + op.valueAtTime(0, false) + " @1=" + op.valueAtTime(1, false) + " @2=" + op.valueAtTime(2, false));
            if (op.numKeys !== 3) fail("op numKeys " + op.numKeys + " != 3");
            else {
                if (!near(op.valueAtTime(0, false), 10, 0.5)) fail("op@0 != 10");
                if (!near(op.valueAtTime(1, false), 100, 0.5)) fail("op@1 != 100");
                if (!near(op.valueAtTime(2, false), 40, 0.5)) fail("op@2 != 40");
            }
        }
        app.project.save(new File(args.resaved));
    } catch (e) {
        fail("EXC " + e.toString() + " line=" + e.line);
    }

    var done = new File(args.done);
    done.encoding = "UTF-8";
    done.open("w");
    done.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n"));
    done.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
