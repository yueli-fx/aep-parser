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

    try {
        app.open(new File(dir + "booyah-clone.aep"));
        w("items=" + app.project.numItems);
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (!(it instanceof CompItem)) continue;
            w("COMP \"" + it.name + "\" " + it.width + "x" + it.height + " " + it.duration + "s layers=" + it.numLayers);
            for (var li = 1; li <= it.numLayers; li++) {
                var l = it.layer(li);
                w("  L" + l.index + " \"" + l.name + "\" blend=" + l.blendingMode);
                try {
                    var contents = l.property("ADBE Root Vectors Group");
                    if (contents) dumpGroup(contents, 2);
                } catch (e) { w("    (no contents: " + e + ")"); }
            }
        }
        w("OK");
    } catch (e) { w("EXC " + e.toString() + " line=" + (e.line || "?")); }

    var f = new File(dir + "verify.done");
    f.open("w"); f.write(L.join("\n")); f.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
