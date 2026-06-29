// Acceptance verify for DuplicateComposition. Reads comp_duplicate_args.json
// {input, done, resaved}. Opens the Go-built file and checks that the duplicate
// comp is a real, distinct CompItem with a deep-cloned layer list whose
// intra-comp parent ref was remapped to its OWN layer, and whose layer sources
// are SHARED with the source comp (not re-created). Then resaves for the Go side.
// No render — the op's surface is comp/layer structure + refs, read from the DOM.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/comp_duplicate_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    function note(m) { log.push("  " + m); }
    function near(a, b, eps) { return Math.abs(a - b) <= eps; }

    function namesOf(comp) {
        var out = [];
        for (var i = 1; i <= comp.numLayers; i++) out.push(comp.layer(i).name);
        return out;
    }

    try {
        app.open(new File(args.input));

        var orig = null, dup = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (!(it instanceof CompItem)) continue;
            if (it.name === "ORIG") orig = it;
            else if (it.name === "DUP") dup = it;
        }

        if (!orig) fail("ORIG comp missing");
        if (!dup) fail("DUP comp missing");

        if (orig && dup) {
            // Distinct items.
            if (orig.id === dup.id) fail("ORIG and DUP share id=" + orig.id + " (not a real duplicate)");
            else note("ORIG.id=" + orig.id + " DUP.id=" + dup.id);

            // Deep-cloned layer list, same names top-to-bottom.
            var on = namesOf(orig), dn = namesOf(dup);
            note("ORIG layers=[" + on.join(",") + "] DUP layers=[" + dn.join(",") + "]");
            if (on.join(",") !== "A,B") fail("ORIG layers=[" + on.join(",") + "], want [A,B]");
            if (dn.join(",") !== "A,B") fail("DUP layers=[" + dn.join(",") + "], want [A,B]");

            if (dup.numLayers === 2 && orig.numLayers === 2) {
                // Parent remap: in BOTH comps B(layer 2).parent must be that comp's
                // own A(layer 1) — proves the dup's ref was remapped, not left
                // pointing at the source comp's A.
                function checkParent(comp, label) {
                    var b = comp.layer(2); // B
                    if (b.name !== "B") { fail(label + " layer(2)=" + b.name + ", want B"); return; }
                    if (!b.parent) { fail(label + " B has no parent (expected A)"); return; }
                    note(label + " B.parent={name:" + b.parent.name + ",index:" + b.parent.index + ",comp:" + b.parent.containingComp.name + "}");
                    if (b.parent.name !== "A") fail(label + " B.parent.name=" + b.parent.name + ", want A");
                    if (b.parent.index !== 1) fail(label + " B.parent.index=" + b.parent.index + ", want 1");
                    if (b.parent.containingComp.name !== comp.name)
                        fail(label + " B.parent lives in comp " + b.parent.containingComp.name + ", want " + comp.name + " (remap failed)");
                }
                checkParent(orig, "ORIG");
                checkParent(dup, "DUP");

                // Source sharing: DUP's layers reference the SAME solid sources as
                // ORIG (CompItem.duplicate() shares sources, doesn't re-create).
                for (var li = 1; li <= 2; li++) {
                    var os = orig.layer(li).source, ds = dup.layer(li).source;
                    if (!os || !ds) { fail("layer " + li + " missing source (orig=" + os + " dup=" + ds + ")"); continue; }
                    note("layer " + li + " source id orig=" + os.id + " dup=" + ds.id);
                    if (os.id !== ds.id) fail("layer " + li + " source not shared: ORIG.id=" + os.id + " DUP.id=" + ds.id);
                }
            }

            // Comp settings deep-cloned.
            if (dup.width !== orig.width) fail("DUP width=" + dup.width + ", want " + orig.width);
            if (dup.height !== orig.height) fail("DUP height=" + dup.height + ", want " + orig.height);
            if (!near(dup.frameRate, orig.frameRate, 0.01)) fail("DUP frameRate=" + dup.frameRate + ", want " + orig.frameRate);
            if (!near(dup.duration, orig.duration, 0.001)) fail("DUP duration=" + dup.duration + ", want " + orig.duration);
            note("DUP settings w=" + dup.width + " h=" + dup.height + " fps=" + dup.frameRate + " dur=" + dup.duration);
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
