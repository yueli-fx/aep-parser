// flightdeck/showcase/camera-light/verify.jsx — open camera_light.aep and dump the
// camera / light layer types + default option values to .done. 📋 READBACK
// showcase: no png. Confirms NewCameraLayer / NewLightLayer produce AE-accepted,
// correctly-typed layers. Run via scripts/ae-worker/ae_run.ps1.
(function () {
    var dir = "e:/projects/tools/aep-parser/flightdeck/showcase/camera-light/";
    var inAep = new File(dir + "camera_light.aep");
    var done = new File(dir + "camera_light.done");
    var log = [];
    var ok = false;
    try {
        app.open(inAep);
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "CameraLight") { comp = it; break; }
        }
        if (!comp) throw new Error("comp not found");
        log.push("comp layers=" + comp.numLayers);

        var cam = comp.layer("Cam01");
        log.push("Cam01 instanceof CameraLayer=" + (cam instanceof CameraLayer));
        log.push("Cam01 zoom=" + readVal(cam, "ADBE Camera Options Group", "ADBE Camera Zoom"));
        log.push("Cam01 depthOfField=" + readVal(cam, "ADBE Camera Options Group", "ADBE Camera Depth of Field"));
        log.push("Cam01 focusDistance=" + readVal(cam, "ADBE Camera Options Group", "ADBE Camera Focus Distance"));
        log.push("Cam01 aperture=" + readVal(cam, "ADBE Camera Options Group", "ADBE Camera Aperture"));
        log.push("Cam01 blurLevel=" + readVal(cam, "ADBE Camera Options Group", "ADBE Camera Blur Level"));

        var light = comp.layer("Light01");
        log.push("Light01 instanceof LightLayer=" + (light instanceof LightLayer));
        log.push("Light01 lightType=" + light.lightType + " (" + lightTypeName(light.lightType) + ")");
        log.push("Light01 intensity=" + readVal(light, "ADBE Light Options Group", "ADBE Light Intensity"));

        ok = (cam instanceof CameraLayer) && (light instanceof LightLayer) && comp.numLayers === 3;
    } catch (e) { log.push("ERROR: " + e.toString() + " line=" + e.line); }
    done.open("w");
    done.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
    done.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}

    function lightTypeName(v) {
        if (v === LightType.PARALLEL) return "PARALLEL";
        if (v === LightType.SPOT) return "SPOT";
        if (v === LightType.POINT) return "POINT";
        if (v === LightType.AMBIENT) return "AMBIENT";
        return "?";
    }
    function readVal(lyr, grpMatch, propMatch) {
        try {
            var g = lyr.property(grpMatch);
            if (!g) return "(no group)";
            var pr = g.property(propMatch);
            if (!pr) return "(no prop)";
            return "" + pr.value;
        } catch (e) { return "(err)"; }
    }
})();
