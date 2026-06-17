// Probe for the 3D-enable foundation: open a from-scratch shape layer whose
// Is3D bit (ldta 0x26 bit2) we flipped, and report exactly what AE makes of it
// — does AE accept it, treat it as 3D, and (the open RE question) materialize
// the 3D transform channels (Z position / Orientation / X·Y rotation) that a
// 2D from-scratch layer never carried? Reads 3d_enable_args.json {input, done,
// resaved}.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/3d_enable_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    function note(m) { log.push("  " + m); }

    function has(group, mn) {
        try { return group.property(mn) !== null; } catch (e) { return false; }
    }

    try {
        app.open(new File(args.input));

        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "3DENABLE") { comp = it; break; }
        }
        if (!comp) { fail("comp 3DENABLE not found"); }
        else {
            note("comp layers=" + comp.numLayers);
            var box = null;
            for (var li = 1; li <= comp.numLayers; li++) {
                if (comp.layer(li).name === "BOX") { box = comp.layer(li); break; }
            }
            if (!box) { fail("BOX missing"); }
            else {
                note("BOX threeDLayer=" + box.threeDLayer);
                if (box.threeDLayer !== true) fail("BOX not 3D (threeDLayer=" + box.threeDLayer + ")");

                var xf = box.property("ADBE Transform Group");
                var pos = xf.property("ADBE Position");
                var dims = pos ? pos.value.length : 0;
                note("Position dims=" + dims + " value=" + (pos ? pos.value.toString() : "nil"));
                if (dims !== 3) fail("Position not 3D (dims=" + dims + ") — AE did not materialize Z");
                var oriOK = has(xf, "ADBE Orientation"), rxOK = has(xf, "ADBE Rotate X"), ryOK = has(xf, "ADBE Rotate Y");
                var matOK = box.property("ADBE Material Options Group") !== null;
                note("has Orientation=" + oriOK + " RotateX=" + rxOK + " RotateY=" + ryOK + " matOption=" + matOK);
                if (!oriOK) fail("Orientation channel not materialized");
                if (!rxOK) fail("Rotate X channel not materialized");
                if (!ryOK) fail("Rotate Y channel not materialized");
                if (!matOK) fail("Material Options group not materialized");
            }
        }

        app.project.save(new File(args.resaved));
    } catch (e) {
        fail("EXC " + e.toString() + " line=" + e.line);
    }

    var done = new File(args.done);
    done.open("w");
    done.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n"));
    done.close();

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
