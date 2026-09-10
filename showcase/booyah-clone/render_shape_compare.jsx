// render_shape_compare.jsx — render frames 0..N-1 of a named comp from BOTH the
// original and the clone, to <folder>/<srcPrefix>_fNN.png and <clonePrefix>_fNN.png.
// One AE session renders both files for an apples-to-apples per-frame diff.
// Sidecar render_shape_compare.txt (UTF-8), one field per line:
//   line1: original .aep absolute path
//   line2: clone .aep absolute path
//   line3: comp name
//   line4: frame count N (renders frames 0..N-1)
//   line5: output folder (absolute, trailing slash)
//   line6: source prefix (e.g. 01_source)
//   line7: clone prefix  (e.g. 01_copy)
// Writes render_shape_compare.done on completion.
(function () {
    var dir = (new File($.fileName).parent.fsName + "/");
    var log = [];

    function pad2(n) { return (n < 10 ? "0" : "") + n; }

    function findComp(name) {
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === name) return it;
        }
        return null;
    }

    function renderSeq(aepPath, compName, nFrames, folder, prefix) {
        app.open(new File(aepPath));
        var comp = findComp(compName);
        if (!comp) { log.push("FAIL no comp " + compName + " in " + aepPath); return; }
        app.project.bitsPerChannel = 8;
        var fd = comp.frameDuration;
        log.push(prefix + ": comp " + comp.name + " fps=" + comp.frameRate.toFixed(4) + " frameDur=" + fd.toFixed(6));
        for (var f = 0; f < nFrames; f++) {
            try { app.purge(PurgeTarget.ALL_CACHES); } catch (e) {}
            var t = f * fd;
            comp.saveFrameToPng(t, new File(folder + prefix + "_f" + pad2(f) + ".png"));
            $.sleep(120);
        }
        log.push(prefix + ": rendered " + nFrames + " frames");
        try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    }

    try {
        var sf = new File(dir + "render_shape_compare.txt");
        sf.encoding = "UTF-8"; sf.open("r");
        var origPath = (sf.readln() || "").replace(/[\r\n]+$/, "");
        var clonePath = (sf.readln() || "").replace(/[\r\n]+$/, "");
        var compName = (sf.readln() || "").replace(/[\r\n]+$/, "");
        var nFrames = parseInt((sf.readln() || "31").replace(/[\r\n]+$/, ""), 10);
        var folder = (sf.readln() || "").replace(/[\r\n]+$/, "");
        var srcPrefix = (sf.readln() || "01_source").replace(/[\r\n]+$/, "");
        var clonePrefix = (sf.readln() || "01_copy").replace(/[\r\n]+$/, "");
        sf.close();
        renderSeq(origPath, compName, nFrames, folder, srcPrefix);
        renderSeq(clonePath, compName, nFrames, folder, clonePrefix);
    } catch (e) { log.push("EXC " + e.toString() + " line=" + e.line); }

    var m = new File(dir + "render_shape_compare.done"); m.encoding = "UTF-8"; m.open("w"); m.write(log.join("\n")); m.close();
    try { app.quit(); } catch (e) {}
})();
