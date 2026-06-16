# Cockpit — aep-parser

**Last updated**: 2026-06-17 by claude（两个能力 roadmap 全收口归档→ `archive/specs/`（from-scratch-mg S1–S6 闭环 · remaining-capability 6 波次全完）。本会话已 ship：3D RotateX/Orientation/RotateZ render-gate + Orientation 端序/otda 修复、音频效果波 +10、text Expressible Selector 翻案。残项迁 cockpit backlog。）

**Active focus**: **需求驱动稳态**（知识库地基重建 arc 已完结）。**知识单一家 = flightdeck + CLAUDE.md**(auto-memory 已退役);**能力真相源 = capindex**(`go run ./cmd/capindex -q <词>` / `docs/capabilities.{json,md}`,CI 强制写/做面零漏标)。库能力主线早已全收口、需求驱动;机制库 parse-the-clone + synthesis-insert + animate;每渲染类双版本 AE ship-gate(红线4)。两个能力 roadmap 已归档(`archive/specs/`),残项见下 backlog,真相源=capindex。火焰=番外(见 `incidents/procedural-fx-over-vector.md`)。

## 进行中

<!-- AUTO:inprogress -->

<!-- /AUTO -->

## 下一步

**➡ 需求驱动稳态——无 active arc。等新需求,或从下列残项挑。** 能力查询:`go run ./cmd/capindex -q <词>`。

**剩余 backlog（非阻塞,需求驱动;两个能力 roadmap 已全收口归档→ `archive/specs/`,真相源=capindex）**：
- effects：layer-ref 第二波(+4:3D Glasses/Warp Stabilizer/Timewarp/CC Particle World,按 Displacement Map 物化流程)。
- mask：**animated mask path**（reachable,与 lhd3 >4kf 容量分页耦合,om-s 时间表）· maskFeatherFalloff（位置未 RE,可能不可达）。
- expr：linear()/ease()/valueAtTime remap（内容无关已证,边际低）。

**搁置（用户决定）**：Essential Graphics 进阶 + EG 面板崩溃未修 RE。
**独立线（按需）**：Render Queue Set* slice-5~8（Alpha）· capindex 收尾小项（render-queue tag cosmetic · 11 manual-gate orphan 恢复 ae-accept,需跑 AE）。

## Backlog

- 泛型 `DuplicateItem` · `ImportComposition` · **Property synthesis**（暂搁大 feature，`incidents/transform-group-default-omission.md`）。
- fixture/RE-gated + deferred R-only（DisplayColorSpace / ValueText / environmentLayer / ligature 等）详 `specs/deferred-backlog.md` + `plans/coverage.md` § 暂搁/不可达。

## Hanging tasks

无。
