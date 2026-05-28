# V2.1 Foundation — 从 0 创建 AE 项目 + 空 Composition

> **Status**: design approved, ready for implementation plan
> **Phase**: V2.1（V2 总目标 5 个 sub-project 的第一个）
> **Date**: 2026-05-22

## Context

aep-parser 当前能力 = 读 `.aep` + length-preserving 原地修改字段。V2 阶段扩展为**结构性写**：从零创建新对象。整个 V2 分解成 6 个独立 sub-project：

```
V2.1  Foundation               ← 本文档
      aep.NewProject + Composition.NewComposition + AE 25 ship gate
V2.2  系统型 Layer 创建        (Solid / Camera / Light)
V2.3  Footage import + 引用型 Layer
V2.4  属性树结构性扩展         (AddEffect / AddMask / NewKeyframe-on-empty)
V2.5  Text + Shape Layer       (btds + btdk + shape primitive 合成)
V2.6  Marker / Folder / 收尾
```

V2.1 是地基，其它 5 个 sub-project 全依赖它（spec 与 plan 单独再开）。

## Goals + Scope

### 目标

用户能 `aep.NewProject()` 拿到一个全新的 empty Project，调 `proj.NewComposition(name, w, h, fps, duration)` 加任意数量空 composition，`WriteAEP` 出来的 `.aep` 用 AE 25 打开正常 — 看到所有 comp，名字 / 尺寸 / 帧率 / 时长 / 默认设置全对。

### In Scope

- ✅ `aep.NewProject(target ...AETarget) *Project` — 嵌入 AE 2020 / 2022 / 2025 三份 saved-empty 作 base，用户按 target 选；不传 = 默认 AE 2020（最大兼容）
- ✅ `proj.NewComposition(name, w, h, fps, duration) (*Composition, error)`
- ✅ 自动分配 `Composition.ID`（`Project.nextItemID` 计数器）
- ✅ 可选 comp 字段（BGColor / PixelAspect / ResolutionFactor / Shutter*）走既有 `Set*` 方法
- ✅ `WriteAEP` 序列化新 comp（既有代码会重算父 LIST size，无需改）
- ✅ AE 2020 + AE 2025 ship gate（半自动测试，需各自版本 AE 实开）

### Out of Scope（留给 V2.2+）

- ❌ Layer 创建（V2.2）
- ❌ Footage import / Folder 创建（V2.3 / V2.6）
- ❌ Marker / Effect / Keyframe / Text / Shape Layer 合成（V2.4-V2.6）
- ❌ Composition 删除 / 复制（V2 收尾时再加）
- ❌ ID 手动指定（auto-only；将来确有需求再加 `aep.WithID(id)` option）
- ❌ AE 2021 / 2023 / 2024 等中间版本 template（用户手头只有 2020/2022/2025；将来按需补；预测 AE 24 等 24+ 系列跟 AE 2022 同形，只差 svap 戳）

### Ship Gate

1. `go vet ./...` clean + 所有既有测试 PASS（不破坏）
2. 新增 Go unit tests pass（详 §Testing）
3. **AE 25 真实打开 New 出来的 `.aep`**：用 `verify_v2_1.jsx` 检查 `app.project.items.length` + name / width / height / frameRate / duration 全对，写 PASS 到 `.done` marker

## Architecture

### 数据流

```
internal/aep/templates/2020.aep    ← AE 2020 saved empty (~8 KB, 24 chunks)
internal/aep/templates/2022.aep    ← AE 2022 saved empty (~8 KB, 30 chunks)
internal/aep/templates/2025.aep    ← AE 2025 saved empty (~8 KB, 30 chunks)
        │
        │ //go:embed
        ↓
aep.NewProject(target) ──► 按 target 选 template
                          ──► FromReader(bytes.NewReader(embeddedTemplate[target])) ──► *Project
        │
        │ proj.NewComposition(name, w, h, fps, duration)
        │   1. validate inputs (return error if invalid)
        │   2. allocItemID() → uint32 (next available)
        │   3. build chunks: idta / Utf8 / cdta / 空 LIST Layr
        │   4. wrap into LIST formType=Item
        │   5. append to root Fold.Children (after existing fdta)
        │   6. parseComposition(newItemList) → *Composition (复用既有 parser)
        │   7. comp.proj = p; append to p.Compositions
        │   8. return *Composition
        ↓
proj.WriteAEP(io.Writer)   ← 既有 rifx 序列化自动重算父 LIST size，无需改
```

### 关键设计选择

1. **3 份独立 template embed**（AE 2020 / 2022 / 2025，各 ~8 KB）。仓内静态 binary 总计 ~24 KB，零 RE 成本。每个 template 含 24-30 个 chunks（项目 header / 偏好 / GPU / 渲染队列 / 工作区 etc.）—— 不预知哪些是 AE 接受文件的必需 chunk，全保留最稳妥。**不优化"共享结构 + 替 svap"**：2022 与 2025 chunk 集相同但 svap 不同（实测），理论可共享 28 个 chunk 只换 svap；但复杂度收益不抵 ~8 KB 重复 binary。

2. **每次 NewProject 重 parse 选中的模板**。简单粗暴：`FromReader(bytes.NewReader(embeddedTemplates[target]))`。8 KB parse 成本 < 1 ms，避免手写 deepClone 又少一份 bug。

3. **新 comp 走 parser 闭环**：构造 chunks 后**不**自己 populate Go struct，而是塞进 RIFX 树后让 `parseComposition` 重 parse 出 `*Composition`。**保证 New 出来的 comp 跟 Open 出来的 100% 同形** —— 同样的 `*rifx.Chunk` 引用、同样的字段、所有 `Set*` 方法立即可用，零分支代码。

4. **ID 分配集中**：`Project.nextItemID uint32`（unexported），`parseProject` 完成后初始化为 `max(已有 ID) + 1`。`NewComposition` 取并递增。混用 `Open + NewComposition` 时 ID 自然不冲突。

5. **新 comp 落点**：append 到 root Fold LIST 的 Children 末尾。Fold 已有 1 个 `fdta` 子 chunk（folder data header）—— 保留在前，新 Item LIST append 在后。AE 不要求 item 严格顺序（实测）。

6. **WriteAEP 不改**：现有 rifx 序列化已经验证会 walk children 递归重算 size header（既有 `WriteAEP` 实现 + 全套 100+ test 实证），append 新 child 自动生效。架构最便宜的地方。

7. **Folder 落点 = root-only**：V2.1 所有 `NewComposition` 都创建在 root Fold LIST。Nested folder（comp 在子 folder 下）留待 V2.6 — 那时 `Project.rootFold` cache 之外再加 `Folder.foldChunk` 引用支持任意 nesting。

## Public API

```go
// AETarget selects which AE version's empty-project skeleton NewProject
// uses as the base. Output .aep files will identify themselves as that
// AE version's format (via the `svap` chunk), so newer AE versions may
// silently upgrade them, while older AE versions refuse to open files
// claiming a newer version.
//
// Underlying values are AE marketing years (2020/2022/2025/…) for
// debuggable panic messages and natural `target >= 2025` comparisons.
//
// AETarget values are NOT forward-compatible — unknown values panic.
// Upgrade the library when targeting a newer AE version.
type AETarget int

const (
    TargetAE2020 AETarget = 2020  // default; max compatibility (any AE 2020+ opens)
    TargetAE2022 AETarget = 2022  // 30 chunks; adds AE 24+ color-mgmt prefs (pcms/PwCs/pdvc)
    TargetAE2025 AETarget = 2025  // 30 chunks; latest tested
)

// NewProject returns a fresh empty Project parsed from the embedded
// AE skeleton matching the requested target.
//
// Optional target arg: zero args = TargetAE2020 (max compatibility). Pass
// at most one target. Subsequent NewComposition calls populate it.
//
// Never returns an error: the embedded templates are build-time trusted;
// parser bugs panic with a "build bug" message (not user-facing).
// Panics on: multiple target args, or unknown AETarget value (forward-incompat).
func NewProject(target ...AETarget) *Project

// NewComposition adds an empty composition to the project's root folder.
// Returns the new *Composition (Layers/Effects/Markers all nil) or an
// error if any required input is invalid.
//
// Required:
//   name        — non-empty string
//   width/height — > 0 (uint16; AE max 30000)
//   frameRate   — > 0 (Hz; 29.97 etc.; whole+frac/65536 encoding handled internally)
//   duration    — > 0 (seconds; converted to whole frames via fps internally)
//
// Optional fields default to AE-typical:
//   BGColor          {0,0,0}
//   PixelAspect      1.0
//   ResolutionFactor {1,1}
//   ShutterAngle     180
//   ShutterPhase     0
//   MotionBlurAdaptiveSampleLimit  128
//   MotionBlurSamplesPerFrame      16
//   WorkAreaStart/End              0 / duration
//   DisplayStartTime               0
//
// Use existing Set* methods to override:
//   comp.SetBGColor([3]uint8{20,30,40})
//   comp.SetPixelAspect(2.0)
//   comp.SetResolutionFactor(2, 2)
//
// Composition.ID is auto-assigned (Project.nextItemID++).
// New comps append to the project's root folder.
func (p *Project) NewComposition(
    name string,
    width, height uint16,
    frameRate, duration float64,
) (*Composition, error)
```

### 用法示例

```go
// 默认 AE 2020 兼容（任何 AE 2020+ 可开）
proj := aep.NewProject()

// 或显式指定目标版本
proj25 := aep.NewProject(aep.TargetAE2025)

main, err := proj.NewComposition("Main", 1920, 1080, 29.97, 10)
if err != nil { log.Fatal(err) }
main.SetBGColor([3]uint8{20, 30, 40})

bg, err := proj.NewComposition("BG_loop", 1920, 1080, 30, 5)
if err != nil { log.Fatal(err) }
bg.SetResolutionFactor(2, 2)

out, _ := os.Create("out.aep")
if err := proj.WriteAEP(out); err != nil { log.Fatal(err) }
// AE 2020 / 2022 / 2025 全可打开 → 看到 "Main" + "BG_loop" 两个空 comp
```

### 既有 API 影响

V2.1 **不改**任何既有公开类型 / 方法 / JSON 输出：

- `aep.Open / FromReader` 行为不变
- `*Project / *Composition / *Layer / *Property` 字段不变
- 所有 `Set*` 方法签名不变
- `WriteAEP / WriteJSON` 实现不变（自动支持新加 comp）

唯一新增：`Project.nextItemID uint32` 非导出字段（不影响 JSON 输出）。

## Internal Design

### 文件组织

```
internal/aep/
  new_project.go              ── aep.NewProject(target) + AETarget enum + 嵌入 3 份 template
  new_composition.go          ── Project.NewComposition() + chunk builders
  new_project_test.go         ── NewProject Go unit（每个 target 各一组）
  new_composition_test.go     ── NewComposition unit + roundtrip + 拒绝路径
  templates/
    2020.aep                  ── //go:embed AE 2020 saved empty (~8 KB, 24 chunks)
    2022.aep                  ── //go:embed AE 2022 saved empty (~8 KB, 30 chunks)
    2025.aep                  ── //go:embed AE 2025 saved empty (~8 KB, 30 chunks)
```

### Project.nextItemID + Project.rootFold

```go
// types_core.go 加 2 个非导出字段
type Project struct {
    // ... existing ...
    nextItemID uint32      // 下一个可分配的 Item ID
    rootFold   *rifx.Chunk // 根 Fold LIST 的引用，parse 时 cache，O(1) 查找
}

// parse.go: parseProject 收尾扫一遍 + cache rootFold
func (p *Project) initDerived(rifxRoot *rifx.Chunk) {
    // ID counter
    var max uint32
    for _, c := range p.Compositions { if c.ID > max { max = c.ID } }
    for _, f := range p.Footage      { if f.ID > max { max = f.ID } }
    for _, fo := range p.Folders     { if fo.ID > max { max = fo.ID } }
    p.nextItemID = max + 1

    // rootFold cache — 找根 Egg! 下第一个 formType=Fold 的 LIST
    for _, c := range rifxRoot.Children {
        if c.IsList() && c.FormType == rifx.IDFold {
            p.rootFold = c
            break
        }
    }
}

func (p *Project) allocItemID() uint32 {
    id := p.nextItemID
    p.nextItemID++
    return id
}
```

`p.rootFold` 在 V2.1 只被 `NewComposition` 用；V2.x（AddFootage / AddFolder）共享同一 cache。V2.6 加 Folder 时，**不**走 rootFold cache —— nested item 走具体 folder 的 Fold LIST，到时分开处理。

### Chunk builders（unexported）

**重要前置**：抽 `internal/aep/cdta_layout.go` 统一 parser / writer / builder 三方共享 offsets：

```go
// cdta_layout.go - cdta / idta 字段偏移单一权威源
const (
    // cdta layout (204 bytes)
    cdtaResolutionFactorX = 0x00
    cdtaResolutionFactorY = 0x02
    cdtaTickRate          = 0x08
    cdtaWorkAreaStart     = 0x1C  // dividend
    cdtaWorkAreaStartDiv  = 0x20
    cdtaWorkAreaEnd       = 0x24  // dividend; 0xFFFFFFFF = sentinel "use duration"
    cdtaWorkAreaEndDiv    = 0x28
    cdtaBGColor           = 0x34  // R / G / B（3 字节）
    cdtaFlagsByte8A       = 0x8A  // Draft3D bit 0
    cdtaFlagsByte8B       = 0x8B  // HideShyLayers / CompMotionBlur / ...
    cdtaWidth             = 0x8C
    cdtaHeight            = 0x8E
    cdtaPixelAspectNum    = 0x90
    cdtaPixelAspectDen    = 0x94
    cdtaFrameRateWhole    = 0x9C
    cdtaFrameRateFrac     = 0x9E
    cdtaDisplayStartTime  = 0xA4
    cdtaDisplayStartDiv   = 0xA8
    cdtaShutterAngle      = 0xAE
    cdtaDuration          = 0xB0
    cdtaShutterPhase      = 0xB4
    cdtaMotionBlurAdaptive = 0xC4
    cdtaMotionBlurSamples  = 0xC8
    cdtaSize              = 0xCC  // 总长

    // idta layout (84 bytes) - 部分 RE，其余字段在 RE-3 确认前不要相信
    idtaTypeCode = 0x00  // uint16 BE; 0x04 = comp (RE 已确认)
    idtaItemID   = 0x14  // uint32 BE; tentative: RE-3 pending —— 此偏移来自 existing fixture dump 观察，未 byte-level 确认是 ID 还是其它含义。RE-3 完成前 builder 实现不应 hardcode 该常量到 spec API，应留在 plan 中决定
    idtaLabel    = 0x3A  // uint8; tentative: RE-3 pending
    idtaSize     = 84
    // 其它 75 字节：RE-3 在 plan day 1 dump既有 fixture 对照填默认值
)
```

`write_composition.go` 现在散落的 hex 数字（`@0x8C` / `@0xC4` etc.）迁移成常量引用。parser / builder 同样引这些常量。**单一权威源** = 改 layout 改一处。

然后 builders：

```go
// new_composition.go

// buildCompIdta 构造 84-byte idta（comp item 元数据）。
// 详细字段 RE 状态见 plan RE-3。
func buildCompIdta(itemID uint32) []byte

// buildCompIdpc 构造 idpc chunk（per-item GUID）。
// 实现策略由 plan RE-2 决定（固定常量 vs crypto/rand）。
func buildCompIdpc() []byte

// buildCompIide 构造 4-byte iide chunk（item index entry header；
// 来自 template 内既有 fixture dump 的默认值）。
func buildCompIide() []byte

// buildCompCdta 构造 204-byte cdta，5 个必填 + AE 默认值。
// 用 cdta_layout.go 的常量；不直接写 hex offsets。
func buildCompCdta(w, h uint16, fps, duration float64) []byte

// buildEmptyLayrList 构造空 LIST formType=Layr（0 children）。
func buildEmptyLayrList() *rifx.Chunk

// buildCompItem 包成完整 LIST formType=Item。
// Children 顺序（按 AE 空 fixture dump 实测）：
//   iide / idpc / idta / Utf8 / LIST(dats) / cdta / cdrp / comr / 空 LIST Layr
//
// 各 child 来源策略：
//   • iide               builder 生成（4 字节固定值，从既有 fixture dump）
//   • idpc               builder 生成（RE-2 决定：crypto/rand v4 UUID）
//   • idta               builder 生成（含 type/ID/label，RE-3 完整 layout 后填默认值）
//   • Utf8               builder 生成（comp name UTF-8 字符串）
//   • LIST(dats)         template 复制（既有 fixture 内 dats 含 numS chunk，默认值固定）
//   • cdta               builder 生成（5 必填 + AE 默认值，用 cdta_layout.go 常量）
//   • cdrp / comr        template 复制（1 字节 enum，AE 25 默认 0；语义未 RE）
//   • LIST(Layr)         builder 生成（空 children，准备 V2.2 加 layer）
//
// "template 复制" = 从 templates/2025.aep（或 plan 决定的 fixture）的 dummy comp Item LIST
// 读出对应 child 的原始字节，作为 builder 的固定默认资源。这条策略保证：我们未 RE 出语
// 义但 AE 接受的 chunk 不被遗漏 / 错填。
func buildCompItem(itemID uint32, name string, cdta []byte) *rifx.Chunk
```

### NewComposition 主流程

```go
func (p *Project) NewComposition(
    name string, w, h uint16, fps, duration float64,
) (*Composition, error) {
    // 1. 校验
    if name == "" {
        return nil, fmt.Errorf("composition name cannot be empty")
    }
    if w == 0 || h == 0 {
        return nil, fmt.Errorf("composition size must be > 0 (got %dx%d)", w, h)
    }
    if fps <= 0 {
        return nil, fmt.Errorf("frame rate must be > 0 (got %g)", fps)
    }
    if duration <= 0 {
        return nil, fmt.Errorf("duration must be > 0 (got %g)", duration)
    }

    // 2. 分配 ID
    id := p.allocItemID()

    // 3. 合成 chunks
    cdtaBytes := buildCompCdta(w, h, fps, duration)
    itemList := buildCompItem(id, name, cdtaBytes)

    // 4. append 到 cached root Fold LIST
    if p.rootFold == nil {
        return nil, fmt.Errorf("internal: project missing root Fold (template malformed?)")
    }
    p.rootFold.Children = append(p.rootFold.Children, itemList)

    // 5. 复用 parseComposition
    comp, err := parseComposition(itemList, &p.Warnings)
    if err != nil {
        return nil, fmt.Errorf("internal: re-parsing new composition: %v", err)
    }
    comp.proj = p

    // 6. 注册
    p.Compositions = append(p.Compositions, comp)

    return comp, nil
}
```

第 5 步重 parse 是关键 —— 保证 New 出来的 Composition 跟 Open 出来的同形。

### NewProject 实现

```go
//go:embed templates/2020.aep
var embeddedTemplate2020 []byte

//go:embed templates/2022.aep
var embeddedTemplate2022 []byte

//go:embed templates/2025.aep
var embeddedTemplate2025 []byte

func NewProject(target ...AETarget) *Project {
    t := TargetAE2020 // default
    switch len(target) {
    case 0:
        // use default
    case 1:
        t = target[0]
    default:
        panic(fmt.Sprintf("aep: NewProject accepts at most one target, got %d", len(target)))
    }

    var tmpl []byte
    switch t {
    case TargetAE2020:
        tmpl = embeddedTemplate2020
    case TargetAE2022:
        tmpl = embeddedTemplate2022
    case TargetAE2025:
        tmpl = embeddedTemplate2025
    default:
        panic(fmt.Sprintf("aep: unknown AETarget %d (forward-incompat; upgrade library)", int(t)))
    }

    p, err := FromReader(bytes.NewReader(tmpl))
    if err != nil {
        panic(fmt.Sprintf("aep: corrupt embedded template for AE %d (build bug): %v", int(t), err))
    }
    return p
}
```

不 cache，每次重 parse。8 KB 模板 < 1 ms 成本，省一份手写 deepClone 代码。**对未知 AETarget 值 panic** —— 防用户传未来枚举值（用户该升库不该悄悄退化）。

## Testing Strategy

### Tier 1 — Go unit tests（CI 必跑）

| Test | 验证什么 |
| --- | --- |
| `TestNewProject` | 默认 target (AE 2020) parse 成功、0 comp / 0 footage / 0 folders、nextItemID > 0、BPC = 8、`len(p.Warnings) == 0`（template 必须 parse 零警告） |
| `TestNewProjectAllTargets` | 三个 target（2020/2022/2025）各 parse 成功；输出 svap 字节 == **常量化的 expected 值**（防 template 被悄悄替换） |
| `TestNewProjectRejectsMultipleTargets` | `NewProject(t1, t2)` panic |
| `TestNewProjectRejectsUnknownTarget` | `NewProject(AETarget(9999))` panic |
| `TestEmptyLayrListPreserved` | New comp 的 Layr LIST 重 parse 后 `len(Layers) == 0`；roundtrip 不引入伪 sentinel layer |
| `TestNewCompositionFields` | name / w / h / fps / duration / BGColor / PixelAspect / ResolutionFactor / ID 都对 |
| `TestNewCompositionRoundtrip` | NewProject → NewComp ×2 → SetBGColor / SetResolutionFactor → WriteAEP → FromReader → 所有字段对 |
| `TestNewCompositionRejects` | name="" / w=0 / fps=0 / duration=0 都返回 error |
| `TestNewCompositionIDsUnique` | 连续 NewComposition 多次，ID 不重复 |
| `TestNewCompositionOnOpenedProject` | `Open(...)` 加 `NewComposition` 后，新 ID > 既有 max ID |
| `TestNewProjectsIndependent` | `p1 := NewProject(); p2 := NewProject(); p1.NewComp(...)` 不影响 p2 |
| `TestEmbeddedTemplateNotCorrupt` | `NewProject()` 不 panic + 返回非 nil（防嵌入资源损坏静默通过 `go build`） |

### Tier 2 — AE 2020 + AE 2025 ship gate（半自动，本地必跑）

跑两遍：用 AE 2020 / AE 2025 各自 AfterFX.exe 打开对应 target 生成的 .aep。AE 2022 不重复（结构与 AE 2025 同形，svap 不同；testing matrix 覆盖头尾即可）。

`test_data/verify_v2_1.jsx`（**路径参数化** + **try/catch 保证 .done 必写出** + **AE save normalization 二次验证**）：

JSX 通过 `AfterFX.exe -r verify_v2_1.jsx <input-aep> <done-path> <resaved-aep>` 接受 3 个绝对路径参数，不再硬编码：

```jsx
(function () {
    // ExtendScript: -r 命令行参数从 $.global.arguments / app.arguments 取（AE 版本差异）
    // 简化：用 BridgeTalk-style 或文件协议传参；这里展示语义，plan 里定具体取参方式
    var args = parseArgs(); // [<input>, <done>, <resaved>]
    var inFile = new File(args[0]);
    var doneFile = new File(args[1]);
    var resavedFile = new File(args[2]);
    var log = [];
    var ok = false;

    try {
        app.openProject(inFile);
        log.push("items.length=" + app.project.items.length);
        ok = (app.project.items.length === 2);

        if (ok) {
            var c = app.project.items[1]; // 1-indexed
            // AE save normalization 检测 — 这些值是 Go 端 SetX 写的原值，
            // 若 AE 任意一个 silently canonicalize / round 就会 fail
            ok = (c.name === "Main"
                  && c.width === 1920 && c.height === 1080
                  && Math.abs(c.frameRate - 29.97) < 1e-5    // RE-4 严格精度
                  && Math.abs(c.duration - 10) < 1e-3
                  && Math.abs(c.shutterAngle - 180) < 1e-3
                  && Math.abs(c.pixelAspect - 1.0) < 1e-5
                  && c.bgColor[0] === 20 && c.bgColor[1] === 30 && c.bgColor[2] === 40);
            log.push("Main: name=" + c.name + " " + c.width + "x" + c.height
                     + " fps=" + c.frameRate + " dur=" + c.duration
                     + " bg=[" + c.bgColor.join(",") + "]");
        }

        // Reopen-save-resave-reopen tier:
        // (a) AE 能开 ≠ AE 认结构合法 —— 第二次 open 才暴露 silent rewrite
        // (b) 验证 AE save normalization：所有字段在 resave 后仍跟我们写的相等
        if (ok) {
            app.project.save(resavedFile);
            app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
            app.openProject(resavedFile);
            ok = (app.project.items.length === 2);
            var c2 = app.project.items[1];
            ok = ok && c2.name === "Main"
                 && Math.abs(c2.frameRate - 29.97) < 1e-5
                 && Math.abs(c2.duration - 10) < 1e-3
                 && c2.bgColor[0] === 20;
            log.push("resaved items.length=" + app.project.items.length
                     + " resaved-fps=" + c2.frameRate);
        }
    } catch (e) {
        ok = false;
        log.push("ERROR: " + e.toString());
    }

    // .done 必写出，try/catch 保证 Go 端永远不会 timeout hang
    try {
        doneFile.open("w");
        doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
        doneFile.close();
    } catch (e2) { /* 写不出 .done 算 Go 端 timeout */ }
})();
```

**关键测试增强**：
- **路径参数化** —— 不硬编码 `e:/projects/...`，Go 端用 `os.TempDir() / t.TempDir()` 构 absolute 路径传给 JSX，CI / 其它机器可用
- **try/catch 包整个测试体** —— JSX 崩了也写 FAIL（不让 Go 端 hang）
- **reopen-save-resave-reopen 二次验证** —— AE 第一次能 open 不等于 AE 认结构合法；resave 后再 open 才暴露 silent rewrite / drop field
- **AE save normalization 字段对比** —— 不只验 `items.length`，还断言 frameRate / duration / shutterAngle / pixelAspect / bgColor 跟 Go 端 SetX 原值精确相等（若 AE 默默 round 29.97 → 30.0 立刻 fail）
- frame rate 精度 `< 1e-5` —— round-trip canonicalization 验证（见 RE-4）

**Plan 实现细节**（不在 spec 锁死）：
- JSX 取参方式（AfterFX `-r` 是否支持 cmdline args；可能需经临时 .json 文件传参）
- Reopen-save 那段抽 JSX helper，给 V2.2+ ship gate 复用
- Go test 用 `t.TempDir()` 生成 .aep / .done / .resaved 路径，传入 JSX

Go test：

```go
func TestV2_1_AEShipGate_AE2020(t *testing.T) {
    if os.Getenv("AE_SHIP_GATE") == "" {
        t.Skip("set AE_SHIP_GATE=1 with AE 2020 installed to run")
    }
    // 1. NewProject(TargetAE2020) + NewComposition(Main + BG_loop) + WriteAEP → v2_1_2020.aep
    // 2. rm -f v2_1_2020.done
    // 3. AE 2020 AfterFX.exe -r verify_v2_1.jsx
    // 4. 等 .done 出现（with timeout）
    // 5. 读内容，assert 第一行 == "PASS"
}

func TestV2_1_AEShipGate_AE2025(t *testing.T) {
    // 同上，用 TargetAE2025 + AE 2025 AfterFX.exe
}
```

CI 无 AE 跳过；本地 `AE_SHIP_GATE=1 go test -run TestV2_1_AEShipGate -v` 运行。

### Tier 3 — Byte-level fixture diff（不做）

跟 AE 保存的对照 `.aep` 字节对比。**跳过**：AE 会写 timestamp / UUID / session-state 字段，我们故意不复制，diff 会一堆假阳性。Tier 2 的 AE 实开足够 cover 语义正确性。

## Invariants（实现期 & 长期）

1. **`NewProject()` 永远成功**或 panic。错误信息含 "build bug" 字样。
2. **`Composition.ID` 全局唯一**，跟既有 Compositions / Footage / Folders 都不冲突。
3. **New 出来的 Composition 跟 Open 出来的同形** —— 走同一个 `parseComposition`。所有现有 `Set*` 方法立即可用。
4. **WriteAEP 输出文件 AE 25 可重开**（Tier 2 验证）。
5. **NewComposition 输入合法性检查在第一行做**，错误信息含**字段名 + 实际值**（用户能直接看出哪里错）。
6. **Template 永远是 build-time trusted bytes**。任何 runtime 失败 = 我们 parser bug，panic 而非 error。

7. **RIFX chunk tree 是 single source of truth**。`Project.Compositions / Footage / Folders` 是 typed index（performance / API ergonomics），不是真相。任何 mutation 必须先动 chunk tree，typed index 同步 append / remove；典型实现：mutation 后调 `parseComposition` 重 parse 出新 `*Composition` 再 append 到 typed slice。理论上 `WriteAEP → FromReader` 应该能从 chunk tree 重建出等价的 typed collections。这条 invariant 决定未来 V2.x `DeleteComposition / MoveFolder / ImportFootage` 不会失控。

8. **`Project.rootFold` 是 derived cache，never ownership**。永远指向当前 live chunk tree 内的根 Fold LIST。任何会替换 root chunk tree 的操作（理论上 V2.x 的 ReplaceRoot / deep copy / 重 parse）必须**重置或重 init** `p.rootFold`。**绝对禁止** `newProj.rootFold = oldProj.rootFold` 这种共享。

9. **`nextItemID` 严格单调递增，永不 reuse**。即便未来 V2.x `DeleteComposition` 删了 id=5，下次 `NewComposition` 也不该用 5。Adobe / AVID / NLE 软件惯例：删过的 id 可能被表达式 / 缓存 / proxy 引用，复用会产生 silent malfunction。保留 id 单调上行只损失 4 字节 / 删除，零风险。

10. **Mutation atomicity**：任何 mutation 操作（NewComposition / 未来 AddLayer / AddEffect）要么 chunk tree + typed index **全成功**，要么**完全 rollback** 到 mutation 前状态。典型实现：
    ```go
    oldLen := len(p.rootFold.Children)
    p.rootFold.Children = append(p.rootFold.Children, newItemList)
    comp, err := parseComposition(newItemList, ...)
    if err != nil {
        p.rootFold.Children = p.rootFold.Children[:oldLen]  // rollback
        return nil, fmt.Errorf("internal: ...")
    }
    p.Compositions = append(p.Compositions, comp)
    return comp, nil
    ```

11. **Warnings-as-failure for mutation paths**：`NewProject / NewComposition` 等构造路径若让 `parseComposition` 等产生新 warning（`len(p.Warnings)` 增加），视为 **builder bug → rollback + return internal error**，不是 silent degrade。理由：parser 给 warning 的字段对 New path 来说应该零警告（我们刚 build 出来字段，不应有 layout drift）；若产生 warning，必然是 builder 漏字段 / 顺序错。

12. **Unknown chunks must survive byte-for-byte**：任何 mutation 必须 preserve 跟操作无关的 unknown chunks 原始字节不变。这是 parser/writer 项目的生命线 —— 我们 RE 了的字段只占总字节小部分，未 RE 字段是 AE 未来 feature 的承载，不动它们才能让 AE 重开认账。架构上 `rifx.WriteAEP` 已经如此（unknown chunk pass-through）；本 invariant 显式写出来防未来 V2.x mutation API 误覆盖。

## Critical RE Prerequisites（plan Day-1 blocking 任务）

下面 4 项必须在写代码前用 fixture / template dump 给出明确结论。任一未解决都可能导致 AE ship gate 失败或 .aep 数据语义不正确。

### RE-1: `fdta` chunk 语义（**高危**）

**问题**：根 Fold LIST 含 14-byte `fdta` child（folder data header）。AE 2020 = `00 00 00 00 00 00 00 00 00 00 f4 22 00 00`（末尾 `f422` 非零），AE 2022 / 2025 = 全 0。若 fdta 编码了 child count / next-item-id 指针 / 拓扑信息，append 新 Item 后 fdta 过期 → AE 打开可能丢 comp 或弹错。

`f422` 可能不只是 child count；也可能是 flags / timestamp / allocator hint / folder capability bitset。不能简单地做 "diff item-count" 推断。

**Plan 必做（dump 矩阵）**：
- empty project（已有 templates/2020.aep / 2025.aep）
- 1 comp project
- 2 comp project
- 删 comp（先建 2 个再删 1 个保存）
- nested folder（建 folder，里面放 comp）

逐字节 diff，分类：
- **monotonic** 字段（每加 1 个 item 自增）→ 大概率 count / id allocator
- **topology-sensitive** 字段（删 / nest 时变）→ 大概率 layout pointer
- **static** 字段（永远不变）→ 可忽略

结论 1（fdta 完全无害）：文档记录证据 + "builder 不动 fdta"。
结论 2（有 monotonic / topology 字段）：spec / plan 加 `updateFdtaOnAppend(...)` 步骤，每次 NewComposition 都重写。

### RE-2: `idpc` 是 per-item persistent key（**高危**）

**问题**：comp Item LIST children 含 `idpc` chunk。AE 内部多处可能引用 item persistent key：expressions / Essential Graphics / Dynamic Link / Render Queue / proxies / 跨项目引用。若 builder 塞全零 / 固定常量，多 comp 共享同一 key 破坏唯一性，AE 不会立即报错但**后面 feature silently malfunction**。

**Default 策略 — `crypto/rand` 优先**：

"过度唯一"几乎不会出错；"错误复用"风险高且 silent。**默认实现走 `crypto/rand` 生成**，即使后续 RE 显示 AE 不严格要求唯一，也保持随机（zero downside）。

**Plan 必做**：
- 模板 dump 显示 idpc = `0000000000000000`（8 bytes 全 0）—— 这看起来跟"persistent key"语义矛盾。RE 确认 idpc 是否：
  - (a) 真 per-item UUID（builder 用 `crypto/rand`）
  - (b) opaque session-state 字段（AE 重写时填充，builder 可填全 0）
  - (c) 8-byte vs 16-byte（实际长度可能比 hex dump 显示的多）
- 多 comp fixture dump 比较 idpc 是否唯一
- 测试：故意写两个 idpc 相同的 comp，看 AE 行为

### RE-3: `idta` 完整 84-byte layout + golden fixture

**问题**：我们只 RE 出 `@0x00` (type) / `@0x14` (ID) / `@0x3A` (label)。剩余 ~75 字节字节级含义未知。Builder 塞错可能 AE 打不开。

**Plan 必做**：
- 从 template 内（如有 dummy）/ 既有 fixture 的 idta 对照填默认值字节
- 把 byte layout 表落到 `internal/aep/cdta_layout.go`（见 §Internal Design）的 idta 段
- **生成 golden fixture `fixtures/comp_item_children.txt`**：记录 buildCompItem 产出的 children **exact order / 每个 chunk size / LIST nesting**。这是 Adobe binary format 最危险点之一 —— "看起来顺序不重要" 但未来某 AE 版本 silently 开始依赖。Golden fixture 防 accidental reorder：测试比较 builder 输出跟 fixture 一致。

### RE-4: WorkArea sentinel + frame-rate canonicalization table

**问题**：
- `WorkAreaEnd dividend = 0xFFFFFFFF` 是 AE 的"使用 Duration"sentinel；新 comp 应该写 sentinel 还是真值？
- frame rate encoding 在 cdta 里是 `whole + frac/65536`。但 NTSC fractions 准确值是 `30000/1001` —— 若用户 `SetFrameRate(29.97)` vs `SetFrameRate(30000.0/1001.0)`，写出来 binary 不同？round-trip 后 reopen 是 29.97 还是 29.9699859？

**Plan 必做（freeze canonical table）**：

定一份 NTSC canonical 映射，所有公开 `SetFrameRate` / `NewComposition` 经此 normalize：

```go
// internal/aep/framerate_canonical.go (规划中)
var ntscCanonical = map[float64]struct{ whole, frac uint16 }{
    23.976: {23, computeFracExact(24000, 1001)},
    29.97:  {29, computeFracExact(30000, 1001)},
    59.94:  {59, computeFracExact(60000, 1001)},
    // 容差 < 1e-3 内的输入都映射到这些 canonical 值
}
```

否则未来：tests flaky / AE save rewrite / binary diff impossible。

- AE 新建 comp 默认 WorkAreaEnd 是 `0xFFFFFFFF` sentinel 还是 duration 真值？dump 确认（新建 1 个 / 删几个 keyframe 后保存，对比 WorkArea bytes）
- `SetFrameRate(29.97)` → WriteAEP → reopen 后精确等于 29.97（误差 < 1e-5）测试断言
- 23.976 / 59.94 同测，确认 canonical 映射有效

---

## Resolved（不再 open）

- **Composition.proj back-pointer**：架构段第 5 步已写 `comp.proj = p`，不是 open question
