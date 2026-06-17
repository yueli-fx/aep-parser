// Acceptance verify for the three remaining Offset Paths sub-streams (priority-3
// shape remaining): the AE-default-elided `ADBE Vector Offset Line Join` (enum),
// `Offset Miter Limit` (scalar), `Offset Copy Offset` (scalar), each materialized
// from scratch by synthesis-insert (spliced in canonical order). Reads
// mg_offset_extras_args.json {input, done, resaved, png}. Proves AE accepts the
// file, reads back each value off its card, resaves, and renders frame 0 so Go
// can assert on actual pixels (corner shape for Line Join / Miter Limit; offset
// extent for Copy Offset). Red line 4.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/mg_offset_extras_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    function note(m) { log.push("  " + m); }

    function offsetOf(comp, layerName) {
        var lyr = null;
        for (var li = 1; li <= comp.numLayers; li++) if (comp.layer(li).name === layerName) lyr = comp.layer(li);
        if (!lyr) { fail(layerName + " layer missing"); return null; }
        return lyr.property("ADBE Root Vectors Group").property(1)
            .property("ADBE Vectors Group").property("ADBE Vector Filter - Offset");
    }

    try {
        app.open(new File(args.input));
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "OFFEXTRAS") { comp = it; break; }
        }
        if (!comp) { fail("comp OFFEXTRAS not found"); }
        else {
            var bevel = offsetOf(comp, "BEVEL");
            if (bevel) {
                var lj = bevel.property("ADBE Vector Offset Line Join").value;
                note("BEVEL Line Join=" + lj);
                if (Math.abs(lj - 3) > 0.01) fail("BEVEL Line Join=" + lj + ", want 3 (Bevel)");
            }
            var ml = offsetOf(comp, "MITERLIM");
            if (ml) {
                var mlim = ml.property("ADBE Vector Offset Miter Limit").value;
                note("MITERLIM Miter Limit=" + mlim);
                if (Math.abs(mlim - 1) > 0.01) fail("MITERLIM Miter Limit=" + mlim + ", want 1");
            }
            var c2 = offsetOf(comp, "COPY2");
            if (c2) {
                var co = c2.property("ADBE Vector Offset Copy Offset").value;
                var cp = c2.property("ADBE Vector Offset Copies").value;
                note("COPY2 Copies=" + cp + " Copy Offset=" + co);
                if (Math.abs(co - 2) > 0.01) fail("COPY2 Copy Offset=" + co + ", want 2");
                if (Math.abs(cp - 3) > 0.01) fail("COPY2 Copies=" + cp + ", want 3");
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
