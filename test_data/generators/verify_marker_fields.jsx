// Ship-gate verify for batch-9 Marker field setters. Reads marker_fields_args.json
// {input, done, resaved}. Reads back each comp marker's fields from AE's
// MarkerValue DOM (matched by comment), asserts the Go-written values, resaves.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/marker_fields_args.json");
    argsFile.open("r"); var raw = argsFile.read(); argsFile.close();
    var args = eval("(" + raw + ")");
    var ok = true, log = [];
    function fail(m) { ok = false; log.push("  FAIL " + m); }
    function eq(label, got, want) { if (got !== want) fail(label + "=" + got + " want " + want); else log.push("  " + label + "=" + got); }
    function near(label, got, want) { if (Math.abs(got - want) > 0.02) fail(label + "=" + got + " want " + want); else log.push("  " + label + "=" + got); }

    try {
        app.open(new File(args.input));
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "RE_CM") { comp = it; break; }
        }
        if (!comp) { fail("RE_CM comp not found"); }
        else {
            var mp = comp.markerProperty;
            log.push("  numKeys=" + mp.numKeys);
            var seen = {};
            for (var k = 1; k <= mp.numKeys; k++) {
                var mv = mp.keyValue(k), tt = mp.keyTime(k);
                seen[mv.comment] = true;
                if (mv.comment === "M1") {
                    eq("M1.chapter", mv.chapter, "ch1");
                    eq("M1.cuePointName", mv.cuePointName, "cue1");
                    eq("M1.url", mv.url, "http://e.x");
                    eq("M1.frameTarget", mv.frameTarget, "ft1");
                    near("M1.duration", mv.duration, 1.5);
                    eq("M1.label", mv.label, 4);
                } else if (mv.comment === "M2") {
                    near("M2.time(SetTime)", tt, 3.5);
                } else if (mv.comment === "M3") {
                    near("M3.time(SetFrameTime)", tt, 4.0);
                } else if (mv.comment === "M4") {
                    near("M4.duration(SetFrameDuration)", mv.duration, 1.5);
                }
            }
            for (var ci = 1; ci <= 4; ci++) if (!seen["M" + ci]) fail("marker M" + ci + " missing");
        }
        app.project.save(new File(args.resaved));
    } catch (e) { fail("EXC " + e.toString()); }

    var done = new File(args.done);
    done.open("w"); done.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n")); done.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
