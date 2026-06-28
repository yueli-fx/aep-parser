# Embedded-template architecture, v2_2 naming & version portability

SUMMARY: Embedded-template architecture, v2_2 naming & version portability
READ WHEN: 想理解库怎么造形状/效果(为什么必须基于 .bin);看到 v2_2/v22 命名困惑;纠结 effect/shape 跨 AE 版本要不要分版本号;评估从零生成 vs embed-body
RECHECK WHEN: embed-template 模板布局重组(2026-06-17 已分子目录),或写策略从 embed-body 转向从零生成,或 per-version 模板机制落地

---

> 三个会话调查的合并总结(2026-06-17,用户提问驱动)。本条保留为架构知识：模板命名、版本分裂触发条件、以及为什么从真实 AE 输出克隆模板比逐字节手搓安全。

## 1. 构建形状/效果 = 基于 `.bin`,不是从零生成字节

**核心**:库的写策略是 **「parse-the-clone + 覆写值」**,不是「从字节零生成」。

| 类别 | 是否基于 .bin | 说明 |
|---|---|---|
| **效果(effects)** | **100% 必须** | 每个效果 = 一个 `effect_*.bin`(从 AE 存过该效果的真 .aep 抽出的 `(tdmn, sspc)` 对)。`AddEffect` 克隆 + 改 host id + splice。**没法从零造效果**——参数树太大太多 opaque/版本相关字段。只有 RE 进库的效果可加。 |
| **形状(shapes)** | **基本都是**(连 Rect/Ellipse/Fill/Stroke) | `lowerRectNode` 注释道破:*"From-scratch construction triggered silent drop; embedding the canonical body + overwriting Size cdat … is the validator-safe path."* 纯从零拼字节造 Rect,**AE 会 silent-drop**。矢量滤镜(Trim/Repeater/ZigZag/Star…)同理。这套叫 **embed-body vein**。 |
| **脚手架 + 数值** | **否,纯生成** | comp 头(cdta)、layer 列表、Layr 头、Transform 通道值、关键帧流、工程头、mask atom → `back_*.go`/`lower_layer.go` 直接造 `rifx.Chunk{}`。形状的**值**(Size/颜色/关键帧)也是算出来覆写进 .bin 槽位。 |

**为什么这么设计**:AE 是**严格校验器**——chunk 格式有大量未公开/版本相关/默认即省略(elided)字段,任何字节结构跟它 canonical 输出不一致的,要么判损坏、要么静默丢。既有教训包括 Layr 结构不完整导致 shape layer silent-drop、以及多层工程中某些层被 AE 静默丢弃。逐字节逆出来从零生成既不可行也脆。「拿真实例克隆 + 只改看懂的值槽」是双版本 ship-gate 反复验证的安全路径。

**代价**:只能定制模板里**存在且映射过**的值槽。模板里被 elide 的字段「没槽可写」——要支持就 RE 一份更丰富 fixture 重抽 .bin(`synthesis-insert` 那套:splice 一个 AE-native leaf 再复位默认)。

**`.bin` 怎么来(可复现,tracked 进 git)**:`tmp_debug/extract_{effect_lib,shape_bodies,text_animator}` 等工具从 `test_data/re_*.jsx`(JSX 在 AE 里 author 出结构)存的 .aep 抽出。加新效果/选择器都要先跑 AE fixture 再 extract。

## 2. `v2_2` / `v22` = 项目里程碑代号,不是 AE 2022

**用户被它误导过,是真包袱。**(2026-06-17 重构已清:文件名剥 `v2_2_shape_`、Go 标识符 `v22Shape*`→`shape*`。)

- AE 版本在本仓写成 4 位年份:`TargetAE2020/2022/2025`、seed `2020.aep…2025.aep`。**AE 2022 是 `2022.aep`/`TargetAE2022`,跟 `v2_2` 拼写都不同。**
- `v2_2` 是开发阶段:**V2.1 地基 → V2.2 从零创建图层/形状层 → V3 结构性 mutation**。这是项目历史代号,不是 AE 版本号。
- 旧标题里出现过 *"V2.2 ship gate FAIL — Layr 结构…"*；保留该说法只为解释命名来源。
- 所以 `v22ShapeRectBody` / 旧 `v2_2_shape_*.bin` = 「V2.2 阶段 RE 出的形状体模板」,跟 AE 版本无关。
- **保留的 prose**:讲 V2.2/V2.3 里程碑历史的注释是准确背景,不算包袱,未清。

## 3. 跨版本:effect/shape body 单一 + 实测可移植,版本敏感性在 seed 层

**用户顾虑(对的)**:AE 升级可能改某效果/形状初始字段,使 AE-2020 body 在新版本被丢/读错。

**项目现状(查证)**:
- **版本敏感的东西在 seed 层按版本分**:`mutate_project_new.go` `case TargetAE2025` → `templates/project/2025.aep`。工程头/版本指纹全在 seed。
- **effect/shape body 单一(全 AE-2020 抽)**,靠**双版本 ship-gate(红线6)持续实测**可移植:`TestAddEffect_AEShipGate_AE2020`+`_AE2025` 把整库灌两个真 AE,断言都接受+读回正确。**不假设,每个模板持续证明。**
- **真出现的版本差异靠版本条件 CODE**(不是模板变体):ldta 按 target 补 160B(2020/22)/164B(2025);lhd3 关键帧容量 AE2025 更严(>4kf,被 2025 gate 抓出后修)。
- **没有任何 per-version 的 effect/shape 模板**;`effectTemplateFiles` 是 `map[string]string`。

**为什么 body 能单一**:内置效果/形状序列化 schema 跨版本稳定(AE 向后兼容,2025 能开 2020 工程)。版本间真正变的是**新增字段**(AE24 fontCapsOption / AE23 trackMatte),那些在图层/文档层、单独处理,不在这些 body 里。

**要不要加版本号?——现在不加(YAGNI),触发条件明确**:某模板的某版本 gate 真红了才加,且分级响应(局部字段差异→代码条件分支;整份模板差异→按 AE 版本键化模板 map + 覆盖文件)。**双版本 gate 就是版本分裂探测器。**

**诚实边界**:① 可移植性是**实测**非 Adobe 承诺,大改某 schema 会断、但 gate 先报警;② 只 gate **bookend(2020+2025)**,中间 2021/2023/2024 假设向后兼容未逐个 render-gate。
