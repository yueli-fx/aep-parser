---
status: active
when_to_read: 新增/调试任何写 btdk PostScript point/measurement 值的 text setter(run 或 paragraph 的 indent/spacing/shift/scale/width/tsume);Go round-trip 绿但 AE DOM 读回的值≈写入值/65536;判断某 btdk key 该用 FormatPSReal 还是 FormatPSNumber
applies_to: [text, btdk, FormatPSReal, FormatPSNumber, 16.16-fixed-point, 65536, false-green, red-line-1, SetRunTsume, SetParagraphIndent, point-value, coolType, roundtrip-not-ae]
last_updated: 2026-06-17
resolved_by:
---

# btdk point-measurement setter 必须 FormatPSReal 否则 AE 读成 value/65536 (假绿)

## Signature
- symptom: `AE DOM 读回的 text 属性值 ≈ 写入值/65536(如写 20 → AE 读 0.000305=20/65536);Go round-trip 却绿`
- error_type: —
- where: `internal/serializer/back_layer.go` SetRun*/SetParagraph* 经 splicePSValue
- trigger: 用 codec.FormatPSNumber 写一个 AE 当 REAL 读的 btdk point/measurement key

## 症状/复现

写某 text run/paragraph 的 point 值(em points / 缩放 / 字距),`Go round-trip 绿`(parser 读回写入值),但 **AE DOM 读回 ≈ 写入值 / 65536**。例:`SetParagraphFirstLineIndent(20)` → AE `textDocument.firstLineIndent` = 0.000305 = 20/65536;`SetRunTsume(50)` → 0.00076 = 50/65536。**红线1 假绿温床**:Go 自读回对、从没让真 AE 消化。

## 根因

btdk 是 PostScript 文本体。`FormatPSNumber(20)` 写 `"20"`(整数无小数点);**AE/CoolType 把一个无小数点的 number 当 16.16 定点整数读 → value/65536**。`FormatPSReal(20)` 写 `"20.0"`(强制带小数点)→ AE 当 REAL 读 → 20.0。

- **point/measurement key 必须 FormatPSReal**:baselineShift `/9`、leading `/5`、h/vScale `/7`、strokeWidth、tsume `/36`、paragraph indent/spacing `/1../5`。
- **整数/枚举 key 用 FormatPSNumber 是对的**:tracking `/8`(实测 render-gate 过,确为 int)、justification/leadingType/direction(枚举 int)、fontIndex。
- **bool key** 写 `"true"/"false"`。
- 区分靠该 key 在 AE 里是不是「带小数的测量量」。拿不准 → 建单 run/单段 from-scratch text + 该 setter,AE DOM 读回:若 ≈ /65536 就是漏了 FormatPSReal。

## 修法

把该 key 的写入从 `codec.FormatPSNumber(v)` 换成 `codec.FormatPSReal(v)`。**Go parser 读数值不受小数点影响,round-trip 不破**(已验)。修后必须双版本 AE gate 复验(DOM 值验;段落 indent 的 leftMargin/rightMargin AE 无 DOM,靠 resave-preservation)。

## Cases
- 2026-06-17 首次 — 批7 `SetRunTsume`(`/36`,FormatPSNumber→Real),`TestTextRun_AEShipGate`。
- 2026-06-17 复发 — 批7b 5 个 `SetParagraph{FirstLineIndent,StartIndent,EndIndent,SpaceBefore,SpaceAfter}`(`/1../5`,FormatPSNumber→Real),`TestTextPara_AEShipGate`。两次都是「Go round-trip 绿 + AE /65536」同款假绿。
