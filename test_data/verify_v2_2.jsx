// test_data/verify_v2_2.jsx
//
// V2.2 AE ship gate — opens an .aep produced by NewProject + 3 NewShapeLayer
// (canonical shape graph, spec §5.2) and validates the runtime semantics
// AE re-derives match what Go API set.
//
// Driven by Go test (TestV2_2_AEShipGate_*). Same args.json + done-file
// protocol as verify_v2_1.jsx.
//
// args.json (fixed path: test_data/v2_2_args.json) contains:
//   {"input":"<abs>","done":"<abs>","resaved":"<abs>"}
//
// Everything in try/catch — .done MUST be written for the Go side not to
// timeout-hang.

(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_args.json");
    var args = {};
    var doneFile = null;
    var log = [];
    var ok = false;

    function check(name, cond) {
        if (!cond) { log.push("FAIL: " + name); return false; }
        log.push("OK:   " + name);
        return true;
    }
    function approxEq(arr1, arr2, tol) {
        if (arr1.length !== arr2.length) return false;
        for (var i = 0; i < arr1.length; i++) {
            if (Math.abs(arr1[i] - arr2[i]) > tol) return false;
        }
        return true;
    }
    function layerByName(comp, n) {
        for (var i = 1; i <= comp.layers.length; i++) {
            if (comp.layers[i].name === n) return comp.layers[i];
        }
        return null;
    }
    function findContent(layer, matchName) {
        var contents = layer.property("ADBE Root Vectors Group");
        for (var i = 1; i <= contents.numProperties; i++) {
            var p = contents.property(i);
            if (p.matchName === matchName) return p;
        }
        return null;
    }

    try {
        argsFile.open("r");
        var s = argsFile.read();
        argsFile.close();
        args = eval("(" + s + ")");
        doneFile = new File(args.done);

        // Pre-clean any prior project so app.open doesn't prompt about
        // unsaved changes from a leftover session.
        try {
            if (app.project) {
                app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
            }
        } catch (ePre) { /* no prior project — fine */ }

        app.open(new File(args.input));
        var c = app.project.items[1];
        log.push("opened comp: " + c.name + " layers=" + c.layers.length);

        var A = layerByName(c, "A_RectFill_Animated");
        var B = layerByName(c, "B_EllipseStroke_Static");
        var C = layerByName(c, "C_PathFillStroke_Static");

        var checks = [
            check("layer count 3", c.layers.length === 3),
            check("layer A exists", A !== null),
            check("layer B exists", B !== null),
            check("layer C exists", C !== null)
        ];

        if (A) {
            var rectA = findContent(A, "ADBE Vector Shape - Rect");
            var fillA = findContent(A, "ADBE Vector Graphic - Fill");
            checks.push(check("A has Rect", rectA !== null));
            checks.push(check("A has Fill", fillA !== null));

            if (rectA) {
                var sizeProp = rectA.property("ADBE Vector Rect Size");
                checks.push(check("A Rect Size kf count 2", sizeProp.numKeys === 2));
                if (sizeProp.numKeys >= 2) {
                    checks.push(check("A Rect Size kf[0]", approxEq(sizeProp.keyValue(1), [50, 50], 1e-3)));
                    checks.push(check("A Rect Size kf[1]", approxEq(sizeProp.keyValue(2), [300, 200], 1e-3)));
                }
            }
            if (fillA) {
                var colorProp = fillA.property("ADBE Vector Fill Color");
                checks.push(check("A Fill Color kf count 2", colorProp.numKeys === 2));
                if (colorProp.numKeys >= 2) {
                    var kf0 = colorProp.keyValue(1);
                    checks.push(check("A Fill Color kf[0] red", approxEq([kf0[0], kf0[1], kf0[2]], [1, 0, 0], 1e-3)));
                }
            }
            var posA = A.property("ADBE Transform Group").property("ADBE Position");
            checks.push(check("A Position kf count 2", posA.numKeys === 2));
            if (posA.numKeys >= 2) {
                checks.push(check("A Position kf[1]", approxEq(posA.keyValue(2), [500, 300], 1e-3)));
            }
        }

        if (B) {
            var ellB = findContent(B, "ADBE Vector Shape - Ellipse");
            var strokeB = findContent(B, "ADBE Vector Graphic - Stroke");
            checks.push(check("B has Ellipse", ellB !== null));
            checks.push(check("B has Stroke", strokeB !== null));
            if (ellB) {
                var sizeB = ellB.property("ADBE Vector Ellipse Size").value;
                checks.push(check("B Ellipse Size", approxEq(sizeB, [150, 150], 1e-3)));
            }
            if (strokeB) {
                var widthB = strokeB.property("ADBE Vector Stroke Width").value;
                checks.push(check("B Stroke Width 5", widthB === 5));
            }
        }

        if (C) {
            var pathC = findContent(C, "ADBE Vector Shape - Group");
            var fillC = findContent(C, "ADBE Vector Graphic - Fill");
            var strokeC = findContent(C, "ADBE Vector Graphic - Stroke");
            checks.push(check("C has Path", pathC !== null));
            checks.push(check("C has Fill", fillC !== null));
            checks.push(check("C has Stroke", strokeC !== null));
            if (pathC) {
                var shp = pathC.property("ADBE Vector Shape").value;
                checks.push(check("C Path vertex count 4", shp.vertices.length === 4));
                checks.push(check("C Path closed", shp.closed === true));
            }
            if (fillC) {
                var colC = fillC.property("ADBE Vector Fill Color").value;
                checks.push(check("C Fill green", approxEq([colC[0], colC[1], colC[2]], [0, 1, 0], 1e-3)));
            }
        }

        ok = true;
        for (var k = 0; k < checks.length; k++) {
            if (!checks[k]) { ok = false; break; }
        }
    } catch (e) {
        ok = false;
        log.push("ERROR: " + e.toString());
    }

    try {
        if (!doneFile) doneFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_test.done");
        doneFile.open("w");
        doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
        doneFile.close();
    } catch (e2) { /* swallow */ }

    // Tear down AE cleanly so it doesn't prompt user about unsaved
    // changes on the next launch. Order matters: close project (no save)
    // first, then quit.
    try {
        if (app.project) {
            app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        }
    } catch (eClose) { /* swallow */ }
    try {
        app.quit();
    } catch (eQuit) { /* swallow */ }
})();
