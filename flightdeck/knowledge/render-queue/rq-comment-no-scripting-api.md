# ⚠ RenderQueueItem.comment 无 ScriptingAPI — ship-gate 走"接受+保留"而非 readback

SUMMARY: RenderQueueItem.comment 无 ScriptingAPI — ship-gate 走"接受+保留"而非 readback
READ WHEN: 给 RenderQueueItem 加 comment/字段 setter；设计某个"无 ScriptingAPI 的二进制字段"的 ship-gate；想用 JSX readback 验证 RQ 字段却拿到 undefined

---

## 事实

`RenderQueueItem` 在**任何** AE 版本都**没有 `comment` 属性**（核对 `charts/Types-for-Adobe/AfterEffects/8.0..26.0/index.d.ts`，class 成员里没有 `comment:`）。它是纯二进制字段 = Render Queue 面板的 **Comment 列**，AE 持久化进 `.aep` 的 `RCom` chunk，但 ScriptingAPI 不暴露读/写。

AE 2020 JSX `renderQueue.item(1).comment` 返回 `undefined`（不是 `""`），是这个 API 根本不存在的信号——freshly-added RQ item 也是 `undefined`，一眼可辨。

跟 [[camera-filmsize-ldta-write-blocked]] / [[variable-fonts-write-noop]] 同族（AE 持久化但 ScriptingAPI 隐藏），但**结论相反**：那两个是写不动，这个**写得动**——见下。

## 写得动：AE 接受 + byte 层保留

`RenderQueueItem.SetComment` 插入/替换 `RCom`（wrapper leaf 内嵌单 `Utf8`：`"Utf8"+BE u32 len+UTF-8 payload`，奇数 payload 补偶 pad；内层 Utf8 永远偶长 → RCom 自身永不需 pad）。双版本 ship-gate（AE 2020 + AE 2025）均：

1. **接受**：打开 Go 写的文件不崩、不报数据损坏、`numItems` 正常读出。
2. **保留**：AE resave 后 `RCom` 仍在，内嵌 Utf8 解出原 comment（byte-identical）。

## 教训：binary-only 字段的 ship-gate 方法论

字段无 readback API 时，常规"Go 写→AE 读回比对"gate 路径走不通（`comment` 永远 undefined，假阴性）。改用：

- **接受**：JSX 只断言 `app.open` 不抛 + `numItems>=1`，PASS = 打开成功（排除 mode 1/2/3 损坏，见 [[ae25-acceptance-gate]]）。
- **保留**：JSX 无条件 resave；**Go 端**解 resaved 文件的目标 chunk，byte 比对（证明 AE 没 silent-drop）。
- **base fixture 故意不含该字段**（这里：AE-2020-native RQ item 无 comment），让 gate 跑**插入**路径（最易被 AE 拒/丢的结构性添加），而非只测替换。

实现：`verify_rq_comment.jsx`（接受+resave）+ `render_queue_comment_shipgate_test.go`（resaved RCom 解码比对）+ `build_rq_ae2020.jsx`（base fixture builder）。

## 别踩

- 不要因为 `comment` 返回 undefined 就以为写失败——那是 API 缺失，不是写错。先 dump resaved 字节确认。
- 不要拿 py-aep 的 `RenderQueueItem.comment` property 当 AE ScriptingAPI 的证据——它是 py-aep 的二进制模型抽象，直接读 RCom，不经 AE 脚本层。
