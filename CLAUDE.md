# aep-parser — AI 协作规范

Go 实现的 Adobe After Effects `.aep` 二进制解析器，对照 boltframe/aftereffects-aep-parser 重写 + 增量。**读取下限 AE 2020**。

> workshop 约定见 workshop-workflow skill；新会话第一件事按 skill entry checklist 跑（先读 `workshop/board.md`）。

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

1. **写回默认 length-preserving**。改字段不准动 chunk 大小。少数 length-variable 例外（name / comment / expression / 字体名 / 文本字符串）`WriteAEP` 会重算父 LIST size + 内嵌 LIST btdk size header。
2. **public API 不动**。重构 / 拆文件 / 移函数都不能改 exported 类型 / 方法签名 / JSON 输出字段。
3. **`internal/aep` 单 package**。不引子包（会强制 API 重排）。
4. **嵌入资源目录命名复数**：`internal/aep/templates/`（非 `template/`）。Go `//go:embed` 限制资源必须在 package 同目录或子目录。

非显然内部不变量（gotcha 错题集）：
- TickRate per-composition，非全局 → `scars/tickrate-per-composition.md`
- Keyframe 字节布局两种，必走 `layoutFor` → `scars/keyframe-byte-layout-dispatcher.md`
- 所有 Set\* mutate 共享 chunk bytes，调用方自己锁 → `scars/concurrency-unsafe-shared-chunk-bytes.md`
- chunk ID 大小写敏感（Tdb4 ≠ tdb4）→ `scars/chunk-id-case-tdb4.md`

## 入口命令

- 校验: `go vet ./... && go test ./...`
- 单测: `go test ./internal/aep/ -run 'TestX' -v`
- 详细操作（tmp_debug 工具表 / fixture 验证 / ship-gate）: `workshop/playbooks/`

## 工作风格

- 代码优先，设计讨论精简
- 重构 / 迁移**先读源码再动**，不瞎猜
- 不写注释，除非 WHY 不明显

## 文档地图

- 字段覆盖矩阵 + 暂搁 / 不可达 / negative findings: `workshop/plans/coverage.md` + `coverage-detail.md`
- API 同步表（改任何 public API 必读）: `workshop/playbooks/commits.md` § 命令一致性
- 测试惯例 / 验证流程 / tmp_debug 工具表: `workshop/playbooks/verify.md`
- JSX RE 工作流 + ship-gate + RE fixture 双轨 + Types-for-Adobe 参考: `workshop/playbooks/re-fixture.md`
- 当前里程碑: V2.1 (NewProject/Composition) + V2.2 (ShapeLayer) 已 ship；V3 (runtime IR + capability matrix) 规划中，详 `workshop/specs/2026-05-22-v3-direction.md`
