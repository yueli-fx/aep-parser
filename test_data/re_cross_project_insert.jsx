// RE fixture builder for V3 Phase 5C.1 — cross-Project InsertLayer ship-gate.
//
// Produces TWO separate .aep files per mode (src + dest), where the src file
// holds "src_comp" with at least one AV layer and the dest file holds
// "dest_comp" with an unrelated layer (and, for dedup mode, a matching
// file-backed footage item).
//
// Mode via $.getenv("RE_XPROJ_MODE") in { footage | precomp | dedup }.
//
// NOTE on file-backed footage path:
//   Modes "footage" and "dedup" import a real image file as footage.
//   Path used: e:/projects/tools/aep-parser/image.png
//   (a 3 KB PNG that lives in the project root, gitignored via /image.png).
//   If the controller needs to substitute a different path, change imagePath
//   at the top of this script.
//
// Outputs:
//   test_data/re_xproj_src_<mode>.aep
//   test_data/re_xproj_dest_<mode>.aep
//   test_data/re_xproj_<mode>.done  (last line: PASS or FAIL <reason>)
//
// AE 2020 compatible (no AE 23+ APIs).
(function () {
    // ---- Real image file path used for file-backed footage ----
    var imagePath = "e:/projects/tools/aep-parser/image.png";

    var mode = $.getenv("RE_XPROJ_MODE");
    if (mode === null || mode === "") mode = "footage";

    var validModes = { footage: true, precomp: true, dedup: true };
    if (!validModes[mode]) {
        var errFile = new File("e:/projects/tools/aep-parser/test_data/re_xproj_unknown.done");
        errFile.open("w");
        errFile.write("ERR unknown mode: " + mode);
        errFile.close();
        return;
    }

    var srcFile  = new File("e:/projects/tools/aep-parser/test_data/re_xproj_src_"  + mode + ".aep");
    var destFile = new File("e:/projects/tools/aep-parser/test_data/re_xproj_dest_" + mode + ".aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/re_xproj_"      + mode + ".done");
    var log = ["mode=" + mode];
    try { log.push("ae=" + app.version); } catch (e) {}

    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }

    var failReason = null;
    function importImage() {
        var io = new ImportOptions(new File(imagePath));
        return app.project.importFile(io);
    }

    // ======================== BUILD SRC PROJECT ========================
    step("src_fresh_project", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        app.project.save(srcFile);
    });

    if (mode === "footage") {
        // src_comp: one layer sourced from a file-backed footage item (image.png).
        step("src_build_footage", function () {
            var footageItem = importImage();
            var srcComp = app.project.items.addComp("src_comp", 1920, 1080, 1, 5, 24);
            srcComp.layers.add(footageItem);
        });
    } else if (mode === "precomp") {
        // src_comp: one layer whose source is inner_comp (a precomp AV layer).
        // inner_comp: a small comp with a solid filler.
        step("src_build_precomp", function () {
            var innerComp = app.project.items.addComp("inner_comp", 320, 240, 1, 3, 24);
            innerComp.layers.addSolid([0.4, 0.6, 0.8], "inner_filler", 100, 100, 1);
            var srcComp = app.project.items.addComp("src_comp", 1920, 1080, 1, 5, 24);
            srcComp.layers.add(innerComp);
        });
    } else if (mode === "dedup") {
        // src_comp: one layer sourced from file-backed footage at imagePath.
        step("src_build_dedup", function () {
            var footageItem = importImage();
            var srcComp = app.project.items.addComp("src_comp", 1920, 1080, 1, 5, 24);
            srcComp.layers.add(footageItem);
        });
    }

    step("src_save", function () {
        app.project.save(srcFile);
    });

    // ======================== BUILD DEST PROJECT ========================
    step("dest_fresh_project", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        app.project.save(destFile);
    });

    if (mode === "footage") {
        // dest_comp: one unrelated solid. Must NOT already reference imagePath.
        step("dest_build_footage", function () {
            var destComp = app.project.items.addComp("dest_comp", 1920, 1080, 1, 5, 24);
            destComp.layers.addSolid([0.2, 0.8, 0.2], "dest_unrelated", 100, 100, 1);
        });
    } else if (mode === "precomp") {
        // dest_comp: one unrelated solid. inner_comp is absent from dest.
        step("dest_build_precomp", function () {
            var destComp = app.project.items.addComp("dest_comp", 1920, 1080, 1, 5, 24);
            destComp.layers.addSolid([0.9, 0.5, 0.1], "dest_unrelated", 100, 100, 1);
        });
    } else if (mode === "dedup") {
        // dest_comp: one unrelated solid PLUS the SAME imagePath imported as
        // footage. The cross-Project insert must dedup to this existing item
        // rather than creating a second footage entry with the same Path.
        step("dest_build_dedup", function () {
            var sameFootage = importImage(); // same path P as src
            var destComp = app.project.items.addComp("dest_comp", 1920, 1080, 1, 5, 24);
            destComp.layers.addSolid([0.1, 0.1, 0.9], "dest_unrelated", 100, 100, 1);
            // Add the footage as a disabled layer so it's firmly in the project
            // (AE sometimes drops unreferenced footage on save — holding it in a
            // layer prevents that).
            var hold = destComp.layers.add(sameFootage);
            hold.enabled = false;
        });
    }

    step("dest_save", function () {
        app.project.save(destFile);
    });

    step("done_write", function () {
        if (failReason) {
            log.push("FAIL " + failReason);
        } else {
            log.push("PASS");
        }
        doneFile.open("w");
        doneFile.write(log.join("\n"));
        doneFile.close();
    });

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
