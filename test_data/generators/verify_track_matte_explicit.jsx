// Acceptance verify for AE23+ explicit-source track matte. Reads
// track_matte_explicit_args.json {input, done, resaved}. Four targets:
//   TA = SetTrackMatteSource(MA, Luma)  → LUMA, matte=MA
//   TB = SetTrackMatteLayer(MB, Alpha)  → ALPHA, matte=MB
//   TC = set then ClearTrackMatteLayer  → NO_TRACK_MATTE
//   TD = set then RemoveTrackMatte      → NO_TRACK_MATTE
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/track_matte_explicit_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    function note(m) { log.push("  " + m); }
    function mname(L) { return (L.trackMatteLayer == null) ? "null" : L.trackMatteLayer.name; }

    try {
        app.open(new File(args.input));
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "EMATTE") { comp = it; break; }
        }
        if (!comp) { fail("comp EMATTE not found"); }
        else {
            var byName = {};
            for (var li = 1; li <= comp.numLayers; li++) byName[comp.layer(li).name] = comp.layer(li);

            function chk(tname, wantType, wantMatte) {
                var L = byName[tname];
                if (!L) { fail(tname + " missing"); return; }
                note(tname + " trackMatteType=" + L.trackMatteType + " matte=" + mname(L));
                if (L.trackMatteType !== wantType) fail(tname + " trackMatteType=" + L.trackMatteType + ", want " + wantType);
                if (wantMatte !== null && mname(L) !== wantMatte) fail(tname + " matte=" + mname(L) + ", want " + wantMatte);
            }
            chk("TA", TrackMatteType.LUMA, "MA");
            chk("TB", TrackMatteType.ALPHA, "MB");
            chk("TC", TrackMatteType.NO_TRACK_MATTE, null);
            chk("TD", TrackMatteType.NO_TRACK_MATTE, null);
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
