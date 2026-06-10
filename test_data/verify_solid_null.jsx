// Ship-gate verify for NewSolidLayer / NewNullLayer / NewAdjustmentLayer.
// Reads solid_null_args.json {input, done, resaved}. Opens the Go-written
// file, confirms AE accepts it and that the comp contains: a solid "Red BG"
// (1280x720, color ~[0.75,0.25,0.5]), a null "Controller" (100x100 source,
// nullLayer flag), and an adjustment layer "Grade" (comp-sized source,
// adjustmentLayer flag). Resaves so the Go side can confirm AE kept them.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/solid_null_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  " + m); }
    function near(a, b) { return Math.abs(a - b) < 1e-4; }

    try {
        app.open(new File(args.input));

        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem) { comp = it; break; }
        }
        if (!comp) {
            fail("no CompItem in project");
        } else {
            var byName = {};
            var names = [];
            for (var li = 1; li <= comp.numLayers; li++) {
                var L = comp.layer(li);
                byName[L.name] = L;
                names.push(L.name);
            }
            log.push("  layers=[" + names.join(", ") + "]");

            var solid = byName["Red BG"];
            if (!solid) fail("solid 'Red BG' missing");
            else {
                if (!solid.source) fail("solid has no source item");
                else {
                    var ms = solid.source.mainSource;
                    if (!(ms instanceof SolidSource)) fail("solid mainSource is not SolidSource");
                    else {
                        log.push("  solid color=" + ms.color.toString() +
                            " dims=" + solid.source.width + "x" + solid.source.height +
                            " srcName=" + solid.source.name);
                        if (!near(ms.color[0], 0.75) || !near(ms.color[1], 0.25) || !near(ms.color[2], 0.5))
                            fail("solid color " + ms.color.toString() + " != [0.75,0.25,0.5]");
                        if (solid.source.width !== 1280 || solid.source.height !== 720)
                            fail("solid dims " + solid.source.width + "x" + solid.source.height + " != 1280x720");
                        if (solid.source.name !== "Red BG")
                            fail("solid source name " + solid.source.name + " != Red BG");
                    }
                }
                if (solid.nullLayer) fail("solid has nullLayer flag");
                if (solid.adjustmentLayer) fail("solid has adjustmentLayer flag");
            }

            var nul = byName["Controller"];
            if (!nul) fail("null 'Controller' missing");
            else {
                log.push("  null flag=" + nul.nullLayer +
                    " dims=" + nul.source.width + "x" + nul.source.height);
                if (!nul.nullLayer) fail("'Controller' nullLayer flag not set");
                if (nul.source.width !== 100 || nul.source.height !== 100)
                    fail("null dims " + nul.source.width + "x" + nul.source.height + " != 100x100");
            }

            var adj = byName["Grade"];
            if (!adj) fail("adjustment 'Grade' missing");
            else {
                log.push("  adj flag=" + adj.adjustmentLayer +
                    " dims=" + adj.source.width + "x" + adj.source.height);
                if (!adj.adjustmentLayer) fail("'Grade' adjustmentLayer flag not set");
                if (adj.source.width !== 1280 || adj.source.height !== 720)
                    fail("adjustment dims " + adj.source.width + "x" + adj.source.height + " != 1280x720");
            }
        }

        app.project.save(new File(args.resaved));
    } catch (e) {
        fail("EXC " + e.toString());
    }

    var done = new File(args.done);
    done.open("w");
    done.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n"));
    done.close();

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
