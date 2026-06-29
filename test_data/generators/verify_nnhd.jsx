(function () {
    var base = "e:/projects/tools/aep-parser/test_data/";
    var log = [];
    function chk(name, got, want) {
        var ok = (String(got) === String(want));
        log.push((ok ? "OK  " : "NO  ") + name + " got=" + got + " want=" + want);
    }

    try {
        var pf = new File(base + "nnhd_go_verify.aep");
        app.open(pf);
        var pr = app.project;
        chk("timeDisplayType", pr.timeDisplayType.toString(), TimeDisplayType.FRAMES.toString());
        chk("framesCountType", pr.framesCountType.toString(), FramesCountType.FC_START_0.toString());
        chk("framesUseFeetFrames", pr.framesUseFeetFrames, true);
        chk("feetFramesFilmType", pr.feetFramesFilmType.toString(), FeetFramesFilmType.MM16.toString());
        chk("transparencyGridThumbnails", pr.transparencyGridThumbnails, true);
    } catch (e) {
        log.push("ERR open/readback -> " + e.toString());
    }

    var marker = new File(base + "verify_nnhd.done");
    marker.open("w");
    marker.write(log.join("\n"));
    marker.close();

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
