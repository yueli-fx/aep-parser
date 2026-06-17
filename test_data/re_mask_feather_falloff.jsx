// RE probe for maskFeatherFalloff (feather decay curve enum). Granular try/catch
// per step so we learn exactly what AE 2020 scripting exposes: is the global
// MaskFeatherFalloff enum defined? does the property assignment take? does the
// readback differ between the two masks? Then a Go byte-diff of the two atoms in
// the saved file pinpoints where (if anywhere) falloff is stored.
(function () {
    var dir = "e:/projects/tools/aep-parser/test_data/";
    var outFile = new File(dir + "re_mask_feather_falloff.aep");
    var log = [];
    var ok = true;
    function step(label, fn) {
        try { var r = fn(); log.push("OK  " + label + (r !== undefined ? " => " + r : "")); return r; }
        catch (e) { log.push("ERR " + label + " -> " + e.toString()); ok = false; return undefined; }
    }
    function rectShape(x, y) {
        var s = new Shape();
        s.vertices = [[x, y], [x + 200, y], [x + 200, y + 200], [x, y + 200]];
        s.closed = true;
        return s;
    }
    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        log.push("typeof MaskFeatherFalloff = " + (typeof MaskFeatherFalloff));
        if (typeof MaskFeatherFalloff !== "undefined") {
            log.push("FFO_SMOOTH=" + MaskFeatherFalloff.FFO_SMOOTH + " FFO_LINEAR=" + MaskFeatherFalloff.FFO_LINEAR);
        }
        var comp = app.project.items.addComp("FFO", 1920, 1080, 1, 5, 30);
        var solid = comp.layers.addSolid([0.5, 0.5, 0.5], "S", 1920, 1080, 1);
        var mp = solid.property("ADBE Mask Parade");

        var m1 = mp.addProperty("ADBE Mask Atom");
        m1.name = "LINEAR";
        step("m1 setShape", function () { m1.property("ADBE Mask Shape").setValue(rectShape(100, 100)); return "ok"; });
        step("m1 read default falloff", function () { return m1.maskFeatherFalloff; });
        step("m1 set LINEAR", function () { m1.maskFeatherFalloff = MaskFeatherFalloff.FFO_LINEAR; return "ok"; });
        step("m1 read after set", function () { return m1.maskFeatherFalloff; });

        var m2 = mp.addProperty("ADBE Mask Atom");
        m2.name = "SMOOTH";
        step("m2 setShape", function () { m2.property("ADBE Mask Shape").setValue(rectShape(400, 400)); return "ok"; });
        step("m2 leave SMOOTH (default)", function () { return m2.maskFeatherFalloff; });

        step("compare", function () {
            return "m1=" + m1.maskFeatherFalloff + " m2=" + m2.maskFeatherFalloff + " differ=" + (m1.maskFeatherFalloff !== m2.maskFeatherFalloff);
        });

        app.project.save(outFile);
        log.push("saved " + outFile.fsName);
    } catch (e) {
        log.push("EXC " + e.toString() + " line=" + e.line);
        ok = false;
    }
    var marker = new File(dir + "re_mask_feather_falloff.done");
    marker.open("w");
    marker.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n"));
    marker.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
