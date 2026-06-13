// flightdeck/showcase/keyframe-channels/render.jsx — open keyframe_channels.aep,
// render the MID time (t=2s) to keyframe_channels.png so every channel shows its
// INTERPOLATED (non-endpoint) value, and dump each layer's t=2s channel value to
// the .done log as a readback cross-check. Run via scripts/ae_run.ps1.
(function () {
    var dir = "e:/projects/tools/aep-parser/flightdeck/showcase/keyframe-channels/";
    var inAep = new File(dir + "keyframe_channels.aep");
    var png = new File(dir + "keyframe_channels.png");
    var done = new File(dir + "keyframe_channels.done");
    var log = [];
    var ok = false;
    try {
        app.open(inAep);
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "KeyframeChannels") { comp = it; break; }
        }
        if (!comp) throw new Error("comp not found");
        log.push("comp layers=" + comp.numLayers);
        // Readback each animated channel's value AT t=2s (the mid frame) so the
        // log proves AE computed the interpolated value, not an endpoint.
        for (var li = 1; li <= comp.numLayers; li++) {
            var lyr = comp.layer(li);
            var tr = lyr.property("ADBE Transform Group");
            if (!tr) continue;
            var parts = [lyr.name];
            parts.push("pos=" + fmt(tr.property("ADBE Position").valueAtTime(2.0, false)));
            parts.push("scale=" + fmt(tr.property("ADBE Scale").valueAtTime(2.0, false)));
            parts.push("rot=" + fmt(tr.property("ADBE Rotate Z").valueAtTime(2.0, false)));
            parts.push("opac=" + fmt(tr.property("ADBE Opacity").valueAtTime(2.0, false)));
            log.push(parts.join(" "));
        }
        app.project.bitsPerChannel = 8;
        comp.saveFrameToPng(2.0, png); // t=2s (mid of the 4s span)
        $.sleep(2500);
        log.push(png.exists ? ("png " + png.length + " bytes") : "PNG MISSING");
        ok = png.exists;
    } catch (e) { log.push("ERROR: " + e.toString() + " line=" + e.line); }
    done.open("w");
    done.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
    done.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}

    function fmt(v) {
        if (v instanceof Array) {
            var s = [];
            for (var i = 0; i < v.length; i++) s.push(Math.round(v[i] * 10) / 10);
            return "[" + s.join(",") + "]";
        }
        return "" + (Math.round(v * 10) / 10);
    }
})();
