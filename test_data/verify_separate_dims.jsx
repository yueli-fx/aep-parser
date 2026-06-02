// Ship-gate verify for Property.SetDimensionsSeparated.
// Opens the Go-written separated .aep, asserts AE reads the Position leader as
// dimensions-separated and the per-axis Position_0/1/2 followers carry the
// migrated X/Y/Z, writes PASS/FAIL + diagnostics to the .done file, resaves.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/separate_dims_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")"); // {input, done, resaved, x, y, z}

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  " + m); }

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
            var layer = comp.layer(1);
            var tg = layer.property("ADBE Transform Group");
            var pos = tg.property("ADBE Position");
            log.push("  dimensionsSeparated=" + pos.dimensionsSeparated);
            if (!pos.dimensionsSeparated) fail("dimensionsSeparated != true");

            var names = ["ADBE Position_0", "ADBE Position_1", "ADBE Position_2"];
            var wants = [args.x, args.y, args.z];
            for (var k = 0; k < 3; k++) {
                var f = tg.property(names[k]);
                if (!f) { fail(names[k] + " missing"); continue; }
                log.push("  " + names[k] + "=" + f.value);
                if (Math.abs(f.value - wants[k]) > 0.001) {
                    fail(names[k] + " value " + f.value + " != " + wants[k]);
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
