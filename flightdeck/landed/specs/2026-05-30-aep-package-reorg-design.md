# aep package 重组 — 命名轴收口 + back_* 缝固化（方案①，带方案②铺路）

**Status**: Pending（spec，未写 plan）。Drafted 2026-05-30。
**关联**: V3 方向 `specs/2026-05-22-v3-direction.md`（M1/M7/M8）、incident `concurrency-unsafe-shared-chunk-bytes.md`、CLAUDE.md 硬约束 #1/#2/#3/#5。

---

## §0 约束发现 / 范围校正（先读，这条最重要）

立项时设想是「`internal/scene` + `internal/serializer` 包对半切」（V3 M8）。**深挖后确认：在「不破 Stable API + 不做方案②延迟写」的前提下，Go 语义不允许这种对半切。** 论据：

1. `internal/aep` **自身就是 API 包**，无顶层 re-export wrapper（`main/` 直接 import）。
2. 公共 API 是**方法式**且属 Stable（硬约束#2）：`(p *Project) WriteAEP` / `WriteJSON`、`(l *Layer) SetVisible`、`(c *Composition) NewShapeLayer` …
3. **Go 铁律：方法必须与其类型同包。** 故 `Project/Composition/Layer` 的类型定义 + 其 `WriteAEP/Set*/New*` 方法焊死同包。这些方法方体调 `lower_*/write_*`；若把 lower/write 挪进 serializer → scene→serializer，而 parse/write 又 operate on scene 类型 → serializer→scene → **循环依赖**。
4. 破环只剩两条路，均被本次范围排除：**改自由函数 `serializer.Write(p,w)`（破 Stable API，违#2）** 或 **方案②（纯 scene + 延迟写，= V3 重写）**。
5. 即便强切，硬约束#5（opaque 必须 byte-identical round-trip）逼 scene 仍要携带 opaque shard、serializer 仍要保留原始 chunk 树 —— 「纯净」是审美/心智收益，非总耦合下降。

**结论**：本次 = **单包内沿统一命名轴重组**，而非物理分包。这仍 100% 命中四个痛点（见 §1）。物理分包是方案②/V3 的事，本次把断层线**预先固化**（§6/§7），令日后 ② 从「70 文件大解耦」缩成「已分离区域内的局部手术」（make the change easy, then make the easy change）。

> **未来破环路径（备忘，非本次）**：方案② 可经「`lower_`/`write_` 改接口入参（依赖倒置）」破除上述循环依赖 —— 届时论据 3 失效，物理分包重新可行。本节结论**限定于「本次迭代 + 不破 API + 不做②」**，并非断言「物理分包永远不可能」。日后读者勿误读。

---

## §1 目标 / 非目标

**四个痛点（立项确认全中）→ 本次如何解：**

| 痛点 | 解法 |
|---|---|
| 找不到文件 / 命名漂移 | §2 单一 `<stage>_<domain>` 命名轴，杀掉 parse/write/lower×layer/property/shape×new/delete/move 三套正交轴 |
| 单文件太大 | §4 拆 `types_core`(1043)、`layer_accessors`(946)、test 巨文件(1789/1376/1105) |
| 职责边界糊 | §2 stage 前缀即职责；§3 每文件单一 stage×domain |
| 测试组织混乱 | §5 testutil/testhelpers 归并、fixture 与用例解耦 |

**非目标（YAGNI / 明确排除）：**
- 不拆 `internal/scene` / `internal/serializer` 物理包（见 §0）。
- 不动方案②：不改 Set* 即时 patch 语义、不引入延迟写/侧表。
- **零行为变更**：公共 API 签名/类型/JSON 字段不变（硬约束#2）；写回字节 byte-identical round-trip（硬约束#1/#5）。
- 不顺手重构无关逻辑、不改算法。纯文件归位 + 重命名 + 机械拆分。
- 不新建 Go 子包（`codec_` 仅是**前缀标记**，非 package，见 §7）。

---

## §2 命名轴：`<stage>_<domain>[_<detail>].go`

**stage 永远在前**，7 个前缀映射 4 个概念阶段：

| 前缀 | 阶段 | 职责 | 依赖 |
|---|---|---|---|
| `scene_` | 模型 | 运行时类型定义 / accessors / views / convenience / capability / IR | rifx（经 back） |
| `codec_` | 模型·纯叶 | 无 scene 耦合的纯字节/值编解码（方案②抽包预备，见 §7） | 仅 rifx/[]byte |
| `parse_` | 读 | chunk → scene | scene, rifx |
| `lower_` | 写·合成 | scene → 新 chunk（New/mutate 用） | scene, rifx |
| `write_` | 写·发射 | scene(+back) → 字节（含 WriteAEP/WriteJSON、length-preserving patch） | scene, rifx |
| `back_` | 写·缝 | `*Backrefs` chunk 引用结构（**方案②断层线**，见 §6） | rifx |
| `mutate_` | 改 | 结构性操作 new/delete/insert/move/duplicate/sync | scene, lower, rifx |

`domain` ∈ {project, composition, layer, property, shape, text, footage, mask, marker, keyframe, item, effect}。

> **lower_ 与 write_ 不合并**：二者已是干净阶段前缀（共 13 文件无需动），合并纯属 churn。文档上二者同属「serialize 阶段·两子角色」，CLAUDE.md 原文那条 `serialize_*` 据此更正为「serialize 阶段 = lower_ + write_」。

**两条硬定义（防边界再次糊化，落 lint 与 review 准则）：**

- **`mutate_` 仅限 scene graph 结构性变更**：new / delete / insert / move / duplicate / sync。`Set*` / `Patch*` 等 **length-preserving 原地字节 patch 一律不属 mutate_**。
- **`Set*` 归 `write_`**：判准则 —— **凡 length-preserving in-place patch（含 `Set*`）→ `write_`；凡需 `lower_`/重建 chunk 的结构性操作 → `mutate_`**。此规则同时约束「日后新增方法该往哪写」，非仅描述现状。
- **`codec_` 零 scene 耦合（最硬）**：不得引用任何 scene 类型（`Project`/`Composition`/`Layer`/`Property`/`ShapeNode`/`Footage`…），**连非指针参数、连函数名（如 `GradientFromProperty`）都禁**。`codec_*` 只处理值对象 / 字节流 / rifx 结构 —— 这是「方案② 时能直接整体搬走」的前提。

### §2.1 允许依赖矩阵（采纳 gpt：写死方向，防未来新环）

不只规定「scene 禁 rifx / codec 禁 scene」，而把**允许的依赖方向**全列出，由 §6 AST 守卫执行。行=from，列=to，✓=允许，✗=禁止：

| from \ to | scene | codec | parse | lower | write | back | mutate | rifx |
|---|---|---|---|---|---|---|---|---|
| **scene**  | ✓ | ✓ | ✗ | ✗ | ✗ | ✓¹ | ✗ | **✗** |
| **codec**  | ✗ | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ | ✓ |
| **parse**  | ✓ | ✓ | ✓ | ✗ | ✗ | ✓ | ✗ | ✓ |
| **lower**  | ✓ | ✓ | ✗ | ✓ | ✗ | ✓ | ✗ | ✓ |
| **write**  | ✓ | ✓ | ✗ | ✓ | ✓ | ✓ | ✗ | ✓ |
| **back**   | ✗ | ✗ | ✗ | ✗ | ✗ | ✓ | ✗ | ✓ |
| **mutate** | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |

> ¹ scene→back：仅指 scene 结构体持有 `back *xBackrefs` 字段（类型包内可见，**不触发 rifx import**）。scene 经由该字段间接握 chunk，但 scene_*.go 自身**禁 import rifx**（§6）。
>
> 关键不变量：**没有任何阶段依赖 `mutate`**（mutate 是顶层）；`back` 只依赖 rifx（chunk 侧叶子）；`codec` 只依赖 rifx（值侧叶子）。这套偏序是 DAG，杜绝 `mutate→write→mutate` 之类新环。

---

## §3 文件迁移映射

**已合规（30 文件）— 不改名**：`parse_*`(11)、`write_*`(9)、`lower_*`(4)、`back_*`(6)。

**改名（domain/operation 轴 → stage 轴）：**

`scene_`（模型/accessor）：
```
types_core.go          → 拆 (见 §4)
types_features.go      → scene_features.go
layer_accessors.go     → 拆 (见 §4)
layer_convenience.go   → scene_layer_convenience.go
layer_matte.go         → scene_layer_matte.go
layer_property_groups.go → scene_layer_property_groups.go
frame_time_accessors.go  → scene_frame_time.go
composition_convenience.go → scene_composition_convenience.go
composition_views.go   → scene_composition_views.go
project_views.go       → scene_project_views.go
project_settings.go    → scene_project_settings.go
footage_convenience.go → scene_footage_convenience.go
footage_source.go      → scene_footage_source.go
property_group.go      → scene_property_group.go
property_stream.go     → scene_property_stream.go   (泛型 PropertyStream[T]；纯机械移动不引入新类型设计；P1 纯度扫描定 codec_ vs scene_ 归属)
property_defaults.go   → scene_property_defaults.go
property_flags.go      → scene_property_flags.go
property_units.go      → scene_property_units.go
shape_graph.go         → scene_shape_graph.go        (ShapeNode/VectorGroup/RectNode… IR)
text_types.go          → scene_text.go
application.go         → scene_application.go
capability_matrix.go   → scene_capability.go         (AETarget capability matrix, V3 M7)
```

`codec_`（纯叶，方案②抽包预备）：
```
framerate_canonical.go → codec_framerate.go
postscript.go          → codec_postscript.go
gradient.go            → codec_gradient.go
cdta_layout.go         → codec_cdta_layout.go
ldta_layout.go         → codec_ldta_layout.go
```
> 迁移时**校验纯度**：`codec_*` 不得 import/引用 scene 大类型（Project/Composition/Layer/Property/ShapeLayer/Footage）。若某文件实际引用，回退为 `scene_`（已知 `property_units` 即此情形，故归 scene_）。

`write_`（serialize·发射）：
```
json.go → write_json.go   (WriteJSON 路径)
```

`mutate_`（结构性操作，统一 new/delete/insert/move/duplicate/sync 五套 operation 前缀）：
```
new_project.go         → mutate_project_new.go
new_composition.go     → mutate_composition_new.go
new_layer.go           → mutate_layer_new.go
delete_layer.go        → mutate_layer_delete.go
insert_layer.go        → mutate_layer_insert.go
move_layer.go + layer_move.go → mutate_layer_move.go   (合并 2 文件；**已确认函数名零冲突**：move_layer=`(c)MoveLayer`+`indexOfChunk`，layer_move=`(l)MoveToX`+`locate*`)
duplicate_layer.go     → mutate_layer_duplicate.go
duplicate_composition.go → mutate_composition_duplicate.go
sync_shape_layers.go   → mutate_shape_sync.go
import_closure.go      → mutate_import_closure.go
```

**已冻结归属（不再"实现时定"）：**
- `hydrate_shape.go` → **`parse_shape_hydrate.go`**（已查实：全部函数 `chunk + parseCtx → scene IR`，注释自述「Inverse of lower」，**无 lower/write/mutate 副作用 = 纯读**）。归 `parse_` 后受 §6 lint 约束：禁止引入任何 mutate 副作用。
- `lower_item_siblings.go`：保留 `lower_` 前缀（已合规）。
- **Set\* 方法**（`SetVisible` 等，现居 `write_layer.go`，length-preserving 原地 patch）：**留在 `write_` 阶段**（见 §2 硬定义：length-preserving in-place patch → `write_`），不挪 `mutate_`。

---

## §4 大文件拆分

| 源 | 行 | 拆为 |
|---|---|---|
| `types_core.go` | 1043 | `scene_project.go` / `scene_composition.go` / `scene_layer.go` / `scene_property.go`（边界按**类型名**切：`type Project` / `type Composition` / `type Layer` / `type Property` 各成一文件，连同各自的方法/常量。**不用行号**——行号是写作快照，会因注释/空行偏移失效） |
| `layer_accessors.go` | 946 | 按 concern 拆：`scene_layer_accessors.go`（transform/timing）+ `scene_layer_property_access.go`（property 树导航）。目标每文件 < ~450 行 |
| `project_settings.go` | 647 | 仅改名 `scene_project_settings.go`（体量可接受，不强拆） |

测试巨文件（§5 一并处理）：`layer_test.go`(1789)、`shape_layer_shipgate_test.go`(1376)、`text_test.go`(1105)、`composition_test.go`(777)、`keyframe_test.go`(645) → 按 feature/子项拆，单文件目标 < ~600 行。

> 拆分纪律：**纯移动**，import + 函数体逐字搬，不改逻辑。每拆一个文件即 `go build` + 相关 `go test` 绿。

---

## §5 测试基建归并

- **测试命名不强制 `<stage>_<domain>_test.go`，允许 feature-first**（如 `shape_trimpaths_test.go` / `text_animator_test.go` / `keyframe_hold_test.go`）—— 测试按「测什么功能」定位比按「测哪个阶段」更自然，强套 stage 轴反而让测试变大杂烩。源码强制 stage 轴，测试只要求下面三条：
  - 单文件不过大（目标 < ~600 行）；
  - 共享 helper 一处定义、收敛：`testhelpers_test.go` / `ship_gate_helpers_test.go` → 并入 `testutil_*_test.go` 体系，禁重复；
  - fixture 路径常量/loader 集中到 `testutil_fixtures_test.go`，用例只引名字。
- ship-gate 测试（`*_shipgate_test.go` / `AE_SHIP_GATE` 门控）保持独立可单跑；拆分后 **AE matchName 字符串**（如 `ADBE Vector Shape`）与 `go test -run` pattern 均不变。

---

## §6 `back_*` 缝固化（方案②断层线）

`back_*.go` 已是 chunk 耦合的隔离区（每 scene 类型一个 `back *xBackrefs`）。本次**固化为契约 + 加 lint 守卫**，令日后方案②「把 back 指针反转成 serializer 侧表」成为局部手术：

1. **契约注释**（每个 `back_*.go` 顶部）：声明「这是 scene↔chunk 唯一耦合点；方案②时本文件整体迁入 serializer 侧表」。
2. **lint 守卫**（新增 `arch_boundary_test.go`）——把 §2.1 依赖矩阵变成 CI 可执行约束（方案②的增量护栏）。断言：
   - **scene_ 禁 `import rifx`**（采纳 gpt，强于「禁 `*rifx.Chunk`」）：禁整个 `import ".../internal/rifx"`，因 `rifx.List`/`rifx.ChunkID`/`rifx.Reader` 会从别处渗回。
   - **codec_ 禁引用任何 scene 类型**（§2 硬定义：含非指针参数、含函数名）。
   - **全阶段遵守 §2.1 矩阵**（尤其无人依赖 `mutate`、`back`/`codec` 仅依赖 rifx）。
   - **实现选 AST，非 grep**（采纳 gpt round-2，**推翻**第一轮的 grep-first）。理由已被本次调查坐实：grep 既 **过匹配注释**（`types_core.go` 的 `rifx.Chunk` 仅出现在注释，被误判）、又 **漏匹配**（`project_views.go` 实际用 `rifx.ChunkID` 却因分组 import 被精确 grep 漏掉；别名 `r "...rifx"` / `type Chunk = rifx.Chunk` 更必漏）。`go/parser`+`go/ast` 扫 import spec + selector expr + 类型引用，几十行，成本低且无盲区。
   - **白名单由 AST 守卫自身在 P1 枚举**（一并解决「底数未知」）：grep 初估 scene_ 候选中真 import rifx 者约 **4–6 个**（`types_features` / `project_settings` / `property_group` / `property_flags` 确认，`project_views` 疑似，`types_core` 实为注释误报）—— 但**准确清单以 AST 守卫首次运行输出为准**，不靠人工数。
   - **codec_ 纯度检查并入同一 AST 守卫**（采纳 ds）：纯度从「迁移时一次性手检」升级为 CI 持续断言。
   - **清零期限**：白名单先放行、不阻塞重组；**§9 P6 必须清零**，无法清零项列技术原因并作为「方案②/V3 前置解耦项」提交 V3 spec —— 防「逐步」无限拖延。

---

## §7 `codec_` 纯叶 — 抽包预备（不抽包）

`codec_*.go`（§3）是无 scene 耦合的纯编解码。**本次只加前缀标记，不建 Go 子包**（YAGNI：当前够格者少，建包是 churn）。

方案②/V3 真正需要 `internal/serializer` 时，`codec_*` 是**第一批零成本平移**的文件（serializer 直接复用）。前缀即「抽包就绪」信号。

---

## §8 不变量与验证（硬约束守卫）

每一步（每改/拆一个文件）必须：
1. `go build ./...` 绿。
2. `go vet ./... && go test -count=1 ./internal/aep/...` 全绿（含现有全部 ship-gate 单测的非门控部分）。
3. **行为零变更证明**：重组**前**对一组 fixture 跑 `Open → WriteAEP → bytes`，存基线；重组**后**重跑，断言 **byte-identical**（硬约束#1/#5 round-trip）。基线脚本放 `tmp_debug/reorg_roundtrip_baseline/`，内含 `README.md`（用法 + 过期策略 + fixture 清单）；基线记录所用 **fixture 文件的 git hash** 确保可复现。
4. 公共 API diff：重组前后 exported 声明集合**零 diff**（硬约束#2）。**具体工具**：`go doc -all ./internal/aep > before.txt` → 重组后 `> after.txt` → `diff` 须空（`go doc -all` 列全部 exported decl，稳定可比；**不用 `go tool nm`** —— 那是链接器符号，含内部/位置噪音，会假阳/假阴）。另跑 `go build ./main/...` 确认下游编译不破。
5. **AE 双版本 ship-gate 不必重跑**（无结构性写路径变更，纯文件移动）—— 但**收尾跑一次** `AE_SHIP_GATE=1` 全套作为终验，确认零回归（自验留痕，无需人工，见 feedback `ae-ship-gate-self-serve`）。

---

## §9 落地路径（strangler，步步绿）

分阶段，每阶段独立可 commit、suite 全绿：

- **P1 命名轴·叶子优先 + 基线 + 守卫**：迁 `codec_*`（纯、低风险，含 `property_stream` 纯度扫描定归属）+ 建 §6 **AST 守卫**（先白名单全放行）。
  - **P1 exit criteria（硬）**：① `tmp_debug/reorg_roundtrip_baseline/` 脚本就位、对**全部** fixture 跑通并存基线（含 git hash）—— 不就位则后续「前后比对」无依据。② AST 守卫跑通并**输出权威白名单**（scene_ 真违 rifx-import 的文件清单 + codec_ 违 scene 引用的清单），作为 P3/P6 工作量基准。
- **P2 命名轴·mutate**：统一 5 套 operation 前缀 → `mutate_*`，合并 `move_layer`+`layer_move`（已确认无冲突）。
- **P3 命名轴·scene**：domain-accessor 文件批量 → `scene_*`。
- **P4 大文件拆分**：`types_core` / `layer_accessors`（§4）。
- **P5 测试基建**：§5 归并 + test 巨文件拆分。
- **P6 收口**：§6 lint 白名单清零（无法清零项列原因 + 提交 V3 spec 作前置解耦项）、CLAUDE.md 更新（§10）、终验 AE 双版本 ship-gate、cockpit/manifest 落地。

> 每阶段是纯机械变换 + 绿灯，可随时中断且仓库始终可 ship。
>
> **关于阶段顺序**（回应 gpt「scene 应先于 mutate 以免两轮 rename churn」）：**同包内文件改名 = 零符号 churn** —— 所有 symbol 包内可见，引用不随文件名变，仅 `git mv`，不触碰任何引用处。故阶段序由**风险**驱动而非 churn：codec-first 是因纯度扫描 + 基线建立 blast radius 最小、最该先验。codec→mutate→scene 与 codec→scene→mutate 在 churn 上等价，此处取前者。

---

## §10 CLAUDE.md 更新（本次必带）

硬约束 #3 **原文**为：

> 3. **`internal/aep` 单 package**。V3 用**文件名规约**（`scene_*.go` / `serialize_*.go` / `parse_*.go` / `back_*.go`）+ lint 维护内部边界，不用子包 —— 子包带来的 re-export 噪音 > 隔离收益。

更正为（diff）：
- **保留**「当前仍单 package」结论，但**给出 §0 的真实理由**（Go 方法同包 + Stable API，非「re-export 噪音」这个 V0 论据）。
- **命名轴升级为 §2 的 7 前缀方案**（`scene_/codec_/parse_/lower_/write_/back_/mutate_`），取代原文 `scene_*/serialize_*/parse_*/back_*` 四前缀的**未实现**描述。
- 注明 `serialize` 阶段 = `lower_*`（生成新 chunk）+ `write_*`（发射字节，含 length-preserving patch）两子角色。**附具体提醒文案**（采纳 ds，免后人翻源码）：「新建/结构性产物走 `lower_`（从 scene 合成全新 chunk，如 `lowerShapeLayer`）；既有 chunk 原地改字节走 `write_`（如 `SetVisible` patch `ldta`）。判据：要不要无中生有一个 chunk —— 要→`lower_`，不要→`write_`。」
- 注明 `codec_*` = 方案②抽包预备纯叶（禁引用 scene 类型）；`back_*` = scene↔chunk 断层线 + §6 lint 守卫（scene_ 禁 import rifx）。
- 指向本 spec 与 V3 方向 spec 作为分包演进依据；注明「物理分包非永不可能，方案② 经接口隔离即可破环」（见 §0 备忘）。

---

## §11 风险

| 风险 | 缓解 |
|---|---|
| 大批改名引入隐性行为变更 | §8 byte-identical round-trip 基线 + API symbol 零 diff，每步验证 |
| `codec_` 纯度误判（实际引用 scene） | §3 迁移时校验，违例回退 `scene_` |
| test 拆分漏带 helper / fixture 路径 | §5 helper 集中 + 每步 `go test` 绿 |
| git history 因改名变难追 | **每批 `git mv` 独立 commit**（不与逻辑改动混）、依赖 git `--find-renames` 默认阈值保 rename 链；拆分文件在 commit message 标注源文件 |
| 范围蔓延到方案② | §1 非目标硬边界 + §6 只固化缝不反转 |

---

## §12 决策记录

- 2026-05-30 立项：用户要求「停下来重构」，四痛点全中。
- 2026-05-30 范围校正：方案① 物理分包被 Go 方法/非导出字段双墙挡死（§0），收敛为「单包命名轴重组 + 缝固化」。用户同意方案①、要求带方案②铺路 → §6/§7/§10。
- 2026-05-30 外审整合：三份外部 AI review（`safety-reviews/ds` `safety-reviews/claude` `safety-reviews/gpt`）均 **Approve with minor changes**。disposition 见 §13。

---

## §13 外审 disposition（safety-reviews/{ds,claude,gpt}）

**采纳（已改入本 spec）：**
- codec_ 硬化：禁引用任何 scene 类型（含非指针参数/函数名）→ §2 (gpt/ds)
- §6 lint 升级：scene_ 禁整个 `import rifx`（非仅 `*rifx.Chunk`）+ 白名单 → §6 (gpt)
- mutate_ 硬定义 + Set*/write_ 判准则 → §2 (gpt/claude)
- 测试放宽 feature-first 命名 → §5 (gpt)
- hydrate_shape 冻结归 parse_（已查实纯读）→ §3 (ds)
- move 合并无冲突标注（已查实）→ §3 (claude)
- property_stream 去「V3 IR」暗示 + P1 纯度扫描 → §3 (ds/claude)
- §8 fixture git hash + baseline README；§9 P1 exit criteria 绑基线 → §8/§9 (claude/ds)
- §6 lint 选 grep-first；白名单 P6 清零期限 + 无法清零项提交 V3 spec → §6/§9 (claude/ds)
- §10 引 #3 原文 + diff + lower_/write_ 触发提醒 → §10 (claude/ds)
- §0 接口隔离破环备忘 → §0 (claude)
- git mv 独立 commit + rename 追踪 → §11 (ds/claude)

**反驳（不采纳，附技术理由）：**
- gpt：P 阶段 scene 先于 mutate「避免两轮 rename churn」→ **不采纳**。前提有误：**同包内文件改名零符号 churn**（引用不随文件名变，仅 `git mv`）。阶段序由风险驱动，codec-first 保留。已在 §9 写明。

**措辞修正（已改）：** §2 back_ 行 `*Backrefs` 渲染；§5 matchName 歧义（澄清为 AE matchName 字符串 + `-run` pattern）；§10 缺 #3 原文。

### 第二轮外审 disposition（2026-05-30，三家均 Approved for planning）

**采纳（已改入）：**
- **依赖矩阵** → 新增 §2.1（gpt）：写死 7 阶段允许依赖方向，DAG 杜绝 `mutate→write→mutate` 新环。
- **AST 守卫，推翻 grep-first** → §6/§9（gpt）：本次调查坐实 grep 过匹配注释（`types_core` rifx.Chunk 仅注释）+ 漏匹配分组/别名 import（`project_views`）。
- **白名单底数由 AST 守卫枚举** → §6/§9（claude）：解决「P1 exit criteria 工作量未知」；初估 4–6 个，准确以守卫输出为准。
- **§4 拆分用类型名非行号** → §4（claude）：行号会因注释/空行偏移失效。
- **§8 symbol dump 指定工具** → §8（claude）：`go doc -all` 比 `go tool nm` 准（后者含链接器噪音）。
- **codec_ 纯度并入 AST 守卫（CI 持续）** → §6（ds）：非一次性手检。
- **§10 lower_/write_ 触发提醒给具体文案** → §10（ds）。

**无新阻塞项。** 三家一致认为本 spec 已从「重命名提案」升级为「可机械执行 + CI 可守卫 + 对方案②有铺路价值」的工程契约，可进 plan。
