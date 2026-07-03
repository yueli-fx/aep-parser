// flightdeck/showcase/rain/render.jsx — open rain.aep and save the rain showcase
// at t=2.0s, where falling streak keyframes fill the frame.
(function () {
    var dir = "e:/projects/tools/aep-parser/flightdeck/showcase/rain/";
    var log = [];
    try {
        app.open(new File(dir + "rain.aep"));
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "RAIN") { comp = it; break; }
        }
        if (!comp) { log.push("FAIL: comp RAIN not found"); }
        else {
            app.project.bitsPerChannel = 8;
            try { app.purge(PurgeTarget.ALL_CACHES); } catch (e) {}
            comp.saveFrameToPng(2.0, new File(dir + "rain.png"));
            $.sleep(2000);
            log.push("rendered rain.png");
        }
    } catch (e) { log.push("EXC " + e.toString() + " line=" + e.line); }
    var m = new File(dir + "rain.done"); m.open("w"); m.write(log.join("\n")); m.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
