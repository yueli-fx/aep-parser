# AEP 格式与工程知识 / Format and engineering knowledge

这里公开了 151 篇专题笔记，保留字节布局、属性行为、版本差异、调试方法与验证证据。它们是研究与实现资料；具体可用接口以当前代码和能力矩阵为准。

151 topic notes covering byte layouts, property behavior, version differences, debugging, and verification. These are research and implementation references; consult current code and the capability matrix for supported interfaces.

## 建议阅读顺序 / Reading paths

- **理解工程 / Understand a project**：[分析流程](techniques/understand-a-project.md) → [通用技法](techniques/fx-techniques.md) → [火焰](techniques/build-good-fire.md)、[glitch](techniques/build-good-glitch.md)。
- **理解实现 / Understand the implementation**：[架构地图](architecture/internal-codebase-map.md) → [模板构建策略](effects/embed-template-architecture.md) → [写回与交付验证](workflow/delivery-contract.md)。
- **扩展字段 / Extend field support**：[fixture 流程](workflow/re-fixture.md) → 对应领域笔记 → [验证流程](workflow/verify.md)。

[能力矩阵](../capabilities.md) · [公开 SDK 示例](../../examples/sdk/README.md) · [展示工程](../../showcase/README.md)

## 按领域浏览 / Browse topics

### architecture

- [Internal codebase map](architecture/internal-codebase-map.md)
- [Public Go SDK interface](architecture/public-sdk-interface.md)
- [Unified CLI interface](architecture/unified-cli-interface.md)

### coding

- [comments checklist](coding/comments.md)

### comp

- [⚠ Recipe comp `draft_3d` is profile/roundtrip evidence, not AE DOM evidence](comp/recipe-draft-3d-profile.md)
- [Recipe comp Motion Graphics template name](comp/recipe-motion-graphics-template-name.md)

### composition

- [⚠ setAlternateSource 自动包 wrapper precomp（AE 脚本 quirk）](composition/altsource-wrapper-precomp.md)
- [⚠ Camera FilmSize 在 ldta @0x98 但 ScriptingAPI 不可写](composition/camera-filmsize-ldta-write-blocked.md)
- [⚠ cdta @0xB0 是 shutter 角度 360° 参考常量，**不是 duration**](composition/cdta-0xB0-shutter-ref-not-duration.md)
- [⚠ SetDraft3D false-green: 字节与 AE 原生逐位一致,AE DOM 仍读 draft3d=false](composition/comp-setdraft3d-false-green.md)
- [⚠ comp SetFrameRate 不 rescale duration ticks (与 SetDuration 同 comp → AE 读错时长)](composition/comp-setframerate-no-duration-rescale.md)
- [⚠ nnhd display settings: AE reads legacy nhed, not nnhd (+ feet = frames-per-foot)](composition/nnhd-display-settings-layout-re.md)
- [⚠ deriveTickRate 3× off for NTSC modern-AE files (×1000/scale was bogus)](composition/ntsc-tickrate-derive-3x-off.md)
- [Recipe Comp Background Color](composition/recipe-background-color.md)
- [Recipe Comp Comment](composition/recipe-comment.md)
- [Recipe Comp Display Start Time](composition/recipe-display-start-time.md)
- [Recipe Comp Frame Blending](composition/recipe-frame-blending.md)
- [Recipe Comp Hide Shy Layers](composition/recipe-hide-shy-layers.md)
- [Recipe Comp Label](composition/recipe-label.md)
- [Recipe Comp Motion Blur Enabled](composition/recipe-motion-blur-enabled.md)
- [Recipe Comp Motion Blur Settings](composition/recipe-motion-blur-settings.md)
- [Recipe Comp Pixel Aspect](composition/recipe-pixel-aspect.md)
- [Recipe Comp Preserve Nested Frame Rate](composition/recipe-preserve-nested-frame-rate.md)
- [Recipe Comp Preserve Nested Resolution](composition/recipe-preserve-nested-resolution.md)
- [Recipe Comp Renderer](composition/recipe-renderer.md)
- [Recipe Comp Resolution Factor](composition/recipe-resolution-factor.md)
- [Recipe Comp Work Area](composition/recipe-work-area.md)
- [⚠ shutter setter 在 AE 端触发 work-area divisor 重编码](composition/shutter-side-effect-divisors.md)
- [⚠ TickRate 是 per-composition 的，不是全局常量](composition/tickrate-per-composition.md)

### docgen

- [⚠ docgen 单包扫描 + 类型别名 → 静默空文档（且对账测试假绿）](docgen/docgen-alias-blindspot-false-green.md)
- [⚠ //nolint:jargon on exported doc comments leaks into docgen public docs](docgen/docgen-nolint-jargon-leaks-into-docs.md)

### effect

- [⚠ Recipe effect param expressions need a materializing value](effect/recipe-effect-param-expressions.md)

### effects

- [⚠ AddEffect — Effect Parade splice RE + ship findings](effects/add-effect-splice-re.md)
- [⚠ Effect param elision — value-keyed persistence + synthesis-lite path](effects/effect-param-elision-synthesis-lite.md)
- [Embedded-template architecture, v2_2 naming & version portability](effects/embed-template-architecture.md)
- [⚠ Pseudo control labels are bound to AE's system ANSI codepage (CJK not portable)](effects/pseudo-control-label-ansi-codepage.md)
- [⚠ Pseudo controls: AE renders the PANEL from the pard — the decoded field map + the render rules](effects/pseudo-control-render-field-map.md)
- [Pseudo Effect 从零生成 — 支线完成记录 + RE 参考](effects/pseudo-effect-continuation-handoff.md)
- [⚠ Pseudo layer-picker tdpi clobbered by shared-core retarget; false-green gate from 0/1-based index](effects/pseudo-layer-picker-tdpi-retarget-clobber.md)

### essential-graphics

- [⚠ Essential Graphics 写路径 RE（绑定拓扑 + 实现 gotcha）](essential-graphics/essential-graphics-write-re.md)
- [Recipe Essential Graphics controller slice](essential-graphics/recipe-essential-graphics-controller.md)

### keyframe-expression

- [⚠ 表达式激活 = tdb4 @0x77/@0x78 字节对 — 历史反语义解读翻案](keyframe-expression/expression-enable-byte-pair.md)
- [⚠ lhd3 容量字段按 4-keyframe 分页 — AE 2025 校验 reject](keyframe-expression/lhd3-keyframe-capacity-pages.md)
- [⚠ Separate Dimensions 写机制 — RE findings](keyframe-expression/separate-dimensions-write-mechanics.md)

### layer

- [⚠ DeleteLayer RE — AE 行为契约](layer/ae-deletelayer-re.md)
- [⚠ DuplicateLayer RE — AE 行为契约](layer/ae-duplicatelayer-re.md)
- [⚠ NewCameraLayer / NewLightLayer — embed-whole-Layr create](layer/camera-light-layer-create-re.md)
- [⚠ 3D-enable: flipping the Is3D bit alone makes AE materialize a full 3D layer](layer/layer-3d-enable-bit-materializes.md)
- [⚠ From-scratch layer replication silently drops static non-default transform channels](layer/layer-replication-drops-static-transform-channels.md)
- [⚠ Layer.SetComment from-scratch 层 AE 读回空 — 真因 = ldta @0x3C has-comment flag(+cmta double-NUL),非插入位置](layer/layer-setcomment-cmta-append-position.md)
- [⚠ SetStretch false-green: AE recomputes stretch from in/out span](layer/layer-setstretch-ae-recomputes-span.md)
- [⚠ SetTimeRemapEnabled false-green: AE needs 2 identity keyframes, not a static value](layer/layer-settimeremapenabled-needs-keyframes.md)
- [⚠ Layer Styles tdsb bit0 = render-time enabled bit](layer/layer-styles-tdsb-enabled-bit.md)
- [⚠ material-advanced + 3D-geometry props 脚本不可 setValue(父级隐藏)→ 双版本不可达 AE-gate](layer/material-advanced-props-hidden.md)
- [⚠ 单合成多 ShapeLayer 被 AE silent-drop（layer-level sibling chunks 缺失）](layer/multi-layer-silent-drop.md)
- [⚠ New layer types — scoping → ALL SHIPPED (2026-06-10)](layer/new-layer-types-scoping.md)
- [⚠ Precomp 层 = AV 层 + ldta @0x28 SourceID → CompItem（复用 newTemplatedLayer）](layer/precomp-layer-source-id-re.md)
- [Recipe Adjustment Layer](layer/recipe-adjustment-layer.md)
- [Recipe Layer Advanced Switches](layer/recipe-advanced-switches.md)
- [⚠ Recipe 2D Anchor Point keyframes profile as 3D `x,y,0`](layer/recipe-anchor-point-keyframe-profile-3d.md)
- [Recipe Layer Auto Orient](layer/recipe-auto-orient.md)
- [Recipe Camera Layer](layer/recipe-camera-layer.md)
- [Recipe Camera Options](layer/recipe-camera-options.md)
- [Recipe Layer Comment](layer/recipe-comment.md)
- [Recipe Layer Common Switches](layer/recipe-common-switches.md)
- [Recipe Layer Label](layer/recipe-label.md)
- [Recipe Light Layer](layer/recipe-light-layer.md)
- [Recipe Light Options](layer/recipe-light-options.md)
- [Recipe Layer Motion Blur](layer/recipe-motion-blur.md)
- [Recipe Layer Null Flag](layer/recipe-null-flag.md)
- [Recipe Null Layer](layer/recipe-null-layer.md)
- [⚠ Recipe opacity keyframes author as percent, profile as unit opacity](layer/recipe-opacity-keyframe-profile-units.md)
- [Recipe Layer Parent](layer/recipe-parent.md)
- [⚠ Recipe 2D Position keyframes profile as 3D `x,y,0`](layer/recipe-position-keyframe-profile-3d.md)
- [Recipe Layer Quality And Blending](layer/recipe-quality-blending.md)
- [⚠ Recipe scale keyframes author as percent, profile as 3D unit scale](layer/recipe-scale-keyframe-profile-units.md)
- [Recipe Layer Shy](layer/recipe-shy.md)
- [Recipe Layer Timing](layer/recipe-timing.md)
- [⚠ Recipe transform expressions require a reopen-backed property](layer/recipe-transform-expressions.md)
- [⚠ Recipe keyframe ease influence is a fraction, not AE UI percent](layer/recipe-transform-keyframe-ease.md)
- [⚠ SetLayerTransform encodes an AV/precomp layer's Anchor Point as fraction-of-source, not pixels](layer/setlayertransform-av-anchor-fraction.md)
- [⚠ From-scratch shape layer renders blank: layer Position defaults to (0,0), not comp-center](layer/shape-layer-position-default-offscreen.md)
- [⚠ V2.2 ship gate FAIL — Layr 结构 + Transform schema + tdum/tduM 缺](layer/v2-2-aelayer-structure.md)

### parse-serialize

- [⚠ 外部 AE 版本降级器 CEP 工具 RE](parse-serialize/2026-06-09-ae-version-downgrader-re.md)
- [⚠ Chunk ID 大小写敏感 — Tdb4 ≠ tdb4](parse-serialize/chunk-id-case-tdb4.md)
- [⚠ AE 24/25 ldta 第三种长度 — 找不到 (negative finding)](parse-serialize/ldta-length-third-variant-not-found.md)
- [⚠ initDerived must walk LAYER IDs when computing nextItemID](parse-serialize/nextitemid-must-include-layer-ids.md)
- [⚠ lnrb / lnrp flag chunks — 不是空 chunk，位置敏感；lnrp ScriptingAPI readback quirk](parse-serialize/project-flag-chunks-lnrb-lnrp.md)

### property

- [⚠ Project/Composition/Layer/Property/Keyframe 全部 mutate 共享 chunk bytes](property/concurrency-unsafe-shared-chunk-bytes.md)
- [⚠ PropertyBase Remove / Duplicate / MoveTo — AE 行为契约 + chunk 机制](property/property-indexed-group-structural-re.md)
- [⚠ Runtime-only fields（AE 不持久化到 .aep）](property/runtime-only-fields.md)
- [⚠ Transform Group 属性比 py-aep 少 = AE 默认值省略，不是 parser bug](property/transform-group-default-omission.md)

### render-queue

- [⚠ Render Queue delete 写机制 — RE findings](render-queue/render-queue-delete-mechanics.md)
- [⚠ RenderQueueItem.comment 无 ScriptingAPI — ship-gate 走"接受+保留"而非 readback](render-queue/rq-comment-no-scripting-api.md)

### security

- [不可信 AEP 输入安全 checklist](security/untrusted-aep-input.md)

### shape

- [⚠ AddMask — Mask Parade from-scratch atom RE + ship findings](shape/add-mask-create-re.md)
- [⚠ AE 2020 把 164B ShapeLayer ldta 判为损坏并跳过该层](shape/ae2020-shape-ldta-164-corrupt.md)
- [⚠ Gradient fill write (SetGradient) — RE + ship findings](shape/gradient-fill-write-re.md)
- [⚠ Mask 真几何在 shph bbox，`Mask.Vertices` 只 surface 归一化 ldat（读侧 gap）](shape/mask-shph-bbox-read-gap.md)
- [⚠ Shape-path keyframe write — Phase 0 RE](shape/path-keyframe-write-re.md)
- [Recipe Fill Blend Mode Profile Enum](shape/recipe-fill-blend-mode-profile-enum.md)
- [Recipe Fill Composite Order Profile Enums](shape/recipe-fill-composite-order-profile-enums.md)
- [Recipe Fill Rule Profile Enums](shape/recipe-fill-rule-profile-enums.md)
- [Recipe Gradient Fill Profile Enums](shape/recipe-gradient-fill-profile-enums.md)
- [Recipe Gradient Stroke Profile Enums](shape/recipe-gradient-stroke-profile-enums.md)
- [Recipe Merge Paths Profile Enums](shape/recipe-merge-paths-profile-enums.md)
- [⚠ Recipe Offset Paths line_join authors as string, profile reads numeric enum](shape/recipe-offset-paths-profile-enums.md)
- [Recipe Polystar Profile Fields](shape/recipe-polystar-profile-fields.md)
- [Recipe Repeater Profile Enums](shape/recipe-repeater-profile-enums.md)
- [Recipe Stroke Composite Order Profile Enums](shape/recipe-stroke-composite-order-profile-enums.md)
- [Recipe Stroke Dashes Profile Fields](shape/recipe-stroke-dashes-profile-fields.md)
- [Recipe Stroke Style Profile Enums](shape/recipe-stroke-style-profile-enums.md)
- [Recipe Stroke Taper Profile Fields](shape/recipe-stroke-taper-profile-fields.md)
- [Recipe Stroke Wave Profile Fields](shape/recipe-stroke-wave-profile-fields.md)
- [Recipe Wiggle Paths points authors as string, profile reads numeric enum](shape/recipe-wiggle-paths-profile-enums.md)
- [Recipe Wiggle Transform profile names mix Xform and shared modulation names](shape/recipe-wiggle-transform-profile-names.md)
- [⚠ Recipe ZigZag points authors as string, profile reads numeric enum](shape/recipe-zigzag-profile-enums.md)
- [ShapeLayer transform 必须写入 runtime stream](shape/shape-layer-transform-runtime-sync.md)
- [⚠ Stroke Line Cap / Line Join / Miter Limit — RE findings](shape/stroke-line-cap-join-miter-re.md)
- [⚠ Trim Paths (`ADBE Vector Filter - Trim`) — 矢量滤镜走既有 shape-body 模板 vein](shape/trim-paths-vector-filter-re.md)

### showcase

- [showcase 产出规约 — 大阶段示例 + 用户审核 — checklist](showcase/showcase.md)

### techniques

- [往 data/samples/ 加参考工程 — 文件夹约定 + 分类 — checklist](techniques/add-reference-sample.md)
- [造一个好火焰 — plugin-free 原生配方 + 参数影响对照 — checklist](techniques/build-good-fire.md)
- [Build Good Glitch — glitch phenomenon recipe](techniques/build-good-glitch.md)
- [Build Good Rain — rain phenomenon recipe](techniques/build-good-rain.md)
- [程序化视觉技法库(三轴本体 · 技法层)](techniques/fx-techniques.md)
- [⚠ 生成火焰/烟/能量类视觉：程序化效果链 > 矢量形状+模糊](techniques/procedural-fx-over-vector.md)
- [内化一个参考工程 — 5 步流水线（每喂一个 .aep 跑一遍） — checklist](techniques/understand-a-project.md)

### text

- [⚠ btdk point-measurement setter 必须 FormatPSReal 否则 AE 读成 value/65536 (假绿)](text/btdk-point-value-needs-formatpsreal.md)
- [⚠ 手动 kerning 首次启用 = 结构性添加](text/kerning-first-enable.md)
- [⚠ Text Animators (kinetic typography) — create-from-scratch RE + 落地](text/text-animator-create-re.md)
- [⚠ Text btdk length-variable write — scoping findings](text/text-btdk-length-variable-write-scoping.md)
- [⚠ Text-style expected_profile colors use normalized RGBA, not 0..255 property arrays](text/text-style-profile-color-units.md)
- [⚠ Property.ValueText（§3H）—— 通用实现不可达，need external schema DB](text/valuetext-needs-schema-db.md)
- [⚠ Variable fonts axes 写 = ScriptingAPI 不存在该字段](text/variable-fonts-write-noop.md)

### workflow

- [⚠ AE ship-gate flakes in interactive sessions: occlusion + crash-recovery cascade](workflow/ae-automation-occlusion-crashstate.md)
- [⚠ AE drops unknown RIFX chunks on resave (custom metadata stash)](workflow/ae-drops-unknown-chunks-on-resave.md)
- [⚠ AE 接受自建 .aep 的多阶段 gate + canonical seed 策略](workflow/ae25-acceptance-gate.md)
- [⚠ aeoracle render must validate PNG files after saveFrameToPng](workflow/aeoracle-saveframe-output-validation.md)
- [aeoracle slow frames need request-level frame timeout and progress-dialog ignore](workflow/aeoracle-slow-frame-script-progress.md)
- [交付准则 — 什么算「能用 / 合法交付」 — checklist](workflow/delivery-contract.md)
- [AE 效果参数字典 — dump 流程 + 使用指南 — checklist](workflow/effects-dict.md)
- [Folder usage policy for generated evidence](workflow/folder-usage-policy.md)
- [⚠ Go scratch dirs under the module are included in ./...](workflow/go-scratch-dirs-under-module.md)
- [⚠ Full-repo gofmt gate does not match the current baseline](workflow/gofmt-baseline-scope.md)
- [JSX generator cleanup provenance](workflow/jsx-generator-cleanup-provenance.md)
- [⚠ JSX -r 多轮累积 duplicate comps](workflow/jsx-state-leak.md)
- [开源许可证与商业边界](workflow/open-source-license.md)
- [Project operating rules](workflow/project-operating-rules.md)
- [JSX RE 工作流 + ship-gate — checklist](workflow/re-fixture.md)
- [验证流程 — checklist](workflow/verify.md)
