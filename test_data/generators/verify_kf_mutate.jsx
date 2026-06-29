// Acceptance verify for the keyframe MUTATE setters. Reads kf_mutate_args.json
// {input, done, resaved}. Opens the Go-built file, reads back each layer's
// mutated keyframe data via the AE DOM (interp type / temporal ease / time /
// value / spatial tangent / static value / numKeys), then resaves for the Go
// side. No render — these setters' surface is the keyframe data; mg_ease already
// render-proves eased motion.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/kf_mutate_args.json");
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
            if (it instanceof CompItem && it.name === "KFMUT") { comp = it; break; }
        }
        if (!comp) { fail("comp KFMUT not found"); }
        else {
            note("comp=" + comp.name + " layers=" + comp.numLayers);
            var byName = {};
            for (var li = 1; li <= comp.numLayers; li++) byName[comp.layer(li).name] = comp.layer(li);

            // EASE: interpolation + temporal ease.
            var ease = byName["EASE"];
            if (!ease) { fail("EASE missing"); }
            else {
                var ep = ease.transform.position;
                note("EASE numKeys=" + ep.numKeys);
                if (ep.numKeys !== 2) fail("EASE numKeys=" + ep.numKeys + ", want 2");
                else {
                    var oInterp = ep.keyOutInterpolationType(1);
                    var iInterp = ep.keyInInterpolationType(2);
                    note("EASE kf1 outInterp=" + oInterp + " kf2 inInterp=" + iInterp);
                    if (oInterp !== KeyframeInterpolationType.BEZIER) fail("EASE kf1 outInterp=" + oInterp + ", want BEZIER");
                    if (iInterp !== KeyframeInterpolationType.BEZIER) fail("EASE kf2 inInterp=" + iInterp + ", want BEZIER");
                    var oEase = ep.keyOutTemporalEase(1)[0];
                    var iEase = ep.keyInTemporalEase(2)[0];
                    note("EASE kf1 outEase inf=" + oEase.influence + " kf2 inEase inf=" + iEase.influence);
                    if (!near(oEase.influence, 85, 2)) fail("EASE kf1 out influence=" + oEase.influence + ", want ~85");
                    if (!near(iEase.influence, 60, 2)) fail("EASE kf2 in influence=" + iEase.influence + ", want ~60");
                }
            }

            // VALT: time + value + frame-time.
            var valt = byName["VALT"];
            if (!valt) { fail("VALT missing"); }
            else {
                var vp = valt.transform.position;
                note("VALT numKeys=" + vp.numKeys);
                if (vp.numKeys !== 3) fail("VALT numKeys=" + vp.numKeys + ", want 3");
                else {
                    var t2 = vp.keyTime(2);
                    var t3 = vp.keyTime(3);
                    var v2 = vp.keyValue(2);
                    note("VALT kf2 time=" + t2 + " value=[" + v2[0] + "," + v2[1] + "] kf3 time=" + t3);
                    if (!near(t2, 1.6, 0.03)) fail("VALT kf2 time=" + t2 + ", want ~1.6");
                    if (!near(t3, 3.0, 0.03)) fail("VALT kf3 time=" + t3 + ", want ~3.0 (SetFrameTime 90)");
                    if (!near(v2[0], 850, 1) || !near(v2[1], 560, 1)) fail("VALT kf2 value=[" + v2[0] + "," + v2[1] + "], want [850,560]");
                }
            }

            // TAN: spatial tangents.
            var tan = byName["TAN"];
            if (!tan) { fail("TAN missing"); }
            else {
                var tp = tan.transform.position;
                note("TAN numKeys=" + tp.numKeys);
                if (tp.numKeys !== 2) fail("TAN numKeys=" + tp.numKeys + ", want 2");
                else {
                    var ot = tp.keyOutSpatialTangent(1);
                    var it = tp.keyInSpatialTangent(2);
                    note("TAN kf1 outTangent=[" + ot[0] + "," + ot[1] + "] kf2 inTangent=[" + it[0] + "," + it[1] + "]");
                    if (!near(ot[0], 300, 2) || !near(ot[1], -250, 2)) fail("TAN kf1 outTangent=[" + ot[0] + "," + ot[1] + "], want [300,-250]");
                    if (!near(it[0], -300, 2) || !near(it[1], -250, 2)) fail("TAN kf2 inTangent=[" + it[0] + "," + it[1] + "], want [-300,-250]");
                }
            }

            // STAT: static value mutate (no keyframes).
            var stat = byName["STAT"];
            if (!stat) { fail("STAT missing"); }
            else {
                var spv = stat.transform.position;
                note("STAT numKeys=" + spv.numKeys + " value=[" + spv.value[0] + "," + spv.value[1] + "]");
                if (spv.numKeys !== 0) fail("STAT should be static, numKeys=" + spv.numKeys);
                if (!near(spv.value[0], 420, 1) || !near(spv.value[1], 950, 1)) fail("STAT value=[" + spv.value[0] + "," + spv.value[1] + "], want [420,950]");
            }

            // INS / DEL: keyframe count after structural mutate.
            var ins = byName["INS"];
            if (!ins) { fail("INS missing"); }
            else {
                note("INS numKeys=" + ins.transform.position.numKeys);
                if (ins.transform.position.numKeys !== 3) fail("INS numKeys=" + ins.transform.position.numKeys + ", want 3");
            }
            var del = byName["DEL"];
            if (!del) { fail("DEL missing"); }
            else {
                note("DEL numKeys=" + del.transform.position.numKeys);
                if (del.transform.position.numKeys !== 2) fail("DEL numKeys=" + del.transform.position.numKeys + ", want 2");
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
