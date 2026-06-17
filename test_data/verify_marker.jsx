// Ship-gate verify for comp marker structural add/remove (P3 §3G). Opens the
// Go-written file (RE_CM had its first marker removed + a new marker added at
// addedTime) and asserts AE reads expectMarkers comp markers, with the added
// marker (addedComment @ addedTime) and the survivor (survComment) both
// present. Reads marker_args.json {input, done, resaved, compName,
// expectMarkers, addedComment, addedTime, survComment}. Writes PASS/FAIL +
// diagnostics, resaves.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/marker_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  " + m); }

    try {
        app.open(new File(args.input));
        var proj = app.project;
        var comp = null;
        for (var i = 1; i <= proj.numItems; i++) {
            var it = proj.item(i);
            if (it instanceof CompItem && it.name === args.compName) { comp = it; break; }
        }
        if (!comp) {
            fail("comp " + args.compName + " not found");
        } else {
            var mp = comp.markerProperty;
            log.push("  numKeys=" + mp.numKeys);
            if (mp.numKeys !== args.expectMarkers) {
                fail("numKeys " + mp.numKeys + " != " + args.expectMarkers);
            }
            var foundAdded = false, foundSurv = false;
            for (var k = 1; k <= mp.numKeys; k++) {
                var mv = mp.keyValue(k);
                var tt = mp.keyTime(k);
                log.push("  key" + k + " t=" + tt + " comment=" + mv.comment);
                if (mv.comment === args.addedComment && Math.abs(tt - args.addedTime) < 0.01) {
                    foundAdded = true;
                }
                if (mv.comment === args.survComment) {
                    foundSurv = true;
                }
            }
            if (!foundAdded) {
                fail("added marker (comment=" + args.addedComment + " t=" + args.addedTime + ") not found");
            }
            if (!foundSurv) {
                fail("survivor marker (comment=" + args.survComment + ") not found");
            }
        }
        app.project.save(new File(args.resaved));
    } catch (e) {
        fail("EXC " + e.toString());
    }

    var done = new File(args.done);
    done.open("w");
    done.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n"));
    done.close();

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
