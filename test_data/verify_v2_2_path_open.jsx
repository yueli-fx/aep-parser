// test_data/verify_v2_2_path_open.jsx
//
// Open shape-path ship-gate — opens an .aep with NewShapeLayer + AddPath (an
// OPEN polyline, SetClosed(false)) + AddFill. Asserts AE accepted the layer
// (no silent-drop / crash) and re-saves so the Go side can decode the re-saved
// shph closed flag + ldat anchors. The path's OPEN-ness is verified Go-side
// (ExtendScript shape .value throws "division by zero", same as the closed gate).
//
// args.json fixed path: test_data/v2_2_path_open_args.json — {"input","done","resaved"}

(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_path_open_args.json");
    var args = {};
    var doneFile = null;
    var log = [];
    var ok = false;

    function check(name, cond) {
        if (!cond) { log.push("FAIL: " + name); return false; }
        log.push("OK:   " + name);
        return true;
    }
    function layerByName(comp, n) {
        for (var i = 1; i <= comp.layers.length; i++) {
            if (comp.layers[i].name === n) return comp.layers[i];
        }
        return null;
    }
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

        try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (ePre) {}

        app.open(new File(args.input));
        var c = app.project.items[1];
        log.push("opened comp: " + c.name + " layers=" + c.layers.length);

        var L = layerByName(c, "Path_Open");
        var checks = [
            check("layer count 1", c.layers.length === 1),
            check("Path_Open layer exists (not dropped/crashed)", L !== null)
        ];
        if (L) {
            var root = L.property("ADBE Root Vectors Group");
            var pg = findByMatchName(root, "ADBE Vector Shape - Group");
            var pathProp = findByMatchName(root, "ADBE Vector Shape");
            checks.push(check("Path group present", pg !== null));
            checks.push(check("Path shape stream present", pathProp !== null));
        }

        ok = true;
        for (var k = 0; k < checks.length; k++) {
            if (!checks[k]) { ok = false; break; }
        }

        if (ok && args.resaved) {
            app.project.save(new File(args.resaved));
            log.push("resaved to " + args.resaved);
        }
    } catch (e) {
        ok = false;
        log.push("ERROR: " + e.toString());
    }

    try {
        if (!doneFile) doneFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_path_open_test.done");
        doneFile.open("w");
        doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
        doneFile.close();
    } catch (e2) {}

    try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (eClose) {}
    try { app.quit(); } catch (eQuit) {}
})();
