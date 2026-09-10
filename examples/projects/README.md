# Project recipes / 组合工程配方

把多个基础能力组合成可修改的练习工程。三个配方均以 AE2020 为写入目标，1920×1080、30 fps，无第三方插件依赖。

Three editable practice projects combining basic capabilities, targeting AE2020 at 1920×1080 and 30 fps, without third-party plugin dependencies.

| 配方 / Recipe | 内容 / Contents | 可以改什么 / Try changing |
| --- | --- | --- |
| [lower-third.json](lower-third.json) | 人物姓名与职位、色条、底板，错开时间的位移和淡入 / Name, role, accent, panel, staggered entry | 姓名、职位、进场延迟、配色 / Text, entry timing, colors |
| [animated-title.json](animated-title.json) | 主标题、副标题、下划线，位移进出与透明度动画 / Title, subtitle, underline, entrance and exit | 标题、副标题、停留时间 / Copy and hold duration |
| [progress-bar.json](progress-bar.json) | 文字、轨道、从左向右伸展的进度条 / Label, track, left-anchored growing bar | 动画时长、长度、进度节点 / Duration, width, keyframes |

从仓库根目录运行 / Run from the repository root:

```sh
go run ./cmd/aeprecipe validate -recipe examples/projects/lower-third.json -json
go run ./cmd/aeprecipe compile -recipe examples/projects/lower-third.json -out tmp/lower-third.aep -json
go run ./cmd/aep inspect -in tmp/lower-third.aep
```

把文件名换成 `animated-title` 或 `progress-bar` 即可生成其他工程。也可以使用[公开 SDK 编译示例](../sdk/compile/main.go)。

Substitute `animated-title` or `progress-bar` for the other projects, or use the [public SDK compile example](../sdk/compile/main.go).

配方内的 `expected_profile` 检查合成、图层与文字图层数量。验证级别为 Recipe 校验、编译和重新解析；本批新组合工程尚未做 AE 画面验收，字体显示取决于目标机器。已有视觉验收记录的案例见 [showcase](../../showcase/README.md)。

Each `expected_profile` checks composition, layer, and text-layer counts. These new combinations are checked through Recipe validation, compilation, and reparsing; they have not received AE visual acceptance. Font appearance depends on the target machine. See [showcase](../../showcase/README.md) for existing visual verification records.
