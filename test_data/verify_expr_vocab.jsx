// Acceptance verify for MG roadmap S2 follow-up: expression VOCABULARY
// (loopOut / wiggle / cross-layer reference). Reads expr_vocab_args.json
// {input, done, resaved, png}. Proves AE accepts the file and EVALUATES each
// idiom (valueAtTime), resaves, and renders t=2.5s for the Go side's pixel
// assertions (red line 4).
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/expr_vocab_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var T = 2.5;
    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    function note(m) { log.push("  " + m); }
    function near(a, b, tol) { return Math.abs(a - b) <= tol; }
    function v2s(v) { return "[" + v[0] + "," + v[1] + "]"; } // index elements — never coerce an Array directly (ExtendScript "÷0" throw)

    try {
        app.open(new File(args.input));

        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "EXPRVOCAB") { comp = it; break; }
        }
        if (!comp) { fail("comp EXPRVOCAB not found"); }
        else {
            note("comp=" + comp.name + " layers=" + comp.numLayers);
            var byName = {};
            for (var li = 1; li <= comp.numLayers; li++) byName[comp.layer(li).name] = comp.layer(li);

            // LEAD — static link target.
            var lead = byName["LEAD"];
            if (!lead) { fail("LEAD missing"); }
            else {
                var lp = lead.transform.position.valueAtTime(T, false);
                note("LEAD pos@" + T + "=" + v2s(lp));
                if (!near(lp[0], 480, 1) || !near(lp[1], 250, 1)) fail("LEAD pos=" + v2s(lp) + ", want [480,250]");
            }

            // LINK — cross-layer reference + vector arithmetic.
            var link = byName["LINK"];
            if (!link) { fail("LINK missing"); }
            else {
                var lkp = link.transform.position;
                note("LINK expr='" + lkp.expression + "' enabled=" + lkp.expressionEnabled);
                if (!lkp.expressionEnabled) fail("LINK expressionEnabled=false");
                var v = lkp.valueAtTime(T, false);
                note("LINK pos@" + T + "=" + v2s(v));
                if (!near(v[0], 480, 1) || !near(v[1], 500, 1)) fail("LINK pos=" + v2s(v) + ", want [480,500] (LEAD.pos+[0,250]) — cross-layer ref not evaluating");
            }

            // LOOP — expression on a keyframed property.
            var loop = byName["LOOP"];
            if (!loop) { fail("LOOP missing"); }
            else {
                var lop = loop.transform.position;
                note("LOOP expr='" + lop.expression + "' enabled=" + lop.expressionEnabled + " numKeys=" + lop.numKeys);
                if (!lop.expressionEnabled) fail("LOOP expressionEnabled=false");
                if (lop.numKeys !== 2) fail("LOOP numKeys=" + lop.numKeys + ", want 2");
                var lv = lop.valueAtTime(T, false);
                note("LOOP pos@" + T + "=" + v2s(lv));
                // cycle 1s, phase 0.5 → linear midpoint x=900; without loop AE holds x=1500.
                if (!near(lv[0], 900, 30)) fail("LOOP x=" + lv[0] + ", want ≈900 (loopOut cycle); ≈1500 = loop not evaluating");
            }

            // WIG — procedural / time-varying.
            var wig = byName["WIG"];
            if (!wig) { fail("WIG missing"); }
            else {
                var wp = wig.transform.position;
                note("WIG expr='" + wp.expression + "' enabled=" + wp.expressionEnabled);
                if (!wp.expressionEnabled) fail("WIG expressionEnabled=false");
                var w1 = wp.valueAtTime(T, false);
                var w0 = wp.valueAtTime(1.0, false);
                note("WIG pos@" + T + "=" + v2s(w1) + "  pos@1.0=" + v2s(w0));
                if (near(w1[0], 960, 0.5) && near(w1[1], 950, 0.5)) fail("WIG pos=" + v2s(w1) + " unchanged from anchor [960,950] — wiggle not evaluating");
                if (near(w1[0], w0[0], 0.5) && near(w1[1], w0[1], 0.5)) fail("WIG pos identical at t=1.0 and t=" + T + " — wiggle not time-varying");
            }

            // SLD — expression resolving an effect-parameter reference. Read the
            // effect via the parade group (the .effect() DOM accessor is flaky
            // here); the expression itself uses effect(1) in AE's engine.
            var sld = byName["SLD"];
            if (!sld) { fail("SLD missing"); }
            else {
                var parade = sld.property("ADBE Effect Parade");
                if (!parade || parade.numProperties < 1) { fail("SLD Slider Control dropped (parade empty)"); }
                else {
                    note("SLD effect[1] matchName='" + parade.property(1).matchName + "'");
                    var sp = sld.transform.position;
                    note("SLD expr='" + sp.expression + "' enabled=" + sp.expressionEnabled);
                    if (!sp.expressionEnabled) fail("SLD expressionEnabled=false");
                    // The expression `effect(1)("Slider")` evaluating to x=880 IS
                    // the proof the slider value (880) is reachable by reference.
                    var sv = sp.valueAtTime(T, false);
                    note("SLD pos@" + T + "=" + v2s(sv));
                    if (!near(sv[0], 880, 1)) fail("SLD x=" + sv[0] + ", want 880 (effect-param ref); ≈960 = ref not evaluating");
                }
            }
        }

        app.project.save(new File(args.resaved));

        if (comp && args.png) {
            app.project.bitsPerChannel = 8;
            comp.saveFrameToPng(T, new File(args.png));
            $.sleep(2000);
            var pngFile = new File(args.png);
            if (pngFile.exists) { note("frame png " + pngFile.length + " bytes"); }
            else { fail("saveFrameToPng wrote no file"); }
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
