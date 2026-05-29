// test_data/verify_v2_2_ellipse.jsx
//
// V2.2.1 focused Ellipse ship-gate — opens an .aep produced by NewProject +
// NewShapeLayer + AddEllipse + AddFill and validates AE accepted the Ellipse
// (did NOT silent-drop it) and re-derives the Size/Position values Go set via
// the embed+overwrite path.
//
// Driven by Go test (TestV2_2_Ellipse_AEShipGate_*). args.json fixed path:
//   test_data/v2_2_ellipse_args.json — {"input":"<abs>","done":"<abs>","resaved":"<abs>"}
//
// NOTE: AE ExtendScript throws "Numeric result invalid (division by zero?)"
// when reading `.value` / `.valueAtTime` on Ellipse Size/Position — this hits
// even AE-native fixtures, so it's an API quirk, not a file defect. We instead
// prove AE accepted our overwritten values by re-saving and letting the Go
// side parse the re-saved cdat (values differ from the [200,100]/[50,30]
// tolerance fixture so a no-op overwrite is caught).
//
// Everything in try/catch — .done MUST be written or the Go side timeout-hangs.

(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_ellipse_args.json");
    var args = {};
    var doneFile = null;
    var log = [];
    var ok = false;

    function check(name, cond) {
        if (!cond) { log.push("FAIL: " + name); return false; }
        log.push("OK:   " + name);
        return true;
    }
    function approxEq(a, b, tol) {
        if (a.length !== b.length) return false;
        for (var i = 0; i < a.length; i++) {
            if (Math.abs(a[i] - b[i]) > tol) return false;
        }
        return true;
    }
    function layerByName(comp, n) {
        for (var i = 1; i <= comp.layers.length; i++) {
            if (comp.layers[i].name === n) return comp.layers[i];
        }
        return null;
    }
    // Recursive search: shapes live nested under Root Vectors Group →
    // Vector Group → Vectors Group → <shape>, so a flat scan misses them.
    function findByMatchName(group, matchName) {
        if (!group || !group.numProperties) return null;
        for (var i = 1; i <= group.numProperties; i++) {
            var p = group.property(i);
            if (p.matchName === matchName) return p;
            if (p.numProperties && p.numProperties > 0) {
                var hit = findByMatchName(p, matchName);
                if (hit) return hit;
            }
        }
        return null;
    }

    try {
        argsFile.open("r");
        var s = argsFile.read();
        argsFile.close();
        args = eval("(" + s + ")");
        doneFile = new File(args.done);

        try {
            if (app.project) {
                app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
            }
        } catch (ePre) { /* no prior project — fine */ }

        app.open(new File(args.input));
        var c = app.project.items[1];
        log.push("opened comp: " + c.name + " layers=" + c.layers.length);

        var L = layerByName(c, "Ellipse_Static");
        var checks = [
            check("layer count 1", c.layers.length === 1),
            check("Ellipse_Static layer exists (not silent-dropped)", L !== null)
        ];

        if (L) {
            var root = L.property("ADBE Root Vectors Group");
            var ell = findByMatchName(root, "ADBE Vector Shape - Ellipse");
            var fill = findByMatchName(root, "ADBE Vector Graphic - Fill");
            checks.push(check("Ellipse shape present (not dropped)", ell !== null));
            checks.push(check("Fill present", fill !== null));
            // Value correctness is proven Go-side from the re-saved file (see
            // header note on the .value/.valueAtTime AE quirk). Here we just
            // re-save so the Go side can parse AE's serialized Ellipse cdat.
        }

        ok = true;
        for (var k = 0; k < checks.length; k++) {
            if (!checks[k]) { ok = false; break; }
        }

        // Re-save so Go can verify AE retained the overwritten Size/Position.
        if (ok && args.resaved) {
            app.project.save(new File(args.resaved));
            log.push("resaved to " + args.resaved);
        }
    } catch (e) {
        ok = false;
        log.push("ERROR: " + e.toString());
    }

    try {
        if (!doneFile) doneFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_ellipse_test.done");
        doneFile.open("w");
        doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
        doneFile.close();
    } catch (e2) { /* swallow */ }

    try {
        if (app.project) {
            app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        }
    } catch (eClose) { /* swallow */ }
    try {
        app.quit();
    } catch (eQuit) { /* swallow */ }
})();
