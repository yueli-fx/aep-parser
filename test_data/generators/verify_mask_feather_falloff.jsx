// Acceptance verify for SetFeatherFalloff (mask mkif @0x03). Reads
// mask_falloff_args.json {input, done, resaved, comp, layer}. Proves AE reads
// back the Go-written mask's maskFeatherFalloff as FFO_LINEAR (coincidence-proof
// that @0x03 is the falloff byte, not just a value round-trip), resaves it.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/mask_falloff_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    function note(m) { log.push("  " + m); }

    try {
        app.open(new File(args.input));

        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === args.comp) { comp = it; break; }
        }
        if (!comp) { fail("comp " + args.comp + " not found"); }
        else {
            var card = null;
            for (var li = 1; li <= comp.numLayers; li++) if (comp.layer(li).name === args.layer) card = comp.layer(li);
            if (!card) { fail("layer " + args.layer + " missing"); }
            else {
                var mk = card.property("ADBE Mask Parade").property(1);
                if (!mk) { fail("no mask"); }
                else {
                    var ffo = mk.maskFeatherFalloff;
                    var wantLinear = MaskFeatherFalloff.FFO_LINEAR;
                    note("maskFeatherFalloff=" + ffo + " (FFO_LINEAR=" + wantLinear + " FFO_SMOOTH=" + MaskFeatherFalloff.FFO_SMOOTH + ")");
                    if (ffo !== wantLinear) fail("maskFeatherFalloff=" + ffo + ", want FFO_LINEAR (" + wantLinear + ")");
                }
            }
        }

        app.project.save(new File(args.resaved));
    } catch (e) {
        fail("EXC " + e.toString() + " line=" + e.line);
    }

    var done = new File(args.done);
    done.open("w");
    done.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n"));
    done.close();

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
