# Board — aep-parser

> 文档分工见 [../CLAUDE.md](../CLAUDE.md) 场景触发器表。本文件 = 现在在做啥 + 接下来 + 最近 5 条归档 + PASS count 单一权威源。

**Last updated**: 2026-05-26 by claude (py-aep parity Phase 1 全闭环 — 9/9 task ship；PASS = **202 / 0 FAIL**；+27 PASS vs 起点 175；~80 新 API)
**Active focus**: py-aep parity Phase 1 done — 等用户批量 review；下一程 Phase 2

## Next session

1. **用户批量 review P1** — 9 个 commit (`17e4d15` → `17e2391`)，spec [`specs/2026-05-26-py-aep-parity-design.md`](specs/2026-05-26-py-aep-parity-design.md)，plan [`plans/2026-05-26-py-aep-parity-p1-plan.md`](plans/2026-05-26-py-aep-parity-p1-plan.md)
2. **AE 2020/2025 ship gate 1D**: 用户开 AE 跑 `re_cameralight.aep` 修改 setting 后 save → re-open 校验 AE 接受 (lnrb / ExpressionEngine / AudioSampleRate 等)
3. P2 plan 起：nnhd 全展开 + CMS + Gradient + 类型化 Sources + PropertyGroup 链式 (~5 phase 主题)。spec [`specs/2026-05-26-py-aep-parity-design.md`](specs/2026-05-26-py-aep-parity-design.md) §3.Phase 2

## In flight

无 — Phase 1 闭环；等用户 review + P2 决策。

## Blockers

- **AE ship gate 1D 待用户操作** — 8 个 Project setting setter 全 roundtrip-tested Go-side，AE 接受验证留下次开 AE 时跑

## Deferred

- **V2.2.1 ShapeLayer 拓展**（Ellipse/Path/Stroke embed bytes / Fill Color 编码 RE / keyframe 持久化）— 跟 py-aep parity 并行；用户用 AE create fixture 后可起
- **V3 capability framework** — Layer.Remove/Duplicate/Move/PropertyBase 结构性 ops 都靠这套；进 Phase 3
- `environmentLayer` 360° 素材 / `ligature` OT liga 字体 / `maskFeatherFalloff` 位置未 RE
- Composition.SetRenderer / ldta 零值区 probe / Footage proxy / Project nhed 扩展 — 见 [coverage.md](plans/coverage.md) "剩余可探方向"

## Recently finished

- **2026-05-26 py-aep parity Phase 1 全闭环 (PASS 175 → 202 +27, ~80 新 API)** — 9 个 task 全 ship: 1A CompItem filter (15) / 1F Comp convenience (3) / 1E Layer convenience (13) / 1I Application wrapper (4) / 1C Frame-time accessor (20) / 1G Property tdb4 flags (7) / 1H Footage convenience + sspc bug fix (6+) / 1B Project views + EffectNames (4) / 1D Project settings chunks (8). 修复 latent bug: 真实 AE sspc 222B 布局 Width/Height 在 @0x20/@0x24 (之前一直读 0).
- **2026-05-26 workshop 格式 v0.8.0 对齐 + py-aep parity 路线落地** — board.md 重组按 v0.8.0 模板；specs/plans 加 YYYY-MM-DD- 前缀；V2.1 入 finish/；新增 py-aep parity spec + Phase 1 plan（PASS 175 不变）。
- **2026-05-26 V2.2 Phase 6 docs sync + v0.8.0 frontmatter migration** — `docs/shape.md` 加 V2.2 alpha Builder API section + 限制声明；`coverage.md` + `v2-2 spec §6.4a/§6.5/§10` finalize；10 scars + 3 playbooks 全补 frontmatter。PASS 175 不变。
- **2026-05-25 V2.2 Phase 5 iter-8** — embed Rect+Fill body bytes（`v2_2_shape_rect_body.bin` 448B + `v2_2_shape_fill_body.bin` 426B）；variants #2-7 全 PASS layers=1。Phase 5 ship gate 完整闭环。PASS 175。
- **2026-05-25 V2.2 Phase 5 iter-7** — transplant 法 isolate + embed Transform Group bytes（`v2_2_transform_group_body.bin` 1842B）；variant #2 empty ShapeLayer 首次 PASS layers=1，跨第一道 silent-drop 闸门。

## Hanging tasks

无。
