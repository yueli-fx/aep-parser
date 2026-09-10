// showcase/animated-path/render.jsx — open animated_path.aep, render the
// MID time (t=2s) to animated_path.png (the path is a SQUARE there = morph
// midpoint), and dump the path's vertices at t=0 / t=2 / t=4 to the .done log so
// the bar→square→bar morph is numerically proven. Run via scripts/ae-worker/ae_run.ps1.
(function () {
    var dir = (new File($.fileName).parent.fsName + "/");
    var inAep = new File(dir + "animated_path.aep");
    var png = new File(dir + "animated_path.png");
    var done = new File(dir + "animated_path.done");
    var log = [];
    var ok = false;
    try {
        app.open(inAep);
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "AnimatedPath") { comp = it; break; }
        }
        if (!comp) throw new Error("comp not found");
        log.push("comp layers=" + comp.numLayers);
        // Find the ADBE Vector Shape path property on the Morph layer and dump its
        // bounding extent at three times so the interpolation is visible as numbers.
        var morph = comp.layer("Morph");
        var shp = findByMatch(morph.property("ADBE Root Vectors Group"), "ADBE Vector Shape");
        if (!shp) throw new Error("path property not found");
        var times = [0, 2, 4];
        for (var ti = 0; ti < times.length; ti++) {
            var sh = shp.valueAtTime(times[ti], false);
            log.push("t=" + times[ti] + " extent=" + extent(sh.vertices));
        }
        app.project.bitsPerChannel = 8;
        comp.saveFrameToPng(2.0, png); // t=2s (mid) → square
        $.sleep(2500);
        log.push(png.exists ? ("png " + png.length + " bytes") : "PNG MISSING");
        ok = png.exists;
    } catch (e) { log.push("ERROR: " + e.toString() + " line=" + e.line); }
    done.open("w");
    done.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
    done.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}

    // width×height of a vertex set, rounded — proves the bar↔square↔bar morph.
    function extent(verts) {
        var minX = 1e9, maxX = -1e9, minY = 1e9, maxY = -1e9;
        for (var i = 0; i < verts.length; i++) {
            minX = Math.min(minX, verts[i][0]); maxX = Math.max(maxX, verts[i][0]);
            minY = Math.min(minY, verts[i][1]); maxY = Math.max(maxY, verts[i][1]);
        }
        return Math.round(maxX - minX) + "x" + Math.round(maxY - minY);
    }
    function findByMatch(grp, mn) {
        for (var i = 1; i <= grp.numProperties; i++) {
            var pr = grp.property(i);
            if (!pr) continue;
            if (pr.matchName === mn) return pr;
            if (pr.numProperties && pr.numProperties > 0) {
                var hit = findByMatch(pr, mn);
                if (hit) return hit;
            }
        }
        return null;
    }
})();
