// Acceptance verify for layer-set transform statics + frame-time setters. Reads
// layer_xform_args.json {input, done, resaved}. Reads the solid layer's DOM
// transform values + in/out/start times, then resaves for the Go side.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/layer_xform_args.json");
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
            if (it instanceof CompItem && it.name === "XFORM") { comp = it; break; }
        }
        if (!comp) { fail("comp XFORM not found"); }
        else {
            var L = comp.layer("L");
            if (!L) { fail("layer L not found"); }
            else {
                var pos = L.transform.position.value;
                var anc = L.transform.anchorPoint.value;
                var scl = L.transform.scale.value;
                var op = L.transform.opacity.value;
                note("position=[" + pos[0] + "," + pos[1] + "] anchor=[" + anc[0] + "," + anc[1] + "] scale=[" + scl[0] + "," + scl[1] + "] opacity=" + op);
                note("startTime=" + L.startTime + " inPoint=" + L.inPoint + " outPoint=" + L.outPoint);

                if (!near(pos[0], 400, 1) || !near(pos[1], 300, 1)) fail("position=[" + pos[0] + "," + pos[1] + "], want [400,300]");
                if (!near(anc[0], 25, 1) || !near(anc[1], 25, 1)) fail("anchorPoint=[" + anc[0] + "," + anc[1] + "], want [25,25]");
                if (!near(scl[0], 125, 0.5) || !near(scl[1], 125, 0.5)) fail("scale=[" + scl[0] + "," + scl[1] + "], want [125,125]");
                if (!near(op, 60, 1)) fail("opacity=" + op + ", want 60");
                // AE inPoint/outPoint are comp-absolute = startTime + the
                // source-relative trim our SetInPoint/SetOutPoint writes. With
                // startTime 0.5: in = 0.5 + 1.0 = 1.5, out = 0.5 + 5.0 = 5.5.
                if (!near(L.startTime, 0.5, 0.02)) fail("startTime=" + L.startTime + ", want 0.5");
                if (!near(L.inPoint, 1.5, 0.02)) fail("inPoint=" + L.inPoint + ", want 1.5 (startTime 0.5 + in 1.0)");
                if (!near(L.outPoint, 5.5, 0.02)) fail("outPoint=" + L.outPoint + ", want 5.5 (startTime 0.5 + out 5.0)");
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
