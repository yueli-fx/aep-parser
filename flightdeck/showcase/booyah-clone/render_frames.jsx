// render_frames.jsx — render N frames of the first (or named) comp of an .aep to
// <prefix>_t<ms>.png, for clone-vs-original mid-frame pixel comparison of comp①.
// Sidecar render_frames.txt (UTF-8), one field per line:
//   line1: absolute .aep path
//   line2: comp name ("" = first comp)
//   line3: comma-separated times in seconds (e.g. "0.3,0.5,0.8")
//   line4: absolute output prefix (PNGs -> <prefix>_t300.png etc., ms)
// Run via scripts/ae_run.ps1; writes render_frames.done on completion.
(function () {
    var dir = "e:/projects/tools/aep-parser/flightdeck/showcase/booyah-clone/";
    var log = [];
    try {
        var sf = new File(dir + "render_frames.txt");
        sf.encoding = "UTF-8";
        sf.open("r");
        var aepPath = (sf.readln() || "").replace(/[\r\n]+$/, "");
        var compName = (sf.readln() || "").replace(/[\r\n]+$/, "");
        var timesCsv = (sf.readln() || "").replace(/[\r\n]+$/, "");
        var prefix = (sf.readln() || "").replace(/[\r\n]+$/, "");
        sf.close();
        var times = timesCsv.split(",");

        app.open(new File(aepPath));
        var comp = null, first = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (!(it instanceof CompItem)) continue;
            if (!first) first = it;
            if (compName && it.name === compName) { comp = it; break; }
        }
        if (!comp) comp = first;
        if (!comp) { log.push("FAIL: no comp"); }
        else {
            app.project.bitsPerChannel = 8;
            for (var k = 0; k < times.length; k++) {
                try { app.purge(PurgeTarget.ALL_CACHES); } catch (e) {}
                var t = parseFloat(times[k]);
                var ms = Math.round(t * 1000);
                comp.saveFrameToPng(t, new File(prefix + "_t" + ms + ".png"));
                $.sleep(800);
                log.push("rendered " + comp.name + " t=" + t + " -> " + ms);
            }
        }
    } catch (e) { log.push("EXC " + e.toString() + " line=" + e.line); }
    var m = new File(dir + "render_frames.done"); m.encoding = "UTF-8"; m.open("w"); m.write(log.join("\n")); m.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
