(function () {
    var base = "e:/projects/tools/aep-parser/test_data/";
    var log = [];
    function has(obj, k) { try { return (k in obj); } catch (e) { return "ERR:" + e; } }

    app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
    app.newProject();
    var pr = app.project;

    log.push("timecodeDefaultBase in project? " + has(pr, "timecodeDefaultBase"));
    try { log.push("  value=" + pr.timecodeDefaultBase); } catch (e) { log.push("  read ERR " + e); }
    log.push("transparencyGridThumbnails in project? " + has(pr, "transparencyGridThumbnails"));
    try { log.push("  value=" + pr.transparencyGridThumbnails); } catch (e) { log.push("  read ERR " + e); }

    // Dump all enumerable project keys for reference.
    var keys = [];
    for (var k in pr) keys.push(k);
    log.push("project keys: " + keys.join(","));

    var marker = new File(base + "probe_proj_props.done");
    marker.open("w");
    marker.write(log.join("\n"));
    marker.close();

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
