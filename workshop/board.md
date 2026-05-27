# Board — aep-parser

**Last updated**: 2026-05-27 by claude (P2c PropertyGroup hierarchy MVP ship)
**Active focus**: py-aep parity — P1 / P2a (sans Task 5) / P2b (sans DisplayColorSpace) / P2c (MVP) all ship

## Next session

**待 user 决定 next direction**：
- **P2c followup**: Property metadata reads — `DefaultValue / LastValue / NbOptions / MinValue / MaxValue / UnitsText / PropertyControlType / PropertyValueType / PropertyIndex / PropertyDepth / ParentProperty` — 需 tdb4 字节 RE 或 schema port
- **P2b Task 3**: Gradient XML — 等 user 提供 gradient fixture
- **P2b Task 4**: Property metadata — 跟 P2c followup 重叠
- **V3**: runtime IR + capability framework — 方向性规划

**并行候选**：V2.2.1 ShapeLayer 拓展 / V3 brainstorm。

## In flight

无。

## Blockers

无。

## Deferred

- **V2.2.1 ShapeLayer 拓展**（Ellipse/Path/Stroke embed bytes / Fill Color 编码 RE / keyframe 持久化）— 跟 py-aep parity 并行；用户用 AE create fixture 后可起
- **V3 capability framework** — Layer.Remove/Duplicate/Move/PropertyBase 结构性 ops 都靠这套；进 Phase 3
- `environmentLayer` 360° 素材 / `ligature` OT liga 字体 / `maskFeatherFalloff` 位置未 RE
- Composition.SetRenderer / ldta 零值区 probe / Footage proxy / Project nhed 扩展 — 见 [coverage.md](plans/coverage.md) "剩余可探方向"
- **`linearizeWorkingSpace` ScriptingAPI quirk** — chunk byte 跟 AE 自己写一致但 ScriptingAPI 读不到 true，归 OCIO/CMS-联动；详 [scars/project-flag-chunks-lnrb-lnrp.md](scars/project-flag-chunks-lnrb-lnrp.md)

## Recently finished

- **2026-05-27 P2c PropertyGroup hierarchy MVP** — 新 `AEPropertyGroup` 类型 + `PropertyBase` interface (注：跟 V2.2 ShapeLayer 的 `PropertyGroup` escape-hatch 类型同名冲突，故加 `AE` 前缀)。parser 在原 flat 输出（`Layer.Properties / Effects / Markers / Masks`）旁边并行构建 tdgp 层级镜像，叶子用 tdbs chunk identity 解析到 *同一个 `*Property` instance*（指针一致，flat 与 tree 视图变更互通）。Layer API: `PropertyTree() / PropertyGroupByMatchName(name) / PropertyByPath(...matchNames)` + 11 typed group accessor（Transform / Audio / LayerStyles / EffectsParade / MaskParade / TextProperties / CameraOptions / LightOptions / MaterialOptions / GeometryOptions / ShapeContents）。Group API: `Property(matchName) / Group(matchName) / ChildByIndex(i) / NumProperties / ParentGroup / PropertyByPath`。5 新 PASS 测试 (cameralight fixture / effects-parade mirror / standalone nil-safe / not-found graceful / leaf identity)。PASS 224 (+5)，go vet clean。docs 同步：spec §2.3+§2.4 PropertyGroup ops 表行从 ❌/🟢 → ✅ done。**剩余 P2c followup**: `DefaultValue / LastValue / NbOptions / MinValue / MaxValue / UnitsText / PropertyControlType / PropertyValueType / PropertyIndex / PropertyDepth / ParentProperty` — 需 tdb4 RE 或 schema port。
- **2026-05-27 P2a/P2b review fixup**（reviewer pass）— 删 `Project.ImportPlaceholder` (opti tag AE 拒收，仅合成字节)、`DisplayColorSpace()` stub (永远返回 "None")；CMS setters 在 `cmsUtf8 == nil` 时拒写而非自动创建 (chunk 容器位置未 RE)；`SetColorManagementSystem`/`SetLutInterpolationMethod` 加 enum 校验；`cmsSettings` JSON 解析失败时 emit `Warnings`；新内部测试覆盖 LockedRatio reader/writer + CMS enum 校验 + JSON 解析错误路径；删手写 `contains`/`findSubstring`（重复两份）改 `strings.Contains`。**测试 219 PASS / 7 SKIP / 0 FAIL**（baseline 220 → 219；删 4 placeholder tests + 8 validation subtests + 加 4 新测试，含 subtests 净 -8）。docs 同步：spec / coverage / coverage-detail / p2a-plan / p2b-plan 全部 mark Task 5 + DisplayColorSpace 为 🗑️ deferred。
- **2026-05-27 py-aep parity doc sync** — spec §2.1-2.5 表全面刷新：30+ 行从 ❌ 更新到 ✅/done（P1 1D setting chunks / P1 1B project views / P1 1E layer convenience / P1 1F comp convenience / P1 1G tdb4 flags / P1 1H footage convenience / P2a Tasks 1-5 / P2b Tasks 1-2）。coverage-detail.md 同步：Project 域 +12 行（nnhd/CMS/setting chunks/views）、Layer 域 +3 行（3DModel/ReplaceSource/LightSource 修正）、Property 域 LockedRatio + tdb4 flag 修正、Footage 域 +4 行（P1 1H items）、KeyframeEase 写回状态修正。PASS 220。
- **2026-05-27 P2a Task 1+2+3+4+5 ship** — ThreeDModelLayer R / LightSource R/W / LockedRatio R/W / ReplaceSource R/W / ImportPlaceholder** — Task 1: `LayerType3DModel` 枚举 + `inferLayerType` ldta byte `@0x83 == 0x05` 派发 + `Layer.IsThreeDModelLayer()` typed accessor。Task 2: `Layer.LightSource() / SetLightSource(target *Layer)` 镜像 py-aep `LightLayer.light_source`，底层走 ldta `@0x28`（与 AV `SourceID` 共用 slot），sentinel `0xFFFFFFFF` = 无源，6 个验证错误路径（non-light caller / self / Light-target / Camera-target / 3D-target / cross-comp）全 PASS。Task 3: `Property.LockedRatio() / SetLockedRatio(v bool)`；底层走 tdsb `@0x02 bit 4`；parser 新加 `Property.tdsb` 私有 ref；IDTdsb 常量加到 rifx。Task 4: `Layer.ReplaceSource(target AVItem, fixExpressions bool)`；底层走既有 `SetSource` 路径；fixExpressions=true 时记 warning。Task 5: `Project.ImportPlaceholder(name, width, height, frameRate, duration)`；NewComposition 同模式，原子 mutation + 警告回滚。新增 ~5 个 PASS test 函数，无 fixture 依赖（synthetic byte-dispatch + synthetic Composition+Layers）。docs/layer.md + coverage.md 同步。**未跑 ship-gate**（P2a 全闭环，P2b 待续）。
- **2026-05-27 ae_run.ps1 wrapper 全闭环 (Phase 5-6, Task 15-18)** — V2.1/V2.2 ship-gate Go 端走 `runAeRunShipGate(t, ...)` 共享 helper (`ship_gate_helpers_test.go`)，删 deadline polling loop（ps1 owns timeout）。**Cross-version smoke PASS**：`AE_SHIP_GATE=1 AE2020_EXE=".../AE 2025/AfterFX.exe" go test -run TestV2_1_AEShipGate_AE2020` — wrapper OCR 检测+自动消化 convert 对话框，ship-gate 通过。V2.2 ship-gate (用 Ellipse/Path/Stroke) t.Skip 标 V2.2.1 deferred（docs/shape.md:260 明确 silent-drop 限制）。playbook re-fixture.md § GDI 自动化 from planned → shipped。归档 plan 到 plans/finish/。PASS 202 不变。

## Hanging tasks

无。
