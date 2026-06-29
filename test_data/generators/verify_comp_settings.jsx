// Ship-gate verify for batch-6 Composition setters. Reads comp_settings_args.json
// {input, done, resaved}. Opens the Go-written file and reads back the comp-level
// settings AE ingested across 5 from-scratch comps (CfgA scalars/bools+workarea,
// CfgB workarea start/end frame, CfgC workarea duration frame, CfgD displayStart
// frame, CfgE displayStart time). Resaves so Go can confirm survival.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/comp_settings_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL " + m); }
    function near(label, got, want, tol) {
        if (tol === undefined) tol = 0.01;
        if (Math.abs(got - want) > tol) fail(label + "=" + got + " want " + want);
        else log.push("  " + label + "=" + got);
    }
    function eqBool(label, got, want) {
        if (!!got !== !!want) fail(label + "=" + got + " want " + want);
        else log.push("  " + label + "=" + got);
    }

    try {
        app.open(new File(args.input));
        var byName = {};
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem) byName[it.name] = it;
        }
        function need(n) { if (!byName[n]) { fail("comp " + n + " missing"); return null; } return byName[n]; }

        var A = need("CfgA");
        if (A) {
            near("A.width", A.width, 1600); near("A.height", A.height, 900);
            near("A.duration", A.duration, 8, 0.02);
            near("A.pixelAspect", A.pixelAspect, 2.0, 0.01);
            near("A.resX", A.resolutionFactor[0], 2); near("A.resY", A.resolutionFactor[1], 2);
            near("A.bgR", A.bgColor[0], 1.0, 0.02); near("A.bgG", A.bgColor[1], 0.502, 0.02); near("A.bgB", A.bgColor[2], 0.0, 0.02);
            near("A.shutterAngle", A.shutterAngle, 360);
            near("A.shutterPhase", A.shutterPhase, -90);
            near("A.mbSamples", A.motionBlurSamplesPerFrame, 32);
            near("A.mbAdaptive", A.motionBlurAdaptiveSampleLimit, 256);
            eqBool("A.motionBlur", A.motionBlur, true);
            eqBool("A.frameBlending", A.frameBlending, true);
            eqBool("A.hideShy", A.hideShyLayers, true);
            eqBool("A.preserveFR", A.preserveNestedFrameRate, true);
            eqBool("A.preserveRes", A.preserveNestedResolution, true);
            near("A.workStart", A.workAreaStart, 2.0, 0.02);
            near("A.workDur", A.workAreaDuration, 4.0, 0.02);
        }
        var B = need("CfgB");
        if (B) {
            near("B.workStart", B.workAreaStart, 1.0, 0.02);
            near("B.workDur", B.workAreaDuration, 3.0, 0.02);
        }
        var C = need("CfgC");
        if (C) {
            near("C.workDur", C.workAreaDuration, 1.5, 0.02);
        }
        var D = need("CfgD");
        if (D) {
            near("D.displayStartTime", D.displayStartTime, 0.5, 0.02);
            near("D.displayStartFrame", D.displayStartFrame, 15);
        }
        var E = need("CfgE");
        if (E) {
            near("E.displayStartTime", E.displayStartTime, 2.0, 0.02);
        }
        var F = need("CfgF");
        if (F) {
            near("F.frameRate", F.frameRate, 24);
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
