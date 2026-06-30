// flightdeck/showcase/glitch/render.jsx — open glitch.aep and save a frame at the
// peak-tear moment (t=1.5s, Max H Displacement keyframe = 60) as glitch.png for
// eyeballing. Force 8bpc + $.sleep so the async write flushes.
// Run via scripts/ae-worker/ae_run.ps1 after `go run ./flightdeck/showcase/glitch`.
(function () {
    var dir = "e:/projects/tools/aep-parser/flightdeck/showcase/glitch/";
    var log = [];
    try {
        app.open(new File(dir + "glitch.aep"));
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "GLITCH") { comp = it; break; }
        }
        if (!comp) { log.push("FAIL: comp GLITCH not found"); }
        else {
            app.project.bitsPerChannel = 8;
            try { app.purge(PurgeTarget.ALL_CACHES); } catch (e) {}
            comp.saveFrameToPng(1.5, new File(dir + "glitch.png"));
            $.sleep(2000);
            log.push("rendered glitch.png");
        }
    } catch (e) { log.push("EXC " + e.toString() + " line=" + e.line); }
    var m = new File(dir + "glitch.done"); m.open("w"); m.write(log.join("\n")); m.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
