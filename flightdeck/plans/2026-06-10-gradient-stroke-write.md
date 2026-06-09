---
status: active
summary: GradientStroke R/W 实现：提取 G-Stroke 模板 → scene(enum+node+wiring) → 抽共享 lower/hydrate helper(fill 重构) → GradientStroke lower+reader → facade → 单测 → 双版本 ship-gate(color+alpha) → aepdemo+coverage。对齐 GradientFill，8 task TDD
last_updated: 2026-06-10
implements: specs/2026-06-10-gradient-stroke-write.md
---

# GradientStroke R/W 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 给 shape layer 加 gradient stroke 的读写（`ADBE Vector Graphic - G-Stroke`），只 model gradient color/alpha stops，对称已 ship 的 GradientFill。

**Architecture:** reader 补 G-Stroke 节点解 gradient（局部降级）；writer 加 `GradientStrokeNode`，克隆从 fixture 提取的 G-Stroke 模板 + 覆写 prop.map XML。fill 的 lower/hydrate/setter 本体抽成共享 helper、stroke 复用（第二个 gradient 节点 = DRY 阈值到）。双版本 AE ship-gate（color + alpha）。

**Tech Stack:** Go；`internal/{scene,serializer,codec,rifx,aep}` 多包；`//go:embed` 模板；AE 2020/2025 ship-gate via `scripts/ae_run.ps1` + JSX。

**设计源：** `flightdeck/specs/2026-06-10-gradient-stroke-write.md`（两轮 review 定稿）。

---

### Task 1: 提取 G-Stroke 模板字节

**Files:**
- Create: `internal/serializer/templates/v2_2_shape_gradstroke_body.bin`
- Reference: `test_data/v2_2_gradient_src.aep`（已 dump 实证含 G-Stroke 节点 + GCst）

提取工具参考既有 `tmp_debug/extract_shape_bodies`（它已为 gradfill 提取过 `v2_2_shape_gradfill_body.bin`）。

- [ ] **Step 1: 看 extract_shape_bodies 怎么提取 gradfill 的**

Run: `go run ./tmp_debug/dump_chunks/main.go test_data/v2_2_gradient_src.aep | grep -n -B2 -A6 "G-Stroke"`
Expected: 看到 `tdmn = ADBE Vector Graphic - G-Stroke` 后跟 `[LIST tdgp]`（节点 body），body 内含 `ADBE Vector Grad Colors → GCst → GCky → Utf8`。记下 G-Stroke 节点 body（紧跟该 tdmn 的 LIST tdgp）。

- [ ] **Step 2: 提取 G-Stroke body 到 .bin**

扩展 `tmp_debug/extract_shape_bodies/main.go`：定位 `tdmn == "ADBE Vector Graphic - G-Stroke"` 后的 LIST(tdgp) payload，`rifx.WriteChunk` 写到 `internal/serializer/templates/v2_2_shape_gradstroke_body.bin`（与 gradfill 提取逻辑一致，只改 match-name 与输出路径）。

Run: `go run ./tmp_debug/extract_shape_bodies/main.go`（按其现有用法）
Expected: 生成 `internal/serializer/templates/v2_2_shape_gradstroke_body.bin`

- [ ] **Step 3: 验证提取的 body 可被 rifx 解析回**

Run: `go run ./tmp_debug/dump_chunks/main.go internal/serializer/templates/v2_2_shape_gradstroke_body.bin`
Expected: 输出含 `ADBE Vector Grad Colors`、`[LIST GCst]`、`[LIST GCky]`、`Utf8`。若缺 GCst → no-go，停止并按 spec §模板提取兜底处理。

- [ ] **Step 4: Commit**

```bash
git add internal/serializer/templates/v2_2_shape_gradstroke_body.bin tmp_debug/extract_shape_bodies/main.go
git commit -m "feat(gradstroke): extract G-Stroke body template from fixture"
```

---

### Task 2: scene — enum + GradientStrokeNode + 共享 setter（fill 重构）+ AddGradientStroke

**Files:**
- Modify: `internal/scene/scene_shape_graph.go`（enum line 23 后；setter line 417-457 抽 free function；AddGradientFill line 97 后加 AddGradientStroke）
- Modify: `internal/scene/scene_wiring.go`（line 102 后加 setter）

- [ ] **Step 1: 写失败测试（scene 单测）**

Create/append `internal/scene/scene_gradient_stroke_test.go`:

```go
package scene

import "testing"

func TestGradientStrokeNode_Defaults(t *testing.T) {
	n := NewGradientStrokeNode()
	if n.Kind() != ShapeKindGradientStroke {
		t.Fatalf("Kind = %v, want ShapeKindGradientStroke", n.Kind())
	}
	if got := len(n.Gradient().ColorStops); got != 2 {
		t.Fatalf("default color stops = %d, want 2", got)
	}
}

func TestGradientStrokeNode_SetColorStops_Validation(t *testing.T) {
	n := NewGradientStrokeNode()
	if err := n.SetColorStops([]GradientColorStop{{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0, 0}}}); err == nil {
		t.Fatal("want error for <2 stops, got nil")
	}
	ok := []GradientColorStop{
		{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0, 0}},
		{Offset: 1, Midpoint: 0.5, Color: [3]float64{0, 0, 1}},
	}
	if err := n.SetColorStops(ok); err != nil {
		t.Fatalf("SetColorStops valid: %v", err)
	}
	if n.Gradient().ColorStops[1].Color != [3]float64{0, 0, 1} {
		t.Fatal("color stop not applied")
	}
}

func TestAddGradientStroke(t *testing.T) {
	g := NewVectorGroup()
	n, err := g.AddGradientStroke()
	if err != nil {
		t.Fatal(err)
	}
	if g.Children[len(g.Children)-1] != n {
		t.Fatal("AddGradientStroke did not append node")
	}
}
```

- [ ] **Step 2: 运行确认失败**

Run: `go test ./internal/scene/ -run TestGradientStroke -v`
Expected: FAIL（编译错误：`NewGradientStrokeNode` / `ShapeKindGradientStroke` / `AddGradientStroke` undefined）

- [ ] **Step 3: 加 enum 值**

`internal/scene/scene_shape_graph.go`，在 `ShapeKindGradientFill` 行后：

```go
	ShapeKindGradientFill                      // `ADBE Vector Graphic - G-Fill`
	ShapeKindGradientStroke                    // `ADBE Vector Graphic - G-Stroke`
```

（删掉原 `// V2.3+ candidates: ... GradientStroke ...` 注释里的 GradientStroke 一项，因已实现。）

- [ ] **Step 4: 抽共享 setter free function + 重构 fill + 加 GradientStrokeNode**

`internal/scene/scene_shape_graph.go`。先把 `GradientFillNode.SetColorStops` / `SetAlphaStops` 的本体抽成 free function（fill 改调它）：

```go
// setGradientColorStops validates (≥2 stops; offset/midpoint/color ∈ [0,1])
// and replaces g.ColorStops. Shared by GradientFill / GradientStroke.
func setGradientColorStops(g *codec.Gradient, stops []GradientColorStop) error {
	if len(stops) < 2 {
		return fmt.Errorf("SetColorStops: need ≥ 2 stops, got %d", len(stops))
	}
	for i, s := range stops {
		if err := checkUnit("offset", i, s.Offset); err != nil {
			return err
		}
		if err := checkUnit("midpoint", i, s.Midpoint); err != nil {
			return err
		}
		for c, v := range s.Color {
			if v < 0 || v > 1 {
				return fmt.Errorf("SetColorStops: stop %d color[%d] = %g out of range [0,1]", i, c, v)
			}
		}
	}
	g.ColorStops = append([]GradientColorStop(nil), stops...)
	return nil
}

// setGradientAlphaStops validates (≥2 stops; offset/midpoint/alpha ∈ [0,1])
// and replaces g.AlphaStops. Shared by GradientFill / GradientStroke.
func setGradientAlphaStops(g *codec.Gradient, stops []GradientAlphaStop) error {
	if len(stops) < 2 {
		return fmt.Errorf("SetAlphaStops: need ≥ 2 stops, got %d", len(stops))
	}
	for i, s := range stops {
		if err := checkUnit("offset", i, s.Offset); err != nil {
			return err
		}
		if err := checkUnit("midpoint", i, s.Midpoint); err != nil {
			return err
		}
		if s.Alpha < 0 || s.Alpha > 1 {
			return fmt.Errorf("SetAlphaStops: stop %d alpha = %g out of range [0,1]", i, s.Alpha)
		}
	}
	g.AlphaStops = append([]GradientAlphaStop(nil), stops...)
	return nil
}
```

把 `GradientFillNode.SetColorStops` / `SetAlphaStops` 改成：

```go
func (n *GradientFillNode) SetColorStops(stops []GradientColorStop) error {
	return setGradientColorStops(n.gradient, stops)
}
func (n *GradientFillNode) SetAlphaStops(stops []GradientAlphaStop) error {
	return setGradientAlphaStops(n.gradient, stops)
}
```

加 `GradientStrokeNode`（紧邻 GradientFillNode 之后）：

```go
// GradientStrokeNode — `ADBE Vector Graphic - G-Stroke`. Models the gradient's
// color + alpha stops only, symmetric to GradientFillNode. Stroke geometry
// (width / cap / join / dashes / taper / wave) + ramp geometry are NOT modeled
// (kept at the extracted template's values; deferred). See the design spec.
type GradientStrokeNode struct {
	gradient *codec.Gradient
}

// NewGradientStrokeNode constructs a default 2-stop black→white gradient stroke.
func NewGradientStrokeNode() *GradientStrokeNode {
	return &GradientStrokeNode{gradient: defaultGradient()}
}

func (n *GradientStrokeNode) Kind() ShapeNodeKind { return ShapeKindGradientStroke }

// Gradient returns the live gradient (color + alpha stops).
func (n *GradientStrokeNode) Gradient() *Gradient { return n.gradient }

// SetColorStops replaces the gradient's color stops (≥2; ranges in [0,1]).
func (n *GradientStrokeNode) SetColorStops(stops []GradientColorStop) error {
	return setGradientColorStops(n.gradient, stops)
}

// SetAlphaStops replaces the gradient's alpha stops (≥2; ranges in [0,1]).
func (n *GradientStrokeNode) SetAlphaStops(stops []GradientAlphaStop) error {
	return setGradientAlphaStops(n.gradient, stops)
}

// Properties returns the escape-hatch β view.
func (n *GradientStrokeNode) Properties() *PropertyGroup {
	return &PropertyGroup{Name: "Gradient Stroke"}
}

// AddGradientStroke appends a default-valued GradientStrokeNode (2-stop
// black→white gradient) and returns it. Set stops via SetColorStops /
// SetAlphaStops.
func (g *VectorGroup) AddGradientStroke() (*GradientStrokeNode, error) {
	n := NewGradientStrokeNode()
	g.Children = append(g.Children, n)
	return n, nil
}
```

- [ ] **Step 5: 加 reader 用的 wiring setter**

`internal/scene/scene_wiring.go`，line 102 后：

```go
func SetGradientStrokeNodeGradient(n *GradientStrokeNode, g *Gradient) { n.gradient = g }
```

- [ ] **Step 6: 立即全仓检查 ShapeNodeKind switch（spec 要求，紧跟 enum 改动）**

Run: `grep -rn "case \*GradientFillNode\|ShapeKindGradientFill\|case ShapeKind" internal/`
Expected: 找出所有按 ShapeNodeKind / 节点类型分派的 switch。逐个确认是否需要 GradientStroke case（lower switch 在 Task 3 加；reader switch 在 Task 4 加；此处只登记，别遗漏）。

- [ ] **Step 7: 运行测试通过**

Run: `go test ./internal/scene/ -run TestGradientStroke -v && go test ./internal/scene/ -v`
Expected: PASS（新测试 + 既有 fill 测试仍绿，证明 setter 重构无回归）

- [ ] **Step 8: Commit**

```bash
git add internal/scene/scene_shape_graph.go internal/scene/scene_wiring.go internal/scene/scene_gradient_stroke_test.go
git commit -m "feat(gradstroke): scene GradientStrokeNode + shared gradient setters"
```

---

### Task 3: serializer/lower — 抽 lowerGradientStops（fill 重构）+ GradientStroke lower

**Files:**
- Modify: `internal/serializer/lower_shape_node.go`（embed line 46 后；once vars line 49 区；shapeMatchNames line 235；lowerShapeNode switch line 257；lowerGradientFillNode line 717）

- [ ] **Step 1: 写失败的 lower 单测**

Create/append `internal/serializer/lower_gradient_stroke_test.go`:

```go
package serializer

import (
	"testing"

	"github.com/example/aep-parser/internal/scene"
)

func TestLowerGradientStrokeNode_HasGradColors(t *testing.T) {
	n := scene.NewGradientStrokeNode()
	if err := n.SetColorStops([]scene.GradientColorStop{
		{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0, 0}},
		{Offset: 0.5, Midpoint: 0.5, Color: [3]float64{0, 1, 0}},
		{Offset: 1, Midpoint: 0.5, Color: [3]float64{0, 0, 1}},
	}); err != nil {
		t.Fatal(err)
	}
	body, err := LowerShapeNodeForTest(n)
	if err != nil {
		t.Fatalf("lower: %v", err)
	}
	// The lowered body must carry the Grad Colors XML with the 3 custom stops.
	xml := findGradientStopsXML(body, "ADBE Vector Grad Colors")
	if xml == "" {
		t.Fatal("lowered G-Stroke body has no Grad Colors XML")
	}
	g := codecParse(t, xml)
	if len(g.ColorStops) != 3 {
		t.Fatalf("lowered stops = %d, want 3", len(g.ColorStops))
	}
}
```

（`codecParse` 助手：`func codecParse(t *testing.T, xml string) *codec.Gradient { g := codec.ParseGradientXML(xml); if g == nil { t.Fatal("parse nil") }; return g }`，import codec；若已有等价助手则复用。）

- [ ] **Step 2: 运行确认失败**

Run: `go test ./internal/serializer/ -run TestLowerGradientStroke -v`
Expected: FAIL（`lowerShapeNode: unsupported kind` 或编译错误 `cloneShapeGradStrokeBody` undefined）

- [ ] **Step 3: 加 embed + once vars**

`internal/serializer/lower_shape_node.go`，line 46 后加 embed：

```go
//go:embed templates/v2_2_shape_gradstroke_body.bin
var v22ShapeGradStrokeBodyBytes []byte
```

once vars 区（line 49 的 var 块内）加：

```go
	v22ShapeGradStrokeOnce  sync.Once
	v22ShapeGradStrokeCache *rifx.Chunk
	v22ShapeGradStrokeErr   error
```

- [ ] **Step 4: 抽共享 lowerGradientStops + 重构 fill + 加 clone/lower for stroke**

把 `lowerGradientFillNode` 的覆写本体抽成 helper：

```go
// lowerGradientStops overwrites the Grad Colors stops XML in a gradient body
// (G-Fill or G-Stroke) with the runtime gradient. No-op if gradient is nil.
// Shared by lowerGradientFillNode / lowerGradientStrokeNode.
func lowerGradientStops(body *rifx.Chunk, gradient *codec.Gradient) *rifx.Chunk {
	if gradient != nil {
		overwriteGradientStopsXML(body, "ADBE Vector Grad Colors", codec.EncodeGradientXML(gradient))
	}
	return body
}
```

`lowerGradientFillNode` 改成：

```go
func lowerGradientFillNode(n *GradientFillNode, _ *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapeGradFillBody()
	if err != nil {
		return nil, err
	}
	return lowerGradientStops(body, n.Gradient()), nil
}
```

加 clone + lower for stroke（紧邻 cloneShapeGradFillBody / lowerGradientFillNode）：

```go
// cloneShapeGradStrokeBody returns a clone of the gradient-stroke template
// (templates/v2_2_shape_gradstroke_body.bin). Like the gradfill template it
// carries only `ADBE Vector Grad Colors`; stroke geometry stays at the
// extracted fixture values (deferred).
func cloneShapeGradStrokeBody() (*rifx.Chunk, error) {
	v22ShapeGradStrokeOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(v22ShapeGradStrokeBodyBytes))
		if err != nil {
			v22ShapeGradStrokeErr = fmt.Errorf("parse v22ShapeGradStrokeBodyBytes: %w", err)
			return
		}
		v22ShapeGradStrokeCache = ch
	})
	if v22ShapeGradStrokeErr != nil {
		return nil, v22ShapeGradStrokeErr
	}
	return cloneChunk(v22ShapeGradStrokeCache), nil
}

// lowerGradientStrokeNode emits a gradient-stroke body from the embedded
// template, overwriting the Grad Colors stops XML. Body logic is identical to
// lowerGradientFillNode; the only difference is which template is cloned.
func lowerGradientStrokeNode(n *GradientStrokeNode, _ *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapeGradStrokeBody()
	if err != nil {
		return nil, err
	}
	return lowerGradientStops(body, n.Gradient()), nil
}
```

- [ ] **Step 5: 加 match-name + switch case**

`shapeMatchNames`（line 235）加：

```go
	ShapeKindGradientStroke: "ADBE Vector Graphic - G-Stroke",
```

`lowerShapeNode` switch（line 269 的 GradientFillNode case 后）加：

```go
	case *GradientStrokeNode:
		return lowerGradientStrokeNode(node, ctx)
```

- [ ] **Step 6: 运行测试通过**

Run: `go test ./internal/serializer/ -run TestLowerGradientStroke -v && go test ./internal/serializer/ -v`
Expected: PASS（新测试 + 既有 gradfill 测试仍绿，证明 lower 重构无回归）

- [ ] **Step 7: Commit**

```bash
git add internal/serializer/lower_shape_node.go internal/serializer/lower_gradient_stroke_test.go
git commit -m "feat(gradstroke): lower path + shared lowerGradientStops helper"
```

---

### Task 4: reader — 抽 hydrateGradientStops（fill 重构）+ G-Stroke hydrate + switch case

**Files:**
- Modify: `internal/serializer/parse_shape_hydrate.go`（collectShapeKids switch line 117；hydrateGradientFillNode line 258）

- [ ] **Step 1: 写失败的 round-trip 单测**

Append `internal/serializer/parse_gradient_stroke_test.go`（同包 `serializer`，直接 lower→hydrate 验 round-trip，不经 collectShapeKids wrapper）:

```go
package serializer

import (
	"testing"

	"github.com/example/aep-parser/internal/scene"
)

// lowerGradientStrokeNode produces the node body; hydrateGradientStrokeNode
// must read the 3 stops back out (both in-package, no rifx wrapper needed).
func TestGradientStroke_RoundTrip(t *testing.T) {
	n := scene.NewGradientStrokeNode()
	if err := n.SetColorStops([]scene.GradientColorStop{
		{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0, 0}},
		{Offset: 0.5, Midpoint: 0.5, Color: [3]float64{0, 1, 0}},
		{Offset: 1, Midpoint: 0.5, Color: [3]float64{0, 0, 1}},
	}); err != nil {
		t.Fatal(err)
	}
	body, err := LowerShapeNodeForTest(n)
	if err != nil {
		t.Fatal(err)
	}
	got := hydrateGradientStrokeNode(body, nil)
	if got == nil {
		t.Fatal("hydrateGradientStrokeNode returned nil")
	}
	if len(got.Gradient().ColorStops) != 3 {
		t.Fatalf("round-trip color stops = %d, want 3", len(got.Gradient().ColorStops))
	}
}
```

> `collectShapeKids` 的 G-Stroke case 分派（Task 4 Step 4）由 Task 7 的 full-parse ship-gate（`aep.Open` 走完整 reader）间接验证；此单测聚焦 lower↔hydrate 本体。

- [ ] **Step 2: 运行确认失败**

Run: `go test ./internal/serializer/ -run TestGradientStroke_RoundTrip -v`
Expected: FAIL（hydrate 无 G-Stroke case → got nil）

- [ ] **Step 3: 抽共享 hydrateGradientStops + 重构 fill**

`internal/serializer/parse_shape_hydrate.go`，加 helper：

```go
// hydrateGradientStops decodes the Grad Colors stops XML from a gradient body
// (G-Fill or G-Stroke). Returns nil if absent (caller keeps the default node).
func hydrateGradientStops(body *rifx.Chunk) *codec.Gradient {
	if xml := findGradientStopsXML(body, "ADBE Vector Grad Colors"); xml != "" {
		return codec.ParseGradientXML(xml)
	}
	return nil
}
```

`hydrateGradientFillNode` 改成调它（保持局部降级语义）：

```go
func hydrateGradientFillNode(body *rifx.Chunk, _ *parseCtx) *GradientFillNode {
	n := scene.NewGradientFillNode()
	if g := hydrateGradientStops(body); g != nil {
		scene.SetGradientFillNodeGradient(n, g)
	}
	return n
}
```

> 若 `hydrateGradientFillNode` 当前用的是包内 `NewGradientFillNode`（无 `scene.` 前缀），保持其现有写法一致，别改前缀风格——只替换 stops 解码为 `hydrateGradientStops`。

- [ ] **Step 4: 加 hydrateGradientStrokeNode + switch case**

加（紧邻 hydrateGradientFillNode）：

```go
// hydrateGradientStrokeNode reads a G-Stroke body back into a runtime
// GradientStrokeNode. Local-degrade like G-Fill: missing Grad Colors → default
// node (node stays visible, parse never fails).
func hydrateGradientStrokeNode(body *rifx.Chunk, _ *parseCtx) *GradientStrokeNode {
	n := scene.NewGradientStrokeNode()
	if g := hydrateGradientStops(body); g != nil {
		scene.SetGradientStrokeNodeGradient(n, g)
	}
	return n
}
```

`collectShapeKids` switch（line 117 的 G-Fill case 后）加：

```go
		case "ADBE Vector Graphic - G-Stroke":
			if n := hydrateGradientStrokeNode(payload, ctx); n != nil {
				g.Children = append(g.Children, n)
			}
```

> 若 hydrate 函数用包内类型名（无 `scene.` 前缀），对齐 `hydrateGradientFillNode` 现有的引用风格。

- [ ] **Step 5: 运行测试通过**

Run: `go test ./internal/serializer/ -run TestGradientStroke -v && go test ./internal/serializer/ -v`
Expected: PASS（round-trip + 既有 fill hydrate 测试仍绿）

- [ ] **Step 6: Commit**

```bash
git add internal/serializer/parse_shape_hydrate.go internal/serializer/parse_gradient_stroke_test.go
git commit -m "feat(gradstroke): reader hydrate + shared hydrateGradientStops helper"
```

---

### Task 5: facade — 暴露 GradientStrokeNode 类型别名

**Files:**
- Modify: `internal/aep/aliases.go`（或 facade 暴露 shape 类型别名处——先 grep `GradientFillNode` 定位）

- [ ] **Step 1: 定位 facade 怎么暴露 GradientFillNode**

Run: `grep -rn "GradientFillNode\|VectorGroup =" internal/aep/aliases.go internal/aep/*.go`
Expected: 找到 `type GradientFillNode = scene.GradientFillNode`（或等价别名）。`AddGradientStroke` 是 `VectorGroup` 方法，VectorGroup 已暴露则方法自动可用，只需补节点类型别名。

- [ ] **Step 2: 加 GradientStrokeNode 别名（对称 GradientFillNode）**

在同文件 GradientFillNode 别名旁加：

```go
type GradientStrokeNode = scene.GradientStrokeNode
```

- [ ] **Step 3: 编译验证**

Run: `go build ./... && go vet ./...`
Expected: 无错误

- [ ] **Step 4: Commit**

```bash
git add internal/aep/aliases.go
git commit -m "feat(gradstroke): expose GradientStrokeNode via facade alias"
```

---

### Task 6: 全量编译 + 测试门

**Files:** 无新增（验证门）

- [ ] **Step 1: 全量 vet + test**

Run: `go vet ./... && go test ./...`
Expected: 全 PASS。若有按 ShapeNodeKind 的 exhaustive switch 漏了 GradientStroke（Task 2 Step 6 登记的），此处会暴露——补齐后重跑。

- [ ] **Step 2: 边界守卫测试（DAG import）**

Run: `go test ./internal/aep/ -run TestArch`
Expected: PASS（未引入逆向 import）

- [ ] **Step 3: Commit（若有补漏）**

```bash
git add -A
git commit -m "test(gradstroke): full vet+test green; fill switches exhaustive"
```

（无改动则跳过。）

---

### Task 7: 双版本 AE ship-gate（color + alpha）

**Files:**
- Create: `test_data/verify_v2_2_gradstroke.jsx`
- Create: `internal/aep/shape_gradient_stroke_shipgate_test.go`

- [ ] **Step 1: 写 JSX 验证脚本**

复制 `test_data/verify_v2_2_gradient.jsx` → `test_data/verify_v2_2_gradstroke.jsx`，改两处：
1. 期望的 layer 名 `Grad_Static` → `GradStroke_Static`
2. 节点查找 `ADBE Vector Graphic - G-Fill` → `ADBE Vector Graphic - G-Stroke`
3. **顺手验 color stop 数 = 3**（若原 JSX 未验，加 `gradColors.value` 的 stop 计数检查，写进 done 文件 PASS/FAIL）

先 `Read test_data/verify_v2_2_gradient.jsx` 看其确切结构再改（保持 args JSON 读取 / done 写入 / resave 逻辑不变）。

- [ ] **Step 2: 写 ship-gate test（对称 GradientFill）**

Create `internal/aep/shape_gradient_stroke_shipgate_test.go`:

```go
// AE ship gate for gradient stroke (AddGradientStroke). Builds one ShapeLayer
// (rect + gradient stroke, 3 custom color stops + non-default alpha ramp), has
// AE open+resave it, and decodes the resaved Grad Colors stops via the read
// path. Gated by AE_SHIP_GATE.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/codec"
)

func runV2_2GradientStrokeShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "v2_2_gradstroke.aep")
	resavedAEP := filepath.Join(tempDir, "v2_2_gradstroke.resaved.aep")
	doneFile := filepath.Join(tempDir, "v2_2_gradstroke.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/v2_2_gradstroke_gate_args.json`

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, err := aep.NewShapeLayer(comp, "GradStroke_Static")
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	rect, _ := l.RootGroup().AddRect()
	_ = rect.SetSize([2]float64{200, 100})
	gs, _ := l.RootGroup().AddGradientStroke()
	if err := gs.SetColorStops([]codec.GradientColorStop{
		{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0, 0}},
		{Offset: 0.5, Midpoint: 0.5, Color: [3]float64{0, 1, 0}},
		{Offset: 1, Midpoint: 0.5, Color: [3]float64{0, 0, 1}},
	}); err != nil {
		t.Fatalf("SetColorStops: %v", err)
	}
	// Non-default alpha ramp — exercises the alpha write path (else untested).
	if err := gs.SetAlphaStops([]codec.GradientAlphaStop{
		{Offset: 0, Midpoint: 0.5, Alpha: 1.0},
		{Offset: 1, Midpoint: 0.5, Alpha: 0.3},
	}); err != nil {
		t.Fatalf("SetAlphaStops: %v", err)
	}

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP))
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	jsxPath := `E:/projects/tools/aep-parser/test_data/verify_v2_2_gradstroke.jsx`
	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	if lines := strings.SplitN(string(content), "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("gradient stroke ship gate FAIL:\n%s", string(content))
	}

	proj, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("open resaved: %v", err)
	}
	var grad *codec.Gradient
	for _, c := range proj.Compositions {
		for _, ly := range c.Layers {
			for _, pr := range ly.Properties {
				if pr.MatchName == "ADBE Vector Grad Colors" && pr.Gradient != nil {
					grad = pr.Gradient
				}
			}
		}
	}
	if grad == nil {
		t.Fatal("resaved file has no decodable gradient — AE dropped the stops?")
	}
	if len(grad.ColorStops) != 3 {
		t.Fatalf("resaved color stops = %d, want 3:\n%+v", len(grad.ColorStops), grad.ColorStops)
	}
	wants := [3][3]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}
	for i, w := range wants {
		got := grad.ColorStops[i].Color
		for c := 0; c < 3; c++ {
			if abs(got[c]-w[c]) > 0.02 {
				t.Errorf("resaved color stop %d = %v, want ~%v", i, got, w)
				break
			}
		}
	}
	// Alpha path: last stop alpha ~0.3.
	if n := len(grad.AlphaStops); n >= 2 {
		if a := grad.AlphaStops[n-1].Alpha; abs(a-0.3) > 0.02 {
			t.Errorf("resaved last alpha = %g, want ~0.3", a)
		}
	}
}

func TestV2_2_GradientStroke_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2GradientStrokeShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_GradientStroke_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2GradientStrokeShipGate(t, aep.TargetAE2020, aeExe)
}
```

> `abs` / `runAeRunShipGate` 已在 `internal/aep` 测试包存在（`shape_gradient_shipgate_test.go` 同包），直接复用，不要重定义。

- [ ] **Step 3: 跑 AE 2025 ship-gate**

Run: `AE_SHIP_GATE=1 go test ./internal/aep/ -run TestV2_2_GradientStroke_AEShipGate_AE2025 -v -timeout 600s`
（PowerShell: `$env:AE_SHIP_GATE=1; go test ./internal/aep/ -run TestV2_2_GradientStroke_AEShipGate_AE2025 -v -timeout 600s`）
Expected: PASS。AE 2025 冷启动可能需 warm retry（见 memory `feedback_ae_ship_gate_self_serve`）。failure 时按 incident `ae-automation-occlusion-crashstate.md` 抓 dialog。

- [ ] **Step 4: 跑 AE 2020 ship-gate**

Run: `$env:AE_SHIP_GATE=1; go test ./internal/aep/ -run TestV2_2_GradientStroke_AEShipGate_AE2020 -v -timeout 600s`
Expected: PASS（跨版本：AE25-shaped gradient 被 AE 2020 接受，同 GradientFill）

- [ ] **Step 5: Commit**

```bash
git add test_data/verify_v2_2_gradstroke.jsx internal/aep/shape_gradient_stroke_shipgate_test.go
git commit -m "test(gradstroke): dual-version AE ship-gate (color + alpha) PASS"
```

---

### Task 8: aepdemo + 文档/coverage 同步

**Files:**
- Modify: `cmd/aepdemo/main.go`（line 78-84 的 AddGradientFill 演示旁）
- Modify: `flightdeck/plans/coverage.md` + `coverage-detail.md` + `flightdeck/cockpit.md`
- Regenerate: `docs/*`（docgen）

- [ ] **Step 1: aepdemo 加对称演示**

`cmd/aepdemo/main.go`，在 ④ 渐变填充演示后加一个图层：

```go
	ls := layer("⑤ 渐变描边 AddGradientStroke(红→蓝)", ...)
	rs, _ := ls.RootGroup().AddRect()
	must(rs.SetSize([2]float64{300, 190}))
	gst, _ := ls.RootGroup().AddGradientStroke()
	must(gst.SetColorStops([]aep.GradientColorStop{
		{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0, 0}},
		{Offset: 1, Midpoint: 0.5, Color: [3]float64{0, 0, 1}},
	}))
```

> 先 `Read cmd/aepdemo/main.go:60-90` 对齐 `layer(...)` 辅助签名与 `aep.GradientColorStop` 的实际别名名（可能是 `codec.GradientColorStop`）。

Run: `go run ./cmd/aepdemo`（按其用法）→ 确认无 panic、产出含 gradient stroke 图层。

- [ ] **Step 2: docgen 重生成**

Run: `go run ./cmd/docgen -manifest docs/docgen.json`（或 `go generate ./cmd/docgen`）
Expected: `docs/*.gen.md` 含新 `AddGradientStroke` / `GradientStrokeNode`。

- [ ] **Step 3: coverage / cockpit 同步**

`flightdeck/plans/coverage.md`：在「子项⑭ Gradient fill」行后加「子项⑮」行（**对齐既有格式**）：

```markdown
#### V2.2.1 子项⑮ (2026-06-10) — Gradient stroke (AddGradientStroke: color + alpha stops, R/W)
- `(g *VectorGroup) AddGradientStroke()` → `GradientStrokeNode`；`SetColorStops` / `SetAlphaStops`（复用 fill 的共享 setter）+ reader hydrate（G-Stroke 节点）。模板克隆 + 覆写 Grad Colors XML（复用 `lowerGradientStops`）。**AE 2020+2025 双版本 ship-gate PASS**（color + alpha）。stroke 几何 + ramp geometry deferred（同 fill）。
```

`coverage-detail.md` 对应补一行（同格式）。`cockpit.md`：`## 下一步` / Active focus 反映 GradientStroke 已 ship（plan 翻 done 后 landing 会同步，此处先记一句）。

- [ ] **Step 4: 最终验证门**

Run: `go vet ./... && go test ./...`
Expected: 全 PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/aepdemo/main.go docs/ flightdeck/plans/coverage.md flightdeck/plans/coverage-detail.md flightdeck/cockpit.md
git commit -m "docs(gradstroke): aepdemo demo + coverage 子项⑮ + docgen regen"
```

---

## 完成后

- 本 plan 代表一个可交付 feature arc（GradientStroke R/W），8 task 全绿 + 双版本 ship-gate PASS 后翻 `status: done`，跑 `/flightdeck:landing` 归档 + 同步 cockpit（rules「大计划完成即 landing」）。
- 同 spec（`2026-06-10-gradient-stroke-write.md`）一并归档。
