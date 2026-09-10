// showcase/text/render.jsx — open text.aep, render frame 0 → text.png.
// Also dumps each text layer's source string + font size to the .done log so the
// default text style is visible. Run via scripts/ae-worker/ae_run.ps1.
(function () {
    var dir = (new File($.fileName).parent.fsName + "/");
    var inAep = new File(dir + "text.aep");
    var png = new File(dir + "text.png");
    var done = new File(dir + "text.done");
    var log = [];
    var ok = false;
    try {
        app.open(inAep);
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "TextShowcase") { comp = it; break; }
        }
        if (!comp) throw new Error("comp not found");
        log.push("comp layers=" + comp.numLayers);
        for (var li = 1; li <= comp.numLayers; li++) {
            var L = comp.layer(li);
            try {
                var td = L.property("ADBE Text Properties").property("ADBE Text Document").value;
                log.push("  " + L.name + ": \"" + td.text + "\" size=" + td.fontSize);
            } catch (e1) { log.push("  " + L.name + ": (no text doc)"); }
        }
        app.project.bitsPerChannel = 8;
        comp.saveFrameToPng(0, png);
        $.sleep(2500);
        log.push(png.exists ? ("png " + png.length + " bytes") : "PNG MISSING");
        ok = png.exists;
    } catch (e) { log.push("ERROR: " + e.toString() + " line=" + e.line); }
    done.open("w");
    done.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
    done.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
