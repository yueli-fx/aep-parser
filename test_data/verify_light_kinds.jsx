(function () {
    var base = "e:/projects/tools/aep-parser/test_data/";
    var log = [];
    function want(name) {
        if (name === "L_parallel") return LightType.PARALLEL;
        if (name === "L_spot") return LightType.SPOT;
        if (name === "L_point") return LightType.POINT;
        if (name === "L_ambient") return LightType.AMBIENT;
        return null;
    }
    try {
        app.open(new File(base + "light_kinds_verify.aep"));
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            if (app.project.item(i) instanceof CompItem) { comp = app.project.item(i); break; }
        }
        if (!comp) throw new Error("no comp");
        for (var j = 1; j <= comp.numLayers; j++) {
            var ly = comp.layer(j);
            if (!(ly instanceof LightLayer)) continue;
            var got = ly.lightType;
            var w = want(ly.name);
            var ok = (got === w);
            log.push((ok ? "OK  " : "NO  ") + ly.name + " lightType=" + got + " want=" + w);
        }
    } catch (e) { log.push("ERR " + e.toString()); }

    var marker = new File(base + "verify_light_kinds.done");
    marker.open("w");
    marker.write(log.join("\n"));
    marker.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
