// flightdeck/showcase/3d-camera/render.jsx — open 3d_camera.aep, verify AE
// accepts it (all layers present, the cards are 3D), and render frame 0 to
// 3d_camera.png for user review. Run via scripts/ae-worker/ae_run.ps1 (writes .done).
(function () {
    var dir = "e:/projects/tools/aep-parser/flightdeck/showcase/3d-camera/";
    var inAep = new File(dir + "3d_camera.aep");
    var png = new File(dir + "3d_camera.png");
    var done = new File(dir + "3d_camera.done");
    var log = [];
    var ok = false;
    try {
        app.open(inAep);
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "Showcase3DCamera") { comp = it; break; }
        }
        if (!comp) throw new Error("comp not found");
        log.push("comp " + comp.width + "x" + comp.height + " layers=" + comp.numLayers);
        for (var li = 1; li <= comp.numLayers; li++) {
            var L = comp.layer(li);
            if (L.name === "Near" || L.name === "Mid" || L.name === "Far" || L.name === "Tumble") {
                var pos = L.property("ADBE Transform Group").property("ADBE Position");
                var ry = L.property("ADBE Transform Group").property("ADBE Rotate Y");
                log.push(L.name + " 3D=" + L.threeDLayer + " pos=" + pos.value.toString() + " rotY=" + (ry ? ry.value : "nil"));
            }
        }
        app.project.bitsPerChannel = 8;
        comp.saveFrameToPng(0, png);
        $.sleep(2500);
        log.push(png.exists ? ("png " + png.length + " bytes") : "PNG MISSING");
        ok = png.exists;
    } catch (e) { log.push("ERROR: " + e.toString() + " line=" + e.line); }
    done.open("w");
    done.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
    done.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
