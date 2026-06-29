// Acceptance verify for the from-scratch orbit animation demo. Reads
// orbit_demo_args.json {input, done, resaved}. Opens the Go-built file and
// proves AE ACCEPTS it (no hang — if AE hung, .done never appears and the
// ae_run wrapper times out), keeps all 5 layers, and that the dots' rotation
// animation is live (rotation differs across time). Resaves so the Go side can
// confirm the keyframes survive AE's own re-encode.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/orbit_demo_args.json");
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
            if (it instanceof CompItem && it.name === "ORBIT") { comp = it; break; }
        }
        if (!comp) { fail("comp ORBIT not found"); }
        else {
            note("comp=" + comp.name + " " + comp.width + "x" + comp.height +
                 " " + comp.frameRate + "fps " + comp.duration + "s layers=" + comp.numLayers);
            if (comp.numLayers !== 5) fail("expected 5 layers, got " + comp.numLayers);

            var byName = {};
            for (var li = 1; li <= comp.numLayers; li++) {
                var L = comp.layer(li);
                byName[L.name] = L;
                note("layer[" + li + "] " + L.name);
            }
            for (var n = 0; n < ["BG", "Ring", "Dot1", "Dot2", "Dot3"].length; n++) {
                var want = ["BG", "Ring", "Dot1", "Dot2", "Dot3"][n];
                if (!byName[want]) fail("missing layer " + want);
            }

            // Animation must be live: each dot's rotation differs across time.
            var dotNames = ["Dot1", "Dot2", "Dot3"];
            for (var k = 0; k < dotNames.length; k++) {
                var L = byName[dotNames[k]];
                if (!L) continue;
                var rot = L.property("Transform").property("Rotation");
                var r0 = rot.valueAtTime(0, false);
                var r3 = rot.valueAtTime(3, false);
                note(dotNames[k] + ".rot t0=" + r0 + " t3=" + r3);
                if (Math.abs(r3 - r0) < 1) fail(dotNames[k] + " rotation not animated");
            }
        }

        app.project.save(new File(args.resaved));

        // Render-pixel proof (delivery-contract red line 4): save frame 0 as
        // PNG for the Go side to sample. Force 8bpc first — at 32bpc float
        // saveFrameToPng emits linear-encoded (dark) pixels. $.sleep lets the
        // async write flush before app.quit can race it.
        if (comp && args.png) {
            app.project.bitsPerChannel = 8;
            comp.saveFrameToPng(0, new File(args.png));
            $.sleep(2000);
            var pngFile = new File(args.png);
            if (pngFile.exists) { note("frame png " + pngFile.length + " bytes"); }
            else { fail("saveFrameToPng wrote no file"); }
        }
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
