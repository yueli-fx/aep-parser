// flightdeck/showcase/essential-graphics/verify.jsx — open essential_graphics.aep
// and dump the comp's Essential Graphics DOM (template name + controller count +
// each controller name) to .done. 📋 READBACK showcase: no png, the user reads the
// log. Run via scripts/ae_run.ps1.
(function () {
    var dir = "e:/projects/tools/aep-parser/flightdeck/showcase/essential-graphics/";
    var inAep = new File(dir + "essential_graphics.aep");
    var done = new File(dir + "essential_graphics.done");
    var log = [];
    var ok = false;
    try {
        app.open(inAep);
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "EssentialGraphics") { comp = it; break; }
        }
        if (!comp) throw new Error("comp not found");
        log.push("motionGraphicsTemplateName=" + comp.motionGraphicsTemplateName);
        var cnt = comp.motionGraphicsTemplateControllerCount;
        log.push("controllerCount=" + cnt);
        for (var k = 1; k <= cnt; k++) {
            var nm = null;
            try { nm = comp.getMotionGraphicsTemplateControllerName(k); }
            catch (e1) { try { nm = comp.motionGraphicsTemplateControllerName(k); } catch (e2) { nm = "(name API n/a)"; } }
            log.push("controller#" + k + "=" + nm);
        }
        ok = (comp.motionGraphicsTemplateName === "Showcase EG Template") && (cnt === 1);
    } catch (e) { log.push("ERROR: " + e.toString() + " line=" + e.line); }
    done.open("w");
    done.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
    done.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
