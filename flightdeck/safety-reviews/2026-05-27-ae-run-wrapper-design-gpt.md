2. 当前 spec 里最大的风险点

这里是重点。

2.1 最大风险：SendKeys 焦点竞争

这是当前 spec 最大的不稳定因素。

你自己已经意识到了：

SendKeys simple but race

但我觉得风险比 spec 描述的大。

因为 AE 很容易：

startup splash 抢焦点
plugin window 抢焦点
modal activation 延迟
RDP session 焦点漂移

你现在：

OCR 命中
 -> SendKeys

缺少：

SetForegroundWindow(hwnd)
WaitActive(hwnd)
Verify focus

否则：

Enter 发给 timeline
Space 发给 preview
Tab 发给 project panel

都会发生。

建议（强烈）

在 spec 里补：

SendKeys 前必须：
1. SetForegroundWindow(hwnd)
2. Sleep 100~300ms
3. Verify GetForegroundWindow()==hwnd

否则后面会出现大量 flaky。

2.2 “窗口标题不是 AE 主窗口” 不够稳

AE 的 modal 有时：

owner 是主窗口
title 为空
title 与主窗口一致
class 特殊

仅靠：

title != main window

可能不够。

建议补：

GetWindow(hwnd, GW_OWNER)
GetClassName(hwnd)
WS_EX_DLGMODALFRAME

至少留接口。

否则后面你会发现：

某些 modal enum 不出来

然后又要返工 detection 层。

2.3 OCR 不应该作为 primary identifier

这个是架构层建议。

现在 spec 有点：

OCR -> match rule

但实际上：

OCR 容易 locale 漂
不同 DPI 会误识别
中英文混排会变
字体抗锯齿会影响结果

真正稳定的 identifier 应该优先：

window class
automation id
button text
window title

OCR 应该是：

最后 fallback

不是 primary key。

建议调整（重要）

把 rule schema 改成：

{
  "name": "...",
  "windowTitle": [],
  "windowClass": [],
  "ocrMatch": [],
  "action": ...
}

匹配顺序：

title/class
 -> OCR fallback

这会让系统稳定很多。

2.4 “Save changes” 的 Tab 序列很脆

这是第二大风险。

{TAB}{TAB}{ENTER}

太依赖：

locale
button order
focus default
accessibility behavior

AE 不同版本可能变。

更稳方案

建议 spec 留一个 future path：

preferred:
  UIAutomation Invoke(buttonName)

fallback:
  SendKeys

即使 v1 不做 UIA，也要把 architecture 留出来。

否则：

以后换 UIA 时 dispatch schema 会重构。

2.5 “done 出现后 wait 1s” 有 race

这里也有风险。

可能：

.done created
但 file handle 还没 flush

或者：

JSX 还在 app.quit
AE modal 正在出现

建议：

不是：

wait 1s

而是：

poll file stable size/mtime

例如：

连续 500ms size 不变

再 cleanup。

更 deterministic。

3. 建议补强的地方
3.1 建议加 “rule cooldown”

否则可能：

dialog 还没消失
loop 再次 OCR
再发一次 Enter

建议：

最近 N 秒内同 hwnd + same rule 不重复触发

比如：

cooldown = 2s

很重要。

3.2 建议记录 action log

dump 里最好再有：

actions.log

例如：

12:00:01 matched convert-old-project
12:00:01 sent {ENTER}

后面 debug 会非常有价值。

3.3 建议定义 “modal detection cadence”

你写：

500ms cadence

但 OCR 本身可能 200~500ms。

建议明确：

single-threaded acceptable

否则以后有人可能想并发 OCR。

3.4 建议明确 process tree

AE 有时会：

CEP helper
dynamiclinkmanager
crash reporter

你写：

pid 属于 AE 进程树

建议明确：

root AfterFX pid only?
还是 owner pid?

否则 enum filtering 后面会歧义。

4. 我认为最该提前改的一个地方

如果只改一个，我会建议：

把 dispatch rule 从 OCR-only 改成 multi-signal

即：

{
  "name": "...",
  "windowTitle": [],
  "windowClass": [],
  "ocrMatch": [],
  "keys": ...
}

因为这是：

最影响长期稳定性
最难后补
最影响 false positive rate

的一层。

现在改最便宜。

5. 总体评价

这是一个：

可以进入 planning/implementation 的 spec

不是草稿级了。

尤其优秀的是：

边界控制好
不侵入 ship-gate contract
forensic 意识强
fail-loud 思维正确
rollout 路径清晰

真正需要警惕的是：

GUI automation 最大敌人不是 OCR，
而是 focus / activation / timing race。

目前 spec 对：

foreground ownership
modal activation
repeat-trigger suppression

写得还不够硬。

把这些补上后，这份 spec 就会非常扎实。