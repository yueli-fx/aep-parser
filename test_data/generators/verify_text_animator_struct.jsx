// Ship-gate verify for Text Animator structural ops (Remove / MoveTo /
// Duplicate) on the "ADBE Text Animators" indexed group. Reads
// text_animator_struct_args.json {input, done, resaved, expect}:
//   expect = array of the driven-leaf match-names AE should read back, in order,
//            one per animator (each test animator carries exactly one leaf:
//            Opacity / Skew / Fill Color), after Go applied the mutation.
// Opens the Go-written input, finds the TXT layer's Text Animators group, tags
// each animator by which driven leaf carries its distinctive non-default value
// (Opacity=50 / Skew=20 / Fill Color=blue — AE's ScriptingAPI exposes the FULL
// ~120-leaf schema under "ADBE Text Animator Properties", so property(1) is
// always "Anchor Point 3D"; the materialized leaf must be found by value), and
// compares the sequence to expect. Proves AE ACCEPTS the restructured group
// (no data-loss / corrupt) and kept the order; resaves for the Go side.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/text_animator_struct_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  " + m); }

    // Identify an animator by its single non-default driven leaf value.
    function tagOf(anim) {
        var props = anim.property("ADBE Text Animator Properties");
        if (!props) return "?";
        try {
            var op = props.property("ADBE Text Opacity");
            if (op && Math.abs(op.value - 50) < 0.5) return "ADBE Text Opacity";
        } catch (e) {}
        try {
            var sk = props.property("ADBE Text Skew");
            if (sk && Math.abs(sk.value - 20) < 0.5) return "ADBE Text Skew";
        } catch (e) {}
        try {
            var fc = props.property("ADBE Text Fill Color");
            if (fc) { var v = fc.value; if (v[2] > 0.5 && v[0] < 0.5) return "ADBE Text Fill Color"; }
        } catch (e) {}
        return "?";
    }

    try {
        app.open(new File(args.input));

        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem) { comp = it; break; }
        }
        if (!comp) {
            fail("no CompItem in project");
        } else {
            var txt = null;
            for (var li = 1; li <= comp.numLayers; li++) {
                if (comp.layer(li).name === "TXT") { txt = comp.layer(li); break; }
            }
            if (!txt) {
                fail("no TXT layer");
            } else {
                var animators = txt.property("ADBE Text Properties").property("ADBE Text Animators");
                if (!animators) {
                    fail("no Text Animators group");
                } else {
                    var got = [];
                    for (var k = 1; k <= animators.numProperties; k++) {
                        got.push(tagOf(animators.property(k)));
                    }
                    log.push("  animators=[" + got.join(", ") + "]");
                    log.push("  expect=[" + args.expect.join(", ") + "]");
                    if (got.length !== args.expect.length) {
                        fail("count " + got.length + " != expected " + args.expect.length);
                    } else {
                        for (var j = 0; j < got.length; j++) {
                            if (got[j] !== args.expect[j]) {
                                fail("index " + j + ": " + got[j] + " != " + args.expect[j]);
                            }
                        }
                    }
                }
            }
        }

        app.project.save(new File(args.resaved));
    } catch (e) {
        fail("EXC " + e.toString());
    }

    var done = new File(args.done);
    done.open("w");
    done.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n"));
    done.close();

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
