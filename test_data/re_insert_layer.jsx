// RE fixture for V3 Phase 5C — Composition.InsertLayer (cross-comp clone, same-Project).
//
// Produces 3 mode pairs (before/after) the Go ship-gate compares against:
//   basic    — compA: 1 solid "Src" no parent/no matte; compB: 1 unrelated solid.
//              After: compB has Src clone inserted at index 1.
//   footage  — compA: 1 AV layer over Footage F1; compB: 1 unrelated solid + F1
//              already imported. After: compB has Src clone refs same F1.
//   precomp  — compA: 1 AV layer whose source is precomp compC; compB: 1
//              unrelated solid + compC. After: compB has Src clone refs compC.
//
// Mode via $.getenv("RE_INSERT_MODE"); state ("before" | "after") via
// $.getenv("RE_INSERT_STATE").
//
// Outputs:
//   test_data/re_insert_layer_<mode>_<state>.aep
//   test_data/re_insert_layer_<mode>_<state>.done
//
// AE 2020 compatible (no AE 23+ APIs).
(function () {
    var mode  = $.getenv("RE_INSERT_MODE");
    var state = $.getenv("RE_INSERT_STATE");
    if (mode === null || mode === "")  mode  = "basic";
    if (state === null || state === "") state = "before";
    var validModes = "basic|footage|precomp";
    var validStates = "before|after";
    if (validModes.indexOf(mode) === -1 || validStates.indexOf(state) === -1) {
        var errFile = new File("e:/projects/tools/aep-parser/test_data/re_insert_layer_unknown.done");
        errFile.open("w");
        errFile.write("ERR unknown mode/state: " + mode + "/" + state);
        errFile.close();
        return;
    }

    var outFile  = new File("e:/projects/tools/aep-parser/test_data/re_insert_layer_" + mode + "_" + state + ".aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/re_insert_layer_" + mode + "_" + state + ".done");
    var log = ["mode=" + mode, "state=" + state];
    try { log.push("ae=" + app.version); } catch (e) {}

    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }

    var compA, compB, compC, footageF1, src, dst, clone;

    step("fresh_project", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        app.project.save(outFile);
    });

    step("setup_shared_items", function () {
        if (mode === "precomp") {
            compC = app.project.items.addComp("compC_precomp", 320, 240, 1, 5, 24);
            compC.layers.addSolid([0.5, 0.5, 0.5], "compC_filler", 100, 100, 1);
        }
        // footage mode: solid footage is created in setup_compA (ItemCollection
        // has no addSolid; LayerCollection.addSolid creates the FootageItem in a
        // "Solids" folder, then footageF1 = src.source captures the shared asset).
    });

    step("setup_compA", function () {
        compA = app.project.items.addComp("compA_src_holder", 1920, 1080, 1, 5, 24);
        if (mode === "basic") {
            src = compA.layers.addSolid([1, 0, 0], "Src", 100, 100, 1);
        } else if (mode === "footage") {
            src = compA.layers.addSolid([0.25, 0.25, 0.75], "Src_footage", 200, 200, 1);
            footageF1 = src.source;
        } else if (mode === "precomp") {
            src = compA.layers.add(compC);
            src.name = "Src_precomp";
        }
    });

    step("setup_compB", function () {
        compB = app.project.items.addComp("compB_dest", 1920, 1080, 1, 5, 24);
        compB.layers.addSolid([0, 1, 0], "B_unrelated", 100, 100, 1);
        if (mode === "footage") {
            // pre-touch compB's relationship with F1 (used elsewhere already)
            var extra = compB.layers.add(footageF1);
            extra.enabled = false;
        } else if (mode === "precomp") {
            var extra2 = compB.layers.add(compC);
            extra2.enabled = false;
        }
    });

    if (state === "after") {
        step("perform_insert", function () {
            // Insert clone of compA's Src at top of compB.
            clone = src.copyToComp(compB);
            // copyToComp inserts at the top of dest (index 1 in AE 1-based).
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
})();
