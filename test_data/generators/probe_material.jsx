// One-off probe: dump every layer's materialOption children (matchName + value)
// in re_material_options.aep. Per-property try/catch + safe stringify so one
// bad .value (the ExtendScript 数字结果无效 concat throw) doesn't abort the dump.
(function () {
    var out = [];
    function P(s) { out.push(s); }
    function safeVal(pr) {
        try {
            var v = pr.value;
            if (v === undefined || v === null) return "" + v;
            if (v instanceof Array) {
                var parts = [];
                for (var k = 0; k < v.length; k++) { parts.push("" + v[k]); }
                return "[" + parts.join(",") + "]";
            }
            return "" + v;
        } catch (e) { return "ERR:" + e; }
    }
    try {
        app.open(new File("E:/projects/tools/aep-parser/test_data/fixtures/re_material_options.aep"));
        P("AE version: " + app.version);
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "MAT_OPT") { comp = it; break; }
        }
        if (!comp) { P("NO MAT_OPT comp"); }
        else {
            for (var li = 1; li <= comp.numLayers; li++) {
                var L = comp.layer(li);
                P("LAYER[" + li + "] name=" + L.name + " threeD=" + L.threeDLayer);
                var mo = null;
                try { mo = L.materialOption; } catch (e) { P("  no materialOption: " + e); }
                if (mo) {
                    for (var pi = 1; pi <= mo.numProperties; pi++) {
                        try {
                            var pr = mo.property(pi);
                            P("  [" + pi + "] mn=" + pr.matchName + " | val=" + safeVal(pr));
                        } catch (e3) {
                            P("  [" + pi + "] PROP-ERR:" + e3);
                        }
                    }
                }
            }
        }
    } catch (e) {
        P("EXC " + e.toString());
    }
    var done = new File("E:/projects/tools/aep-parser/test_data/probe_material.done");
    done.open("w"); done.write(out.join("\n")); done.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
