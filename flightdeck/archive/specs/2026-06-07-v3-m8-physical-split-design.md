---
status: done
summary: V3 M8 方案② 真·物理分包设计（A 先行 + B′ back-ref 接口）：scene/serializer/codec 物理拆包，back-ref 作 scene 内 writer 接口（serializer 实现）→ 保留全部方法 API（不破 API）、scene 编译期零 rifx；eager length-preserving patch 经接口；opaque 延后到 C；全程保 byte-exact 回归门
note: brainstorm + 三家外审两轮整合。方向 = A 先行（保 byte-exact，C 日后独立）+ B′（back-ref 接口、不破 API、无侧表/无 Document god-object）。
---

> **✅ 已实现并落地（2026-06-09）** —— M8 物理分包按本设计完成（plan `archive/plans/2026-06-07-v3-m8-physical-split-plan.md`，逐 commit 见 git log）。opaque「C 阶段」未单独做：当前 round-trip 已 byte-identical（opaque 经现有 back-ref 保留），无独立 C 需求。

# V3 M8 方案② — 真·物理分包设计（A 先行 + B′ back-ref 接口）

**Status**: active（brainstorm + 三家外审两轮已整合，未写 plan）。Drafted 2026-06-07。
**关联**: `2026-05-22-v3-direction.md`（M8 / scene→rifx 残留表）、`landed/specs/2026-05-30-aep-package-reorg-design.md`（方案① §0 Go 语义墙 + §6 back_ 缝固化）、`2026-05-27-v3-deep-think.md`（Q1/Q5/Q7 + opaque shard）、CLAUDE.md 硬约束 #1/#2/#3/#5/#6。

---

## §0 决策记录 + 核心设计约束（先读，定范围）

立项问题：V3 M8 的「真·物理分包」（独立 `internal/scene` + `internal/serializer`），被方案① §0 的 **Go 语义墙**挡住——方法必须与类型同包，`(p *Project) WriteAEP` 方法体调 serializer + serializer 读 Project = 导入环。方案① 当时收敛为「单包命名轴重组」，把破环留给方案②。本 spec 设计方案②。

### §0.1 brainstorm + 外审收敛的四个抉择

1. **分包动机 = 硬编译边界**。把 `arch_boundary_test.go`（AST 测试守卫，可禁用、非气密）升级为 Go 编译器强制：scene 逻辑*永远*碰不到 serializer/rifx 内部。

2. **破环路线 = back-ref 接口依赖倒置（B′），保留方法 API**。**这推翻了 brainstorm 初期「自由函数·破 API」的决策**——理由是当初对接口倒置的定价错误：我误以为「保 API 必须镜像 serializer 读到的*每个* read accessor → 接口爆炸」。真相是 serializer 在物理分包后**直接 import scene、具体读字段**，无需 read-accessor 接口；**只有写回（eager patch）那一小撮操作需要接口**（每个 back-ref 类型一个 writer 接口，有界）。故 `WriteAEP`/`Set*` 可保留为 scene 方法，方法体调 `p.back.X(...)`（scene 本地接口），serializer impl 具体读 scene、写 chunk —— 无环、边界硬、**API 不破**。

3. **back-ref / eager 写路径 = B′（scene 内 writer 接口 + serializer 实现）**。**这推翻了初期「侧表（A）」的决策**——外审两轮火力几乎全打在 A 的「Document 侧表 / 指针 identity map / Document god-object」上（见 §10）。B′ 让 back-ref 成为 scene 对象上的接口字段（今天是 `back *xBackrefs` 具体类型，B′ 改成 `back XWriter` 接口），impl 在 serializer。**无侧表、无指针 identity map、无 doc handle、无 Document god-object**，从根上消解那一整类外审。

4. **范围/契约 = A 先行，C 日后独立决**。保 byte-exact round-trip 契约（82-fixture 回归门全程当安全网），不改契约、不切懒重生（C = lazy regen + 契约放宽为「AE 接受 + 语义等价」，是日后独立 arc）。

### §0.2 A→C 排序论证（为何不直接 big-bang C）

证据（brainstorm 期三组并行调查，数据见 §9 证据表）确立：

- **本 arc（A 先行 + B′）为 C 提供可复用基础设施**（外审 gpt#178-189 纠「必经之路」过强：已证「scene/rifx 边界有价值 + serializer 持 chunk 有价值 + eager patch 保回归门」三项，**未证**「C 必用同一组 writer 接口」——writer 接口编码当前 eager 行为，C 改懒重生可能重构之，见 C-7）。但三项地基 C 确定要用：C（懒重生）消不掉 serializer 侧 per-object 字节存储（opaque chunk 不能住 scene = rifx import，破边界），本 arc 立起的「scene 零 rifx 边界 + serializer 持 chunk 树」正是 C 的地基。差别只在写语义（本 arc 原地 patch / C 重生+opaque 拼接）。
- **big-bang C 把三风险耦进一次不可验证的跳跃**（分包 + 改契约 + 懒重生），且在填 ~5–10 处 partial-decode 坑时丢掉 byte-identical 这个每 commit 跑的廉价回归门。
- **本 arc 先行**：在做最危险的结构动作（back-ref 接口化 + 物理分包）时全程保留 byte-identical 门当安全网；等边界立住、AE ship-gate + semantic-diff 覆盖到位，再独立 arc 切 C。

> 注（外审 ds#36/gpt#57）：「做 A 不浪费一行」措辞过强，已弱化为「本 arc 的基础设施（scene 边界 / back-ref 接口 / serializer 持 chunk 树）C 全要；C 可能重定义 mutation 语义，但不会推翻这些地基」。

### §0.3 核心设计约束（编译边界 + 双源一致性 + 对象生命周期）

物理分包要成立，下列约束是**前置不变量**，非事后审计（外审 claude#2 / gpt#1 / ds 多条）：

- **C-1 双源一致性（single logical truth, dual physical representation）**：chunk 树（serializer 侧，根在 projectWriter.root）是**序列化权威源**；scene 值是**逻辑权威源**。**支持的 mutation 入口是 `Set*` 方法 + 结构性 op（New*/Delete*，亦经 attached 对象）**。**写序保证原子性（外审 gpt/ds/claude 同指）：先做可失败的 chunk patch，成功后才更新 scene 值** —— `if err := recv.back.X(v); err != nil { return err }; recv.field = v`。patch 失败则 scene 值不变、chunk 不变，无半改窗口。length-variable splice 同理（先构新子树并就位，成功后更新 scene），失败不 swap。`WriteAEP` 从 chunk 树发射；`WriteJSON` 从 scene 值导出；二者一致**仅因每个 mutation 双写**。**直接写 scene 导出字段（如 `layer.Visible = true`）绕过 chunk patch、不被序列化** —— **当前无编译期/lint 机械防护，靠 code-review 纪律**（与今天单包代码同性质：scene 字段本就 exported、Set* 双写、直写绕过；**B′ 不加重此风险，只是把它跨包暴露**）。lint 可行性留 plan，不在此承诺。
- **C-2 对象归属 + attach 时机（B′ 下大幅简化）**：每个 scene 对象的 `back` 在**构造期一次性注入**（Parse 解出对象时 / New* 建对象时），**不支持 late-attach / re-attach / detach**。故对象天然两类且**不混用**：①**attached**（Parse/New* 产出，全程经 back 双写）②**detached**（纯 scene 孤立构造，永不序列化）。无跨对象侧表 → 无「拿 A 文档的 Layer 调 B 文档」隐患。**消除外审 ds#6/gpt#50-63 的「detached 改值后再 attach 致 chunk 不同步」窗口**：因 attach 仅在构造期、不存在「先 detached 改值再 attach」路径。
- **C-3 Set\* 的 nil-back 语义（外审 claude#3）**：`back == nil`（detached 孤立对象）时，Set* 仅更新 scene 值并**返回 nil**（逻辑可用、不 panic）；不是编程错误。仅当 back 非 nil 但 patch 失败才返回 error（此时按 C-1 序：scene 值未改）。文档明写两分支。
- **C-4 Parse 原子性（外审 gpt#79）**：Parse 全程成功才返回 `*Project`（其下所有 back 已 attach）；任一步失败返回 `(nil, err)`，**绝不暴露半成品**。
- **C-5 并发（外审 ds/gpt + CLAUDE.md 既有约束）**：沿用 `incidents/concurrency-unsafe-shared-chunk-bytes.md` —— back 接口 impl 与 scene 值非线程安全，调用方自己锁；不承诺更多。
- **C-6 新增 scene 字段必须 eager materialize（外审 gpt#49）**：scene 零 rifx 靠「解析期把所有需要的值解进 scene 字段」。新增字段若按需读 chunk 会重引 rifx 依赖 → 禁止；新增字段一律解析期 eager 解出。§6 CI 边界断言兜底（scene import 集回归即红）。
- **C-7 scene 携带 write-back 协议 = 已知取舍（外审 gpt#11-18/#162-167 诚实记录）**：B′ 让 scene 持有 `XWriter` 接口字段 + exported `AttachWriter` plumbing，故 **scene 不是「纯模型层」——它携带 eager-patch 回写协议**；编译隔离（scene⊥rifx）成立，但 scene 的 mutation 语义**依赖 serializer 提供实现**（架构未完全独立）。这是**为「保 API + 最小 churn + 编译边界」付的价**（A 的替代是 scene 纯净但 doc-handle 病毒穿线 + Document god-object，churn 更大）。接口方法名（`SetVisible`/`SetLinearBlending`…）编码的是**当前 eager-patch 行为**；**C 阶段若改懒重生，这些 writer 接口很可能重构**（故见 §0.2：本 arc 是 C 的可复用基础设施，非「同一组接口必然延用」）。

---

## §1 包拓扑与依赖 DAG

五个包，箭头 = 允许 import，整体无环（DAG）：

```
internal/aep（薄 facade，顶层，下游入口）
  ├─imports→ internal/serializer   # Open/FromReader 委托 Parse
  └─imports→ internal/scene        # 类型别名再导出

internal/serializer（唯一同时见 scene + rifx 的包；实现 scene 的 writer 接口）
  ├─imports→ internal/scene        # 具体读 scene 字段 + 实现 XWriter 接口
  ├─imports→ internal/codec
  └─imports→ internal/rifx

internal/scene（纯运行时模型 + writer 接口定义）
  └─imports→ internal/codec        # 仅取值类型/枚举；零 rifx、零 serializer

internal/codec（纯值/字节 codec）
  └─（无 import：实测 codec_*.go 不 import rifx，纯 []byte/值）

internal/rifx（RIFX framing，不变）
  └─ 依赖：无
```

**编译期强制的不变量（取代 AST 守卫；§6 CI 断言全列校验）：**

- `scene` 直接 import 集 = {codec}。**编译器**保证 scene 永远碰不到 rifx/serializer —— 硬编译边界兑现。
- `codec` 直接 import 集 = ∅（rifx 无关；实测纠正了 brainstorm 初稿误标的 codec→rifx）。**约束：codec 的导出 API 不得暴露 rifx 类型**（只收发 `[]byte`/值对象）——这保证 scene（import codec）的代码里**永不出现 rifx 标识符**，使「scene⊥rifx 编译期硬」对 codec 未来演进也稳健（回应 ds#74/#104：codec 不返回 `*rifx.Chunk`，故 scene 无法经 codec 间接见 rifx 类型）。
- `serializer` 是唯一同时见 scene + rifx 的包。
- 反向边禁止：serializer 不依赖 aep、codec 不依赖 scene（§6 CI 全查，非只查 scene，回应外审 gpt#41-44）。

**文件搬迁（基于证据三的清单）：**

| 现 `internal/aep/` | 去向 |
|---|---|
| scene 类型定义 + 逻辑 accessor + 纯图级 mutation + **writer 接口定义** + `WriteJSON`（纯 scene 导出，回应 gpt#8-11） | `internal/scene` |
| 纯 codec（`codec_*`，实测零 rifx import） | `internal/codec` |
| `parse_*` / `lower_*` / `write_*` / `back_*`（10 backref struct 改实现 XWriter 接口）/ `mutate_*` 字节侧 + chunk 树 owner | `internal/serializer` |
| `Open`/`FromReader` 入口 + 类型别名 | `internal/aep` facade |

**API 形态（B′：不破）：**

- `aep.Open(path) (*Project, error)` / `aep.FromReader(io.ReadSeeker) (*Project, error)` —— 签名不变（现 API 即此，回应 ds io.Reader 质疑：现本就 ReadSeeker）。
- `proj.WriteAEP(w)` / `proj.WriteJSON(w)` —— 保留为 Project 方法。WriteAEP 体 = `p.back.WriteAEP(w)`（接口），serializer impl 读 scene + 写 root；WriteJSON 纯 scene 导出，留 scene 包。
- `layer.SetVisible(v)` / `prop.SetStaticValue(v)` 等 —— 全保留方法签名，体 = 改 scene 值 + `recv.back.X(...)`（C-1/C-3）。**~550 调用点零 churn**。

**§1 决策（默认拍，review 可推翻）：**

- **D1.1 保留 `internal/aep` facade**：下游单 import + `Open` 入口集中；facade = 类型别名（`type Project = scene.Project` 等）+ `Open`/`FromReader` 委托。**已知局限（外审三家）**：类型别名会让 godoc/import path 穿透到 `internal/scene`，**facade 不提供封装隔离**，仅提供 ergonomics + 单入口。鉴于无外部消费者（证据一），此局限可接受；硬边界目标是 scene⊥rifx 编译隔离，**非**对下游隐藏 scene。
- **D1.2 不引入 Document god-object**：B′ 下 chunk 树由 projectWriter（scene.Project 的 back impl）持有，无需独立 Document 容器统管侧表/registry（A 才需要）。

---

## §2 back-ref 接口（B′）+ Parse/Write 机制

### §2.1 writer 接口（scene 定义，serializer 实现）

scene 为每个持 back-ref 的类型定义一个 writer 接口；scene 对象持该接口（取代今天的 `back *xBackrefs` 具体字段）：

```go
// —— internal/scene ——
type Project struct {
    Compositions []*Composition
    Footages     []*Footage
    // … 纯逻辑导出字段（读 + 经 Set* 写，见 C-1）
    back ProjectWriter   // nil = 孤立构造（C-3）
}

// ProjectWriter 是 scene↔chunk 的唯一耦合点的逻辑契约；serializer 实现。
// 方法面 = Project 级 back-ref 操作（WriteAEP/项目级 Set*）。
type ProjectWriter interface {       // 方法必须 exported：serializer 跨包实现
    WriteAEP(w io.Writer) error
    SetLinearBlending(bool) error
    // … 项目级 length-preserving patch 操作（有界）
}

type Layer struct {
    Visible bool
    Name    string
    // …
    back LayerWriter
}
type LayerWriter interface {
    SetVisible(bool) error
    SetName(string) error
    // … layer 级 ldta patch 操作
}
// Property / Composition / Footage / Marker / Mask / Keyframe / AEPropertyGroup 同构。
```

10 个 `*Backrefs` struct（compositionBackrefs / layerBackrefs / propertyBackrefs / keyframeBackrefs / markerBackrefs / maskBackrefs / footageBackrefs / projectBackrefs / renderQueueBackrefs / propertyGroupBackrefs）迁入 serializer，**内容不变**（仍持 `*rifx.Chunk`，projectBackrefs 仍持 `root` —— 实测今天即如此），**改为实现对应 XWriter 接口**。

> **接口粒度**：每类一个 writer 接口，方法 = 该类 eager-patch 操作集。总方法数 ≈ back-ref setter 子集（~60–100，**非** 283——283 含大量纯图 setter，见 §3）。字节逻辑全在 serializer impl，**不泄进 scene**（scene 方法体只「改值 + 调接口」）。精确接口拆分（含 group property / effect param setter 归属）入 plan 首步分类表（回应 claude#1）。

### §2.2 attach 协议（plumbing 代价）

serializer 需把 impl 注入 scene 对象的 `back`。因跨包，需 scene 暴露注入点：

```go
// —— internal/scene —— exported plumbing（仅 internal/ 可见，非真公共 API）
func (p *Project) AttachWriter(w ProjectWriter) { p.back = w }
func (l *Layer)   AttachWriter(w LayerWriter)   { l.back = w }
// …
```

**代价（外审采纳，诚实记录）**：writer 接口 + `AttachWriter` 必须 exported（serializer 跨包实现/调用）→ scene 包 exported 面带 ~9 接口 + ~9 AttachWriter + ~60–100 接口方法的 plumbing 噪音。鉴于 `internal/`、无外部消费者，可接受；远小于 A 的 550 调用点 churn + Document god-object。

**AttachWriter 暴露面（外审 claude/gpt#36-49 采纳）**：Go 无 friend-package，exported `AttachWriter` 对**所有 internal/ 包（含 `cmd/`）可见** → `cmd` 理论上能 `layer.AttachWriter(nil)` 破坏 back。**接受此暴露为 out-of-contract**：这是「保方法 API + 跨包」的语言级必然代价（任何「方法留 scene、impl 在他包」的切法都需要它）。缓解仅契约层（doc 注明 AttachWriter 是 serializer-only plumbing，非消费者 API）；不重复 attach / 不 detach（C-2：构造期一次性）。若 plan 阶段认为暴露不可接受，退路是把构造也收进 facade 的 `Open`/`New*`（消费者永不直接见 AttachWriter），代价是 scene 无法独立构造 attached 对象——plan 决。

### §2.3 Parse 流程（原子，C-4）

`serializer.Parse(r io.ReadSeeker) (*scene.Project, error)`：解析 chunk 树 → 构造 `scene.*` 纯值对象（eager 解出所有字段，C-6）→ 为每对象建 `xBackrefs`（持 chunk）→ `obj.AttachWriter(xBackrefs)`。全程成功才返回 Project；任一步失败 `(nil, err)`，不暴露半成品（C-4）。构造序：先建被引用者（Footage）再建引用者（Composition/Layer），back attach 紧跟各对象构造（回应 ds#30 递归引用序）。`aep.Open` 薄包装之。

**attach 完整性断言（外审 gpt#143-153 采纳）**：Parse / New* 返回前跑一个 invariant 检查——遍历 scene 树，断言每个**应 attached** 的对象 `back != nil`（漏 attach 即 fail）。这补上 byte-identical（无-mutation，抓不到 attach 接线错）+ setter 单测（未必覆盖全部对象）的盲区。debug build / 测试常开。

### §2.4 Set\* + 从零建

- **Set\***（scene 方法，C-1 patch-first 原子序 / C-3 nil-back）：
  ```go
  func (l *Layer) SetVisible(v bool) error {
      if l.back == nil {            // detached 孤立对象：仅改值（C-3）
          l.Visible = v
          return nil
      }
      if err := l.back.SetVisible(v); err != nil { // 先做可失败的 ldta patch
          return err                                // 失败：scene 值不变，无半改窗口
      }
      l.Visible = v                                 // patch 成功后才同步 scene 值
      return nil
  }
  ```
- **length-variable**（name/comment/expression/text/keyframe insert）：serializer impl 在接口方法内 splice chunk 子树；`WriteAEP` 由既有 `rifx.Chunk.PayloadSize()` + 递归 `Write` 重算父 LIST size（机制不变，回应 ds/gpt「标记未定义」——**无独立标记机制**，size 重算是 Write 时全树重算，非脏标记扫描）。
- **从零建（New* / 结构性 mutate）**：走 serializer 的 lower 路径建 chunk 子树 + 建 xBackrefs + `obj.AttachWriter(...)`，使新对象与 Parse 对象共享同一 back attach 协议（回应 gpt#33-36「Parse 世界 / New 世界统一注册」）。**API 位置决策（回应 gpt#64-75/ds#36）**：建 chunk 的结构性构造器**住 serializer**（scene 不能 import rifx），**facade re-export 为消费者入口**（`aep.NewShapeLayer(...)` → `serializer.NewShapeLayer`），与 `aep.Open` 同层，消费者只见 facade。纯 scene 图构造（无 chunk、detached）可留 scene。结构性 op 仍遵 V2.1 atomic（warnings-as-failure + rollback），约束#6 ship-gate 不变。
- **结构性删除**（DeleteLayer 等）：移除 scene 对象即断 back 引用，serializer 侧 chunk 子树由 lower/splice 移除；无侧表/registry 需手动清（B′ 下随对象 GC，回应 gpt#78）。

---

## §3 Set\* 分类（哪些需 writer 接口方法）

证据三：~283 个 `Set*`、~550 调用点、getter 几乎全 eager（额外 churn 低）。**B′ 下方法签名全不变，churn 主要是「把 back-ref setter 的字节逻辑从方法体抽到 serializer impl」，调用点零改。**

**两类划分（plan 首步出精确表，回应 claude#1 / ds#21）：**

- **back-ref setter（需 writer 接口方法）**：经 `*rifx.Chunk` 做 length-preserving patch 者——Layer/Composition/Property/Footage/Marker/Mask/Project/RQ/OM 的 length-preserving setter。方法体迁为「改 scene 值 + 调接口」，字节逻辑入 serializer impl。
- **纯图 setter（留 scene，无接口）**：仅改 scene 值、序列化经 lower 重建（非原地 patch）者。**关键澄清（回应 ds#36 + SetStaticValue funnel）**：shape/stroke/fill 的 `RectNode.SetSize` 等**委托到 `Property.SetStaticValue`**——而 `Property.SetStaticValue` *是* back-ref setter（经 cdat patch）。故 shape setter **不**独立碰 chunk，但其底层 `Property.SetStaticValue` 经 Property 的 back 接口完成 patch。B′ 下这天然成立：`rect.SetSize(v)` → `prop.SetStaticValue(v)` → `prop.back.WriteStaticValue(v)`，全程方法、无 doc handle。**这正是 B′ 相对 A 的决定性优势**（A 会让 doc handle 病毒式穿透 fluent 形状 API）。
- **lazy flag 读**（IsSpatial/IsAnimated 等读 tdb4）：解析期预解进 scene bool 字段（C-6）。**外审 gpt#49/ds 提示语义变更风险**：求值时机从按需→eager。这些 flag 是 tdb4 纯 bit 读、无副作用、不依赖未解字段（已核 `scene_property_flags.go`），eager 化行为等价；plan 阶段逐个确认无条件依赖。

---

## §4 opaque shard —— 延后到 C（本 arc 不填）

证据二：opaque 基建已存在（`propertyBackrefs.opaque` / `layerBackrefs.opaque`）但当前 nil；partial-decode 散点 ~5–10 处。

**外审 ds#49/#51 + gpt 指出（采纳）**：本 arc 写路径是 verbatim+patch，**不消费** opaque shard → 本 arc 的 byte-identical round-trip **无法验证** opaque 收集是否正确（「round-trip 相同」对未消费的 opaque 是必要非充分）。在无法验证的情况下填 opaque = 投机性未验证工作。

**故本 arc 不填 opaque**；只确保架构留有位置（`xBackrefs` 仍带 opaque 字段，随 §2 迁入 serializer）。opaque 的正确收集 + 消费挪入 **C**（那时它被重生路径消费，且 round-trip 可验证其正确性）。

> A→C「必经之路」论证（§0.2）改由「scene 零 rifx 边界 + back-ref 接口 + serializer 持 chunk 树」三项地基支撑，**不依赖 opaque 填充**（回应 ds#36：去掉 opaque 填充后 on-ramp 论证仍成立）。

---

## §5 迁移分期（strangler，步步绿，全程保 byte-exact 门）

**核心排序约束（外审 claude#6 提请上移至此 / §0 已并入 C-1 周边）**：scene 结构体只要还持具体 `back *xBackrefs`（内含 `*rifx.Chunk`），就不可能成为零 rifx 的独立包。所以**先在单包内把 `back` 从具体类型改成接口（逻辑解耦），再把物理分包做成机械 `git mv`**。危险的是 P2（接口化解耦），P3（分包）是机械收割。

- **P0 基线**：`tmp_debug/` 建分包 round-trip 基线脚本（全 82 fixture 跑 `Open→Write→bytes` 存指纹 + 记 fixture git hash）。沿用方案① §8 套路。**exit**：基线就位。
- **P1 抽 `internal/codec`**：`codec_*` 纯叶平移成真包（实测零 rifx import）。**facade re-alias 任何 exported codec 符号**（如 `Gradient`/`GradientColorStop`，它们在公共 API）以保零-diff（回应 ds#73）。**exit**：`go build ./...` + 82-fixture byte-identical 绿 + `go doc -all` facade 零-diff。
- **P2 单包内 back-ref 接口化（最危险，byte-exact 门 + setter 单测当安全网）**：**仍在单包 `internal/aep`**。定义 9 个 XWriter 接口 + AttachWriter；把 10 个 `xBackrefs` 改为实现接口；scene 类型 `back` 字段由 `*xBackrefs` 改 `XWriter` 接口；back-ref setter 方法体改「改值 + 调接口」（每改一组，连同其内部调用点同 commit，`go test` 绿——回应 ds#60/gpt「调用点同步」）。**末态**：scene 结构体只引用 XWriter 接口、零 `*rifx.Chunk`/零具体 backref（grep `\.back\.` 仅见接口调用）。**安全网**：byte-identical（verbatim 路径）+ **既有 setter 单测**（每个 Set* 有断言，覆盖「改值 + patch」正确性——byte-identical 抓不到 setter 逻辑，回应 gpt#54）。**abort 条件 + 退路**（回应 gpt#62/claude/ds#61）：若某类 back-ref 接口化后 setter 单测无法保绿且非调用点同步问题 → 暂停该类、记录原因。**退路（不回退整个 P2）**：该类**暂留具体 `back *xBackrefs`**（单包内不动），其余类照常接口化；P3 分包时，未接口化的类**阻塞 scene 包独立**（因仍持 rifx）→ 要么把该类的*类型定义*暂留 serializer 侧（scene 缺该类，能力受限），要么对该类用局部化的 handle（窄范围 A）。10 类里 ≤2-3 类卡住时取前者（缺类）或混合；plan 评估哪些类高风险并先验证。B′ 是 concrete→interface 的机械替换，预期全类适用，abort 仅为安全阀。
- **P3 物理分包（机械收割）**：scene 已只引用接口 → `git mv` `scene_*`（含 XWriter 接口定义 + WriteJSON）进 `internal/scene`、serializer 侧（parse_/lower_/write_/back_/mutate_ + xBackrefs impl）进 `internal/serializer`；建 `internal/aep` facade。**编译边界此刻由编译器强制**。**P3 非纯零风险（回应 gpt#62-68 / claude#5）**：跨包后 export 可见性变化、init/global-var 初始化顺序、测试辅助代码失效需逐一处理；故 P3 exit 含显式核实。**exit**：各包独立编译 + **全 DAG CI 断言**（scene 不直接 import rifx/serializer、serializer 不 import aep、codec 不 import scene——回应 gpt#41-44）+ byte-identical 全绿 + `go doc -all` facade 公共面 diff（B′ 下应**仅** codec 别名 + plumbing 接口增量，核心 R/W 方法零变）。
- **P4 下游切换 + 收口**：`cmd/aepdemo` 验证编译（B′ 下 API 不变，预期零改或极小）；删/改 `arch_boundary_test.go`（scene⊥rifx 已编译期保证，AST 守卫降级为 serializer 包内命名轴 lint）；CLAUDE.md 硬约束#3 更新（单包 → 多包 + 新 DAG + B′ 接口破环）；终验 **AE 2020 + 2025 双版本 ship-gate 全套**（约束#6，自验留痕）；cockpit/INDEX 落地。

> 阶段序由风险驱动：codec（独立纯叶）→ **P2 单包内接口化（真正的活）** → P3 机械分包 → P4 收口。回归门（P0）必须在 P2 之前就位。**回滚（回应 gpt#62/ds#84）**：每阶段独立 commit，失败 `git revert` 回上一绿态；P2 内每组 setter 独立 commit，单组失败不污染其余。

---

## §6 验证 / 测试 / ship-gate / 边界强制

每阶段（每搬/改一组）必须：

1. `go build ./...` 绿。
2. `go vet ./... && go test -count=1 ./...` 全绿（含 ship-gate 非门控部分 + **既有 setter 单测**——P2 的主安全网）。
3. **byte-identical round-trip**（约束#1/#5）：对 82 fixture 跑 **无 mutation 的 `Open→Write→bytes`**，断言与 P0 基线指纹相同。**澄清（回应 ds#71/#48）**：此门是「解析-再发射」不变性证明，**非**「SetName 后字节不变」（length-variable 操作字节本就变；那条由 setter 单测覆盖）。
4. **公共 API diff**（约束#2）：`go doc -all` facade 前后对比。**P1 零-diff**（codec 抽包 + facade 别名）；**P2/P3 diff 应仅 = plumbing 接口/AttachWriter 增量**（B′ 不破核心 R/W 方法签名），逐条核对无意外。
5. **全 DAG 边界 CI 断言**（把硬边界做成可执行）：`go list -deps`/AST 检查——`internal/scene` 直接 import 集 ⊆ {codec}（不含 rifx/serializer）、`internal/serializer` 不 import `internal/aep`、`internal/codec` 不 import `internal/scene`。回应 gpt#41-44「DAG 不变量验证不完整」。注：这是 CI 断言（可被改脚本绕过），但**真正的硬保证是 Go 编译器**——scene 一旦误 import rifx 直接编译失败。
6. **终验 AE 双版本 ship-gate**（约束#6）：P4 跑全套 `AE_SHIP_GATE=1`（AE 2020 + 2025），确认纯结构搬迁零回归。无结构性写路径语义变更，故中途不必逐阶段跑（回应 ds#52：中途 byte-identical + setter 单测已是廉价回归网，AE 门贵，收尾跑）。

---

## §7 非目标 / 范围

- **不破核心 R/W API**：B′ 保留 `Open→*Project`/`WriteAEP`/`Set*` 全签名（plumbing 接口除外，且仅 internal/）。
- **不改契约**：byte-exact round-trip 仍硬不变量；不切懒重生（C 日后独立 arc）。
- **不改写语义**：eager length-preserving patch 逻辑原样迁入 serializer impl，不重写；length-variable 路径不动。
- **不填 opaque shard**（§4，C 的活）；只保留架构位置。
- **零行为变更**：纯接口化 + 搬迁 + facade；唯一预期 API 增量是 plumbing 接口（internal/）。
- **不扩** capability matrix / shape graph / effect schema（V3 其他 M）；不顺手重构无关逻辑。

---

## §8 风险

| 风险 | 缓解 |
|---|---|
| 双源一致性漂移（scene 值 vs chunk）：直接写 scene 字段绕过 patch | C-1 契约：唯一 mutation 入口是 Set*；scene 字段语义=读+经 Set* 写；plan 评估能否加 lint 防直写 |
| back-ref 接口化引入隐性行为变更 | P0 byte-identical 基线在 P2 之前就位 + 既有 setter 单测（抓 setter 逻辑）；每组独立 commit |
| Set* 分类误判（纯图 vs back-ref） | P2 首步精确 grep 分类（`\.back\.`/接口调用）入 plan 评审表；shape→Property.SetStaticValue 链已澄清（§3） |
| writer 接口面/plumbing 污染 scene exported 面 | 接受（internal/、无外部消费者）；远小于 A 的 churn；接口粒度 plan 细化 |
| P3 跨包 export/init 顺序/测试辅助失效 | P3 exit 显式核实 init/global-var + 测试辅助迁移（不假设零风险，回应 gpt/claude） |
| lazy flag eager 化语义变更 | 已核 tdb4 flag 无副作用/无条件依赖；plan 逐个确认 |
| 内存：Parse 后 scene + chunk 树 + backref 并存 | B′ 无侧表/registry（比 A 省）；今天 scene 已持 back，增量仅接口间接层，marginal |
| AE 双版本 ship-gate 回归 | §6 P4 终验全套（约束#6，自验留痕） |
| 范围蔓延到 C | §0.1 抉择4 + §4 opaque 延后硬边界 |

---

## §9 brainstorm 期证据表（A 先行 + B′ 依据）

| 调查 | 结论 | 关键出处 |
|---|---|---|
| byte-exact 是否产品契约 | **内部回归工具**，非产品承诺；README/CLAUDE 只诺 length-preserving + opaque + AE 可重开；无外部消费者（仅 cmd/aepdemo 写新文件、cmd/docgen 读 Go 源） | README:4/37,CLAUDE:24/32,cmd/* |
| C 重生脆弱面 | **可控 ~5–10 散点**（dimsep `@0x08/@0x10`、tdum/tduM、ldta padding、keyframe `@0x07`、cdta 尾、btds/btdk、effect sub-chunk）；opaque 基建已存在但 nil | 各 incident,back_*.go |
| back-ref/Set* churn | ~283 Set*、~550 调用点（~75 非测试+~475 测试）、10 backref struct、getter 全 eager；**B′ 下调用点零 churn（方法签名不变）** | back_*.go,全仓 grep |
| 项目事实校验（外审分诊） | codec_*.go **不 import rifx**；`Open(path)`+`FromReader(io.ReadSeeker)` 现状；shape setter 经 `Property.SetStaticValue`（back-ref） | grep 实证 |

---

## §10 外审 disposition（tmp/{ds,gpt,claude}.txt，三轮）

> 沿用项目惯例（reorg-spec §13 外审 disposition）。**外审不了解项目全貌，已用项目事实校验**。

**采纳（已改入本版）：**
- **B′ 取代 A**（侧表 → back-ref 接口）：从根消解 ds#10/#12（侧表失效/指针 map key）、gpt#12-20（Document 角色过载）、gpt#53-56（opaque+backref 指针 identity 耦合）、gpt#67-78（`Open→doc.Scene` 读路径剧变）、ownership/跨 Document 误用、claude#2（指针 identity 阻断）→ §0.1/§2。
- **双源一致性不变量** → C-1（gpt#1/#21-25、ds 双真相）。
- **nil-back 错误语义** → C-3（claude#3）。
- **Parse 原子性** → C-4（gpt#79-82）。
- **新增字段 eager materialize** → C-6（gpt#49）。
- **WriteJSON 留 scene**（纯 scene 导出，不污染 serializer）→ §1 搬迁表（gpt#8-11）。
- **opaque 延后到 C**（本 arc 不消费→无法验证→不填）→ §4（ds#49/#51、gpt）。
- **P2 安全网 = setter 单测**（byte-identical 抓不到 setter 逻辑）→ §5 P2/§6（gpt#54）。
- **全 DAG CI 断言**（非只查 scene）→ §6（gpt#41-44）。
- **P3 非零风险**（export/init/测试辅助）→ §5 P3（gpt#62-68、claude#5）。
- **byte-identical 澄清**（无 mutation 的解析-发射不变性）→ §6（ds#71/#48）。
- **P1 零-diff 需 facade re-alias codec 公共符号** → §5 P1（ds#73）。
- **abort/回滚条件** → §5（gpt#62、ds#84）。
- **codec 不 import rifx**（实测纠正初稿）+ **自审修正上移 §0/§5** → §1/§5（gpt#3-7、ds#45/#66、claude#6）。
- **lazy flag eager 化语义风险** → §3/§8（gpt#49、ds#26）。
- **「做 A 不浪费一行」措辞过强** → §0.2 弱化（ds#36、gpt#57-61）。
- **status idea→active**（已是 Active focus）+ **P5→P4 编号统一** → frontmatter/§5（ds#64/#70/#97）。
- **facade 别名穿透 = 不提供封装**（仅 ergonomics）→ D1.1 显式承认（三家）。

**驳回（项目事实不成立）：**
- ds「Open 应支持 io.Reader、ReadSeeker 是 breaking」→ 现 API 本就 `FromReader(io.ReadSeeker)`，无回退问题。
- ds#66「`go list -deps scene` 含 rifx 必失败」→ codec 实测不 import rifx，scene 传递依赖不含 rifx，断言成立。
- 早轮 ds#94「§1 无 D1.1/D1.2 标签」→ 标签已在 §1（误读）。

**留给 plan（非 spec 缺陷）：**
- Set* 精确分类规则（group property / effect param setter 归属）→ plan 首步出分类表（claude#1/ds#21）。
- writer 接口精确粒度/方法清单 → plan（§2.1）。
- fixture rebaseline 责任/触发 → plan（ds#56）。
- 内存峰值实测 → plan（gpt#90，B′ 下增量更小）。
- 直写 scene 字段的 lint 防护可行性 → plan（C-1）。

**第三轮 disposition（精度/诚实加固，无新架构异议）：**
- 采纳：Set* **patch-first 原子序**（gpt#116-134/ds#34/claude#13）→ C-1 + §2.4 示例改写；**C-1 无机械防护、靠 code-review**（同今天单包性质，B′ 不加重）显式声明（gpt#3-10/claude#13）；**attach 仅构造期、消除 detached-改值-再 attach 窗口**（ds#6/gpt#50-63）→ C-2/C-3；**C-7 scene 携带 write-back 协议=已知取舍**（gpt#11-18/#162-167）；**「必经之路」弱化为「可复用基础设施」**（gpt#178-189）；**接口方法须 exported**（ds#28，修 `setLinearBlending`→`SetLinearBlending`）；**AttachWriter 对 cmd/ 暴露=out-of-contract + facade 收口退路**（claude/gpt#36-49）→ §2.2；**attach 完整性断言**（gpt#143-153）→ §2.3；**New* 位置=serializer+facade re-export**（gpt#64-75/ds#36）→ §2.4；**codec 不暴露 rifx 类型**（ds#74/#104）→ §1；**P2 abort 退路**（claude/ds#61）→ §5。
- 驳回：ds#50「§8 列 opaque 收集遗漏风险」→ §8 无此行（B′ 已删，opaque 延后到 C）；ds#112「status active 与未写 plan 矛盾」→ flightdeck 语义 active=在办，active spec 待 plan 是常态；ds「未附外审原文」→ 原文在 `tmp/{ds,gpt,claude}.txt`，§10 逐条标号引用。
- 留 plan：writer 接口精确方法计数与稳定性（gpt#21-26）；接口化后测试改 mock/类型断言的影响（gpt#86-97）；P3 init/global-var 顺序的具体核实法（claude#8）。

> **收敛判定**：三轮外审已从「架构异议」（轮1-2 驱动 A→B′ pivot）收敛到「精度/诚实措辞」（轮3）。剩余 open 项均为 plan 级实现细节或已显式记录的 B′ 取舍（C-7 / AttachWriter 暴露 / C-1 无机械防护），非 spec 设计缺陷。判定 spec 设计层已稳，宜进 writing-plans。

---

## §11 决策记录

- 2026-06-07 brainstorm：用户选 V3 M8 方案② 为下一焦点。四抉择初定：①动机=硬编译边界 ②路线=自由函数·破 API ③back-ref=侧表（A）④范围=A 先行、C 日后。
- 2026-06-07 三组并行证据 → 确立「本 arc 为 C 提供可复用基础设施」（非强「必经之路」），否决 big-bang C。
- 2026-06-07 三家外审第 1-2 轮 → **推翻 ②③**：②破 API→**不破 API**（接口倒置定价纠错）；③侧表→**B′ back-ref 接口**（消解侧表/指针 identity/god-object 一整类外审）。①④不变。
- 2026-06-07 三家外审第 3 轮 → 精度/诚实加固（patch-first 原子序、C-1 无机械防护显式化、attach 仅构造期、C-7 取舍记录、AttachWriter 暴露面、attach 完整性断言、New* 位置、codec 不暴露 rifx），**无新架构异议**；判定收敛，宜进 plan。§10 disposition 逐条。

## 关联文档

- `2026-05-22-v3-direction.md` — M1–M8 框架 + scene→rifx 残留表（M8 前置解耦已清零）
- `landed/specs/2026-05-30-aep-package-reorg-design.md` — 方案① §0 Go 语义墙 + §6 back_ 缝固化（本 spec 直接前置）
- `2026-05-27-v3-deep-think.md` — Q1 软迁移 / Q5 opaque shard / Q7 eager+lazy 共存
- `CLAUDE.md` § 硬约束 #1/#2/#3/#5/#6 · `incidents/concurrency-unsafe-shared-chunk-bytes.md`（C-5）
- `incidents/separate-dimensions-write-mechanics.md` — `@0x08/@0x10` 缓存字段（partial-decode 样例）
