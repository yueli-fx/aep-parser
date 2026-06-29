// Ship-gate verify for RenderQueue.AddItem. Opens the Go-written file (one item
// cloned+appended for a new comp) and asserts AE reads expectItems items and the
// last item's comp matches addedComp. Reads rq_add_args.json {input, done,
// resaved, expectItems, addedComp}. Writes PASS/FAIL + diagnostics, resaves.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/rq_add_args.json");
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
        if (rq.numItems >= args.expectItems) {
            var c = rq.item(args.expectItems).comp;
            log.push("  item(" + args.expectItems + ").comp=" + (c ? c.name : "null"));
            if (args.addedComp && (!c || c.name !== args.addedComp)) {
                fail("added comp " + (c ? c.name : "null") + " != " + args.addedComp);
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
