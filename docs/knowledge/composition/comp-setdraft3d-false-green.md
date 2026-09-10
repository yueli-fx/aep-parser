# ⚠ SetDraft3D false-green: 字节与 AE 原生逐位一致,AE DOM 仍读 draft3d=false

SetDraft3D false-green: 字节与 AE 原生逐位一致,AE DOM 仍读 draft3d=false

## Signature
- symptom: `Composition.SetDraft3D(true)` 写 cdta @0x8A bit0 = 1,Go round-trip 绿,但 AE reopen 读 `comp.draft3d=false`
- error_type: —  (false-green / DOM 不反映 stored 字节)
- where: `Composition.SetDraft3D` → `setCdtaFlagBit(flagDraft3D)` (back_composition.go);probe = TestCompIdta_AEShipGate_*(DRAFT comp,已从 gate 移除)
- trigger: 补验 arc 批21(2026-06-17)comp_idta gate 想把 SetDraft3D 一起升

## 症状/复现

comp_idta gate 三项(comment/label/draft3d)一起验。AE2025 读回:CMT.comment=OK、LBL.label=OK、**DRAFT.draft3d=false → NO**。

**比普通 false-green 更刁**:RE 用 AE2020 原生 fixture `re_comp_idta.aep`(`comp.draft3d=true` 后存盘),diff DRAFT vs BARE,**唯一非 item-ID 差异就是 cdta @0x8A: 00→01**。即我们写的字节与 **AE 自己写的逐位一致**,AE 照样 reopen 读 false。

## 根因(推断)

draft3d 是 **derived/runtime 态**,AE 的 `comp.draft3d` DOM 属性不从 cdta @0x8A bit0 单独反映——同 lnrp(`app.project.linearizeWorkingSpace` 即便 AE 自存也读 false,见 [project-flag-chunks-lnrb-lnrp](../parse-serialize/project-flag-chunks-lnrb-lnrp.md))、同 SetTimeRemapEnabled(需 2 identity 关键帧而非裸字节,见 [layer-settimeremapenabled-needs-keyframes](../layer/layer-settimeremapenabled-needs-keyframes.md))。可能联动其它状态(3D 层存在性 / 渲染器 / session 态),单写 @0x8A bit0 不足以让 DOM 认。批6 早观察到"@0x8A draft3d DOM 不反映(其余 5 个 @0x8B flag 全反映)",本批坐实。

> 注:**未单独验 AE 原生 DRAFT comp 自身 reopen 是否读 true**(那需再开一次 AE,低 ROI)。但"我们字节 = AE 字节、AE 读我们的=false"已足够判定:单靠 @0x8A bit0 不可 DOM 值验。

## 当前处理

- SetDraft3D 留 **verify=roundtrip**,boundary 记此 DOM 假绿 + acceptance 封顶。tier 仍 stable(字节写入正确、byte-preservation 有效;只是 DOM 不可值验,非写错)。
- 不纳入 comp_idta gate(SetComment + SetLabel 已双版本 ae-accept)。
- 升级路径(若需):RE AE 原生 DRAFT 自身 reopen 行为 + 找 draft3d 真正的 DOM-绑定状态(可能需 3D 层载体或额外 flag),非单字节写。

## Cases
- 2026-06-17 补验 arc 批21 首次坐实。同 gate 的 SetComment(idta @0x39 + cmta 位置修复)/ SetLabel 正常双版本升 ae-accept;唯 draft3d false-green。false-green 家族:[layer-setstretch-ae-recomputes-span](../layer/layer-setstretch-ae-recomputes-span.md) · [layer-settimeremapenabled-needs-keyframes](../layer/layer-settimeremapenabled-needs-keyframes.md) · [project-flag-chunks-lnrb-lnrp](../parse-serialize/project-flag-chunks-lnrb-lnrp.md)(lnrp readback quirk)。
