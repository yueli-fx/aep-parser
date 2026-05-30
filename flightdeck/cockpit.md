# Cockpit — aep-parser

**Last updated**: 2026-05-30 by claude（`internal/aep` package 重组 **landed on main**：`<stage>_<domain>` 7 前缀命名轴（scene_/codec_/parse_/lower_/write_/back_/mutate_）+ 拆 types_core/layer_accessors + 测试按 feature 拆 + AST 边界守卫 `arch_boundary_test.go`（scene_ 禁 import rifx / codec_ 禁 scene 类型）。零行为变更：byte-identical 115 fixture + API-set 不变 + AE 双版本 ship-gate 24/24 PASS。详见 logbook + `landed/specs/2026-05-30-aep-package-reorg-design.md`。）
**Active focus**: 无 active 实现线。

## Next session

1. **注释纪律清理 pass**（comments.md §6，**单独分支**）：`internal/aep` 源文件约 **233 处** §3 违规（`spec §` / `Phase N` / `Inv-N` / `iter N` / 日期戳 / 历史考古等），系预存债（重组只原样搬运注释）。跨 ~30 文件，逐条删/改写/搬 commit-msg；按 comments.md §6 grep 收口。

> **长线 backlog**（大 arc 或缺 runtime setter）：

1. **Path keyframe**（逐帧 bezier）— V2.3+ 级大 arc。
2. **Layr Transform 3D 通道**（Orientation / Rotate X/Y / Position_Z）— 需先有 3D layer 支持（runtime 无 3D switch，V2.3）。
3. **Stroke Line Cap/Join/Miter**（enum/scalar）— 中等价值。groundwork：stroke body **不含**这些 slot（默认值被 elide），matchName **不是** `ADBE Vector Stroke Line Cap`（JSX property-not-found）。需先查真实 matchName + 富化 stroke body + 加 runtime model 字段/enum/setter。
4. **Gradient W**（SetGradient）— 大 arc。groundwork：默认 gradient 被 AE elide（须自定义 stops 强制 emit）；无现成 in-repo fixture；存储 `GCst > GCky > Utf8(prop.map XML)`（parse_properties.go）。需 gradient-fill fixture + XML RE + 序列化器 + API + 双版本 gate。
5. **Fill/Stroke BlendMode·CompositeOrder、Rect/Ellipse Direction** — enum，低价值，缺 runtime setter。

**其它候选**：泛型 `DuplicateItem`（无 scripting API）、`ImportComposition`（需求驱动）。其余 deferred R-only（DisplayColorSpace / ValueText 等）见 logbook § Deferred。

## Hanging tasks

无。
