---
status: idea
summary: V3 M8 方案② 真·物理分包设计（A 先行）：internal/scene + internal/serializer + internal/codec 三包拆分，free-function/Document API，eager patch 迁 serializer 侧表，opaque shard 作 C 的 on-ramp；保 byte-exact 回归门，C（懒重生·改契约）作日后独立 arc
note: brainstorm 产出，未写 plan。方向已与用户确认（A 先行 + facade 保留 + Document 写中枢）；§1 骨架已逐段过，§2–§6 直接落文档。前置：M8 scene→rifx 白名单清零（2026-06-07 已落）
---

# V3 M8 方案② — 真·物理分包设计（A 先行）

**Status**: idea（brainstorm 产出，未写 plan）。Drafted 2026-06-07。
**关联**: `2026-05-22-v3-direction.md`（M8 / scene→rifx 残留表）、`landed/specs/2026-05-30-aep-package-reorg-design.md`（方案① §0 Go 语义墙 + §6 back_ 缝固化）、`2026-05-27-v3-deep-think.md`（Q1/Q5/Q7 + opaque shard）、CLAUDE.md 硬约束 #1/#2/#3/#5/#6。

---

## §0 决策记录（先读，定范围）

立项问题：V3 M8 的「真·物理分包」（独立 `internal/scene` + `internal/serializer` Go 包），被方案① §0 的 **Go 语义墙**挡住——方法必须与类型同包，`(p *Project) WriteAEP` 方法体调 serializer + serializer 读 Project 字段 = 导入环。方案① 当时收敛为「单包命名轴重组」，把破环留给方案②。本 spec 设计方案②。

**brainstorm 期四个抉择（与用户逐一确认）：**

1. **分包动机** = **硬编译边界**。把当前 `arch_boundary_test.go`（AST 测试守卫，可禁用、非气密）升级为 Go 编译器强制的隔离：scene 逻辑*永远*碰不到 serializer/rifx 内部。

2. **破环路线** = **自由函数 · 破 Stable 方法 API**（非接口依赖倒置）。scene 类型住 `internal/scene`，serializer 单向 import scene，写入走 free function / Document 方法。破 `WriteAEP` 等方法签名（约束#2），接受 compat shim / 版本跳。理由：接口倒置要镜像 serializer 读到的*每个* accessor，接口面爆炸；自由函数路线接口面最小、设计最简。

3. **back-ref / eager 写路径** = **侧表 + 保 eager**（选项 A）。back-ref（持 `*rifx.Chunk`）从 scene 类型迁入 serializer 侧 Document 的 per-object 侧表（以 scene 指针为 key）；`Set*` 变 Document 方法，经侧表做 length-preserving 字节 patch。scene 零 rifx import。

   否决「懒重生·纯值 scene」（选项 C）的当下理由见下「A→C 排序」。

4. **范围/契约** = **A 先行，C 日后独立决**。保 byte-exact round-trip 契约（82-fixture 回归门全程当安全网），不在本 arc 改契约。

**A→C 排序论证（本设计的核心判断）：**

证据（brainstorm 期三组并行调查，见 §8 证据表）确立：

- **A 是 C 的必经之路，不是对立选项**。opaque chunk 不能住 scene（=rifx import，破边界），所以 C 也必须有 serializer 侧的 per-object 字节存储。A 的「back-ref 侧表」与 C 的「opaque shard 侧表」是同一个容器；差别只在写语义（A 原地 patch / C 重生+opaque 拼接）。A 的全部管道（Document、free function、侧表）C 全要。**做 A 不浪费一行**。
- **big-bang C 把三个风险耦进一次不可验证的跳跃**（分包 + 改契约 + 懒重生），且在填 ~5–10 处 partial-decode 坑时丢掉 byte-identical 这个每 commit 跑的廉价回归门。
- **A 先行**：在做最危险的结构动作（分包 + back-ref 迁侧表）时全程保留 byte-identical 门当安全网；顺手填 opaque shard（= C 的 on-ramp）；等边界立住、AE ship-gate + semantic-diff 覆盖到位，再在独立 follow-on 有意识地放宽契约切 C。

> 一句话：追 V3 终态（高收益）的正确姿势是 **A 先立边界 + 填 opaque on-ramp，再独立 arc 切 C**，而非一次性赌掉回归门。

---

## §1 包拓扑与依赖 DAG

五个包，箭头 = 允许 import，整体无环（DAG）：

```
internal/aep（薄 facade，顶层，下游唯一入口）
  ├─imports→ internal/serializer
  └─imports→ internal/scene

internal/serializer（唯一同时见 scene + rifx 的包；序列化知识全集中）
  ├─imports→ internal/scene
  ├─imports→ internal/codec
  └─imports→ internal/rifx

internal/scene（纯运行时模型）
  └─imports→ internal/codec        # 仅取值类型/枚举；零 rifx、零 serializer

internal/codec（纯值/字节 codec，§7 已标纯叶）
  └─imports→ internal/rifx          # 仅 []byte / Chunk 值；零 scene 类型

internal/rifx（RIFX framing，不变）
  └─ 依赖：无
```

**编译期强制的不变量（取代 AST 守卫）：**

- `scene` 的 import 集 = {codec}。**编译器**保证 scene 永远碰不到 rifx/serializer —— 这就是「硬编译边界」目标的兑现。
- `codec` 的 import 集 = {rifx}，零 scene 类型（§7 已铺）。
- `serializer` 是唯一同时见 scene + rifx 的包。
- `aep` facade 顶层，无人依赖 → 无环。

**文件搬迁（基于证据三的清单）：**

| 现 `internal/aep/` | 去向 |
|---|---|
| scene 类型定义 + 逻辑 accessor + 图级结构 mutation（`scene_*`） | `internal/scene` |
| 纯 codec（`codec_*`，§7 已标） | `internal/codec` |
| `parse_*` / `lower_*` / `write_*` / `back_*`（10 backref struct）/ `mutate_*` 字节侧 | `internal/serializer` |
| `Open`/`Write` 入口 + 类型别名 | `internal/aep` facade |

**API 形态变化（破点，但保留方法手感）：**

- `proj.WriteAEP(w)` → `doc.Write(w)`（`Document` 方法，合法——Document 是 serializer 类型，持 chunk 树）。
- `layer.SetVisible(v)` → `doc.SetLayerVisible(layer, v)`（Document 方法，经侧表查 back-ref 做 length-preserving patch）。
- 读不变：`doc.Scene.Compositions[0].Layers[1].Position()`（scene 纯值，getter 全 eager，无需 doc）。

**§1 决策（默认拍，可在 review 推翻）：**

- **D1.1 保留 `internal/aep` facade**（vs 下游直接 import scene+serializer）：下游 `cmd/aepdemo` 单 import 不变，churn 最小。facade = 类型别名（`type Project = scene.Project`、`type Document = serializer.Document`）+ `aep.Open`/`aep.Write` 薄包装。
- **D1.2 写操作集中到 `Document` 方法**（`doc.SetLayerVisible(...)`）：保留方法调用手感，写入中枢单一，便于 §2 侧表查找。

---

## §2 Document 容器 + back-ref 侧表 + Parse/Write 签名

**核心结构（serializer 包）：**

```go
// Document 配对「纯逻辑 scene 视图」与「序列化 source-of-truth（chunk 树）」，
// 并持有 scene→chunk 的 back-ref 侧表 + opaque shard。下游持 *Document。
type Document struct {
    Scene *scene.Project   // 纯逻辑视图（零 rifx）
    root  *rifx.Chunk      // 序列化 source-of-truth（verbatim + patch，约束#1/#5）

    // back-ref 侧表：以 scene 指针 identity 为 key（scene 对象在 []*T 里指针稳定）
    layerBack    map[*scene.Layer]*layerBackrefs
    propBack     map[*scene.Property]*propertyBackrefs
    compBack     map[*scene.Composition]*compositionBackrefs
    kfBack       map[*scene.Keyframe]*keyframeBackrefs
    markerBack   map[*scene.Marker]*markerBackrefs
    maskBack     map[*scene.Mask]*maskBackrefs
    footageBack  map[*scene.Footage]*footageBackrefs
    grpBack      map[*scene.AEPropertyGroup]*propertyGroupBackrefs
    projBack     *projectBackrefs
    rqBack       *renderQueueBackrefs
    opaque       opaqueRegistry  // §4：per-object 未解码 sibling 字节
}
```

10 个 `*Backrefs` struct（compositionBackrefs / layerBackrefs / propertyBackrefs / keyframeBackrefs / markerBackrefs / maskBackrefs / footageBackrefs / projectBackrefs / renderQueueBackrefs / propertyGroupBackrefs）原样从 `back_*.go` 迁入 serializer，**内容不变**（仍持 `*rifx.Chunk`），只是从「scene 类型的 `back` 字段」改成「Document 侧表的 map value」。

**入口签名：**

```go
func Parse(r io.ReadSeeker) (*Document, error)   // chunk→scene + 填侧表 + 填 opaque
func (d *Document) Write(w io.Writer) error       // 写 d.root（重算 LIST size）
func (d *Document) WriteJSON(w io.Writer) error   // 走 d.Scene 单向导出
```

**Parse 流程：** 解析 chunk 树 → 构造 `scene.*` 纯值对象 → 每解出一个对象，把其 backref struct 存进对应 map（`d.layerBack[layer] = &layerBackrefs{ldta: …}`）→ 填 opaque（§4）。scene 对象只持逻辑值，永不持 chunk。

**Set\*（Document 方法）双写：**

```go
func (d *Document) SetLayerVisible(l *scene.Layer, v bool) error {
    br := d.layerBack[l]
    if br == nil {
        return fmt.Errorf("SetLayerVisible: layer not tracked by this Document")
    }
    // 1) length-preserving 字节 patch（逻辑原样从旧 write_layer.go 搬来）
    patchLdtaVisible(br.ldta, v)
    // 2) 同步 scene 字段，保证后续读一致
    l.Visible = v
    return nil
}
```

**Write 流程：** length-preserving setter 在 set 时已就地改 `br.*.Data`（slice 引用即树内字节）；length-variable（name/comment/expression/text/keyframe insert）在 set 时已 splice `root` 子树 + 标记。`d.Write` = `d.root.Write(w)`，由既有 `rifx.Chunk.PayloadSize()` + 递归 `Write` 重算所有父 LIST size（机制不变，证据三确认 size 重算已集中在 rifx 层）。

**从零建（NewProject / NewComposition / NewShapeLayer 等结构性 new）：** 走 `mutate_*` lower 路径建 chunk 子树 + 同步往 Document 侧表注册 backref，使新对象也可被后续 Set* 命中。结构性 op 仍遵 V2.1 atomic（warnings-as-failure + rollback），约束#6 ship-gate 不变。

---

## §3 Set\* → Document 方法（API 破点 + facade + compat）

证据三量化：~283 个 `Set*`、~550 调用点（~75 非测试 + ~475 测试）、getter 几乎全 eager（额外 churn 低）。

**转换规则：**

- 每个经 back-ref 做字节 patch 的 `(recv) SetX(args)` → `(d *Document) SetRecvX(recv, args)`，方法体 = 旧逻辑 + 侧表查找 + scene 字段同步。
- 纯 scene 图级 setter（不碰 chunk，如 shape IR 构建 `RectNode.SetSize`）**留在 scene 包**——它们不需要 back-ref，零 rifx，本就属 scene。证据三的 283 里含大量这类（RectNode/StrokeNode/FillNode/…）；这部分*不动*。
- getter 全留 scene（eager 值，已解进 scene 字段）。少数 lazy flag 读（IsSpatial/IsAnimated 等读 tdb4）：解析期预解进 scene bool 字段，消除 lazy 依赖。

> **澄清**：283 是「名字以 Set 开头」的总数，**真正需迁的是「经 back-ref 字节 patch」的子集**（Layer/Composition/Property/Footage/Marker/Mask/Project/RQ/OM 的 length-preserving setter）。shape/stroke/fill 等 IR builder setter 留 scene。plan 阶段第一步即精确分类这两类（grep `\.back\.` / `d\.\w+Back\[`）。

**facade（`internal/aep`）：**

```go
type Project    = scene.Project
type Composition = scene.Composition
type Layer       = scene.Layer
// … 全部公共 scene 类型 + 枚举别名
type Document   = serializer.Document

func Open(path string) (*Document, error)        // = serializer.Parse(file)
func OpenReader(r io.ReadSeeker) (*Document, error)
```

**compat / 破 API 处置（约束#2）：**

- 这是显式的 Alpha-breaking 重构；commit message 标 `BREAKING`。
- `cmd/aepdemo`（唯一非测试消费者）改 `proj.WriteAEP(f)` → `doc.Write(f)`、`layer.SetX` → `doc.SetX`。
- 无外部消费者（证据一确认）→ 不做长期 shim；如需平滑，可在 facade 临时保留 `func (d *Document) WriteAEP(w) error { return d.Write(w) }` 一轮后删。

---

## §4 opaque shard 填充（C 的 on-ramp）

证据二：opaque 基建已存在（`propertyBackrefs.opaque` / `layerBackrefs.opaque`）但**当前为 nil**；partial-decode 散点 ~5–10 处（dimsep `@0x08/@0x10` 缓存字段、tdum/tduM 未解码、ldta padding、keyframe `@0x07` header、cdta 尾 padding、btds/btdk opaque list、effect sub-chunk）。

**本 arc 做（轻量，不改写语义）：**

- Parse 时把每个 scene 对象**未解码的 sibling/tail chunk** 收进 Document 的 `opaqueRegistry`（per-object，以 scene 指针为 key），与 backref 侧表并列。
- 当前写路径（verbatim + patch）**不用** opaque shard（chunk 树已含原字节）；填它纯为 C 铺路——C 改成「从 scene 重生 + 原位拼回 opaque」时直接取用。
- **本 arc 不碰** ~5–10 处 partial-decode 的字节级重生逻辑（那是 C 的活）。只确保 opaque 收集**完整**（round-trip 仍 byte-identical 即证明收集无遗漏）。

> §4 是「填 on-ramp」而非「切 C」。若 review 认为 YAGNI，可降级为本 arc 不填、留 C 时再填（§1 范围选项三）；当前默认填，因边界刚立时填 opaque 的 blast radius 最小。

---

## §5 迁移分期（strangler，步步绿，全程保 byte-exact 门）

每阶段独立可 commit、suite 全绿、byte-identical round-trip 不破。

**核心排序洞察（自审修正）**：scene 结构体只要还持 `back *layerBackrefs`（内含 `*rifx.Chunk`），就**不可能**成为零 rifx 的独立包——所以「先抽 scene、再迁 backref」是不可能的；**backref 迁侧表这个动作本身，才是让 scene 可抽出的前提**。故正解是**先在单包内做逻辑解耦（backref → Document 侧表），把 scene 结构体的 rifx 耦合清零，再把物理分包做成近乎机械的 `git mv`**（make the change easy, then make the easy change——同方案① 哲学）。危险的是 P2（解耦），P3（分包）只是机械收割。

- **P0 基线**：`tmp_debug/` 建分包 round-trip 基线脚本（全 82 fixture 跑 `Open→Write→bytes` 存指纹 + 记 fixture git hash）。沿用方案① §8 套路。**exit**：基线就位，否则后续比对无依据。
- **P1 抽 `internal/codec`**：`codec_*` 纯叶平移成真包（§7 已验纯度，零 scene 耦合）。真包但独立、低风险、零 API 影响。**exit**：`go build ./...` + 82-fixture byte-identical 绿。
- **P2 单包内逻辑解耦（最危险，byte-exact 门当安全网）**：**仍在单包 `internal/aep`**，不动包边界。引入 `Document`；把 10 个 backref struct 从「scene 类型的 `back` 字段」逐类迁入「Document 侧表 map」；把 back-ref 子集 `Set*` 逐个改成 Document 方法（每迁一个 `go test` 绿）；填 opaque（§4）。**末态**：scene 结构体零 `*rifx.Chunk`/零 backref 字段（grep `\.back\.` 清零），但仍单包。**exit**：byte-identical 全绿 + scene 结构体 rifx 引用 grep 清零。
- **P3 物理分包（机械收割）**：此时 scene 结构体已 rifx-free → `git mv` `scene_*` 进 `internal/scene`、serializer 侧（parse_/lower_/write_/back_/mutate_ + Document）进 `internal/serializer`；建 `internal/aep` facade（类型别名 + Open/Write）。**编译边界此刻由编译器强制**。因耦合已在 P2 斩断，本阶段以机械移动为主。**exit**：各包独立编译 + `go list -deps ./internal/scene` 不含 rifx/serializer（编译期边界 CI 断言）+ byte-identical 全绿 + `go doc -all` facade 公共面 diff 逐条核对（破点是预期 diff）。
- **P4 下游切换 + 收口**：`cmd/aepdemo` 切新 API（`doc.Write`/`doc.Set*`）；删旧 `WriteAEP` 方法（或留一轮 shim）；删/改 `arch_boundary_test.go`（scene⊥rifx 已由包边界编译期保证，AST 守卫降级为 serializer 包内命名轴 lint）；CLAUDE.md 硬约束#3 更新（单包 → 多包 + 新 DAG）；终验 **AE 2020 + 2025 双版本 ship-gate 全套**（约束#6，自验留痕）；cockpit/INDEX 落地。

> 阶段序由风险驱动：codec（独立纯叶）→ **P2 单包内 backref 解耦（真正的活）** → P3 机械分包 → P4 收口。回归门（P0）必须在 P2（解耦）之前就位——这是硬要求。

---

## §6 验证 / 测试 / ship-gate / 边界强制

每阶段（每搬一个文件 / 每批改名）必须：

1. `go build ./...` 绿。
2. `go vet ./... && go test -count=1 ./...` 全绿（含 ship-gate 非门控部分）。
3. **byte-identical round-trip**（约束#1/#5）：对 82 fixture 重跑 `Open→Write→bytes`，断言与 P0 基线指纹相同。这是分包全程的安全网。
4. **公共 API diff**（约束#2）：`go doc -all` 前后对比。**P1 须零 diff**（codec 纯搬迁）；**P2 起引入预期破点**——diff 必须逐条等于设计的破点集（`Set*`→Document 方法，`WriteAEP` 方法→`Document.Write`），无意外增删。
5. **边界编译期强制**：scene 包 import 集断言（可加一个 `go list -deps` 检查或 CI 步骤：`internal/scene` 的依赖闭包不含 `internal/rifx`、`internal/serializer`）。这是把「硬编译边界」做成 CI 可执行断言。
6. **终验 AE 双版本 ship-gate**（约束#6）：P5 跑全套 `AE_SHIP_GATE=1`（AE 2020 + 2025），确认纯结构搬迁零回归。无结构性写路径语义变更，故中途不必逐阶段跑，收尾跑一次。

---

## §7 非目标 / 范围

- **不改契约**：byte-exact round-trip 仍是硬不变量；不切懒重生（C 是日后独立 arc）。
- **不改写语义**：eager length-preserving patch 逻辑原样搬迁，不重写；length-variable 路径不动。
- **不做接口依赖倒置**：走 free-function/Document 路线（§0 抉择 2）。
- **零行为变更（P1–P2、P4–P5）**：纯搬迁 + 改名 + facade；唯一预期变更是 P3/P4 的 API 破点（方法→Document 方法/free function）。
- **不顺手重构无关逻辑**、不改算法、不扩 capability matrix / shape graph / effect schema（V3 其他 M）。
- **不填 partial-decode 重生逻辑**（C 的活）；§4 只收集 opaque，不消费。

---

## §8 风险

| 风险 | 缓解 |
|---|---|
| back-ref 迁侧表引入隐性行为变更 | P0 byte-identical 基线在 P2（迁侧表）之前就位，每步比对 |
| 指针 identity 侧表失效（scene 对象被值拷贝/重分配） | scene 对象在 `[]*T` 里以指针存放，identity 稳定；plan 阶段审计有无按值传递 scene struct 的路径 |
| Set\* 分类误判（把纯图 setter 当 back-ref setter 迁，或反之） | P2 第一步精确 grep 分类（`\.back\.` / 侧表访问），分类表入 plan 评审 |
| API 破点波及测试 ~475 处 churn | 机械改写；facade 别名降低 import churn；分阶段每步 `go test` 绿 |
| opaque 收集遗漏（§4） | byte-identical round-trip 即遗漏探测器——漏收一个 chunk 则指纹必变 |
| 下游 `cmd/aepdemo` 编译破 | P4 同步切换 + `go build ./cmd/...` 门 |
| 范围蔓延到 C（懒重生/改契约） | §0/§7 硬边界；§4 只填 on-ramp 不消费 |
| AE 双版本 ship-gate 回归 | §6 P5 终验全套（约束#6，自验留痕） |

**brainstorm 期证据表（A 先行依据）：**

| 调查 | 结论 | 关键出处 |
|---|---|---|
| byte-exact 是否产品契约 | **内部回归工具**，非产品承诺；README/CLAUDE 只诺 length-preserving + opaque + AE 可重开；无外部消费者 | README:4/37,CLAUDE:24/32,cmd/* |
| C 重生脆弱面 | **可控 ~5–10 散点**；opaque 基建已存在但当前 nil，要 C 须先填 | 各 incident,back_*.go |
| A churn | **变更面高/数据访问面低**：~283 Set*、~550 调用点（~75 非测试+~475 测试）、10 backref struct、getter 全 eager；~1–2 周 | back_*.go,全仓 grep |

---

## §9 决策记录

- 2026-06-07 brainstorm：用户选 V3 M8 方案②（真·物理分包）为下一焦点。四抉择确认：①动机=硬编译边界 ②路线=自由函数·破 API ③back-ref=侧表+保 eager（A）④范围=A 先行、C 日后独立。
- 2026-06-07 三组并行证据调查 → 确立「A 是 C 的必经之路」，否决 big-bang C（三风险耦合 + 丢回归门）。
- §1 骨架与用户逐段过；§1 两开放点按推荐默认拍（D1.1 保 facade、D1.2 Document 写中枢），待 spec review 确认。

## 关联文档

- `2026-05-22-v3-direction.md` — M1–M8 框架 + scene→rifx 残留表（M8 前置解耦已清零）
- `landed/specs/2026-05-30-aep-package-reorg-design.md` — 方案① §0 Go 语义墙 + §6 back_ 缝固化（本 spec 的直接前置）
- `2026-05-27-v3-deep-think.md` — Q1 软迁移 / Q5 opaque shard / Q7 eager+lazy 共存
- `CLAUDE.md` § 硬约束 #1/#2/#3/#5/#6
- `incidents/separate-dimensions-write-mechanics.md` — `@0x08/@0x10` 缓存字段（partial-decode 样例）
