// flightdeck/showcase/booyah-clone/verify.jsx — open booyah-clone.aep and dump every
// comp's DOM to verify.done (plain text), confirming AE actually ingested the
// shapes/effects (spec §3 结构保真; catches silent-drop). Iterates ALL comps so no
// Japanese-name matching is needed. String-only output (no JSON.stringify, which can
// throw on DOM enum values). Run via scripts/ae_run.ps1, -Done = verify.done.
(function () {
    var dir = "e:/projects/tools/aep-parser/flightdeck/showcase/booyah-clone/";
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
                w("  L" + l.index + " \"" + l.name + "\" blend=" + l.blendingMode + srcName);
                try {
                    var contents = l.property("ADBE Root Vectors Group");
                    if (contents) dumpGroup(contents, 2);
                } catch (e) { w("    (no contents: " + e + ")"); }
                try { dumpEffects(l); } catch (e) { w("    (no fx: " + e + ")"); }
            }
        }
        w("OK");
    } catch (e) { w("EXC " + e.toString() + " line=" + (e.line || "?")); }

    var f = new File(dir + "verify.done");
    f.open("w"); f.write(L.join("\n")); f.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
