# V2.2.1 Ellipse embed-bytes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 `AddEllipse()` 写出的 Ellipse 被 AE 2020+2025 接受（不再 silent-drop），Size + Position 静态值经 embed+overwrite 正确持久化，双版本 ship-gate PASS。

**Architecture:** 镜像已 ship 的 Rect 路径——AE 存一份含 Ellipse 的 tolerance fixture → `extract_shape_bodies` 抽出 canonical LIST(tdgp) body → `//go:embed` 成 `.bin` → `lowerEllipseNode` clone embed body + `overwriteShapeStreamCdat` 原位覆盖 Size/Position cdat。byte 级 RE 已证明不可行（semantic silent-drop），embed+overwrite 是唯一路径。

**Tech Stack:** Go 1.x、`internal/rifx`（RIFX chunk 树）、ExtendScript JSX（AE 自动化）、`scripts/ae_run.ps1`（AE 无人值守 + modal 自动消除）。

参照 spec：`flightdeck/specs/2026-05-29-v2-2-1-ellipse-embed-design.md`。

---

## File Structure

- **Create** `tmp_debug/gen_shape_ellipse_tolerance.jsx` — AE 端造含 Ellipse 的 tolerance fixture。
- **Create** `test_data/v2_2_shape_ellipse_tolerance.aep` (+`.done`) — AE 产物（不入 git 视既有惯例；fixture 入 git 与否照 test_data 现状）。
- **Modify** `tmp_debug/extract_shape_bodies/main.go` — `extraction` 加 `srcPath` 字段 + Ellipse 入口。
- **Create** `internal/aep/templates/v2_2_shape_ellipse_body.bin` — 提取产物，embed 资源。
- **Modify** `internal/aep/lower_shape_node.go` — embed 声明 + `cloneShapeEllipseBody()` + 重写 `lowerEllipseNode`（:188-200）。
- **Modify** `internal/aep/lower_shape_node_test.go` — Ellipse cdat 覆盖单测（替换占位的 `TestLowerEllipseNode_DispatcherWorks`）。
- **Create** `test_data/verify_v2_2_ellipse.jsx` — 聚焦 Ellipse ship-gate 的 AE 端 assert 脚本。
- **Modify** `internal/aep/shape_layer_shipgate_test.go` — 加聚焦 Ellipse ship-gate 测试（不 skip）；原 3 层 canonical 注释更新。
- **Modify** `docs/shape.md`、`flightdeck/flight-plans/coverage.md`、`flightdeck/cockpit.md`、`flightdeck/manifest.md` — 收口。

---

## Task 1: AE 端造 Ellipse tolerance fixture

**Files:**
- Create: `tmp_debug/gen_shape_ellipse_tolerance.jsx`
- Produces: `test_data/v2_2_shape_ellipse_tolerance.aep` + `.done`

- [ ] **Step 1: 写 JSX（照抄 gen_shape_tolerance.jsx 骨架，换成 Ellipse）**

```javascript
// tmp_debug/gen_shape_ellipse_tolerance.jsx
//
// V2.2.1 Ellipse embed-bytes — produce v2_2_shape_ellipse_tolerance.aep.
// 1 ShapeLayer with a nested VectorGroup containing 1 Ellipse + 1 Fill.
// Ellipse Size AND Position set to non-default values so AE persists both
// cdat slots (default-valued slots get elided → nothing to overwrite later).
// Source for extract_shape_bodies → templates/v2_2_shape_ellipse_body.bin.
//
// Run (AE 2025):
//   "E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe" -r \
//     "E:/projects/tools/aep-parser/tmp_debug/gen_shape_ellipse_tolerance.jsx"
// Wait for test_data/v2_2_shape_ellipse_tolerance.done — first line PASS/FAIL.
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_shape_ellipse_tolerance.aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_shape_ellipse_tolerance.done");
    var log = [];
    var ok = false;
    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        var c = app.project.items.addComp("EllipseTolerance", 1920, 1080, 1, 5, 30);
        var s = c.layers.addShape();
        s.name = "EllipseNested";
        var root = s.property("ADBE Root Vectors Group");
        var sub = root.addProperty("ADBE Vector Group");
        var subContents = sub.property("ADBE Vectors Group");
        var ell = subContents.addProperty("ADBE Vector Shape - Ellipse");
        ell.property("ADBE Vector Ellipse Size").setValue([200, 100]);
        ell.property("ADBE Vector Ellipse Position").setValue([50, 30]);
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
    try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e3) {}
    try { app.quit(); } catch (e4) {}
})();
```

- [ ] **Step 2: 经 ae_run.ps1 自助跑 AE（无人值守）**

Run（PowerShell；参照 `flightdeck/checklists/re-fixture.md` § GDI 自动化 / `scripts/ae_run.ps1` 的实际签名，若签名不同以脚本为准）:
```
pwsh scripts/ae_run.ps1 -AeExe "E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe" -Jsx "E:/projects/tools/aep-parser/tmp_debug/gen_shape_ellipse_tolerance.jsx" -Done "E:/projects/tools/aep-parser/test_data/v2_2_shape_ellipse_tolerance.done"
```
Expected: 进程 exit 0，`test_data/v2_2_shape_ellipse_tolerance.done` 首行 `PASS`，`.aep` 已生成。
冷启 flake（splash/About 撞屏 → exit 2）：先跑一个已知-good fixture 热身再重试（已知问题，非数据）。

- [ ] **Step 3: 确认产物**

Run: `git status --porcelain test_data/v2_2_shape_ellipse_tolerance.aep` + 用 dump 工具核对层存在。
Expected: `.aep` 存在；首行 PASS。

- [ ] **Step 4: Commit**

```bash
git add tmp_debug/gen_shape_ellipse_tolerance.jsx test_data/v2_2_shape_ellipse_tolerance.aep test_data/v2_2_shape_ellipse_tolerance.done
git commit -m "test(aep): V2.2.1 Ellipse tolerance fixture (JSX + AE-saved aep)"
```

---

## Task 2: 提取 Ellipse body bytes（含 Position-elide 风险闸门）

**Files:**
- Modify: `tmp_debug/extract_shape_bodies/main.go`
- Produces: `internal/aep/templates/v2_2_shape_ellipse_body.bin`

- [ ] **Step 1: 给 extraction 加 srcPath 字段 + Ellipse 入口**

把 `main.go` 的 `extraction` struct 与 `extractions` 表改成每条带源 fixture（保留 Rect/Fill 原读 tolerance.aep 不变以防回归）：

```go
type extraction struct {
	srcPath   string
	matchName string
	outPath   string
}

var extractions = []extraction{
	{"test_data/v2_2_shape_tolerance.aep", "ADBE Vector Shape - Rect", "internal/aep/templates/v2_2_shape_rect_body.bin"},
	{"test_data/v2_2_shape_tolerance.aep", "ADBE Vector Graphic - Fill", "internal/aep/templates/v2_2_shape_fill_body.bin"},
	{"test_data/v2_2_shape_ellipse_tolerance.aep", "ADBE Vector Shape - Ellipse", "internal/aep/templates/v2_2_shape_ellipse_body.bin"},
}
```

并把 `main()` 里写死的 `os.Open("test_data/v2_2_shape_tolerance.aep")` + 单次 `rifx.Parse` 改为按 `ex.srcPath` 逐条打开解析：

```go
func main() {
	for _, ex := range extractions {
		f, err := os.Open(ex.srcPath)
		if err != nil {
			panic(err)
		}
		root, err := rifx.Parse(f)
		f.Close()
		if err != nil {
			panic(err)
		}
		body := findShapeBody(root, ex.matchName)
		if body == nil {
			fmt.Printf("MISS: %s not found in %s\n", ex.matchName, ex.srcPath)
			continue
		}
		out, err := os.Create(ex.outPath)
		if err != nil {
			panic(err)
		}
		if err := body.Write(out); err != nil {
			panic(err)
		}
		stat, _ := out.Stat()
		out.Close()
		fmt.Printf("wrote %s (%d bytes, %d children)\n", ex.outPath, stat.Size(), len(body.Children))
	}
}
```

- [ ] **Step 2: 跑提取**

Run: `go run ./tmp_debug/extract_shape_bodies`
Expected: 三行 `wrote ...`，含 `wrote internal/aep/templates/v2_2_shape_ellipse_body.bin (N bytes, M children)`，**无 MISS**。

- [ ] **Step 3: 风险闸门 — 确认 body 含 Size + Position cdat**

Run（用既有 dump 工具，如 `go run ./tmp_debug/dump_chunks <bin>`，或临时 inline 检查）核对 `v2_2_shape_ellipse_body.bin` 内含 `ADBE Vector Ellipse Size` 与 `ADBE Vector Ellipse Position` 两个 tdmn，且各自后随的 tdbs LIST 内有非空 cdat。
Expected: 两个 stream 都在，都有 cdat。

**决策点**：
- 两个都在 → 继续 Task 3 做 Size+Position。
- **若 Position 被 AE elide（缺 cdat）** → 触发 spec §4 fallback：Task 3 只覆盖 Size，Ellipse Position 转 runtime-only（与 Rect 一致）；在 godoc/docs 标注，ship-gate 只验 Size。后续 Task 同步删去 Position 相关断言。

- [ ] **Step 4: Commit**

```bash
git add tmp_debug/extract_shape_bodies/main.go internal/aep/templates/v2_2_shape_ellipse_body.bin
git commit -m "build(aep): extract Ellipse body bytes from tolerance fixture (extract tool srcPath)"
```

---

## Task 3: lowerEllipseNode 改 embed+overwrite（TDD）

**Files:**
- Modify: `internal/aep/lower_shape_node.go:188-200` + embed/clone 声明区
- Test: `internal/aep/lower_shape_node_test.go`（替换 `TestLowerEllipseNode_DispatcherWorks`）

- [ ] **Step 1: 写失败测试 —— 断言 Size+Position cdat 被覆盖为输入值**

替换 `lower_shape_node_test.go` 中的 `TestLowerEllipseNode_DispatcherWorks`（:31-40）为：

```go
func TestLowerEllipseNode_OverwritesSizeAndPosition(t *testing.T) {
	e := aep.NewEllipseNode()
	_ = e.SetSize([2]float64{321, 123})
	_ = e.SetPosition([2]float64{40, 60})
	chunk, err := aep.LowerShapeNodeForTest(e)
	if err != nil {
		t.Fatal(err)
	}
	if chunk == nil || !chunk.IsList() || chunk.FormType != rifx.IDTdgp {
		t.Fatalf("expected LIST(tdgp), got %+v", chunk)
	}
	assertEllipseStreamCdat(t, chunk, "ADBE Vector Ellipse Size", []float64{321, 123})
	assertEllipseStreamCdat(t, chunk, "ADBE Vector Ellipse Position", []float64{40, 60})
}

// assertEllipseStreamCdat finds the tdmn matching name inside body, descends
// to its tdbs cdat, and asserts the leading f64 BE values equal want.
func assertEllipseStreamCdat(t *testing.T, body *rifx.Chunk, name string, want []float64) {
	t.Helper()
	kids := body.Children
	for i := 0; i+1 < len(kids); i++ {
		if kids[i].ID == rifx.IDTdmn && trimTestNUL(string(kids[i].Data)) == name {
			tdbs := kids[i+1]
			if !tdbs.IsList() || tdbs.FormType != rifx.IDTdbs {
				t.Fatalf("%s: next chunk not LIST(tdbs)", name)
			}
			for _, ch := range tdbs.Children {
				if ch.ID == rifx.IDCdat {
					for j, w := range want {
						got := math.Float64frombits(binary.BigEndian.Uint64(ch.Data[j*8 : j*8+8]))
						if got != w {
							t.Errorf("%s[%d] = %v, want %v", name, j, got, w)
						}
					}
					return
				}
			}
			t.Fatalf("%s: no cdat under tdbs", name)
		}
	}
	t.Fatalf("%s: tdmn not found in lowered body", name)
}

func trimTestNUL(s string) string {
	if n := indexNUL(s); n >= 0 {
		return s[:n]
	}
	return s
}
```

加 import：`"encoding/binary"`、`"math"`（在 `lower_shape_node_test.go` import 块）。

> 注：若 Task 2 决策点判定 Position 被 elide，则删去 Position 那行断言及 `SetPosition` 调用。

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./internal/aep/ -run TestLowerEllipseNode_OverwritesSizeAndPosition -v`
Expected: FAIL —— 当前 from-scratch `lowerEllipseNode` 发的 cdat 值不等于 321/123（或结构不符）。

- [ ] **Step 3: 加 embed + clone（照抄 Rect）**

在 `lower_shape_node.go` embed 区（:28-32 附近）加：
```go
//go:embed templates/v2_2_shape_ellipse_body.bin
var v22ShapeEllipseBodyBytes []byte
```
在 `var (...)` once/cache 区（:34-42）加：
```go
	v22ShapeEllipseOnce  sync.Once
	v22ShapeEllipseCache *rifx.Chunk
	v22ShapeEllipseErr   error
```
加 clone 函数（仿 `cloneShapeRectBody`）：
```go
func cloneShapeEllipseBody() (*rifx.Chunk, error) {
	v22ShapeEllipseOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(v22ShapeEllipseBodyBytes))
		if err != nil {
			v22ShapeEllipseErr = fmt.Errorf("parse v22ShapeEllipseBodyBytes: %w", err)
			return
		}
		v22ShapeEllipseCache = ch
	})
	if v22ShapeEllipseErr != nil {
		return nil, v22ShapeEllipseErr
	}
	return cloneChunk(v22ShapeEllipseCache), nil
}
```

- [ ] **Step 4: 重写 lowerEllipseNode（:188-200）为 embed+overwrite**

```go
// lowerEllipseNode emits an Ellipse shape body using V2.2.1 embedded tolerance
// bytes (templates/v2_2_shape_ellipse_body.bin). Same rationale as
// lowerRectNode — from-scratch emit triggered AE silent-drop; embedded
// canonical body + cdat overwrite for Size/Position is the validator-safe path.
//
// V2.2.1 limitations: Direction stays AE default; animated Size/Position use
// the first keyframe value as a static fallback.
func lowerEllipseNode(e *EllipseNode, _ *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapeEllipseBody()
	if err != nil {
		return nil, err
	}
	sz := e.size.static
	if e.size.mode == StreamModeAnimated && len(e.size.keyframes) > 0 {
		sz = e.size.keyframes[0].Value
	}
	overwriteShapeStreamCdat(body, "ADBE Vector Ellipse Size", encodeF64sBE(sz[0], sz[1]))
	ps := e.position.static
	if e.position.mode == StreamModeAnimated && len(e.position.keyframes) > 0 {
		ps = e.position.keyframes[0].Value
	}
	overwriteShapeStreamCdat(body, "ADBE Vector Ellipse Position", encodeF64sBE(ps[0], ps[1]))
	return body, nil
}
```

> 确认 `EllipseNode` 的字段名是 `size` / `position`（见 shape_graph.go:147-177）；若实际字段名不同，以源码为准。若 Task 2 判定 Position elide，删去 Position 段。

- [ ] **Step 5: 跑测试确认通过 + 全包回归**

Run: `go test ./internal/aep/ -run TestLowerEllipseNode_OverwritesSizeAndPosition -v`
Expected: PASS。
Run: `go vet ./... && go test -count=1 ./internal/aep/...`
Expected: 全绿（含既有 Rect/Fill/VectorGroup 测试 + preservation 测试无回归）。

- [ ] **Step 6: Commit**

```bash
git add internal/aep/lower_shape_node.go internal/aep/lower_shape_node_test.go
git commit -m "feat(aep): Ellipse embed+overwrite (Size+Position persist) — V2.2.1 (Alpha)"
```

---

## Task 4: 聚焦 Ellipse 双版本 ship-gate

**Files:**
- Create: `test_data/verify_v2_2_ellipse.jsx`
- Modify: `internal/aep/shape_layer_shipgate_test.go`

- [ ] **Step 1: 写 AE 端 assert JSX**

`test_data/verify_v2_2_ellipse.jsx` —— 读 args.json（`{input, done}`），open input.aep，assert：comp 存在、含名为 `Ellipse_Static` 的 ShapeLayer 在 `comp.layers`、该层 Contents 下可达 `ADBE Vector Shape - Ellipse`、其 Size==[200,100] 且 Position==[50,30]；写 `.done` 首行 PASS/FAIL。结构照抄现有 `verify_v2_2.jsx` 的 open/assert/done 骨架（assert-based，无 byte 基线）。

```javascript
// test_data/verify_v2_2_ellipse.jsx — V2.2.1 focused Ellipse ship-gate.
(function () {
    var args = JSON.parse(File("e:/projects/tools/aep-parser/test_data/v2_2_ellipse_args.json").read ?
        readFile("e:/projects/tools/aep-parser/test_data/v2_2_ellipse_args.json") : "{}");
    // (实现以现有 verify_v2_2.jsx 的 args 读取/done 写出方式为准——见该文件)
    var doneFile, ok = false, log = [];
    try {
        app.open(File(args.input));
        var p = app.project, comp = null;
        for (var i = 1; i <= p.numItems; i++) if (p.item(i) instanceof CompItem) { comp = p.item(i); break; }
        if (!comp) throw "no comp";
        var layer = null;
        for (var j = 1; j <= comp.numLayers; j++) if (comp.layer(j).name === "Ellipse_Static") { layer = comp.layer(j); break; }
        if (!layer) throw "Ellipse_Static layer dropped";
        var ell = findProp(layer.property("ADBE Root Vectors Group"), "ADBE Vector Shape - Ellipse");
        if (!ell) throw "Ellipse shape dropped";
        var sz = ell.property("ADBE Vector Ellipse Size").value;
        var ps = ell.property("ADBE Vector Ellipse Position").value;
        if (Math.abs(sz[0]-200) > 0.5 || Math.abs(sz[1]-100) > 0.5) throw "Size=" + sz;
        if (Math.abs(ps[0]-50) > 0.5 || Math.abs(ps[1]-30) > 0.5) throw "Position=" + ps;
        ok = true; log.push("ellipse OK");
    } catch (e) { log.push("ERROR: " + e.toString()); }
    writeDone(args.done, ok, log);
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); app.quit(); } catch (e2) {}
    function findProp(grp, mn) { /* recursive matchName search — copy from verify_v2_2.jsx */ }
    function writeDone(path, ok, log) { var f = File(path); f.open("w"); f.write((ok?"PASS\n":"FAIL\n")+log.join("\n")); f.close(); }
})();
```
> 以现有 `test_data/verify_v2_2.jsx` 的 args 解析 / `findProp` / `writeDone` 实现为准复用，避免重写易错的 ExtendScript 细节。

- [ ] **Step 2: 写聚焦 ship-gate Go 测试（不 skip）**

在 `shape_layer_shipgate_test.go` 加（仿 `runV2_2ShipGate`，但单层 Ellipse+Fill、用新 verify JSX + 新 args 路径）：

```go
func runV2_2EllipseShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "v2_2_ellipse.aep")
	doneFile := filepath.Join(tempDir, "v2_2_ellipse.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/v2_2_ellipse_args.json`

	p := aep.NewProject(target)
	comp, err := p.NewComposition("Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, _ := comp.NewShapeLayer("Ellipse_Static")
	ell, _ := l.RootGroup().AddEllipse()
	_ = ell.SetSize([2]float64{200, 100})
	_ = ell.SetPosition([2]float64{50, 30})
	fill, _ := l.RootGroup().AddFill()
	_ = fill.SetColor([4]float64{0.5, 0.5, 0.5, 1})

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q}`, toFwd(inputAEP), toFwd(doneFile))
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	jsxPath := `E:/projects/tools/aep-parser/test_data/verify_v2_2_ellipse.jsx`
	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.SplitN(string(content), "\n", 2)
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("ellipse ship gate FAIL:\n%s", string(content))
	}
}

func TestV2_2_Ellipse_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2EllipseShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_Ellipse_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2EllipseShipGate(t, aep.TargetAE2020, aeExe)
}
```

并把原 3 层 `TestV2_2_AEShipGate_AE2020/_AE2025` 的 skip 注释更新为：`"Path/Stroke embed bytes 仍 deferred（Ellipse 已单独 ship,见 TestV2_2_Ellipse_AEShipGate_*）"`。

- [ ] **Step 3: 编译确认（无 AE 时 skip-clean）**

Run: `go vet ./... && go test ./internal/aep/ -run TestV2_2_Ellipse_AEShipGate -v`
Expected: 不设 `AE_SHIP_GATE` 时两个测试 SKIP（编译通过、逻辑无误）。

- [ ] **Step 4: 跑双版本 ship-gate（AE 自助）**

Run:
```
$env:AE_SHIP_GATE=1; go test ./internal/aep/ -run TestV2_2_Ellipse_AEShipGate_AE2025 -v
$env:AE_SHIP_GATE=1; go test ./internal/aep/ -run TestV2_2_Ellipse_AEShipGate_AE2020 -v
```
Expected: 两者 PASS（`.done` 首行 PASS）。冷启 flake → warm retry。
若 FAIL 且报 Ellipse dropped 或 Position 不符 → 回 spec §4 fallback（降 Size-only，重跑）。

- [ ] **Step 5: Commit**

```bash
git add test_data/verify_v2_2_ellipse.jsx internal/aep/shape_layer_shipgate_test.go
git commit -m "test(aep): focused Ellipse double-version ship-gate — AE 2020+2025 PASS"
```

---

## Task 5: 文档收口 + flightdeck 更新

**Files:**
- Modify: `internal/aep/lower_shape_node.go`（godoc 已在 Task 3 step 4 改）
- Modify: `docs/shape.md`、`flightdeck/flight-plans/coverage.md`、`flightdeck/cockpit.md`、`flightdeck/manifest.md`

- [ ] **Step 1: docs/shape.md** —— 把 Ellipse 从 "V2.2.1 deferred" 表移到已支持；标注：Size+Position 静态持久化、Direction=AE 默认、动画走 first-kf static fallback。

- [ ] **Step 2: coverage.md** —— V2.2 ShapeLayer 段更新 Ellipse 行（从 "silent drop" → "✅ Size+Position embed"）。

- [ ] **Step 3: cockpit.md** —— Active focus 改为 "V2.2.1 Ellipse embed-bytes 已 ship（双版本 ship-gate PASS）；下一子项 Path/Stroke/Fill-Color/keyframe 待排"；Next session 列剩余子项；自验留痕记 ship-gate done 文件。

- [ ] **Step 4: manifest.md** —— 若分支待合并，记一行 awaiting-review；合并后清。

- [ ] **Step 5: 最终全绿复核**

Run: `go vet ./... && go test -count=1 ./internal/aep/...`
Expected: 全绿。

- [ ] **Step 6: Commit**

```bash
git add docs/shape.md flightdeck/
git commit -m "docs(aep): V2.2.1 Ellipse embed-bytes ship — shape.md + coverage + cockpit"
```

---

## Self-Review

**Spec coverage（对 design §1-§5）**：
- §1 Ellipse 被接受 + Size/Position 持久化 → Task 3+4 ✓
- §3.1 fixture → Task 1 ✓；§3.2 提取 → Task 2 ✓；§3.3 lower → Task 3 ✓；§3.4 ship-gate → Task 4 ✓；§3.5 收口 → Task 5 ✓
- §4 Position-elide fallback → Task 2 step 3 决策点 + Task 3/4 的 fallback 注记 ✓；cold-start flake → Task 1 step 2 / Task 4 step 4 ✓
- §5 DoD 1-6 → Task 3 step 5（vet+test）、Task 2（bin）、Task 4（ship-gate）、Task 5（docs/cockpit）✓

**Placeholder scan**：JSX 中 `findProp`/`writeDone`/args 读取标注"复用现有 verify_v2_2.jsx 实现"——这是有意复用既有正确实现而非占位（重写 ExtendScript 易错）；执行时直接 copy 现成函数体。其余步骤均含完整代码/命令/期望输出。

**Type consistency**：`cloneShapeEllipseBody` / `v22ShapeEllipseBodyBytes` / `overwriteShapeStreamCdat` / `encodeF64sBE` / `EllipseNode.size/.position` / `SetSize/SetPosition/AddEllipse/RootGroup/NewShapeLayer/NewComposition/AddFill/SetColor` 均与源码现状一致；ship-gate helper `runAeRunShipGate` 沿用现有签名。
