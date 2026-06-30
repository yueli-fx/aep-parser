// flightdeck/showcase/stroke-detail/render.jsx — open stroke_detail.aep, dump a
// terse per-layer summary (layer count + each shape layer's stroke detail
// properties where they can be cheaply read), then render frame 0 to
// stroke_detail.png for review. Run via scripts/ae-worker/ae_run.ps1.
(function () {
    var dir = "e:/projects/tools/aep-parser/flightdeck/showcase/stroke-detail/";
    var inAep = new File(dir + "stroke_detail.aep");
    var png = new File(dir + "stroke_detail.png");
    var done = new File(dir + "stroke_detail.done");
    var log = [];
    var ok = false;
    try {
        app.open(inAep);
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "StrokeDetailShowcase") { comp = it; break; }
        }
        if (!comp) throw new Error("comp not found");
        log.push("comp " + comp.width + "x" + comp.height + " layers=" + comp.numLayers);
        // Per-layer: walk the shape root and report stroke detail params so the
        // gen.go values can be cross-checked against what AE actually parsed.
        for (var li = 1; li <= comp.numLayers; li++) {
            var lyr = comp.layer(li);
            var root = lyr.property("ADBE Root Vectors Group");
            if (!root) { log.push(lyr.name + " :: (not a shape layer)"); continue; }
            var strokes = [];
            collectStrokes(root, strokes);
            if (strokes.length === 0) { log.push(lyr.name + " :: (no stroke)"); continue; }
            for (var si = 0; si < strokes.length; si++) {
                log.push(lyr.name + " :: " + describeStroke(strokes[si]));
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

    // Recursively gather every "ADBE Vector Graphic - Stroke" property group.
    function collectStrokes(grp, out) {
        for (var i = 1; i <= grp.numProperties; i++) {
            var pr = grp.property(i);
            if (!pr) continue;
            if (pr.matchName === "ADBE Vector Graphic - Stroke") { out.push(pr); continue; }
            if (pr.numProperties && pr.numProperties > 0) collectStrokes(pr, out);
        }
    }

    // Pull the stroke detail scalars that are easy to read via match-name.
    function describeStroke(s) {
        var parts = [];
        parts.push("W=" + rd(s, "ADBE Vector Stroke Width"));
        parts.push("cap=" + rd(s, "ADBE Vector Stroke Line Cap"));
        parts.push("join=" + rd(s, "ADBE Vector Stroke Line Join"));
        parts.push("miter=" + rd(s, "ADBE Vector Stroke Miter Limit"));
        parts.push("taperEndW=" + rd(s, "ADBE Vector Taper End Width"));
        parts.push("waveAmt=" + rd(s, "ADBE Vector Taper Wave Amount"));
        parts.push("dash=" + rd(s, "ADBE Vector Stroke Dash 1"));
        return parts.join(" ");
    }

    // Read a property value by match-name anywhere under the stroke; "-" if absent.
    function rd(s, mn) {
        var p = findByMatch(s, mn);
        if (!p) return "-";
        try { return "" + p.value; } catch (e) { return "?"; }
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
