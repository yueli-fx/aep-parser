// flightdeck/showcase/booyah-clone/render.jsx — open booyah-clone.aep and save a
// frame as booyah-clone.png for eyeballing / pixel comparison against the original.
// Renders the comp named in render_target.txt (UTF-8 sidecar, avoids embedding
// Japanese names in this source); falls back to the FIRST comp if the sidecar is
// absent/unmatched. 8bpc + purge + sleep so the async write flushes. Run via
// scripts/ae_run.ps1 after the generator; -Done = render.done.
(function () {
    var dir = "e:/projects/tools/aep-parser/flightdeck/showcase/booyah-clone/";
    var T = 1.0; // time (seconds)
    var log = [];
    try {
        var want = "";
        var tf = new File(dir + "render_target.txt");
        if (tf.exists) { tf.open("r"); want = (tf.read() || "").replace(/[\r\n]+$/, ""); tf.close(); }

        app.open(new File(dir + "booyah-clone.aep"));
        var comp = null, first = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (!(it instanceof CompItem)) continue;
            if (!first) first = it;
            if (want && it.name === want) { comp = it; break; }
        }
        if (!comp) comp = first;
        if (!comp) { log.push("FAIL: no comp"); }
        else {
            app.project.bitsPerChannel = 8;
            try { app.purge(PurgeTarget.ALL_CACHES); } catch (e) {}
            comp.saveFrameToPng(T, new File(dir + "booyah-clone.png"));
            $.sleep(2000);
            log.push("rendered " + comp.name + " t=" + T);
        }
    } catch (e) { log.push("EXC " + e.toString() + " line=" + e.line); }
    var m = new File(dir + "render.done"); m.open("w"); m.write(log.join("\n")); m.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
