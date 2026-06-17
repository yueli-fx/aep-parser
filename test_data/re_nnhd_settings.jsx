(function () {
    var base = "e:/projects/tools/aep-parser/test_data/";
    var log = [];
    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }

    // Probe defaults of a fresh AE 2020 project.
    step("defaults", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        log.push("  default timeDisplayType=" + app.project.timeDisplayType.toString());
        log.push("  default framesCountType=" + app.project.framesCountType.toString());
        log.push("  default feetFramesFilmType=" + app.project.feetFramesFilmType.toString());
        log.push("  default framesUseFeetFrames=" + app.project.framesUseFeetFrames);
    });

    // For each field, save the two enum values from identical set+save flows so
    // a variant-vs-variant byte diff isolates exactly the storage location.
    function saveWith(file, fn) {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        app.project.save(new File(base + file));
        fn();
        app.project.save();
    }

    step("time_timecode", function () {
        saveWith("re_nnhd_time_tc.aep", function () {
            app.project.timeDisplayType = TimeDisplayType.TIMECODE;
            log.push("  set->readback timeDisplayType=" + app.project.timeDisplayType.toString());
        });
    });
    step("time_frames", function () {
        saveWith("re_nnhd_time_fr.aep", function () {
            app.project.timeDisplayType = TimeDisplayType.FRAMES;
            log.push("  set->readback timeDisplayType=" + app.project.timeDisplayType.toString());
        });
    });

    step("frames_start0", function () {
        saveWith("re_nnhd_fc_s0.aep", function () {
            app.project.framesCountType = FramesCountType.FC_START_0;
            log.push("  set->readback framesCountType=" + app.project.framesCountType.toString());
        });
    });
    step("frames_start1", function () {
        saveWith("re_nnhd_fc_s1.aep", function () {
            app.project.framesCountType = FramesCountType.FC_START_1;
            log.push("  set->readback framesCountType=" + app.project.framesCountType.toString());
        });
    });

    step("feet_mm35", function () {
        saveWith("re_nnhd_feet_mm35.aep", function () {
            app.project.framesUseFeetFrames = true;
            app.project.feetFramesFilmType = FeetFramesFilmType.MM35;
            log.push("  set->readback feetFramesFilmType=" + app.project.feetFramesFilmType.toString());
        });
    });
    step("feet_mm16", function () {
        saveWith("re_nnhd_feet_mm16.aep", function () {
            app.project.framesUseFeetFrames = true;
            app.project.feetFramesFilmType = FeetFramesFilmType.MM16;
            log.push("  set->readback feetFramesFilmType=" + app.project.feetFramesFilmType.toString());
        });
    });

    step("write_done_marker", function () {
        var marker = new File(base + "re_nnhd_settings.done");
        marker.open("w");
        marker.write(log.join("\n"));
        marker.close();
    });

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
