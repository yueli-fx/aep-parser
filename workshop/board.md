# Board — aep-parser

**Last updated**: 2026-05-27 by claude (P2a Task 1+2+3+4 ship — ThreeDModelLayer R / LightSource R/W / LockedRatio R/W / ReplaceSource R/W；Task 5 留下次)
**Active focus**: py-aep parity P2a — Task 5 ImportPlaceholder 继续

## Next session

**继续 P2a — Task 5 ImportPlaceholder**：详 [`plans/2026-05-27-py-aep-parity-p2a-plan.md`](plans/2026-05-27-py-aep-parity-p2a-plan.md) § Task 5。

进度（本次会话）：
- ✅ **Task 1 (2J)** — `LayerType3DModel` 枚举 + `inferLayerType` byte 0x05 派发 + `Layer.IsThreeDModelLayer()` accessor + synthetic byte-dispatch test。
- ✅ **Task 2 (2I)** — `Layer.LightSource() / SetLightSource(target *Layer)`（AE 24+ Environment-type light）；底层走 ldta `@0x28`（与 AV `SourceID` 共用 slot 按 Type 重解释），sentinel `0xFFFFFFFF` = 无源。py-aep `LightLayer.light_source` parity 含 6 个 validation 错误路径 test。
- ✅ **Task 3 (2D)** — `Property.LockedRatio() / SetLockedRatio(v bool)`；底层走 tdsb `@0x02 bit 4`；parser 新加 `Property.tdsb` 私有 ref；IDTdsb 常量加到 rifx。length-preserving，roundtrip test + fallback test 全 PASS。
- ✅ **Task 4 (2F)** — `Layer.ReplaceSource(target AVItem, fixExpressions bool)`；底层走既有 `SetSource` 路径；fixExpressions=true 时记 warning。roundtrip + warning test 全 PASS。
- ⏸ **Task 5** — 未开工。

**未提交残留**（待 user 处理或下次 batch commit）：
- 前一 session 大堆 session-exit 残留（CLAUDE.md 重构 / 4 新 scar / ship_gate_helpers / spec 更新 / ae_run plan moved to finish/ 等）—— 我**没动** WT 里这些，本次只提交了 Task 1+2 代码 + 我新加的 board/coverage/docs 行。
- 注意 `workshop/plans/coverage.md` 的 commit 含 prior session 2 行未提交内容（4D 32bpc / mkif 残余字节 negative findings），跟我的 P2a 1+2 row 混在一起。

**并行候选**（不阻塞 P2a）：V2.2.1 ShapeLayer 拓展 / V3 brainstorm。

## In flight

- **py-aep parity P2a** — plan [`plans/2026-05-27-py-aep-parity-p2a-plan.md`](plans/2026-05-27-py-aep-parity-p2a-plan.md)（ready to execute）；spec [`specs/2026-05-26-py-aep-parity-design.md`](specs/2026-05-26-py-aep-parity-design.md) § 2.1-2.5

## Blockers

无。

## Deferred

- **V2.2.1 ShapeLayer 拓展**（Ellipse/Path/Stroke embed bytes / Fill Color 编码 RE / keyframe 持久化）— 跟 py-aep parity 并行；用户用 AE create fixture 后可起
- **V3 capability framework** — Layer.Remove/Duplicate/Move/PropertyBase 结构性 ops 都靠这套；进 Phase 3
- `environmentLayer` 360° 素材 / `ligature` OT liga 字体 / `maskFeatherFalloff` 位置未 RE
- Composition.SetRenderer / ldta 零值区 probe / Footage proxy / Project nhed 扩展 — 见 [coverage.md](plans/coverage.md) "剩余可探方向"
- **`linearizeWorkingSpace` ScriptingAPI quirk** — chunk byte 跟 AE 自己写一致但 ScriptingAPI 读不到 true，归 OCIO/CMS-联动；详 [scars/project-flag-chunks-lnrb-lnrp.md](scars/project-flag-chunks-lnrb-lnrp.md)

## Recently finished

- **2026-05-27 P2a Task 1+2+3+4 ship — ThreeDModelLayer R / LightSource R/W / LockedRatio R/W / ReplaceSource R/W** — Task 1: `LayerType3DModel` 枚举 + `inferLayerType` ldta byte `@0x83 == 0x05` 派发 + `Layer.IsThreeDModelLayer()` typed accessor。Task 2: `Layer.LightSource() / SetLightSource(target *Layer)` 镜像 py-aep `LightLayer.light_source`，底层走 ldta `@0x28`（与 AV `SourceID` 共用 slot），sentinel `0xFFFFFFFF` = 无源，6 个验证错误路径（non-light caller / self / Light-target / Camera-target / 3D-target / cross-comp）全 PASS。Task 3: `Property.LockedRatio() / SetLockedRatio(v bool)`；底层走 tdsb `@0x02 bit 4`；parser 新加 `Property.tdsb` 私有 ref；IDTdsb 常量加到 rifx。Task 4: `Layer.ReplaceSource(target AVItem, fixExpressions bool)`；底层走既有 `SetSource` 路径；fixExpressions=true 时记 warning。新增 ~4 个 PASS test 函数，无 fixture 依赖（synthetic byte-dispatch + synthetic Composition+Layers）。docs/layer.md + coverage.md 同步。**未跑 ship-gate**（P2a 仅 Task 5 需要）。
- **2026-05-27 ae_run.ps1 wrapper 全闭环 (Phase 5-6, Task 15-18)** — V2.1/V2.2 ship-gate Go 端走 `runAeRunShipGate(t, ...)` 共享 helper (`ship_gate_helpers_test.go`)，删 deadline polling loop（ps1 owns timeout）。**Cross-version smoke PASS**：`AE_SHIP_GATE=1 AE2020_EXE=".../AE 2025/AfterFX.exe" go test -run TestV2_1_AEShipGate_AE2020` — wrapper OCR 检测+自动消化 convert 对话框，ship-gate 通过。V2.2 ship-gate (用 Ellipse/Path/Stroke) t.Skip 标 V2.2.1 deferred（docs/shape.md:260 明确 silent-drop 限制）。playbook re-fixture.md § GDI 自动化 from planned → shipped。归档 plan 到 plans/finish/。PASS 202 不变。
- **2026-05-27 CLAUDE.md / architecture.md 大重构** — CLAUDE.md 62→47 行（删跟 workshop-workflow skill 重复内容：场景触发器 / board 维护规则 / 进来第一件事 / 工作风格 #4；加 § 数据流 + 4 个 scar 指针）。architecture.md 全删，内容拆：数据流→CLAUDE.md；4 个不变量（chunk-id case / TickRate per-comp / keyframe layout dispatcher / concurrency unsafe）→ 4 个新 scar；测试惯例 → playbooks/verify.md；RE 双轨 + Types-for-Adobe → playbooks/re-fixture.md；已知非完美区 → coverage.md。3 处 spec back-ref 修。
- **2026-05-27 ae_run.ps1 wrapper Phase 1-4 + Task 14 smoke PASS** — `scripts/` 4 文件齐 (AeRun.Lib.ps1 / ae_run.ps1 / ocr_helper.ps1 / ae_dialog_rules.json) + 29 Pester unit test PASS。Task 14 clean smoke 用 AE 2025 跑 `smoke_ae_run.jsx`，wrapper 自动 OCR 检测+消化 "崩溃修复选项" 对话框（ESC 拒安全模式），JSX 写 `.done(PASS)`。新增 2 scar：`pwsh-7-no-winrt` + `windows-media-ocr-cjk-glyph-spacing`。规则表 5 条。
- **2026-05-26 P1 1D AE ship-gate 闭环 + lnrb/lnrp RE 修复** — 第一版 lnrb/lnrp 让 AE 报"文件数据丢失"。bisect 隔出真凶 (8/8 variants)。AE-saved fixture 对比发现 chunk 是 1B `0x01` 非空, 位置紧跟 `cpid`. 修后 ship-gate AE 2020/2025 全 PASS, AE-visible 值跟 Go-side byte-identical. 新增 scar `project-flag-chunks-lnrb-lnrp.md` + playbook re-fixture.md ship-gate 章节 (3 失败模式表 + 版本匹配规则 + GDI 自动化路线图).
## Hanging tasks

无。
