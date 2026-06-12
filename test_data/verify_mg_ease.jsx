// Acceptance verify for MG roadmap S1: temporal-ease keyframes + 6-keyframe
// stream, from scratch. Reads mg_ease_args.json {input, done, resaved, png}.
// Proves AE accepts the file, reads back the ease (keyOutTemporalEase
// influence ≈ 90%, BEZIER interp on the eased sides), counts all 6 SCL
// keyframes, resaves for the Go side, and renders the MID-FRAME (t=2s) so Go
// can assert the eased layer lags linear on actual pixels (red line 4).
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/mg_ease_args.json");
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
            if (it instanceof CompItem && it.name === "MGEASE") { comp = it; break; }
        }
        if (!comp) { fail("comp MGEASE not found"); }
        else {
            note("comp=" + comp.name + " layers=" + comp.numLayers);
            if (comp.numLayers !== 4) fail("expected 4 layers, got " + comp.numLayers);

            var byName = {};
            for (var li = 1; li <= comp.numLayers; li++) byName[comp.layer(li).name] = comp.layer(li);

            var eas = byName["EAS"];
            if (!eas) { fail("EAS missing"); }
            else {
                var pos = eas.transform.position;
                note("EAS numKeys=" + pos.numKeys);
                if (pos.numKeys !== 2) fail("EAS numKeys=" + pos.numKeys + ", want 2");
                var outEase = pos.keyOutTemporalEase(1)[0];
                var inEase = pos.keyInTemporalEase(2)[0];
                note("EAS kf1 outEase speed=" + outEase.speed + " influence=" + outEase.influence);
                note("EAS kf2 inEase speed=" + inEase.speed + " influence=" + inEase.influence);
                if (Math.abs(outEase.influence - 90) > 2) fail("kf1 out influence=" + outEase.influence + ", want ~90");
                if (pos.keyOutInterpolationType(1) !== KeyframeInterpolationType.BEZIER)
                    fail("kf1 out interp=" + pos.keyOutInterpolationType(1) + ", want BEZIER");
                if (pos.keyInInterpolationType(2) !== KeyframeInterpolationType.BEZIER)
                    fail("kf2 in interp=" + pos.keyInInterpolationType(2) + ", want BEZIER");
                // The eased value at t=2 must lag the linear midpoint (950).
                var v2 = pos.valueAtTime(2, false);
                note("EAS value@2s=[" + v2[0] + "," + v2[1] + "]");
                if (v2[0] > 800) fail("EAS x@2s=" + v2[0] + " — ease not applied (linear would be 950)");
            }

            var scl = byName["SCL"];
            if (!scl) { fail("SCL missing"); }
            else {
                note("SCL numKeys=" + scl.transform.position.numKeys);
                if (scl.transform.position.numKeys !== 6) fail("SCL numKeys=" + scl.transform.position.numKeys + ", want 6");
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
