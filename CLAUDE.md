# aep-parser — AI 协作规范

Go 实现的 Adobe After Effects `.aep` 二进制解析器，对照 boltframe/aftereffects-aep-parser 重写 + 增量。**读取下限 AE 2020**。

> flightdeck 约定见 flightdeck-workflow skill；新会话第一件事按 skill entry checklist 跑（先读 `flightdeck/cockpit.md` + `flightdeck/manifest.md`）。

## 数据流

```
io.ReadSeeker
   ↓
internal/rifx        ── 通用 RIFX chunk 树 (磁盘字节 ↔ Chunk 树)
   ↓
internal/aep/*       ── AEP 语义层 (Chunk 树 → Project / Composition / Layer / ...)
   ↓
*aep.Project ──┬── *aep.WriteAEP(io.Writer)    二进制写回 (length-preserving)
               └── *aep.WriteJSON(io.Writer)   JSON 单向导出（无 ReadJSON）
```

**关键边界**：`internal/rifx` 不知道 AEP 语义，只懂 RIFX 容器。`internal/aep` 不直接读字节流，只通过 `rifx.Chunk` 操作。两层职责分清，不交叉。

## 硬约束（不可破）

1. **写回 default 是 length-preserving**。改字段不准动 chunk 大小；少数 length-variable 例外（name / comment / expression / 字体名 / 文本字符串）`WriteAEP` 会重算父 LIST size + 内嵌 LIST btdk size header。结构性 ops（NewComposition / NewShapeLayer / V3 的 Layer.Remove 等）走 V2.1 atomic invariants：warnings-as-failure + rollback to pre-call state + AE 双版本 ship-gate 验证。详 `incident-reports/ae25-acceptance-gate.md`。
2. **public API 分级**：
   - **Stable**: V1 核心 + V2.1 NewProject/NewComposition + 已通过双版本 ship-gate 的 R/W 字段。重构不能动签名 / 类型 / JSON 字段。
   - **Alpha**: 显式标 alpha 或 deferred 的（V2.2 ShapeLayer / 未 ship-gate 的新加 API）。可改可删，commit message 标 BREAKING。
   - review 时撤销新加但已知 broken 的 API（如曾删 `ImportPlaceholder`）不算违反此约束。
3. **`internal/aep` 单 package + 文件名命名轴 + AST 边界守卫**。仍单 package，但**理由是 Go 语义**：公共 API 是 Stable 的方法式签名（`(p *Project) WriteAEP` / `(l *Layer) SetVisible` …），方法必须与类型同包；强拆 scene/serializer 包会 scene↔serializer 循环依赖。物理分包属 V3 M8（需先解 chunk 耦合或引接口隔离）。内部边界靠**文件名命名轴**维护：`<stage>_<domain>`，7 前缀 = `scene_`（运行时模型/accessor）、`codec_`（无 scene 耦合的纯字节/值编解码，方案②抽包预备）、`parse_`（chunk→scene）、`lower_`（scene→新 chunk）+ `write_`（发射字节/length-preserving patch）= serialize 阶段两子角色、`back_`（`*Backrefs` chunk 引用结构 = scene↔chunk 断层线）、`mutate_`（new/delete/insert/move/duplicate/sync 结构性操作；`Set*` 原地 patch 归 `write_`）。守卫 `internal/aep/arch_boundary_test.go`（AST）强制 `scene_` 禁 import rifx + `codec_` 禁引用 scene 类型；scene→rifx 残留见白名单（5 项，V3 M8 清零，详 `specs/2026-05-22-v3-direction.md`）。详 `specs/2026-05-30-aep-package-reorg-design.md`。
4. **嵌入资源目录命名复数**：`internal/aep/templates/`（非 `template/`）。Go `//go:embed` 限制资源必须在 package 同目录或子目录。
5. **Opaque preservation**（V2.2 教训）：parser 未解的 chunk 必须 byte-identical round-trip。scene types 携带 opaque shard，serializer 原位重发。任何 "regenerate from scene" 路径必须保留它，否则 AE 会 silent-drop。
6. **AE 接受 gate**：任何新结构性写路径（NewX / DeleteX / DuplicateX / V3 mutation API）必须跑 AE 2020 + AE 2025 双版本 ship-gate 才算 ship。详 `incident-reports/ae25-acceptance-gate.md` + `checklists/re-fixture.md` § GDI 自动化。

非显然内部不变量（gotcha 错题集）：
- TickRate per-composition，非全局 → `incident-reports/tickrate-per-composition.md`
- Keyframe 字节布局两种，必走 `layoutFor` → `incident-reports/keyframe-byte-layout-dispatcher.md`
- 所有 Set\* mutate 共享 chunk bytes，调用方自己锁 → `incident-reports/concurrency-unsafe-shared-chunk-bytes.md`
- chunk ID 大小写敏感（Tdb4 ≠ tdb4）→ `incident-reports/chunk-id-case-tdb4.md`

## 入口命令

- 校验: `go vet ./... && go test ./...`
- 单测: `go test ./internal/aep/ -run 'TestX' -v`
- 详细操作（tmp_debug 工具表 / fixture 验证 / ship-gate）: `flightdeck/checklists/`

## 工作风格

- **一律用中文跟用户交流**（含解释、提案、报告、commit body 可英文按既有惯例）
- 代码优先，设计讨论精简
- 重构 / 迁移**先读源码再动**，不瞎猜
- 不写注释，除非 WHY 不明显

## 文档地图

- 字段覆盖矩阵 + 暂搁 / 不可达 / negative findings: `flightdeck/flight-plans/coverage.md` + `coverage-detail.md`
- API 同步表（改任何 public API 必读）: `flightdeck/checklists/commits.md` § 命令一致性
- 测试惯例 / 验证流程 / tmp_debug 工具表: `flightdeck/checklists/verify.md`
- JSX RE 工作流 + ship-gate + RE fixture 双轨 + Types-for-Adobe 参考: `flightdeck/checklists/re-fixture.md`
- 当前里程碑: V2.1 (NewProject/Composition) + V2.2 (ShapeLayer) 已 ship；V3 (runtime IR + capability matrix) 规划中，详 `flightdeck/specs/2026-05-22-v3-direction.md`
