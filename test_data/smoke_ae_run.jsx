// Minimal smoke for scripts/ae_run.ps1 -- no .aep IO, just .done write + quit.
// Reads $TEMP/ae_run_smoke.done path from args.json (Go-side ship-gate pattern).
(function () {
    var doneFile = new File("$DONE_PLACEHOLDER$");
    var log = ["PASS"];
    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        log.push("OK fresh_project");
    } catch (e) {
        log.push("ERR fresh_project -> " + e.toString());
    }

    try {
        doneFile.open("w");
        doneFile.write(log.join("\n"));
        doneFile.close();
    } catch (e) {
        // best effort
    }

    // tear down so AE exits cleanly
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
