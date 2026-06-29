// Acceptance verify for >4-vertex path geometry (priority-1 leftover): a Go-built
// n=5 pentagon as both a MASK and a SHAPE path. Reads mg_pentagon_args.json
// {input, done, resaved, png}. Proves AE accepts >4-vertex geometry, reads back
// 5 vertices on each, resaves, and renders frame 0 so Go can assert the pentagon
// silhouette (red line 4).
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/mg_pentagon_args.json");
    argsFile.open("r"); var raw = argsFile.read(); argsFile.close();
    var args = eval("(" + raw + ")");
    var log = []; var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    function note(m) { log.push("  " + m); }
    try {
        app.open(new File(args.input));
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "PENTA") { comp = it; break; }
        }
        if (!comp) { fail("comp PENTA not found"); }
        else {
            var card = null, sh = null;
            for (var li = 1; li <= comp.numLayers; li++) {
                if (comp.layer(li).name === "CARD") card = comp.layer(li);
                if (comp.layer(li).name === "PENTASHAPE") sh = comp.layer(li);
            }
            if (!card) { fail("CARD missing"); }
            else {
                var shp = card.property("ADBE Mask Parade").property(1).property("ADBE Mask Shape").value;
                note("mask verts=" + shp.vertices.length + " closed=" + shp.closed);
                if (shp.vertices.length !== 5) fail("mask verts=" + shp.vertices.length + ", want 5");
            }
            if (!sh) { fail("PENTASHAPE missing"); }
            else {
                function findShape(g) {
                    for (var k = 1; k <= g.numProperties; k++) {
                        var pr = g.property(k);
                        if (pr.matchName === "ADBE Vector Shape") return pr;
                        if (pr.numProperties && pr.numProperties > 0) { var r = findShape(pr); if (r) return r; }
                    }
                    return null;
                }
                var sp = findShape(sh.property("ADBE Root Vectors Group"));
                if (!sp) { fail("shape path ADBE Vector Shape not found (dropped?)"); }
                else { note("shape path verts=" + sp.value.vertices.length + " closed=" + sp.value.closed);
                       if (sp.value.vertices.length !== 5) fail("shape verts=" + sp.value.vertices.length + ", want 5"); }
            }
        }
        app.project.save(new File(args.resaved));
        if (comp && args.png) {
            app.project.bitsPerChannel = 8;
            comp.saveFrameToPng(0, new File(args.png));
            $.sleep(2000);
            var pf = new File(args.png);
            if (pf.exists) { note("frame png " + pf.length + " bytes"); } else { fail("saveFrameToPng wrote no file"); }
        }
    } catch (e) { fail("EXC " + e.toString() + " line=" + e.line); }
    var done = new File(args.done);
    done.open("w"); done.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n")); done.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
