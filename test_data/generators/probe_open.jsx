// Generic open-probe: opens the .aep named in probe_open_args.json {input},
// reports comp count (or the open error) to probe_open.done. Used to bisect an
// AE open-reject without the full verify logic.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/probe_open_args.json");
    argsFile.open("r"); var raw = argsFile.read(); argsFile.close();
    var args = eval("(" + raw + ")");
    var out = [];
    try {
        app.open(new File(args.input));
        var n = 0;
        for (var i = 1; i <= app.project.numItems; i++) {
            if (app.project.item(i) instanceof CompItem) n++;
        }
        out.push("OPENED comps=" + n + " items=" + app.project.numItems);
    } catch (e) {
        out.push("OPEN-ERR " + e.toString());
    }
    var done = new File("e:/projects/tools/aep-parser/test_data/probe_open.done");
    done.open("w"); done.write("PASS\n" + out.join("\n")); done.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
