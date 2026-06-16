// Acceptance verify for the text-styling render gate. Reads
// mg_text_style_args.json {done, ver, jobs:[{name, input, png}]}.
//
// Mirrors the proven showcase render.jsx pattern exactly: one fresh AE per job
// (the Go harness launches AE once per knob), open the single-comp project,
// saveFrameToPng(0) its sole comp, $.sleep to let the async write land, check.
// No app.purge / unique-time / openInViewer — those were found to PRODUCE stale
// frames; the plain showcase pattern renders correctly.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/mg_text_style_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    function note(m) { log.push("  " + m); }

    try {
        for (var j = 0; j < args.jobs.length; j++) {
            var job = args.jobs[j];

            app.open(new File(job.input));
            app.project.bitsPerChannel = 8;

            var comp = null;
            for (var i = 1; i <= app.project.numItems; i++) {
                var it = app.project.item(i);
                if (it instanceof CompItem && it.name.indexOf("TS_") === 0) { comp = it; break; }
            }
            if (!comp) { fail(job.name + ": comp not found"); continue; }

            // DOM readback (diagnostic — Go measures pixels, not these)
            var tl = null;
            for (var li = 1; li <= comp.numLayers; li++) {
                try { if (comp.layer(li).property("ADBE Text Properties")) { tl = comp.layer(li); break; } } catch (e) {}
            }
            if (tl) {
                var doc = tl.property("ADBE Text Properties").property("ADBE Text Document").value;
                note(job.name + " size=" + doc.fontSize + " trk=" + doc.tracking +
                     " just=" + doc.justification + " lead=" + doc.leading + "(auto " + doc.autoLeading + ")");
            }
            // Diagnostic: comp internal id + dims + render time. If ids collide
            // across separate single-comp projects, AE's disk frame cache (keyed
            // by comp-id+time) serves a stale frame — the cross-contamination bug.
            note(job.name + " comp.id=" + comp.id + " " + comp.width + "x" + comp.height + " t=" + job.time);

            var png = new File(job.png);
            if (png.exists) png.remove();
            comp.saveFrameToPng(job.time, png);
            $.sleep(2500);
            if (!png.exists) fail(job.name + " PNG MISSING");
            else note(job.name + " png " + png.length + " bytes");
        }
    } catch (e) { fail("EXC " + e.toString() + " line=" + e.line); }

    var done = new File(args.done);
    done.open("w");
    done.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n"));
    done.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
