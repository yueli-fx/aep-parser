// flightdeck/showcase/structural-ops/render.jsx — open structural_ops.aep, dump
// the comp's layer roster (proof of the Delete/Duplicate/Move structural result)
// plus demo-D's Position "Separate Dimensions" state (proof of
// SetDimensionsSeparated, which has no single-frame pixel signature), then render
// frame 0 to structural_ops.png for review. Run via scripts/ae-worker/ae_run.ps1.
//
// Layout is a CENTRED CONCENTRIC bullseye (solids can't be repositioned — see
// gen.go). Verifiable results:
//   Delete:    C_mid_green ABSENT; C_inner_pink + C_outer_pink present →
//              render shows the green ring gone (pink shows through).
//   Move:      B_violet_back now ABOVE B_amber_front (lower AE index) → the
//              centre 220px square renders VIOLET (was amber before the move).
//   Duplicate: A_dup_src AND A_dup_clone both present (clone overlaps the source
//              exactly — same footage colour+size, centred — so it is a purely
//              structural proof, no extra pixels in the render).
//   Separate:  D_sep_shape Position dimensionsSeparated == true.
(function () {
    var dir = "e:/projects/tools/aep-parser/flightdeck/showcase/structural-ops/";
    var inAep = new File(dir + "structural_ops.aep");
    var png = new File(dir + "structural_ops.png");
    var done = new File(dir + "structural_ops.done");
    var log = [];
    var ok = false;
    try {
        app.open(inAep);
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "StructuralOpsShowcase") { comp = it; break; }
        }
        if (!comp) throw new Error("comp not found");
        log.push("comp " + comp.width + "x" + comp.height + " layers=" + comp.numLayers);

        // Layer roster (top of stack = AE index 1).
        var hasMidGreen = false, hasInnerPink = false, hasOuterPink = false;
        var hasClone = false, hasSrc = false;
        var amberIdx = -1, violetIdx = -1;
        for (var li = 1; li <= comp.numLayers; li++) {
            var nm = comp.layer(li).name;
            log.push("  L" + li + " " + nm);
            if (nm === "C_mid_green") hasMidGreen = true;
            if (nm === "C_inner_pink") hasInnerPink = true;
            if (nm === "C_outer_pink") hasOuterPink = true;
            if (nm === "A_dup_clone") hasClone = true;
            if (nm === "A_dup_src") hasSrc = true;
            if (nm === "B_amber_front") amberIdx = li;
            if (nm === "B_violet_back") violetIdx = li;
        }
        log.push("DELETE check: C_mid_green absent = " + (!hasMidGreen) +
            " (C_inner_pink=" + hasInnerPink + " C_outer_pink=" + hasOuterPink + ")");
        log.push("DUPLICATE check: A_dup_src AND A_dup_clone present = " + (hasSrc && hasClone));
        log.push("MOVE check: B_violet_back above B_amber_front (violetIdx<amberIdx) = " +
            (violetIdx > 0 && amberIdx > 0 && violetIdx < amberIdx) +
            " (violetIdx=" + violetIdx + " amberIdx=" + amberIdx + ")");

        // SetDimensionsSeparated proof: inspect demo-D shape's Position leader.
        var d = null;
        for (var si = 1; si <= comp.numLayers; si++) {
            if (comp.layer(si).name === "D_sep_shape") { d = comp.layer(si); break; }
        }
        if (d) {
            var pos = d.property("ADBE Transform Group").property("ADBE Position");
            log.push("D_sep_shape Position dimensionsSeparated=" + pos.dimensionsSeparated);
            try {
                var xDim = d.property("ADBE Transform Group").property("ADBE Position_0");
                log.push("D_sep_shape X-axis (Position_0) present=" + (xDim != null) +
                    (xDim ? (" name='" + xDim.name + "'") : ""));
            } catch (e) { log.push("D_sep_shape Position_0 lookup: " + e.toString()); }
        } else {
            log.push("D_sep_shape not found for dimension check");
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
