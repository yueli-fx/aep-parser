---
status: active
when_to_read: 实现/调试 Layer.SetComment;给 from-scratch 层(NewSolidLayer/NewShapeLayer 等无原生 cmta 的层)设 comment 后 AE 读回空;排查"Go round-trip 绿但 AE DOM comment 丢";给任意层级插入新 cmta/任何 sibling chunk 时纠结插入位置;评估 item-level SetComment(comp/footage)是否同坑
applies_to: [layer, set-comment, cmta, append-position, layr-list, false-green, from-scratch, red-line-1, red-line-7, roundtrip-not-ae, sibling-slot, ship-gate, item-comment]
last_updated: 2026-06-17
resolved_by:
---

# Layer.SetComment 在无原生 cmta 的层上 append cmta 到 Layr LIST 末尾 → AE 读回空(假绿)

## Signature
- symptom: `NewSolidLayer` 等 from-scratch 层调 `SetComment("x")` 后 WriteAEP,AE 打开 DOM `layer.comment` 读回**空串**;Go 自 parse 读回正确(假绿)。同一 fixture 里 `SetName`(也是 length-variable)、`SetStartTime`、`SetParent` 全部 AE 接受 + DOM 正确——单单 comment 丢。
- error_type: silent-drop(AE 不报错,只是忽略该 cmta)
- where: `internal/serializer/back_layer.go::SetComment` —— `commentChunk == nil` 分支 `b.layrList.Children = append(..., newCmta)`
- trigger: roundtrip→ae-accept 补验 arc 批3(2026-06-17),layer_av_fields3 fixture 首次给 from-scratch solid 设 comment 时撞到。

## 根因

`SetComment` 插入新 cmta 时 **append 到 Layr LIST 的 Children 末尾**(back_layer.go:385)。AE 期望 layer 的 cmta 在 Layr LIST 内的**特定 sibling 槽位**(未 RE 确认精确 index,但显然不是末尾——末尾在属性组 tdgp / 效果 parade 之后)。AE 解析 Layr 时不在预期位置找到 cmta → 静默忽略。

对比为何 `SetName` 成功:它**替换现有 Utf8 name chunk 的 Data**(位置不变),不新增 chunk。`SetComment` 在**有原生 cmta 的 parsed layer** 上同样走"替换 Data"分支(back_layer.go:379),位置不变 → 应当 OK(未单独验)。只有**无原生 cmta**(from-scratch 层 / 真实层从未设过 comment)才走 append 分支 → 坏。

这是红线1(Go round-trip 绿 ≠ AE 接受)+ 红线7(from-scratch 弱项)的双重活标本:库强项"局部改有原生 comment 的真实工程"可能没事,弱项"from-scratch 插新 cmta"暴露。

## 连带嫌疑:item-level SetComment 同款 append

`internal/serializer/write_item.go::setItemComment`(comp / footage 的 Item LIST comment)**也是 append 到末尾**(write_item.go:29)。Item LIST 子结构与 Layr 不同,末尾是否被 AE 认未验——但写法一模一样,**补验 comp/project 域时须重点验 item comment 的插入路径**(很可能同坑)。

## 修法(待做,需 RE)

1. RE AE 原生位置:JSX 建 solid + `layer.comment="X"` → 存盘 → Go parse,dump Layr LIST children 的 ID 顺序,看 cmta 的 index(相对 ldta / Utf8 name / tdgp 的位置)。
2. 把 append 改成 **插在正确槽位**(大概率紧跟 ldta 或 name Utf8 之后,在属性组之前)。
3. 重跑批3 把 SetComment 加回 fixture,双版本 gate 确认 AE DOM comment 非空。
4. 同步修 + 验 `setItemComment`(item 版)。

## 现状(诚实标注)

- `Layer.SetComment` 维持 `verify=roundtrip`,boundary 已标 from-scratch 假绿 + 指向本 incident。
- 批3 ship-gate(`layer_av_fields3`)已把 SetComment 剥离,只 gate SetName/SetStartTime/SetParent(三者双版本 PASS)。

## Cases
- 2026-06-17 补验 arc 批3 首次发现。bisection 天然:同 fixture 4 setter,3 绿 1 红,红的就是 comment,根因 = 唯一走 append 新 chunk 路径的那个。
