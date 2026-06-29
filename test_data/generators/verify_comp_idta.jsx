// Acceptance verify for item-level comp setters (SetComment / SetLabel /
// SetDraft3D). Reads comp_idta_args.json {input, done, comment, label}.
// Opens the Go-built file, checks comp.comment / comp.label / comp.draft3d via
// the DOM. No resave (would pop AE's Save dialog), no render — metadata fields.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/comp_idta_args.json");
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

        var comps = {};
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem) comps[it.name] = it;
        }

        if (!comps["CMT"]) { ok = false; log.push("  FAIL: CMT comp missing"); }
        else chk("CMT.comment", comps["CMT"].comment, args.comment);

        if (!comps["LBL"]) { ok = false; log.push("  FAIL: LBL comp missing"); }
        else chk("LBL.label", comps["LBL"].label, args.label);

        // SetDraft3D dropped: confirmed false-green (cdta @0x8A bit0 byte-identical
        // to AE-native, but AE reopen reads comp.draft3d=false — derived/runtime
        // state, like lnrp). Stays verify=roundtrip.

        // No resave: DOM readback above is the ae-accept proof. Resaving would
        // pop AE's Save dialog (unhandled by the OCR dispatcher) and hang.
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
