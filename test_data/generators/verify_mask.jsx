// Ship-gate verify for AddMask on the Mask Parade. Reads mask_args.json
// {input, done, resaved, effectCount, masks:[{name, closed, inverted, vertices:[[x,y]..]}]}:
//   masks = expected masks in parade order on the first layer with a Mask Parade.
//   effectCount = expected effect count on that layer (-1 to skip the check).
// Opens the Go-written input, reads each mask's name / maskMode / inverted /
// shape (closed + vertices, layer px, eps 0.5), compares to expectations,
// writes PASS/FAIL + diagnostics to .done, and resaves so the Go side can
// confirm AE kept the masks.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/mask_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

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
        var layer = null, parade = null;
        if (!comp) {
            fail("no CompItem in project");
        } else {
            for (var li = 1; li <= comp.numLayers; li++) {
                var p = comp.layer(li).property("ADBE Mask Parade");
                if (p && p.numProperties > 0) { layer = comp.layer(li); parade = p; break; }
            }
        }
        if (!parade) {
            fail("no layer with a non-empty Mask Parade");
        } else {
            log.push("  layer=" + layer.name + " masks=" + parade.numProperties);
            if (parade.numProperties !== args.masks.length) {
                fail("mask count " + parade.numProperties + " != expected " + args.masks.length);
            }
            var n = Math.min(parade.numProperties, args.masks.length);
            for (var mi = 0; mi < n; mi++) {
                var mask = parade.property(mi + 1);
                var want = args.masks[mi];
                log.push("  mask[" + mi + "] name=" + mask.name + " mode=" + mask.maskMode + " inverted=" + mask.inverted);
                if (mask.name !== want.name) fail("mask " + mi + " name " + mask.name + " != " + want.name);
                if (mask.maskMode !== MaskMode.ADD) fail("mask " + mi + " mode " + mask.maskMode + " != ADD");
                if (mask.inverted !== want.inverted) fail("mask " + mi + " inverted " + mask.inverted);
                var sh = mask.property("ADBE Mask Shape").value;
                log.push("  mask[" + mi + "] closed=" + sh.closed + " verts=" + sh.vertices.join(";"));
                if (sh.closed !== want.closed) fail("mask " + mi + " closed " + sh.closed + " != " + want.closed);
                if (sh.vertices.length !== want.vertices.length) {
                    fail("mask " + mi + " vertex count " + sh.vertices.length + " != " + want.vertices.length);
                } else {
                    for (var vi = 0; vi < sh.vertices.length; vi++) {
                        var dx = Math.abs(sh.vertices[vi][0] - want.vertices[vi][0]);
                        var dy = Math.abs(sh.vertices[vi][1] - want.vertices[vi][1]);
                        if (dx > 0.5 || dy > 0.5) {
                            fail("mask " + mi + " v" + vi + " [" + sh.vertices[vi] + "] != [" + want.vertices[vi] + "]");
                        }
                    }
                }
            }
            if (args.effectCount >= 0) {
                var fx = layer.property("ADBE Effect Parade");
                var fxCount = fx ? fx.numProperties : 0;
                log.push("  effects=" + fxCount);
                if (fxCount !== args.effectCount) fail("effect count " + fxCount + " != " + args.effectCount);
            }
        }

    } catch (e) {
        fail("EXC " + e.toString());
    }
    // Resave in its own try so a readback EXC above still produces the
    // resaved file (the Go side asserts on it for preservation).
    try {
        app.project.save(new File(args.resaved));
    } catch (e) {
        fail("EXC-save " + e.toString());
    }

    var done = new File(args.done);
    done.open("w");
    done.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n"));
    done.close();

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
