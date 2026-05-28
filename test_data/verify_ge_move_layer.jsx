// V3 Phase 4 Task 3 — ship-gate verify JSX. Opens a Go-emitted
// ge_move_layer_<mode>.aep in AE and checks (a) AE doesn't throw
// "file data missing", (b) numLayers == 3 (preserved — no add/remove),
// (c) layer name order matches the expected mode permutation.
//
// Inputs:
//   $.getenv("GE_MOVE_MODE")    one of: first_to_last | last_to_first | mid_swap
//
// Outputs:
//   test_data/verify_ge_move_layer_<mode>.done
//      first line:  mode + AE version
//      subsequent lines: per-step OK / ERR + post-open layer dump
//      LAST line:   PASS or FAIL <reason>
(function () {
    var mode = $.getenv("GE_MOVE_MODE");
    if (mode === null || mode === "") mode = "first_to_last";

    // Expected post-move name order (AE 1-based, comp.layer(1) = top).
    // Baseline: [L1_top, L2_mid, L3_bot] (AE indices 1, 2, 3).
    var modeSpec = {
        first_to_last: { want: ["L2_mid", "L3_bot", "L1_top"] },
        last_to_first: { want: ["L3_bot", "L1_top", "L2_mid"] },
        mid_swap:      { want: ["L2_mid", "L1_top", "L3_bot"] }
    };
    var spec = modeSpec[mode];
    if (!spec) {
        var errFile = new File("e:/projects/tools/aep-parser/test_data/verify_ge_move_layer_unknown.done");
        errFile.open("w");
        errFile.write("FAIL unknown mode: " + mode);
        errFile.close();
        return;
    }

    var aepPath  = "e:/projects/tools/aep-parser/test_data/ge_move_layer_" + mode + ".aep";
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/verify_ge_move_layer_" + mode + ".done");
    var log = ["mode=" + mode];
    try { log.push("ae=" + app.version); } catch (e) {}

    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }

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
            throw e;
        }
    });

    if (openErr !== null) {
        failReason = "AE rejected file: " + openErr;
    } else {
        step("inspect_layers", function () {
            if (!app.project) throw new Error("app.project is null after open");
            var comp = null;
            for (var i = 1; i <= app.project.numItems; i++) {
                var it = app.project.items[i];
                if (it instanceof CompItem && it.name === "DeleteTest") {
                    comp = it;
                    break;
                }
            }
            if (comp === null) throw new Error("comp 'DeleteTest' not found in project");
            log.push("  comp.numLayers=" + comp.numLayers + " (want 3)");
            for (var j = 1; j <= comp.numLayers; j++) {
                var li = comp.layer(j);
                var info = "    [" + j + "] " + li.name;
                if (li.parent) info += " parent=" + li.parent.name;
                log.push(info);
            }
            if (comp.numLayers !== 3) {
                throw new Error("layer count mismatch: got " + comp.numLayers + ", want 3");
            }
            for (var k = 0; k < spec.want.length; k++) {
                var aeIdx = k + 1;
                var got = comp.layer(aeIdx).name;
                if (got !== spec.want[k]) {
                    throw new Error("layer @ AE idx " + aeIdx + ": got '" + got + "', want '" + spec.want[k] + "'");
                }
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
