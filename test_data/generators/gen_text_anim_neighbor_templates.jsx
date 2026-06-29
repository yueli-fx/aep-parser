// RE template generator — author one "ADBE Text Animators" group per free-
// neighbor leaf, each with that single leaf + Range Start/End/Offset all set
// non-default so AE persists (materializes) every cdat slot. Each is saved to
// its own .aep so extract_text_animator can pull the verbatim group body into
// internal/serializer/templates/text_animators_<key>_body.bin.
//
// One AE invocation authors ALL templates (loop over specs, fresh project each).
// Free neighbors confirmed by probe_text_anim_neighbors.jsx: all 1D scalar
// (vtype 6417, same layout as Opacity/Rotation) except Stroke Color (vtype 6418
// color, same as Fill Color). Each leaf is given a distinct non-default value.
//
// Output per spec key: test_data/re_text_animator_<key>.aep  (+ shared .done log)
(function () {
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/gen_text_anim_neighbor_templates.done");
    var log = [];
    try { log.push("ae=" + app.version); } catch (e) {}

    // key -> [matchName, value]  (number = 1D scalar, array = color [r,g,b,a])
    var specs = [
        ["fillopacity",   "ADBE Text Fill Opacity",   50],
        ["strokeopacity", "ADBE Text Stroke Opacity", 50],
        ["strokecolor",   "ADBE Text Stroke Color",   [0.2, 0.4, 0.8, 1]],
        ["strokewidth",   "ADBE Text Stroke Width",   12],
        ["skew",          "ADBE Text Skew",           20],
        ["rotx",          "ADBE Text Rotation X",     45],
        ["roty",          "ADBE Text Rotation Y",     45]
    ];

    function authorOne(key, mn, val) {
        var out = new File("e:/projects/tools/aep-parser/test_data/re_text_animator_" + key + ".aep");
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        app.project.save(out);

        var comp   = app.project.items.addComp("TXANIM_" + key, 1280, 720, 1, 5, 24);
        var tlayer = comp.layers.addText("ABCDEF");
        var txtProps  = tlayer.property("ADBE Text Properties");
        var animators = txtProps.property("ADBE Text Animators");
        var anim      = animators.addProperty("ADBE Text Animator");
        var sels      = anim.property("ADBE Text Selectors");
        var sel       = sels.addProperty("ADBE Text Selector");
        var props     = anim.property("ADBE Text Animator Properties");

        var leaf = props.addProperty(mn);
        leaf.setValue(val);

        // All three Range slots non-default so AE materializes each cdat.
        sel.property("ADBE Text Percent Start").setValue(20);
        sel.property("ADBE Text Percent End").setValue(80);
        sel.property("ADBE Text Percent Offset").setValue(10);

        app.project.save();
        log.push("OK  " + key + " (" + mn + ") -> " + out.fsName);
    }

    for (var i = 0; i < specs.length; i++) {
        try { authorOne(specs[i][0], specs[i][1], specs[i][2]); }
        catch (e) { log.push("ERR " + specs[i][0] + " -> " + e.toString()); }
    }

    try { doneFile.open("w"); doneFile.write(log.join("\n")); doneFile.close(); } catch (e) {}
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
