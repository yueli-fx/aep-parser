// Visible-3D ship gate: prove a from-scratch 3D layer (z=0, enabled via the
// Is3D bit) genuinely responds to a camera. Reads 3d_cam_dolly_args.json
// {input, done, pngNear, pngFar}. The input has a 3D BOX (z=0) + a Go-created
// camera. We dolly the camera Z between two renders; Go measures the box width
// — a genuine 3D layer scales with camera distance (near=big, far=small), a 2D
// layer would render at a constant size and ignore the camera entirely.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/3d_cam_dolly_args.json");
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
            if (it instanceof CompItem && it.name === "CAM3D") { comp = it; break; }
        }
        if (!comp) { fail("comp CAM3D not found"); }
        else {
            note("layers=" + comp.numLayers);
            var cam = null, box = null;
            for (var li = 1; li <= comp.numLayers; li++) {
                var L = comp.layer(li);
                if (L instanceof CameraLayer) cam = L;
                if (L.name === "BOX") box = L;
            }
            if (!cam) fail("camera missing");
            if (!box) fail("BOX missing");
            else note("BOX threeDLayer=" + box.threeDLayer);
            if (cam) {
                var cp = cam.property("ADBE Transform Group").property("ADBE Position");
                note("camera default pos=" + cp.value.toString());
                app.project.bitsPerChannel = 8;
                // NEAR: camera close to z=0 plane.
                cp.setValue([960, 540, -700]);
                comp.saveFrameToPng(0, new File(args.pngNear));
                // FAR: camera pulled far back.
                cp.setValue([960, 540, -2400]);
                comp.saveFrameToPng(0, new File(args.pngFar));
                $.sleep(1500);
                note("pngNear " + new File(args.pngNear).length + "B  pngFar " + new File(args.pngFar).length + "B");
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
