// test_data/generators/verify_flame_demo.jsx — open the Go-built flame + render a frame to
// PNG for Go-side pixel checks. Mirrors verify_orbit_demo.jsx's proven render
// mechanism: force 8bpc (32bpc float saveFrameToPng emits dark linear pixels) +
// $.sleep so the async PNG write flushes before app.quit races it. Renders the
// frame at the time given in args.t (seconds).
(function () {
    var a = new File("e:/projects/tools/aep-parser/test_data/generated/args/flame_demo_args.json");
    a.open("r"); var args = eval("(" + a.read() + ")"); a.close();
    var log = []; var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    function note(m) { log.push("  " + m); }
    try {
        app.open(new File(args.input));
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "FLAME") { comp = it; break; }
        }
        if (!comp) { fail("comp FLAME not found"); }
        else {
            note("comp " + comp.width + "x" + comp.height + " " + comp.duration + "s layers=" + comp.numLayers);
            app.project.bitsPerChannel = 8;
            var t = (args.t !== undefined) ? args.t : comp.duration / 2;
            comp.saveFrameToPng(t, new File(args.png));
            $.sleep(2000);
            var pngFile = new File(args.png);
            if (pngFile.exists) { note("rendered t=" + t + " -> " + pngFile.length + " bytes"); }
            else { fail("saveFrameToPng wrote no file"); }
            // optional second frame (motion check)
            if (args.png2 !== undefined && args.t2 !== undefined) {
                comp.saveFrameToPng(args.t2, new File(args.png2));
                $.sleep(2000);
                if (!(new File(args.png2)).exists) fail("saveFrameToPng wrote no png2");
                else note("rendered t2=" + args.t2);
            }
        }
    } catch (e) { fail("EXC " + e.toString() + " line=" + e.line); }
    var d = new File(args.done); d.open("w"); d.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n")); d.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
