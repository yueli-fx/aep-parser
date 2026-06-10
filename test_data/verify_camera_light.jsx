// Ship-gate verify for NewCameraLayer / NewLightLayer. Reads
// camera_light_args.json {input, done, resaved}. Opens the Go-written file,
// confirms AE accepts it and that the comp contains a CameraLayer "Cam1" and a
// LightLayer "Light1", resaves so the Go side can confirm AE kept them.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/camera_light_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  " + m); }

    try {
        app.open(new File(args.input));

        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem) { comp = it; break; }
        }
        if (!comp) {
            fail("no CompItem in project");
        } else {
            var cam = null, light = null, names = [];
            for (var li = 1; li <= comp.numLayers; li++) {
                var L = comp.layer(li);
                var kind = "av";
                if (L instanceof CameraLayer) { kind = "cam"; cam = L; }
                else if (L instanceof LightLayer) { kind = "light"; light = L; }
                names.push(L.name + "(" + kind + ")");
            }
            log.push("  layers=[" + names.join(", ") + "]");
            if (!cam) fail("no CameraLayer found");
            else if (cam.name !== "Cam1") fail("camera name " + cam.name + " != Cam1");
            if (!light) fail("no LightLayer found");
            else if (light.name !== "Light1") fail("light name " + light.name + " != Light1");
        }

        app.project.save(new File(args.resaved));
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
