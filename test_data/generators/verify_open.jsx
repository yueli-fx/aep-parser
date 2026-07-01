// test_data/generators/verify_open.jsx
//
// Generic AE-open verifier for Phase 5 ship-gate bisection. Opens the input
// aep, reports PASS if no exception, FAIL + error message otherwise.
//
// args.json (fixed path: test_data/verify_open_args.json):
//   {"input":"<abs>","done":"<abs>"}

(function () {
    var argsPath = $.getenv("AE_OPEN_ARGS");
    var argsFile = new File(argsPath && argsPath.length > 0 ? argsPath : "e:/projects/tools/aep-parser/test_data/generated/args/verify_open_args.json");
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
            if (app.project) {
                app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
            }
        } catch (ePre) {}

        app.open(new File(args.input));
        log.push("project items.length=" + app.project.items.length);
        for (var i = 1; i <= app.project.items.length; i++) {
            var it = app.project.items[i];
            var t = "?";
            try { t = it.typeName; } catch (eT) {}
            log.push("  items[" + i + "]: name=" + it.name + " type=" + t);
        }
        var c = null;
        for (var j = 1; j <= app.project.items.length; j++) {
            if (app.project.items[j] instanceof CompItem) { c = app.project.items[j]; break; }
        }
        if (c) {
            log.push("comp=" + c.name + " layers.length=" + c.layers.length);
            for (var k = 1; k <= c.layers.length; k++) {
                log.push("  layer[" + k + "]: name=" + c.layers[k].name + " enabled=" + c.layers[k].enabled);
            }
        } else {
            log.push("NO CompItem in project");
        }
        ok = true;
    } catch (e) {
        ok = false;
        log.push("ERROR: " + e.toString());
    }

    try {
        if (!doneFile) doneFile = new File("e:/projects/tools/aep-parser/test_data/verify_open.done");
        doneFile.open("w");
        doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
        doneFile.close();
    } catch (e2) {}

    try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (eClose) {}
    try { app.quit(); } catch (eQuit) {}
})();
