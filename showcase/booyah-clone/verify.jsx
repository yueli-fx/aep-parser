// showcase/booyah-clone/verify.jsx — open booyah-clone.aep and dump every
// comp's DOM to verify.done (plain text), confirming AE actually ingested the
// shapes/effects (spec §3 结构保真; catches silent-drop). Iterates ALL comps so no
// Japanese-name matching is needed. String-only output (no JSON.stringify, which can
// throw on DOM enum values). Run via scripts/ae-worker/ae_run.ps1, -Done = verify.done.
(function () {
    var dir = (new File($.fileName).parent.fsName + "/");
    var L = [];
    function w(s) { L.push(String(s)); }

    function dumpGroup(g, depth) {
        var pad = ""; for (var k = 0; k < depth; k++) pad += "  ";
        w(pad + "[G] " + g.matchName + " (" + g.numProperties + ")");
        if (depth >= 7) return;
        for (var i = 1; i <= g.numProperties; i++) {
            var c = g.property(i);
            if (c.propertyType === PropertyType.PROPERTY) {
                w(pad + "  [P] " + c.matchName);
            } else {
                dumpGroup(c, depth + 1);
            }
        }
    }

    // dumpEffects walks a layer's Effect Parade: each effect's matchName plus, per
    // leaf param, its static value and (if expression-able) the expression + enabled
    // flag. Confirms AE ingested effect params/expressions (comp ③ Fractal Noise etc.).
    function dumpEffects(layer) {
        var parade;
        try { parade = layer.property("ADBE Effect Parade"); } catch (e) { return; }
        if (!parade || parade.numProperties < 1) return;
        for (var e = 1; e <= parade.numProperties; e++) {
            var fx = parade.property(e);
            w("    [FX] " + fx.matchName + " (" + fx.numProperties + ")");
            for (var p = 1; p <= fx.numProperties; p++) {
                var pr;
                try { pr = fx.property(p); } catch (e0) { continue; }
                if (pr.propertyType !== PropertyType.PROPERTY) continue;
                var line = "      [FP] " + pr.matchName;
                try {
                    if (pr.canSetExpression && pr.expressionEnabled) {
                        line += " expr=\"" + pr.expression + "\" on";
                    } else if (pr.numKeys > 0) {
                        line += " keys=" + pr.numKeys + " v0=" + String(pr.keyValue(1));
                    } else {
                        line += " v=" + String(pr.value);
                    }
                } catch (ex) { line += " (val? " + ex.toString() + ")"; }
                w(line);
            }
        }
    }

    // dumpTransformExpr reports whether a layer's Transform Position/Opacity carry an
    // ENABLED expression (the wiggle RGB jitter on ⑧ / flicker on ⑤ L2). expressionEnabled
    // is the real AE-acceptance signal for SetExpression (Go round-trip can be a false
    // green per red-line-1 / incident expression-enable-byte-pair).
    function dumpTransformExpr(layer) {
        var tg;
        try { tg = layer.property("ADBE Transform Group"); } catch (e) { return; }
        if (!tg) return;
        var names = ["ADBE Position", "ADBE Opacity"];
        for (var n = 0; n < names.length; n++) {
            var pr;
            try { pr = tg.property(names[n]); } catch (e1) { continue; }
            if (!pr) continue;
            try {
                if (pr.canSetExpression && pr.expressionEnabled) {
                    var samp = "";
                    var ts = [0.4, 1.0, 1.7, 2.6];
                    for (var t = 0; t < ts.length; t++) {
                        samp += " @" + ts[t] + "=" + String(pr.valueAtTime(ts[t], false));
                    }
                    w("    [TX] " + names[n] + " expr=\"" + pr.expression + "\" ON" + samp);
                }
            } catch (ex) { w("    [TX] " + names[n] + " (err " + ex.toString() + ")"); }
        }
    }

    try {
        app.open(new File(dir + "booyah-clone.aep"));
        w("items=" + app.project.numItems);
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (!(it instanceof CompItem)) continue;
            w("COMP \"" + it.name + "\" " + it.width + "x" + it.height + " " + it.duration + "s layers=" + it.numLayers);
            for (var li = 1; li <= it.numLayers; li++) {
                var l = it.layer(li);
                var srcName = "";
                try { if (l.source && l.source.name) srcName = " src=\"" + l.source.name + "\""; } catch (es) {}
                var nMasks = 0;
                try { var mp = l.property("ADBE Mask Parade"); if (mp) nMasks = mp.numProperties; } catch (em) {}
                w("  L" + l.index + " \"" + l.name + "\" blend=" + l.blendingMode + " masks=" + nMasks + srcName);
                try {
                    var contents = l.property("ADBE Root Vectors Group");
                    if (contents) dumpGroup(contents, 2);
                } catch (e) { w("    (no contents: " + e + ")"); }
                try { dumpEffects(l); } catch (e) { w("    (no fx: " + e + ")"); }
                try { dumpTransformExpr(l); } catch (e) { w("    (no tx: " + e + ")"); }
            }
        }
        w("OK");
    } catch (e) { w("EXC " + e.toString() + " line=" + (e.line || "?")); }

    var f = new File(dir + "verify.done");
    f.open("w"); f.write(L.join("\n")); f.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
