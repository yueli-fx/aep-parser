// Fixture for RE'ing Property.alternateSource (AE 18+ Media Replacement /
// Essential Properties workflow).
//
//   Match names per docsforadobe:
//     ADBE Layer Overrides         = Essential Properties top-level group on AVLayer
//     ADBE Layer Source Alternate  = the alt-source Property (read via canSetAlternateSource)
//
//   Project layout:
//     ALT_SRC_A (precomp)            — original source for the swappable layer
//     ALT_SRC_B (precomp)            — alternate source we swap to
//     RE_ALT_SOURCE (precomp)        — wraps ALT_SRC_A in an inner layer; that layer is
//                                      marked for media replacement via
//                                      addToMotionGraphicsTemplateAs.
//     RE_ALT_SOURCE_MAIN (comp)      — uses RE_ALT_SOURCE twice:
//                                        baseline_no_alt → no override
//                                        with_alt_b      → setAlternateSource(ALT_SRC_B)
//
//   Note: addToMotionGraphicsTemplate* needs Master Properties enabled (AE 17+
//   default). canAddToMotionGraphicsTemplate() returns false if the precomp has
//   layers already promoted via the legacy Essential Graphics workflow only —
//   we use the modern path so it should succeed.

(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/re_altsource_ae24.aep");
    var log = [];
    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }

    step("redirect_save", function () { app.project.save(outFile); });

    var srcA, srcB;
    step("add_src_a", function () {
        srcA = app.project.items.addComp("ALT_SRC_A", 320, 240, 1, 5, 24);
        srcA.layers.addSolid([1, 0, 0], "innerA", 320, 240, 1);
    });
    step("add_src_b", function () {
        srcB = app.project.items.addComp("ALT_SRC_B", 320, 240, 1, 5, 24);
        srcB.layers.addSolid([0, 1, 0], "innerB", 320, 240, 1);
    });

    var wrapper;
    step("add_wrapper", function () {
        wrapper = app.project.items.addComp("RE_ALT_SOURCE", 320, 240, 1, 5, 24);
    });
    step("add_wrapper_layer_and_mark", function () {
        var L = wrapper.layers.add(srcA);
        L.name = "swappable_inner";
        log.push("  can_add=" + L.canAddToMotionGraphicsTemplate(wrapper));
        if (L.canAddToMotionGraphicsTemplate(wrapper)) {
            var ok = L.addToMotionGraphicsTemplateAs(wrapper, "MediaSlot");
            log.push("  addToMGT result=" + ok);
        }
    });

    var parent;
    step("add_parent", function () {
        parent = app.project.items.addComp("RE_ALT_SOURCE_MAIN", 320, 240, 1, 5, 24);
    });
    step("add_parent_baseline", function () {
        var L = parent.layers.add(wrapper);
        L.name = "baseline_no_alt";
        // Dump match names of this instance's properties for sanity.
        var ep = L.property("ADBE Layer Overrides");
        log.push("  baseline_layer_overrides=" + (ep ? ep.matchName + "/" + ep.numProperties : "<null>"));
        if (ep) {
            for (var i = 1; i <= ep.numProperties; i++) {
                var p = ep.property(i);
                log.push("    [" + i + "] " + p.matchName + " name='" + p.name + "' canSetAlt=" + (p.canSetAlternateSource === undefined ? "?" : p.canSetAlternateSource));
            }
        }
    });
    step("add_parent_with_alt", function () {
        var L = parent.layers.add(wrapper);
        L.name = "with_alt_b";
        var ep = L.property("ADBE Layer Overrides");
        if (!ep) { throw new Error("no ADBE Layer Overrides on with_alt_b"); }
        var altProp = null;
        for (var i = 1; i <= ep.numProperties; i++) {
            var p = ep.property(i);
            if (p.matchName === "ADBE Layer Source Alternate") { altProp = p; break; }
        }
        if (!altProp) { throw new Error("no ADBE Layer Source Alternate child"); }
        log.push("  altProp.canSetAlt=" + altProp.canSetAlternateSource);
        if (!altProp.canSetAlternateSource) { throw new Error("canSetAlternateSource=false"); }
        altProp.setAlternateSource(srcB);
        var post = altProp.alternateSource;
        log.push("  post_alt=" + (post ? post.name + " id=" + post.id : "<null>"));
    });

    step("save", function () { app.project.save(); });

    step("write_done_marker", function () {
        var marker = new File("e:/projects/tools/aep-parser/test_data/re_altsource_ae24.done");
        marker.open("w");
        marker.write(log.join("\n"));
        marker.close();
    });
})();
