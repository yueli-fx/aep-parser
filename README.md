# aep-parser

**不启动 After Effects，用 Go 创建、修改和生成真正的 `.aep` 工程。**

**中文** · [English](README.en.md) · [日本語](README.ja.md)

## 几行代码，做一个带动画的 AE 工程

下面是[完整可运行示例](examples/authoring/main.go)的核心代码。在本仓库中运行，使用 `internal/aep` 创作 API；`must` / `check` 是示例自定义的错误处理辅助函数，定义见下方。

### 创建工程，添加文字和色块

```go
import "github.com/yueli-fx/aep-parser/internal/aep"

project := aep.NewProject(aep.TargetAE2020)
comp := must(aep.NewComposition(project, "Hello AEP", 1920, 1080, 30, 5))

must(aep.NewTextLayer(comp, "Title"))
must(aep.NewSolidLayer(comp, "Card", 640, 360, [3]float64{0.1, 0.8, 0.7}))
must(aep.NewSolidLayer(comp, "Draft", 100, 100, [3]float64{1, 0, 0}))
```

一个 **1080p、30 fps、5 秒**的合成，三个图层。接下来直接改文字、加效果。

上面的 `must` 接收 `(返回值, error)`，检查错误后取出返回值；`check` 用于只返回 `error` 的操作。这两个函数定义在 `main` 外，不属于 Go 内置函数或本库 API：

```go
// 示例脚本遇到错误就停止，避免带着错误继续写工程。
func must[T any](value T, err error) T {
    check(err)
    return value
}

func check(err error) {
    if err != nil {
        panic(err)
    }
}
```

### 添加高斯模糊，修改文字和效果参数

```go
// 在内存中重新解析，让新建图层进入可编辑状态，无需启动 AE。
project = must(aep.Reopen(project))
comp = project.Compositions[0]
check(comp.LayerByName("Title").SetText("HELLO, AEP"))

card := comp.LayerByName("Card")
blur := must(aep.AddEffect(card, aep.EffectGaussianBlur))
must(aep.SetEffectParam(card, blur, "Blurriness", 30.0))
```

这里把模糊量设为 **30**。参数可直接写英文 `"Blurriness"` 或中文 `"模糊度"`，库会解析为对应的内部标识。接口仅接受参数名称，不接受内部编号。高斯模糊只是 [231 种可添加效果模板](internal/serializer/mutate_effect_add.go)之一；[效果示例](showcase/effects/gen.go)还展示了阴影、描边、调色等参数写法。

### 让模糊动起来，删除草稿层，保存工程

```go
// 第一秒从模糊 30 变成 0。
must(aep.AnimateEffectParam(card, blur, "Blurriness",
    []aep.ScalarKeyframe{{Time: 0, Value: 30}, {Time: 1, Value: 0}}))

for index, layer := range comp.Layers {
    if layer.Name == "Draft" {
        check(aep.DeleteLayer(comp, index)) // 从 0 开始的索引；这里删除 Solid 图层。
        break
    }
}

file := must(os.Create("hello.aep"))
check(project.WriteAEP(file))
check(file.Close())
```

得到一个包含文字、色块和模糊关键帧的 `.aep`。**创建、修改和写文件全程不需要安装 AE**；打开编辑与画面渲染交给 AE。完整程序另含写后回读检查，并拒绝覆盖已有文件。

### 现在就跑

```sh
git clone https://github.com/yueli-fx/aep-parser.git
cd aep-parser
go run ./examples/authoring -out tmp/hello.aep
go run ./cmd/aep inspect -in tmp/hello.aep
```

需要 Go 1.25.12 或同系列更新补丁版本。上面的创作 API 用于仓库内程序；接入自己的 Go module 时，使用[公开 SDK 示例](examples/sdk/README.md)中的 `Open`、`Export`、`Compile` 等入口。

## 不止一个模糊色块

- **[程序化火焰](showcase/procedural-fx/gen.go)**：三层噪声火舌、分层色温、羽化遮罩、Add 合成与 Glow，生成持续翻腾的火焰工程。[验收记录](showcase/procedural-fx/INDEX.md)
- **[3D 相机场景](showcase/3d-camera/gen.go)**：从零创建相机和不同深度的 3D 卡片，控制 Z 视差与 Y 轴透视旋转。[验收记录](showcase/3d-camera/INDEX.md)
- **[矢量形状动画](showcase/shape-filters/INDEX.md)**：Trim Paths、Repeater、ZigZag、Wiggle 等 11 类滤镜及组合。
- **[表达式驱动](showcase/expressions/INDEX.md)**：跨层引用、loopOut、wiggle 与 slider 控制动画。

[全部 25 个 showcase →](showcase/README.md) · [人物信息条、动态标题、进度条配方 →](examples/projects/README.md)

不想写 Go，也可以直接编译 JSON 配方：

```sh
go run ./cmd/aeprecipe compile -recipe examples/projects/animated-title.json -out tmp/title.aep -json
```

## 已有工程也能读、比较和迁移

```sh
# 列出合成、图层、效果、关键帧等工程信息
go run ./cmd/aep profile -in project.aep

# 比较修改前后的工程
go run ./cmd/aep diff -expected before.aep -actual after.aep

# 将支持的工程语义迁移到 AE2025
go run ./cmd/aep migrate -in source.aep -target AE2025 -out tmp/migrated.aep
```

批量盘点、提取属性、保留式修改数值参数，见[五个公开 SDK 程序](examples/sdk/README.md)。想读懂一个复杂工程用了哪些技法，可以生成[HTML 技法报告](docs/self-hosted-reports.md)，查看逐层效果栈、复刻步骤和学习任务。

## 这背后实现了多少

**500 个函数/方法能力条目 · 231 种效果模板 · 16 个能力领域 · AE2020–AE2025 六个写入目标。**

从 RIFX 容器、嵌套 chunk 和字节布局，到图层结构、2D/3D 变换、文字样式、形状与渐变、关键帧与缓动、表达式、效果和遮罩，都已有对应实现。逆向发现和踩过的坑整理成了 **151 篇公开知识笔记**，另有 **151 个原子 Recipe 示例**可运行。

统计来自[能力索引](docs/capabilities.json)与[效果注册表](internal/serializer/mutate_effect_add.go)：382 个方法 + 118 个函数，其中 438 项标记 stable、62 项 alpha；266 项标注 ae-accept，163 项标注 render-pixel。这些数字统计底层能力和已有验证记录，具体接口与边界见[能力矩阵](docs/capabilities.md)。[迁移回归记录](registry/evidence/versioned-aep-migration/migration_matrix_verify/smoke_all/ledger.md)包含 906 个组合：900 通过、6 跳过、0 失败。

## 继续探索

| 想做什么 | 入口 |
| --- | --- |
| 用 Go / HTTP 集成 | [使用指南](docs/usage.md) · [SDK 示例](examples/sdk/README.md) |
| 用 JSON 批量生成工程 | [Recipe 示例](examples/recipes) · [字段参考](docs/recipe.md) · [Schema](docs/recipe_schema.json) |
| 学习 AEP 逆向 | [知识库](docs/knowledge/README.md) · [API 参考](docs/README.md) |
| 查支持范围与兼容性 | [能力矩阵](docs/capabilities.md) · [兼容与限制](docs/usage.md#兼容范围与限制) |
| 参与开发 | [贡献指南](CONTRIBUTING.md) · [验证流程](docs/knowledge/workflow/verify.md) · [发布审核](docs/open-source-audit.md) |

项目处于预发布阶段。解析与生成核心为纯 Go，可构建于 Windows、macOS、Linux。知识笔记主要为中文，生成的 API / Recipe 参考主要为英文。新增组合示例已通过生成与回读检查，尚未单独做 AE 画面验收；各 showcase 的宿主验证情况见各自记录。

## 许可证

**[PolyForm Noncommercial 1.0.0](LICENSE)**。许可范围内的非商业用途免费；商业使用、收费转售或商业集成等超出免费许可的用途，须先取得单独书面商业授权。

本项目公开源码，属于 **source-available**。免费范围、机构例外、商业授权申请和生成内容说明见 **[使用与商业授权](LICENSING.md)**。第三方声明见 [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md)，安全问题见 [SECURITY.md](SECURITY.md)。

本项目不隶属于 Adobe。Adobe、After Effects 及相关商标属于各自权利人。
