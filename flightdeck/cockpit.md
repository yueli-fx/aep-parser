# Cockpit — aep-parser

**Last updated**: 2026-06-16 by claude（**effects 扩库 arc DONE**:AddEffect 库 41→193(+152,5 波双版本 ship-gate)——wave5 +38·wave6 +23·wave7 +45(Lumetri+Lightning+整个 Cycore CC 家族)·wave8 layer-ref 家族(Displacement Map/Compound Blur render-pixel + CC Vector Blur accept,走 SetEffectLayerParam)·wave9 +43 大 probe sweep(keying/simulation/utility + CC 剩余)。机制零变更(splice + tdpi retarget)。全套 build/vet/test 绿,8 新 ship-gate 全 PASS。commit c89782f..f98cfe3。到不了 ~300 的差额有实证:模态弹框 3 个 + foreign-tdpi layer-ref 4 个 + 音频 ~18(未扫)+ AE2020 已移除/2021+ 新增不在 2020-floor。详 `incidents/add-effect-splice-re`。）

**Active focus**: **需求驱动稳态**（知识库地基重建 arc 已完结）。**知识单一家 = flightdeck + CLAUDE.md**(auto-memory 已退役);**能力真相源 = capindex**(`go run ./cmd/capindex -q <词>` / `docs/capabilities.{json,md}`,CI 强制写/做面零漏标;getter/const 豁免)。coverage.md=残值(暂搁/不可达/negative)、coverage-detail=AE-attribute 参考矩阵。—— **库能力主线**早已全收口、需求驱动(roadmap 1-6 + Text Animators 全 ship,详 `specs/2026-06-14-remaining-capability-roadmap.md`);机制库 parse-the-clone + synthesis-insert + animate;每渲染类双版本 AE ship-gate(红线4)。火焰=番外(程序化效果链,见 incident)。

## 进行中

<!-- AUTO:inprogress -->
- [2026-06-12-from-scratch-mg-roadmap.md](specs/2026-06-12-from-scratch-mg-roadmap.md) — 终极目标：不开 AE、纯 Go 从零生成完整 MG 动画工程（AI 直接产出 .aep）。按交付准则逐 slice 确权（每 slice 渲染像素级双版本 gate）：S1 ease 关键帧+规模 gate → S2 表达式激活 RE → S3 Trim Paths → S4 precomp 嵌套 → S5 Repeater/gradient 方向/圆角
- [2026-06-14-remaining-capability-roadmap.md](specs/2026-06-14-remaining-capability-roadmap.md) — 模板/真实 .aep 起点的未做能力清单，按优先级排序：动画关键帧 > 3D 图层 > 形状图层剩余 > mask > 表达式 > 文字图层。非 from-scratch（已有基础模板规避 silent-drop）。基本图形搁置。
- [2026-06-16-capability-index.md](specs/2026-06-16-capability-index.md) — 源码内 aep:cap 结构化 tag → cmd/capindex 自动生成可秒查的能力+API 索引(JSON+人读表) + version(minver) 支持,CI 防漂移;取代手写易过期的 coverage.md
<!-- /AUTO -->

## 下一步

**➡ 知识库地基重建 arc 已完结(capindex P1+P2 + knowledge-consolidation E/D/C/A/B 全 done)。回归需求驱动库工作。**

**可选收尾小项(非阻塞)**:
- capindex:render-queue tag directive-first→END(cosmetic,功能等价、go/doc 已 strip)· orphan(11 manual-gate op 如 DeleteLayer 编码 JSX gate 为 Go test → 恢复 ae-accept,需跑 AE)。
- landing housekeeping:capability-index spec(graduate)可 graduate-to-docs;knowledge-consolidation(done,一次性整合)可归档。
**库能力 backlog**(需求驱动,见下)· **能力查询**:`go run ./cmd/capindex -q <词>`。

**原库能力 backlog（非阻塞,需求驱动;火焰=番外搁置）**：
- 文字：动画器 5 类全 ✓ + **animate leaf 全收口** ✓ + **免费近邻收口** ✓（2026-06-16：~~Fill Opacity~~ ~~Stroke Opacity~~ ~~Stroke Width~~ ~~Stroke Color~~ ~~Skew~~ 5 个双版本 render-gate PASS；**Rotation X/Y evidence-based defer**——2D 层视觉惰性 bbox 三帧全同，需逐字 3D，facade 保留标 Alpha/write-only + round-trip 自验）；**structural op（Remove/Dup/Move）双版本 gate PASS** ✓ + **Range Advanced（SetTextRangeAdvanced，Amount render-gate PASS）** ✓ + **多 Selector（AddTextRangeSelector）** ✓ + **Wiggly Selector（AddTextWigglySelector，双版本 render-gate PASS via 时间变化签名）** ✓（2026-06-16）；**selector 家族收口**。**Expressible Selector = evidence-defer**（Amount 表达式驱动，库表达式未渲染验证）。详 `incidents/text-animator-create-re.md`
- shape：**矢量滤镜家族彻底收口**（2026-06-16）——所有 elided 可写子流全 ship，synthesis-insert 累计 15×。ZigZag Points · Twist Center · Offset Line Join/Miter/Copy Offset · Repeater Order（8×，2026-06-15）+ **Wiggle Paths/Transform 调制全收**（9–15×，2026-06-16）：Roughen Points + Correlation 双版本 render-pixel gate（`TestMGWiggleMod`,Points 用弦偏离 d=10 抓直vs曲、Correlation 用 spread 塌缩）;Temporal/Spatial Phase(两 filter)+ Wiggler Correlation 双版本 roundtrip gate（`TestMGWiggleModRT`,噪声相位无 categorical 像素,验 AE接受+值读回+resave）。Gradient stroke Dashes/Taper/Wave + Blend Mode/Composite Order 早已 ship。**shape 无剩余可写子流**。详 `trim-paths-vector-filter-re.md`
- mask：maskFeatherFalloff（位置未 RE，可能不可达）
- effects：**扩库 arc DONE**（库 193,wave5-9 双版本 ship-gate）。剩余可加（需求驱动）：**音频效果波**（+~18,需音频层探针）· **layer-ref 第二波**（+4：3D Glasses/Warp Stabilizer/Timewarp/CC Particle World,按 Displacement Map 物化流程）· per-effect typed helper（低价值）。模态弹框 3 个（Apply Color LUT/PS Arbitrary Map/Numbers）headless 不可达;AE2020-floor 限制部分 2021+ 效果不可达
- expr：linear()/ease()/valueAtTime remap（内容无关已证，边际低）
- 3D：RotateX/Orientation/RotateZ 补 gate（同路径，低优先）

**搁置（用户决定）**：Essential Graphics 进阶 + EG 面板崩溃未修 RE。
**独立线（按需）**：Render Queue Set* slice-5~8（Alpha）。

## Backlog

- **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持。
- 泛型 `DuplicateItem` · `ImportComposition` · **Property synthesis**（暂搁大 feature，`incidents/transform-group-default-omission.md`）。
- fixture/RE-gated + deferred R-only（DisplayColorSpace / ValueText / environmentLayer / ligature 等）详 `specs/deferred-backlog.md` + `plans/coverage.md` § 暂搁/不可达。

## Hanging tasks

无。
