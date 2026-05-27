# Board — aep-parser

**Last updated**: 2026-05-27 by claude (doc sync — spec + coverage-detail parity table refresh)
**Active focus**: py-aep parity — P1/P2a/P2b all done; P2c+ or V3 next

## Next session

**Doc sync — spec + coverage-detail 全面刷新**：spec §2.1-2.5 表 30+ 行从 ❌ 更新到 ✅/done（P1 1D/B/E/F/G/H + P2a + P2b 已 ship 的全部标记）。coverage-detail.md 同步：Project 域 nnhd/CMS/setting chunks、Layer 域 convenience helpers、Property 域 tdb4 flag readers、Footage 域 P1 1H items、KeyframeEase 写回状态。PASS 220。

**待 user 决定 next direction**：
- **P2c**: PropertyGroup hierarchy (`layer.Effects()` etc.) — unblocked, 中等工作量
- **P2b Task 3**: Gradient XML — 等 user 提供 gradient fixture
- **P2b Task 4**: Property metadata — 需 pard chunk reader + specs.py schema port
- **V3**: runtime IR + capability framework — 方向性规划

**未提交残留**（待 user 处理或下次 batch commit）：
- 前一 session 大堆 session-exit 残留（CLAUDE.md 重构 / 4 新 scar / ship_gate_helpers / spec 更新 / ae_run plan moved to finish/ 等）—— 我**没动** WT 里这些，本次只提交了 P2b Task 1+2 代码 + 我新加的 board/coverage/docs 行。

**并行候选**（不阻塞 P2a）：V2.2.1 ShapeLayer 拓展 / V3 brainstorm。

## In flight

无。P1 全闭环、P2a 全闭环、P2b Task 1-2 ship（Task 3-4 deferred）。等 user 决定 next direction。

## Blockers

无。

## Deferred

- **V2.2.1 ShapeLayer 拓展**（Ellipse/Path/Stroke embed bytes / Fill Color 编码 RE / keyframe 持久化）— 跟 py-aep parity 并行；用户用 AE create fixture 后可起
- **V3 capability framework** — Layer.Remove/Duplicate/Move/PropertyBase 结构性 ops 都靠这套；进 Phase 3
- `environmentLayer` 360° 素材 / `ligature` OT liga 字体 / `maskFeatherFalloff` 位置未 RE
- Composition.SetRenderer / ldta 零值区 probe / Footage proxy / Project nhed 扩展 — 见 [coverage.md](plans/coverage.md) "剩余可探方向"
- **`linearizeWorkingSpace` ScriptingAPI quirk** — chunk byte 跟 AE 自己写一致但 ScriptingAPI 读不到 true，归 OCIO/CMS-联动；详 [scars/project-flag-chunks-lnrb-lnrp.md](scars/project-flag-chunks-lnrb-lnrp.md)

## Recently finished

- **2026-05-27 py-aep parity doc sync** — spec §2.1-2.5 表全面刷新：30+ 行从 ❌ 更新到 ✅/done（P1 1D setting chunks / P1 1B project views / P1 1E layer convenience / P1 1F comp convenience / P1 1G tdb4 flags / P1 1H footage convenience / P2a Tasks 1-5 / P2b Tasks 1-2）。coverage-detail.md 同步：Project 域 +12 行（nnhd/CMS/setting chunks/views）、Layer 域 +3 行（3DModel/ReplaceSource/LightSource 修正）、Property 域 LockedRatio + tdb4 flag 修正、Footage 域 +4 行（P1 1H items）、KeyframeEase 写回状态修正。PASS 220。
- **2026-05-27 P2a Task 1+2+3+4+5 ship** — ThreeDModelLayer R / LightSource R/W / LockedRatio R/W / ReplaceSource R/W / ImportPlaceholder** — Task 1: `LayerType3DModel` 枚举 + `inferLayerType` ldta byte `@0x83 == 0x05` 派发 + `Layer.IsThreeDModelLayer()` typed accessor。Task 2: `Layer.LightSource() / SetLightSource(target *Layer)` 镜像 py-aep `LightLayer.light_source`，底层走 ldta `@0x28`（与 AV `SourceID` 共用 slot），sentinel `0xFFFFFFFF` = 无源，6 个验证错误路径（non-light caller / self / Light-target / Camera-target / 3D-target / cross-comp）全 PASS。Task 3: `Property.LockedRatio() / SetLockedRatio(v bool)`；底层走 tdsb `@0x02 bit 4`；parser 新加 `Property.tdsb` 私有 ref；IDTdsb 常量加到 rifx。Task 4: `Layer.ReplaceSource(target AVItem, fixExpressions bool)`；底层走既有 `SetSource` 路径；fixExpressions=true 时记 warning。Task 5: `Project.ImportPlaceholder(name, width, height, frameRate, duration)`；NewComposition 同模式，原子 mutation + 警告回滚。新增 ~5 个 PASS test 函数，无 fixture 依赖（synthetic byte-dispatch + synthetic Composition+Layers）。docs/layer.md + coverage.md 同步。**未跑 ship-gate**（P2a 全闭环，P2b 待续）。
- **2026-05-27 ae_run.ps1 wrapper 全闭环 (Phase 5-6, Task 15-18)** — V2.1/V2.2 ship-gate Go 端走 `runAeRunShipGate(t, ...)` 共享 helper (`ship_gate_helpers_test.go`)，删 deadline polling loop（ps1 owns timeout）。**Cross-version smoke PASS**：`AE_SHIP_GATE=1 AE2020_EXE=".../AE 2025/AfterFX.exe" go test -run TestV2_1_AEShipGate_AE2020` — wrapper OCR 检测+自动消化 convert 对话框，ship-gate 通过。V2.2 ship-gate (用 Ellipse/Path/Stroke) t.Skip 标 V2.2.1 deferred（docs/shape.md:260 明确 silent-drop 限制）。playbook re-fixture.md § GDI 自动化 from planned → shipped。归档 plan 到 plans/finish/。PASS 202 不变。
- **2026-05-27 CLAUDE.md / architecture.md 大重构** — CLAUDE.md 62→47 行（删跟 workshop-workflow skill 重复内容：场景触发器 / board 维护规则 / 进来第一件事 / 工作风格 #4；加 § 数据流 + 4 个 scar 指针）。architecture.md 全删，内容拆：数据流→CLAUDE.md；4 个不变量（chunk-id case / TickRate per-comp / keyframe layout dispatcher / concurrency unsafe）→ 4 个新 scar；测试惯例 → playbooks/verify.md；RE 双轨 + Types-for-Adobe → playbooks/re-fixture.md；已知非完美区 → coverage.md。3 处 spec back-ref 修。
- **2026-05-27 ae_run.ps1 wrapper Phase 1-4 + Task 14 smoke PASS** — `scripts/` 4 文件齐 (AeRun.Lib.ps1 / ae_run.ps1 / ocr_helper.ps1 / ae_dialog_rules.json) + 29 Pester unit test PASS。Task 14 clean smoke 用 AE 2025 跑 `smoke_ae_run.jsx`，wrapper 自动 OCR 检测+消化 "崩溃修复选项" 对话框（ESC 拒安全模式），JSX 写 `.done(PASS)`。新增 2 scar：`pwsh-7-no-winrt` + `windows-media-ocr-cjk-glyph-spacing`。规则表 5 条。
## Hanging tasks

无。
