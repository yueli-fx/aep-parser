// Acceptance verify for layer-set frame-blend setters (SetFrameBlendEnabled /
// SetFrameBlendPixelMotion). Reads frameblend_args.json {input, done}. Opens the
// input, finds the MAIN comp's two precomp layers, and DOM-reads frameBlendingType:
//   LMIX → FrameBlendingType.FRAME_MIX   (enabled, pixelMotion=false)
//   LPXM → FrameBlendingType.PIXEL_MOTION (enabled, pixelMotion=true)
// The FRAME_MIX vs PIXEL_MOTION contrast proves BOTH bits (enable + mode). NO
// resave (idta-batch lesson). Frame blending is valid on precomp/comp layers, so
// no external video file is needed.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/frameblend_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    function note(m) { log.push("  " + m); }
    function fbName(t) {
        if (t === FrameBlendingType.NO_FRAME_BLEND) return "NO_FRAME_BLEND";
        if (t === FrameBlendingType.FRAME_MIX) return "FRAME_MIX";
        if (t === FrameBlendingType.PIXEL_MOTION) return "PIXEL_MOTION";
        return "?(" + t + ")";
    }

    try {
        app.open(new File(args.input));

        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "MAIN") { comp = it; break; }
        }
        if (!comp) { fail("comp MAIN not found"); }
        else {
            var mix = comp.layer("LMIX");
            var pxm = comp.layer("LPXM");
            if (!mix || !pxm) { fail("LMIX/LPXM layer not found"); }
            else {
                var tm = mix.frameBlendingType, tp = pxm.frameBlendingType;
                note("LMIX frameBlendingType=" + fbName(tm) + " LPXM=" + fbName(tp));
                if (tm !== FrameBlendingType.FRAME_MIX) fail("LMIX=" + fbName(tm) + ", want FRAME_MIX");
                if (tp !== FrameBlendingType.PIXEL_MOTION) fail("LPXM=" + fbName(tp) + ", want PIXEL_MOTION");
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
