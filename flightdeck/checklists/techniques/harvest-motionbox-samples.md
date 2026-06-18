---
status: active
when_to_read: 想批量采集 motion-box.net（或同类 Nuxt+Firebase 前端）的 AE 参考工程到 samples/；下载按钮无静态链接、点击不触发下载、要找文件直链；firestore REST 取数据被 403 挡；要给 understand-a-project 流水线补充原料工程
applies_to: [sample-harvest, motionbox, motion-box.net, nuxt, firebase-storage, firestore, vue-state, dlURL, download-reverse-engineer, aep-samples, reference-projects, playwright, rest-403, no-reprint, license]
last_updated: 2026-06-18
---

# 采集 motion-box.net 的 AE 参考工程到 samples/

> 给 [[understand-a-project]] 流水线喂原料：从 MOTIONBOX（https://motion-box.net/，日本 motion graphics 社区，作者免费分享工程）批量拉真实 AE 工程当学习样本。机制可推广到**任何 Nuxt + Firebase 前端**。
> 产物落 `samples/motionbox/<题材>/<项目>/`（`.aep` + `cover.gif` + `readme.md`）。`/samples/` 整体 gitignore = 纯本地、不再分发（尊重转载授权）。

## 站点机制（逆向结论，省得重新摸）

- **下载键无静态链接**：是 Vue 的 `<a class="main-btn">`（无 href），点击走 JS handler。Playwright 点它**不产生可见下载**（需登录/弹窗机制），别在这死磕。
- **`.aep` 直链藏在页面 Vue 组件状态**：`boxInfo.dlURL` = 带 token 的 firebase storage 公开链（`posts/<id>/DLfile/<名>.aep?alt=media&token=…`），**公开可 curl，无需登录**。`thumnailURL`=gif 封面、`previewURL`=mp4 预览，同样直链。
- **Firestore REST 直读被安全规则挡**：`GET firestore.googleapis.com/v1/projects/motion-box-44ad6/databases/(default)/documents/<col>/<id>` → 403（collection 存在但拒读）；`posts` → 404（真正 collection 名不是 posts）。**别走 REST，从渲染好的页面 Vue state 抠**。
- **列表是 top-N 排行**：`/popular` 初始 12 条，底部「もっと見る」= `div.index-btn.sub-btn`，**Playwright locator 点击**每次 +12（页内 `el.click()` 有时点不动，用 locator）。点到按钮 `offsetParent===null` 即到底（popular 全量约 149 卡 / Vue 数组 ~173 含重复，按 `id` 去重）。

## 步骤

1. **载全量**：Playwright 开 `/popular` → 循环 locator 点 `div.index-btn.sub-btn` 到按钮消失。
2. **抽数组**：`browser_evaluate` 遍历 DOM 找 `el.__vue__.$data` 里最大的「含 `id`+`dlURL` 的数组」，导出每项 `{id,title,fileExt,dlURL,thumnailURL,env,description,tags_name,dlCount,likePostCount,no_commercial,no_reprint}`（按 `id` 去重）。`creater` 是 firestore 引用（循环对象），别 JSON.stringify 整个 boxInfo。
3. **过滤**：只留 `fileExt ∈ {aep, zip}`。**跳过非 AE**：`prproj`(Premiere) `aup`(Audition) `prfpset` `blend`(Blender) `dfx`(Fusion) `c4d` `drp`(DaVinci) `mp4` 等。
4. **下载（id 缓存、幂等）**：curl `dlURL` → `tmp/mb_cache/<id>.<ext>`（已存且非空则跳过——重跑不重下、对服务器友好，token 暂稳定）；封面 `thumnailURL` → `<id>.gif`。
5. **构树**：题材分类（关键词启发式：glitch/transition/text-telop/vj/reel/template/generative/logo，兜底 motion-graphics）→ `samples/motionbox/<题材>/<slug>/`。zip **只取 .aep**，并**剔除 macOS AppleDouble 垃圾**（`._*` 文件 + `__MACOSX/`，否则会混入 178~234B 假 .aep）。
6. **每项 readme.md**：来源网址 + 题材 + DL/Like + 标签 + **授权（`no_commercial`/`no_reprint`）** + 采集日期 + 制作环境 + 作者原文。总 `samples/motionbox/README.md` 按题材列表。
7. **验真**：抽样 `go run ./cmd/aepdissect <f.aep>` 能解析（防错误页/损坏）；查无 `<20KB` 异常小文件。

## 脚本

`tmp/fetch_motionbox_all.py`（一次性、gitignore）= 步骤 4–6 全自动。清单 `tmp/motionbox_all.json`（步骤 2 导出）。改分类口径只需重跑——字节在 `tmp/mb_cache/` 不会重下。**要重采或换站点时，照本文重建脚本即可**（脚本本身是 scratch，不进 git）。

## ⚠ 授权 / 礼仪

- 各工程 `no_reprint`(転載禁止) / `no_commercial` 必须记进 readme；因 `/samples/` 不进 git，本地学习/解析合规，**勿再分发**。
- curl 间加小 sleep；用 id 缓存避免重复 hammer。版权归原作者。

## 推广

同套路适用任何 **Nuxt + Firebase storage** 站点：渲染页 → 抠 Vue state 里的 storage 直链（带 token）→ 直接 curl，绕开「点击下载」黑盒与被锁的 REST。关键是**找到承载下载链的那个 state 字段**（这里是 `boxInfo.dlURL`）。
