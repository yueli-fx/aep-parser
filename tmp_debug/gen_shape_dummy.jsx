// tmp_debug/gen_shape_dummy.jsx
// 跨 AE 版本生成 V2.2 RE 所需的 ShapeLayer fixture 系列。
// args.json (固定路径 tmp_debug/gen_shape_args.json):
//   {"out": "<abs path>.aep", "kind": "<scenario>", "done": "<abs path>.done"}
//
// scenarios:
//   "empty"        — 单 ShapeLayer，root group 空 (RE-S1/S2/S3)
//   "1rect"        — root group + 1 Rect (RE-S4)
//   "1ellipse"     — root group + 1 Ellipse (RE-S5a)
//   "1ellipse_set" — root group + 1 Ellipse w/ explicit Size/Position setValue (RE-S5a 编码)
//   "1path_4vtx"   — root group + 1 Path (4 顶点 closed, 无 tangent) (RE-S5b)
//   "1fill"        — root group + 1 Fill (RE-S5c)
//   "1fill_set"    — root group + 1 Fill w/ explicit Color/Opacity setValue (RE-S5c 编码)
//   "1stroke"      — root group + 1 Stroke (RE-S5d)
//   "1stroke_set"  — root group + 1 Stroke w/ explicit Color/Width/Opacity setValue (RE-S5d 编码)
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
            case "1ellipse_set":
                var e = contents.addProperty("ADBE Vector Shape - Ellipse");
                // 自动探测 sub-property matchNames (plan §4.3 未验证)
                for (var i = 1; i <= e.numProperties; i++) {
                    log.push("ellipse.child[" + i + "]: matchName=" + e.property(i).matchName + " name=" + e.property(i).name);
                }
                var eSize = e.property("ADBE Vector Ellipse Size");
                if (eSize) { eSize.setValue([120, 80]); log.push("set Size [120,80]"); }
                else { log.push("WARN: no ADBE Vector Ellipse Size"); }
                var ePos = e.property("ADBE Vector Ellipse Position");
                if (ePos) { ePos.setValue([10, 20]); log.push("set Position [10,20]"); }
                else { log.push("WARN: no ADBE Vector Ellipse Position"); }
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
            case "1fill_set":
                var fl = contents.addProperty("ADBE Vector Graphic - Fill");
                for (var fi = 1; fi <= fl.numProperties; fi++) {
                    log.push("fill.child[" + fi + "]: matchName=" + fl.property(fi).matchName + " name=" + fl.property(fi).name);
                }
                var flColor = fl.property("ADBE Vector Fill Color");
                if (flColor) { flColor.setValue([1, 0, 0, 1]); log.push("set Color [1,0,0,1]"); }
                else { log.push("WARN: no ADBE Vector Fill Color"); }
                var flOp = fl.property("ADBE Vector Fill Opacity");
                if (flOp) { flOp.setValue(75); log.push("set Opacity 75"); }
                else { log.push("WARN: no ADBE Vector Fill Opacity"); }
                break;
            case "1stroke":
                contents.addProperty("ADBE Vector Graphic - Stroke");
                break;
            case "1stroke_set":
                var st = contents.addProperty("ADBE Vector Graphic - Stroke");
                for (var si = 1; si <= st.numProperties; si++) {
                    log.push("stroke.child[" + si + "]: matchName=" + st.property(si).matchName + " name=" + st.property(si).name);
                }
                var stColor = st.property("ADBE Vector Stroke Color");
                if (stColor) { stColor.setValue([0, 0, 1, 1]); log.push("set Color [0,0,1,1]"); }
                else { log.push("WARN: no ADBE Vector Stroke Color"); }
                var stWidth = st.property("ADBE Vector Stroke Width");
                if (stWidth) { stWidth.setValue(5); log.push("set Width 5"); }
                else { log.push("WARN: no ADBE Vector Stroke Width"); }
                var stOp = st.property("ADBE Vector Stroke Opacity");
                if (stOp) { stOp.setValue(80); log.push("set Opacity 80"); }
                else { log.push("WARN: no ADBE Vector Stroke Opacity"); }
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
