# 往 samples/ 加参考工程 — 文件夹约定 + 分类 — checklist

SUMMARY: 往 samples/ 加参考工程 — 文件夹约定 + 分类
READ WHEN: 往 samples/ 加参考工程当学习样本（建文件夹该长啥样、放哪些文件、怎么分类、readme 写啥）；批量采集后做归置；纠结某工程算哪个题材

---

> 给 [[understand-a-project]] 流水线攒原料的**通用归置约定**（与具体站点无关）。站点专属的「怎么扒下载链」见 `references/<站点>-harvest.md`（如 [[motionbox-harvest]]）。

## 目录形态（一项目一文件夹）

```
samples/<来源>/<题材>/<项目slug>/
    <原名>.aep        # 工程（保留原始文件名，日文 OK；zip 只取 .aep）
    cover.gif         # 封面（站点缩略图 / 预览首帧）
    readme.md         # 见下模板
samples/<来源>/README.md   # 该来源总表：来源站点 + 采集方式指针 + 按题材列全部项目
```

- `slug` = 标题转 ascii（小写、非字母数字→`-`、截 ~48 字符）；纯非 ascii 标题则用站点 `id` 兜底。
- `<来源>` 按出处命名（如 `motionbox/`）。`/samples/` **整体 gitignore** = 纯本地、不随仓库分发（尊重转载授权）。

## readme.md 模板（每项目）

```markdown
# <标题>
![cover](cover.gif)
- **来源**: <原始网址>
- **题材分类**: <见下>
- **下载数 / Like**: <n> / <n>     # 有则记，做热度参考
- **标签**: <原站标签>
- **授权**: 商用:<可/禁> · 転載:<可/禁⚠>   # no_commercial / no_reprint，必记
- **采集日期**: YYYY-MM-DD（<来源>公开下载，仅作本库解析/技法学习用）
## 工程文件
- `<名>.aep`  (<KB>)
## 制作环境
（原站 env，无则注「未注明」）
## 作者说明（原文）
（原站 description 原文照录）
```
末尾若 `no_reprint`：加一行 `> ⚠ 転載禁止 — 仅本地学习，勿再分发`。

## 题材分类（关键词启发式，兜底 motion-graphics）

对 `标题 + 标签 + 描述` 小写后按**优先序**首个命中：

| 题材 | 命中词（含日文） |
|---|---|
| `glitch` | glitch · グリッチ |
| `transition` | transition · トランジション · wipe · 切替 |
| `text-telop` | telop · テロップ · タイポ · typograph · kinetic · 字幕 · 歌詞 · タイトル · 文字 · テキスト |
| `vj` | vj · ループ · loop · realtime · リアルタイム |
| `reel` | reel · リール · showreel |
| `template` | template · テンプレート · 構図 · preset · プリセット · rig |
| `generative` | generative · ジェネレ · procedural · particle · particular · trapcode · パーティクル · node |
| `logo` | logo · ロゴ |
| `motion-graphics` | （兜底——标签太雷同时多数落这里，正常） |

> 启发式必然有偏（如标题含 REEL 但描述提「文字」会落 text-telop）。**分类只是归置、不是真相**；要重排：改本表口径 → 重跑采集脚本的构树步（字节有 id 缓存、不重下）。

## 卫生 / 验真（务必做）

1. **剔除 macOS AppleDouble 垃圾**：zip 常带 `._*` 文件 + `__MACOSX/`，会混入 178~234B 的假 `.aep`——`find <dir> -name '._*' -delete` + 删 `__MACOSX`。
2. **只取 AE**：`.aep` + `.zip`(内 .aep)；跳过 `prproj/blend/aup/dfx/c4d/drp/prfpset/mp4` 等非 AE。
3. **验真**：抽样 `go run ./cmd/aepdissect <f.aep>` 能解析（防错误页/损坏）；查无 `<20KB` 异常小文件（多半是错误页）。
4. **授权**：`no_reprint`/`no_commercial` 逐项记进 readme；版权归原作者。

## 下游

样本备齐 → 用 [[understand-a-project]] 流水线挑工程做三轴理解 / 抽技法。
