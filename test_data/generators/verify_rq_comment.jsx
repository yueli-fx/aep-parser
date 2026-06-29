// test_data/generators/verify_rq_comment.jsx
//
// RenderQueueItem.SetComment ship-gate verifier. Opens a Go-built .aep whose
// first render-queue item had a comment written by SetComment (insert path:
// a fresh RCom chunk) and proves AE ACCEPTS it (no data-loss / corrupt), then
// resaves so the Go side can confirm AE PRESERVED the inserted RCom byte-wise.
//
// NOTE: RenderQueueItem has no `comment` property in the ScriptingAPI (any
// version) — it is a binary-only field (the panel "Comment" column). So this
// gate cannot read the value back via script; it verifies acceptance +
// resave-preservation instead. The `comment` probe below is logged for info
// (always undefined) but never gates.
//
// PASS = file opened and at least one render-queue item is present.
//
// args.json fixed path: test_data/rq_comment_args.json
//   {"input","done","resaved","expected"}

(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/rq_comment_args.json");
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

        try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (ePre) {}

        app.open(new File(args.input));
        var rq = app.project.renderQueue;
        log.push("numItems=" + rq.numItems);
        ok = (rq.numItems >= 1);
        if (ok) {
            var rqi = rq.item(1);
            log.push("comp=" + (rqi.comp ? rqi.comp.name : "(none)"));
            log.push("comment-probe=[" + rqi.comment + "]"); // always undefined (no API), info only
            log.push("expected=[" + (args.expected || "") + "]");
        } else {
            log.push("FAIL: render queue empty");
        }

        if (ok && args.resaved) {
            app.project.save(new File(args.resaved));
            log.push("resaved to " + args.resaved);
        }
    } catch (e) {
        ok = false;
        log.push("ERROR: " + e.toString());
    }

    try {
        if (!doneFile) doneFile = new File("e:/projects/tools/aep-parser/test_data/verify_rq_comment.done");
        doneFile.open("w");
        doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
        doneFile.close();
    } catch (e2) {}

    try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (eClose) {}
    try { app.quit(); } catch (eQuit) {}
})();
