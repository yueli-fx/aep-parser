// render_solo_layers.jsx — solo each layer of a comp one at a time and render a
// single frame, for BOTH original and clone, to isolate which precomp layer's
// composite contribution differs.
// Sidecar render_solo_layers.txt (UTF-8):
//   line1: original .aep
//   line2: clone .aep
//   line3: comp name
//   line4: frame index (e.g. 18)
//   line5: output folder (trailing slash)
(function () {
    var dir = (new File($.fileName).parent.fsName + "/");
    var log = [];
    function findComp(name) {
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === name) return it;
        }
        return null;
    }
    function soloRender(aepPath, compName, frame, folder, tag) {
        app.open(new File(aepPath));
        var comp = findComp(compName);
        if (!comp) { log.push("FAIL no comp " + compName); return; }
        app.project.bitsPerChannel = 8;
        var t = frame * comp.frameDuration;
        for (var li = 1; li <= comp.numLayers; li++) {
            for (var k = 1; k <= comp.numLayers; k++) comp.layer(k).solo = (k === li);
            try { app.purge(PurgeTarget.ALL_CACHES); } catch (e) {}
            comp.saveFrameToPng(t, new File(folder + tag + "_L" + li + "_f" + frame + ".png"));
            $.sleep(150);
            log.push(tag + " L" + li + " (" + comp.layer(li).name + ") start=" + comp.layer(li).startTime.toFixed(4) + " op=" + comp.layer(li).property("ADBE Transform Group").property("ADBE Opacity").valueAtTime(t, false));
        }
        for (var k2 = 1; k2 <= comp.numLayers; k2++) comp.layer(k2).solo = false;
        try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    }
    try {
        var sf = new File(dir + "render_solo_layers.txt");
        sf.encoding = "UTF-8"; sf.open("r");
        var origPath = (sf.readln() || "").replace(/[\r\n]+$/, "");
        var clonePath = (sf.readln() || "").replace(/[\r\n]+$/, "");
        var compName = (sf.readln() || "").replace(/[\r\n]+$/, "");
        var frame = parseInt((sf.readln() || "18").replace(/[\r\n]+$/, ""), 10);
        var folder = (sf.readln() || "").replace(/[\r\n]+$/, "");
        sf.close();
        soloRender(origPath, compName, frame, folder, "src");
        soloRender(clonePath, compName, frame, folder, "cln");
    } catch (e) { log.push("EXC " + e.toString() + " line=" + e.line); }
    var m = new File(dir + "render_solo_layers.done"); m.encoding = "UTF-8"; m.open("w"); m.write(log.join("\n")); m.close();
    try { app.quit(); } catch (e) {}
})();
