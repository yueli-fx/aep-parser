// Ship-gate verify for RenderQueue.RemoveItem. Opens the Go-written file (one
// item removed) and asserts AE reads exactly one render-queue item whose comp
// matches the expected survivor. Reads rq_delete_args.json {input, done,
// resaved, expectItems, survivorComp}. Writes PASS/FAIL + diagnostics, resaves.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/rq_delete_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  " + m); }

    try {
        app.open(new File(args.input));
        var rq = app.project.renderQueue;
        log.push("  numItems=" + rq.numItems);
        if (rq.numItems !== args.expectItems) {
            fail("numItems " + rq.numItems + " != " + args.expectItems);
        }
        if (rq.numItems >= 1) {
            var c = rq.item(1).comp;
            log.push("  item(1).comp=" + (c ? c.name : "null"));
            if (args.survivorComp && (!c || c.name !== args.survivorComp)) {
                fail("survivor comp " + (c ? c.name : "null") + " != " + args.survivorComp);
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
