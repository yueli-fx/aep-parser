// Effect-library AUDIO wave: probe + capture the audio-processing effects (which
// can only attach to a layer that HAS audio — hence an imported mp3, not a solid).
// Imports test_data/generated/fixtures/_audio_probe.mp3, adds it to a comp as a layer, then sweeps a
// candidate match-name list with canAddProperty/addProperty, logging the STORED
// match-name for each, and saves re_effect_audio.aep for tmp_debug/extract_effect_lib.
(function () {
    var dir = "e:/projects/tools/aep-parser/test_data/";
    var mp3 = dir + "_audio_probe.mp3";
    var outFile = new File(dir + "re_effect_audio.aep");
    var log = [];

    var wanted = [
        "ADBE Aud Reverse",        // Backwards
        "ADBE Aud BT",             // Bass & Treble
        "ADBE Aud Delay",          // Delay
        "ADBE Aud_Flange",         // Flange & Chorus
        "ADBE Aud Flange",
        "ADBE Aud HiLo",           // High-Low Pass
        "ADBE Aud HighLow",
        "ADBE Aud Modulator",      // Modulator
        "ADBE Param EQ",           // Parametric EQ
        "ADBE Aud Parametric EQ",
        "ADBE Aud Reverb",         // Reverb
        "ADBE Aud Stereo Mixer",   // Stereo Mixer
        "ADBE Aud Tone",           // Tone
        "ADBE Aud Spectrum",
        "ADBE Backwards"
    ];

    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        var io = new ImportOptions(new File(mp3));
        var foot = app.project.importFile(io);
        log.push("imported " + foot.name + " hasAudio=" + foot.hasAudio + " dur=" + foot.duration);
        var comp = app.project.items.addComp("efxaudio", 1920, 1080, 1, 5, 30);
        var aud = comp.layers.add(foot);
        log.push("audio layer '" + aud.name + "' hasAudio=" + aud.hasAudio);
        var parade = aud.property("ADBE Effect Parade");
        for (var i = 0; i < wanted.length; i++) {
            var mn = wanted[i];
            var canAdd = false;
            try { canAdd = parade.canAddProperty(mn); } catch (e) { canAdd = "?"; }
            if (canAdd !== true) { log.push("skip " + mn + " (canAdd=" + canAdd + ")"); continue; }
            try {
                var fx = parade.addProperty(mn);
                log.push("OK   " + mn + " -> stored=" + fx.matchName);
            } catch (e2) {
                log.push("FAIL " + mn + " (canAdd=" + canAdd + ") " + e2.toString());
            }
        }
        log.push("parade numProps=" + parade.numProperties);
        for (var j = 1; j <= parade.numProperties; j++) {
            log.push("  [" + j + "] " + parade.property(j).matchName + "  name=" + parade.property(j).name);
        }
        app.project.save(outFile);
        log.push("saved " + outFile.fsName);
    } catch (e3) {
        log.push("EXC " + e3.toString() + " line=" + e3.line);
    }

    var marker = new File(dir + "re_effect_audio.done");
    marker.open("w");
    marker.write(log.join("\n"));
    marker.close();

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
