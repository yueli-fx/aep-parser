(function () {
    var log = [];
    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }
    step("redirect_save", function () {
        app.project.save(new File("e:/projects/tools/aep-parser/test_data/fixtures/re_wave2_ae24.aep"));
    });
    step("dump_dropframe", function () {
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem) {
                log.push("  " + it.name + " dropFrame=" + it.dropFrame + " displayStartFrame=" + it.displayStartFrame + " displayStartTime=" + it.displayStartTime);
            }
        }
    });
    var marker = new File("e:/projects/tools/aep-parser/test_data/verify_dropframe.done");
    marker.open("w");
    marker.write(log.join("\n"));
    marker.close();
})();
