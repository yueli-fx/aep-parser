// verify_timing.jsx — dump comp① layer-0 timing + rect Size valueAtTime from AE's
// own DOM (ground truth) for both original and clone. Sidecar verify_timing.txt:
//   line1: absolute .aep path
//   line2: comp name
// Writes verify_timing.done (UTF-8) with the DOM dump.
(function () {
    var dir = "e:/projects/tools/aep-parser/flightdeck/showcase/booyah-clone/";
    var out = [];
    try {
        var sf = new File(dir + "verify_timing.txt");
        sf.encoding = "UTF-8"; sf.open("r");
        var aepPath = (sf.readln() || "").replace(/[\r\n]+$/, "");
        var compName = (sf.readln() || "").replace(/[\r\n]+$/, "");
        sf.close();

        app.open(new File(aepPath));
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === compName) { comp = it; break; }
        }
        if (!comp) { out.push("FAIL no comp " + compName); }
        else {
            var L = comp.layer(1);
            out.push("layer=" + L.name + " startTime=" + L.startTime + " inPoint=" + L.inPoint + " outPoint=" + L.outPoint + " stretch=" + L.stretch);
            // descend root vectors group -> collect rects
            var contents = L.property("ADBE Root Vectors Group");
            function walkRects(grp) {
                for (var k = 1; k <= grp.numProperties; k++) {
                    var p = grp.property(k);
                    if (p.matchName === "ADBE Vector Shape - Rect") {
                        var sz = p.property("ADBE Vector Rect Size");
                        var kt = [];
                        for (var n = 1; n <= sz.numKeys; n++) {
                            kt.push(sz.keyTime(n).toFixed(4) + ":" + sz.keyValue(n).toString());
                        }
                        out.push("  RECT Size " + sz.numKeys + "kf [" + kt.join(" | ") + "]");
                    } else if (p.propertyType === PropertyType.INDEXED_GROUP || p.propertyType === PropertyType.NAMED_GROUP) {
                        try { walkRects(p); } catch (e) {}
                    }
                }
            }
            walkRects(contents);
        }
    } catch (e) { out.push("EXC " + e.toString() + " line=" + e.line); }
    var m = new File(dir + "verify_timing.done"); m.encoding = "UTF-8"; m.open("w"); m.write(out.join("\n")); m.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
