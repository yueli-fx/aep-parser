// dump_layer_time.jsx — dump per-layer TIME properties (start/in/out/stretch/
// timeRemap/source) for comp ⑤ in BOTH original and clone, to find the nested-
// comp time-mapping difference on the animated precomp layers.
// Sidecar dump_layer_time.txt: line1 orig, line2 clone, line3 comp name.
(function () {
    var dir = "e:/projects/tools/aep-parser/flightdeck/showcase/booyah-clone/";
    var out = [];
    function findComp(name) {
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === name) return it;
        }
        return null;
    }
    function dump(tag, aepPath, compName) {
        app.open(new File(aepPath));
        var comp = findComp(compName);
        out.push("===== " + tag + " " + compName + " fps=" + (comp ? comp.frameRate.toFixed(4) : "?") + " =====");
        if (!comp) { out.push("no comp"); return; }
        for (var li = 1; li <= comp.numLayers; li++) {
            var L = comp.layer(li);
            var src = L.source;
            var srcInfo = src ? (src.name + " dur=" + src.duration.toFixed(4) + (src instanceof CompItem ? " fps=" + src.frameRate.toFixed(4) : "")) : "<none>";
            var tr = "n/a";
            try { tr = L.timeRemapEnabled; } catch (e) {}
            out.push("L" + li + " start=" + L.startTime.toFixed(6) + " in=" + L.inPoint.toFixed(6) + " out=" + L.outPoint.toFixed(6) +
                     " stretch=" + L.stretch + " timeRemap=" + tr + " src=[" + srcInfo + "]");
        }
        try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    }
    try {
        var sf = new File(dir + "dump_layer_time.txt");
        sf.encoding = "UTF-8"; sf.open("r");
        var o = (sf.readln() || "").replace(/[\r\n]+$/, "");
        var c = (sf.readln() || "").replace(/[\r\n]+$/, "");
        var cn = (sf.readln() || "").replace(/[\r\n]+$/, "");
        sf.close();
        dump("ORIG", o, cn);
        dump("CLONE", c, cn);
    } catch (e) { out.push("EXC " + e.toString() + " line=" + e.line); }
    var m = new File(dir + "dump_layer_time.done"); m.encoding = "UTF-8"; m.open("w"); m.write(out.join("\n")); m.close();
    try { app.quit(); } catch (e) {}
})();
