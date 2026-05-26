# Board — aep-parser

> 文档分工见 [../CLAUDE.md](../CLAUDE.md) 场景触发器表。本文件 = 现在在做啥 + 接下来 + 最近 5 条归档 + PASS count 单一权威源。

**Last updated**: 2026-05-26 by claude (P1 全闭环 + AE 2020/2025 ship-gate PASS — 9/9 task ship + lnrb/lnrp RE 修复 + scar；PASS = **202 / 0 FAIL**；+27 vs 起点 175；~80 新 API)
**Active focus**: P1 闭环 done — ready for batch review；下一程 Phase 2 起 plan

## Next session

1. **用户批量 review P1**：11 个 commit (`17e4d15` → `ca5e5a8`)，spec [`specs/2026-05-26-py-aep-parity-design.md`](specs/2026-05-26-py-aep-parity-design.md)，archived plan [`plans/finish/2026-05-26-py-aep-parity-p1-plan.md`](plans/finish/2026-05-26-py-aep-parity-p1-plan.md)
2. 起 Phase 2 plan：nnhd 全展开 + CMS + Gradient + 类型化 Sources + PropertyGroup 链式 (~5 大主题)。spec §3.Phase 2
3. （并行候选）`scripts/ae_run.ps1` 起 GDI 自动化 wrapper — 截图认对话框 + SendKeys 自动点掉 convert / save-changes 弹框，让 ship-gate 真正 unattended。playbook re-fixture.md "GDI / 屏幕截图自动化" 段落有路线图

## In flight

无 — Phase 1 闭环；等用户 review + P2 决策。

## Blockers

无。

## Deferred

- **V2.2.1 ShapeLayer 拓展**（Ellipse/Path/Stroke embed bytes / Fill Color 编码 RE / keyframe 持久化）— 跟 py-aep parity 并行；用户用 AE create fixture 后可起
- **V3 capability framework** — Layer.Remove/Duplicate/Move/PropertyBase 结构性 ops 都靠这套；进 Phase 3
- `environmentLayer` 360° 素材 / `ligature` OT liga 字体 / `maskFeatherFalloff` 位置未 RE
- Composition.SetRenderer / ldta 零值区 probe / Footage proxy / Project nhed 扩展 — 见 [coverage.md](plans/coverage.md) "剩余可探方向"
- **`linearizeWorkingSpace` ScriptingAPI quirk** — chunk byte 跟 AE 自己写一致但 ScriptingAPI 读不到 true，归 OCIO/CMS-联动；详 [scars/project-flag-chunks-lnrb-lnrp.md](scars/project-flag-chunks-lnrb-lnrp.md)

## Recently finished

- **2026-05-26 P1 1D AE ship-gate 闭环 + lnrb/lnrp RE 修复** — 第一版 lnrb/lnrp 让 AE 报"文件数据丢失"。bisect 隔出真凶 (8/8 variants)。AE-saved fixture 对比发现 chunk 是 1B `0x01` 非空, 位置紧跟 `cpid`. 修后 ship-gate AE 2020/2025 全 PASS, AE-visible 值跟 Go-side byte-identical. 新增 scar `project-flag-chunks-lnrb-lnrp.md` + playbook re-fixture.md ship-gate 章节 (3 失败模式表 + 版本匹配规则 + GDI 自动化路线图).
- **2026-05-26 py-aep parity Phase 1 全闭环 (PASS 175 → 202 +27, ~80 新 API)** — 9 个 task 全 ship: 1A CompItem filter (15) / 1F Comp convenience (3) / 1E Layer convenience (13) / 1I Application wrapper (4) / 1C Frame-time accessor (20) / 1G Property tdb4 flags (7) / 1H Footage convenience + sspc bug fix (6+) / 1B Project views + EffectNames (4) / 1D Project settings chunks (8). 修复 latent bug: 真实 AE sspc 222B 布局 Width/Height 在 @0x20/@0x24 (之前一直读 0).
- **2026-05-26 workshop 格式 v0.8.0 对齐 + py-aep parity 路线落地** — board.md 重组按 v0.8.0 模板；specs/plans 加 YYYY-MM-DD- 前缀；V2.1 入 finish/；新增 py-aep parity spec + Phase 1 plan（PASS 175 不变）。
- **2026-05-26 V2.2 Phase 6 docs sync + v0.8.0 frontmatter migration** — `docs/shape.md` 加 V2.2 alpha Builder API section + 限制声明；`coverage.md` + `v2-2 spec §6.4a/§6.5/§10` finalize；10 scars + 3 playbooks 全补 frontmatter。PASS 175 不变。
- **2026-05-25 V2.2 Phase 5 iter-8** — embed Rect+Fill body bytes（`v2_2_shape_rect_body.bin` 448B + `v2_2_shape_fill_body.bin` 426B）；variants #2-7 全 PASS layers=1。Phase 5 ship gate 完整闭环。PASS 175。

## Hanging tasks

无。
