(function () {
    var base = "e:/projects/tools/aep-parser/test_data/";
    var log = [];
    try {
        app.open(new File(base + "resfactor_go_verify.aep"));
        // find first comp
        var c = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            if (app.project.item(i) instanceof CompItem) { c = app.project.item(i); break; }
        }
        if (!c) { log.push("NO comp found"); }
        else {
            var rf = c.resolutionFactor; // numeric indexing only, never "" + rf
            var ok = (rf[0] === 2 && rf[1] === 2);
            log.push((ok ? "OK  " : "NO  ") + "resolutionFactor got=" + rf[0] + "x" + rf[1] + " want=2x2");
        }
    } catch (e) { log.push("ERR " + e.toString()); }

    var marker = new File(base + "verify_resfactor.done");
    marker.open("w");
    marker.write(log.join("\n"));
    marker.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
