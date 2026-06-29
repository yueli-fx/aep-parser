// Authors test_data/fixtures/re_audio_levels.aep: an audio layer (imported mp3) whose
// Audio Levels property is MATERIALIZED to a non-default value ([3,3] dB). A
// from-scratch / default audio layer ELIDES Audio Levels (default-omission), so
// the scene Layer.SetAudioLevels setter — which mutates an existing property —
// has nothing to write. This carrier gives it a present property to mutate, and
// the non-default author value (≠ the gate's target) means a no-op mutate would
// be caught. Run with AE2020 (low version → openable by both AE2020 and AE2025).
// Reuses the wave-10 audio carrier mp3 (test_data/generated/fixtures/_audio_probe.mp3).
(function () {
    var dir = "e:/projects/tools/aep-parser/test_data/";
    var out = new File(dir + "re_audio_levels.aep");
    var log = [];
    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        var io = new ImportOptions(new File(dir + "_audio_probe.mp3"));
        var foot = app.project.importFile(io);
        var comp = app.project.items.addComp("AUDLV", 1920, 1080, 1, 5, 30);
        var aud = comp.layers.add(foot);
        // Materialize Audio Levels at a non-default placeholder.
        var lv = aud.property("ADBE Audio Group").property("ADBE Audio Levels");
        lv.setValue([3, 3]);
        log.push("audio layer '" + aud.name + "' hasAudio=" + aud.hasAudio +
                 " levels=[" + lv.value[0] + "," + lv.value[1] + "]");
        app.project.save(out);
        log.push("saved " + out.fsName);
    } catch (e) { log.push("EXC " + e.toString() + " line=" + e.line); }
    var m = new File(dir + "re_audio_levels.done"); m.open("w"); m.write(log.join("\n")); m.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
