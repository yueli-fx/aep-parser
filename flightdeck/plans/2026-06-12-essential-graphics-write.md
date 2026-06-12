---
status: active
summary: Essential Graphics 写路径：SetMotionGraphicsTemplateName + AddEssentialProperty（CCtl×3 代 CIF + OvG2/CPrp 绑定 + override 值流），Phase-0 RE 纯代码已完成
last_updated: 2026-06-12
---

# Essential Graphics W

deferred-backlog「仅存的真未实现写线头」。读侧已 ship（CIF3→CCtl: Name/Type/UUID + 模板名）。

## Phase 0 — RE（✅ 2026-06-12 纯代码完成，零 AE 调用）

工具：`tmp_debug/eg_uuid_scan`（UUID 跨 chunk 定位）+ `tmp_debug/eg_dump`（CIF*/OvG2 叶子 payload dump）。
样本：`test_data/eg_{slider,color,text_source_text,multiple}_controller*.aep`（AE-native，py-aep golden 配套）。

### 绑定拓扑（三处协同写）

```
Item(comp)
  LIST:CIFO ─┐
  LIST:CIF2 ─┼─ 三代面板，内容 byte-identical（AE 兼容写）
  LIST:CIF3 ─┘   ├ LIST:CpS2 {CsCt=1, Utf8 模板名, Utf8 "en_US"}
                 ├ LIST:CapS {CsCt=1, CapL=0, Utf8 模板名}      ← 无 locale
                 ├ CPTm (8B = 00..01) · CROI (8B zero)
                 ├ CcCt (u32 BE = controller 数)
                 └ LIST:CCtl × N（面板顺序）

Layr/tdgp parade 内（所有层默认存在，含空层）:
  tdmn "ADBE Layer Overrides" + LIST:OvG2 + LIST:tdgp
    OvG2: CprC (u32 LE! 计数) + LIST:CPrp × N，每个只含 Utf8 36B UUID
    tdgp: override 值流，tdmn=<叶参数 matchName> + tdbs{tdsb,tdsn,tdb4,cdat,tdum,tduM}
    （CPrp 顺序 ↔ tdgp 子项顺序对应）
```

注意字节序混用：CprC/CsCt/CcCt 观测为 **LE**（01000000），CTyp/CCId/CLId 为 **BE**（00000002）。实现时逐 chunk 核对。

### CCtl 布局（per CTyp）

公共头：CpS2{名+locale} + CapS{名} + Utf8 36B UUID + CTyp (u32 BE)。

| CTyp | 类型 | 值 chunk |
|---|---|---|
| 2 slider | CVal/CDef f64 + Smin/Smax f64（观测 0..100） |
| 4 color | CVal/CDef 16B = 4×f32（观测 [1,0,0,1]） |
| 6 text | Utf8 值 + Utf8 默认 + CFEd/CSEd/CFEd (1B flags) + Utf8 能力 JSON（capPropFontEdit…） + Utf8 comp-link JSON（{"compId":-1,…}） + CTov (4B) |
| 1 checkbox / 5 point / 13 dropdown | 未 dump（实现时跑 eg_dump 对应 fixture） |

公共尾：CprC (=1) + LIST:CPrp{ CCId (u32 BE = comp item ID), CLId (u32 BE = 宿主层 layer ID), Utf8 = **JSON 属性路径** }。

属性路径 JSON 形如：
```json
{"0":{"index":4294967295,"matchName":"ADBE Effect Parade"},
 "1":{"index":1,"matchName":"ADBE Brightness & Contrast 2"},
 "2":{"index":1,"matchName":"ADBE Brightness & Contrast 2-0001"}}
```
index=4294967295(-1)=singleton（如 Transform Group / Opacity）；非 -1 的语义待实现时与我们 property tree 对照确证（疑为组内 dynamic index）。Transform 属性可暴露（`ADBE Transform Group/ADBE Opacity` 实测在册）。

### 已证事实

- CCId = comp item ID（fixture "primary" id=1 实测）；CLId = 宿主 Layr layerID（15/13 实测）。
- 一层多 controller：OvG2 CprC=3 + CPrp×3，全指同层。
- 无 EG 的 comp 也带三代 CIF*（CcCt=0 无 CCtl）→ AE-native comp 默认有壳；**Go-built comp（embed 模板）是否带 CIF*/OvG2 需核**（无则结构性创建）。
- 无 override 层：OvG2 = CprC(0)，tdgp 空（tdsb+tdsn+GroupEnd）。

## Phase 1 — SetMotionGraphicsTemplateName（W）

- comp 有 CIF*：length-variable Utf8 改写 ×6（3 代 × CpS2+CapS），父 LIST size 重算（同 name/comment 既有 length-variable 通道）。
- comp 无 CIF*（Go-built）：结构性创建三代空壳（CpS2/CapS/CPTm/CROI/CcCt=0）。
- Go 测试：round-trip 读回 + opaque preservation；ship-gate 搭车 Phase 2。

## Phase 2 — AddEssentialProperty（结构性，V1 范围）

V1 类型：**slider（效果标量参）+ color + checkbox**（值 chunk 简单；与 AddEffect/SetEffectParam vein 自然衔接）。text/dropdown/point deferred。

```
aep.AddEssentialProperty(comp, layer, prop, displayName) → (UUID 或 controller, error)
```

写三处（atomic invariants：warnings-as-failure + rollback）：
1. 三代 CIF* 各 append CCtl + CcCt++（无 CIF* 先建壳）。
2. 宿主层 OvG2：CprC++ + append CPrp{Utf8 UUID}。
3. 宿主层 Layer Overrides tdgp：插入 override 值流（tdmn 叶 matchName + tdbs，克隆源属性 tdbs 形态）。
+ UUID 生成（crypto/rand v4）；CVal/CDef/Smin/Smax 从源属性取。

实现位置：`internal/serializer/mutate_essential_graphics.go`；facade 自由函数。

## Phase 3 — 双版本 ship-gate

- Go-built（from-scratch）工程：NewComposition + 层 + AddEffect(slider) + AddEssentialProperty → AE 2020/2025 开 + JSX 读回 `comp.motionGraphicsTemplateControllerCount/Name`（防 ID coincidence，按 gate-fixture-id-coincidence 教训用 all-Go-built 变体）。
- fixture-mutate 变体：eg fixture 上追加 controller → 读回 N+1。
- 实验项：AE 是否容忍只写 CIF3（若容忍可简化）；先按三代全写，gate 绿后可做减法实验。

### 补充已证（2026-06-12 续）

- checkbox (CTyp=1): CVal/CDef 1B bool；point (5): CVal/CDef 16B = 2×f64；dropdown (13): CVal/CDef 4B u32 (1-based) + LIST:StVc{StVS 计数 LE + Utf8×N 选项 label}，源为 Pseudo/@@… 自定义 dropdown 效果。
- **Go-built comp 已带三代空 CIF\* 壳**（embed 模板出自 zh_CN AE，模板名「未命名」+zh_CN locale）→ Phase 1 对 Go-built comp 也是 length-variable 改写，无需结构性建壳。
- Go-built 层 OvG2 **不齐**（probe: 2 层只 1 个 OvG2——embed 模板差异）→ AddEssentialProperty 需 parade auto-create 兜底（`tmp_debug/eg_gobuilt_probe` 复测）。

## 开放问题（实现期解）

- [ ] 路径 JSON 的 index 非 -1 语义（vs 我们 property tree 的 child index）
- [ ] override tdbs 的 cdat 值 = EG 面板覆盖值 or 源值副本？（对 .mogrt 消费有意义，对 AE 接受性大概率无碍）
- [ ] CIF* 模板名 locale（en_US/zh_CN）是否需与写值一致改写
