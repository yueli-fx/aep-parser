---
status: active
when_to_read: 实现/调试 Layer.SetComment;给 from-scratch 层(NewSolidLayer/NewShapeLayer 等无原生 cmta 的层)设 comment 后 AE 读回空;排查"Go round-trip 绿但 AE DOM comment 丢";给任意层级插入新 cmta/任何 sibling chunk 时纠结插入位置;评估 item-level SetComment(comp/footage)是否同坑
applies_to: [layer, set-comment, cmta, append-position, layr-list, false-green, from-scratch, red-line-1, red-line-7, roundtrip-not-ae, sibling-slot, ship-gate, item-comment]
last_updated: 2026-06-17
resolved_by: codec.EncodeCmta double-NUL + back_layer.SetComment ldta @0x3C flag(批3 双版本 gated)
---

# Layer.SetComment from-scratch 层 AE 读回空 — 真因 = ldta @0x3C has-comment flag(+cmta double-NUL),非插入位置

## Signature
- symptom: `NewSolidLayer` 等 from-scratch 层调 `SetComment("x")` 后 WriteAEP,AE 打开 DOM `layer.comment` 读回**空串**;Go 自 parse 读回正确(假绿)。同一 fixture 里 `SetName`(也是 length-variable)、`SetStartTime`、`SetParent` 全部 AE 接受 + DOM 正确——单单 comment 丢。
- error_type: silent-drop(AE 不报错,只是忽略该 cmta)
- where: `internal/serializer/back_layer.go::SetComment` —— `commentChunk == nil` 分支 `b.layrList.Children = append(..., newCmta)`
- trigger: roundtrip→ae-accept 补验 arc 批3(2026-06-17),layer_av_fields3 fixture 首次给 from-scratch solid 设 comment 时撞到。

## 根因(RE 后真相 — 位置假说被推翻)

最初猜"append 位置错",**RE 推翻**:dump AE 原生 re_layer_comment.aep,cmta **也在 Layr LIST 末尾**(index 4,在 ldta/Utf8/tdgp/Gide 之后)——位置和我们一模一样。真因是两个字节级差异:

1. **cmta 少一个 NUL**:AE 原生 cmta = `content + \x00\x00`(double NUL,15 字符 comment → 17 字节);`EncodeCmta` 原本只写 single NUL(16 字节)。
2. **ldta @0x3C has-comment flag**:AE 原生设了 comment 的 layer ldta @0x3C = 0x01;from-scratch 层 = 0x00。**这才是关键** —— 单独把 cmta 改成 double-NUL(字节与 AE 逐字节一致)AE **仍读回空**,加上 @0x3C=1 才认。@0x3C 是此前未解的 ldta 字节(邻位 @0x3D=label 已知)。

诊断关键两步:① round-trip AE 原生文件 **byte-identical**(opaque preservation 完好)→ 排除 parse/write 破坏,锁定 SetComment 写路径;② `diff_ldta` 比对 from-scratch vs AE 原生 ldta,@0x3C 是唯一与 comment 相关的差异(其余差异是 from-scratch solid 默认值,name/startTime/parent 都被 AE 接受证明无关)。

红线1 活标本:cmta 字节全对(double-NUL)但缺一个 ldta flag,AE 静默丢——"值/字节对 ≠ AE 认"再添一例。`SetName` 没栽是因为 AE 原生 layer 本就有 name slot,无需额外 flag。

## 连带嫌疑:item-level SetComment 同款 append

`internal/serializer/write_item.go::setItemComment`(comp / footage 的 Item LIST comment)**也是 append 到末尾**(write_item.go:29)。Item LIST 子结构与 Layr 不同,末尾是否被 AE 认未验——但写法一模一样,**补验 comp/project 域时须重点验 item comment 的插入路径**(很可能同坑)。

## 修法(已修,批3 双版本 gated)

- `codec.EncodeCmta`:非空 comment 终止符 single → **double NUL**(empty 仍 single)。
- `back_layer.SetComment`:写/清 cmta 后设 ldta **@0x3C = 1(非空)/0(空)**。
- 批3 ship-gate(`layer_av_fields3`)把 SetComment 加回,AE 双版本 DOM `layer.comment` readback 非空 + resave 存活 PASS。`Layer.SetComment` roundtrip→ae-accept。

## item-level setItemComment(comp/footage)仍待补

`write_item.go::setItemComment` 也走 `EncodeCmta`(现已 double-NUL),但 item 挂在 Item LIST、没有 ldta —— 若 AE 同样需要某 **idta flag** 标记 item comment 存在,setItemComment 可能仍假绿(comp/footage SetComment 当前 verify=roundtrip,未真 AE 验)。补 comp/project 域时用同法 RE:AE 原生设 comp comment → 存盘 → diff idta 找 has-comment flag。

## Cases
- 2026-06-17 补验 arc 批3 首次发现。bisection 天然:同 fixture 4 setter,3 绿 1 红,红的就是 comment,根因 = 唯一走 append 新 chunk 路径的那个。
