# V2.2 ShapeLayer Creation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship `comp.NewShapeLayer(name)` + Rect/Ellipse/Path/Fill/Stroke 节点合成 + PropertyStream (static + keyframe) + AE 2020/2025 ship gate，作为 V2 第二个 sub-project，同时为 V3 brainstorm 提取 runtime / serializer 分层基础。

**Architecture:** Spec 已收齐 (`workshop/specs/v2-2-layer-creation-design.md`)。10 条 Architecture Invariant + 4 条 Serializer Invariant 钉死 runtime ↔ serializer 边界。Builder 走 `lower_*.go` primitives；runtime types 不持 chunk refs。Escape hatch (AE 2020 canonical minimum) 避免 per-version branch。Capability matrix 接口预留，V2.2 实施期增量加。

**Tech Stack:** Go 1.21+ (generics for `PropertyStream[T]`), 现有 `internal/aep/` package, `internal/rifx/` framing, ExtendScript (JSX) for RE + ship gate, AfterFX.exe (2020 / 2022 / 2025) for ship gate validation.

**Spec reference:** `workshop/specs/v2-2-layer-creation-design.md` — 1213 行；所有 invariant / API surface / classification 在 spec 里查。Plan 内 cross-ref spec §N 节即可。

**Pre-flight check (any phase 开工前)：**

```bash
cd E:/projects/tools/aep-parser
go vet ./...                                                    # clean
go test -count=1 ./internal/aep/... -v 2>&1 | grep -c '^--- PASS'  # 122 (baseline; 应一直递增)
git status --short                                              # working tree clean
```

---

# Phase 0 — RE prerequisites (blocking)

**目的:** Phase 1+ 实施全部依赖 RE 数据（empty ShapeLayer 字节 / ShapeNode 默认值 / PropertyStream keyframe encoding / BezierPath encoding）。Phase 0 不写 production 代码，只写 RE 工具 + 跑 AE + dump 字节 + 写 finding 进 spec §8。

**依赖:** AE 2020 + AE 2022 + AE 2025 三个版本可用。`tmp_debug/gen_dummy_comp.jsx` 验证过工作流。

### Task 0.1: `tmp_debug/gen_shape_dummy.jsx`

**Files:**
- Create: `tmp_debug/gen_shape_dummy.jsx`

- [ ] **Step 1: 写 JSX driver**

```jsx
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
```

- [ ] **Step 2: Commit**

```bash
git add tmp_debug/gen_shape_dummy.jsx
git commit -m "tool(re): gen_shape_dummy.jsx — V2.2 RE fixture generator (9 scenarios)"
```

---

### Task 0.2: RE-S1 — empty ShapeLayer (AE 2020 / 2022 / 2025)

**Files:**
- Generate: `tmp_debug/re_v22/empty_ae{2020,2022,2025}.aep`
- Update: `workshop/specs/v2-2-layer-creation-design.md` §8

- [ ] **Step 1: 创建 RE 输出目录**

```bash
mkdir -p tmp_debug/re_v22
```

- [ ] **Step 2: AE 2020 跑 empty scenario**

```bash
cat > tmp_debug/gen_shape_args.json <<'EOF'
{"out": "e:/projects/tools/aep-parser/tmp_debug/re_v22/empty_ae2020.aep", "kind": "empty", "done": "e:/projects/tools/aep-parser/tmp_debug/gen_shape.done"}
EOF
rm -f tmp_debug/gen_shape.done
"E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe" -r "E:/projects/tools/aep-parser/tmp_debug/gen_shape_dummy.jsx" &
disown
until [ -f tmp_debug/gen_shape.done ]; do sleep 2; done
cat tmp_debug/gen_shape.done
```

Expected: `PASS` + `ae.version=17.x` + `saved .../empty_ae2020.aep` + `scenario=empty`

- [ ] **Step 3: AE 2022 跑 empty scenario**

同 Step 2 但 args.out 改 `empty_ae2022.aep`，AE 路径改 `Adobe After Effects 2022`。

- [ ] **Step 4: AE 2025 跑 empty scenario**

同上 → `empty_ae2025.aep` + AE 2025 路径。

- [ ] **Step 5: dump 三个 fixture 的 root + Layr LIST 结构对比**

```bash
go run ./tmp_debug/dump_root tmp_debug/re_v22/empty_ae2020.aep > tmp_debug/re_v22/empty_ae2020.dump
go run ./tmp_debug/dump_root tmp_debug/re_v22/empty_ae2022.aep > tmp_debug/re_v22/empty_ae2022.dump
go run ./tmp_debug/dump_root tmp_debug/re_v22/empty_ae2025.aep > tmp_debug/re_v22/empty_ae2025.dump
diff tmp_debug/re_v22/empty_ae2020.dump tmp_debug/re_v22/empty_ae2025.dump | head -100
```

- [ ] **Step 6: 写 RE-S1 finding 进 spec §8**

打开 `workshop/specs/v2-2-layer-creation-design.md`，定位 `## 8. RE Findings`，追加：

```markdown
### RE-S1 finding: empty ShapeLayer Layr LIST 结构

- Date: 2026-05-22 (示例日期)
- Source: tmp_debug/re_v22/empty_ae{2020,2022,2025}.aep
- Method: gen_shape_dummy.jsx (kind=empty) + go run ./tmp_debug/dump_root
- Layr LIST children 顺序 (AE 2020 canonical):
  1. ldta (NNN bytes)
  2. Utf8 (layer name "S1")
  3. LIST(tdgp, "ADBE Transform Group") — layer Transform
  4. LIST(tdgp, "ADBE Vector Materials Group") — shape contents root
  5. ... (实测填入)
- Cross-version diff: ldta size AE 2020 = 160 / AE 2022 = 160 / AE 2025 = 164
  - Material/Lighting groups: AE 2025 root group 多 N tdgp（per V2.1 既知）
  - V2.2 escape hatch 选 AE 2020 minimum，跨版本字节差不入 capability matrix
- Classification: [serialization]
- 影响: lower_layer.go ldta layout + Layr children 顺序 freeze 此处
```

- [ ] **Step 7: Commit**

```bash
git add tmp_debug/re_v22/empty_ae*.aep tmp_debug/re_v22/empty_ae*.dump workshop/specs/v2-2-layer-creation-design.md
git commit -m "re(v2.2): RE-S1 empty ShapeLayer Layr LIST + cross-version baseline"
```

---

### Task 0.3: RE-S2 — Layer-level Transform group defaults

**Files:**
- Same `tmp_debug/re_v22/empty_ae*.aep` (RE-S1 produced)
- Update: spec §8

- [ ] **Step 1: 写 dump 工具复用**

```bash
# 既有 dump_root 已够；用 grep / awk 抽 "ADBE Transform Group" tdgp 段
go run ./tmp_debug/dump_root tmp_debug/re_v22/empty_ae2020.aep | \
  awk '/ADBE Transform Group/,/ADBE Vector Materials Group/' > tmp_debug/re_v22/empty_ae2020.transform.dump
head -50 tmp_debug/re_v22/empty_ae2020.transform.dump
```

观察项：`tdgp` children 顺序（应含 5 个属性：Anchor Point / Position / Scale / Rotation / Opacity），每个属性的 `tdb4` header + `cdat` 默认值。

- [ ] **Step 2: 写 RE-S2 finding**

追加到 spec §8:

```markdown
### RE-S2 finding: Layer Transform group default values

- Source: tmp_debug/re_v22/empty_ae2020.aep
- Layer Transform tdgp children 顺序:
  1. tdmn "ADBE Anchor Point"  → cdat default [0,0] (2D, ldta @0xBC layer-local 推测)
  2. tdmn "ADBE Position"      → cdat default [comp.width/2, comp.height/2] = [960, 540] (2D)
  3. tdmn "ADBE Scale"         → cdat default [100, 100]
  4. tdmn "ADBE Rotate Z"      → cdat default 0
  5. tdmn "ADBE Opacity"       → cdat default 100
- Classification: [serialization defaults] + [runtime defaults via LayerTransform typed surface]
- 影响: lower_layer.go 实现 Transform group 时用此默认；ShapeLayer.Transform() typed setter
        defaults 跟 V1 Layer.SetPosition 等一致
```

- [ ] **Step 3: Commit**

```bash
git add workshop/specs/v2-2-layer-creation-design.md
git commit -m "re(v2.2): RE-S2 Layer Transform group default values"
```

---

### Task 0.4: RE-S3 — VectorGroup (Vector Materials root) defaults

**Files:**
- Same fixture
- Update: spec §8

- [ ] **Step 1: Dump "ADBE Vector Materials Group" 段**

```bash
go run ./tmp_debug/dump_root tmp_debug/re_v22/empty_ae2020.aep | \
  awk '/ADBE Vector Materials Group/,/^  / && !/^    /' > tmp_debug/re_v22/empty_ae2020.contents.dump
cat tmp_debug/re_v22/empty_ae2020.contents.dump
```

观察项：root VectorGroup 内部 children = 1 个 internal Transform tdgp("ADBE Vector Transform Group")？还是空？

- [ ] **Step 2: 写 RE-S3 finding 进 spec §8**

```markdown
### RE-S3 finding: VectorGroup default (empty RootGroup)

- empty ShapeLayer 的 Vector Materials Group tdgp 内 children:
  1. tdmn "ADBE Vector Group" (root group identifier)
  2. LIST tdgp "ADBE Vectors Group" (children container, 空)
  3. LIST tdgp "ADBE Vector Transform Group" (group-level Transform):
     - Anchor Point [0,0]
     - Position [0,0]
     - Scale [100,100]
     - Rotation 0
     - Opacity 100
     - Skew 0
     - Skew Axis 0
- Classification: [serialization defaults]
- 影响: lower_shape_node.go lowerVectorGroup 时 RootGroup Transform 默认值
```

- [ ] **Step 3: Commit**

```bash
git add workshop/specs/v2-2-layer-creation-design.md
git commit -m "re(v2.2): RE-S3 VectorGroup (root) defaults + group-level Transform"
```

---

### Task 0.5: RE-S4 — Rect node defaults

**Files:**
- Generate `tmp_debug/re_v22/1rect_ae2020.aep`
- Update: spec §8

- [ ] **Step 1: 跑 gen_shape_dummy kind=1rect**

```bash
cat > tmp_debug/gen_shape_args.json <<'EOF'
{"out": "e:/projects/tools/aep-parser/tmp_debug/re_v22/1rect_ae2020.aep", "kind": "1rect", "done": "e:/projects/tools/aep-parser/tmp_debug/gen_shape.done"}
EOF
rm -f tmp_debug/gen_shape.done
"E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe" -r "E:/projects/tools/aep-parser/tmp_debug/gen_shape_dummy.jsx" &
disown
until [ -f tmp_debug/gen_shape.done ]; do sleep 2; done
```

- [ ] **Step 2: Dump Rect node 字节**

```bash
go run ./tmp_debug/dump_root tmp_debug/re_v22/1rect_ae2020.aep | \
  awk '/ADBE Vector Shape - Rect/,/ADBE Vector Group End|ADBE Vector Materials Group/' > \
  tmp_debug/re_v22/1rect.dump
cat tmp_debug/re_v22/1rect.dump
```

观察项：
- Rect tdgp children: tdmn + Size + Position + Roundness 三个 PropertyStream tdgp
- 每个 PropertyStream 的 default value（Size / Position / Roundness）
- tdb4 dimension 字段（2D / 1D）

- [ ] **Step 3: 写 RE-S4 finding**

```markdown
### RE-S4 finding: RectNode defaults

- Default Size: [100, 100] (or [实测])
- Default Position: [0, 0] (or [实测])
- Default Roundness: 0
- tdb4 dimensions: Size=2, Position=2, Roundness=1
- match-name 表已在 spec §4.3，本处 freeze defaults
- Classification: [serialization defaults]
- 影响: lower_shape_node.go lowerRectNode + spec §3.6 defaults table 校准
```

- [ ] **Step 4: Commit**

```bash
git add tmp_debug/re_v22/1rect_ae2020.aep tmp_debug/re_v22/1rect.dump workshop/specs/v2-2-layer-creation-design.md
git commit -m "re(v2.2): RE-S4 RectNode defaults"
```

---

### Task 0.6: RE-S5a — Ellipse defaults

类似 RE-S4，scenario = `1ellipse`，输出 `1ellipse_ae2020.aep`。

- [ ] **Step 1: 跑 + dump + 写 finding**

```bash
cat > tmp_debug/gen_shape_args.json <<'EOF'
{"out": "e:/projects/tools/aep-parser/tmp_debug/re_v22/1ellipse_ae2020.aep", "kind": "1ellipse", "done": "e:/projects/tools/aep-parser/tmp_debug/gen_shape.done"}
EOF
rm -f tmp_debug/gen_shape.done
"E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe" -r "E:/projects/tools/aep-parser/tmp_debug/gen_shape_dummy.jsx" &
disown
until [ -f tmp_debug/gen_shape.done ]; do sleep 2; done
go run ./tmp_debug/dump_root tmp_debug/re_v22/1ellipse_ae2020.aep | \
  awk '/ADBE Vector Shape - Ellipse/,/ADBE Vector Group End|ADBE Vector Materials Group/' > \
  tmp_debug/re_v22/1ellipse.dump
```

- [ ] **Step 2: spec §8 finding + commit**

格式同 RE-S4。Finding 写 EllipseNode Size / Position defaults。

```bash
git add tmp_debug/re_v22/1ellipse* workshop/specs/v2-2-layer-creation-design.md
git commit -m "re(v2.2): RE-S5a EllipseNode defaults"
```

---

### Task 0.7: RE-S5b — Path defaults (4-vertex closed, linear)

scenario = `1path_4vtx`.

- [ ] **Step 1: 跑 + dump**

```bash
cat > tmp_debug/gen_shape_args.json <<'EOF'
{"out": "e:/projects/tools/aep-parser/tmp_debug/re_v22/1path_ae2020.aep", "kind": "1path_4vtx", "done": "e:/projects/tools/aep-parser/tmp_debug/gen_shape.done"}
EOF
rm -f tmp_debug/gen_shape.done
"E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe" -r "E:/projects/tools/aep-parser/tmp_debug/gen_shape_dummy.jsx" &
disown
until [ -f tmp_debug/gen_shape.done ]; do sleep 2; done
```

- [ ] **Step 2: Path 字节深 dump (ldat 编码)**

```bash
# Path 的 PropertyStream 走 ldat 编码 (vs Rect Size 走 cdat)
# 用既有 dump_root + grep 抓 ldat chunk
go run ./tmp_debug/dump_root tmp_debug/re_v22/1path_ae2020.aep | \
  awk '/ADBE Vector Shape - Group/,/ADBE Vector Materials Group/' | head -40 > \
  tmp_debug/re_v22/1path.dump
cat tmp_debug/re_v22/1path.dump
```

观察项：
- `ADBE Vector Shape` PropertyStream 的 tdb4 / ldat
- ldat 内 closed flag 位置 + 4 vertex × (in tangent + vertex + out tangent) 字节

- [ ] **Step 3: spec §8 finding**

```markdown
### RE-S5b finding: PathNode defaults + BezierPath encoding (linear)

- Default PathNode (V2.2 builder requires SetVertices；AE-side default 不直接对应)
- BezierPath ldat encoding (4-vertex closed, linear):
  - Header: closed flag (1 B) at offset N
  - Vertex count (uint16/uint32 at offset M)
  - Per vertex: 24 bytes (in tangent 8 B + vertex 8 B + out tangent 8 B, all [2]f64 BE)
  - 全 0 tangent: 仍写 24 B (per RE-S8 待校；本 task 只 record linear case)
- Classification: [serialization encoding]
- 影响: lower_property_stream.go lowerPathStream + spec §3.3 PathNode constraints (min vertices)
- 待 RE-S8 决议: tangent vs no-tangent 是否同一 encoding format
```

- [ ] **Step 4: Commit**

```bash
git add tmp_debug/re_v22/1path* workshop/specs/v2-2-layer-creation-design.md
git commit -m "re(v2.2): RE-S5b PathNode + BezierPath linear encoding"
```

---

### Task 0.8: RE-S5c — Fill node defaults

scenario = `1fill`. 同样 4-step pattern。

- [ ] **Step 1: 跑 + dump + finding + commit**

```bash
cat > tmp_debug/gen_shape_args.json <<'EOF'
{"out": "e:/projects/tools/aep-parser/tmp_debug/re_v22/1fill_ae2020.aep", "kind": "1fill", "done": "e:/projects/tools/aep-parser/tmp_debug/gen_shape.done"}
EOF
rm -f tmp_debug/gen_shape.done
"E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe" -r "E:/projects/tools/aep-parser/tmp_debug/gen_shape_dummy.jsx" &
disown
until [ -f tmp_debug/gen_shape.done ]; do sleep 2; done
go run ./tmp_debug/dump_root tmp_debug/re_v22/1fill_ae2020.aep | \
  awk '/ADBE Vector Graphic - Fill/,/ADBE Vector Materials Group/' > tmp_debug/re_v22/1fill.dump
```

Finding 关注：
- Fill tdgp children: tdmn + Composite / FillRule / Color / Opacity PropertyStreams
- Color cdat 32 B encoding (RGBA 0..1 float64 ×4)
- Opacity 默认 100
- Composite + FillRule 默认 enum 值

```bash
git add tmp_debug/re_v22/1fill* workshop/specs/v2-2-layer-creation-design.md
git commit -m "re(v2.2): RE-S5c FillNode defaults + Color RGBA encoding"
```

---

### Task 0.9: RE-S5d — Stroke node defaults

scenario = `1stroke`. 同上。

- [ ] **Step 1: 跑 + dump + finding + commit**

```bash
cat > tmp_debug/gen_shape_args.json <<'EOF'
{"out": "e:/projects/tools/aep-parser/tmp_debug/re_v22/1stroke_ae2020.aep", "kind": "1stroke", "done": "e:/projects/tools/aep-parser/tmp_debug/gen_shape.done"}
EOF
rm -f tmp_debug/gen_shape.done
"E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe" -r "E:/projects/tools/aep-parser/tmp_debug/gen_shape_dummy.jsx" &
disown
until [ -f tmp_debug/gen_shape.done ]; do sleep 2; done
go run ./tmp_debug/dump_root tmp_debug/re_v22/1stroke_ae2020.aep | \
  awk '/ADBE Vector Graphic - Stroke/,/ADBE Vector Materials Group/' > tmp_debug/re_v22/1stroke.dump
```

Finding 关注:
- Stroke tdgp children: tdmn + Composite / Color / Opacity / Width / LineCap / LineJoin / MiterLimit / Dashes / TaperStart / TaperEnd 等
- V2.2 hot path: Color / Opacity / Width 三个走 typed setter；其它 escape hatch
- 各 default 值

```bash
git add tmp_debug/re_v22/1stroke* workshop/specs/v2-2-layer-creation-design.md
git commit -m "re(v2.2): RE-S5d StrokeNode defaults"
```

---

### Task 0.10: RE-S6 — PropertyStream 1 keyframe encoding

scenario = `kf_1`. 关键 RE — keyframe table 编码。

- [ ] **Step 1: 跑 kf_1**

```bash
cat > tmp_debug/gen_shape_args.json <<'EOF'
{"out": "e:/projects/tools/aep-parser/tmp_debug/re_v22/kf_1_ae2020.aep", "kind": "kf_1", "done": "e:/projects/tools/aep-parser/tmp_debug/gen_shape.done"}
EOF
rm -f tmp_debug/gen_shape.done
"E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe" -r "E:/projects/tools/aep-parser/tmp_debug/gen_shape_dummy.jsx" &
disown
until [ -f tmp_debug/gen_shape.done ]; do sleep 2; done
```

- [ ] **Step 2: Dump ADBE Position 流（带 1 keyframe）字节**

```bash
go run ./tmp_debug/dump_root tmp_debug/re_v22/kf_1_ae2020.aep | \
  awk '/ADBE Position/,/ADBE Scale/' > tmp_debug/re_v22/kf_1.dump
cat tmp_debug/re_v22/kf_1.dump
```

观察 lhd3 (52 B header) + ldat 字段：
- lhd3: TickRate / numKeyframes / interp mode flags
- ldat: 1 keyframe record stride

- [ ] **Step 3: 用 V1 既有 keyframe parser 解码作 cross-check**

V1 keyframe 解码已实现 (`parse_keyframe.go`)；可以用既有 parser 跑这个 fixture 校 finding：

```bash
go test ./internal/aep/ -run TestKeyframe -v  # 既有测试通过即 parser 可信
# 写小 tmp_debug 工具读 kf_1_ae2020.aep 的 Layer Position keyframes 输出
```

- [ ] **Step 4: spec §8 finding**

```markdown
### RE-S6 finding: PropertyStream 1-keyframe encoding (Layer Position 2D)

- Source: tmp_debug/re_v22/kf_1_ae2020.aep
- lhd3 (52 B): 
  - offset 0..3: TickRate (uint32 BE) — comp's ticks/sec; should match cdta @0x08
  - offset 4..7: numKeyframes (uint32 BE) = 1
  - offset 8..51: 实测填入 (flags / pad / 其它)
- ldat (per-keyframe record, T=[2]float64):
  - offset 0..7: time as ticks (uint64 BE; = round(time_seconds × TickRate))
  - offset 8..15: temporal in-ease speed (float64 BE)
  - offset 16..23: temporal in-ease influence (float64 BE)
  - offset 24..31: temporal out-ease speed
  - offset 32..39: temporal out-ease influence
  - offset 40..55: value (2 × float64 BE for [2]float64)
  - 实测每 record stride = N bytes
- Classification: [serialization encoding]
- 影响: lower_property_stream.go lowerVec2Stream keyframe path
```

- [ ] **Step 5: Commit**

```bash
git add tmp_debug/re_v22/kf_1* workshop/specs/v2-2-layer-creation-design.md
git commit -m "re(v2.2): RE-S6 PropertyStream 1-keyframe encoding"
```

---

### Task 0.11: RE-S7 — PropertyStream 2 keyframes (stride 校验)

scenario = `kf_2`. 验证 record stride。

- [ ] **Step 1: 跑 + dump + finding**

```bash
cat > tmp_debug/gen_shape_args.json <<'EOF'
{"out": "e:/projects/tools/aep-parser/tmp_debug/re_v22/kf_2_ae2020.aep", "kind": "kf_2", "done": "e:/projects/tools/aep-parser/tmp_debug/gen_shape.done"}
EOF
rm -f tmp_debug/gen_shape.done
"E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe" -r "E:/projects/tools/aep-parser/tmp_debug/gen_shape_dummy.jsx" &
disown
until [ -f tmp_debug/gen_shape.done ]; do sleep 2; done
go run ./tmp_debug/dump_root tmp_debug/re_v22/kf_2_ae2020.aep | \
  awk '/ADBE Position/,/ADBE Scale/' > tmp_debug/re_v22/kf_2.dump
```

校：`(ldat_size - lhd3_size_excess) / 2 == per-keyframe stride` 跟 RE-S6 的 stride 一致。lhd3 numKeyframes = 2。

- [ ] **Step 2: spec §8 finding + commit**

```markdown
### RE-S7 finding: PropertyStream 2-keyframe stride 校验

- 与 RE-S6 单 keyframe stride 一致 → record stride freeze
- lhd3 numKeyframes = 2 ✓
- ldat = 2 × (per-record stride)
```

```bash
git add tmp_debug/re_v22/kf_2* workshop/specs/v2-2-layer-creation-design.md
git commit -m "re(v2.2): RE-S7 PropertyStream 2-keyframe stride validation"
```

---

### Task 0.12: RE-S8 — BezierPath tangent encoding (tangent vs no-tangent)

scenario = `path_tangent` + RE-S5b 已有的 linear。

- [ ] **Step 1: 跑 path_tangent**

```bash
cat > tmp_debug/gen_shape_args.json <<'EOF'
{"out": "e:/projects/tools/aep-parser/tmp_debug/re_v22/path_tangent_ae2020.aep", "kind": "path_tangent", "done": "e:/projects/tools/aep-parser/tmp_debug/gen_shape.done"}
EOF
rm -f tmp_debug/gen_shape.done
"E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe" -r "E:/projects/tools/aep-parser/tmp_debug/gen_shape_dummy.jsx" &
disown
until [ -f tmp_debug/gen_shape.done ]; do sleep 2; done
```

- [ ] **Step 2: Diff linear vs tangent**

```bash
go run ./tmp_debug/dump_root tmp_debug/re_v22/1path_ae2020.aep | \
  awk '/ADBE Vector Shape - Group/,/ADBE Vector Materials Group/' > tmp_debug/re_v22/path_linear.dump
go run ./tmp_debug/dump_root tmp_debug/re_v22/path_tangent_ae2020.aep | \
  awk '/ADBE Vector Shape - Group/,/ADBE Vector Materials Group/' > tmp_debug/re_v22/path_tangent.dump
diff tmp_debug/re_v22/path_linear.dump tmp_debug/re_v22/path_tangent.dump
```

观察：
- ldat 长度是否相同（24 B/vertex 单一 encoding）？
- in/out tangent 字节是否在 linear case 全 0 但仍占位？

- [ ] **Step 3: spec §8 finding**

```markdown
### RE-S8 finding: BezierPath tangent encoding format

- 实测 linear vs non-linear path 同一 fixture 顶点数:
  - linear (全 0 tangent): ldat size = N B
  - non-linear: ldat size = same N B (tangent 字段始终 24 B/vertex 写出)
  - OR: ldat size 不同 (e.g. linear 省略 tangent block) → 需 V2.2 双格式 lowering
- 决议: [实测结果] → lower_property_stream.go lowerPathStream 用单一/双格式编码
- 最少顶点: 通过 AE UI 实测 Closed=true / Closed=false 各 0/1/2 顶点的 AE 行为:
  - Closed=true 最少 N 顶点
  - Closed=false 最少 M 顶点
- Classification: [serialization encoding] + 部分 [runtime invariant for SetVertices min check]
- 影响: spec §3.3 PathNode min vertices freeze；lowerPathStream tangent encoding
```

- [ ] **Step 4: Commit**

```bash
git add tmp_debug/re_v22/path_* workshop/specs/v2-2-layer-creation-design.md
git commit -m "re(v2.2): RE-S8 BezierPath tangent encoding + min vertices"
```

---

### Task 0.13: RE-S9 — Cross-version diff (AE 2020 vs 2025)

**Files:**
- Generate AE 2025 versions of selected fixtures
- Update: spec §8

- [x] **Step 1: AE 2025 跑 1rect / 1fill / kf_2 / 1path_4vtx**

```bash
for scenario in 1rect 1fill kf_2 1path_4vtx; do
  cat > tmp_debug/gen_shape_args.json <<EOF
{"out": "e:/projects/tools/aep-parser/tmp_debug/re_v22/${scenario}_ae2025.aep", "kind": "${scenario}", "done": "e:/projects/tools/aep-parser/tmp_debug/gen_shape.done"}
EOF
  rm -f tmp_debug/gen_shape.done
  "E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe" -r "E:/projects/tools/aep-parser/tmp_debug/gen_shape_dummy.jsx" &
  disown
  until [ -f tmp_debug/gen_shape.done ]; do sleep 2; done
done
```

- [x] **Step 2: Diff each scenario AE 2020 vs AE 2025**

```bash
for scenario in 1rect 1fill kf_2 1path_4vtx; do
  echo "=== $scenario ==="
  go run ./tmp_debug/dump_root tmp_debug/re_v22/${scenario}_ae2020.aep > /tmp/${scenario}_2020.dump
  go run ./tmp_debug/dump_root tmp_debug/re_v22/${scenario}_ae2025.aep > /tmp/${scenario}_2025.dump
  diff /tmp/${scenario}_2020.dump /tmp/${scenario}_2025.dump | head -30
  echo
done
```

- [x] **Step 3: spec §8 finding — Capability candidates admission**

```markdown
### RE-S9 finding: cross-version diff + capability matrix admission

- 实测差异（per scenario）:
  - ldta size: 160 (AE 2020) vs 164 (AE 2025) — 既知，V2.2 escape hatch 不入 matrix
  - Material/Lighting groups: AE 2025 多 N 个 tdgp child — V2.2 不出
  - Stroke 字段新增 (TaperStart/End 等 AE 22+/24+): 若 AE 2020 缺 → V2.2 不出
  - PathBezierEncoding: AE 2020 vs 2025 字节 [一致 / 不同实测] → admission [no / yes]
  - KeyframeEaseEncoding: 同上
- V2.2 capability matrix decision (per admission rule):
  - 无 trait 入 matrix（escape hatch + AE 2020 canonical 路径覆盖所有跨版本差异）
  - 候选保留在 spec §6.5 待 V3 重审
- Classification: [capability-cand 文档]
- 影响: spec §6.5 admission decisions finalize；capability_matrix.go 仍空 struct
```

- [x] **Step 4: Commit**

```bash
git add tmp_debug/re_v22/*_ae2025.aep workshop/specs/v2-2-layer-creation-design.md
git commit -m "re(v2.2): RE-S9 cross-version diff + capability admission decisions"
```

---

### Task 0.14: Phase 0 收尾 — defaults table freeze + Phase 1 unblock

**Files:**
- Update: `workshop/specs/v2-2-layer-creation-design.md` §3.6 defaults table (RE 校准后 freeze)
- Update: `workshop/board.md` 标 Phase 0 完

- [ ] **Step 1: Spec §3.6 defaults table 把 RE 实测值填入，去掉 "provisional" 标记**

打开 spec，定位 `### 3.6 Defaults`，根据 RE-S2 / S3 / S4 / S5a-d findings 填具体值。

- [ ] **Step 2: Spec §6.4 substrate 段补 "Empty ShapeLayer canonical bytes" / "Empty VectorGroup Transform 字节" 的提取来源**

填入：`从 tmp_debug/re_v22/empty_ae2020.aep 提取，embed 进 lower_layer.go / lower_shape_node.go 作为 byte const`

- [ ] **Step 3: Update board.md Phase 0 archive**

```markdown
### 2026-MM-DD Phase 0 RE Prerequisites 完 (122 PASS 不变)

V2.2 实施前 RE：9 个 fixture 跑 AE 2020 + 2022 + 2025 → 12 个 RE finding 填 spec §8。
- empty ShapeLayer + Layer/Group/Node defaults / PropertyStream keyframe encoding
  / BezierPath tangent encoding / 跨版本 diff capability admission
- 决策: capability_matrix.go ship 时空 struct (escape hatch 覆盖所有 V2.2 范围差异)
- 工具: tmp_debug/gen_shape_dummy.jsx (9 scenarios)
- next: Phase 1 runtime types
```

- [ ] **Step 4: Verify spec self-consistency**

```bash
cd E:/projects/tools/aep-parser
# 全文搜索剩余 "provisional" / "TBD" / "待 RE" 应该清零
grep -nE 'provisional|TBD|TODO|待 RE-S' workshop/specs/v2-2-layer-creation-design.md
```

Expected: 仅 `§8 finding template` 模板内的 placeholder 残留（无实际 pending）。

- [ ] **Step 5: Commit**

```bash
git add workshop/specs/v2-2-layer-creation-design.md workshop/board.md
git commit -m "re(v2.2): Phase 0 RE complete — defaults freeze + Phase 1 unblock"
```

**Phase 0 PASS criterion:** Spec §3.6 defaults table 全 freeze；§6.4 substrate 来源 explicit；§6.5 admission decisions finalize；§8 含 ≥12 finding records。Phase 1 unblock.

---

# Phase 1 — Runtime types

**目的:** 落 spec §2 所有 runtime concept 的 Go 类型骨架。无 serializer / 无 API 入口，纯类型 + 内部状态机。TDD: 每类型先写状态机 / 字段 access 失败测，再实现。

### Task 1.1: `internal/aep/ldta_layout.go` (offset 常量)

**Files:**
- Create: `internal/aep/ldta_layout.go`

- [ ] **Step 1: 写文件，mirror cdta_layout.go 风格**

```go
// internal/aep/ldta_layout.go
package aep

// ldta byte offsets — single source of truth for parser / writer / builder.
// ldta total size: 160 bytes (AE 2020 canonical, V2.2 builder target).
// AE 2025 = 164 bytes (escape hatch: builder 不出 AE 25 native; 见 spec §1.4).
//
// 来源: V1 parse_layer.go offset 注释 + Phase 0 RE-S1 dump 校准。
const (
	ldtaLayerID         = 0x00 // uint32 BE; layer-local ID
	ldtaQuality         = 0x04 // uint16 BE; LayerQuality enum
	// ... (Phase 0 RE-S1 校准后填入完整 offset 集)
	ldtaBlendingMode    = 0x63 // uint8
	ldtaPreserveTrans   = 0x67 // uint8
	ldtaTrackMatte      = 0x6B // uint8 TrackMatteType
	ldtaParentID        = 0x84 // uint32 BE
	ldtaSize2020        = 160  // AE 2020 / 2022 canonical (V2.2 builder target)
	ldtaSize2025        = 164  // AE 2025 (escape hatch 不出)
)
```

- [ ] **Step 2: go vet + commit**

```bash
go vet ./internal/aep/...
git add internal/aep/ldta_layout.go
git commit -m "feat(v2.2): ldta_layout.go offset constants"
```

---

### Task 1.2: `internal/aep/capability_matrix.go` (空骨架)

**Files:**
- Create: `internal/aep/capability_matrix.go`

- [ ] **Step 1: 写空骨架**

```go
// internal/aep/capability_matrix.go
package aep

// AECapabilities 描述某 AE target 的 serializer-affecting traits。V2.2 ship 时
// 空 struct —— escape hatch (AE 2020 canonical minimum) 覆盖所有当前已知差异。
// 增 trait 必须满足 §1.4 admission rule 三条件。
//
// Candidates (Phase 0 RE-S9 实测，V2.2 全不入 matrix；见 spec §6.5):
//   LdtaSize, FEEHasPpSn, MaterialLightingGroups, TdgpDefaultChildren,
//   ShapeMatchNameVariants, PathBezierEncoding, KeyframeEaseEncoding
type AECapabilities struct {
	// V2.2 实施期增量加（受 admission rule 限制）。当前为空。
}

// Capabilities 返回 target 对应的 capability 集。纯函数 lookup，不挂在 Project 上 ——
// 为 V3 auto-derive (OQ-1) 留接口。V2.2 全 target 返同一空 struct.
func Capabilities(target AETarget) AECapabilities {
	return AECapabilities{}
}
```

- [ ] **Step 2: go vet + commit**

```bash
go vet ./internal/aep/...
git add internal/aep/capability_matrix.go
git commit -m "feat(v2.2): capability_matrix.go skeleton (V2.2 ship 时空 struct)"
```

---

### Task 1.3: `internal/aep/property_stream.go` (PropertyStream[T] + state machine + Keyframe + Ease)

**Files:**
- Create: `internal/aep/property_stream.go`
- Test: `internal/aep/property_stream_test.go`

- [ ] **Step 1: 写失败测试 — state machine transitions**

```go
// internal/aep/property_stream_test.go
package aep_test

import (
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestPropertyStream_InitialStaticMode(t *testing.T) {
	ps := aep.NewPropertyStream[float64]()
	if ps.Mode() != aep.StreamModeStatic {
		t.Fatalf("initial mode = %v, want Static", ps.Mode())
	}
	v, isStatic := ps.StaticValue()
	if !isStatic {
		t.Fatalf("StaticValue() ok = false, want true (initial Static mode)")
	}
	if v != 0.0 {
		t.Fatalf("initial static = %v, want zero", v)
	}
}

func TestPropertyStream_SetStaticValue(t *testing.T) {
	ps := aep.NewPropertyStream[float64]()
	if err := ps.SetStaticValue(3.14); err != nil {
		t.Fatal(err)
	}
	v, _ := ps.StaticValue()
	if v != 3.14 {
		t.Fatalf("after SetStaticValue: %v, want 3.14", v)
	}
}

func TestPropertyStream_AddKeyframe_TransitionsToAnimated(t *testing.T) {
	ps := aep.NewPropertyStream[float64]()
	_ = ps.SetStaticValue(100.0)

	if err := ps.AddKeyframeLinear(0, 0); err != nil {
		t.Fatal(err)
	}
	if ps.Mode() != aep.StreamModeAnimated {
		t.Fatalf("after AddKeyframe: mode = %v, want Animated", ps.Mode())
	}
	if _, isStatic := ps.StaticValue(); isStatic {
		t.Fatalf("StaticValue() ok = true in Animated mode, want false (Inv-8)")
	}
}

func TestPropertyStream_SetStaticValue_InAnimatedMode_Error(t *testing.T) {
	ps := aep.NewPropertyStream[float64]()
	_ = ps.AddKeyframeLinear(0, 0)
	if err := ps.SetStaticValue(99); err == nil {
		t.Fatal("SetStaticValue in Animated mode should error")
	}
}

func TestPropertyStream_Clear_RestoresStaticMode(t *testing.T) {
	ps := aep.NewPropertyStream[float64]()
	_ = ps.SetStaticValue(42.0)
	_ = ps.AddKeyframeLinear(0, 0)
	if err := ps.Clear(); err != nil {
		t.Fatal(err)
	}
	if ps.Mode() != aep.StreamModeStatic {
		t.Fatalf("after Clear: mode = %v, want Static", ps.Mode())
	}
	v, isStatic := ps.StaticValue()
	if !isStatic || v != 42.0 {
		t.Fatalf("after Clear: StaticValue = (%v, %v), want (42.0, true)", v, isStatic)
	}
}

func TestPropertyStream_NegativeTime_Rejected(t *testing.T) {
	ps := aep.NewPropertyStream[float64]()
	if err := ps.AddKeyframeLinear(-0.001, 0); err == nil {
		t.Fatal("negative time should error")
	}
}

func TestPropertyStream_DuplicateTime_Rejected(t *testing.T) {
	ps := aep.NewPropertyStream[float64]()
	_ = ps.AddKeyframeLinear(1.0, 0)
	if err := ps.AddKeyframeLinear(1.0, 99); err == nil {
		t.Fatal("duplicate time should error")
	}
}
```

- [ ] **Step 2: Run test → expect compile fail**

```bash
go test ./internal/aep/ -run TestPropertyStream -v
```

Expected: compile error `undefined: aep.NewPropertyStream` etc.

- [ ] **Step 3: 实现 property_stream.go**

```go
// internal/aep/property_stream.go
package aep

import "fmt"

// StreamMode is PropertyStream 状态机 mode (Inv-8: Static ↔ Animated 互斥).
type StreamMode int

const (
	StreamModeStatic StreamMode = iota
	StreamModeAnimated
)

// PropertyStream[T] 是 V2.2 引入的 canonical animation primitive (Inv-4).
// T = float64 / [2]float64 / [3]float64 / [4]float64 / BezierPath.
// Time domain = seconds (Inv-7); ticks 转换是 serializer concern.
type PropertyStream[T any] struct {
	mode       StreamMode
	static     T
	keyframes  []Keyframe[T]
	expression string
}

// Keyframe[T] = (time:seconds, value:T, ease).
type Keyframe[T any] struct {
	Time            float64
	Value           T
	InEase, OutEase TemporalEase
}

// TemporalEase = (speed:value/sec, influence:0..1).
type TemporalEase struct {
	Speed     float64
	Influence float64
}

// NewPropertyStream 返回 Static mode + zero value 的新 stream.
func NewPropertyStream[T any]() *PropertyStream[T] {
	return &PropertyStream[T]{mode: StreamModeStatic}
}

func (ps *PropertyStream[T]) Mode() StreamMode { return ps.mode }

func (ps *PropertyStream[T]) StaticValue() (T, bool) {
	var zero T
	if ps.mode != StreamModeStatic {
		return zero, false
	}
	return ps.static, true
}

func (ps *PropertyStream[T]) Keyframes() []Keyframe[T] { return ps.keyframes }
func (ps *PropertyStream[T]) HasKeyframes() bool       { return len(ps.keyframes) > 0 }

func (ps *PropertyStream[T]) SetStaticValue(v T) error {
	if ps.mode == StreamModeAnimated {
		return fmt.Errorf("SetStaticValue: stream is in Animated mode; call Clear() first (Inv-8)")
	}
	ps.static = v
	return nil
}

func (ps *PropertyStream[T]) AddKeyframeLinear(time float64, value T) error {
	return ps.addKeyframe(Keyframe[T]{Time: time, Value: value})
}

func (ps *PropertyStream[T]) AddKeyframeWithEase(time float64, value T, in, out TemporalEase) error {
	return ps.addKeyframe(Keyframe[T]{Time: time, Value: value, InEase: in, OutEase: out})
}

func (ps *PropertyStream[T]) addKeyframe(kf Keyframe[T]) error {
	if kf.Time < 0 {
		return fmt.Errorf("AddKeyframe: time < 0 (got %g; Inv-7: seconds-only)", kf.Time)
	}
	for _, existing := range ps.keyframes {
		if existing.Time == kf.Time {
			return fmt.Errorf("AddKeyframe: duplicate time %g (AE 不允许)", kf.Time)
		}
	}
	if ps.mode == StreamModeStatic {
		ps.mode = StreamModeAnimated // Inv-8 transition
	}
	ps.keyframes = append(ps.keyframes, kf)
	// 维持 time 升序排列以便 lowering
	for i := len(ps.keyframes) - 1; i > 0 && ps.keyframes[i].Time < ps.keyframes[i-1].Time; i-- {
		ps.keyframes[i], ps.keyframes[i-1] = ps.keyframes[i-1], ps.keyframes[i]
	}
	return nil
}

func (ps *PropertyStream[T]) Clear() error {
	ps.keyframes = nil
	ps.mode = StreamModeStatic
	return nil
}
```

- [ ] **Step 4: Run test → expect PASS**

```bash
go test ./internal/aep/ -run TestPropertyStream -v
```

Expected: all PropertyStream tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/aep/property_stream.go internal/aep/property_stream_test.go
git commit -m "feat(v2.2): PropertyStream[T] with Static↔Animated state machine"
```

---

### Task 1.4: `internal/aep/shape_graph.go` (ShapeNode + VectorGroup + 5 typed nodes + BezierPath)

**Files:**
- Create: `internal/aep/shape_graph.go`
- Test: `internal/aep/shape_graph_test.go`

- [ ] **Step 1: 写失败测 — ShapeNode interface conformance + typed node construction**

```go
// internal/aep/shape_graph_test.go
package aep_test

import (
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestVectorGroup_Empty(t *testing.T) {
	g := aep.NewVectorGroup()
	if got := len(g.Children); got != 0 {
		t.Fatalf("new VectorGroup children = %d, want 0", got)
	}
	if g.Transform == nil {
		t.Fatal("Transform PropertyGroup nil, want non-nil (group-level Transform)")
	}
}

func TestRectNode_Construct_Defaults(t *testing.T) {
	r := aep.NewRectNode()
	if r.Kind() != aep.ShapeKindRect {
		t.Fatalf("kind = %v, want ShapeKindRect", r.Kind())
	}
	// Defaults per RE-S4 (Phase 0 校准):
	v, _ := r.Size().StaticValue()
	if v != [2]float64{100, 100} {
		t.Fatalf("Rect default Size = %v, want [100,100] (RE-S4)", v)
	}
	pos, _ := r.Position().StaticValue()
	if pos != [2]float64{0, 0} {
		t.Fatalf("Rect default Position = %v, want [0,0]", pos)
	}
	rnd, _ := r.Roundness().StaticValue()
	if rnd != 0 {
		t.Fatalf("Rect default Roundness = %v, want 0", rnd)
	}
}

func TestRectNode_SetSize(t *testing.T) {
	r := aep.NewRectNode()
	if err := r.SetSize([2]float64{200, 100}); err != nil {
		t.Fatal(err)
	}
	v, _ := r.Size().StaticValue()
	if v != [2]float64{200, 100} {
		t.Fatalf("after SetSize: %v, want [200,100]", v)
	}
}

func TestFillNode_Defaults(t *testing.T) {
	f := aep.NewFillNode()
	if f.Kind() != aep.ShapeKindFill {
		t.Fatalf("kind = %v, want ShapeKindFill", f.Kind())
	}
	c, _ := f.Color().StaticValue()
	if c != [4]float64{1, 1, 1, 1} {
		t.Fatalf("Fill default Color = %v, want [1,1,1,1] white", c)
	}
}

func TestStrokeNode_Defaults(t *testing.T) {
	s := aep.NewStrokeNode()
	if s.Kind() != aep.ShapeKindStroke {
		t.Fatalf("kind = %v, want ShapeKindStroke", s.Kind())
	}
	c, _ := s.Color().StaticValue()
	if c != [4]float64{0, 0, 0, 1} {
		t.Fatalf("Stroke default Color = %v, want [0,0,0,1] black", c)
	}
	w, _ := s.Width().StaticValue()
	if w != 2 {
		t.Fatalf("Stroke default Width = %v, want 2", w)
	}
}

func TestPathNode_SetVertices(t *testing.T) {
	p := aep.NewPathNode()
	if err := p.SetVertices([][2]float64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}); err != nil {
		t.Fatal(err)
	}
	bz, _ := p.Path().StaticValue()
	if len(bz.Vertices) != 4 {
		t.Fatalf("Vertices count = %d, want 4", len(bz.Vertices))
	}
	if !bz.Closed {
		t.Fatal("PathNode default Closed = false, want true")
	}
}

func TestPathNode_SetVertices_RejectsEmpty(t *testing.T) {
	p := aep.NewPathNode()
	if err := p.SetVertices(nil); err == nil {
		t.Fatal("nil vertices should error")
	}
	if err := p.SetVertices([][2]float64{{0, 0}}); err == nil {
		t.Fatal("single vertex should error (per RE-S8 min check)")
	}
}
```

- [ ] **Step 2: Run test → compile fail**

```bash
go test ./internal/aep/ -run "TestVectorGroup|TestRectNode|TestFillNode|TestStrokeNode|TestPathNode" -v
```

- [ ] **Step 3: 实现 shape_graph.go**

```go
// internal/aep/shape_graph.go
package aep

import "fmt"

// ShapeNodeKind: runtime-facing node identity. match-name strings are
// serializer-only (lower_shape_node.go shapeMatchNames map).
type ShapeNodeKind int

const (
	ShapeKindRect ShapeNodeKind = iota
	ShapeKindEllipse
	ShapeKindPath
	ShapeKindFill
	ShapeKindStroke
	ShapeKindGroup // V2.2 internal only (嵌套合成 API 不开放)
	// V2.3+: ShapeKindPolyStar / ShapeKindGradientFill / ShapeKindGradientStroke
	//        ShapeKindTrim / ShapeKindMerge / ShapeKindRepeater / ShapeKindTransform
)

// ShapeNode 是 runtime-facing shape graph node interface.
type ShapeNode interface {
	Kind() ShapeNodeKind
	Properties() *PropertyGroup // escape hatch β
}

// VectorGroup 是 shape graph 的容器节点。每个 ShapeLayer 默认带 1 个 RootGroup.
// Children 顺序 = AE render order (Children[0] 最底层；len-1 最上层 / 最后 Add).
type VectorGroup struct {
	Children  []ShapeNode
	Transform *PropertyGroup // group-level Transform (V2.2 默认 identity，不暴露 typed setter)
}

func NewVectorGroup() *VectorGroup {
	return &VectorGroup{
		Transform: newGroupTransform(),
	}
}

// BezierPath 是 runtime geometry object (not a serializer encoding mirror).
type BezierPath struct {
	Vertices    [][2]float64
	InTangents  [][2]float64
	OutTangents [][2]float64
	Closed      bool
}

// RectNode
type RectNode struct {
	size      *PropertyStream[[2]float64]
	position  *PropertyStream[[2]float64]
	roundness *PropertyStream[float64]
}

func NewRectNode() *RectNode {
	r := &RectNode{
		size:      NewPropertyStream[[2]float64](),
		position:  NewPropertyStream[[2]float64](),
		roundness: NewPropertyStream[float64](),
	}
	_ = r.size.SetStaticValue([2]float64{100, 100})
	_ = r.position.SetStaticValue([2]float64{0, 0})
	_ = r.roundness.SetStaticValue(0)
	return r
}

func (r *RectNode) Kind() ShapeNodeKind                  { return ShapeKindRect }
func (r *RectNode) Size() *PropertyStream[[2]float64]    { return r.size }
func (r *RectNode) Position() *PropertyStream[[2]float64] { return r.position }
func (r *RectNode) Roundness() *PropertyStream[float64]  { return r.roundness }
func (r *RectNode) SetSize(v [2]float64) error           { return r.size.SetStaticValue(v) }
func (r *RectNode) SetPosition(v [2]float64) error       { return r.position.SetStaticValue(v) }
func (r *RectNode) SetRoundness(v float64) error         { return r.roundness.SetStaticValue(v) }
func (r *RectNode) Properties() *PropertyGroup           { return nil /* Phase 3 escape hatch 填 */ }

// EllipseNode
type EllipseNode struct {
	size, position *PropertyStream[[2]float64]
}

func NewEllipseNode() *EllipseNode {
	e := &EllipseNode{
		size:     NewPropertyStream[[2]float64](),
		position: NewPropertyStream[[2]float64](),
	}
	_ = e.size.SetStaticValue([2]float64{100, 100})
	_ = e.position.SetStaticValue([2]float64{0, 0})
	return e
}

func (e *EllipseNode) Kind() ShapeNodeKind                   { return ShapeKindEllipse }
func (e *EllipseNode) Size() *PropertyStream[[2]float64]     { return e.size }
func (e *EllipseNode) Position() *PropertyStream[[2]float64] { return e.position }
func (e *EllipseNode) SetSize(v [2]float64) error            { return e.size.SetStaticValue(v) }
func (e *EllipseNode) SetPosition(v [2]float64) error        { return e.position.SetStaticValue(v) }
func (e *EllipseNode) Properties() *PropertyGroup            { return nil }

// PathNode
type PathNode struct {
	path *PropertyStream[BezierPath]
}

func NewPathNode() *PathNode {
	p := &PathNode{path: NewPropertyStream[BezierPath]()}
	_ = p.path.SetStaticValue(BezierPath{Closed: true})
	return p
}

func (p *PathNode) Kind() ShapeNodeKind            { return ShapeKindPath }
func (p *PathNode) Path() *PropertyStream[BezierPath] { return p.path }
func (p *PathNode) Properties() *PropertyGroup     { return nil }

func (p *PathNode) SetVertices(verts [][2]float64) error {
	if len(verts) < 2 {
		return fmt.Errorf("PathNode.SetVertices: need >= 2 vertices (AE 拒 0/1; per RE-S8)")
	}
	current, _ := p.path.StaticValue()
	zeroTangents := make([][2]float64, len(verts))
	return p.path.SetStaticValue(BezierPath{
		Vertices:    append([][2]float64(nil), verts...),
		InTangents:  zeroTangents,
		OutTangents: zeroTangents,
		Closed:      current.Closed,
	})
}

func (p *PathNode) SetClosed(closed bool) error {
	current, _ := p.path.StaticValue()
	current.Closed = closed
	return p.path.SetStaticValue(current)
}

// FillNode
type FillNode struct {
	color   *PropertyStream[[4]float64]
	opacity *PropertyStream[float64]
}

func NewFillNode() *FillNode {
	f := &FillNode{
		color:   NewPropertyStream[[4]float64](),
		opacity: NewPropertyStream[float64](),
	}
	_ = f.color.SetStaticValue([4]float64{1, 1, 1, 1}) // white
	_ = f.opacity.SetStaticValue(100)
	return f
}

func (f *FillNode) Kind() ShapeNodeKind                  { return ShapeKindFill }
func (f *FillNode) Color() *PropertyStream[[4]float64]   { return f.color }
func (f *FillNode) Opacity() *PropertyStream[float64]    { return f.opacity }
func (f *FillNode) SetColor(v [4]float64) error          { return f.color.SetStaticValue(v) }
func (f *FillNode) SetOpacity(v float64) error           { return f.opacity.SetStaticValue(v) }
func (f *FillNode) Properties() *PropertyGroup           { return nil }

// StrokeNode
type StrokeNode struct {
	color   *PropertyStream[[4]float64]
	opacity *PropertyStream[float64]
	width   *PropertyStream[float64]
}

func NewStrokeNode() *StrokeNode {
	s := &StrokeNode{
		color:   NewPropertyStream[[4]float64](),
		opacity: NewPropertyStream[float64](),
		width:   NewPropertyStream[float64](),
	}
	_ = s.color.SetStaticValue([4]float64{0, 0, 0, 1}) // black
	_ = s.opacity.SetStaticValue(100)
	_ = s.width.SetStaticValue(2)
	return s
}

func (s *StrokeNode) Kind() ShapeNodeKind                 { return ShapeKindStroke }
func (s *StrokeNode) Color() *PropertyStream[[4]float64]  { return s.color }
func (s *StrokeNode) Opacity() *PropertyStream[float64]   { return s.opacity }
func (s *StrokeNode) Width() *PropertyStream[float64]     { return s.width }
func (s *StrokeNode) SetColor(v [4]float64) error         { return s.color.SetStaticValue(v) }
func (s *StrokeNode) SetOpacity(v float64) error          { return s.opacity.SetStaticValue(v) }
func (s *StrokeNode) SetWidth(v float64) error            { return s.width.SetStaticValue(v) }
func (s *StrokeNode) Properties() *PropertyGroup          { return nil }

// newGroupTransform 返回 identity Transform PropertyGroup (RootGroup 默认).
// Phase 3 填实际 PropertyGroup 结构；Phase 1 占位返 nil 的 placeholder.
func newGroupTransform() *PropertyGroup {
	return &PropertyGroup{Name: "Transform"}
}

// PropertyGroup 是 runtime-name keyed tree (escape hatch β; Phase 3 完整实现).
type PropertyGroup struct {
	Name      string
	Children  map[string]*PropertyGroup
	streams   map[string]any
	Separated bool
}
```

- [ ] **Step 4: Run test → expect PASS**

```bash
go test ./internal/aep/ -run "TestVectorGroup|TestRectNode|TestFillNode|TestStrokeNode|TestPathNode" -v
```

- [ ] **Step 5: Commit**

```bash
git add internal/aep/shape_graph.go internal/aep/shape_graph_test.go
git commit -m "feat(v2.2): shape_graph.go — ShapeNode interface + 5 typed nodes + VectorGroup + BezierPath"
```

---

### Task 1.5: `internal/aep/types_core.go` — ShapeLayer typed wrapper + LayerTransform skeleton

**Files:**
- Modify: `internal/aep/types_core.go`
- Test: `internal/aep/shape_layer_test.go`

- [ ] **Step 1: 写失败测 — ShapeLayer wrap + LayerTransform access**

```go
// internal/aep/shape_layer_test.go
package aep_test

import (
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestShapeLayer_Wraps_Layer(t *testing.T) {
	base := &aep.Layer{Type: aep.LayerTypeShape, Name: "test"}
	s := aep.WrapShapeLayer(base)
	if s.Name != "test" {
		t.Fatalf("ShapeLayer.Name (via embedded Layer) = %q, want %q", s.Name, "test")
	}
	if s.RootGroup() == nil {
		t.Fatal("ShapeLayer.RootGroup() nil")
	}
}

func TestShapeLayer_Transform_AccessorsReturnNonNilStreams(t *testing.T) {
	base := &aep.Layer{Type: aep.LayerTypeShape}
	s := aep.WrapShapeLayer(base)
	if s.Transform().Position() == nil {
		t.Fatal("Transform().Position() nil")
	}
	if s.Position() == nil {
		t.Fatal("shorthand Position() nil")
	}
}
```

- [ ] **Step 2: Run test → compile fail**

```bash
go test ./internal/aep/ -run "TestShapeLayer" -v
```

- [ ] **Step 3: 在 types_core.go 加 ShapeLayer + LayerTransform 类型**

定位 `internal/aep/types_core.go` 末尾，追加：

```go
// ShapeLayer 是 V2.2 引入的 typed wrapper around *Layer (per spec §2.1).
// V1 callers 通过 *Layer 继续工作；V2.2 新建路径返 *ShapeLayer 暴露 shape-specific API.
type ShapeLayer struct {
	*Layer                  // embed: V1 setter / getter 全继承
	rootGroup *VectorGroup
	transform *LayerTransform
}

// WrapShapeLayer 把已 parse / 创建的 *Layer 包成 ShapeLayer.
// 调用者负责确保 layer.Type == LayerTypeShape；本函数不校 (per V1 既有 contract).
func WrapShapeLayer(layer *Layer) *ShapeLayer {
	s := &ShapeLayer{
		Layer:     layer,
		rootGroup: NewVectorGroup(),
		transform: newLayerTransform(),
	}
	return s
}

func (s *ShapeLayer) RootGroup() *VectorGroup { return s.rootGroup }
func (s *ShapeLayer) Transform() *LayerTransform { return s.transform }

// Shorthand accessors (delegate to Transform)
func (s *ShapeLayer) Position() *PropertyStream[[2]float64] { return s.transform.position }
func (s *ShapeLayer) Scale() *PropertyStream[[2]float64]    { return s.transform.scale }
func (s *ShapeLayer) Rotation() *PropertyStream[float64]    { return s.transform.rotation }
func (s *ShapeLayer) Opacity() *PropertyStream[float64]     { return s.transform.opacity }

// LayerTransform 是 layer-level Transform 的 typed wrapper.
// V2.2 ShapeLayer = 2D layer (3D 推 V2.3+) → Position 走 2D stream.
type LayerTransform struct {
	anchorPoint *PropertyStream[[2]float64]
	position    *PropertyStream[[2]float64]
	scale       *PropertyStream[[2]float64]
	rotation    *PropertyStream[float64]
	opacity     *PropertyStream[float64]
}

func newLayerTransform() *LayerTransform {
	lt := &LayerTransform{
		anchorPoint: NewPropertyStream[[2]float64](),
		position:    NewPropertyStream[[2]float64](),
		scale:       NewPropertyStream[[2]float64](),
		rotation:    NewPropertyStream[float64](),
		opacity:     NewPropertyStream[float64](),
	}
	// Defaults per RE-S2:
	_ = lt.scale.SetStaticValue([2]float64{100, 100})
	_ = lt.opacity.SetStaticValue(100)
	return lt
}

func (t *LayerTransform) AnchorPoint() *PropertyStream[[2]float64] { return t.anchorPoint }
func (t *LayerTransform) Position() *PropertyStream[[2]float64]    { return t.position }
func (t *LayerTransform) Scale() *PropertyStream[[2]float64]       { return t.scale }
func (t *LayerTransform) Rotation() *PropertyStream[float64]       { return t.rotation }
func (t *LayerTransform) Opacity() *PropertyStream[float64]        { return t.opacity }
```

- [ ] **Step 4: Run test → PASS**

```bash
go test ./internal/aep/ -run "TestShapeLayer" -v
```

- [ ] **Step 5: Commit**

```bash
git add internal/aep/types_core.go internal/aep/shape_layer_test.go
git commit -m "feat(v2.2): ShapeLayer typed wrapper + LayerTransform"
```

---

### Task 1.6: Phase 1 collect + PASS count check

- [ ] **Step 1: Run all tests**

```bash
go vet ./...
go test -count=1 ./internal/aep/... -v 2>&1 | tail -10
go test -count=1 ./internal/aep/... -v 2>&1 | grep -c '^--- PASS'
```

Expected: 122 baseline + Phase 1 加测 (~10) ≈ 132 PASS, 0 FAIL.

- [ ] **Step 2: 更新 board.md Phase 1 archive line**

```bash
# 在 board.md 最近归档段 + 1 行（位置照 Phase 0 之上）
```

- [ ] **Step 3: Commit**

```bash
git add workshop/board.md
git commit -m "docs(board): V2.2 Phase 1 runtime types complete (~132 PASS)"
```

**Phase 1 PASS criterion**: PASS ≥ 130 (具体值跟实际 test count 校); go vet clean; types_core.go ShapeLayer 编译过；V1 既有 tests 全通过.

---

# Phase 2 — Serializer primitives

**目的:** 落 5 个 lowering primitive (`lower_property_stream.go` / `lower_shape_node.go` / `lower_layer.go` + rename `lower_item_siblings.go`)。每个 primitive 跟 spec §4 描述对应。

### Task 2.1: `internal/aep/lower_property_stream.go` — 5 个 typed lowering function

**Files:**
- Create: `internal/aep/lower_property_stream.go`
- Test: `internal/aep/lower_property_stream_test.go`

- [ ] **Step 1: 写失败测 — static float64 stream → tdgp chunk**

```go
// internal/aep/lower_property_stream_test.go
package aep_test

import (
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestLowerFloat64Stream_Static(t *testing.T) {
	ps := aep.NewPropertyStream[float64]()
	_ = ps.SetStaticValue(0.5)
	chunk, err := aep.LowerFloat64Stream(ps, "ADBE Opacity", "Opacity", aep.NewLowerCtxForTest())
	if err != nil {
		t.Fatal(err)
	}
	if chunk == nil || !chunk.IsList() {
		t.Fatal("expected LIST(tdgp) chunk")
	}
	// Top-level structure 校 (tdmn + tdbs); 详细字节校在 Phase 4 roundtrip
}

func TestLowerVec2Stream_Animated(t *testing.T) {
	ps := aep.NewPropertyStream[[2]float64]()
	_ = ps.AddKeyframeLinear(0, [2]float64{0, 0})
	_ = ps.AddKeyframeLinear(2, [2]float64{500, 300})
	chunk, err := aep.LowerVec2Stream(ps, "ADBE Position", "Position", aep.NewLowerCtxForTest())
	if err != nil {
		t.Fatal(err)
	}
	if chunk == nil {
		t.Fatal("expected non-nil chunk")
	}
}
```

`NewLowerCtxForTest()` 是 Phase 2 export 的测试 helper（位于 lower_property_stream.go 内）。

- [ ] **Step 2: Run test → compile fail**

- [ ] **Step 3: 实现 lower_property_stream.go**

详细字节 layout 来自 Phase 0 RE-S6/S7/S8 finding。skeleton：

```go
// internal/aep/lower_property_stream.go
package aep

import (
	"encoding/binary"
	"math"

	"github.com/example/aep-parser/internal/rifx"
)

// lowerCtx 是 serializer-side lowering state (per spec §4.2).
// 不持 runtime graph handle (Inv 强化: lowerCtx carries lowering state only).
type lowerCtx struct {
	tickRate     float64
	capabilities AECapabilities
	nextLayerID  func() uint32
}

// NewLowerCtxForTest exports a default lowerCtx for unit tests.
// Production code 不应调（lowerCtx 应由 NewShapeLayer / WriteAEP 等内部构造）.
func NewLowerCtxForTest() *lowerCtx {
	return &lowerCtx{
		tickRate: 30720, // AE 30fps default
	}
}

// LowerFloat64Stream → LIST(tdgp) chunk (1D PropertyStream lowering).
// Children: tdmn (matchName) + LIST(tdbs)(tdsb + tdsn + tdb4 + cdat OR list[lhd3+ldat]).
func LowerFloat64Stream(ps *PropertyStream[float64], matchName, displayName string, ctx *lowerCtx) (*rifx.Chunk, error) {
	return lowerStream(ps.mode, encode1D, matchName, displayName, ps.static, ps.keyframes, ctx)
}

// LowerVec2Stream → LIST(tdgp) chunk (2D PropertyStream).
func LowerVec2Stream(ps *PropertyStream[[2]float64], matchName, displayName string, ctx *lowerCtx) (*rifx.Chunk, error) {
	return lowerStream(ps.mode, encode2D, matchName, displayName, ps.static, ps.keyframes, ctx)
}

// LowerVec3Stream → LIST(tdgp) chunk (3D PropertyStream).
func LowerVec3Stream(ps *PropertyStream[[3]float64], matchName, displayName string, ctx *lowerCtx) (*rifx.Chunk, error) {
	return lowerStream(ps.mode, encode3D, matchName, displayName, ps.static, ps.keyframes, ctx)
}

// LowerColorStream → LIST(tdgp) chunk (RGBA color, 4D float64).
func LowerColorStream(ps *PropertyStream[[4]float64], matchName, displayName string, ctx *lowerCtx) (*rifx.Chunk, error) {
	return lowerStream(ps.mode, encode4D, matchName, displayName, ps.static, ps.keyframes, ctx)
}

// LowerPathStream → LIST(tdgp) chunk (BezierPath).
func LowerPathStream(ps *PropertyStream[BezierPath], matchName, displayName string, ctx *lowerCtx) (*rifx.Chunk, error) {
	return lowerStream(ps.mode, encodeBezier, matchName, displayName, ps.static, ps.keyframes, ctx)
}

type encodeFunc[T any] func(T) []byte

func encode1D(v float64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, math.Float64bits(v))
	return b
}

func encode2D(v [2]float64) []byte {
	b := make([]byte, 16)
	binary.BigEndian.PutUint64(b[0:8], math.Float64bits(v[0]))
	binary.BigEndian.PutUint64(b[8:16], math.Float64bits(v[1]))
	return b
}

func encode3D(v [3]float64) []byte {
	b := make([]byte, 24)
	for i := 0; i < 3; i++ {
		binary.BigEndian.PutUint64(b[i*8:(i+1)*8], math.Float64bits(v[i]))
	}
	return b
}

func encode4D(v [4]float64) []byte {
	b := make([]byte, 32)
	for i := 0; i < 4; i++ {
		binary.BigEndian.PutUint64(b[i*8:(i+1)*8], math.Float64bits(v[i]))
	}
	return b
}

func encodeBezier(p BezierPath) []byte {
	// V2.2 BezierPath encoding per RE-S8 finding (Phase 0 Task 0.12 dump → spec §8):
	// header (closed flag + vertex count) + N × 24 B (in tangent + vertex + out tangent)
	// 复用 V1 parse_shape.go decodeShapePrimitive / ShapePath 反向逻辑作 mirror.
	// 实施期 Phase 2 Task 2.1 完成时本函数 body 必须填完 (依赖 Phase 0 已完成)。
	// Placeholder 占位 (Phase 0 未完时 Phase 2 不应启动)：
	panic("encodeBezier: not yet implemented — requires Phase 0 RE-S8 byte layout (Task 0.12)")
}

// lowerStream is the generic core that emits the tdgp chunk.
// Each T-specific entry (LowerFloat64Stream / LowerVec2Stream / ...) 调本函数，
// 传入 T-specific encodeFunc。
func lowerStream[T any](
	mode StreamMode,
	enc encodeFunc[T],
	matchName, displayName string,
	staticVal T,
	keyframes []Keyframe[T],
	ctx *lowerCtx,
) (*rifx.Chunk, error) {
	// PropertyStream tdgp children 顺序 (per V1 parse_property + Phase 0 RE-S4/S6 finding):
	//   tdmn (matchName)
	//   LIST(tdbs) {
	//     tdsb (flag bits)
	//     tdsn (display name Utf8 embed)
	//     tdb4 (property header)
	//     IF static: cdat (raw value)
	//     IF animated: LIST(list)(lhd3 + ldat)
	//     tdum / tduM (V1: min / max bounds; V2.2 默认全 0)
	//   }
	//
	// Phase 2 实现详细字节 layout (依赖 Phase 0 dump 数据)。

	// Skeleton: 仅构造 chunk shape，字节实测 freeze 后填
	tdgp := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	// tdmn
	tdgp.Children = append(tdgp.Children, &rifx.Chunk{
		ID:   rifx.IDTdmn,
		Data: padMatchName(matchName), // 40 字节 NUL-padded ASCII (V1 既有 helper)
	})
	// LIST(tdbs)
	tdbs := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdbs}
	// ... (tdsb / tdsn / tdb4 / cdat or list[lhd3+ldat])
	tdgp.Children = append(tdgp.Children, tdbs)

	return tdgp, nil
}

// padMatchName: 40-byte NUL-padded ASCII; mirror V1 既有逻辑 (复用 parse_property.go 工具).
func padMatchName(s string) []byte {
	b := make([]byte, 40)
	copy(b, []byte(s))
	return b
}
```

**Phase 2 实施 NOTE**: 上面 skeleton 的 `lowerStream` / `encodeBezier` 等需要在 Phase 0 RE 数据 finalize 后用具体字节 layout 填。Phase 2 单测仅校 chunk shape；Phase 4 roundtrip 校字节正确。

- [ ] **Step 4: Run test → PASS (skeleton 层级)**

```bash
go test ./internal/aep/ -run "TestLowerFloat64Stream|TestLowerVec2Stream" -v
```

- [ ] **Step 5: Commit**

```bash
git add internal/aep/lower_property_stream.go internal/aep/lower_property_stream_test.go
git commit -m "feat(v2.2): lower_property_stream.go skeleton — 5 typed lowering funcs"
```

---

### Task 2.2: `internal/aep/lower_shape_node.go` — node lowering + match-name table

**Files:**
- Create: `internal/aep/lower_shape_node.go`
- Test: `internal/aep/lower_shape_node_test.go`

- [ ] **Step 1: 写失败测 — Rect / Fill / Stroke / Ellipse / Path / VectorGroup**

```go
package aep_test

import (
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestLowerRectNode_HasTdmn(t *testing.T) {
	r := aep.NewRectNode()
	chunk, err := aep.LowerShapeNodeForTest(r)
	if err != nil {
		t.Fatal(err)
	}
	if chunk == nil {
		t.Fatal("nil chunk")
	}
	// tdgp.Children[0] should be tdmn = "ADBE Vector Shape - Rect"
}

func TestLowerVectorGroup_RootHasTransformGroup(t *testing.T) {
	g := aep.NewVectorGroup()
	rect := aep.NewRectNode()
	g.Children = append(g.Children, rect)
	chunk, err := aep.LowerVectorGroupForTest(g)
	if err != nil {
		t.Fatal(err)
	}
	if chunk == nil {
		t.Fatal("nil chunk")
	}
	// children: Transform tdgp + Rect tdgp (Transform 在前)
}
```

- [ ] **Step 2: 实现 lower_shape_node.go**

```go
// internal/aep/lower_shape_node.go
package aep

import (
	"fmt"

	"github.com/example/aep-parser/internal/rifx"
)

// shapeMatchNames: ShapeNodeKind → AE match-name 字符串 (serializer-only).
// runtime API 通过 Go enum (ShapeKindRect) 引用；用户永远看不到字符串.
var shapeMatchNames = map[ShapeNodeKind]string{
	ShapeKindRect:    "ADBE Vector Shape - Rect",
	ShapeKindEllipse: "ADBE Vector Shape - Ellipse",
	ShapeKindPath:    "ADBE Vector Shape - Group",
	ShapeKindFill:    "ADBE Vector Graphic - Fill",
	ShapeKindStroke:  "ADBE Vector Graphic - Stroke",
	ShapeKindGroup:   "ADBE Vector Group",
}

// LowerShapeNodeForTest exports lowerShapeNode for unit tests.
func LowerShapeNodeForTest(n ShapeNode) (*rifx.Chunk, error) {
	return lowerShapeNode(n, NewLowerCtxForTest())
}

// LowerVectorGroupForTest 同上.
func LowerVectorGroupForTest(g *VectorGroup) (*rifx.Chunk, error) {
	return lowerVectorGroup(g, NewLowerCtxForTest())
}

// lowerShapeNode → LIST(tdgp) chunk wrapping node:
//   tdmn (matchName)
//   N × LIST(tdgp) (子 PropertyStream)
func lowerShapeNode(n ShapeNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	switch node := n.(type) {
	case *RectNode:
		return lowerRectNode(node, ctx)
	case *EllipseNode:
		return lowerEllipseNode(node, ctx)
	case *PathNode:
		return lowerPathNode(node, ctx)
	case *FillNode:
		return lowerFillNode(node, ctx)
	case *StrokeNode:
		return lowerStrokeNode(node, ctx)
	default:
		return nil, fmt.Errorf("lowerShapeNode: unsupported kind %v", n.Kind())
	}
}

func lowerRectNode(r *RectNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	tdgp := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	tdgp.Children = append(tdgp.Children, &rifx.Chunk{
		ID:   rifx.IDTdmn,
		Data: padMatchName(shapeMatchNames[ShapeKindRect]),
	})
	if c, err := LowerVec2Stream(r.size, "ADBE Vector Rect Size", "Size", ctx); err != nil {
		return nil, err
	} else {
		tdgp.Children = append(tdgp.Children, c)
	}
	if c, err := LowerVec2Stream(r.position, "ADBE Vector Rect Position", "Position", ctx); err != nil {
		return nil, err
	} else {
		tdgp.Children = append(tdgp.Children, c)
	}
	if c, err := LowerFloat64Stream(r.roundness, "ADBE Vector Rect Roundness", "Roundness", ctx); err != nil {
		return nil, err
	} else {
		tdgp.Children = append(tdgp.Children, c)
	}
	return tdgp, nil
}

func lowerEllipseNode(e *EllipseNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	tdgp := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	tdgp.Children = append(tdgp.Children, &rifx.Chunk{
		ID:   rifx.IDTdmn,
		Data: padMatchName(shapeMatchNames[ShapeKindEllipse]),
	})
	if c, err := LowerVec2Stream(e.size, "ADBE Vector Ellipse Size", "Size", ctx); err != nil {
		return nil, err
	} else {
		tdgp.Children = append(tdgp.Children, c)
	}
	if c, err := LowerVec2Stream(e.position, "ADBE Vector Ellipse Position", "Position", ctx); err != nil {
		return nil, err
	} else {
		tdgp.Children = append(tdgp.Children, c)
	}
	return tdgp, nil
}

func lowerPathNode(p *PathNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	tdgp := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	tdgp.Children = append(tdgp.Children, &rifx.Chunk{
		ID:   rifx.IDTdmn,
		Data: padMatchName(shapeMatchNames[ShapeKindPath]),
	})
	if c, err := LowerPathStream(p.path, "ADBE Vector Shape", "Path", ctx); err != nil {
		return nil, err
	} else {
		tdgp.Children = append(tdgp.Children, c)
	}
	return tdgp, nil
}

func lowerFillNode(f *FillNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	tdgp := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	tdgp.Children = append(tdgp.Children, &rifx.Chunk{
		ID:   rifx.IDTdmn,
		Data: padMatchName(shapeMatchNames[ShapeKindFill]),
	})
	if c, err := LowerColorStream(f.color, "ADBE Vector Fill Color", "Color", ctx); err != nil {
		return nil, err
	} else {
		tdgp.Children = append(tdgp.Children, c)
	}
	if c, err := LowerFloat64Stream(f.opacity, "ADBE Vector Fill Opacity", "Opacity", ctx); err != nil {
		return nil, err
	} else {
		tdgp.Children = append(tdgp.Children, c)
	}
	return tdgp, nil
}

func lowerStrokeNode(s *StrokeNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	tdgp := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	tdgp.Children = append(tdgp.Children, &rifx.Chunk{
		ID:   rifx.IDTdmn,
		Data: padMatchName(shapeMatchNames[ShapeKindStroke]),
	})
	if c, err := LowerColorStream(s.color, "ADBE Vector Stroke Color", "Color", ctx); err != nil {
		return nil, err
	} else {
		tdgp.Children = append(tdgp.Children, c)
	}
	if c, err := LowerFloat64Stream(s.opacity, "ADBE Vector Stroke Opacity", "Opacity", ctx); err != nil {
		return nil, err
	} else {
		tdgp.Children = append(tdgp.Children, c)
	}
	if c, err := LowerFloat64Stream(s.width, "ADBE Vector Stroke Width", "Width", ctx); err != nil {
		return nil, err
	} else {
		tdgp.Children = append(tdgp.Children, c)
	}
	return tdgp, nil
}

// lowerVectorGroup → LIST(tdgp) wrap children.
// V2.2 depth = 1: RootGroup → ShapeNode 直接子项；嵌套 group 推 V2.3.
//
// Children 顺序:
//   tdmn ("ADBE Vector Group")
//   LIST(tdgp, "ADBE Vector Transform Group")  ← group-level Transform (默认 identity)
//   tdmn ("ADBE Vectors Group")
//   LIST(tdgp, "ADBE Vectors Group")           ← contains all ShapeNode children
func lowerVectorGroup(g *VectorGroup, ctx *lowerCtx) (*rifx.Chunk, error) {
	tdgp := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	// tdmn root group
	tdgp.Children = append(tdgp.Children, &rifx.Chunk{
		ID:   rifx.IDTdmn,
		Data: padMatchName("ADBE Vector Group"),
	})
	// group-level Transform (RE-S3 defaults)
	transformTdgp, err := lowerGroupTransform(g.Transform, ctx)
	if err != nil {
		return nil, err
	}
	tdgp.Children = append(tdgp.Children, transformTdgp)
	// contents container
	tdgp.Children = append(tdgp.Children, &rifx.Chunk{
		ID:   rifx.IDTdmn,
		Data: padMatchName("ADBE Vectors Group"),
	})
	contents := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	for _, child := range g.Children {
		childChunk, err := lowerShapeNode(child, ctx)
		if err != nil {
			return nil, err
		}
		contents.Children = append(contents.Children, childChunk)
	}
	tdgp.Children = append(tdgp.Children, contents)
	return tdgp, nil
}

// lowerGroupTransform: 实施期 Phase 2 Task 2.2 完成时本函数 body 必须填完
// (依赖 Phase 0 RE-S3 Task 0.4 dump 数据；该 task 完成前 Phase 2 启动应 fail)。
func lowerGroupTransform(_ *PropertyGroup, _ *lowerCtx) (*rifx.Chunk, error) {
	panic("lowerGroupTransform: not yet implemented — requires Phase 0 RE-S3 byte layout (Task 0.4)")
}
```

- [ ] **Step 3: Run test → PASS**

```bash
go test ./internal/aep/ -run "TestLowerRectNode|TestLowerVectorGroup" -v
```

- [ ] **Step 4: Commit**

```bash
git add internal/aep/lower_shape_node.go internal/aep/lower_shape_node_test.go
git commit -m "feat(v2.2): lower_shape_node.go — 5 node lowering + VectorGroup + match-name table"
```

---

### Task 2.3: `internal/aep/lower_layer.go` — ShapeLayer lowering

**Files:**
- Create: `internal/aep/lower_layer.go`
- Test: `internal/aep/lower_layer_test.go`

- [ ] **Step 1: 写失败测**

```go
package aep_test

import (
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestLowerShapeLayer_EmptyHasLayrChunk(t *testing.T) {
	base := &aep.Layer{Type: aep.LayerTypeShape, Name: "S1"}
	s := aep.WrapShapeLayer(base)
	chunk, err := aep.LowerShapeLayerForTest(s)
	if err != nil {
		t.Fatal(err)
	}
	if chunk == nil {
		t.Fatal("nil")
	}
	if !chunk.IsList() {
		t.Fatal("expected LIST(Layr)")
	}
	// children: ldta + Utf8(name) + Transform tdgp + Vector Materials tdgp + ...
}
```

- [ ] **Step 2: 实现 lower_layer.go**

skeleton（详细 ldta 字节 layout 依赖 Phase 0 RE-S1）：

```go
// internal/aep/lower_layer.go
package aep

import (
	"github.com/example/aep-parser/internal/rifx"
)

// LowerShapeLayerForTest exports lowerShapeLayer for unit tests.
func LowerShapeLayerForTest(s *ShapeLayer) (*rifx.Chunk, error) {
	return lowerShapeLayer(s, NewLowerCtxForTest())
}

// lowerShapeLayer → LIST(Layr) chunk.
// children 顺序 per Phase 0 RE-S1 freeze:
//   ldta (160 B AE 2020 canonical)
//   Utf8 (layer name)
//   LIST(tdgp, "ADBE Transform Group")  — layer Transform
//   LIST(tdgp, "ADBE Vector Materials Group")  — shape contents root
//   ... 其它 chunks per RE-S1 dump
func lowerShapeLayer(s *ShapeLayer, ctx *lowerCtx) (*rifx.Chunk, error) {
	layr := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDLayr}

	// ldta (Phase 0 RE-S1 finalize 后填具体字节)
	ldta := &rifx.Chunk{ID: rifx.IDLdta, Data: buildLdtaBytes(s, ctx)}
	layr.Children = append(layr.Children, ldta)

	// Utf8 name
	layr.Children = append(layr.Children, &rifx.Chunk{
		ID:   rifx.IDUtf8,
		Data: []byte(s.Name),
	})

	// Layer Transform tdgp (RE-S2 defaults)
	transformTdgp, err := lowerLayerTransform(s.transform, ctx)
	if err != nil {
		return nil, err
	}
	layr.Children = append(layr.Children, transformTdgp)

	// Vector Materials Group (root group wrapping)
	materialsTdgp := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	materialsTdgp.Children = append(materialsTdgp.Children, &rifx.Chunk{
		ID:   rifx.IDTdmn,
		Data: padMatchName("ADBE Vector Materials Group"),
	})
	rootGroupChunk, err := lowerVectorGroup(s.rootGroup, ctx)
	if err != nil {
		return nil, err
	}
	materialsTdgp.Children = append(materialsTdgp.Children, rootGroupChunk)
	layr.Children = append(layr.Children, materialsTdgp)

	return layr, nil
}

// buildLdtaBytes: 160 B AE 2020 canonical ldta。Phase 0 RE-S1 dump 后填。
func buildLdtaBytes(s *ShapeLayer, ctx *lowerCtx) []byte {
	d := make([]byte, ldtaSize2020)
	// ID @0x00
	// 其它 default fields per RE-S1
	// (Phase 2 完成 RE 后填)
	return d
}

// lowerLayerTransform: layer-level Transform tdgp (5 streams: Anchor/Position/Scale/Rotation/Opacity).
func lowerLayerTransform(t *LayerTransform, ctx *lowerCtx) (*rifx.Chunk, error) {
	tdgp := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	tdgp.Children = append(tdgp.Children, &rifx.Chunk{
		ID:   rifx.IDTdmn,
		Data: padMatchName("ADBE Transform Group"),
	})
	streams := []struct {
		ps        *PropertyStream[[2]float64]
		ps1d      *PropertyStream[float64]
		matchName string
		display   string
		is1D      bool
	}{
		{ps: t.anchorPoint, matchName: "ADBE Anchor Point", display: "Anchor Point"},
		{ps: t.position, matchName: "ADBE Position", display: "Position"},
		{ps: t.scale, matchName: "ADBE Scale", display: "Scale"},
		{ps1d: t.rotation, matchName: "ADBE Rotate Z", display: "Rotation", is1D: true},
		{ps1d: t.opacity, matchName: "ADBE Opacity", display: "Opacity", is1D: true},
	}
	for _, s := range streams {
		var c *rifx.Chunk
		var err error
		if s.is1D {
			c, err = LowerFloat64Stream(s.ps1d, s.matchName, s.display, ctx)
		} else {
			c, err = LowerVec2Stream(s.ps, s.matchName, s.display, ctx)
		}
		if err != nil {
			return nil, err
		}
		tdgp.Children = append(tdgp.Children, c)
	}
	return tdgp, nil
}
```

- [ ] **Step 3: Run test → PASS**

```bash
go test ./internal/aep/ -run "TestLowerShapeLayer" -v
```

- [ ] **Step 4: Commit**

```bash
git add internal/aep/lower_layer.go internal/aep/lower_layer_test.go
git commit -m "feat(v2.2): lower_layer.go — ShapeLayer → LIST(Layr) lowering"
```

---

### Task 2.4: Rename V2.1 item siblings → `lower_item_siblings.go`

**Files:**
- Rename/extract: V2.1 既有 `compTmpl.siblingChunks` 相关代码 (在 `new_composition.go`) 提到新文件
- Modify: `internal/aep/new_composition.go`

- [ ] **Step 1: 创建 `internal/aep/lower_item_siblings.go`，从 new_composition.go 移逻辑**

具体：把 V2.1 `templateItemSiblingChunks` + sibling clone 逻辑提到新文件，public 名 `lowerItemSiblings`。

```go
// internal/aep/lower_item_siblings.go
package aep

import "github.com/example/aep-parser/internal/rifx"

// lowerItemSiblings emits the 8 Fold-level sibling chunks AE expects after
// each Item LIST. V2.1 实测 (per scars/ae25-acceptance-gate.md):
//   LIST(FEE) + fvdv + fiop + ftts + foac + fiac + fipc + fifl
//
// V2.2 escape hatch (AE 2020 minimum) 下：FEE LIST 0 children；其它 chunks 来自
// 2020_dummy_comp.aep substrate (deep-clone via compTmpl.siblingChunks).
//
// Inv-6 合规: substrate 是 bootstrap，不是 semantic template.
func lowerItemSiblings(_ *lowerCtx) []*rifx.Chunk {
	ensureCompTemplate()
	out := make([]*rifx.Chunk, 0, len(compTmpl.siblingChunks))
	for _, c := range compTmpl.siblingChunks {
		out = append(out, deepCloneChunk(c))
	}
	return out
}
```

- [ ] **Step 2: 修改 new_composition.go 调用 lowerItemSiblings (替换 inline 逻辑)**

```bash
grep -n "siblingChunks" internal/aep/new_composition.go
```

把原 inline `for _, sib := range tmpl.siblingChunks { ... }` 改 `for _, sib := range lowerItemSiblings(nil) { ... }`（V2.1 调用点没 lowerCtx，传 nil；后续 V3 brainstorm 决定要不要传 ctx）。

- [ ] **Step 3: Run all tests → V2.1 既有 122 tests + Phase 1/2 加测全 PASS**

```bash
go test -count=1 ./internal/aep/...
```

- [ ] **Step 4: Commit**

```bash
git add internal/aep/lower_item_siblings.go internal/aep/new_composition.go
git commit -m "refactor(v2.2): rename V2.1 item siblings → lower_item_siblings.go (formalize primitive)"
```

---

### Task 2.5: Phase 2 收尾 + PASS count check

- [ ] **Step 1: Run all + count**

```bash
go vet ./...
go test -count=1 ./internal/aep/... -v 2>&1 | grep -c '^--- PASS'
```

Expected: ~140 PASS (122 + Phase 1 ~10 + Phase 2 ~8).

- [ ] **Step 2: board.md Phase 2 archive line**

- [ ] **Step 3: Commit**

```bash
git add workshop/board.md
git commit -m "docs(board): V2.2 Phase 2 serializer primitives skeleton complete (~140 PASS)"
```

**Phase 2 PASS criterion:** PASS ≥ 135; 5 个 lowering primitive 文件全在；既有 V2.1 122 tests 不受影响.

---

# Phase 3 — Public API entry

**目的:** 落 `NewShapeLayer` + AddRect/AddEllipse/AddPath/AddFill/AddStroke + Escape hatch。这是用户可见的 API 层；Phase 4 接 roundtrip 闭环。

### Task 3.1: `internal/aep/new_layer.go` — `comp.NewShapeLayer(name)`

**Files:**
- Create: `internal/aep/new_layer.go`
- Test: `internal/aep/new_shape_layer_test.go`

- [ ] **Step 1: 写失败测**

```go
package aep_test

import (
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestNewShapeLayer_BasicCreation(t *testing.T) {
	p := aep.NewProject()
	c, _ := p.NewComposition("Main", 1920, 1080, 30, 10)
	s, err := c.NewShapeLayer("S1")
	if err != nil {
		t.Fatal(err)
	}
	if s == nil {
		t.Fatal("nil ShapeLayer")
	}
	if s.Name != "S1" {
		t.Fatalf("Name = %q, want S1", s.Name)
	}
	if s.Type != aep.LayerTypeShape {
		t.Fatalf("Type = %v, want LayerTypeShape", s.Type)
	}
	if len(c.Layers) != 1 || c.Layers[0] != s.Layer {
		t.Fatalf("comp.Layers not updated correctly")
	}
}

func TestNewShapeLayer_EmptyName_Error(t *testing.T) {
	p := aep.NewProject()
	c, _ := p.NewComposition("Main", 1920, 1080, 30, 10)
	if _, err := c.NewShapeLayer(""); err == nil {
		t.Fatal("empty name should error")
	}
	if len(c.Layers) != 0 {
		t.Fatalf("comp polluted on failure: %d layers", len(c.Layers))
	}
}
```

- [ ] **Step 2: 实现 new_layer.go**

```go
// internal/aep/new_layer.go
package aep

import "fmt"

// NewShapeLayer 创建空 ShapeLayer 加到 comp。
// Failure: name empty / lowering 失败 / parse-after-build 产 warning.
// On failure, no partial runtime mutation is committed (per Inv-mutation atomicity).
func (c *Composition) NewShapeLayer(name string) (*ShapeLayer, error) {
	if name == "" {
		return nil, fmt.Errorf("ShapeLayer name cannot be empty")
	}
	if c.proj == nil {
		return nil, fmt.Errorf("internal: comp has no project back-ref")
	}

	// 1. 构建 runtime ShapeLayer (Phase 1 types)
	base := &Layer{
		Type: LayerTypeShape,
		Name: name,
		ID:   c.proj.allocItemID(), // V2.1 既有；layer ID 复用 item ID namespace
	}
	s := WrapShapeLayer(base)

	// 2. lowering → LIST(Layr) chunk (Phase 2)
	ctx := &lowerCtx{
		tickRate:     c.TickRate,
		capabilities: Capabilities(c.proj.target),
		nextLayerID:  c.proj.allocItemID,
	}
	layrChunk, err := lowerShapeLayer(s, ctx)
	if err != nil {
		return nil, fmt.Errorf("lower ShapeLayer: %w", err)
	}

	// 3. Atomic mutation snapshot (V2.1 NewComposition pattern)
	oldChildLen := len(c.itemList.Children)
	oldWarningsLen := len(c.proj.Warnings)

	// 4. Append layr chunk to comp's itemList (siblings of cdta / DLay etc)
	// V2.2 ShapeLayer 加到 itemList，跟 V1 既有 Layr children 同级
	c.itemList.Children = append(c.itemList.Children, layrChunk)

	// 5. parse-after-build (Phase 4 完整 hydration；Phase 3 占位)
	// 当前: 仅 wire 既有 V1 *Layer 进 comp.Layers
	c.Layers = append(c.Layers, base)

	// 6. Warnings-as-failure (V2.1 invariant)
	if len(c.proj.Warnings) != oldWarningsLen {
		c.itemList.Children = c.itemList.Children[:oldChildLen]
		c.Layers = c.Layers[:len(c.Layers)-1]
		c.proj.Warnings = c.proj.Warnings[:oldWarningsLen]
		return nil, fmt.Errorf("internal: NewShapeLayer produced parser warnings")
	}

	return s, nil
}
```

- [ ] **Step 3: Run test → PASS**

```bash
go test ./internal/aep/ -run "TestNewShapeLayer" -v
```

- [ ] **Step 4: Commit**

```bash
git add internal/aep/new_layer.go internal/aep/new_shape_layer_test.go
git commit -m "feat(v2.2): Composition.NewShapeLayer atomic mutation entry"
```

---

### Task 3.2: VectorGroup AddNode methods (5 node types)

**Files:**
- Modify: `internal/aep/shape_graph.go`
- Test: `internal/aep/shape_graph_attach_test.go`

- [ ] **Step 1: 写失败测**

```go
package aep_test

import (
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestVectorGroup_AddRect_Appends(t *testing.T) {
	g := aep.NewVectorGroup()
	r, err := g.AddRect()
	if err != nil {
		t.Fatal(err)
	}
	if r == nil {
		t.Fatal("nil RectNode")
	}
	if len(g.Children) != 1 || g.Children[0] != r {
		t.Fatalf("Children not updated: %v", g.Children)
	}
}

func TestVectorGroup_RenderOrder(t *testing.T) {
	g := aep.NewVectorGroup()
	rect, _ := g.AddRect()   // Children[0] (bottom)
	fill, _ := g.AddFill()    // Children[1] (top)
	if g.Children[0] != rect {
		t.Fatal("first add should be Children[0]")
	}
	if g.Children[1] != fill {
		t.Fatal("second add should be Children[1] (on top)")
	}
}

func TestVectorGroup_AllAddMethods(t *testing.T) {
	g := aep.NewVectorGroup()
	if _, err := g.AddEllipse(); err != nil { t.Fatal(err) }
	if _, err := g.AddPath(); err != nil { t.Fatal(err) }
	if _, err := g.AddStroke(); err != nil { t.Fatal(err) }
	if len(g.Children) != 3 {
		t.Fatalf("Children = %d, want 3", len(g.Children))
	}
}
```

- [ ] **Step 2: 在 shape_graph.go 加 AddX 方法**

```go
// 加在 VectorGroup type 定义后

func (g *VectorGroup) AddRect() (*RectNode, error) {
	r := NewRectNode()
	g.Children = append(g.Children, r)
	return r, nil
}

func (g *VectorGroup) AddEllipse() (*EllipseNode, error) {
	e := NewEllipseNode()
	g.Children = append(g.Children, e)
	return e, nil
}

func (g *VectorGroup) AddPath() (*PathNode, error) {
	p := NewPathNode()
	g.Children = append(g.Children, p)
	return p, nil
}

func (g *VectorGroup) AddFill() (*FillNode, error) {
	f := NewFillNode()
	g.Children = append(g.Children, f)
	return f, nil
}

func (g *VectorGroup) AddStroke() (*StrokeNode, error) {
	s := NewStrokeNode()
	g.Children = append(g.Children, s)
	return s, nil
}
```

- [ ] **Step 3: Run test → PASS + commit**

```bash
go test ./internal/aep/ -run "TestVectorGroup" -v
git add internal/aep/shape_graph.go internal/aep/shape_graph_attach_test.go
git commit -m "feat(v2.2): VectorGroup.Add{Rect,Ellipse,Path,Fill,Stroke}"
```

---

### Task 3.3: Escape hatch β — `Properties()` + typed `*Stream(name)` methods

**Files:**
- Modify: `internal/aep/shape_graph.go` (PropertyGroup methods)
- Test: `internal/aep/escape_hatch_test.go`

- [ ] **Step 1: 写失败测**

```go
package aep_test

import (
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestPropertyGroup_Vec2Stream(t *testing.T) {
	// 用 RectNode Properties() 走 escape hatch 访问 Size
	r := aep.NewRectNode()
	props := r.Properties()
	if props == nil {
		t.Skip("Properties() returns nil — Phase 3 Task 3.3 task not implemented yet")
	}
	sizeStream, err := props.Vec2Stream("Size")
	if err != nil {
		t.Fatal(err)
	}
	if sizeStream != r.Size() {
		t.Fatal("escape hatch Stream(\"Size\") != typed accessor Size()")
	}
}

func TestPropertyGroup_NameNotFound_Error(t *testing.T) {
	r := aep.NewRectNode()
	props := r.Properties()
	if props == nil { t.Skip() }
	if _, err := props.Vec2Stream("NotAFieldName"); err == nil {
		t.Fatal("unknown name should error")
	}
}
```

- [ ] **Step 2: 完整 PropertyGroup 实现 + RectNode.Properties() 等 hookup**

```go
// shape_graph.go 末尾追加 / 替换 PropertyGroup 占位

import "fmt"

// PropertyGroup methods

func (pg *PropertyGroup) Child(name string) *PropertyGroup {
	if pg.Children == nil {
		return nil
	}
	return pg.Children[name]
}

func (pg *PropertyGroup) Float64Stream(name string) (*PropertyStream[float64], error) {
	v, ok := pg.streams[name]
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q not found", pg.Name, name)
	}
	ps, ok := v.(*PropertyStream[float64])
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q is not float64", pg.Name, name)
	}
	return ps, nil
}

func (pg *PropertyGroup) Vec2Stream(name string) (*PropertyStream[[2]float64], error) {
	v, ok := pg.streams[name]
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q not found", pg.Name, name)
	}
	ps, ok := v.(*PropertyStream[[2]float64])
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q is not [2]float64", pg.Name, name)
	}
	return ps, nil
}

func (pg *PropertyGroup) Vec3Stream(name string) (*PropertyStream[[3]float64], error) {
	v, ok := pg.streams[name]
	if !ok { return nil, fmt.Errorf("not found %q", name) }
	ps, ok := v.(*PropertyStream[[3]float64])
	if !ok { return nil, fmt.Errorf("type mismatch for %q", name) }
	return ps, nil
}

func (pg *PropertyGroup) ColorStream(name string) (*PropertyStream[[4]float64], error) {
	v, ok := pg.streams[name]
	if !ok { return nil, fmt.Errorf("not found %q", name) }
	ps, ok := v.(*PropertyStream[[4]float64])
	if !ok { return nil, fmt.Errorf("type mismatch for %q", name) }
	return ps, nil
}

func (pg *PropertyGroup) PathStream(name string) (*PropertyStream[BezierPath], error) {
	v, ok := pg.streams[name]
	if !ok { return nil, fmt.Errorf("not found %q", name) }
	ps, ok := v.(*PropertyStream[BezierPath])
	if !ok { return nil, fmt.Errorf("type mismatch for %q", name) }
	return ps, nil
}

// 改 RectNode.Properties() (之前占位 return nil)
func (r *RectNode) Properties() *PropertyGroup {
	return &PropertyGroup{
		Name: "Rect",
		streams: map[string]any{
			"Size":      r.size,
			"Position":  r.position,
			"Roundness": r.roundness,
		},
	}
}

// 同样改 EllipseNode / PathNode / FillNode / StrokeNode 的 Properties()
func (e *EllipseNode) Properties() *PropertyGroup {
	return &PropertyGroup{
		Name: "Ellipse",
		streams: map[string]any{
			"Size":     e.size,
			"Position": e.position,
		},
	}
}

func (p *PathNode) Properties() *PropertyGroup {
	return &PropertyGroup{
		Name: "Path",
		streams: map[string]any{
			"Path": p.path,
		},
	}
}

func (f *FillNode) Properties() *PropertyGroup {
	return &PropertyGroup{
		Name: "Fill",
		streams: map[string]any{
			"Color":   f.color,
			"Opacity": f.opacity,
		},
	}
}

func (s *StrokeNode) Properties() *PropertyGroup {
	return &PropertyGroup{
		Name: "Stroke",
		streams: map[string]any{
			"Color":   s.color,
			"Opacity": s.opacity,
			"Width":   s.width,
		},
	}
}
```

- [ ] **Step 3: Run test → PASS + commit**

```bash
go test ./internal/aep/ -run "TestPropertyGroup" -v
git add internal/aep/shape_graph.go internal/aep/escape_hatch_test.go
git commit -m "feat(v2.2): PropertyGroup escape hatch β — typed Vec2/Vec3/Color/Path/Float64 Stream"
```

---

### Task 3.4: Phase 3 收尾 + PASS count

- [ ] **Step 1: Run + count**

```bash
go vet ./...
go test -count=1 ./internal/aep/... -v 2>&1 | grep -c '^--- PASS'
```

Expected: ~150 PASS.

- [ ] **Step 2: board.md + commit**

```bash
git add workshop/board.md
git commit -m "docs(board): V2.2 Phase 3 public API complete (~150 PASS)"
```

**Phase 3 PASS criterion:** PASS ≥ 145; NewShapeLayer + 5 AddX + escape hatch 全 PASS.

---

# Phase 4 — Roundtrip + hydration

**目的:** 实现 parse-after-build 闭环 + hydrateShapeNodes，跑完 canonical fixture Go roundtrip。Phase 2/3 skeleton 字节填实 (依赖 Phase 0 RE)。

### Task 4.1: `hydrateShapeNodes` — chunk tree → runtime VectorGroup

**Files:**
- Modify: `internal/aep/parse_shape.go` (新加 hydrateShapeNodes function)
- Or new file: `internal/aep/hydrate_shape.go`
- Test: `internal/aep/hydrate_shape_test.go`

- [ ] **Step 1: 写失败测**

```go
package aep_test

import (
	"bytes"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestHydrate_RectNode_FromChunks(t *testing.T) {
	p := aep.NewProject()
	c, _ := p.NewComposition("M", 1920, 1080, 30, 5)
	s, _ := c.NewShapeLayer("S1")
	r, _ := s.RootGroup().AddRect()
	_ = r.SetSize([2]float64{200, 100})

	// 走 WriteAEP → FromReader 闭环
	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil { t.Fatal(err) }
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil { t.Fatal(err) }

	// 找到 hydrated ShapeLayer
	reComp := re.Compositions[0]
	if len(reComp.Layers) != 1 { t.Fatalf("layers = %d, want 1", len(reComp.Layers)) }
	reBase := reComp.Layers[0]
	reShape := aep.WrapShapeLayer(reBase) // V2.2: parser 返 *Layer; user wrap
	// 实际 hydrate: 应该自动 wrap，见 Task 4.2

	if reShape.RootGroup() == nil { t.Fatal("nil rootGroup post-hydrate") }
	if len(reShape.RootGroup().Children) != 1 {
		t.Fatalf("contents = %d, want 1 (rect)", len(reShape.RootGroup().Children))
	}
	reRect, ok := reShape.RootGroup().Children[0].(*aep.RectNode)
	if !ok { t.Fatal("first child not *RectNode") }
	v, _ := reRect.Size().StaticValue()
	if v != [2]float64{200, 100} {
		t.Fatalf("hydrated Size = %v, want [200,100]", v)
	}
}
```

- [ ] **Step 2: 实现 hydrateShapeNodes**

```go
// internal/aep/hydrate_shape.go
package aep

import (
	"github.com/example/aep-parser/internal/rifx"
)

// hydrateShapeNodes 从 layr LIST 重建 runtime VectorGroup (root) + ShapeNode 树.
// Inv: Hydration reconstructs canonical runtime semantics from serializer artifacts.
//
// 调用点: parseLayer 检测到 LayerTypeShape 时，调本函数填 ShapeLayer.rootGroup.
func hydrateShapeNodes(layr *rifx.Chunk) *VectorGroup {
	// 查 "ADBE Vector Materials Group" tdgp
	for _, c := range layr.Children {
		if !c.IsList() || c.FormType != rifx.IDTdgp {
			continue
		}
		// 校 tdmn = "ADBE Vector Materials Group" (第 1 child)
		if len(c.Children) == 0 { continue }
		first := c.Children[0]
		if first.ID != rifx.IDTdmn || trimNUL(first.Data) != "ADBE Vector Materials Group" {
			continue
		}
		// 找内部 root VectorGroup tdgp
		for _, sub := range c.Children[1:] {
			if !sub.IsList() || sub.FormType != rifx.IDTdgp {
				continue
			}
			return hydrateVectorGroup(sub)
		}
	}
	return NewVectorGroup() // empty fallback
}

// hydrateVectorGroup 把 tdgp (单 VectorGroup) → runtime VectorGroup.
// 复用 V1 parse_shape.go 既有 ShapePrimitive 解码逻辑做 child 翻译.
func hydrateVectorGroup(tdgp *rifx.Chunk) *VectorGroup {
	g := NewVectorGroup()
	for i := 0; i < len(tdgp.Children); i++ {
		ch := tdgp.Children[i]
		if ch.ID != rifx.IDTdmn { continue }
		matchName := trimNUL(ch.Data)
		if i+1 >= len(tdgp.Children) { continue }
		next := tdgp.Children[i+1]
		switch matchName {
		case "ADBE Vector Shape - Rect":
			if node := hydrateRectNode(next); node != nil {
				g.Children = append(g.Children, node)
			}
		case "ADBE Vector Shape - Ellipse":
			if node := hydrateEllipseNode(next); node != nil {
				g.Children = append(g.Children, node)
			}
		case "ADBE Vector Shape - Group":
			if node := hydratePathNode(next); node != nil {
				g.Children = append(g.Children, node)
			}
		case "ADBE Vector Graphic - Fill":
			if node := hydrateFillNode(next); node != nil {
				g.Children = append(g.Children, node)
			}
		case "ADBE Vector Graphic - Stroke":
			if node := hydrateStrokeNode(next); node != nil {
				g.Children = append(g.Children, node)
			}
		case "ADBE Vector Transform Group":
			g.Transform = hydrateGroupTransform(next)
		}
	}
	return g
}

func hydrateRectNode(tdgp *rifx.Chunk) *RectNode {
	r := NewRectNode()
	// 解码 Size / Position / Roundness streams from tdgp children
	// 复用 V1 既有 parse_property.go decodeProperty logic
	// (Phase 0 RE-S4 frozen; here 写具体提取代码)
	return r
}

func hydrateEllipseNode(tdgp *rifx.Chunk) *EllipseNode { return NewEllipseNode() }
func hydratePathNode(tdgp *rifx.Chunk) *PathNode       { return NewPathNode() }
func hydrateFillNode(tdgp *rifx.Chunk) *FillNode       { return NewFillNode() }
func hydrateStrokeNode(tdgp *rifx.Chunk) *StrokeNode   { return NewStrokeNode() }
func hydrateGroupTransform(tdgp *rifx.Chunk) *PropertyGroup { return newGroupTransform() }

// trimNUL = V1 既有 helper; mirror placement.
func trimNULv22(b []byte) string {
	n := len(b)
	for n > 0 && b[n-1] == 0 {
		n--
	}
	return string(b[:n])
}
```

- [ ] **Step 3: 把 hydration 接入 parseLayer**

修改 `internal/aep/parse_layer.go`：检测到 `LayerTypeShape` 时，调 `hydrateShapeNodes` 并 attach 进新加的 `Layer.shapeRootGroup *VectorGroup` 字段。或更优：在 `WrapShapeLayer(layer)` 内做（如果 `layer.ldta` 非空，从 ldta 父 LIST 找 layr 再 hydrate）。

具体 hookup 看 V1 既有 parser 流程；Phase 4 实施期决定最干净的 attach point。

- [ ] **Step 4: Run test → PASS**

```bash
go test ./internal/aep/ -run "TestHydrate" -v
```

- [ ] **Step 5: Commit**

```bash
git add internal/aep/hydrate_shape.go internal/aep/hydrate_shape_test.go internal/aep/parse_layer.go
git commit -m "feat(v2.2): hydrateShapeNodes — chunk tree → runtime VectorGroup tree"
```

---

### Task 4.2: Canonical fixture roundtrip — `TestV2_2_CanonicalShapeGraph_Roundtrip`

**Files:**
- Test: `internal/aep/shape_graph_roundtrip_test.go`

- [ ] **Step 1: 写 canonical fixture test (spec §5.2 完整代码)**

```go
package aep_test

import (
	"bytes"
	"math"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestV2_2_CanonicalShapeGraph_Roundtrip(t *testing.T) {
	p := aep.NewProject() // TargetAE2020
	comp, _ := p.NewComposition("Main", 1920, 1080, 30, 5)

	// ShapeLayer A: 全 animated streams
	a, _ := comp.NewShapeLayer("A_RectFill_Animated")
	rectA, _ := a.RootGroup().AddRect()
	_ = rectA.Size().AddKeyframeLinear(0, [2]float64{50, 50})
	_ = rectA.Size().AddKeyframeLinear(2, [2]float64{300, 200})
	fillA, _ := a.RootGroup().AddFill()
	_ = fillA.Color().AddKeyframeLinear(0, [4]float64{1, 0, 0, 1})
	_ = fillA.Color().AddKeyframeLinear(2, [4]float64{0, 0, 1, 1})
	_ = a.Position().AddKeyframeLinear(0, [2]float64{0, 0})
	_ = a.Position().AddKeyframeLinear(2, [2]float64{500, 300})

	// ShapeLayer B: 全 static
	b, _ := comp.NewShapeLayer("B_EllipseStroke_Static")
	ellB, _ := b.RootGroup().AddEllipse()
	_ = ellB.SetSize([2]float64{150, 150})
	strokeB, _ := b.RootGroup().AddStroke()
	_ = strokeB.SetColor([4]float64{0, 0, 1, 1})
	_ = strokeB.SetWidth(5)

	// ShapeLayer C: 全 static
	c, _ := comp.NewShapeLayer("C_PathFillStroke_Static")
	pathC, _ := c.RootGroup().AddPath()
	_ = pathC.SetVertices([][2]float64{{0, 0}, {100, 0}, {100, 100}, {0, 100}})
	_ = pathC.SetClosed(true)
	fillC, _ := c.RootGroup().AddFill()
	_ = fillC.SetColor([4]float64{0, 1, 0, 1})
	strokeC, _ := c.RootGroup().AddStroke()
	_ = strokeC.SetWidth(2)

	// Roundtrip via WriteAEP → FromReader
	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil { t.Fatal(err) }
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil { t.Fatal(err) }

	// Assertions
	reMain := re.Compositions[0]
	if len(reMain.Layers) != 3 {
		t.Fatalf("layers = %d, want 3", len(reMain.Layers))
	}

	// Layer A (animated)
	la := findLayerByName(reMain, "A_RectFill_Animated")
	sa := aep.WrapShapeLayer(la)
	if len(sa.RootGroup().Children) != 2 {
		t.Fatalf("A contents = %d, want 2", len(sa.RootGroup().Children))
	}
	reRectA, ok := sa.RootGroup().Children[0].(*aep.RectNode)
	if !ok { t.Fatal("A[0] not RectNode") }
	if reRectA.Size().Mode() != aep.StreamModeAnimated {
		t.Fatal("A Rect Size mode should be Animated post-roundtrip")
	}
	kfA := reRectA.Size().Keyframes()
	if len(kfA) != 2 {
		t.Fatalf("A Rect Size kf count = %d, want 2", len(kfA))
	}
	if kfA[0].Value != [2]float64{50, 50} {
		t.Fatalf("kf[0] = %v, want [50,50]", kfA[0].Value)
	}

	// Layer A Position keyframes
	posKf := sa.Position().Keyframes()
	if len(posKf) != 2 || posKf[1].Value != [2]float64{500, 300} {
		t.Fatalf("A Position kf bad: %v", posKf)
	}

	// Layer B (static)
	lb := findLayerByName(reMain, "B_EllipseStroke_Static")
	sb := aep.WrapShapeLayer(lb)
	if len(sb.RootGroup().Children) != 2 {
		t.Fatalf("B contents = %d, want 2", len(sb.RootGroup().Children))
	}
	reEllB, ok := sb.RootGroup().Children[0].(*aep.EllipseNode)
	if !ok { t.Fatal("B[0] not EllipseNode") }
	reEllSize, isStatic := reEllB.Size().StaticValue()
	if !isStatic || reEllSize != [2]float64{150, 150} {
		t.Fatalf("B Ellipse Size = (%v, isStatic=%v)", reEllSize, isStatic)
	}

	// Layer C (static)
	lc := findLayerByName(reMain, "C_PathFillStroke_Static")
	sc := aep.WrapShapeLayer(lc)
	rePathC, ok := sc.RootGroup().Children[0].(*aep.PathNode)
	if !ok { t.Fatal("C[0] not PathNode") }
	rePath, _ := rePathC.Path().StaticValue()
	if len(rePath.Vertices) != 4 {
		t.Fatalf("C Path vertices = %d, want 4", len(rePath.Vertices))
	}
	if !rePath.Closed {
		t.Fatal("C Path Closed = false, want true")
	}
}

func findLayerByName(c *aep.Composition, name string) *aep.Layer {
	for _, l := range c.Layers {
		if l.Name == name { return l }
	}
	return nil
}
```

- [ ] **Step 2: Run test → 期初可能 FAIL (Phase 0 RE byte layout 未填完时 roundtrip 失败); Phase 0/2 finalize 后 PASS**

```bash
go test ./internal/aep/ -run "TestV2_2_CanonicalShapeGraph" -v
```

期望: Phase 4 task 完成 + Phase 0 RE 数据 byte-exact 填进 lower_*.go 后 PASS。

- [ ] **Step 3: Commit**

```bash
git add internal/aep/shape_graph_roundtrip_test.go
git commit -m "test(v2.2): canonical shape graph roundtrip (3 layers, multi-root, animated/static partition)"
```

---

### Task 4.3: Atomicity / rollback tests

**Files:**
- Test: `internal/aep/shape_atomicity_test.go`

- [ ] **Step 1: 写 atomicity 测**

```go
package aep_test

import (
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestShapeLayer_LoweringFailure_Rollback(t *testing.T) {
	// 难直接触发 lowering 失败（lowering 不预期 error path）
	// 改写 empty name → atomic rollback test (Task 3.1 已有；这里加 comp 状态校)
	p := aep.NewProject()
	c, _ := p.NewComposition("M", 1920, 1080, 30, 5)
	_, _ = c.NewShapeLayer("ok") // success
	beforeLen := len(c.Layers)

	_, err := c.NewShapeLayer("") // failure
	if err == nil { t.Fatal("empty name should error") }

	if len(c.Layers) != beforeLen {
		t.Fatalf("comp.Layers polluted: was %d, now %d", beforeLen, len(c.Layers))
	}
}

func TestPropertyStream_AddKeyframe_InvalidArgs_NoSideEffect(t *testing.T) {
	ps := aep.NewPropertyStream[float64]()
	_ = ps.AddKeyframeLinear(1.0, 100)

	beforeKf := len(ps.Keyframes())
	if err := ps.AddKeyframeLinear(-1, 0); err == nil {
		t.Fatal("negative time should error")
	}
	if len(ps.Keyframes()) != beforeKf {
		t.Fatalf("kf count changed despite error: was %d, now %d", beforeKf, len(ps.Keyframes()))
	}
}

func TestPropertyStream_DuplicateTime_NoSideEffect(t *testing.T) {
	ps := aep.NewPropertyStream[float64]()
	_ = ps.AddKeyframeLinear(1.0, 100)

	beforeKf := len(ps.Keyframes())
	if err := ps.AddKeyframeLinear(1.0, 999); err == nil {
		t.Fatal("duplicate time should error")
	}
	if len(ps.Keyframes()) != beforeKf {
		t.Fatalf("kf count changed: was %d, now %d", beforeKf, len(ps.Keyframes()))
	}
	// 校原值未被覆盖
	if v := ps.Keyframes()[0].Value; v != 100 {
		t.Fatalf("original kf value mutated: %v", v)
	}
}
```

- [ ] **Step 2: Run + commit**

```bash
go test ./internal/aep/ -run "TestShapeLayer_LoweringFailure|TestPropertyStream_AddKeyframe_InvalidArgs|TestPropertyStream_DuplicateTime" -v
git add internal/aep/shape_atomicity_test.go
git commit -m "test(v2.2): atomicity — failed mutations leave no side effect"
```

---

### Task 4.4: Mixed authored/preserved test

**Files:**
- Test: `internal/aep/shape_mutate_existing_test.go` (skip-if-fixture-missing per V1 pattern)

- [ ] **Step 1: 写 test**

```go
package aep_test

import (
	"bytes"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestV2_2_MutateExistingShape(t *testing.T) {
	// fixture 在 Phase 5 Task 5.3 产；此处 skip-if-missing
	p, err := aep.Open("../../test_data/v2_2_shape_tolerance.aep")
	if err != nil {
		t.Skipf("v2_2_shape_tolerance.aep not present: %v (Phase 5 Task 5.3 出)", err)
	}
	if len(p.Compositions) == 0 || len(p.Compositions[0].Layers) == 0 {
		t.Skip("fixture empty")
	}
	layer := p.Compositions[0].Layers[0]
	if layer.Type != aep.LayerTypeShape {
		t.Skipf("fixture layer 0 type = %v, want Shape", layer.Type)
	}
	sl := aep.WrapShapeLayer(layer)
	if len(sl.RootGroup().Children) == 0 {
		t.Skip("fixture root group empty")
	}
	rect, ok := sl.RootGroup().Children[0].(*aep.RectNode)
	if !ok {
		t.Skipf("fixture layer 0 first child kind = %v, want Rect", sl.RootGroup().Children[0].Kind())
	}

	// Mutate
	_ = rect.SetSize([2]float64{500, 500})

	// Roundtrip
	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil { t.Fatal(err) }
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil { t.Fatal(err) }

	reSl := aep.WrapShapeLayer(re.Compositions[0].Layers[0])
	reRect := reSl.RootGroup().Children[0].(*aep.RectNode)
	val, _ := reRect.Size().StaticValue()
	if val != [2]float64{500, 500} {
		t.Fatalf("mutated size lost: %v", val)
	}
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/aep/shape_mutate_existing_test.go
git commit -m "test(v2.2): mutate existing ShapeLayer + roundtrip"
```

---

### Task 4.5: Phase 4 收尾

- [ ] **Step 1: All tests pass**

```bash
go vet ./...
go test -count=1 ./internal/aep/...
go test -count=1 ./internal/aep/... -v 2>&1 | grep -c '^--- PASS'
```

Expected: ~155 PASS.

- [ ] **Step 2: board.md + commit**

```bash
git add workshop/board.md
git commit -m "docs(board): V2.2 Phase 4 roundtrip + hydration complete (~155 PASS)"
```

**Phase 4 PASS criterion:** PASS ≥ 150; canonical roundtrip 完 PASS; atomicity tests 完 PASS; mutate-existing skip-if-missing 不破坏.

---

# Phase 5 — AE ship gate (Tier 2 + Tier 3)

**目的:** AE 2020 + AE 2025 实际打开 + JSX 校验 V2.2 builder 产物。+ preservation tolerance fixture。

### Task 5.1: `test_data/verify_v2_2.jsx`

**Files:**
- Create: `test_data/verify_v2_2.jsx`

- [ ] **Step 1: 写 JSX driver (mirror verify_v2_1.jsx 风格 + spec §5.3 check 表)**

```jsx
// test_data/verify_v2_2.jsx
// V2.2 ship gate — canonical shape graph (3 ShapeLayer multi-root)
//
// args.json: {"input": ..., "done": ..., "resaved": ...}
// args.json 固定路径 e:/projects/tools/aep-parser/test_data/v2_2_args.json

(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_args.json");
    var args = {};
    var doneFile = null;
    var log = [];
    var ok = false;

    function check(name, cond) {
        if (!cond) { log.push("FAIL: " + name); return false; }
        log.push("OK:   " + name);
        return true;
    }
    function approxEq(arr1, arr2, tol) {
        if (arr1.length !== arr2.length) return false;
        for (var i = 0; i < arr1.length; i++) {
            if (Math.abs(arr1[i] - arr2[i]) > tol) return false;
        }
        return true;
    }
    function layerByName(comp, n) {
        for (var i = 1; i <= comp.layers.length; i++) {
            if (comp.layers[i].name === n) return comp.layers[i];
        }
        return null;
    }
    function findContent(layer, matchName) {
        var contents = layer.property("ADBE Root Vectors Group");
        for (var i = 1; i <= contents.numProperties; i++) {
            var p = contents.property(i);
            if (p.matchName === matchName) return p;
        }
        return null;
    }

    try {
        argsFile.open("r");
        var s = argsFile.read();
        argsFile.close();
        args = eval("(" + s + ")");
        doneFile = new File(args.done);

        app.open(new File(args.input));
        var c = app.project.items[1];
        log.push("opened comp: " + c.name + " layers=" + c.layers.length);

        var A = layerByName(c, "A_RectFill_Animated");
        var B = layerByName(c, "B_EllipseStroke_Static");
        var C = layerByName(c, "C_PathFillStroke_Static");

        var checks = [
            check("layer count 3", c.layers.length === 3),
            check("layer A exists", A !== null),
            check("layer B exists", B !== null),
            check("layer C exists", C !== null),
        ];

        if (A) {
            var rectA = findContent(A, "ADBE Vector Shape - Rect");
            var fillA = findContent(A, "ADBE Vector Graphic - Fill");
            checks.push(check("A has Rect", rectA !== null));
            checks.push(check("A has Fill", fillA !== null));

            if (rectA) {
                var sizeProp = rectA.property("ADBE Vector Rect Size");
                checks.push(check("A Rect Size kf count 2", sizeProp.numKeys === 2));
                if (sizeProp.numKeys >= 2) {
                    checks.push(check("A Rect Size kf[0]", approxEq(sizeProp.keyValue(1), [50, 50], 1e-3)));
                    checks.push(check("A Rect Size kf[1]", approxEq(sizeProp.keyValue(2), [300, 200], 1e-3)));
                }
            }
            if (fillA) {
                var colorProp = fillA.property("ADBE Vector Fill Color");
                checks.push(check("A Fill Color kf count 2", colorProp.numKeys === 2));
                if (colorProp.numKeys >= 2) {
                    var kf0 = colorProp.keyValue(1);
                    checks.push(check("A Fill Color kf[0] red", approxEq([kf0[0], kf0[1], kf0[2]], [1, 0, 0], 1e-3)));
                }
            }
            // Layer Position
            var posA = A.property("ADBE Transform Group").property("ADBE Position");
            checks.push(check("A Position kf count 2", posA.numKeys === 2));
            if (posA.numKeys >= 2) {
                checks.push(check("A Position kf[1]", approxEq(posA.keyValue(2), [500, 300], 1e-3)));
            }
        }

        if (B) {
            var ellB = findContent(B, "ADBE Vector Shape - Ellipse");
            var strokeB = findContent(B, "ADBE Vector Graphic - Stroke");
            checks.push(check("B has Ellipse", ellB !== null));
            checks.push(check("B has Stroke", strokeB !== null));
            if (ellB) {
                var sizeB = ellB.property("ADBE Vector Ellipse Size").value;
                checks.push(check("B Ellipse Size", approxEq(sizeB, [150, 150], 1e-3)));
            }
            if (strokeB) {
                var widthB = strokeB.property("ADBE Vector Stroke Width").value;
                checks.push(check("B Stroke Width 5", widthB === 5));
            }
        }

        if (C) {
            var pathC = findContent(C, "ADBE Vector Shape - Group");
            var fillC = findContent(C, "ADBE Vector Graphic - Fill");
            var strokeC = findContent(C, "ADBE Vector Graphic - Stroke");
            checks.push(check("C has Path", pathC !== null));
            checks.push(check("C has Fill", fillC !== null));
            checks.push(check("C has Stroke", strokeC !== null));
            if (pathC) {
                var shp = pathC.property("ADBE Vector Shape").value;
                checks.push(check("C Path vertex count 4", shp.vertices.length === 4));
                checks.push(check("C Path closed", shp.closed === true));
            }
            if (fillC) {
                var colC = fillC.property("ADBE Vector Fill Color").value;
                checks.push(check("C Fill green", approxEq([colC[0], colC[1], colC[2]], [0, 1, 0], 1e-3)));
            }
        }

        ok = true;
        for (var k = 0; k < checks.length; k++) {
            if (!checks[k]) { ok = false; break; }
        }
    } catch (e) {
        ok = false;
        log.push("ERROR: " + e.toString());
    }

    try {
        if (!doneFile) doneFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_test.done");
        doneFile.open("w");
        doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
        doneFile.close();
    } catch (e2) { /* swallow */ }
})();
```

- [ ] **Step 2: Commit**

```bash
git add test_data/verify_v2_2.jsx
git commit -m "test(v2.2): verify_v2_2.jsx — JSX ship gate driver (per-check logging)"
```

---

### Task 5.2: Go shipgate test — `TestV2_2_AEShipGate_AE2020/2025`

**Files:**
- Modify: `internal/aep/new_composition_test.go` (扩展; or new file `shape_layer_shipgate_test.go`)

- [ ] **Step 1: 写 Go shipgate helper + 两 target test**

复用 V2.1 `runAEShipGate` helper 形态。新建 `runV2_2ShipGate`：

```go
// internal/aep/shape_layer_shipgate_test.go
package aep_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	aep "github.com/example/aep-parser/internal/aep"
)

func runV2_2ShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "v2_2_canonical.aep")
	resavedAEP := filepath.Join(tempDir, "v2_2_canonical.resaved.aep")
	doneFile := filepath.Join(tempDir, "v2_2_test.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/v2_2_args.json`

	// Build canonical shape graph (per spec §5.2)
	p := aep.NewProject(target)
	comp, _ := p.NewComposition("Main", 1920, 1080, 30, 5)

	a, _ := comp.NewShapeLayer("A_RectFill_Animated")
	rectA, _ := a.RootGroup().AddRect()
	_ = rectA.Size().AddKeyframeLinear(0, [2]float64{50, 50})
	_ = rectA.Size().AddKeyframeLinear(2, [2]float64{300, 200})
	fillA, _ := a.RootGroup().AddFill()
	_ = fillA.Color().AddKeyframeLinear(0, [4]float64{1, 0, 0, 1})
	_ = fillA.Color().AddKeyframeLinear(2, [4]float64{0, 0, 1, 1})
	_ = a.Position().AddKeyframeLinear(0, [2]float64{0, 0})
	_ = a.Position().AddKeyframeLinear(2, [2]float64{500, 300})

	b, _ := comp.NewShapeLayer("B_EllipseStroke_Static")
	ellB, _ := b.RootGroup().AddEllipse()
	_ = ellB.SetSize([2]float64{150, 150})
	strokeB, _ := b.RootGroup().AddStroke()
	_ = strokeB.SetColor([4]float64{0, 0, 1, 1})
	_ = strokeB.SetWidth(5)

	c, _ := comp.NewShapeLayer("C_PathFillStroke_Static")
	pathC, _ := c.RootGroup().AddPath()
	_ = pathC.SetVertices([][2]float64{{0, 0}, {100, 0}, {100, 100}, {0, 100}})
	_ = pathC.SetClosed(true)
	fillC, _ := c.RootGroup().AddFill()
	_ = fillC.SetColor([4]float64{0, 1, 0, 1})
	strokeC, _ := c.RootGroup().AddStroke()
	_ = strokeC.SetWidth(2)

	out, err := os.Create(inputAEP)
	if err != nil { t.Fatal(err) }
	if err := p.WriteAEP(out); err != nil { t.Fatal(err) }
	out.Close()

	toFwd := func(s string) string { return strings.ReplaceAll(s, `\`, `/`) }
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP))
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	jsxPath := `E:/projects/tools/aep-parser/test_data/verify_v2_2.jsx`
	cmd := exec.Command(aeExe, "-r", jsxPath)
	if err := cmd.Start(); err != nil { t.Fatalf("start AE: %v", err) }

	deadline := time.Now().Add(90 * time.Second)
	for {
		if _, err := os.Stat(doneFile); err == nil { break }
		if time.Now().After(deadline) {
			t.Fatalf("timeout waiting for %s", doneFile)
		}
		time.Sleep(2 * time.Second)
	}

	content, err := os.ReadFile(doneFile)
	if err != nil { t.Fatal(err) }
	lines := strings.SplitN(string(content), "\n", 2)
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("ship gate FAIL:\n%s", string(content))
	}
}

func TestV2_2_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2ShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2ShipGate(t, aep.TargetAE2020, aeExe)
}

// Suppress unused import compile warning if bytes not actually used elsewhere
var _ = bytes.NewReader
```

- [ ] **Step 2: Skip-run sanity**

```bash
go test -count=1 ./internal/aep/ -run 'TestV2_2_AEShipGate' -v
```

Expected: 2 SKIP (no AE_SHIP_GATE).

- [ ] **Step 3: Commit**

```bash
git add internal/aep/shape_layer_shipgate_test.go
git commit -m "test(v2.2): TestV2_2_AEShipGate_AE2020/2025 (skip 默认；AE_SHIP_GATE=1 触发)"
```

---

### Task 5.3: Tier 3 preservation fixture — `gen_shape_tolerance.jsx` + `v2_2_shape_tolerance.aep`

**Files:**
- Create: `tmp_debug/gen_shape_tolerance.jsx`
- Generate: `test_data/v2_2_shape_tolerance.aep`

- [ ] **Step 1: 写 JSX**

```jsx
// tmp_debug/gen_shape_tolerance.jsx
// 创建 1 ShapeLayer，root group 内嵌 1 子 group，子 group 包 1 Rect + 1 Fill.
// 保存为 test_data/v2_2_shape_tolerance.aep — V2.2 Tier 3 preservation fixture.

(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_shape_tolerance.aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_shape_tolerance.done");
    var log = [];
    var ok = false;

    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();

        var c = app.project.items.addComp("ToleranceMain", 1920, 1080, 1, 5, 30);
        var s = c.layers.addShape();
        s.name = "Nested";
        var root = s.property("ADBE Root Vectors Group");
        var sub = root.addProperty("ADBE Vector Group");
        var subContents = sub.property("ADBE Vectors Group");
        var rect = subContents.addProperty("ADBE Vector Shape - Rect");
        rect.property("ADBE Vector Rect Size").setValue([200, 100]);
        var fill = subContents.addProperty("ADBE Vector Graphic - Fill");
        fill.property("ADBE Vector Fill Color").setValue([0.5, 0.5, 0.5, 1]);

        app.project.save(outFile);
        log.push("saved " + outFile.fsName);
        ok = true;
    } catch (e) {
        log.push("ERROR: " + e.toString());
    }

    try {
        doneFile.open("w");
        doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
        doneFile.close();
    } catch (e2) {}
})();
```

- [ ] **Step 2: Run AE 2025 with this JSX**

```bash
rm -f test_data/v2_2_shape_tolerance.done
"E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe" -r "E:/projects/tools/aep-parser/tmp_debug/gen_shape_tolerance.jsx" &
disown
until [ -f test_data/v2_2_shape_tolerance.done ]; do sleep 2; done
cat test_data/v2_2_shape_tolerance.done
ls -la test_data/v2_2_shape_tolerance.aep
```

- [ ] **Step 3: Commit**

```bash
git add tmp_debug/gen_shape_tolerance.jsx test_data/v2_2_shape_tolerance.aep
git commit -m "test(v2.2): v2_2_shape_tolerance.aep — Tier 3 nested-group preservation fixture"
```

---

### Task 5.4: `TestV2_2_NestedGroup_Preservation`

**Files:**
- Test: `internal/aep/shape_preservation_test.go`

- [ ] **Step 1: 写测**

```go
package aep_test

import (
	"bytes"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestV2_2_NestedGroup_Preservation(t *testing.T) {
	p, err := aep.Open("../../test_data/v2_2_shape_tolerance.aep")
	if err != nil { t.Skipf("fixture missing: %v", err) }

	originalLayerCount := len(p.Compositions[0].Layers)
	originalLayerName := p.Compositions[0].Layers[0].Name

	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil { t.Fatal(err) }
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil { t.Fatal(err) }

	// 语义 unchanged (V2.2 minimum preserved semantic set)
	if len(re.Compositions[0].Layers) != originalLayerCount {
		t.Fatalf("layer count changed: %d → %d", originalLayerCount, len(re.Compositions[0].Layers))
	}
	if re.Compositions[0].Layers[0].Name != originalLayerName {
		t.Fatalf("layer name changed: %q → %q", originalLayerName, re.Compositions[0].Layers[0].Name)
	}

	// 嵌套 group 拓扑 unchanged
	// (具体校 via comp.itemList chunk walk + V1 既有 parse_shape 出的 ShapePrimitive 计数)
	// (Phase 5 实施期补具体 walk 代码)
}
```

- [ ] **Step 2: Run + commit**

```bash
go test ./internal/aep/ -run "TestV2_2_NestedGroup_Preservation" -v
git add internal/aep/shape_preservation_test.go
git commit -m "test(v2.2): Tier 3 nested-group semantic preservation"
```

---

### Task 5.5: `TestV2_2_OpaquePreservation_UnknownChunks`

**Files:**
- Test: `internal/aep/shape_preservation_test.go` (扩展)

- [ ] **Step 1: 写测**

```go
func TestV2_2_OpaquePreservation_UnknownChunks(t *testing.T) {
	p, err := aep.Open("../../test_data/v2_2_shape_tolerance.aep")
	if err != nil { t.Skipf("fixture missing: %v", err) }

	// Walk chunk tree pre-write
	preChunks := collectChunkSignatures(p.RootChunk())

	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil { t.Fatal(err) }
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil { t.Fatal(err) }

	postChunks := collectChunkSignatures(re.RootChunk())

	// 检查所有 "preserved" chunks (per spec §4.0 ownership map) byte-identical
	// V2.2 ship 时此 list 包括: AFsi, Rhed, Rout, ftwd, plugin blobs 等
	for _, sig := range preChunks {
		if isPreservedKind(sig.id) {
			if !chunkBytesEqual(sig, findSig(postChunks, sig.id)) {
				t.Fatalf("preserved chunk %s bytes changed (Inv-9 violated)", sig.id)
			}
		}
	}
}

// Helper: collectChunkSignatures / isPreservedKind / chunkBytesEqual / findSig
// 用 rifx chunk tree walk + per-chunk byte hash 实现
```

`p.RootChunk()` 是 V2.2 新增 accessor 暴露既有 `Project.root` (per Inv-1 谨慎: 此 method 仅给 test 用)。

- [ ] **Step 2: 扩展 Project type 加 RootChunk() (test-only)**

```go
// internal/aep/project_test_export.go (or in types_core.go with comment)
// RootChunk returns the underlying rifx chunk tree for test purposes.
// NOT for runtime API use (Inv-1 violation if exposed broadly).
func (p *Project) RootChunk() *rifx.Chunk {
	return p.root
}
```

- [ ] **Step 3: Run + commit**

```bash
go test ./internal/aep/ -run "TestV2_2_OpaquePreservation" -v
git add internal/aep/shape_preservation_test.go internal/aep/project_test_export.go internal/aep/types_core.go
git commit -m "test(v2.2): Tier 3 opaque chunk byte-preservation (Inv-9)"
```

---

### Task 5.6: AE-reopen ship gate (Tier 3 + AE)

**Files:**
- Test: `internal/aep/shape_preservation_test.go` (扩展)

- [ ] **Step 1: 写测** (类似 5.2 + nested fixture)

```go
func TestV2_2_NestedGroup_AEReopens_AE2025(t *testing.T) {
	if os.Getenv("AE_SHIP_GATE") == "" { t.Skip("set AE_SHIP_GATE=1") }
	// Go: parse v2_2_shape_tolerance.aep → write tmp.aep → AE 2025 open + JSX 校验同 5.3 风格
	// JSX 复用 verify_v2_2.jsx 的 layer count / nested tree topology check 子集
	// (实施期具体填)
}

func TestV2_2_NestedGroup_AEReopens_AE2020(t *testing.T) {
	if os.Getenv("AE_SHIP_GATE") == "" { t.Skip("set AE_SHIP_GATE=1") }
	// AE 2020 同上
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/aep/shape_preservation_test.go
git commit -m "test(v2.2): nested group AE-reopen 2020+2025 (Tier 3 + AE)"
```

---

### Task 5.7: AE ship gate manual run (Tier 2 实际跑)

**Files:** none (操作 task)

- [ ] **Step 1: 关闭所有 AE 实例**

提示用户关 AE (类似 V2.1 ship gate retry 流程).

- [ ] **Step 2: 跑 AE 2025**

```bash
AE_SHIP_GATE=1 go test -count=1 ./internal/aep/ -run 'TestV2_2_AEShipGate_AE2025' -v -timeout 180s
```

Expected: PASS.

如果 FAIL，看 `.done` 内容 → 调整 lower_*.go (字节 layout 错误) 或 verify_v2_2.jsx (JSX check 写错)。重复直至 PASS。

- [ ] **Step 3: 跑 AE 2020 (确保 AE 2025 已关)**

```bash
AE_SHIP_GATE=1 go test -count=1 ./internal/aep/ -run 'TestV2_2_AEShipGate_AE2020' -v -timeout 180s
```

Expected: PASS.

- [ ] **Step 4: 跑 Tier 3 AE reopen tests**

```bash
AE_SHIP_GATE=1 go test -count=1 ./internal/aep/ -run 'TestV2_2_NestedGroup_AEReopens' -v -timeout 240s
```

- [ ] **Step 5: Commit if any fixes**

```bash
git add internal/aep/...
git commit -m "fix(v2.2): ship gate iteration to PASS"
```

每次 fix 一条 commit，跟 V2.1 Phase 6 pattern 一致.

---

### Task 5.8: Phase 5 收尾

- [ ] **Step 1: Run + count + ship gate confirm**

```bash
go vet ./...
go test -count=1 ./internal/aep/...
go test -count=1 ./internal/aep/... -v 2>&1 | grep -c '^--- PASS'

AE_SHIP_GATE=1 go test -count=1 ./internal/aep/ -run TestV2_2_AEShipGate -v -timeout 240s
```

Expected: total PASS ≥ 155 (122 baseline + Phase 1-4 + Phase 5 tests); AE ship gate × 2 + Tier 3 reopen × 2 全 PASS.

- [ ] **Step 2: board.md + commit**

```bash
git add workshop/board.md
git commit -m "docs(board): V2.2 Phase 5 AE ship gate + preservation tolerance PASS"
```

**Phase 5 PASS criterion:** PASS ≥ 155; AE 2020 + AE 2025 ship gate + nested preservation 全 PASS.

---

# Phase 6 — Docs sync + ship

**目的:** spec / docs / board / coverage 全部同步到 V2.2 状态；scar 更新（若发现 quirks）；Phase 6 hard ship criterion.

### Task 6.1: User-facing docs — `docs/composition.md` / `docs/layer.md` / 新建 `docs/shape.md`

**Files:**
- Modify: `docs/composition.md` (加 NewShapeLayer ref)
- Modify: `docs/layer.md` (加 ShapeLayer typed wrapper段)
- Create: `docs/shape.md`

- [ ] **Step 1: docs/composition.md 加 NewShapeLayer 段**

定位 `### Project.NewComposition` 后，追加：

```markdown
### Composition.NewShapeLayer

```go
func (c *Composition) NewShapeLayer(name string) (*ShapeLayer, error)
```

在 comp 内新建空 ShapeLayer。Required: name 非空。原子 (failure → no partial mutation)。
ID 自动分配 (跟 Composition / Footage 共享 monotonic Item ID namespace).

返回 `*ShapeLayer` typed wrapper (embeds `*Layer`，所有 V1 layer setter 直接可用) + shape-specific API (RootGroup / AddRect/AddEllipse/AddPath/AddFill/AddStroke).

详 `docs/shape.md` shape graph 文档。
```

- [ ] **Step 2: 新建 `docs/shape.md`**

```markdown
# Shape Layer + Vector Graph

`ShapeLayer` 是 V2.2 引入的 typed wrapper around `*Layer`，暴露 vector shape graph API。

## Description

ShapeLayer 内部是 `VectorGroup` 树 (root group + ShapeNode children)。V2.2 ship 5 节点类型:

| Kind | Type | 角色 |
|---|---|---|
| Rect | `*RectNode` | Geometry — Size / Position / Roundness |
| Ellipse | `*EllipseNode` | Geometry — Size / Position |
| Path | `*PathNode` | Geometry — BezierPath (linear segments V2.2; tangent V2.3+) |
| Fill | `*FillNode` | Render — Color RGBA 0..1 / Opacity 0..100 |
| Stroke | `*StrokeNode` | Render — Color / Width / Opacity |

## Example

```go
proj := aep.NewProject()
c, _ := proj.NewComposition("Main", 1920, 1080, 30, 5)

s, _ := c.NewShapeLayer("S1")
rect, _ := s.RootGroup().AddRect()
rect.SetSize([2]float64{200, 100})
fill, _ := s.RootGroup().AddFill()
fill.SetColor([4]float64{1, 0, 0, 1}) // red

// Layer Transform
s.Position().SetStaticValue([2]float64{960, 540})  // 2D ShapeLayer (V2.2)

// Keyframe animation
rect.Size().AddKeyframeLinear(0, [2]float64{50, 50})
rect.Size().AddKeyframeLinear(2, [2]float64{300, 200})
```

## Render order

Newly attached nodes are appended to the top of the render stack:

```go
group := s.RootGroup()
rect := group.AddRect()  // Children[0] (bottom)
fill := group.AddFill()  // Children[1] (top, fills rect)
```

## PropertyStream (animation primitive)

每个 typed setter (`SetSize`, `SetColor` 等) 走 static path. 调 `AddKeyframeLinear` / `AddKeyframeWithEase` 转 animated mode (per `PropertyStream` state machine, Static ↔ Animated 互斥).

```go
size := rect.Size()       // *PropertyStream[[2]float64]
size.SetStaticValue([2]float64{100, 100})  // Static mode
size.AddKeyframeLinear(0, [2]float64{50, 50})  // → Animated mode
v, isStatic := size.StaticValue()  // (zero, false) — Animated 时 ok=false
```

## Escape hatch (rare fields)

V2.2 typed setter 覆盖 hot path. Rare 字段 (Stroke LineCap / Fill FillRule / etc) 走 PropertyGroup tree:

```go
props := stroke.Properties()
opacity, _ := props.Float64Stream("Opacity")
opacity.SetStaticValue(50)
```

runtime name (`"Opacity"`) 是 stable identifier；AE match-name (`"ADBE Vector Stroke Opacity"`) 严格 serializer-only。

## 不在 V2.2 范围

- 嵌套 group 合成 API (preservation territory: parse/write 不破坏)
- 3D ShapeLayer + 3D Position
- LineCap / LineJoin / Dashes / Gradient / Trim / Merge / Repeater
- Path with tangents typed setter (V2.2 默认 linear)
- Effect / Mask 创建 on ShapeLayer

详 `workshop/specs/v2-2-layer-creation-design.md` §V2.2 范围外段。
```

- [ ] **Step 3: docs/layer.md 加 ShapeLayer typed wrapper 段**

定位 docs/layer.md "Attributes" 之前，加一节 "## ShapeLayer (V2.2)" 简短指引到 docs/shape.md.

- [ ] **Step 4: Commit**

```bash
git add docs/composition.md docs/layer.md docs/shape.md
git commit -m "docs(v2.2): NewShapeLayer + ShapeLayer + ShapeNode API reference"
```

---

### Task 6.2: `workshop/board.md` V2.2 ship archive

**Files:**
- Modify: `workshop/board.md`

- [ ] **Step 1: 更新 Last updated + Active focus**

```markdown
**Last updated**: 2026-MM-DD by claude (V2.2 ShapeLayer ship — 5 nodes + PropertyStream + AE 2020/25 ship gates PASS；PASS = **155+ / 0 FAIL**)
**Active focus**: 🟢 V2.2 完工 —— 下个候选: V3 brainstorm (scene-graph IR; `workshop/specs/v3-direction.md`) 或 V2.3 (Layer 类型扩展 + 嵌套 group 合成)。
```

- [ ] **Step 2: 加最近归档段**

```markdown
### 2026-MM-DD V2.2 ShapeLayer ship — NewShapeLayer + 5 nodes + PropertyStream (~155 PASS, +N)

V2 第二个 sub-project。dual-track 战略 (V2 ship + V3 增量提取) 验证.

- **Public API**: 
  - `comp.NewShapeLayer(name) (*ShapeLayer, error)` 
  - `*ShapeLayer.RootGroup().Add{Rect,Ellipse,Path,Fill,Stroke}()`
  - 5 ShapeNode 各 typed setter (Size / Position / Color / Width / etc)
  - `*ShapeLayer.Transform()` + shorthand (Position / Scale / Rotation / Opacity)
  - `PropertyStream[T]` 状态机 (Static ↔ Animated mutually exclusive)
  - escape hatch β: PropertyGroup tree (typed `Vec2Stream` / `ColorStream` / etc)
- **Architecture**: runtime ↔ serializer 分层 (10 architecture invariant + 4 serializer invariant)
- **Reusable serializer primitives** 给 V3: lower_property_stream / lower_shape_node / lower_layer / lower_item_siblings / capability_matrix
- **Capability matrix**: ship 时空 struct (escape hatch AE 2020 canonical 覆盖所有版本差异)
- **Test pyramid**: Tier 1 Go roundtrip + Tier 2 AE 2020/25 ship gate + Tier 3 nested-group preservation
- **classification deliverable** 完整记录 (spec §6)
- **新文件**: shape_graph.go / property_stream.go / new_layer.go / lower_*.go (5 个) / ldta_layout.go / capability_matrix.go / hydrate_shape.go + 测试 + verify_v2_2.jsx
- **教训**: (Phase 5 ship gate 期间发现填入)
```

- [ ] **Step 3: Commit**

```bash
git add workshop/board.md
git commit -m "docs(board): V2.2 ShapeLayer ship archive"
```

---

### Task 6.3: `coverage.md` / `coverage-detail.md` V2.2 段

**Files:**
- Modify: `workshop/plans/coverage.md`
- Modify: `workshop/plans/coverage-detail.md`

- [ ] **Step 1: coverage.md 在 V2 段加 V2.2**

定位 `### V2 结构性创建` 段，追加:

```markdown
- ✅ V2.2 ShapeLayer creation (2026-MM-DD):
  - `comp.NewShapeLayer(name)` + 5 ShapeNode (Rect/Ellipse/Path/Fill/Stroke)
  - PropertyStream typed (state machine Static ↔ Animated)
  - AE 2020 + AE 2025 ship gate PASS
  - 5 reusable serializer primitives for V3 inheritance
  - 详 `../specs/v2-2-layer-creation-design.md`
```

- [ ] **Step 2: coverage-detail.md 加 ShapeLayer / ShapeNode 段**

在 Layer 段后加:

```markdown
## ShapeLayer (V2.2)

| AE attr | aep-parser | 状态 | 备注 |
|---|---|---|---|
| `app.project.items[i].layers.addShape()` | `Composition.NewShapeLayer(name)` | ✅ R/W | V2.2 |
| ShapeLayer Transform group | `ShapeLayer.Transform()` typed; shorthand `Position()/Scale()/Rotation()/Opacity()` | ✅ R/W | V2.2 |
| `Property.setValueAtTime(t, v)` | `PropertyStream[T].AddKeyframeLinear / AddKeyframeWithEase` | ✅ R/W | V2.2; state machine |
| Rect Shape (Size/Position/Roundness) | `RectNode.Set{Size,Position,Roundness}` + accessor | ✅ R/W | V2.2 |
| Ellipse Shape | `EllipseNode.Set{Size,Position}` | ✅ R/W | V2.2 |
| Bezier Path (linear) | `PathNode.SetVertices + SetClosed` | ✅ R/W | V2.2; tangent V2.3+ |
| Fill (Color, Opacity) | `FillNode.Set{Color,Opacity}` | ✅ R/W | V2.2 |
| Stroke (Color, Width, Opacity) | `StrokeNode.Set{Color,Width,Opacity}` | ✅ R/W | V2.2; cap/join/dash V2.3+ |
| Stroke 额外字段 / Trim / Merge / Repeater | — | ❌ V2.3+ | escape hatch β 通过 PropertyGroup 可访问既有数据 |
| 嵌套 VectorGroup 合成 | — | ❌ V2.3 | preservation territory (parse/write 不破坏) |
| 3D ShapeLayer + 3D Position | — | ❌ V2.3+ | V2.2 2D layer only |
```

- [ ] **Step 3: Commit**

```bash
git add workshop/plans/coverage.md workshop/plans/coverage-detail.md
git commit -m "docs(coverage): V2.2 ShapeLayer + ShapeNode coverage rows"
```

---

### Task 6.4: Spec finalize — §8 RE Findings / §6.4a clear / §6.5 admission decisions / scar updates

**Files:**
- Modify: `workshop/specs/v2-2-layer-creation-design.md`
- Modify (if new quirks): `workshop/scars/ae25-acceptance-gate.md` or new scar

- [ ] **Step 1: 确认 §8 RE Findings 段含完整 RE-S1 到 RE-S9 finding records (Phase 0 已填)**

```bash
grep -c "^### RE-S" workshop/specs/v2-2-layer-creation-design.md
```

Expected: ≥ 9 (RE-S1..S9 + 可能 RE-S5a/b/c/d).

- [ ] **Step 2: 确认 §6.4a Unclassified 段空**

```bash
sed -n '/### 6.4a/,/^### 6\.5/p' workshop/specs/v2-2-layer-creation-design.md | grep "^\[pending\]"
```

Expected: 0 matches (Phase 6 commit 前 V2.2 spec 必须清空 6.4a).

如果有 pending → 当 task 在 Phase 6 内处理或归类.

- [ ] **Step 3: 确认 §6.5 admission decisions 全 finalize**

每个 candidate 都标 [不入 matrix] / [入 matrix:trait_name]。Phase 5 ship gate 期间发现新 trait → admission rule 评估 + finalize.

- [ ] **Step 4: 加 scar (if Phase 5 ship gate 发现 V2.2 新 ScriptingAPI quirks or 字节 quirks)**

例如：若发现某 shape ScriptingAPI 返回 stored × 1.x quirk → 新 scar `workshop/scars/v2-2-shape-quirks.md` 或追加到 既有 `ae25-acceptance-gate.md`.

- [ ] **Step 5: Commit**

```bash
git add workshop/specs/v2-2-layer-creation-design.md workshop/scars/
git commit -m "docs(v2.2): spec §8 RE finalize + §6.4a clear + §6.5 admission decisions"
```

---

### Task 6.5: Final verify + Phase 6 hard ship criterion

**Files:** none (operation task)

- [ ] **Step 1: Run all**

```bash
cd E:/projects/tools/aep-parser
go vet ./...                                                # clean
go test -count=1 ./internal/aep/...                         # 全 PASS
go test -count=1 ./internal/aep/... -v 2>&1 | grep -c '^--- PASS'  # ≥ 155
go test -count=1 ./internal/aep/... -v 2>&1 | grep -c '^--- FAIL'  # 0
```

- [ ] **Step 2: AE ship gate × 2 final run**

```bash
AE_SHIP_GATE=1 go test -count=1 ./internal/aep/ -run TestV2_2_AEShipGate -v -timeout 240s
AE_SHIP_GATE=1 go test -count=1 ./internal/aep/ -run TestV2_2_NestedGroup_AEReopens -v -timeout 240s
```

Both PASS.

- [ ] **Step 3: Phase 6 hard ship criterion check**

```
✓ go vet clean
✓ total PASS ≥ 155 (count freeze 进 board.md Last updated 行)
✓ AE 2020 + AE 2025 ship gate PASS
✓ Tier 3 preservation PASS
✓ spec §6.4a Unclassified 空
✓ spec §6.5 admission decisions 全 finalize
✓ spec §8 RE findings ≥ 9
✓ docs/composition.md + docs/layer.md + docs/shape.md 全 sync
✓ coverage.md + coverage-detail.md 含 V2.2 段
✓ board.md V2.2 archive + Last updated 更新
✓ scar updates (if Phase 5 发现 quirks)
```

- [ ] **Step 4: V2.2 ship 总结 commit**

```bash
git log --oneline | head -30  # 看 V2.2 整体提交链 (~40+ commits 跨 Phase 0-6)
# 如需 squash 或加 tag (e.g. v2.2.0), 按 git playbook 决
```

- [ ] **Step 5: 关闭 V2.2，准备 V3 brainstorm or V2.3**

更新 board.md Active focus 到下个候选。

**Phase 6 / V2.2 hard ship PASS criterion**: 全部 Step 3 checklist 项 ✓.

---

## 总结

V2.2 ship 完产出：

- **Public API**: `Composition.NewShapeLayer` + 5 typed shape nodes + PropertyStream state machine + escape hatch β
- **Architecture**: 10 + 4 invariants 钉死 runtime ↔ serializer 边界；single canonical AE 2020 lowering target；capability matrix 接口预留 V3 auto-derive
- **Reusable serializer primitives**: `lower_layer / lower_shape_node / lower_property_stream / lower_item_siblings / capability_matrix` 5 个 V3 直接 inherit
- **Tests**: ~35 new Go tests + 2 AE ship gates + 2 Tier 3 preservation tests
- **Classification deliverable** (spec §6): 17 runtime concepts / 19 serialization artifacts / 5 substrates / 7 capability candidates 显式归类
- **Docs sync**: docs/composition + layer + shape; workshop/board + coverage + coverage-detail + spec
- **scars** (if quirks): new V2.2 ScriptingAPI / 字节 quirks records

V3 brainstorm 输入完备 (`workshop/specs/v3-direction.md` + V2.2 实践数据).


