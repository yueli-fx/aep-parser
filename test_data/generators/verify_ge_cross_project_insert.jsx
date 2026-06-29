// V3 Phase 5C.1 — cross-Project InsertLayer ship-gate verify JSX.
//
// Opens a Go-emitted ge_cross_project_insert_<mode>.aep in AE and checks:
//   (a) AE opens the file without throwing (catch open errors → FAIL).
//   (b) "dest_comp" exists; dest_comp.layer(1) is the inserted clone;
//       dest_comp.layer(1).source is non-null.
//   (c) dest_comp.layer(1).source.name is a non-empty string
//       (source resolved to a real item — not a ghost/null).
//   (d) For file-backed sources (FootageItem with .file): source.footageMissing
//       is false (AE found the file on disk).
//   (e) app.project.numItems >= 2 (loose sanity: dest_comp + at least one
//       imported/own item exist).
//   (f) dedup mode only: exactly ONE footage item in the project whose .file
//       basename matches the dedup image filename ("image.png"), confirming the
//       cross-Project import did not create a duplicate footage entry.
//
// Inputs:
//   $.getenv("GE_XPROJ_MODE")   one of: footage | precomp | dedup
//   $.getenv("GE_XPROJ_TAG")    optional .done suffix for per-AE-version
//                                uniqueness (e.g. "ae2020_footage"); defaults
//                                to mode.
//
// Outputs:
//   test_data/generated/ship-gate/verify_ge_cross_project_insert_<tag>.done
//      first line:  mode + AE version
//      subsequent:  per-step OK / ERR lines
//      LAST line:   PASS   or   FAIL <reason>
(function () {
    var mode = $.getenv("GE_XPROJ_MODE");
    if (mode === null || mode === "") mode = "footage";

    var validModes = { footage: true, precomp: true, dedup: true };
    if (!validModes[mode]) {
        var errFile = new File("e:/projects/tools/aep-parser/test_data/generated/ship-gate/verify_ge_cross_project_insert_unknown.done");
        errFile.open("w");
        errFile.write("FAIL unknown mode: " + mode);
        errFile.close();
        return;
    }

    var tag = $.getenv("GE_XPROJ_TAG");
    if (tag === null || tag === "") tag = mode;

    var aepPath  = "e:/projects/tools/aep-parser/test_data/generated/ship-gate/ge_cross_project_insert_" + mode + ".aep";
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/generated/ship-gate/verify_ge_cross_project_insert_" + tag + ".done");
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
        step("inspect_project", function () {
            if (!app.project) throw new Error("app.project is null after open");

            // (e) loose sanity: at least the dest_comp + one other item
            log.push("  app.project.numItems=" + app.project.numItems);
            if (app.project.numItems < 2) {
                throw new Error("project has fewer than 2 items (numItems=" + app.project.numItems + "); expected dest_comp + at least one imported/own item");
            }

            // Locate dest_comp
            var destComp = null;
            for (var i = 1; i <= app.project.numItems; i++) {
                var it = app.project.items[i];
                if (it instanceof CompItem && it.name === "dest_comp") {
                    destComp = it;
                    break;
                }
            }
            if (destComp === null) throw new Error("comp 'dest_comp' not found in project");
            log.push("  dest_comp.numLayers=" + destComp.numLayers);

            // (b) layer(1) is the inserted clone; source is non-null
            if (destComp.numLayers < 1) throw new Error("dest_comp has no layers");
            var clone = destComp.layer(1);
            log.push("  clone.name=" + clone.name);

            var cloneSrc = null;
            try { cloneSrc = clone.source; } catch (e) {}
            if (cloneSrc === null || cloneSrc === undefined) {
                throw new Error("dest_comp.layer(1).source is null/undefined — InsertLayer source not resolved");
            }
            log.push("  clone.source.name=" + cloneSrc.name + " (id=" + cloneSrc.id + ")");

            // (c) source name is non-empty
            if (!cloneSrc.name || cloneSrc.name === "") {
                throw new Error("dest_comp.layer(1).source.name is empty — source item unresolved");
            }

            // (d) file-backed footage: footageMissing must be false
            // Guard: only check FootageItem instances that have a .file property.
            try {
                if (cloneSrc instanceof FootageItem && cloneSrc.file) {
                    log.push("  clone.source.footageMissing=" + cloneSrc.footageMissing);
                    if (cloneSrc.footageMissing) {
                        throw new Error("dest_comp.layer(1).source.footageMissing=true — AE cannot locate the source file on disk");
                    }
                }
            } catch (e) {
                // Re-throw only our own error; ignore AE probe errors on
                // layer kinds that don't support footageMissing.
                if (e.message && e.message.indexOf("footageMissing=true") !== -1) throw e;
            }

            // (f) dedup mode: exactly ONE footage item with basename "image.png"
            if (mode === "dedup") {
                var dedupBasename = "image.png";
                var matchCount = 0;
                for (var j = 1; j <= app.project.numItems; j++) {
                    var item = app.project.items[j];
                    try {
                        if (item instanceof FootageItem && item.file) {
                            var fname = item.file.name; // ExtendScript File.name = basename
                            if (fname === dedupBasename) {
                                matchCount++;
                            }
                        }
                    } catch (e) {}
                }
                log.push("  dedup: footage items matching '" + dedupBasename + "'=" + matchCount);
                if (matchCount !== 1) {
                    throw new Error("dedup mode: expected exactly 1 footage item for '" + dedupBasename + "', got " + matchCount + " — cross-Project import may have created a duplicate");
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
