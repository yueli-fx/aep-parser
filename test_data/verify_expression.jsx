// Acceptance verify for MG roadmap S2: expression activation. Reads
// expression_args.json {input, done, resaved, png}. Proves AE accepts the
// file, the ON layer's rotation expression is enabled and EVALUATES
// (valueAtTime(2) = 180), the OFF layer's identical expression stays
// disabled (valueAtTime(2) = 0), resaves, and renders t=2s for the Go
// side's pixel assertions (red line 4).
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/expression_args.json");
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
            if (it instanceof CompItem && it.name === "EXPR") { comp = it; break; }
        }
        if (!comp) { fail("comp EXPR not found"); }
        else {
            note("comp=" + comp.name + " layers=" + comp.numLayers);
            var byName = {};
            for (var li = 1; li <= comp.numLayers; li++) byName[comp.layer(li).name] = comp.layer(li);

            var on = byName["ON"];
            if (!on) { fail("ON missing"); }
            else {
                var rot = on.transform.rotation;
                note("ON expr='" + rot.expression + "' enabled=" + rot.expressionEnabled);
                if (rot.expression !== "time*90") fail("ON expression text mismatch");
                if (!rot.expressionEnabled) fail("ON expressionEnabled=false — activation bit still wrong");
                var v = rot.valueAtTime(2, false);
                note("ON rotation@2s=" + v);
                if (Math.abs(v - 180) > 0.01) fail("ON rotation@2s=" + v + ", want 180 — expression not evaluating");
            }

            var off = byName["OFF"];
            if (!off) { fail("OFF missing"); }
            else {
                var rot2 = off.transform.rotation;
                note("OFF expr='" + rot2.expression + "' enabled=" + rot2.expressionEnabled);
                if (rot2.expressionEnabled) fail("OFF expressionEnabled=true, want false");
                var v2 = rot2.valueAtTime(2, false);
                note("OFF rotation@2s=" + v2);
                if (Math.abs(v2) > 0.01) fail("OFF rotation@2s=" + v2 + ", want 0 (disabled)");
            }
        }

        app.project.save(new File(args.resaved));

        if (comp && args.png) {
            app.project.bitsPerChannel = 8;
            comp.saveFrameToPng(2, new File(args.png));
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
