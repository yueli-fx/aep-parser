// test_data/generators/probe_flame_params.jsx — dump param matchName↔name for flame effects.
// Go-side fx.Parameters only mirrors numeric matchNames; AE resolves human names.
(function () {
    var dir = "e:/projects/tools/aep-parser/test_data/";
    var log = [];
    function dump(effMN) {
        var comp = app.project.items.addComp("P", 640, 360, 1, 3, 30);
        var sol = comp.layers.addSolid([0, 0, 0], "S", 640, 360, 1);
        var fx;
        try { fx = sol.property("ADBE Effect Parade").addProperty(effMN); }
        catch (e) { log.push(effMN + ": ADD FAIL " + e.toString()); return; }
        log.push("=== " + effMN + " (" + fx.numProperties + " params) ===");
        for (var i = 1; i <= fx.numProperties; i++) {
            var p = fx.property(i);
            log.push("  " + p.matchName + "  ::  " + p.name);
        }
        comp.remove();
    }
    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); app.newProject();
        dump("ADBE Fractal Noise");
        dump("ADBE Turbulent Displace");
        dump("ADBE Tint");
        dump("ADBE Tritone");
        dump("ADBE Glo2");
        dump("ADBE Ramp");
        dump("ADBE 4ColorGradient");
    } catch (e) { log.push("EXC " + e.toString() + " line=" + e.line); }
    var m = new File(dir + "probe_flame_params.done"); m.open("w"); m.write(log.join("\n")); m.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
