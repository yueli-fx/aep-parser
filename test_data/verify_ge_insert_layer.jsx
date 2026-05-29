// V3 Phase 5C — InsertLayer ship-gate verify JSX. Opens a Go-emitted
// ge_insert_layer_<mode>.aep in AE and checks:
//   (a) AE doesn't throw a "file data missing" / corruption signal on open,
//   (b) compB_dest.numLayers == expected (grew by 1 from the clone),
//   (c) the inserted clone (AE 1-based index 1, top of compB) has NO parent
//       and NO track matte — cross-comp clone resets both,
//   (d) the clone references the SAME source item as compA_src_holder's src
//       layer (SourceID preserved verbatim — shared Footage/Comp item).
//
// Inputs:
//   $.getenv("GE_INSERT_MODE")   one of: basic | footage | precomp
//   $.getenv("GE_INSERT_TAG")    optional .done filename suffix so the same
//                                mode run on two AE versions writes distinct
//                                files (e.g. "ae2020_basic"); defaults to mode.
//
// Outputs:
//   test_data/verify_ge_insert_layer_<tag>.done   (tag defaults to mode)
//      first line:  mode + AE version
//      subsequent lines: per-step OK / ERR + post-open layer dump
//      LAST line:   PASS or FAIL <reason>
(function () {
    var mode = $.getenv("GE_INSERT_MODE");
    if (mode === null || mode === "") mode = "basic";

    // compB had: basic=1 (B_unrelated); footage/precomp=2 (B_unrelated + a
    // disabled extra referencing the shared asset). Insert adds 1 at the top.
    var modeSpec = {
        basic:   { wantLayers: 2 },
        footage: { wantLayers: 3 },
        precomp: { wantLayers: 3 }
    };
    var spec = modeSpec[mode];
    if (!spec) {
        var errFile = new File("e:/projects/tools/aep-parser/test_data/verify_ge_insert_layer_unknown.done");
        errFile.open("w");
        errFile.write("FAIL unknown mode: " + mode);
        errFile.close();
        return;
    }

    var tag = $.getenv("GE_INSERT_TAG");
    if (tag === null || tag === "") tag = mode;
    var aepPath  = "e:/projects/tools/aep-parser/test_data/ge_insert_layer_" + mode + ".aep";
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/verify_ge_insert_layer_" + tag + ".done");
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
            var compA = null, compB = null;
            for (var i = 1; i <= app.project.numItems; i++) {
                var it = app.project.items[i];
                if (it instanceof CompItem) {
                    if (it.name === "compA_src_holder") compA = it;
                    else if (it.name === "compB_dest") compB = it;
                }
            }
            if (compA === null) throw new Error("comp 'compA_src_holder' not found");
            if (compB === null) throw new Error("comp 'compB_dest' not found");

            log.push("  compB.numLayers=" + compB.numLayers + " (want " + spec.wantLayers + ")");
            for (var j = 1; j <= compB.numLayers; j++) {
                var li = compB.layer(j);
                var info = "    [" + j + "] " + li.name;
                if (li.parent) info += " parent=" + li.parent.name;
                try { if (li.source) info += " source=" + li.source.name + "(id" + li.source.id + ")"; } catch (e) {}
                try {
                    if (li.trackMatteType && li.trackMatteType !== TrackMatteType.NO_TRACK_MATTE) {
                        info += " trackMatteType=" + li.trackMatteType;
                    }
                } catch (e) {}
                log.push(info);
            }
            if (compB.numLayers !== spec.wantLayers) {
                throw new Error("layer count mismatch: got " + compB.numLayers + ", want " + spec.wantLayers);
            }

            // The clone is inserted at the top (AE 1-based index 1).
            var clone = compB.layer(1);

            // (c) cross-comp clone: no parent, no track matte.
            if (clone.parent) {
                throw new Error("clone @ idx 1: got parent '" + clone.parent.name + "', want no parent (cross-comp reset)");
            }
            try {
                if (clone.trackMatteType && clone.trackMatteType !== TrackMatteType.NO_TRACK_MATTE) {
                    throw new Error("clone @ idx 1: trackMatteType=" + clone.trackMatteType + ", want NO_TRACK_MATTE (cross-comp reset)");
                }
            } catch (e) {
                if (e.message && e.message.indexOf("want NO_TRACK_MATTE") !== -1) throw e;
                // older AE may lack trackMatteType on some layer kinds — ignore probe error
            }

            // (d) source preserved verbatim — same item as compA's src layer.
            if (compA.numLayers < 1) throw new Error("compA has no src layer");
            var srcLayer = compA.layer(1);
            var wantSrc = null, gotSrc = null;
            try { wantSrc = srcLayer.source; } catch (e) {}
            try { gotSrc = clone.source; } catch (e) {}
            if (wantSrc === null) throw new Error("compA src layer has no source");
            if (gotSrc === null) throw new Error("clone has no source");
            if (gotSrc.id !== wantSrc.id) {
                throw new Error("clone.source id=" + gotSrc.id + " ('" + gotSrc.name + "'), want id=" + wantSrc.id + " ('" + wantSrc.name + "') — SourceID must be verbatim");
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
