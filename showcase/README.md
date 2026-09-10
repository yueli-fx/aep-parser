# Showcase

25 个展示方向：从基础图层与形状动画，到 3D 相机、程序化火焰和效果组合。每个方向保留生成器、说明与 AE 验证脚本。

25 demonstration areas, from layers and shape animation to 3D cameras, procedural fire, and effect stacks. Each includes source and its verification notes.

## 先看这些 / Start here

| 展示 | 看点 |
| --- | --- |
| [程序化火焰 / Fire](procedural-fx/INDEX.md) | 多层噪声、遮罩、色温、Glow / Layered noise, masks, color, glow |
| [3D 相机 / Camera](3d-camera/INDEX.md) | 深度视差和透视旋转 / Depth parallax and perspective |
| [形状滤镜 / Shape filters](shape-filters/INDEX.md) | 11 类矢量滤镜与组合 / Eleven vector filter families |
| [渐变 / Gradients](gradient/INDEX.md) | 线性、径向、填充与描边 / Linear/radial fills and strokes |
| [缓动 / Easing](keyframes-ease/INDEX.md) | 不同时间曲线的运动 / Motion with different timing curves |
| [表达式 / Expressions](expressions/INDEX.md) | 跨层引用、loopOut、wiggle / Cross-layer references and animation |
| [文字 / Text](text/INDEX.md) | 文字图层与样式 / Text layers and styling |
| [效果 / Effects](effects/INDEX.md) | 添加效果与参数调整 / Effect insertion and parameters |

[全部方向、历史验收状态和已知边界 / Full index and verification records](INDEX.md)

## 生成工程 / Generate a project

从仓库根目录运行，无需启动 AE / Run from the repository root without AE:

```sh
go run ./showcase/gradient
go run ./showcase/3d-camera
go run ./showcase/procedural-fx
```

输出位置由各目录说明列出，通常为当前展示目录内的 `.aep`。生成工程与渲染图片不提交到 Git。

Outputs are documented per example, usually an `.aep` inside its showcase directory. Generated projects and images are excluded from Git.

渲染或宿主读回需要安装 AE。按对应 `INDEX.md` 使用 `render.jsx` / `verify.jsx` 和 `scripts/ae-worker/ae_run.ps1`。脚本通过自身位置定位展示目录，不要求维护者的本机工程路径。

Rendering/readback requires AE. Follow each `INDEX.md` for its `render.jsx` or `verify.jsx` and the AE worker. Showcase-local paths resolve relative to the script itself.

`booyah-clone` 等参考复刻实验需要自行提供原始输入；不能作为干净 clone 的入门示例。各方向验证状态保持原记录，路径迁移不代表新增宿主验收。

Reference reconstruction experiments such as `booyah-clone` require separately supplied inputs. Moving directories does not change their recorded acceptance status.

[公开 SDK 示例](../examples/sdk/README.md) · [组合工程配方](../examples/projects/README.md) · [逆向知识](../docs/knowledge/README.md)
