// Acceptance verify for SetMaskPath (priority-4 mask): reshaping an existing
// mask's outline (rectangle → triangle, 4→3 vertices). Reads
// mg_mask_path_args.json {input, done, resaved, png}. Proves AE accepts the
// reshaped mask, reads back the new vertex count (3) + closed, resaves, and
// renders frame 0 so Go can assert the revealed region is now triangular (the
// original rectangle's corners are masked out) — red line 4: render the
// capability's surface; a vertex count that only round-trips proves nothing about
// AE's clip.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/mg_mask_path_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    function note(m) { log.push("  " + m); }

    try {
        app.open(new File(args.input));

        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "MASKPATH") { comp = it; break; }
        }
        if (!comp) { fail("comp MASKPATH not found"); }
        else {
            var card = null;
            for (var li = 1; li <= comp.numLayers; li++) if (comp.layer(li).name === "CARD") card = comp.layer(li);
            if (!card) { fail("CARD layer missing"); }
            else {
                var mk = card.property("ADBE Mask Parade").property(1);
                if (!mk) { fail("CARD has no mask"); }
                else {
                    var shp = mk.property("ADBE Mask Shape").value;
                    note("mask verts=" + shp.vertices.length + " closed=" + shp.closed);
                    if (shp.vertices.length !== 3) fail("mask verts=" + shp.vertices.length + ", want 3 (triangle)");
                    if (!shp.closed) fail("mask not closed");
                }
            }
        }

        app.project.save(new File(args.resaved));

        if (comp && args.png) {
            app.project.bitsPerChannel = 8;
            comp.saveFrameToPng(0, new File(args.png));
            $.sleep(2000);
            var pngFile = new File(args.png);
            if (pngFile.exists) { note("frame png " + pngFile.length + " bytes"); }
            else { fail("saveFrameToPng wrote no file"); }
        }
    } catch (e) {
        fail("EXC " + e.toString() + " line=" + e.line);
    }

    var done = new File(args.done);
    done.open("w");
    done.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n"));
    done.close();

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
