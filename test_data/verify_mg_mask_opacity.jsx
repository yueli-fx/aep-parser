// Acceptance verify for Mask Opacity (priority-4 mask). Reads
// mg_mask_opacity_args.json {input, done, resaved, png}. Proves AE accepts the
// from-scratch mask with a synthesis-inserted `ADBE Mask Opacity` leaf, reads
// back each card's mask opacity (100 / 50), resaves, and renders frame 0 so Go
// can assert on pixels that the 50% mask dims its revealed region to ~half
// luminance while the 100% mask reveals full white — red line 4: render the
// capability's surface.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/mg_mask_opacity_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    function note(m) { log.push("  " + m); }

    var want = { "MASKFULL": 100, "MASKHALF": 50 };

    try {
        app.open(new File(args.input));

        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "MASKOP") { comp = it; break; }
        }
        if (!comp) { fail("comp MASKOP not found"); }
        else {
            for (var li = 1; li <= comp.numLayers; li++) {
                var lay = comp.layer(li);
                if (!(lay.name in want)) continue;
                var mk = lay.property("ADBE Mask Parade").property(1);
                if (!mk) { fail(lay.name + " has no mask"); continue; }
                var op = mk.property("ADBE Mask Opacity").value;
                note(lay.name + " maskOpacity=" + op);
                if (Math.abs(op - want[lay.name]) > 1) fail(lay.name + " maskOpacity=" + op + ", want " + want[lay.name]);
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
