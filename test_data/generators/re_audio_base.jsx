// Minimal base fixture for the audio-effect ship-gate: an audio layer (imported
// mp3) with ZERO effects, so the gate can AddEffect each audio effect via Go and
// prove AE accepts the spliced result. No effects here on purpose.
(function () {
    var dir = "e:/projects/tools/aep-parser/test_data/";
    var out = new File(dir + "re_audio_base.aep");
    var log = [];
    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        var io = new ImportOptions(new File(dir + "_audio_probe.mp3"));
        var foot = app.project.importFile(io);
        var comp = app.project.items.addComp("audbase", 1920, 1080, 1, 5, 30);
        var aud = comp.layers.add(foot);
        log.push("audio layer '" + aud.name + "' hasAudio=" + aud.hasAudio + " index=" + aud.index);
        app.project.save(out);
        log.push("saved " + out.fsName);
    } catch (e) { log.push("EXC " + e.toString() + " line=" + e.line); }
    var m = new File(dir + "re_audio_base.done"); m.open("w"); m.write(log.join("\n")); m.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
