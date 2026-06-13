// flightdeck/showcase/masks/render.jsx — open masks.aep, dump each layer's mask
// names / modes / inverted flags / vertex counts (so the gen.go paths can be
// verified), then render frame 0 to masks.png for review. Run via scripts/ae_run.ps1.
(function () {
    var dir = "e:/projects/tools/aep-parser/flightdeck/showcase/masks/";
    var inAep = new File(dir + "masks.aep");
    var png = new File(dir + "masks.png");
    var done = new File(dir + "masks.done");
    var log = [];
    var ok = false;
    try {
        app.open(inAep);
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "MasksShowcase") { comp = it; break; }
        }
        if (!comp) throw new Error("comp not found");
        log.push("comp " + comp.width + "x" + comp.height + " layers=" + comp.numLayers);
        // Dump masks per layer (name, mode, inverted, vertex count) for verification.
        for (var li = 1; li <= comp.numLayers; li++) {
            var lyr = comp.layer(li);
            var mg = lyr.property("ADBE Mask Parade");
            if (mg && mg.numProperties > 0) {
                for (var mi = 1; mi <= mg.numProperties; mi++) {
                    var mk = mg.property(mi); // MaskPropertyGroup
                    var shape = mk.property("ADBE Mask Shape").value; // Shape object
                    var nVerts = shape.vertices ? shape.vertices.length : -1;
                    var modeName = "" + mk.maskMode;
                    log.push(lyr.name + " :: " + mk.name +
                        " | mode=" + modeName +
                        " inverted=" + mk.inverted +
                        " closed=" + shape.closed +
                        " verts=" + nVerts);
                }
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
