# ⚠ Full-repo gofmt gate does not match the current baseline

历史 Go 文件仍有大量 gofmt diff；CI 只对新的公开/安全产品面做格式门禁，全仓格式化必须作为独立机械重写处理。

首次建立 CI 时，`gofmt -l .` 会命中数十个历史文件，横跨 facade、scene、showcase、debug tools 和 vendored/extracted sample。直接把全仓格式检查设为必过会让 CI 在没有新格式问题时也永久失败；直接格式化全部文件则会把大规模无行为 diff 混入产品化提交。

当前 CI 只检查新的公开和安全产品面：根 SDK、`cmd/aep`、`internal/rifx`、`internal/server`、`internal/toolkitcli`。扩展这些 module 时应继续保持 gofmt clean。

若要建立全仓格式门禁，先单独开机械重写包：排除 `data/samples/**` 等外部语料，运行 gofmt，验证全仓测试与生成文档，再用单独 commit 建立新 baseline。不要把它混入功能、修复或发布提交。
