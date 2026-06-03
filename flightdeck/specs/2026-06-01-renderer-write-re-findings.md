---
status: idea
summary: Composition.Renderer W — RE findings（prin 104B / prda 变长，结构性，待实现）
---

# Composition.Renderer W — RE findings (P3, pre-implementation)

**Created**: 2026-06-01（RQ reader/writer arc 收尾时顺手探的）

下一个 P3 结构性候选 `Composition.Renderer` 写。R 已有（`parse_composition.go`：`PRin` LIST → `prin` chunk @offset 4 的 ASCII NUL-padded match-name）。写之前先把 RE 探明记下，省得下次重探。

## chunk 布局

comp 的渲染器存在 Item 下的 `LIST:PRin` 里：
- `prin` chunk：**固定 104B**，renderer match-name 是 ASCII NUL-padded，**@offset 4**。改名是 length-preserving。
- `prda` chunk：renderer-specific 选项，**长度随 renderer 变**：

| renderer (内部名) | UI | prda len |
|---|---|---|
| `ADBE Escher` | Classic 3D | 12 |
| `ADBE Calder` | Advanced 3D | 52 |
| `ADBE Ernst` | Cinema 4D | 20 |
| `ADBE Picasso` | Ray-traced 3D | 16 |

（实测自 py-aep `samples/models/composition/renderer_{classic_3d,advanced_3d,cinema_4d,ray_traced}.aep`，各自单 comp。）

## 结论：Renderer W 是结构性写

换 renderer = 改 prin 名（定长）+ **替换变长 prda chunk** → 父 `PRin` LIST size 变 → **结构性**。按 CLAUDE.md 硬约束 #6 必须跑 AE 2020 + 2025 双版本 ship-gate 才算 ship。

## 实现路线（待执行）

1. 抽 4 个 renderer 的 prda byte 模板（直接从上面 4 个 fixture dump）→ Go table（类似 templates 内嵌）。
2. `(c *Composition) SetRenderer(name string)`：
   - prin @4 写新 match-name（NUL-pad 到 104B 内）
   - 用目标 renderer 的 prda 模板替换现 prda chunk（rifx 自动重算父 LIST size）
   - 走 V2.1 atomic invariants：snapshot + warnings-as-failure + rollback
3. JSX RE fixture：`comp.renderer = "ADBE Escher"` 等，AE 2020+2025 双开校验接受 + readback。

## 开放问题 — 已由 ship-gate RE 钉死（2026-06-01）

实现了 `(c *Composition) SetRenderer(matchName)`（Alpha，commit `420020c`），跑 AE 2025 + AE 2020 ship-gate 后拿到 ground truth：

**1. prin body 结构**：固定 104B，@0..3 常量 `00000000`；match-name NUL-pad @4（字段 [4,52)）；display name NUL-pad @0x34（[52,96)）；@96..103 常量 trailer。**仅两个名字字段 renderer/locale 相关，其余常量** → prin 改名 length-preserving + opaque-preserving（只 patch 两字段，保留版本常量区）。

**2. ⚠ match-name 不是跨版本稳定的引擎标识（核心坑）**：同一引擎在不同 AE 版本存的 prin match-name 完全不同，display name 还是 locale 相关（GBK 中文 vs 英文）：

| 引擎 | AE 2020 scripting | AE 2020 存储 match-name | AE 2020 prda | AE 2025 match-name | AE 2025 prda |
|---|---|---|---|---|---|
| Advanced 3D | `ADBE Advanced 3d` | `ADBE Escher` | 12B | `ADBE Calder` | 52B |
| Classic 3D | `ADBE Standard 3d` | `ABDE 3D DURER` | 12B | `ADBE Escher`(py-aep) | 12B |
| Cinema 4D | `ADBE Ernst` | `ADBE Ernst` | **20B** | `ADBE Ernst` | **20B** |
| Ray-traced 3D | （AE 2020 无此引擎，仅 3 个） | — | — | `ADBE Picasso` | 16B |

- AE 2020 只暴露 3 个 renderer（`ADBE Advanced 3d|ADBE Standard 3d|ADBE Ernst`）；AE 2025 暴露 `ADBE Advanced 3d|ADBE Calder|ADBE Ernst`。
- **`ADBE Ernst`（Cinema 4D）是唯一 match-name + prda（20B）跨版本都一致的引擎**。
- AE 2025 把废弃引擎自动提升：写 `ADBE Escher`/`ADBE Picasso` → readback `ADBE Advanced 3d`；写 `ADBE Calder`/`ADBE Ernst` 精确 round-trip。
- prda 版本相关（Advanced：2020=12B vs 2025=52B）。
- 同一个名 `ADBE Escher` 在 AE 2020=Advanced、在 py-aep(AE2025-era)=Classic display → **名不可信，display name 才提示真引擎，但 display 又 locale 相关**。

**3. ship-gate 结果**：结构写机制（prin 改名 + prda 替换 + WriteAEP size 重算）在 **AE 2020 + AE 2025 都被接受**（开不崩、不损坏、PRin 存活）。AE 2025 全 4 模板 accept（4/4 PASS，Calder/Ernst 精确、Escher/Picasso→Advanced）；AE 2020 验证了跨版本稳定的 Ernst（base=AE2020-native，PASS）。

**4. 当前 Alpha 的局限**：模板表是 **AE-2025-only**（4 个 2025 match-name + 2025 prda）。对 AE 2020 仅 `ADBE Ernst` 正确；`ADBE Calder/Escher/Picasso` 不是 AE 2020 引擎。要全版本正确需 **version-aware 引擎 enum + per-version 模板矩阵**（按 Project 的 AE 版本选模板），是更大 arc。

## 决策：已解决 — 不需要版本矩阵（2026-06-01）

最初误判成「match-name 版本相关 = 需要大重构」。查 py-aep `composition.py` 后纠正（用户指正"看 py-aep 写法"）：

```python
# py-aep: binary prin 存 match_name，ExtendScript 暴露另一套 module name
_RENDERER_BINARY_TO_EXTENDSCRIPT = {
    "ADBE Escher": "ADBE Advanced 3d",  # 唯一有别名
    "ADBE Calder": "ADBE Calder", "ADBE Ernst": "ADBE Ernst", "ADBE Picasso": "ADBE Picasso",
}
```

- **binary match_name（4 个画家代号）是稳定引擎身份**；ExtendScript 名只是表层别名（仅 Escher↔"ADBE Advanced 3d" 不同）。我那个"AE 2020 写 Advanced 3d 存成 Escher"的"矛盾"正是这条映射，非 bug。
- prda 模板按 binary 代号 keyed（Escher 12B/Calder 52B/Ernst 20B/Picasso 16B），**引擎稳定，非版本相关**。「Advanced 2020=12B vs 2025=52B」是因为 2020 的 Advanced=Escher、2025 的 Advanced=Calder（两个不同 binary 引擎）。
- display_name 只 locale/cosmetic（中文 GBK 经典/标准 vs 英文 Classic/Standard），AE 自己重导，无关紧要。
- 各 AE 版本暴露哪些 binary 引擎不同：AE 2020 = Escher(Advanced)/Standard-3d/Ernst；AE 2025 = Calder(Advanced)/Ernst/Picasso，且 load 时把废弃 Escher/Picasso 自动提升为 Advanced 3D。

**落地**（commit 见 git log）：`SetRenderer(name)` 接受 binary 或 ExtendScript 名（py-aep 别名归一），改 match_name+display(length-preserving)+换 prda(structural)。ship-gate **AE 2025 4/4 + AE 2020 (Ernst+Escher) 全绿**，满足硬约束 #6。剩 deferred：DimensionsSeparated / RQ 结构性增删等其它 P3 项。

## 相关

- 现有 R：`parse_composition.go` § Renderer（PRin→prin@4）+ `composition_display_test.go::TestCompositionRendererReal`（注意：本地 `re_renderer.aep` 是 9KB 极简 fixture，无 prin/prda，该测试在本机 vacuous pass）
- 写机制参考：V3 结构性 ops（DeleteLayer/MoveLayer 等）的 snapshot+rollback 模式
- spec §2.2 Renderer 行
