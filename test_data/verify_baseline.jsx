// test_data/verify_baseline.jsx
//
// Per GPT's "epistemic hole" check (workshop/wip/gpt step 3):
// before continuing byte-level patching, verify our probe metric is sound by
// opening tolerance.aep through the SAME measurement pipeline. If
// tolerance.aep also reports layers.length=0, then our 6 iters of "silent
// drop" were measurement artifacts — not bugs in the file.
//
// args.json (fixed path: test_data/verify_baseline_args.json):
//   {"input": "<abs path to .aep>", "done": "<abs path to .done>"}
//
// Output to <done>:
//   PASS/FAIL
//   app.project.items.length=N
//   compCount=N
//   activeItem.name=X (or null/typeName)
//   selectionCount=N
//   for each item:
//     items[i]: name=X type=TypeName class=CtorName
//     if CompItem: layers.length=N
//       for each layer: name=X enabled=B layerType=...

(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/verify_baseline_args.json");
    var args = {};
    var doneFile = null;
    var log = [];
    var ok = false;

    function describeClass(o) {
        try {
            if (o === null || o === undefined) return "null";
            var c = o.constructor;
            if (c && c.name) return c.name;
            return o.toString();
        } catch (e) { return "?"; }
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
        } catch (ePre) {}

        app.open(new File(args.input));
        var proj = app.project;
        log.push("project.items.length=" + proj.items.length);

        // Active item / selection state — what AE considers "the current focus"
        try {
            var ai = proj.activeItem;
            log.push("activeItem=" + (ai ? ai.name + " (class=" + describeClass(ai) + ")" : "null"));
        } catch (eAI) { log.push("activeItem.ERROR: " + eAI.toString()); }
        try {
            var sel = proj.selection;
            log.push("selection.length=" + sel.length);
            for (var si = 0; si < sel.length; si++) {
                log.push("  sel[" + si + "]=" + sel[si].name + " (class=" + describeClass(sel[si]) + ")");
            }
        } catch (eSel) { log.push("selection.ERROR: " + eSel.toString()); }

        // Enumerate ALL items + their basic identity. CompItem gets layer enumeration.
        var compCount = 0;
        for (var i = 1; i <= proj.items.length; i++) {
            var it = proj.items[i];
            var t = "?", cls = "?";
            try { t = it.typeName; } catch (eT) {}
            cls = describeClass(it);
            var line = "items[" + i + "]: name=" + it.name + " type=" + t + " class=" + cls + " id=" + it.id;
            log.push(line);
            if (it instanceof CompItem) {
                compCount++;
                log.push("  comp.layers.length=" + it.layers.length);
                log.push("  comp.duration=" + it.duration + "s  frameRate=" + it.frameRate);
                for (var k = 1; k <= it.layers.length; k++) {
                    var ly = it.layers[k];
                    var lyCls = describeClass(ly);
                    var lyType = "?";
                    try { lyType = ly.matchName; } catch (eLN) {}
                    log.push("    layer[" + k + "]: name=" + ly.name + " class=" + lyCls + " enabled=" + ly.enabled + " index=" + ly.index);
                }
            }
        }
        log.push("compCount=" + compCount);
        ok = true;
    } catch (e) {
        ok = false;
        log.push("ERROR: " + e.toString());
    }

    try {
        if (!doneFile) doneFile = new File("e:/projects/tools/aep-parser/test_data/verify_baseline.done");
        doneFile.open("w");
        doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
        doneFile.close();
    } catch (e2) {}

    try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (eClose) {}
    try { app.quit(); } catch (eQuit) {}
})();
