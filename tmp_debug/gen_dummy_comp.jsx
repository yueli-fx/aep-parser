// Generate a minimal 1-comp .aep saved by the running AE version.
// Output path read from args.json (so same JSX works across AE versions).
//
// args.json layout: {"out": "<abs path>", "done": "<abs path>"}
// Args fixed path: e:/projects/tools/aep-parser/tmp_debug/gen_dummy_args.json
//
// Comp specs: 1920x1080, 30 fps, 12s, 1 sec PAR — matches existing
// 2025_dummy_comp.aep so we can use as a target-version template substitute
// in NewComposition's builder.

(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/tmp_debug/gen_dummy_args.json");
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

        var outFile = new File(args.out);

        // Fresh project (防上轮残留)
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        log.push("ae.version=" + app.version);

        // Add 1 default comp matching AE-25 dummy template's C1
        var c = app.project.items.addComp("C1", 1920, 1080, 1, 12, 30);
        log.push("added C1: " + c.width + "x" + c.height + " fps=" + c.frameRate + " dur=" + c.duration);

        app.project.save(outFile);
        log.push("saved " + outFile.fsName);
        ok = true;
    } catch (e) {
        ok = false;
        log.push("ERROR: " + e.toString());
    }

    try {
        if (!doneFile) doneFile = new File("e:/projects/tools/aep-parser/tmp_debug/gen_dummy.done");
        doneFile.open("w");
        doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
        doneFile.close();
    } catch (e2) { /* swallow */ }
})();
