// verify_precomp_timing.jsx — dump comp ⑤ precomp-layer transform keyTimes from AE's
// own DOM (ground truth), for BOTH original and clone, to verify the 29.97 fps fix
// moved the clone's from-scratch keyframe TIMES onto the original's.
// Sidecar verify_precomp_timing.txt (UTF-8):
//   line1: original .aep absolute path
//   line2: clone .aep absolute path
//   line3: comp name (e.g. プリコンポジション 1)
// Writes verify_precomp_timing.done (UTF-8).
(function () {
    var dir = "e:/projects/tools/aep-parser/flightdeck/showcase/booyah-clone/";
    var out = [];

    function fmtKeys(prop) {
        if (!prop) return "<nil>";
        var n = 0;
        try { n = prop.numKeys; } catch (e) { return "<no-keys " + e.toString() + ">"; }
        if (n === 0) {
            var sv = "?";
            try { sv = prop.value.toString(); } catch (e2) {}
            return "static=" + sv;
        }
        var kt = [];
        for (var i = 1; i <= n; i++) {
            var v = "?";
            try { v = prop.keyValue(i).toString(); } catch (e3) {}
            kt.push(prop.keyTime(i).toFixed(4) + ":" + v);
        }
        return n + "kf [" + kt.join(" | ") + "]";
    }

    function tryProp(grp, mn) {
        try { return grp.property(mn); } catch (e) { return null; }
    }

    function dumpFile(label, aepPath, compName) {
        out.push("===== " + label + " : " + aepPath + " =====");
        app.open(new File(aepPath));
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === compName) { comp = it; break; }
        }
        if (!comp) { out.push("FAIL no comp " + compName); }
        else {
            out.push("comp " + comp.name + " fps=" + comp.frameRate.toFixed(4) + " dur=" + comp.duration.toFixed(4) + " nLayers=" + comp.numLayers);
            for (var li = 1; li <= comp.numLayers; li++) {
                var L = comp.layer(li);
                out.push("L" + (li - 1) + " " + L.name + " start=" + L.startTime.toFixed(4) + " in=" + L.inPoint.toFixed(4) + " out=" + L.outPoint.toFixed(4));
                var tg = tryProp(L, "ADBE Transform Group");
                if (!tg) { out.push("   <no transform group>"); continue; }
                // Position may be separated (Position_0 / Position_1).
                var pos = tryProp(tg, "ADBE Position");
                var posX = tryProp(tg, "ADBE Position_0");
                var posY = tryProp(tg, "ADBE Position_1");
                var op = tryProp(tg, "ADBE Opacity");
                var anc = tryProp(tg, "ADBE Anchor Point");
                out.push("   Pos   " + fmtKeys(pos));
                if (posX || posY) {
                    out.push("   PosX  " + fmtKeys(posX));
                    out.push("   PosY  " + fmtKeys(posY));
                }
                out.push("   Opac  " + fmtKeys(op));
                out.push("   Anchor " + fmtKeys(anc));
            }
        }
        try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    }

    try {
        var sf = new File(dir + "verify_precomp_timing.txt");
        sf.encoding = "UTF-8"; sf.open("r");
        var origPath = (sf.readln() || "").replace(/[\r\n]+$/, "");
        var clonePath = (sf.readln() || "").replace(/[\r\n]+$/, "");
        var compName = (sf.readln() || "").replace(/[\r\n]+$/, "");
        sf.close();
        dumpFile("ORIGINAL", origPath, compName);
        dumpFile("CLONE", clonePath, compName);
    } catch (e) { out.push("EXC " + e.toString() + " line=" + e.line); }

    var m = new File(dir + "verify_precomp_timing.done"); m.encoding = "UTF-8"; m.open("w"); m.write(out.join("\n")); m.close();
    try { app.quit(); } catch (e) {}
})();
