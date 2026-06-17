// Acceptance verify for classic track matte (SetTrackMatte mode byte). Reads
// track_matte_classic_args.json {input, done, resaved}. Target "T" has the
// matte "M" immediately above; AE should read T.trackMatteType == ALPHA.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/track_matte_classic_args.json");
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
            if (it instanceof CompItem && it.name === "MATTE") { comp = it; break; }
        }
        if (!comp) { fail("comp MATTE not found"); }
        else {
            var T = comp.layer("T");
            if (!T) { fail("layer T not found"); }
            else {
                note("T.trackMatteType=" + T.trackMatteType + " (ALPHA=" + TrackMatteType.ALPHA + ")");
                if (T.trackMatteType !== TrackMatteType.ALPHA)
                    fail("T.trackMatteType=" + T.trackMatteType + ", want ALPHA(" + TrackMatteType.ALPHA + ")");
            }
        }
        app.project.save(new File(args.resaved));
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
