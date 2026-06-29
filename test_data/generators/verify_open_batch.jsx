// test_data/generators/verify_open_batch.jsx
//
// Batch variant of verify_open.jsx — one AE launch verifies N .aep files in
// sequence. Avoids the ~10s AE startup × N cost when bisecting variants.
//
// args.json (fixed path: test_data/verify_open_batch_args.json):
//   {"items": [
//     {"input": "<abs>", "done": "<abs>"},
//     ...
//   ]}
//
// Per item: opens the .aep, dumps project items + first CompItem's layers,
// writes (PASS|FAIL) + log to <done>. Closes project (DO_NOT_SAVE_CHANGES)
// between items so state doesn't leak. app.quit() at the very end (single
// process exit).

(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/verify_open_batch_args.json");
    var summary = new File("e:/projects/tools/aep-parser/test_data/generated/fixtures/verify_open_batch_summary.txt");
    var items = [];

    try {
        argsFile.open("r");
        var s = argsFile.read();
        argsFile.close();
        var parsed = eval("(" + s + ")");
        items = parsed.items || [];
    } catch (eArgs) {
        try {
            summary.open("w");
            summary.write("FAIL\nargs read error: " + eArgs.toString());
            summary.close();
        } catch (eSum) {}
        try { app.quit(); } catch (eQ) {}
        return;
    }

    for (var idx = 0; idx < items.length; idx++) {
        var entry = items[idx];
        var log = [];
        var ok = false;
        var doneFile = new File(entry.done);

        try {
            try {
                if (app.project) {
                    app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
                }
            } catch (ePre) {}

            app.open(new File(entry.input));
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
            doneFile.open("w");
            doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
            doneFile.close();
        } catch (eDone) {}

        try {
            if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        } catch (eClose) {}
    }

    try {
        summary.open("w");
        summary.write("DONE\n" + items.length + " items processed");
        summary.close();
    } catch (eSum2) {}

    try { app.quit(); } catch (eQuit) {}
})();
