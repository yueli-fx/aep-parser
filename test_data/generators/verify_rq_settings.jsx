// Acceptance verify for the render-queue settings batch. Reads rq_settings_args.json
// {input, done, resaved}. Opening the input at all proves AE ACCEPTS the file with
// all ~38 RQ settings set to non-defaults (no corrupt / data-loss dialog). Then
// resaves so the Go side can confirm byte-level preservation. RQ binary settings
// have no reliable cross-version ScriptingAPI readback, so acceptance + resave-
// preservation (Go-side) is the verification — same model as verify_rq_comment.jsx.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/rq_settings_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }

    try {
        app.open(new File(args.input));
        if (app.project.renderQueue.numItems < 1) fail("renderQueue.numItems < 1 after open");
        else log.push("  renderQueue.numItems=" + app.project.renderQueue.numItems);
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
