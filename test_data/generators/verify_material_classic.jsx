// Ship-gate verify for batch-5a classic Material Options setters. Reads
// material_classic_args.json {input, done, resaved}. Opens the Go-written file,
// finds the 3D solid, reads back the 8 classic material coefficients AE
// ingested via materialOption match-names, resaves so Go can confirm survival.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/material_classic_args.json");
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
        if (!comp) { fail("no CompItem"); }
        else {
            var L = null;
            for (var li = 1; li <= comp.numLayers; li++) {
                var cand = comp.layer(li);
                if (cand.threeDLayer) { L = cand; break; }
            }
            if (!L) { fail("no 3D layer found"); }
            else {
                log.push("  layer=" + L.name + " threeD=" + L.threeDLayer);
                var mo = L.materialOption;
                function near(label, mn, want) {
                    var pr = mo.property(mn);
                    if (!pr) { fail(label + " property missing"); return; }
                    var got = pr.value;
                    if (Math.abs(got - want) > 0.01) fail(label + "=" + got + " want " + want);
                    else log.push("  " + label + "=" + got);
                }
                near("lightTransmission", "ADBE Light Transmission", 0.55);
                near("acceptsShadows", "ADBE Accepts Shadows", 1);
                near("acceptsLights", "ADBE Accepts Lights", 1);
                near("ambient", "ADBE Ambient Coefficient", 0.66);
                near("diffuse", "ADBE Diffuse Coefficient", 0.44);
                near("specular", "ADBE Specular Coefficient", 0.77);
                near("shininess", "ADBE Shininess Coefficient", 60);
                near("metal", "ADBE Metal Coefficient", 0.88);
            }
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
