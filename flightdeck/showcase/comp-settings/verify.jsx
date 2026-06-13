// flightdeck/showcase/comp-settings/verify.jsx — open comp_settings.aep and dump
// the composition's settings DOM values to the .done log. 📋 READBACK showcase:
// no png, the user reads the values against the gen.go table. Run via ae_run.ps1.
(function () {
    var dir = "e:/projects/tools/aep-parser/flightdeck/showcase/comp-settings/";
    var inAep = new File(dir + "comp_settings.aep");
    var done = new File(dir + "comp_settings.done");
    var log = [];
    var ok = false;
    try {
        app.open(inAep);
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "CompSettings") { comp = it; break; }
        }
        if (!comp) throw new Error("comp not found");
        log.push("shutterAngle=" + comp.shutterAngle + " shutterPhase=" + comp.shutterPhase);
        log.push("frameRate=" + comp.frameRate + " duration=" + comp.duration);
        log.push("motionBlur=" + comp.motionBlur);
        log.push("motionBlurSamplesPerFrame=" + comp.motionBlurSamplesPerFrame);
        log.push("motionBlurAdaptiveSampleLimit=" + comp.motionBlurAdaptiveSampleLimit);
        log.push("bgColor=[" + fmt(comp.bgColor) + "]");
        log.push("workAreaStart=" + comp.workAreaStart + " workAreaDuration=" + comp.workAreaDuration);
        log.push("hideShyLayers=" + comp.hideShyLayers);
        log.push("preserveNestedFrameRate=" + comp.preserveNestedFrameRate);
        ok = true;
    } catch (e) { log.push("ERROR: " + e.toString() + " line=" + e.line); }
    done.open("w");
    done.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
    done.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}

    function fmt(a) {
        var s = [];
        for (var i = 0; i < a.length; i++) s.push(Math.round(a[i] * 1000) / 1000);
        return s.join(",");
    }
})();
