// Probe: dump fdta byte semantics across project topologies in AE 2025.
// Generates 5 fixtures: empty / 1 comp / 2 comp / 1 comp after delete / nested folder.
(function () {
    var outDir = "e:/projects/tools/aep-parser/test_data/fdta_probe/";
    var log = [];
    function save(name) {
        var f = new File(outDir + name);
        app.project.save(f);
        log.push("saved " + name);
    }
    function fresh() {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
    }

    try {
        fresh(); save("AE2025_empty.aep");

        fresh();
        app.project.items.addComp("C1", 1920, 1080, 1, 5, 30);
        save("AE2025_1comp.aep");

        fresh();
        app.project.items.addComp("C1", 1920, 1080, 1, 5, 30);
        app.project.items.addComp("C2", 1920, 1080, 1, 5, 30);
        save("AE2025_2comp.aep");

        fresh();
        var c1 = app.project.items.addComp("C1", 1920, 1080, 1, 5, 30);
        app.project.items.addComp("C2", 1920, 1080, 1, 5, 30);
        c1.remove();
        save("AE2025_1comp_after_delete.aep");

        fresh();
        var folder = app.project.items.addFolder("F1");
        app.project.items.addComp("InsideC", 1920, 1080, 1, 5, 30).parentFolder = folder;
        save("AE2025_nested.aep");
    } catch (e) {
        log.push("ERR " + e.toString());
    }

    var marker = new File("e:/projects/tools/aep-parser/test_data/fdta_probe_AE2025.done");
    marker.open("w");
    marker.write(log.join("\n"));
    marker.close();
})();
