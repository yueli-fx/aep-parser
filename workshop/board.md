# Board — aep-parser

> 文档分工见 [../CLAUDE.md](../CLAUDE.md) 场景触发器表。本文件 = 现在在做啥 + 接下来 + 最近 5 条归档 + PASS count 单一权威源。

**Last updated**: 2026-05-27 by claude (ae_run.ps1 wrapper Phase 1-4 完成 + Task 14 smoke PASS — 14/18 task ship；29 Pester unit test PASS；2 新 scar；PASS = **202 / 0 FAIL** 不变 (scripts-only))
**Active focus**: ae_run.ps1 unattended ship-gate wrapper — Phase 5 partial (Task 14 done)，剩 Task 15-18

## Next session

1. **Task 15 — Go drop-in `shape_layer_shipgate_test.go`**：把 `exec.Command(aeExe, "-r", jsxPath)` 替换为 ps1 invocation；删 deadline loop（ps1 owns timeout）。详 [`plans/2026-05-27-ae-run-wrapper-plan.md`](plans/2026-05-27-ae-run-wrapper-plan.md) Task 15
2. **Task 16 — Go drop-in `new_composition_test.go`**：同 Task 15 但 V2.1 path
3. **Task 17 — 跨版本 smoke**：`AE_SHIP_GATE=1 AE2020_EXE="...AE 2025/...AfterFX.exe" go test ./internal/aep/ -run TestV2_1_AEShipGate_AE2020` — 触发 convert 对话框，wrapper 应自动消化
4. **Task 18 — Playbook + board 闭环**：更新 `playbooks/re-fixture.md` § GDI 自动化（删 planned 标记），归档 plan 到 `plans/finish/`

## In flight

- **ae_run.ps1 wrapper** — Phase 1-4 done (scripts/ 4 files + 29 unit test)；Phase 5 partial (Task 14 ✅)；spec [`specs/2026-05-27-ae-run-wrapper-design.md`](specs/2026-05-27-ae-run-wrapper-design.md), plan [`plans/2026-05-27-ae-run-wrapper-plan.md`](plans/2026-05-27-ae-run-wrapper-plan.md)

## Blockers

无。

## Deferred

- **V2.2.1 ShapeLayer 拓展**（Ellipse/Path/Stroke embed bytes / Fill Color 编码 RE / keyframe 持久化）— 跟 py-aep parity 并行；用户用 AE create fixture 后可起
- **V3 capability framework** — Layer.Remove/Duplicate/Move/PropertyBase 结构性 ops 都靠这套；进 Phase 3
- `environmentLayer` 360° 素材 / `ligature` OT liga 字体 / `maskFeatherFalloff` 位置未 RE
- Composition.SetRenderer / ldta 零值区 probe / Footage proxy / Project nhed 扩展 — 见 [coverage.md](plans/coverage.md) "剩余可探方向"
- **`linearizeWorkingSpace` ScriptingAPI quirk** — chunk byte 跟 AE 自己写一致但 ScriptingAPI 读不到 true，归 OCIO/CMS-联动；详 [scars/project-flag-chunks-lnrb-lnrp.md](scars/project-flag-chunks-lnrb-lnrp.md)

## Recently finished

- **2026-05-27 ae_run.ps1 wrapper Phase 1-4 + Task 14 smoke PASS** — `scripts/` 4 文件齐 (AeRun.Lib.ps1 / ae_run.ps1 / ocr_helper.ps1 / ae_dialog_rules.json) + 29 Pester unit test PASS。Task 14 clean smoke 用 AE 2025 跑 `smoke_ae_run.jsx`，wrapper 自动 OCR 检测+消化 "崩溃修复选项" 对话框（ESC 拒安全模式），JSX 写 `.done(PASS)`。新增 2 scar：`pwsh-7-no-winrt`（pwsh 7 没 WinRT projection，需 powershell.exe 5.1 sub-shell）+ `windows-media-ocr-cjk-glyph-spacing`（OCR 把 CJK 字符空格分开 "修 复 选 项"，Match-Rule 加 whitespace-stripped fallback）。规则表 5 条（convert / save-changes / file-data-missing / perf-prefs / safe-mode-recovery）。PASS 202 不变 (scripts-only).
- **2026-05-26 P1 1D AE ship-gate 闭环 + lnrb/lnrp RE 修复** — 第一版 lnrb/lnrp 让 AE 报"文件数据丢失"。bisect 隔出真凶 (8/8 variants)。AE-saved fixture 对比发现 chunk 是 1B `0x01` 非空, 位置紧跟 `cpid`. 修后 ship-gate AE 2020/2025 全 PASS, AE-visible 值跟 Go-side byte-identical. 新增 scar `project-flag-chunks-lnrb-lnrp.md` + playbook re-fixture.md ship-gate 章节 (3 失败模式表 + 版本匹配规则 + GDI 自动化路线图).
- **2026-05-26 py-aep parity Phase 1 全闭环 (PASS 175 → 202 +27, ~80 新 API)** — 9 个 task 全 ship: 1A CompItem filter (15) / 1F Comp convenience (3) / 1E Layer convenience (13) / 1I Application wrapper (4) / 1C Frame-time accessor (20) / 1G Property tdb4 flags (7) / 1H Footage convenience + sspc bug fix (6+) / 1B Project views + EffectNames (4) / 1D Project settings chunks (8). 修复 latent bug: 真实 AE sspc 222B 布局 Width/Height 在 @0x20/@0x24 (之前一直读 0).
- **2026-05-26 workshop 格式 v0.8.0 对齐 + py-aep parity 路线落地** — board.md 重组按 v0.8.0 模板；specs/plans 加 YYYY-MM-DD- 前缀；V2.1 入 finish/；新增 py-aep parity spec + Phase 1 plan（PASS 175 不变）。
- **2026-05-26 V2.2 Phase 6 docs sync + v0.8.0 frontmatter migration** — `docs/shape.md` 加 V2.2 alpha Builder API section + 限制声明；`coverage.md` + `v2-2 spec §6.4a/§6.5/§10` finalize；10 scars + 3 playbooks 全补 frontmatter。PASS 175 不变。

## Hanging tasks

无。
