# showcase 产出规约 — 大阶段示例 + 用户审核 — checklist

SUMMARY: showcase 产出规约 — 大阶段示例 + 用户审核
READ WHEN: 一个大阶段（独立 work effort / 可交付 feature arc）落地后要产出审核示例；新增一个 showcase 方向文件夹；写或更新 showcase/<方向>/INDEX.md；纠结某次成果算大阶段（必出）还是小阶段（攒批）；想知道 showcase 产物怎么重生成；复刻的工程某帧渲染与原版不一样要逐帧/分层对比定位；想导出 原版-vs-clone 逐帧 PNG 给用户对比

---

> 2026-06-13 用户立规：每个大阶段须从零产出一个示例工程让用户审核；小阶段攒批。2026-07-03 起，showcase 规则正文从 `briefing.md` 内化到本文；briefing 只保留路由入口。

## 核心规则

- **大阶段必出 showcase 供用户审核**。大阶段 = 有独立 work effort / 可交付 feature arc 的能力方向，例如矢量滤镜家族整体、新 layer 类型组、expression 激活、precomp 嵌套、gradient 体系。落地后须在 `flightdeck/showcase/<方向>/` 产出纯 Go 从零生成 + AE 实渲的示例工程并通知用户审核。
- **小阶段不单独出**。单个滤镜、单个 `Set*` 字段、单个 enum 等小 slice 攒到所属方向 showcase 里更新，不新建独立 showcase。
- **视觉能力域必须有眼验面**。任何会改变渲染像素的写能力方向，都必须能在 showcase 或 render-pixel ship-gate 中看到作用面。值 round-trip 绿不等于渲染正确。
- **非视觉域可用读值档**。project、render-queue、io、EG 面板、纯 meta 等无渲染面的能力，可以用 gen + verify.jsx dump DOM 值核对。
- **状态有 review-gate**：agent 自己 AE 实渲 + 眼验后只能标 `待review`；只有用户真机打开/复核通过后才能标 `complete`。
- **capindex 对齐**：capindex tag `verify=render-pixel` 的能力应被某 showcase 方向覆盖；render-pixel 能力没有 showcase/占位时，要补占位或实档。

## 何时出（大阶段 vs 小阶段）

- **大阶段（必出 / 必通知用户审核）** = 有独立 work effort、能作为一个 feature arc 交付的能力方向。例：矢量滤镜家族整体、新 layer 类型组（Solid/Null/Camera/Light…）、expression 激活、precomp 嵌套、gradient 体系。
- **小阶段（不单独出，攒批）** = 单 slice：单个滤镜、单个 `Set*` 字段、单个 enum。落地时只更新所属方向 showcase 的 `gen.go` + `INDEX.md`，不新建文件夹、不单独通知。

## 目录形态（按能力方向分区）

```
flightdeck/showcase/
  INDEX.md                      ← 总览：列所有方向 + 一句话 + 重生成命令
  <方向>/                        ← shape-filters / shape-primitives / keyframes-ease /
    INDEX.md                    ←   expressions / precomp-nesting / gradient / text / layers …
    gen.go                      ← package main，纯 Go 从零构建 <方向>.aep（tracked）
    render.jsx                  ← AE 打开 + saveFrameToPng 出 <方向>.png（tracked）
    <方向>.aep                   ← 产出工程（gitignored）
    <方向>.png                   ← 渲染帧（gitignored）
```

- **tracked**：`INDEX.md` · `gen.go` · `render.jsx`。**gitignored**：`*.aep` · `*.png`（根 `.gitignore` 已加 `flightdeck/showcase/**/*.aep|*.png`）。
- `gen.go` 是 tracked 代码 → **必须保持 `go build ./...` / `go vet ./...` 绿**；它 `package main`，import `internal/aep` facade（与 tmp_debug 同，模块内允许）。
- **目标版本默认 `aep.NewProject(aep.TargetAE2020)`**（读取下限，产出文件 2020+2025 都能开）。**只有能力是 AE 2025 特有时才用 `TargetAE2025`**——双版本 ship-gate 都过的能力一律 2020（用户 2026-06-15 立规；3d-camera 初版误用 2025 已纠）。渲染眼验也优先用 AE2020 跑（与 target 一致）。

## INDEX.md 格式（frontmatter 格式段 + 正文）

frontmatter 写明「测哪个方向」+ 元信息；正文列产物 + 类型 + 布局。模板：

```markdown
---
showcase: <方向 slug>
direction: <一句话：这个方向从零测什么>
capabilities: [<能力1>, <能力2>, …]        # 覆盖的能力点
gates: [<对应 ship-gate 测试名>, …]          # 每个能力的双版本 gate（溯源）
status: complete | partial
last_updated: YYYY-MM-DD
regenerate: "go run ./flightdeck/showcase/<方向>  +  AE render.jsx"
---

# <方向> — showcase

## 这个方向测什么
<2-3 句：从零生成了什么、覆盖哪些能力、为什么这样布局>

## 产物
| 文件 | 类型 | 说明 |
|---|---|---|
| gen.go | 生成器(Go) | 纯 Go 构建 <方向>.aep |
| render.jsx | 渲染脚本 | AE saveFrameToPng → <方向>.png |
| <方向>.aep | 产出工程 (gitignored) | <层数/结构> |
| <方向>.png | 渲染帧 (gitignored) | frame 0 视觉 |

## 布局
<网格/层次表：每个 cell/layer = 哪个能力 + 期望效果，供用户逐项眼验>
```

## 状态 review-gate（`status` 两档）

| 值 | 含义 | 谁能标 |
|---|---|---|
| **`待review`** | agent 已建 + AE 实渲 + 自己眼验，**用户尚未真机复核** | agent（默认落这档） |
| **`complete`** | **用户在真机打开 .aep 验收过** | **仅用户确认后** agent 才翻 |

**agent 自己渲染眼验 ≠ 用户真机验收**（同交付准则「Go round-trip ≠ AE 接受」）。**agent 禁止自行标 `complete`**；新建/更新一律落 `待review` 并通知用户。顶层 `showcase/INDEX.md` 用 `🔍 待review` / `✅ complete` 对应。

## 操作流程（出一个大阶段 showcase）

1. 新方向 → 建 `flightdeck/showcase/<方向>/`，写 `gen.go`（参考既有方向；纯 Go facade 调用，输出到本目录的 `<方向>.aep`）。
2. `go run ./flightdeck/showcase/<方向>` 构建 .aep。
3. 写 `render.jsx`（打开 .aep → `saveFrameToPng(0, …)`），用 `scripts/ae-worker/ae_run.ps1` 跑出 `<方向>.png`。
4. **AE 实渲眼验**（红线4）：Read 渲染 png，确认每个能力的视觉对了——不靠值 round-trip 假绿。
5. 写/更新 `<方向>/INDEX.md`（上面模板，`status: 待review`）+ 顶层 `showcase/INDEX.md` 加一行（`🔍 待review`）。
6. commit tracked 三件（INDEX.md/gen.go/render.jsx）；**通知用户真机复核**。
7. **用户真机验收通过后** → 才把该方向 `status` 翻 `complete`（顶层表同步 ✅）。

## 交付准则对齐

showcase 里每个能力须已过双版本 ship-gate（`delivery-contract.md`）；showcase「组合工程」本身是独立交付项，**产出必经 AE 实渲眼验**，禁止只验值 round-trip 就宣称示例「能看」。

## 复刻保真：原版 vs clone 逐帧对比渲染（定位「渲染不一样」）

复刻一个真实 .aep 后，用户/你发现某帧「长得不一样」时——**别靠 keyTime/值 dump 占卜，直接渲帧看图**（红线4：值对≠渲染对；本工作流揪出了 comp ⑤ 的 tdb4 时基 bug + Scale/Rotation 漏复刻，详 [[layer-replication-drops-static-transform-channels]]）。三件 tracked 工具（`flightdeck/showcase/booyah-clone/`）：

- **`render_shape_compare.jsx`** — 一趟 AE 里先开原版、再开 clone，把某 comp 的 frame `0..N-1` 各导一张 PNG。sidecar `render_shape_compare.txt`（UTF-8，每行一项）：`原版.aep` / `clone.aep` / `comp 名` / 帧数 / 输出目录 / 源前缀 / clone前缀。跑：`pwsh scripts/ae-worker/ae_run.ps1 -AeExe <AE> -Jsx <这个> -Done <.done>`。
- **`render_solo_layers.jsx`** — 同一 comp 把每层逐个 `solo` 单独渲一帧（原版 + clone），**隔离是哪一层不对**。sidecar：原/clone/comp/帧号/输出目录。它还顺手把每层 start/opacity 写进 `.done`。
- **`dump_layer_time.jsx`** — dump 每层 start/in/out/stretch/timeRemap/source（不渲染，快），排除时间属性差异。
- **`tools/debug/dump_precomp_xform`**（Go）— dump 层的静态 Scale/Rotate Z/Anchor 值（parser 读 Scale 为分数）。

**命名给用户对比**：Windows 按名排序，要让「同帧的 source/copy 相邻」必须**帧号在前**——`NN_source.png` / `NN_copy.png`（或 `NN-source`/`NN_copy`），**不是** `source_fNN`（那样所有 source 排一堆、所有 copy 排一堆没法滑动对比）。1 起始编号（`01`=第1帧）。renamed 例：
```bash
for f in $(seq 0 30); do o=$(printf "%02d" $f); n=$(printf "%02d" $((f+1)));
  mv "source_f${o}.png" "${n}_source.png"; mv "copy_f${o}.png" "${n}_copy.png"; done
```

**诊断顺序**：① 叶子 comp 逐帧对比（源动画对不对）→ ② 合成 comp 逐帧（哪几帧差）→ ③ solo 每层（哪层差）→ ④ 对差的层 dump transform 值（`dump_precomp_xform`）/ 时间属性（`dump_layer_time`）→ 找到差异字段。**别在「合成差」时就猜时基**——先 solo 缩小到层，再看该层的具体字段（位置/缩放/旋转/不透明度/源时间）。渲染产物进 `tmp_debug/`（gitignored），JSX/Go 工具 tracked。
