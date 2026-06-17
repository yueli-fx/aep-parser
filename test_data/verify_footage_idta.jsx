// Acceptance verify for FOOTAGE item-level setters (SetComment / SetLabel).
// Reads footage_idta_args.json {input, done, comment, label}. Opens the Go-built
// file, finds the single solid FootageItem (by instanceof, NOT by name — a
// from-scratch solid's Item-level Utf8 is empty, so its DOM name is unreliable),
// and checks footage.comment / footage.label via the DOM. No resave (would pop
// AE's Save dialog), no render — metadata fields.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/footage_idta_args.json");
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

        var footages = [];
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof FootageItem) footages.push(it);
        }

        if (footages.length !== 1) {
            ok = false;
            log.push("  FAIL: expected exactly 1 FootageItem, got " + footages.length);
        } else {
            var f = footages[0];
            log.push("  footage name=" + f.name);
            chk("footage.comment", f.comment, args.comment);
            chk("footage.label", f.label, args.label);
        }
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
