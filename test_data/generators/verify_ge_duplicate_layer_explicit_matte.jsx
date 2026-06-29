// V3 Phase 5B ship-gate verify JSX. Opens
// ge_duplicate_layer_explicit_matte.aep in AE 2025 and confirms:
//   (a) AE doesn't reject the file as corrupt
//   (b) The RE_TRACKMATTE comp has 8 layers (was 7 pre-dup, +1 clone)
//   (c) Layer "mt_alpha_clone" exists and its trackMatteLayer is the
//       same source layer as the original "mt_alpha_to_solidA"'s
//       trackMatteLayer (both should point to solidA)
//   (d) Original "mt_alpha_to_solidA" still has its matte intact
//
// Inputs: none (single mode, no env var)
// Outputs:
//   test_data/generated/ship-gate/verify_ge_duplicate_layer_explicit_matte.done
//      first line:  ae=<version>
//      subsequent lines: per-step OK / ERR + post-open layer dump
//      LAST line:   PASS or FAIL <reason>
(function () {
    var aepPath  = "e:/projects/tools/aep-parser/test_data/generated/ship-gate/ge_duplicate_layer_explicit_matte.aep";
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/generated/ship-gate/verify_ge_duplicate_layer_explicit_matte.done");
    var log = [];
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
                if (it instanceof CompItem && it.name === "RE_TRACKMATTE") {
                    comp = it;
                    break;
                }
            }
            if (comp === null) throw new Error("comp 'RE_TRACKMATTE' not found in project");

            log.push("  comp.numLayers=" + comp.numLayers + " (want 8)");
            for (var j = 1; j <= comp.numLayers; j++) {
                var li = comp.layer(j);
                var info = "    [" + j + "] name='" + li.name + "'";
                try {
                    if (li.source && li.source.name) info += " src='" + li.source.name + "'";
                } catch (e) {}
                try {
                    if (li.trackMatteLayer) info += " matteSrc='" + li.trackMatteLayer.name + "'";
                } catch (e) {}
                if (li.trackMatteType && li.trackMatteType !== TrackMatteType.NO_TRACK_MATTE) {
                    info += " trackMatteType=" + li.trackMatteType;
                }
                log.push(info);
            }

            if (comp.numLayers !== 8) {
                throw new Error("layer count mismatch: got " + comp.numLayers + ", want 8");
            }

            // Find clone and original by name. Clone has Utf8 override set
            // by Go; original is unnamed solid → AE shows its source name.
            var clone = null, original = null;
            for (var k = 1; k <= comp.numLayers; k++) {
                var lk = comp.layer(k);
                if (lk.name === "mt_alpha_clone") clone = lk;
                else if (lk.name === "mt_alpha_to_solidA") original = lk;
            }
            if (clone === null) throw new Error("clone layer 'mt_alpha_clone' not found");
            if (original === null) throw new Error("original layer 'mt_alpha_to_solidA' not found");

            // Both must have explicit trackMatteLayer set.
            var cloneMatte = null, origMatte = null;
            try { cloneMatte = clone.trackMatteLayer; } catch (e) {}
            try { origMatte = original.trackMatteLayer; } catch (e) {}
            if (cloneMatte === null) {
                throw new Error("clone.trackMatteLayer is null/undefined — explicit matte lost on dup");
            }
            if (origMatte === null) {
                throw new Error("original.trackMatteLayer is null/undefined — dup broke original's matte");
            }
            // They must point to the SAME source layer (identity by AE
            // index — AE returns the actual layer object).
            if (cloneMatte.index !== origMatte.index) {
                throw new Error("clone matte source (AE idx " + cloneMatte.index + ") != original matte source (AE idx " + origMatte.index + ") — both should point to same layer");
            }
            // The shared matte source should be "solidA" (or whatever
            // displays as solidA — name match via .name fallback).
            if (cloneMatte.name !== "solidA") {
                throw new Error("matte source layer name: got '" + cloneMatte.name + "', want 'solidA'");
            }

            // Both should retain TrackMatteType.ALPHA (= 5012 in AE 23+).
            if (!clone.trackMatteType || clone.trackMatteType === TrackMatteType.NO_TRACK_MATTE) {
                throw new Error("clone.trackMatteType is NO_TRACK_MATTE — F8 byte-copy failed");
            }
            if (clone.trackMatteType !== original.trackMatteType) {
                throw new Error("clone.trackMatteType=" + clone.trackMatteType + " != original.trackMatteType=" + original.trackMatteType);
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
