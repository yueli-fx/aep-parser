// Template for AE-script-driven RE work.
//
// IMPORTANT: AE has NO app.project.saveAs() — use app.project.save(file)
// where file is a File object. If file is omitted, save() saves to the
// current project location.
//
// To not pollute the user's current .aep, this template:
//   1. Calls app.project.save(outFile) first → redirects future saves
//      to outFile, leaving the user's original .aep on disk untouched.
//   2. Adds the test data in a brand-new comp (doesn't disturb user comps).
//   3. Each step is wrapped in its own try/catch and reported via alert.
//
// Adjust the body of step("test_data", ...) to add whatever test data
// you need for the next RE target.
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/re_out.aep");
    var log = [];
    function step(name, fn) {
        try {
            fn();
            log.push("OK  " + name);
        } catch (e) {
            log.push("ERR " + name + " -> " + e.toString());
        }
    }

    step("save_as_new_file", function () {
        app.project.save(outFile);
    });

    step("add_test_comp", function () {
        app.testComp = app.project.items.addComp("RE_TEST", 200, 200, 1, 5, 30);
    });

    step("test_data", function () {
        // ── Replace this block with whatever data you want to capture ──
        var solid = app.testComp.layers.addSolid([1, 0, 0], "t", 200, 200, 1.0);
        var mask = solid.property("ADBE Mask Parade").addProperty("ADBE Mask Atom");
        var sh = new Shape();
        sh.vertices = [[10, 10], [190, 10], [190, 190], [10, 190]];
        sh.inTangents = [[0, 0], [0, 0], [0, 0], [0, 0]];
        sh.outTangents = [[0, 0], [0, 0], [0, 0], [0, 0]];
        sh.closed = true;
        mask.property("ADBE Mask Shape").setValue(sh);
        var feather = mask.property("ADBE Mask Feather");
        feather.setValueAtTime(0.0, [10, 20]);
        feather.setValueAtTime(2.0, [60, 80]);
    });

    step("save_again", function () {
        app.project.save();
    });

    alert(log.join("\n"));
})();
