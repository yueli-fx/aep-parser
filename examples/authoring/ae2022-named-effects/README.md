# AE2022：按参数名称生成效果验收工程

从仓库根目录运行，无需 AE：

```sh
go run ./examples/authoring/ae2022-named-effects -out tmp/ae2022-named-effects.aep
```

已有同名输出时程序会拒绝覆盖，可用 `-out` 指定新文件名。[生成源码](main.go)只使用参数名称调用 `SetEffectParam`、`AnimateEffectParam`、`AnimateEffectParamVec`，没有参数编号或编号兼容入口。

## 在 AE2022 中检查

1. 打开生成的 `.aep`，双击 **00 OVERVIEW - AE2022 - 4 seconds**。这是 1920×1080、30 fps、4 秒的总览。
2. 对比 **0 秒、2 秒、3.9 秒**。大部分视觉动画在 2 秒到达另一个值，3.9 秒回到初始值。
3. 双击任意单项合成，选中图层，按 **F3** 看效果控件、按 **U** 看关键帧。
4. 打开 **13 CONTROLS - inspect Effect Controls**，选中唯一图层。这里的五个表达式控件用于检查数值与关键帧，没有连接表达式，不会自行改变画面。

总览按从左到右、从上到下排列：

| 第一列 | 第二列 | 第三列 | 第四列 |
| --- | --- | --- | --- |
| 01 原始色块 | 02 英文名高斯模糊 | 03 中文名高斯模糊 | 04 阴影 |
| 05 色调映射 | 06 填充颜色动画 | 07 渐变起点动画 | 08 分形噪声 |
| 09 方向模糊 | 10 亮度/对比度 | 11 Glow | 12 湍流置换 |

## 重点核对

- **02 / 03**：画面应一致，模糊度从 **30 → 0 → 30**。分别使用 `Blurriness` 与 `模糊度`，同时设置了方向枚举和重复边缘像素开关。
- **04**：暖色阴影，方向 **45° → 315° → 45°**；同时设置颜色、不透明度、距离、柔和度与仅阴影开关。
- **06**：红色 → 蓝色 → 红色，验证四分量颜色关键帧。
- **07**：径向渐变起点从左上移到右上再返回，验证二维点关键帧。
- **08 / 12**：Evolution 动画；**09 / 11**：模糊长度和辉光强度动画。
- **13 控件合成**：0–2 秒，Angle **0 → 180°**；滑块 **0 → 100**；颜色红 → 蓝；二维和三维点坐标变化。点参数的源码值按底层坐标比例写入，AE 面板显示像素。

覆盖 **14 个合成、16 个效果实例（15 种效果）、40 个参数，其中 14 个有动画**，包含数值、角度、枚举、开关、颜色、二维点和三维点。

生成器会重新解析最终文件，逐项核对 40 个参数及其关键帧，并输出同名 `.coverage.json` 清单。已通过 Go 写入和回读验证；**2026-09-11：用户反馈 AE2022 本机测试通过**。这是本次组合工程的人工验收反馈，未附逐项截图或自动化宿主报告。

## English

Run the command above, then open the file in AE2022. Start with the `00 OVERVIEW` composition and compare frames at 0, 2, and 3.9 seconds. Open individual compositions and use Effect Controls / U to inspect parameters and keyframes. The English and Chinese Gaussian Blur examples should match. Composition `13 CONTROLS` contains five animated expression controls; these are intentionally not connected to visual properties.

The generator uses parameter names only. It reparses the output and checks all 40 configured parameters before saving. On 2026-09-11, the user reported that local AE2022 testing passed. This records manual acceptance of this combined project; no per-parameter screenshots or automated host report were supplied.
