// RE probe: what makes AE persist (vs default-elide) an effect parameter in
// the saved sspc — a non-default VALUE, or a touched/modified FLAG?
//
// Three solid layers, each with a Gaussian Blur instance:
//   L1 "touched"   — every param set to a non-default value
//                    (+ a Drop Shadow with only 2 of 5 params touched)
//   L2 "resetdef"  — every param setValue(its current default value)
//   L3 "untouched" — added, never touched (control; expect 1 surfaced param)
//
// If L2 persists all params at default values, the typed-param-helper
// extraction strategy is simply "setValue(default) on every param" — templates
// carry every param chunk while keeping AE's defaults. If L2 elides like L3,
// persistence keys off value≠default and templates would carry non-default
// values needing a post-patch.
//
// Run via scripts/ae_run.ps1 against AE 2020.
// Output: test_data/generated/fixtures/re_effect_param_elision.aep + .done
(function () {
    var outFile  = new File("e:/projects/tools/aep-parser/test_data/generated/fixtures/re_effect_param_elision.aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/re_effect_param_elision.done");
    var log = [];
    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }
    try { log.push("ae=" + app.version); } catch (e) {}

    var comp;
    step("fresh_project", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        comp = app.project.items.addComp("ElisionProbe", 1920, 1080, 1, 5, 24);
    });

    function addGB(layerName) {
        var l = comp.layers.addSolid([0.5, 0.5, 0.5], layerName, 100, 100, 1);
        return { layer: l, fx: l.property("ADBE Effect Parade").addProperty("ADBE Gaussian Blur 2") };
    }

    step("L1_touched_all_nondefault", function () {
        var a = addGB("touched");
        var n = a.fx.numProperties;
        log.push("  GB numProperties=" + n);
        for (var i = 1; i <= n; i++) {
            var p = a.fx.property(i);
            log.push("  GB param " + i + ": " + p.matchName + " default=" + p.value);
        }
        a.fx.property("ADBE Gaussian Blur 2-0001").setValue(25);   // Blurriness 0 -> 25
        a.fx.property("ADBE Gaussian Blur 2-0002").setValue(2);    // Dimensions 1 -> 2
        a.fx.property("ADBE Gaussian Blur 2-0003").setValue(true); // Repeat Edge false -> true
        var ds = a.layer.property("ADBE Effect Parade").addProperty("ADBE Drop Shadow");
        var dn = ds.numProperties;
        log.push("  DS numProperties=" + dn);
        for (var j = 1; j <= dn; j++) {
            log.push("  DS param " + j + ": " + ds.property(j).matchName + " default=" + ds.property(j).value);
        }
        ds.property("ADBE Drop Shadow-0004").setValue(20);  // Distance 5 -> 20
        ds.property("ADBE Drop Shadow-0005").setValue(10);  // Softness 0 -> 10
    });

    step("L2_setvalue_defaults", function () {
        var b = addGB("resetdef");
        for (var i = 1; i <= b.fx.numProperties; i++) {
            var p = b.fx.property(i);
            if (p.propertyType === PropertyType.PROPERTY) {
                p.setValue(p.value);
            }
        }
    });

    step("L3_untouched", function () {
        addGB("untouched");
    });

    step("save", function () { app.project.save(outFile); });

    doneFile.open("w");
    doneFile.write("PASS\n" + log.join("\n"));
    doneFile.close();

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
