// RE fixture — Text Animators (kinetic typography engine).
//
// Goal: establish the chunk structure AE emits for a text layer with one
// animator + a Range Selector + one animated property, so we can extract a
// verbatim animator-tdgp template and implement AddTextAnimator (synthesis-
// insert into the "ADBE Text Animators" indexed group, same vein as AddEffect).
//
// Questions:
//   Q1 Where does ADBE Text Animators live in the layer tree (nesting path),
//      and what propertyType is each container (animator / selectors / props)?
//   Q2 What match-names does an animator + range selector + a property carry?
//   Q3 Static fixture: animator with Opacity=0 + Range selector End=50 →
//      structure fully materialized for byte extraction.
//   Q4 Does an animated range Offset (keyframed -100→0) live as a normal
//      keyframed scalar stream (so existing AnimateScalarKeyframes applies)?
//
// Mode: $.getenv("RE_TXANIM_MODE") in {opacity, position, animated}. Default opacity.
//   opacity   : Animator 1 = Opacity 0, Range selector Start=0 End=50 (static)
//   position  : Animator 1 = Position [0,-100], Range selector Start=0 End=50
//   animated  : Animator 1 = Opacity 0, Range selector Offset keyframed -100→0
//
// AE 2020 (project read floor). All match-names are pre-2020 built-ins.
//
// Outputs:
//   test_data/re_text_animator_<mode>.aep
//   test_data/re_text_animator_<mode>.done   (recursive property-tree dump)
(function () {
    var mode = $.getenv("RE_TXANIM_MODE");
    if (mode === null || mode === "") mode = "opacity";
    var valid = "opacity|position|animated|template|postemplate|scaletemplate|rottemplate|colortemplate|animatedvec";
    if (valid.indexOf(mode) === -1) mode = "opacity";

    var base = "e:/projects/tools/aep-parser/test_data/re_text_animator_" + mode;
    var outFile  = new File(base + ".aep");
    var doneFile = new File(base + ".done");
    var log = ["mode=" + mode];
    try { log.push("ae=" + app.version); } catch (e) {}

    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }

    function dumpTree(prop, depth) {
        var pad = "";
        for (var d = 0; d < depth; d++) pad += "  ";
        var line = pad + "[" + prop.matchName + "] name=" + prop.name +
                   " type=" + prop.propertyType;
        try { if (prop.propertyValueType !== undefined) line += " vtype=" + prop.propertyValueType; } catch (e) {}
        try { if (prop.numKeys !== undefined && prop.numKeys > 0) line += " numKeys=" + prop.numKeys; } catch (e) {}
        try {
            if (prop.propertyType === PropertyType.PROPERTY && prop.value !== undefined) {
                line += " value=" + prop.value.toString();
            }
        } catch (e) {}
        log.push(line);
        if (prop.numProperties !== undefined && prop.numProperties > 0 && depth < 12) {
            for (var i = 1; i <= prop.numProperties; i++) {
                try { dumpTree(prop.property(i), depth + 1); } catch (e) { log.push(pad + "  <err " + e + ">"); }
            }
        }
    }

    var comp, tlayer, txtProps, animators, anim, sels, sel, props;

    step("fresh_project", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        app.project.save(outFile);
    });

    step("add_comp_text", function () {
        comp = app.project.items.addComp("TXANIM", 1280, 720, 1, 5, 24);
        tlayer = comp.layers.addText("ABCDEF");
    });

    step("add_animator", function () {
        txtProps  = tlayer.property("ADBE Text Properties");
        animators = txtProps.property("ADBE Text Animators");
        anim      = animators.addProperty("ADBE Text Animator");   // "Animator 1"
        sels      = anim.property("ADBE Text Selectors");
        sel       = sels.addProperty("ADBE Text Selector");        // "Range Selector 1"
        props     = anim.property("ADBE Text Animator Properties");
    });

    step("set_property", function () {
        if (mode === "position" || mode === "postemplate") {
            var pos = props.addProperty("ADBE Text Position 3D");
            // postemplate: all-non-default so every cdat slot persists (template
            // for AddTextPositionAnimator to overwrite parametrically).
            pos.setValue(mode === "postemplate" ? [30, -40, 0] : [0, -100, 0]);
        } else if (mode === "scaletemplate") {
            var sc = props.addProperty("ADBE Text Scale 3D");
            // non-default so every cdat slot persists (template for
            // AddTextScaleAnimator). Default is [100,100,100].
            sc.setValue([40, 60, 100]);
        } else if (mode === "rottemplate") {
            var rot = props.addProperty("ADBE Text Rotation");
            // non-default (default 0) so the cdat slot persists (template for
            // AddTextRotationAnimator). 1D scalar, same layout as Opacity.
            rot.setValue(45);
        } else if (mode === "colortemplate") {
            var fc = props.addProperty("ADBE Text Fill Color");
            // non-default (default [1,0,0,1] red) so the color cdat persists
            // (template for AddTextColorAnimator). Four distinct channels
            // [r,g,b,a]=[0.2,0.4,0.8,1] → decode the on-disk [A,R,G,B]×255 order.
            fc.setValue([0.2, 0.4, 0.8, 1]);
        } else if (mode === "animatedvec") {
            // Ground truth for ANIMATING a 3D/4D animator-properties leaf itself
            // (vs sweeping the Range Offset). Position 3D (spatial), Scale 3D
            // (non-spatial), Fill Color (4D) each keyframed 2 frames so we can
            // diff AE's keyframe-block layout against AnimateVectorKeyframes.
            var pos = props.addProperty("ADBE Text Position 3D");
            pos.setValueAtTime(0, [0, 0, 0]);
            pos.setValueAtTime(2, [100, 50, 0]);
            var sc = props.addProperty("ADBE Text Scale 3D");
            sc.setValueAtTime(0, [100, 100, 100]);
            sc.setValueAtTime(2, [200, 200, 100]);
            var fcv = props.addProperty("ADBE Text Fill Color");
            fcv.setValueAtTime(0, [1, 0, 0, 1]);
            fcv.setValueAtTime(2, [0, 0, 1, 1]);
        } else {
            var op = props.addProperty("ADBE Text Opacity");
            op.setValue(0);
        }
    });

    step("set_selector", function () {
        var start = sel.property("ADBE Text Percent Start");
        var end   = sel.property("ADBE Text Percent End");
        var off   = sel.property("ADBE Text Percent Offset");
        if (mode === "animated") {
            start.setValue(0); end.setValue(100);
            off.setValueAtTime(0, -100);
            off.setValueAtTime(2, 0);
        } else if (mode === "animatedvec") {
            // Full static selection so the animated leaf applies to every char.
            start.setValue(0); end.setValue(100); off.setValue(0);
        } else if (mode === "template" || mode === "postemplate" || mode === "scaletemplate" || mode === "rottemplate" || mode === "colortemplate") {
            // All three set non-default so AE persists every slot — gives
            // AddTextAnimator a template whose Start/End/Offset cdats are all
            // present to overwrite parametrically.
            start.setValue(20); end.setValue(80); off.setValue(10);
        } else {
            start.setValue(0); end.setValue(50); off.setValue(0);
        }
    });

    step("dump_tree", function () {
        log.push("---- text layer property tree ----");
        dumpTree(tlayer.property("ADBE Text Properties"), 0);
        // selector advanced sub-group match-names
        log.push("---- range selector advanced ----");
        try {
            var adv = sel.property("ADBE Text Range Advanced");
            if (adv) dumpTree(adv, 0);
        } catch (e) { log.push("no advanced: " + e); }
    });

    step("save", function () { app.project.save(); });

    try { doneFile.open("w"); doneFile.write(log.join("\n")); doneFile.close(); } catch (e) {}
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
