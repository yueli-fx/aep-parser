(function () {
    var base = "e:/projects/tools/aep-parser/test_data/";
    var log = [];
    function saveWith(file, fn) {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        app.project.save(new File(base + file));
        fn();
        app.project.save();
    }
    try {
        saveWith("re_transp_off.aep", function () {
            app.project.transparencyGridThumbnails = false;
            log.push("off readback=" + app.project.transparencyGridThumbnails);
        });
        saveWith("re_transp_on.aep", function () {
            app.project.transparencyGridThumbnails = true;
            log.push("on readback=" + app.project.transparencyGridThumbnails);
        });
    } catch (e) { log.push("ERR " + e); }

    var marker = new File(base + "re_transparency.done");
    marker.open("w");
    marker.write(log.join("\n"));
    marker.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
