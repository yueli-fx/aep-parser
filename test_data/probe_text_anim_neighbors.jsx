// RE discovery — confirm the "free neighbor" Text Animator leaf match-names,
// propertyValueType, and on-disk value shape, so AddText*Animator can be added
// for each by extracting a per-leaf template (same vein as Opacity/Position/…).
//
// Probes each candidate match-name under "ADBE Text Animator Properties": tries
// addProperty (try/catch), sets a non-default value, records matchName / name /
// propertyType / propertyValueType / value. Saves the .aep so extract tooling
// can read the materialized cdats, and writes a .done dump.
//
// Output:
//   test_data/probe_text_anim_neighbors.aep
//   test_data/probe_text_anim_neighbors.done
(function () {
    var base = "e:/projects/tools/aep-parser/test_data/probe_text_anim_neighbors";
    var outFile  = new File(base + ".aep");
    var doneFile = new File(base + ".done");
    var log = [];
    try { log.push("ae=" + app.version); } catch (e) {}

    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }

    var comp, tlayer, props;

    step("fresh_project", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        app.project.save(outFile);
    });
    step("add_comp_text", function () {
        comp = app.project.items.addComp("PROBE", 1280, 720, 1, 5, 24);
        tlayer = comp.layers.addText("ABCDEF");
        var txtProps  = tlayer.property("ADBE Text Properties");
        var animators = txtProps.property("ADBE Text Animators");
        var anim      = animators.addProperty("ADBE Text Animator");
        props         = anim.property("ADBE Text Animator Properties");
    });

    // candidate: [matchName, valueToSet]  (1D scalar => number, color => [r,g,b,a])
    var cands = [
        ["ADBE Text Fill Opacity",   50],
        ["ADBE Text Stroke Opacity", 50],
        ["ADBE Text Stroke Color",   [0.2, 0.4, 0.8, 1]],
        ["ADBE Text Stroke Width",   12],
        ["ADBE Text Skew",           20],
        ["ADBE Text Skew Axis",      30],
        ["ADBE Text Rotation X",     45],
        ["ADBE Text Rotation Y",     45]
    ];

    for (var i = 0; i < cands.length; i++) {
        (function (mn, val) {
            step("probe " + mn, function () {
                var p = props.addProperty(mn);
                try { p.setValue(val); } catch (e2) { log.push("    setValue err: " + e2); }
                var line = "  FOUND mn=" + p.matchName + " name=" + p.name +
                           " ptype=" + p.propertyType;
                try { line += " vtype=" + p.propertyValueType; } catch (e3) {}
                try { line += " value=" + p.value.toString(); } catch (e4) {}
                log.push(line);
            });
        })(cands[i][0], cands[i][1]);
    }

    step("save", function () { app.project.save(); });

    try { doneFile.open("w"); doneFile.write(log.join("\n")); doneFile.close(); } catch (e) {}
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
