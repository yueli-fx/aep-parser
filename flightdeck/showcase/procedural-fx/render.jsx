// flightdeck/showcase/procedural-fx/render.jsx — open flame.aep and save a
// mid-boil frame (t=2s) as flame.png for eyeballing. Force 8bpc (32bpc float
// saveFrameToPng emits dark linear pixels) + $.sleep so the async write flushes.
// Run via scripts/ae_run.ps1 after `go run ./flightdeck/showcase/procedural-fx`.
(function () {
    var dir = "e:/projects/tools/aep-parser/flightdeck/showcase/procedural-fx/";
    var log = [];
    try {
        app.open(new File(dir + "flame.aep"));
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "FLAME") { comp = it; break; }
        }
        if (!comp) { log.push("FAIL: comp FLAME not found"); }
        else {
            app.project.bitsPerChannel = 8;
            comp.saveFrameToPng(2.0, new File(dir + "flame.png"));
            $.sleep(2000);
            log.push("rendered flame.png");
        }
    } catch (e) { log.push("EXC " + e.toString() + " line=" + e.line); }
    var m = new File(dir + "procedural_fx.done"); m.open("w"); m.write(log.join("\n")); m.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
