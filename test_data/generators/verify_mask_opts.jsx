// Ship-gate verify for batch-10 Mask option setters. Reads mask_opts_args.json
// {input, done, resaved}. Reads back the mask's options from AE's Mask DOM,
// asserts the Go-written values, resaves.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/mask_opts_args.json");
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
            if (it instanceof CompItem) { comp = it; break; }
        }
        if (!comp) { fail("no CompItem"); }
        else {
            var L = null;
            for (var li = 1; li <= comp.numLayers; li++) {
                var cand = comp.layer(li);
                if (cand.property("ADBE Mask Parade") && cand.property("ADBE Mask Parade").numProperties > 0) { L = cand; break; }
            }
            if (!L) { fail("no layer with mask"); }
            else {
                var m = L.mask(1);
                eq("maskMode", m.maskMode, MaskMode.SUBTRACT);
                eq("inverted", m.inverted, true);
                eq("locked", m.locked, true);
                var c = m.color;
                near("color.r", c[0], 1.0); near("color.g", c[1], 0.502); near("color.b", c[2], 0.0);
                eq("maskMotionBlur", m.maskMotionBlur, MaskMotionBlur.ON);
                var fe = m.maskFeather.value;
                near("feather.x", fe[0], 10); near("feather.y", fe[1], 20);
                near("expansion", m.maskExpansion.value, 15);
                eq("closed", m.maskShape.value.closed, false);
            }
        }
        app.project.save(new File(args.resaved));
    } catch (e) { fail("EXC " + e.toString()); }

    var done = new File(args.done);
    done.open("w"); done.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n")); done.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
