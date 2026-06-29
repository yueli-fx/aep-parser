// Acceptance verify for MG roadmap S4: precomp / nested-comp layer from scratch.
// Reads mg_precomp_args.json {input, done, resaved, png}. Proves AE accepts the
// file, that the parent's single layer is a precomp whose source is the CHILD
// comp (not footage / wrong comp), resaves, and renders frame 0 so Go can assert
// on actual pixels that the child comp's content shows through the precomp layer
// (red line 4 — render the capability's visible surface).
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/mg_precomp_args.json");
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

        var parent = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "MGPREC_Parent") { parent = it; break; }
        }
        if (!parent) { fail("comp MGPREC_Parent not found"); }
        else {
            note("parent=" + parent.name + " layers=" + parent.numLayers);
            if (parent.numLayers !== 1) fail("expected 1 layer in parent, got " + parent.numLayers);

            var pre = parent.layer(1);
            note("precomp layer name=" + pre.name);
            var src = pre.source;
            if (!src) fail("precomp layer has no source");
            else {
                note("precomp source name=" + src.name + " isComp=" + (src instanceof CompItem));
                if (!(src instanceof CompItem)) fail("precomp source is not a CompItem");
                if (src.name !== "MGPREC_Child") fail("precomp source=" + src.name + ", want MGPREC_Child");
            }
        }

        app.project.save(new File(args.resaved));

        if (parent && args.png) {
            app.project.bitsPerChannel = 8;
            parent.saveFrameToPng(0, new File(args.png));
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
