// V3 Phase 3 Task 5 — ship-gate verify JSX. Opens a Go-emitted
// ge_duplicate_layer_<mode>.aep in AE and checks (a) AE doesn't throw
// "file data missing" / similar corruption signal, (b) numLayers == 4,
// (c) clone is at the expected AE 1-based index with the supplied name,
// (d) parent-ref expectations per mode (F6 + F7 from the RE scar).
//
// Inputs:
//   $.getenv("GE_DUP_MODE")    one of: solo | dup_parent | dup_child
//
// Outputs:
//   test_data/verify_ge_duplicate_layer_<mode>.done
//      first line:  mode + AE version
//      subsequent lines: per-step OK / ERR + post-open layer dump
//      LAST line:   PASS or FAIL <reason>   (consumed by Go side / shell)
(function () {
    var mode = $.getenv("GE_DUP_MODE");
    if (mode === null || mode === "") mode = "solo";

    // Per-mode expectations (Go producer pushes clone in at source idx;
    // source moves to source+1; counted in AE 1-based indices below).
    //   cloneAeIdx  — where the clone should land after open
    //   cloneName   — caller-supplied name from Go producer
    //   srcAeIdx    — where the original source ended up
    //   srcName     — original source's name (from RE baseline)
    //   cloneParent — expected clone's parent name, or null if no parent
    //   l3ParentChk — for dup_parent: verify L3 still parents to L2_mid
    //                 (F6: child's outgoing ParentID NOT rewritten to clone)
    var modeSpec = {
        solo: {
            cloneAeIdx:  2, cloneName: "L2_clone",
            srcAeIdx:    3, srcName:   "L2_mid",
            cloneParent: null,
            l3ParentChk: null
        },
        dup_parent: {
            cloneAeIdx:  2, cloneName: "L2_clone",
            srcAeIdx:    3, srcName:   "L2_mid",
            cloneParent: null,                  // L2 itself has no parent
            l3ParentChk: { layerAeIdx: 4, parentName: "L2_mid" }
        },
        dup_child: {
            cloneAeIdx:  1, cloneName: "L1_clone",
            srcAeIdx:    2, srcName:   "L1_top",
            cloneParent: "L2_mid",              // F7: clone inherits ParentID
            l3ParentChk: null
        }
    };
    var spec = modeSpec[mode];
    if (!spec) {
        var errFile = new File("e:/projects/tools/aep-parser/test_data/verify_ge_duplicate_layer_unknown.done");
        errFile.open("w");
        errFile.write("FAIL unknown mode: " + mode);
        errFile.close();
        return;
    }

    var aepPath  = "e:/projects/tools/aep-parser/test_data/ge_duplicate_layer_" + mode + ".aep";
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/verify_ge_duplicate_layer_" + mode + ".done");
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
            // Locate the comp (re_delete_layer baseline used "DeleteTest").
            var comp = null;
            for (var i = 1; i <= app.project.numItems; i++) {
                var it = app.project.items[i];
                if (it instanceof CompItem && it.name === "DeleteTest") {
                    comp = it;
                    break;
                }
            }
            if (comp === null) throw new Error("comp 'DeleteTest' not found in project");
            log.push("  comp.numLayers=" + comp.numLayers + " (want 4)");
            for (var j = 1; j <= comp.numLayers; j++) {
                var li = comp.layer(j);
                var info = "    [" + j + "] " + li.name;
                if (li.parent) info += " parent=" + li.parent.name;
                try { if (li.trackMatteLayer) info += " matteSrc=" + li.trackMatteLayer.name; } catch (e) {}
                if (li.trackMatteType && li.trackMatteType !== TrackMatteType.NO_TRACK_MATTE) {
                    info += " trackMatteType=" + li.trackMatteType;
                }
                log.push(info);
            }
            if (comp.numLayers !== 4) {
                throw new Error("layer count mismatch: got " + comp.numLayers + ", want 4");
            }

            // Clone identity check.
            var clone = comp.layer(spec.cloneAeIdx);
            if (clone.name !== spec.cloneName) {
                throw new Error("clone @ AE idx " + spec.cloneAeIdx + ": got name '" + clone.name + "', want '" + spec.cloneName + "'");
            }
            if (spec.cloneParent === null) {
                if (clone.parent) {
                    throw new Error("clone @ AE idx " + spec.cloneAeIdx + ": got parent '" + clone.parent.name + "', want no parent");
                }
            } else {
                if (!clone.parent) {
                    throw new Error("clone @ AE idx " + spec.cloneAeIdx + ": got no parent, want '" + spec.cloneParent + "'");
                }
                if (clone.parent.name !== spec.cloneParent) {
                    throw new Error("clone @ AE idx " + spec.cloneAeIdx + ": got parent '" + clone.parent.name + "', want '" + spec.cloneParent + "'");
                }
            }

            // Source identity check (pushed down by 1).
            var src = comp.layer(spec.srcAeIdx);
            if (src.name !== spec.srcName) {
                throw new Error("source @ AE idx " + spec.srcAeIdx + ": got name '" + src.name + "', want '" + spec.srcName + "'");
            }

            // F6 — only checked for dup_parent.
            if (spec.l3ParentChk) {
                var l3 = comp.layer(spec.l3ParentChk.layerAeIdx);
                if (!l3.parent) {
                    throw new Error("L3 @ AE idx " + spec.l3ParentChk.layerAeIdx + ": got no parent, want '" + spec.l3ParentChk.parentName + "' (F6: child's ParentID NOT rewritten to clone)");
                }
                if (l3.parent.name !== spec.l3ParentChk.parentName) {
                    throw new Error("L3 @ AE idx " + spec.l3ParentChk.layerAeIdx + ": got parent '" + l3.parent.name + "', want '" + spec.l3ParentChk.parentName + "' (F6)");
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
