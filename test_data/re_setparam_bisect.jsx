// Bisect probe for the SetEffectParam control-type AE reject: opens
// tmp_setparam_bisect.aep (built by tmp_debug/ge_setparam_bisect) and reads
// every layer's single materialized effect param in its own try/catch, so one
// run reports exactly which stream AE rejects (and with what error).
(function () {
    var inFile   = new File("e:/projects/tools/aep-parser/test_data/tmp_setparam_bisect.aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/re_setparam_bisect.done");
    var log = [];
    try { log.push("ae=" + app.version); } catch (e) {}

    try {
        app.open(inFile);
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem) { comp = it; break; }
        }
        if (!comp) { log.push("ERR no comp"); }
        for (var li = 1; li <= comp.numLayers; li++) {
            var layer = comp.layer(li);
            try {
                var parade = layer.property("ADBE Effect Parade");
                if (!parade || parade.numProperties === 0) {
                    log.push("LAYER " + layer.name + ": no effects");
                    continue;
                }
                for (var fi = 1; fi <= parade.numProperties; fi++) {
                    var fx = parade.property(fi);
                    var vals = [];
                    for (var pi = 1; pi <= fx.numProperties; pi++) {
                        var prop = fx.property(pi);
                        var entry = prop.matchName;
                        try {
                            if (prop.propertyType === PropertyType.PROPERTY) {
                                entry += "=" + prop.value.toString();
                            }
                        } catch (pe) {
                            entry += " VALUE-ERR " + pe.toString();
                        }
                        vals.push(entry);
                    }
                    log.push("LAYER " + layer.name + " fx=" + fx.matchName + " [" + vals.join(" | ") + "]");
                }
            } catch (le) {
                log.push("LAYER " + layer.name + " ERR " + le.toString());
            }
        }
    } catch (e2) {
        log.push("OPEN-ERR " + e2.toString());
    }

    doneFile.open("w");
    doneFile.write(log.join("\n"));
    doneFile.close();

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e3) {}
    try { app.quit(); } catch (e4) {}
})();
