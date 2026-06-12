# aep-parser — AI 协作规范

Go 实现的 Adobe After Effects `.aep` 二进制解析器，对照 boltframe/aftereffects-aep-parser 重写 + 增量。**读取下限 AE 2020**。

> 项目用 **flightdeck**（deck 在 `flightdeck/`）。会话入口由 SessionStart hook 自动注入接管指令；手动可跑 `/flightdeck:preflight`。第一件事：读 `flightdeck/cockpit.md`（状态/下一步）+ 各 `INDEX.md`。

## 数据流

```
io.ReadSeeker
   ↓
internal/rifx        ── 通用 RIFX chunk 树 (磁盘字节 ↔ Chunk 树)
internal/codec       ── 纯值/字节编解码 (无 AEP 语义、无 rifx)
internal/scene       ── 运行时模型 (Project / Composition / Layer / ...) + writer 接口 (编译期不碰字节)
internal/serializer  ── chunk ⇄ scene (parse / lower / write / back / mutate)，实现 scene 的 writer 接口
internal/aep         ── 薄 facade (公共 API：Open / FromReader / New* / 结构性 op / 类型别名)
   ↓
*aep.Project ──┬── (p).WriteAEP(io.Writer)    二进制写回 (length-preserving)
               └── (p).WriteJSON(io.Writer)   JSON 单向导出（无 ReadJSON）
```

**关键边界**：`internal/rifx` 不知 AEP 语义，只懂 RIFX 容器；`internal/codec` 纯值/字节、不碰 rifx；`internal/scene` 编译期不碰字节（chunk 耦合经 writer 接口倒置到 serializer 实现）；只有 `internal/serializer` 同时见 scene+codec+rifx。DAG 单向、不成环（详 #3）。

## 硬约束（不可破）

1. **写回 default 是 length-preserving**。改字段不准动 chunk 大小；少数 length-variable 例外（name / comment / expression / 字体名 / 文本字符串）`WriteAEP` 会重算父 LIST size + 内嵌 LIST btdk size header。结构性 ops（NewComposition / NewShapeLayer / DeleteLayer 等）走 atomic invariants：warnings-as-failure + rollback to pre-call state + AE 双版本 ship-gate 验证。详 `incidents/ae25-acceptance-gate.md`。
2. **public API 分级**：
   - **Stable（核心 R/W）**: 已通过双版本 ship-gate 的 `Open` / `FromReader` / `WriteAEP` / `WriteJSON` / `Set*` / getter。**签名 / 类型 / JSON 字段不可动**。
   - **Stable（结构性 op）= 语义稳定，调用形态可随包边界重组变化**（M8 方案② 决议，2026-06-09 用户批准，覆盖原「Stable 重构不能动签名」对结构性 op 的部分）：ship-gated 的 `New*` / `Delete*` / `Insert*` / `Move*` / `Duplicate*` / `Add*` / `Remove*` / `SetDimensionsSeparated` 等结构性写路径，**语义契约不变**（chunk 输出 byte-structural 等同、双版本 ship-gate 持续通过）；但**调用形态可在物理分包时从 scene 方法改为 facade 自由函数**（如 `comp.DeleteLayer(i)` → `aep.DeleteLayer(comp, i)`）——因其实现 building chunk 必须住 `internal/serializer`，而方法须与 scene 类型同包又不能访问 serializer（Go 语义墙，详 #3）。此类形态变更：**commit 标 BREAKING + 同步 `flightdeck/checklists/commits.md` API 表**，不算违反 Stable 契约。核心 R/W（Set*/Open/Write）不受此豁免，仍签名稳定。
   - **Alpha**: 显式标 alpha / deferred / 未 ship-gate 的新 API。可改可删，commit message 标 BREAKING。
   - review 时撤销新加但已知 broken 的 API 不算违反此约束。
   - 具体哪些字段属 Stable / Alpha 详 `flightdeck/plans/coverage.md`。
3. **多包物理分层 `internal/{rifx,codec,scene,serializer}` + `aep` facade**（M8 方案② 物理分包已落，2026-06-09：scene 抽 `ac62b25`、serializer 抽 `a347e46`）。DAG 上依赖下、禁逆向/成环：
   - `rifx`（叶）：通用 RIFX chunk 树，不知 AEP 语义。
   - `codec`：纯值/字节编解码（framerate / gradient / property-stream / cdta·ldta layout / render-settings …）。禁 import scene / serializer / rifx / aep。
   - `scene`：运行时模型 + accessor + writer 接口 + `WriteJSON`。**禁 import rifx / serializer / aep**（仅可 import codec）。chunk 耦合经 scene 内定义、serializer 实现的 `XWriter` 接口倒置（B′）——scene 编译期不碰字节。
   - `serializer`：`parse_`（chunk→scene）/ `lower_`+`write_`（scene→chunk · 发射字节/length-preserving patch）/ `back_`（`*Backrefs` chunk 引用 + writer 实现）/ `mutate_`（结构性 new/delete/insert/move/duplicate）。import scene+codec+rifx；**禁 import aep**（防环）。包内仍以 `<stage>_<domain>` 命名轴组织。
   - `aep`：薄 facade（`aliases`/`facade_codec` = 类型·枚举别名；`facade.go` = `Open`/`FromReader` + 结构性 op 自由函数委托 serializer；`scene_application.go` = `Application`）。**公共 API 全经此包**。
   - 公共 R/W 方法（`(p *Project) WriteAEP` / `Set*`）= scene 类型方法，经 writer 接口委托 serializer 实现（保方法式 API）；结构性 op = facade 自由函数（`aep.DeleteLayer(comp, i)`，详 #2）。
   - 边界守卫：`internal/aep/arch_boundary_test.go`（AST 包级 import-DAG 断言，`go test` CI 强制）+ `tmp_debug/dag_boundary`（`go list` 手动核）。
   设计/历程详 `flightdeck/specs/2026-06-07-v3-m8-physical-split-design.md` + `cockpit.md`。
4. **嵌入资源目录命名复数**：`internal/serializer/templates/`（非 `template/`）。Go `//go:embed` 限制资源必须在 package 同目录或子目录——M8 物理分包后随 `lower_`/`mutate_` 居 serializer。
5. **Opaque preservation**（V2.2 教训）：parser 未解的 chunk 必须 byte-identical round-trip。scene types 携带 opaque shard，serializer 原位重发。任何 "regenerate from scene" 路径必须保留它，否则 AE 会 silent-drop。
6. **AE 接受 gate**：任何新结构性写路径（NewX / DeleteX / DuplicateX / V3 mutation API）必须跑 AE 2020 + AE 2025 双版本 ship-gate 才算 ship。详 `incidents/ae25-acceptance-gate.md` + `checklists/re-fixture.md` § GDI 自动化。
7. **交付准则（什么算「能用」）**：只有经 AE 双版本 ship-gate 实测 PASS 的能力、且仅在其 gate 覆盖的**规模 + 组合**边界内，才可对外宣称「能用 / 可交付」。四条认知红线：(a) **Go round-trip 0 警告 ≠ AE 接受**（Go parser 宽松，只验字节读回，不验 AE 引擎消化——`SetExpression` 栽在这）；(b) **`Stable` 有覆盖边界**，超出规模/组合（如 gate 只验 2 关键帧、你用 13 个空间关键帧）即退回未验证；(c) **组合 / 端到端是独立交付项**，单点 gate 不为「从零拼完整工程」背书；(d) **「值 round-trip 绿」≠「渲染正确」**——gate 必须验到能力的「作用面」：**渲染类能力（颜色/opacity/可见效果）的 gate 必须验渲染像素**（render 一帧采样像素），只验值 round-trip = 假绿（shape 颜色栽在这：值全对、gate 全绿，AE 渲染成默认红）。演示 / 示例 / 对用户的能力描述**禁止**混入未验证能力（混入必须当场标红）。本库擅长「读真实 .aep + 局部改 + 写回 / 单点结构性写」，**不擅长从零拼复杂工程**（from-scratch 缺省略 / 表达式不认 / 规模未测）。全文详 `checklists/delivery-contract.md`。

非显然内部不变量（gotcha；逐条详 `flightdeck/incidents/`，入口已加载其 INDEX）：TickRate per-composition（非全局）· Keyframe 两种字节布局必走 `layoutFor` · 所有 Set\* mutate 共享 chunk bytes（调用方自己锁）· chunk ID 大小写敏感（Tdb4 ≠ tdb4）。

## 入口命令

- 校验: `go vet ./... && go test ./...`
- 单测: `go test ./internal/aep/ -run 'TestX' -v`
- 详细操作（tmp_debug 工具表 / fixture 验证 / ship-gate）: `flightdeck/checklists/`

## 工作风格

- **一律用中文跟用户交流**（含解释、提案、报告、commit body 可英文按既有惯例）
- 代码优先，设计讨论精简
- 重构 / 迁移**先读源码再动**，不瞎猜
- **内部实现无注释**，除非 WHY 不明显；但**导出 API 的 doc comment = 文档源**（英文为源，`cmd/docgen` 从中生成 `docs/*.md`）。行内实现注释仍禁；导出符号上方的 doc comment 是文档载体，不算违反。

## 文档地图

- 字段覆盖矩阵 + 暂搁 / 不可达 / negative findings: `flightdeck/plans/coverage.md` + `coverage-detail.md`
- API 同步表（改任何 public API 必读）: `flightdeck/checklists/commits.md` § 命令一致性
- 测试惯例 / 验证流程 / tmp_debug 工具表: `flightdeck/checklists/verify.md`
- JSX RE 工作流 + ship-gate + RE fixture 双轨 + Types-for-Adobe 参考: `flightdeck/checklists/re-fixture.md`
- 当前里程碑 / 进度 / 下一步: `flightdeck/cockpit.md`（不在此留存，避免状态漂移）
