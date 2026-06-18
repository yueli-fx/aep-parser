// flightdeck/showcase/booyah-clone/render.jsx — open booyah-clone.aep and save a
// frame of the named comp (default = top comp メインコンプ！) as booyah-clone.png
// for eyeballing / pixel comparison against the original. 8bpc + purge + sleep so
// the async write flushes. Run via scripts/ae_run.ps1 after the generator.
//
// CAVEAT (resolve at first real render): comp names are Japanese. ExtendScript
// source must be UTF-8 for `it.name === COMP` to match; if matching fails, drive
// COMP via a \u-escaped string or select by index instead.
(function () {
    var dir = "e:/projects/tools/aep-parser/flightdeck/showcase/booyah-clone/";
    var COMP = "メインコンプ！"; // comp to render
    var T = 1.0;                 // time (seconds)
    var log = [];
    try {
        app.open(new File(dir + "booyah-clone.aep"));
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === COMP) { comp = it; break; }
        }
        if (!comp) { log.push("FAIL: comp " + COMP + " not found"); }
        else {
            app.project.bitsPerChannel = 8;
            try { app.purge(PurgeTarget.ALL_CACHES); } catch (e) {}
            comp.saveFrameToPng(T, new File(dir + "booyah-clone.png"));
            $.sleep(2000);
            log.push("rendered booyah-clone.png @" + COMP + " t=" + T);
        }
    } catch (e) { log.push("EXC " + e.toString() + " line=" + e.line); }
    var m = new File(dir + "render.done"); m.open("w"); m.write(log.join("\n")); m.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
