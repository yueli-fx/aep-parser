// Acceptance verify for layer-set audio setters (SetAudioEnabled / SetAudioLevels).
// Reads audio_args.json {input, done}. Opens the input, finds the AUDLV comp's
// audio layer, and DOM-reads audioEnabled + Audio Levels value. NO resave (the
// idta-batch lesson: a verify resave pops AE's Save dialog / pollutes safe-mode
// state; pure open+DOM-readback is the ae-accept proof for a non-visual domain).
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/audio_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    function note(m) { log.push("  " + m); }
    function near(a, b, tol) { return Math.abs(a - b) <= tol; }

    try {
        app.open(new File(args.input));

        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "AUDLV") { comp = it; break; }
        }
        if (!comp) { fail("comp AUDLV not found"); }
        else {
            var L = comp.layer(1);
            if (!L) { fail("audio layer not found"); }
            else {
                var ae = L.audioEnabled;
                var lv = L.property("ADBE Audio Group").property("ADBE Audio Levels").value;
                note("audioEnabled=" + ae + " levels=[" + lv[0] + "," + lv[1] + "]");
                if (ae !== false) fail("audioEnabled=" + ae + ", want false");
                if (!near(lv[0], -8, 0.05) || !near(lv[1], -8, 0.05)) fail("levels=[" + lv[0] + "," + lv[1] + "], want [-8,-8]");
            }
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
