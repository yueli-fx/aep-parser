// test_data/verify_open_multicomp.jsx
//
// Opens the demo aep and reports every CompItem with its layer count, so we can
// confirm AE keeps each comp's single ShapeLayer (no silent drop). PASS only if
// no exception AND every CompItem has layers.length >= 1.
//
// args.json (fixed path: test_data/verify_open_args.json):
//   {"input":"<abs>","done":"<abs>"}

(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/verify_open_args.json");
    var args = {};
    var doneFile = null;
    var log = [];
    var ok = false;

    try {
        argsFile.open("r");
        var s = argsFile.read();
        argsFile.close();
        args = eval("(" + s + ")");
        doneFile = new File(args.done);

        try {
            if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        } catch (ePre) {}

        app.open(new File(args.input));
        log.push("project items.length=" + app.project.items.length);

        var comps = 0, emptyComps = 0;
        for (var i = 1; i <= app.project.items.length; i++) {
            var it = app.project.items[i];
            if (it instanceof CompItem) {
                comps++;
                log.push("comp[" + comps + "] name=" + it.name + " layers=" + it.numLayers);
                if (it.numLayers < 1) emptyComps++;
            }
        }
        log.push("TOTAL comps=" + comps + " emptyComps=" + emptyComps);
        ok = (comps > 0 && emptyComps === 0);
    } catch (e) {
        ok = false;
        log.push("ERROR: " + e.toString());
    }

    try {
        if (!doneFile) doneFile = new File("e:/projects/tools/aep-parser/test_data/verify_open_multicomp.done");
        doneFile.open("w");
        doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
        doneFile.close();
    } catch (e2) {}

    try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (eClose) {}
    try { app.quit(); } catch (eQuit) {}
})();
