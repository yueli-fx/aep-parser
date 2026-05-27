// V3 Phase 2 Task 5 — ship-gate verify JSX. Opens a Go-emitted
// ge_delete_layer_<mode>.aep in AE and checks (a) AE doesn't throw
// "file data missing" / similar corruption signal, (b) the surviving
// layer count matches Go-side expectation.
//
// Inputs:
//   $.getenv("GE_DELETE_MODE")    one of: baseline | middle | parent | matte
//
// Outputs:
//   test_data/verify_ge_delete_layer_<mode>.done
//      first line: mode + AE version
//      subsequent lines: per-step OK / ERR
//      LAST line: PASS or FAIL <reason>   (used by Go-side / shell consumer)
(function () {
    var mode = $.getenv("GE_DELETE_MODE");
    if (mode === null || mode === "") mode = "baseline";

    var expectedLayerCount = {
        baseline:   3,
        middle:     2,
        parent:     2,
        matte_ae20: 2,
        matte_ae25: 2
    };
    var expected = expectedLayerCount[mode];
    if (typeof expected !== "number") {
        var errFile = new File("e:/projects/tools/aep-parser/test_data/verify_ge_delete_layer_unknown.done");
        errFile.open("w");
        errFile.write("FAIL unknown mode: " + mode);
        errFile.close();
        return;
    }

    var aepPath  = "e:/projects/tools/aep-parser/test_data/ge_delete_layer_" + mode + ".aep";
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/verify_ge_delete_layer_" + mode + ".done");
    var log = ["mode=" + mode];
    try { log.push("ae=" + app.version); } catch (e) {}

    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }

    // Result tracking.
    var passed = false;
    var failReason = "";

    step("close_existing_project", function () {
        try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    });

    var openErr = null;
    step("open_ge_file", function () {
        var f = new File(aepPath);
        if (!f.exists) throw new Error("file does not exist: " + aepPath);
        try {
            app.open(f);
        } catch (e) {
            openErr = e.toString();
            throw e; // re-throw so step logs ERR; we'll handle below
        }
    });

    if (openErr !== null) {
        failReason = "AE rejected file: " + openErr;
    } else {
        step("inspect_layers", function () {
            if (!app.project) throw new Error("app.project is null after open");
            // Find the DeleteTest comp.
            var comp = null;
            for (var i = 1; i <= app.project.numItems; i++) {
                var it = app.project.items[i];
                if (it instanceof CompItem && it.name === "DeleteTest") {
                    comp = it;
                    break;
                }
            }
            if (comp === null) throw new Error("comp 'DeleteTest' not found in project");
            log.push("  comp.numLayers=" + comp.numLayers + " (want " + expected + ")");
            for (var j = 1; j <= comp.numLayers; j++) {
                var li = comp.layer(j);
                var info = "    [" + j + "] " + li.name + " id=" + li.id;
                if (li.parent) info += " parent=" + li.parent.name;
                try { if (li.trackMatteLayer) info += " matteSrc=" + li.trackMatteLayer.name; } catch (e) {}
                if (li.trackMatteType && li.trackMatteType !== TrackMatteType.NO_TRACK_MATTE) {
                    info += " trackMatteType=" + li.trackMatteType;
                }
                log.push(info);
            }
            if (comp.numLayers !== expected) {
                throw new Error("layer count mismatch: got " + comp.numLayers + ", want " + expected);
            }
            passed = true;
        });
    }

    if (passed) {
        log.push("PASS");
    } else {
        log.push("FAIL " + (failReason || "see ERR lines above"));
    }

    try {
        doneFile.open("w");
        doneFile.write(log.join("\n"));
        doneFile.close();
    } catch (e) {}

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
