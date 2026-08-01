# AE 效果参数字典 — dump 流程 + 使用指南 — checklist

AE 效果参数字典 — dump 流程 + 使用指南

> **这份字典是什么 / 为什么要它**:.aep 里效果参数只存 matchName(`ADBE Glo2-0002`)+ 无名数值,**parser 读得到存在、读不懂语义**。这份字典把 `matchName → 人类可读名 + 类型 + 默认值` 嚼好,让工具能把"一串无名数字"翻译成"用户在这个效果上调了哪几个旋钮"。是"读懂工程效果意图"的基础设施。
>
> **产物**:`data/effects-dict/effects_<lang>_<ver>.json`(分语言 + 分版本,各 ~470KB)。当前已有 5 版本英文(2020/2022/2023/2024/2025)+ 4 版本中文残档(2020/2023/2024/2025;统一英文后中文留作三语 join 用)。
> **工具**(tracked,可重生成):`scripts/effects-dict/dump_effects_dict.jsx` + `scripts/effects-dict/dump_effects_dict.ps1`(跑全版本 wrapper)+ `scripts/effects-dict/effect_matchnames_seed.txt`(种子 matchName 列表)。

---

## 第 1 部分:如何 dump 一份字典(给 2027/2028 新版本复用)

### 原理(为什么是 seed + addProperty,而不是"列出所有效果")

AE ExtendScript **没有"列出所有已装效果"的 API**。本方案用一份 **matchName 种子列表**(`scripts/effects-dict/effect_matchnames_seed.txt`,439 个,源自 yozya 三语 JSON + 本 reel 出现的第三方),对每个种子在临时 solid 上 `addProperty(matchName)`:成功 → 递归走它的参数树,记 `matchName + 当前语言显示名 + 类型 + 默认值`;失败(该版本没有 / 第三方没装)→ 跳过。**默认值取自刚 add、未改动的效果实例,所以就是 AE 默认值**。

### 标准流程(每个版本)

1. **(只 dump 英文时)把 AE 切英文** —— 改 `E:\adobe\Adobe After Effects <YEAR>\Support Files\AMT\application.xml` 里
   `<Data key="installedLanguages">zh_CN</Data>` → `en_US`(先 `.bak` 备份)。重启 AE 即英文 UI。`zh_CN`=中文,`en_US`=英文。
2. **人工开脚本写权限**(⚠ 见下方注意事项,这步**必须人工**) —— AE 里 `Edit > Preferences > Scripting & Expressions` → 勾 **"Allow Scripts to Write Files and Access Network"** → OK。**每个语言档第一次都要开一次**(切语言=新首选项档,默认关)。
3. **跑 dumper** —— `pwsh -File scripts/effects-dict/dump_effects_dict.ps1`(默认跑 E:\adobe 下 5 个版本;新版本在 `-AeExes` 数组里加 exe 路径)。每版本:kill 残留 → `ae_run.ps1` 跑 jsx → cold-start exit-2 warm-retry ×3。
4. **验证** —— done 日志 `ok=N fail=M skip=2`,产物 `data/effects-dict/effects_<lang>_<ver>.json` 非 0KB(~470KB)。再抽查 name 语言对、跨语言 matchName 集合一致(只差 skip 的 2 个 LUT)、同 key 默认值跨语言相同。

### ⚠ 注意事项(都是踩过的坑,2026-06-18)

- **脚本无法自己开"写文件"权限 —— Adobe 故意防自我提权。** `Pref_SCRIPTING_FILE_NETWORK_SECURITY` 用 `savePrefAsLong`/`savePrefAsString` + `saveToDisk` **都写不进磁盘**(实测盘值不变),因为脚本若能解除自己的写限制这个安全开关就形同虚设。**只能人工**:UI 勾选,或 AE 关闭时手改首选项文件
  `%APPDATA%\Adobe\After Effects\<ver>\Adobe After Effects <ver> Prefs.txt` 里
  `"Pref_SCRIPTING_FILE_NETWORK_SECURITY" = "0"` → `"1"`(注意它是**字符串型**,带引号)。
  - 症状:jsx 在 `File.write` 那行抛 `Permission denied (is Preferences > Scripting & Expressions > Allow Scripts to Write Files... enabled?)`;输出 **0KB** + done 文件空(`open("w")` 能创建空文件 → `ae_run` 误判 exit 0,**别被 exit 0 骗了,要查文件大小**)。
- **切语言 = 全新首选项档。** 中文用 `…Prefs.txt`?不 —— 中文档文件名是 `…设置.txt`,英文档是 `…Prefs.txt`,**两套独立**。所以中文档开过的写权限,切英文后不继承,要重开。
- **LUT 效果加时弹文件选择框,会挂起无人值守 run。** `ADBE Apply Color LUT` / `ADBE Apply Color LUT2` 一 `addProperty` 就弹 LUT 文件框。dumper 里已 `skip` 这俩(它们唯一"参数"是文件路径、无有意义默认值,跳过零损失)。**若新版本冒出别的"加效果即弹框"的效果**,加进 jsx 的 `skip` 表即可。
- **AE 2024 偶发 LUT/脚本超时弹框**:warm-retry(wrapper 已内置 ×3)能兜过;别当真 FAIL。
- **AE 2025 冷启动 splash 可能超 `ae_run` 的 unknown-grace(30s)→ exit-2**:warm-retry 兜;或先手动 warm 一次。
- **新版本(2027/2028)接入清单**:① 装好后确认 `E:\adobe\Adobe After Effects <YEAR>\Support Files\AMT\application.xml` 路径 + 版本号目录(如 `26.x`);② wrapper `-AeExes` 加该 exe;③ 切英文 + 人工开写权限;④ 跑;⑤ 种子列表若该版本有全新效果(本表抓不到),把新 matchName 追加进 `effect_matchnames_seed.txt` 再 dump 一次。
- **真相源**:matchName + 默认值都是**语言无关**的;新版本与旧版本 diff(新增/删效果、参数数变化)直接 `ConvertFrom-Json` 比对即可(实证:2020→2025 多了 7 个 OCIO/CC 效果)。

---

## 第 2 部分:如何使用这个字典

### 文件结构

```json
{
  "aeVersion": "17.7x45", "uiLanguage": "en_US",
  "effectCount": 423, "seedCount": 439,
  "effects": {
    "ADBE Glo2": {
      "name": "Glow",
      "params": {
        "ADBE Glo2-0002": { "name": "Glow Threshold", "type": "1d", "default": 153 },
        "ADBE Glo2-0012": { "name": "Color A", "type": "color", "default": [1,1,1,0] },
        ...
      }
    }
  }
}
```

- `type`:`1d / 2d / 3d / 2dSpatial / 3dSpatial / color / 1dEnum…`(color 默认是 `[A,R,G,B]` 0–1,group 无 default)。
- 子组(如 `ADBE Lumetri` 122 参数)已**递归展平**进同一 `params` 表。

### 三种用法

1. **翻译(matchName → 人类语义)** —— 把工程里读到的 `ADBE Glo2-0002 = 50` 显示成 `Glow Threshold = 50`。直接查 `effects[fxMatchName].params[paramMatchName].name`。

2. **配方信号(default → 识别用户调过哪些参数)⭐ 最大价值** —— AE 持久化规则:**参数值 == 默认就不写进 .aep(elision)**。反过来,**字节里出现的参数 = 用户真改过的**。配上 default 就能高亮:
   ```
   Glow Threshold = 50   [默认 153 → 用户调低了]   ← 配方关键旋钮
   ```
   一个效果 29 个参数,用户只动了 3 个 → 那 3 个非默认值就是这个技法的关键。这是把"读字节"变成"读懂意图"的核心,服务于从工程提取配方/技法。(注:也有少数参数因 synthesis/物化被强制写出,所以"出现≈调过"是强信号非铁律。)

3. **跨语言 / 跨版本 join** —— `matchName` 是**语言无关 + 跨版本稳定**的 key。三语显示名 = 拿同一 matchName 在 `effects_en_US` / `effects_zh_CN`(/ yozya 的 ja)里各取 `name` 拼。版本差异 = 两个版本字典按 matchName 集合 / 参数数 diff。

### 选哪份字典

按**目标工程的 AE 版本**选最接近的 `effects_<lang>_<ver>.json`(参数槽 index→名 在大版本间偶有变化,见 2020↔2025 diff)。无强约束时用最新(2025)英文档。

---

## 交叉链接

- 字典价值：它把 matchName 和默认值翻译成可比较的语义层，是 PARSE→理解 那步的基础设施。
- elision / 物化使"出现≈调过"非绝对；遇到默认参数未写、非默认参数被物化、或 effect param 需要合成时，要用 AE DOM/fixture 对照确认。
- 消费字典的工具:`cmd/aepdissect`(当前读 matchName + 值;接字典后可翻译 + 标非默认 = 待做)。
- yozya 三语基准 JSON(交叉验证用):`E:\projects\yozya\glossary\src\stores\effects_data.json`(2025、原生+Cycore、无第三方;本库 dumper 多了类型 + 子组递归 + 分版本 + 含本机第三方)。
