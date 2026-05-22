// tmp_debug/gen_shape_dummy.jsx
// 跨 AE 版本生成 V2.2 RE 所需的 ShapeLayer fixture 系列。
// args.json (固定路径 tmp_debug/gen_shape_args.json):
//   {"out": "<abs path>.aep", "kind": "<scenario>", "done": "<abs path>.done"}
//
// scenarios:
//   "empty"        — 单 ShapeLayer，root group 空 (RE-S1/S2/S3)
//   "1rect"        — root group + 1 Rect (RE-S4)
//   "1ellipse"     — root group + 1 Ellipse (RE-S5a)
//   "1path_4vtx"   — root group + 1 Path (4 顶点 closed, 无 tangent) (RE-S5b)
//   "1fill"        — root group + 1 Fill (RE-S5c)
//   "1stroke"      — root group + 1 Stroke (RE-S5d)
//   "kf_1"         — root group + 1 Rect, Layer Position 1 keyframe (RE-S6)
//   "kf_2"         — root group + 1 Rect, Layer Position 2 keyframes (RE-S7)
//   "path_tangent" — root group + 1 Path (4 顶点 closed, 非 0 tangent) (RE-S8b)

(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/tmp_debug/gen_shape_args.json");
    var args = {};
    var doneFile = null;
    var log = [];
    var ok = false;

    try {
        argsFile.open("r");
        var s = argsFile.read();
        argsFile.close();
        args = eval("(" + s + ")");
        doneFile = new File(args.done);

        var outFile = new File(args.out);

        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        log.push("ae.version=" + app.version);

        var c = app.project.items.addComp("C1", 1920, 1080, 1, 12, 30);
        var shape = c.layers.addShape();
        shape.name = "S1";
        log.push("created ShapeLayer S1");
        var contents = shape.property("ADBE Root Vectors Group");

        switch (args.kind) {
            case "empty":
                // 不加任何节点
                break;
            case "1rect":
                contents.addProperty("ADBE Vector Shape - Rect");
                break;
            case "1ellipse":
                contents.addProperty("ADBE Vector Shape - Ellipse");
                break;
            case "1path_4vtx":
                var p = contents.addProperty("ADBE Vector Shape - Group");
                var shapeProp = p.property("ADBE Vector Shape");
                var bz = new Shape();
                bz.vertices = [[0,0], [100,0], [100,100], [0,100]];
                bz.inTangents = [[0,0], [0,0], [0,0], [0,0]];
                bz.outTangents = [[0,0], [0,0], [0,0], [0,0]];
                bz.closed = true;
                shapeProp.setValue(bz);
                break;
            case "1fill":
                contents.addProperty("ADBE Vector Graphic - Fill");
                break;
            case "1stroke":
                contents.addProperty("ADBE Vector Graphic - Stroke");
                break;
            case "kf_1":
                contents.addProperty("ADBE Vector Shape - Rect");
                var pos = shape.property("ADBE Transform Group").property("ADBE Position");
                pos.setValueAtTime(0, [960, 540]);
                break;
            case "kf_2":
                contents.addProperty("ADBE Vector Shape - Rect");
                var pos2 = shape.property("ADBE Transform Group").property("ADBE Position");
                pos2.setValueAtTime(0, [0, 0]);
                pos2.setValueAtTime(2, [500, 300]);
                break;
            case "path_tangent":
                var p3 = contents.addProperty("ADBE Vector Shape - Group");
                var shape3 = p3.property("ADBE Vector Shape");
                var bz3 = new Shape();
                bz3.vertices = [[0,0], [100,0], [100,100], [0,100]];
                bz3.inTangents  = [[-10,0], [0,-10], [10,0], [0,10]];
                bz3.outTangents = [[10,0],  [0,10],  [-10,0], [0,-10]];
                bz3.closed = true;
                shape3.setValue(bz3);
                break;
            default:
                throw new Error("unknown kind: " + args.kind);
        }
        log.push("scenario=" + args.kind);

        app.project.save(outFile);
        log.push("saved " + outFile.fsName);
        ok = true;
    } catch (e) {
        ok = false;
        log.push("ERROR: " + e.toString());
    }

    try {
        if (!doneFile) doneFile = new File("e:/projects/tools/aep-parser/tmp_debug/gen_shape.done");
        doneFile.open("w");
        doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
        doneFile.close();
    } catch (e2) { /* swallow */ }
})();
