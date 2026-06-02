# incidents/ — INDEX

<!-- AUTO:incidents -->
- [ae2020-shape-ldta-164-corrupt.md](ae2020-shape-ldta-164-corrupt.md) — active — AE 2020 把 164B ShapeLayer ldta 判为损坏并跳过该层
- [ae25-acceptance-gate.md](ae25-acceptance-gate.md) — active — AE 接受自建 .aep 的多阶段 gate + canonical seed 策略
- [ae-automation-occlusion-crashstate.md](ae-automation-occlusion-crashstate.md) — active — AE ship-gate 交互式会话 flake：遮挡 + 崩溃恢复级联
- [ae-deletelayer-re.md](ae-deletelayer-re.md) — active — DeleteLayer RE — AE 行为契约
- [ae-duplicatelayer-re.md](ae-duplicatelayer-re.md) — active — DuplicateLayer RE — AE 行为契约
- [altsource-wrapper-precomp.md](altsource-wrapper-precomp.md) — active — setAlternateSource 自动包 wrapper precomp（AE 脚本 quirk）
- [camera-filmsize-ldta-write-blocked.md](camera-filmsize-ldta-write-blocked.md) — active — Camera FilmSize 在 ldta @0x98 但 ScriptingAPI 不可写
- [cdta-duration-two-representations.md](cdta-duration-two-representations.md) — active — cdta 两套 duration 表示（frame-count @0xB0 vs time-ratio dividend/divisor）synthetic fixture 会分歧
- [chunk-id-case-tdb4.md](chunk-id-case-tdb4.md) — active — Chunk ID 大小写敏感 — Tdb4 ≠ tdb4
- [concurrency-unsafe-shared-chunk-bytes.md](concurrency-unsafe-shared-chunk-bytes.md) — active — 所有 mutate 共享 chunk bytes，调用方自己锁
- [gradient-fill-write-re.md](gradient-fill-write-re.md) — active — Gradient fill write (SetGradient) — RE + ship findings
- [jsx-state-leak.md](jsx-state-leak.md) — active — JSX -r 多轮累积 duplicate comps
- [kerning-first-enable.md](kerning-first-enable.md) — active — 手动 kerning 首次启用 = 结构性添加
- [keyframe-byte-layout-dispatcher.md](keyframe-byte-layout-dispatcher.md) — active — 关键帧字节布局两种 — 必须走 layoutFor 分发
- [ldta-length-third-variant-not-found.md](ldta-length-third-variant-not-found.md) — active — AE 24/25 ldta 第三种长度 — 找不到 (negative finding)
- [multi-layer-silent-drop.md](multi-layer-silent-drop.md) — active — 单合成多 ShapeLayer 被 AE silent-drop（layer-level sibling chunks 缺失）
- [nextitemid-must-include-layer-ids.md](nextitemid-must-include-layer-ids.md) — active — initDerived 计算 nextItemID 必须 walk LAYER IDs
- [path-keyframe-write-re.md](path-keyframe-write-re.md) — active — Shape-path keyframe write — Phase 0 RE
- [project-flag-chunks-lnrb-lnrp.md](project-flag-chunks-lnrb-lnrp.md) — active — lnrb / lnrp flag chunks — 位置敏感；lnrp ScriptingAPI readback quirk
- [property-indexed-group-structural-re.md](property-indexed-group-structural-re.md) — active — PropertyBase Remove/Duplicate/MoveTo — INDEXED_GROUP 谓词 + 纯 tdmn+payload pair splice（无 count chunk）
- [pwsh-7-no-winrt.md](pwsh-7-no-winrt.md) — active — pwsh 7 无 WinRT projection — 用 powershell.exe 5.1 子壳
- [rq-comment-no-scripting-api.md](rq-comment-no-scripting-api.md) — active — RQItem.comment 无 ScriptingAPI — binary-only 字段 ship-gate 走接受+保留
- [render-queue-delete-mechanics.md](render-queue-delete-mechanics.md) — active — RQ delete = LItm + settings ldat + lhd3 count + Rout 四处联动；settingsBlock 别名需重挂
- [runtime-only-fields.md](runtime-only-fields.md) — active — Runtime-only fields（AE 不持久化到 .aep）
- [separate-dimensions-write-mechanics.md](separate-dimensions-write-mechanics.md) — active — Separate Dimensions 写 = 值迁移 + Position_2 合成 + 组件数分支（非翻 bit）
- [shutter-side-effect-divisors.md](shutter-side-effect-divisors.md) — active — shutter setter 在 AE 端触发 work-area divisor 重编码
- [stroke-line-cap-join-miter-re.md](stroke-line-cap-join-miter-re.md) — active — Stroke Line Cap / Line Join / Miter Limit — RE findings
- [tickrate-per-composition.md](tickrate-per-composition.md) — active — TickRate 是 per-composition 的，不是全局常量
- [transform-group-default-omission.md](transform-group-default-omission.md) — active — Transform Group 比 py-aep 短 = AE 省略默认属性（非 parser bug）；py-aep 合成 schema 我们不合成
- [v2-2-aelayer-structure.md](v2-2-aelayer-structure.md) — active — V2.2 ship gate FAIL — Layr 结构 + Transform schema + tdum/tduM 缺
- [variable-fonts-write-noop.md](variable-fonts-write-noop.md) — active — Variable fonts axes 写 = ScriptingAPI 不存在该字段
- [windows-media-ocr-cjk-glyph-spacing.md](windows-media-ocr-cjk-glyph-spacing.md) — active — Windows.Media.Ocr 用空格分隔 CJK 字形
<!-- /AUTO -->
