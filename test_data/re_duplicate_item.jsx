// RE probe for V3 Phase 5 candidate — Project.DuplicateItem (CompItem.duplicate()).
//
// Goal: learn the chunk-level delta AE writes when duplicating a comp:
//   - does the dup comp get a new item ID? (expected yes)
//   - do the dup's LAYERS get new layer IDs, or keep source IDs?  (crux)
//   - intra-comp parent/matte refs in the dup: remapped to the dup's new
//     layer IDs, or left pointing at source comp's layer IDs?            (crux)
//   - precomp/footage sources: shared (same item) or deep-cloned?
//   - folder/Sfdr + itemList ordering of the new comp.
//
// Scenario (single; state via $.getenv("RE_DUP_STATE") = before|after):
//   F1            — shared solid footage
//   compC_precomp — precomp with 1 filler layer
//   compA_main    — L1 "A_solid" (plain), L2 "A_footage" (over F1),
//                   L3 "A_precomp" (over compC); L1.parent = L3 (intra-comp ref)
//   after: app.project duplicates compA_main → "compA_main 2"
//
// Outputs:
//   test_data/re_duplicate_item_<state>.aep
//   test_data/re_duplicate_item_<state>.done
//
// AE 2020 compatible.
(function () {
    var state = $.getenv("RE_DUP_STATE");
    if (state === null || state === "") state = "before";

    var outFile  = new File("e:/projects/tools/aep-parser/test_data/re_duplicate_item_" + state + ".aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/re_duplicate_item_" + state + ".done");
    var log = ["state=" + state];
    try { log.push("ae=" + app.version); } catch (e) {}

    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }

    var F1, compC, compA, dup;

    step("fresh_project", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        app.project.save(outFile);
    });

    step("setup_precomp_compC", function () {
        compC = app.project.items.addComp("compC_precomp", 320, 240, 1, 5, 24);
        compC.layers.addSolid([0.5, 0.5, 0.5], "compC_filler", 100, 100, 1);
    });

    step("setup_compA_main", function () {
        compA = app.project.items.addComp("compA_main", 1920, 1080, 1, 5, 24);
        var l1 = compA.layers.addSolid([1, 0, 0], "A_solid", 100, 100, 1);
        F1 = l1.source; // capture the shared solid footage
        var l2 = compA.layers.add(F1);
        l2.name = "A_footage";
        var l3 = compA.layers.add(compC);
        l3.name = "A_precomp";
        // intra-comp parent ref: L1 (top) parented to L3 (bottom).
        l1.parent = l3;
    });

    step("dump_before_ids", function () {
        // record source layer indices/names for human cross-ref
        for (var j = 1; j <= compA.numLayers; j++) {
            var li = compA.layer(j);
            var info = "  src[" + j + "] " + li.name;
            if (li.parent) info += " parent=" + li.parent.name;
            try { if (li.source) info += " source=" + li.source.name; } catch (e) {}
            log.push(info);
        }
    });

    if (state === "after") {
        step("duplicate_compA", function () {
            dup = compA.duplicate();
            log.push("  dup.name=" + dup.name + " dup.numLayers=" + dup.numLayers);
            for (var j = 1; j <= dup.numLayers; j++) {
                var li = dup.layer(j);
                var info = "  dup[" + j + "] " + li.name;
                if (li.parent) info += " parent=" + li.parent.name;
                try { if (li.source) info += " source=" + li.source.name; } catch (e) {}
                log.push(info);
            }
        });
    }

    step("save", function () {
        app.project.save(outFile);
    });

    step("write_done", function () {
        doneFile.open("w");
        doneFile.write(log.join("\n"));
        doneFile.close();
    });

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
