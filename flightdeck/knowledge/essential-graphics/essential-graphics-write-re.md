# ⚠ Essential Graphics 写路径 RE（绑定拓扑 + 实现 gotcha）

SUMMARY: Essential Graphics 写路径 RE（绑定拓扑 + 实现 gotcha）
READ WHEN: extending AddEssentialProperty (point/dropdown/text/Transform-sourced controllers, RemoveEssentialProperty); debugging AE rejecting/ignoring a Go-written EG controller; touching CIF*/CCtl/OvG2/CPrp chunks; needing the EG path-JSON index semantics or the CIF-generation / endianness quirks

---

来源：2026-06-12 纯代码 RE（`tmp_debug/eg_uuid_scan` / `eg_dump` / `eg_gobuilt_probe` 扫 `test_data/eg_*.aep` AE-native fixtures），零 AE 调用解开全部格式；双版本 ship-gate `TestEGAdd_AEShipGate_AE2020/_AE2025` PASS 验证。实现：`internal/serializer/mutate_essential_graphics.go` + `write_essential_graphics.go`。

## 绑定拓扑（三处协同，UUID 是关联键）

```
Item(comp)
  LIST:CIFO / CIF2 / CIF3      ← 三代面板，内容必须 byte-identical（AE 兼容写；
                                  漏写某代 → 对应 AE 版本显示旧状态）
    LIST:CpS2 {CsCt, Utf8 模板名, Utf8 locale}
    LIST:CapS {CsCt, CapL=0, Utf8 模板名}      ← 模板名共 2 槽/代 × 3 代 = 6 槽
    CPTm(8B=…01) · CROI(8B 0) · CcCt(u32 BE 控件数)
    LIST:CCtl × N（面板顺序）

Layr 属性 tdgp 内（特殊三元组，非标准 tdmn+tdgp 对）:
  tdmn "ADBE Layer Overrides" + LIST:OvG2 + LIST:tdgp
    OvG2: CprC(u32 LE!) + LIST:CPrp×N{Utf8 36B UUID}
    tdgp: override 值流 ×N（tdmn=叶参数 matchName + LIST:tdbs），与 CPrp 同序
```

CCtl 布局：CpS2{名+locale} + CapS{名} + Utf8 UUID + CTyp(u32 BE) + 按类型值 chunk + CprC(BE 1) + LIST:CPrp{CCId(BE comp item ID), CLId(BE 宿主层 ID), Utf8 路径 JSON}。

按 CTyp 值 chunk：slider(2)=CVal/CDef f64+Smin/Smax f64 · checkbox(1)=CVal/CDef 1B · color(4)=CVal/CDef 16B **4×f32 RGBA 0..1**（注意 ≠ cdat 的 4×f64 ARGB 0..255）· point(5)=2×f64 **像素**（≠ cdat 分数——deferred 的换算坑）· dropdown(13)=u32 1-based + LIST:StVc{StVS 计数 LE + Utf8×N label} · text(6)=双 Utf8 值/默认 + CFEd/CSEd 1B flags + 能力 JSON + comp-link JSON + CTov。

## 实现 gotcha（违反即 AE 拒收/误读）

1. **路径 JSON index = 父组内 0-based 位置**；固定组（Effect Parade 自身 / Transform Group / Opacity）= -1，序列化为 `4294967295`。叶参数 ordinal 含 "-0000" 头对（计 0）。⚠ 易误判：单效果 fixture 的 effect index=1 不是 1-based——是 parade 里前面还有别的效果（grep 过滤会漏兄弟，先 dump 全 parade 再下结论）。
2. **字节序混用**：OvG2 的 CprC + CpS2/CapS 的 CsCt + StVc 的 StVS = **LE**；CcCt、CCtl 内 CprC、CTyp/CCId/CLId、CVal/CDef/Smin/Smax = **BE**。同 ID（CprC）两处不同 endianness，逐位置照抄观测值。
3. **override tdgp 值流所有源类型都有**（含 Transform 源如 ADBE Opacity——grep 时它不带效果名前缀容易漏）。流 = 源参数 tdbs 克隆（已物化）或 synthesis-lite 模板+写当前值（elided，复用 `cloneEffectParamTemplate`/`cloneGenericParamTemplate`）。流内 tdsn **仅在 controller 改名时**携带 display name，否则 "-_0_/-" 占位。
4. **Go-built 工程**：comp 经 embed 模板自带三代空 CIF* 壳（zh_CN locale「未命名」——locale 从所在代的 CpS2 读、别硬编码 en_US）；层的 Layer Overrides 三元组**不齐**（embed 模板差异）→ 需 AE-native 空壳 auto-create（CprC=0 OvG2 + 空 tdgp，锚在 "ADBE Layer Sets"/"ADBE Source Options Group" 前）。
5. CCId=comp item ID / CLId=宿主层 layer ID（`dump_item_ids` 实测钉死）；UUID = 标准 v4 小写 36 字符，AE resave 后保持不变（gate 断言过——可作身份键）。
6. color 的 pard 不存默认值（`parseSinglePard` PCTLColor 空实现）→ color controller 需参数已物化（先 SetEffectParam）。

## Signature

symptom: AE 打开含 Go 写 EG controller 的工程后面板为空/控件丢失/名字显示旧值；或 motionGraphicsTemplateControllerCount 读回 0
error_type: —
where: internal/serializer/mutate_essential_graphics.go · write_essential_graphics.go · comp Item CIF*/CCtl · Layr OvG2/Layer Overrides tdgp
trigger: 三代 CIF* 没写齐 / CprC-CsCt 端序写反 / 路径 JSON index 语义按 1-based 写 / override 值流缺失或 tdsn 误带未改名 display name
