// showcase/project-settings/verify.jsx — open project_settings.aep and
// dump app.project DOM settings to the .done log with an explicit OK/NO check vs
// the expected value. 📋 READBACK showcase: no png, the user reads the log.
// Run via scripts/ae-worker/ae_run.ps1.
(function () {
    var dir = (new File($.fileName).parent.fsName + "/");
    var inAep = new File(dir + "project_settings.aep");
    var done = new File(dir + "project_settings.done");
    var log = [];
    var allOK = true;
    function chk(name, got, want) {
        var pass = (got === want);
        if (!pass) allOK = false;
        log.push(name + "=" + got + " (expect " + want + ") -> " + (pass ? "OK" : "NO"));
    }
    try {
        app.open(inAep);
        var pr = app.project;
        chk("bitsPerChannel", pr.bitsPerChannel, 16);
        chk("linearBlending", pr.linearBlending, true);
        chk("expressionEngine", pr.expressionEngine, "javascript-1.0");
        chk("footageTimecodeDisplayStartType==USE_SOURCE_MEDIA",
            pr.footageTimecodeDisplayStartType === FootageTimecodeDisplayStartType.FTCS_USE_SOURCE_MEDIA, true);
        chk("timeDisplayType==TIMECODE", pr.timeDisplayType === TimeDisplayType.TIMECODE, true);
        chk("framesCountType==FC_START_0", pr.framesCountType === FramesCountType.FC_START_0, true);
        chk("framesUseFeetFrames", pr.framesUseFeetFrames, true);
        chk("feetFramesFilmType==MM35", pr.feetFramesFilmType === FeetFramesFilmType.MM35, true);
    } catch (e) { log.push("ERROR: " + e.toString() + " line=" + e.line); allOK = false; }
    done.open("w");
    done.write((allOK ? "PASS\n" : "FAIL\n") + log.join("\n"));
    done.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
