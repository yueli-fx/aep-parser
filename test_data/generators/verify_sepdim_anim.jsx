// Ship-gate verify for Property.SetDimensionsSeparated on an ANIMATED Position.
// Reads sepdim_anim_args.json {input, done, resaved, mode, n, times[], vx[], vy[], vz[]}:
//   mode "sep"   -> leader dimensionsSeparated=true; Position_0/1/2 each animated
//                   with n keyframes whose time/value match times[]/vx,vy,vz[].
//   mode "merge" -> leader dimensionsSeparated=false; Position animated with n
//                   keyframes whose time/value == times[]/[vx,vy,vz].
// Proves AE ACCEPTS our restructured keyframe streams (no data-loss / corrupt)
// and reads the expected animated values. Writes PASS/FAIL + resaves.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/sepdim_anim_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL " + m); }
    function near(a, b) { return Math.abs(a - b) <= 0.01; }

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
            log.push("  mode=" + args.mode + " n=" + args.n + " dimensionsSeparated=" + pos.dimensionsSeparated);

            if (args.mode === "sep") {
                if (!pos.dimensionsSeparated) fail("dimensionsSeparated != true");
                var axes = [
                    { name: "ADBE Position_0", v: args.vx },
                    { name: "ADBE Position_1", v: args.vy },
                    { name: "ADBE Position_2", v: args.vz }
                ];
                for (var a = 0; a < axes.length; a++) {
                    var f = tg.property(axes[a].name);
                    if (!f) { fail(axes[a].name + " MISSING"); continue; }
                    if (f.numKeys !== args.n) { fail(axes[a].name + " numKeys=" + f.numKeys + " != " + args.n); continue; }
                    for (var k = 1; k <= args.n; k++) {
                        if (!near(f.keyTime(k), args.times[k - 1])) fail(axes[a].name + " k" + k + " time=" + f.keyTime(k));
                        if (!near(f.keyValue(k), axes[a].v[k - 1])) fail(axes[a].name + " k" + k + " value=" + f.keyValue(k) + " != " + axes[a].v[k - 1]);
                    }
                }
            } else { // "merge"
                if (pos.dimensionsSeparated) fail("dimensionsSeparated != false");
                if (pos.numKeys !== args.n) fail("leader numKeys=" + pos.numKeys + " != " + args.n);
                for (var k2 = 1; k2 <= args.n; k2++) {
                    var v = pos.keyValue(k2);
                    if (!near(pos.keyTime(k2), args.times[k2 - 1])) fail("leader k" + k2 + " time=" + pos.keyTime(k2));
                    if (!near(v[0], args.vx[k2 - 1]) || !near(v[1], args.vy[k2 - 1]) || !near(v[2], args.vz[k2 - 1])) {
                        fail("leader k" + k2 + " value=" + v.toString());
                    }
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
