// Acceptance verify for the from-scratch PROJECT settings. Reads
// project_settings_args.json {input, done, resaved}. Opens the Go-built file and
// checks each app.project DOM field against the NON-DEFAULT value the Go side
// set, then resaves for the Go side. No render — project settings have no
// rendered visual; DOM readback is their surface.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/project_settings_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function chk(name, got, want) {
        var pass = (got === want);
        if (!pass) ok = false;
        log.push("  " + name + "=" + got + " (want " + want + ") -> " + (pass ? "OK" : "NO"));
    }

    try {
        app.open(new File(args.input));
        var pr = app.project;

        chk("bitsPerChannel", pr.bitsPerChannel, 16);
        chk("timeDisplayType==FRAMES", pr.timeDisplayType === TimeDisplayType.FRAMES, true);
        chk("framesCountType==FC_START_1", pr.framesCountType === FramesCountType.FC_START_1, true);
        chk("displayStartFrame", pr.displayStartFrame, 1);
        chk("framesUseFeetFrames", pr.framesUseFeetFrames, true);
        chk("feetFramesFilmType==MM16", pr.feetFramesFilmType === FeetFramesFilmType.MM16, true);
        chk("footageTimecodeDisplayStartType==FTCS_USE_SOURCE_MEDIA",
            pr.footageTimecodeDisplayStartType === FootageTimecodeDisplayStartType.FTCS_USE_SOURCE_MEDIA, true);
        chk("linearBlending", pr.linearBlending, true);
        chk("expressionEngine", pr.expressionEngine, "javascript-1.0");

        app.project.save(new File(args.resaved));
    } catch (e) {
        ok = false;
        log.push("  EXC " + e.toString() + " line=" + e.line);
    }

    var done = new File(args.done);
    done.open("w");
    done.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n"));
    done.close();

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
