---
status: active
summary: BuildPseudoEffect 支线已全部完成(Dropdown/Group/Label/Point 坐标/3DPoint 坐标/Layer-picker 均双版本 AE gate 绿)。仅剩 CJK 控件标签的 AE 架构限制(byte-equiv-only,不可 ship-gate)。下方保留 RE 字节布局与 ship-gate 技巧作参考。
when_to_read: 维护 BuildPseudoEffect、新增伪控件类型、或复用 pard/value-entry 字节布局 + ship-gate 技巧时
applies_to: [pseudo-effect, build-pseudo-effect, handoff, continuation, point, layer-picker, dropdown, group, value-entry]
last_updated: 2026-06-20
---

# Pseudo Effect 从零生成 — 支线完成记录 + RE 参考

> 冷启动顺序:先读本篇 → `specs/2026-06-20-pseudo-effect-support.md`(§9 = 从零生成全 RE)→ `incidents/pseudo-control-label-ansi-codepage.md`(CJK 限制)→ `incidents/add-effect-splice-re.md`(splice 机制)。能力真相源:`go run ./cmd/capindex -q "pseudo"`。

## 0. 支线收口(2026-06-20,commit 4fa4a5c + d661bb9)

**全部剩余控件类型已实现 + 双版本 AE ship-gate 绿。** Pseudo Effect Maker 能造的控件类型现在都能纯 Go 离线合成。

- **Dropdown / Group / Label**(commit `4fa4a5c`):pard-only(值条目可省略——AE 从 pard 读默认)。gate `TestBuildPseudoEffectRich_AEShipGate_AE2020/2025`。**决定性发现**:伪效果「组」是**扁平标记控件**(Effect Controls 视觉分组),非属性树嵌套——AE 原生金样本读回同样扁平(组+子控件都是顶层兄弟),仅内置 Compositing Options 真嵌套。
- **Point/3DPoint 自定义坐标 + Layer-picker**(commit `d661bb9`):值条目合成(tdmn + LIST tdbs{tdsb,tdsn,tdb4(124),cdat,[tdpi,tdps]})。gate `TestBuildPseudoEffectValueEntry_AEShipGate_AE2020/2025`。
  - **Point cdat = 坐标空间分数**(RE 决定性:读金样本 point 回 AE = value [0.000025,2500] @ 500px host → cdat [5e-8,5.0],比值 500=维度)。`PointDefault [fx,fy(,fz)]`,实测 [0.25,0.125]→[100,50]、3D 加 z 0.0625→25。
  - **Layer-picker**:pard type 0x00(同 header)+@0x30=2;绑定在值条目 `tdpi`=目标层内部 ID(`aep.Layer.ID`),0→AE 解析为第一层。gate 绑第二层、AE 读回该层索引。
  - per-type tdb4(124B)逐字节抄金样本(@0x10 块=per-dim 常量,非值相关);新增 rifx `IDTdps`。

**唯一未解 = CJK 控件标签**(AE 架构限,详 `incidents/pseudo-control-label-ansi-codepage.md`):标签=pard @0x10 名,按查看机系统 ANSI 码页解码,GBK-pard-name 仅 byte-equivalence 验(本西欧码页机不可 ship-gate)。非缺口、是 AE 限制。

---

## 1. 历史现状(Phase 1 之前,保留作背景)

`BuildPseudoEffect(layer, uid, name, displayName, controls)` —— 纯 Go 从零合成伪效果(无 .ffx / 无 AE / 不 clone 模板),splice 进 Effect Parade。**AE 2020+2025 双版本 gate 绿**。

已支持控件类型 + 自定义:
- **Slider**(0x0a):`Min`/`Max`(=AE minValue/maxValue 的 valid range)+ `Default` —— 实测回读。
- **Color**(0x05):`Color []float64` RGBA 默认。
- **Checkbox**(0x04):`Checked`;parT 尾随 `pdnm`(Utf8 on/off 标签)。
- **Angle**(0x03):`Default`(度)。
- **Point**(0x06)/**Point3D**(0x12):仅结构默认(坐标=原点,**自定义坐标未做**)。
- **控件标签**:pard @0x10 名 **GBK 编码**(`pardNameBytes`)→ 中文 Windows 显示对(byte-equiv AE 原生,非 gate)。
- **效果显示名**`displayName`:值组顶层 tdsn(Utf8)→ CJK gate 绿。

相关 commit:`cb4ab44`(从零合成)`c862542`(自定义值)`22f1449`(CJK 负结论)`6a4a369`(GBK 标签)。

## 2. 剩余要做(本支线的"做完")

按 ROI 排序。**前两项共用「value-entry 合成」底层**(机制已验 AE 接受非 elide 值条目)。

### A. Point / 3DPoint 默认坐标 + Layer-picker(共用 value-entry 合成)
- **坐标不在 pard,在值条目**:`tdmn(controlMN) + LIST tdbs{tdsb + tdsn(Utf8 名) + tdb4(124B,每类型 RE'd) + cdat(坐标值)}`,插在值组 effect-label tdsn 之后、built-in 组之前。
- **value-entry tdb4 已 RE 出**(见 §4),注意 **不等于** 库里 `makeTdb4`(尾常量 `5da8` vs 库的 `7800`,且 @0x08-0x0B 带 effect-param 元数据)。
- cdat 尺寸:dim1=40 · dim2(point)=48 · dim3(3dpoint)=72 · dim4(color)=96。坐标 = N×f64 BE(像素值,可能有 scale,需 RE 确认:point cdat[0:8] 看金样本 0007 = `3e6ad7f2…` 是个小数,要对 AE 实际坐标反推 scale)。
- **Layer-picker**:control_type **0x00**(同 header)+ pard @0x30=2;值条目额外带 `tdpi`(4B,层 id ref)+ `tdps`(4B)。tdpi 绑定见 `add-effect-splice-re.md`。
- ⚠️ value-entry 合成我上轮 **revert 了**(对 CJK 无用),但**机制已证 AE 接受**——重做时从 git 历史 `6a4a369^`(被 revert 前)捞 `synthControlValueEntry` + `pseudoValueEntryTdb4` 字节常量,或重新 RE。

### B. Dropdown(0x07)
- pard:@0x38 last · @0x3C hi16=nb_options/lo16 · parT 尾随 `pdnm`=Utf8「opt1|opt2|opt3」(分隔符 `|`)。金样本 dropdown pdnm hex=`...e98089e9a1b9 31 7c e98089e9a1b9 32`(「选项1|选项2」)。
- `PseudoControl` 加 `Options []string`,join `|` 写 pdnm,nb_options 写 pard。

### C. Group / Label(0x0d / 0x0e)嵌套
- **Group start**(0x0d,@0x04=0x00)… 子控件 … **GroupEnd**(0x0e,@0x04=0x08)配对。
- **Label**(0x0d,@0x04=**0x20**)= 纯文字标签(无值)。
- 需在 parT + 值组都体现嵌套层级。API 设计:`PseudoControl` 加 `Kind: PseudoGroup/PseudoGroupEnd/PseudoLabel`,或嵌套 `Children []PseudoControl`(更友好)。

## 3. 所需材料 / 资源 / 文件

| 类 | 路径 | 用途 |
|---|---|---|
| **RE 真相源** | `test_data/pseudo_rich_demo.aep` | 用户 AE-2022 出的 13 控件金样本,**所有字节布局的神谕**(force-add tracked) |
| 代码主体 | `internal/serializer/mutate_pseudo_effect_build.go` | `BuildPseudoEffect`/`synthPseudoSspc`/`synthControlPard`/`synthPard`/`pardNameBytes` |
| facade | `internal/aep/facade.go` | `BuildPseudoEffect` + `PseudoControl`/`PseudoControlKind` 别名 + cap tag |
| ship-gate test | `internal/aep/pseudo_effect_shipgate_test.go` | `TestBuildPseudoEffect_AEShipGate_AE2020/2025` 模板,照抄加新 test |
| verify JSX | `test_data/verify_pseudo_effect.jsx` | 通用回读校验,支持 `checks:[{idx,prop,expect}]`(value/min/max)+ nameCodes |
| 单测 | `internal/serializer/mutate_pseudo_effect_build_test.go` | GBK byte-equiv(`TestPardNameBytes_GBKMatchesAENative`) |
| **probe 工具**(gitignored,可复用) | `tmp_debug/pseudo-spike/pardfull/` | dump pard 字节(`go run ./tmp_debug/pseudo-spike/pardfull <aep> [matchname过滤]`) |
| | `tmp_debug/pseudo-spike/valdump/` | dump 值组 tdgp 树 |
| | `tmp_debug/pseudo-spike/valtdb4/` `fulltdb4/` | dump value-entry tdb4/cdat |
| | `tmp_debug/pseudo-spike/synthdump/` `cjkdemo/` | 用新代码 build demo .aep |
| | `tmp_debug/pseudo-spike/probe_slider.jsx` | AE 里 dump 控件 value/min/max(开金样本对照) |
| 参考实现 | `flightdeck/references/PseudoEffect/`(gitignored) | rendertom 源 + Scribe.ffx 样本 |
| 外部参考 | aescripts Pseudo Effect Maker · atarabi.github.io/at_script/API/Pseudo | 控件能力对标 |

## 4. RE 出的字节布局(直接抄)

**pard 148 字节**(`synthPard` 已实现 control_type@0x0F + name@0x10 GBK 32B):
- Slider 0x0a:@0x04=0x200 · @0x38 f8 default · @0x68/@0x6C f4 valid(=AE min/max)· @0x70/@0x74 f4 visible · @0x78 f4 default · @0x7C=0x00050003
- Color 0x05:@0x38 ARGB last · @0x3C ARGB default(**非 @0x40**)
- Angle 0x03:@0x38/@0x3C s4 度数 16.16
- Checkbox 0x04:@0x38 u32 last · @0x3C u8 default(1=勾)+ 尾 pdnm
- Point 0x06:@0x3C=0x00050000 · @0x48=0x00640000 结构常量(缺则 AE「range has no values」)
- Dropdown 0x07:@0x38 last · @0x3C(hi16 nb_options)+ 尾 pdnm
- Layer:type 0x00 + @0x30=2
- Group/Label 0x0d(label @0x04=0x20)· GroupEnd 0x0e(@0x04=0x08)

**value-entry tdb4 head**(124B,从金样本 verbatim,@0x10-0x3B 是 opaque f64 scale 块 + @0x3C control-type marker):
- dim1(slider/angle/checkbox):`db990001 0001 0000 0001 0004 0000 5da8` + 0x10起 `3f1a36e2eb1c432d 3ff0…(几个 1.0) … 0404 0000`(@0x3C=control_type)
- dim4(color):`db990004 0007 0001 0002 ff04 0000 5da8` + @0x30尾 `0101 0000`
- dim2(point):`db990002 000f 0003 ffff ff04 0000 5da8` + `…0406 0000`
- dim3(3dpoint):`db990003 000f 0003 ffff ff04 0000 5da8`
- 全量 hex 重抓:`go run ./tmp_debug/pseudo-spike/fulltdb4 test_data/pseudo_rich_demo.aep "Pseudo/711536-NNNN"`

## 5. ship-gate 技巧 / 坑(踩过的)

- **跑 gate**:`$env:AE_SHIP_GATE=1; go test ./internal/aep/ -run 'TestBuildPseudoEffect_AEShipGate_AE2020' -v -count=1`。手动单跑 JSX 用 `pwsh -File scripts/ae_run.ps1 -AeExe "<exe>" -Jsx <jsx> -Done <done> -TimeoutSec 120`(**参数名 `-AeExe` 不是 `-Exe`**)。
- **每次跑前** `Get-Process AfterFX* | Stop-Process -Force`(straggler → exit 6)。
- **AE exe**:2020=`E:\adobe\Adobe After Effects 2020\Support Files\AfterFX.exe`;2025 同构。
- **ExtendScript minValue 坑**:刚 fetch 的伪 slider 属性,直接读 `p.minValue` 返回 stale(=maxValue);**必须先碰 `p.hasMin`** 再读(maxValue 先碰 hasMax)。verify JSX 已处理。
- **属性访问**:用 `fx.property(index)` 比 `fx.property(matchName)` 稳(后者 minValue 也踩坑)。
- **本机码页**:cp1252(西欧),**渲染不了 GBK**——CJK 控件标签**无法在本机 ship-gate**,只能 byte-equivalence 单测 + 用户中文 AE 实机验。CJK demo 给用户:`go run ./tmp_debug/pseudo-spike/cjkdemo` → `tmp/pseudo_cjk_demo.aep`。
- **决定性 RE 法**:同一文件两种读法结果不同时,bisect 到单变量(minValue 坑就是这么定位的:probe 先碰 hasMin 报 -100、gate 冷读报 100)。

## 6. 红线提醒(CLAUDE.md)

- 红线7:只有**双版本 AE ship-gate PASS** 的能力才能宣称"能用",且仅在 gate 覆盖的规模/组合内。CJK 标签是例外(不可 gate,已诚实标 byte-equiv-only)。
- 红线4:渲染类能力 gate 必须验像素。伪效果控件是**参数容器不直接渲染**,值/范围 round-trip 是对的作用面(但 color 默认值我只验了字节布局、没验渲染——如果后续要宣称 color 默认"渲染对"需补像素验)。
- capindex CI 强制:新 API 加 `aep:cap` tag,tier=alpha 需 verify∈{roundtrip|ae-accept|render-pixel}+gate 名。改完 `go generate ./cmd/capindex` + `go run ./cmd/capindex -check`,docgen `go run ./cmd/docgen -manifest docs/docgen.json`。
- commit 直推 main,**永不 push 远端**。
