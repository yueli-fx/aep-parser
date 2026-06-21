// RE template generator — author one "ADBE Text Animators" group per remaining
// Booyah comp ② leaf (Tracking Amount + Character Offset), each with that single
// leaf + Range Start/End/Offset all non-default so AE persists (materializes)
// every cdat slot. Each saved to its own .aep so the body can be extracted into
// internal/serializer/templates/text/text_animators_<key>_body.bin.
//
// Both are 1D scalars (RE'd from the real Booyah project, tmp_debug/probe_text_tracking:
// tdb4 head db990001 = 1 component). Each auto-materializes companion leaves that
// AE adds with it (Tracking Amount → Track Type; Character Offset → Character
// Change Type + Character Range) — left at default, the match-name-targeted
// overwrite ignores them (same as Rotation X/Y's companion Rotation Z).
//
// Output per key: test_data/re_text_animator_<key>.aep  (+ shared .done log)
(function () {
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/gen_text_anim_tracking_charoffset_templates.done");
    var log = [];
    try { log.push("ae=" + app.version); } catch (e) {}

    // key -> [matchName, value]  (both 1D scalar)
    var specs = [
        ["tracking",  "ADBE Text Tracking Amount",  50],
        ["charoffset", "ADBE Text Character Offset", 5]
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
        log.push("OK  " + key + " (" + mn + ") -> " + out.fsName + " ; animProps=" + props.numProperties);
    }

    for (var i = 0; i < specs.length; i++) {
        try { authorOne(specs[i][0], specs[i][1], specs[i][2]); }
        catch (e) { log.push("ERR " + specs[i][0] + " -> " + e.toString()); }
    }

    try { doneFile.open("w"); doneFile.write(log.join("\n")); doneFile.close(); } catch (e) {}
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
