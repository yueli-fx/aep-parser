// Ship-gate verify for Property.SetDimensionsSeparated (both directions).
// Reads separate_dims_args.json {input, done, resaved, mode, x, y, z?}:
//   mode "sep"   -> assert leader dimensionsSeparated=true + Position_0/1(/2) values
//   mode "merge" -> assert leader dimensionsSeparated=false + position.value == [x,y,z]
// Writes PASS/FAIL + diagnostics to .done, resaves.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/separate_dims_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  " + m); }
    function near(a, b) { return Math.abs(a - b) <= 0.001; }

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
            var tg = comp.layer(1).property("ADBE Transform Group");
            var pos = tg.property("ADBE Position");
            log.push("  mode=" + args.mode + " dimensionsSeparated=" + pos.dimensionsSeparated);

            if (args.mode === "merge") {
                if (pos.dimensionsSeparated) fail("dimensionsSeparated != false");
                var v = pos.value;
                log.push("  value=" + v.toString());
                if (!near(v[0], args.x) || !near(v[1], args.y) || !near(v[2], args.z)) {
                    fail("value " + v.toString() + " != [" + args.x + "," + args.y + "," + args.z + "]");
                }
            } else { // "sep"
                if (!pos.dimensionsSeparated) fail("dimensionsSeparated != true");
                var px = tg.property("ADBE Position_0"), py = tg.property("ADBE Position_1");
                log.push("  Position_0=" + (px ? px.value : "MISSING") + " Position_1=" + (py ? py.value : "MISSING"));
                if (!px || !near(px.value, args.x)) fail("Position_0 != " + args.x);
                if (!py || !near(py.value, args.y)) fail("Position_1 != " + args.y);
                if (args.z !== null && args.z !== undefined) {
                    var pz = tg.property("ADBE Position_2");
                    log.push("  Position_2=" + (pz ? pz.value : "MISSING"));
                    if (!pz || !near(pz.value, args.z)) fail("Position_2 != " + args.z);
                }
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
