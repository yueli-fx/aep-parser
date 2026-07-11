# Changelog

本文件记录面向外部用户的重要变化。项目尚未发布首个版本。

## Unreleased

### Added

- 可从模块外导入的根 Go package：`Open`、`Parse`、`Inspect`、`ProfileJSON`、`Write`。
- 统一 `aep` CLI：`inspect`、`profile`、`diff`、`migrate`、`capabilities`。
- RIFX parser 资源预算、typed format/limit errors、畸形输入测试和 fuzz gate。
- HTTP server 并发/timeout/panic 防护与 allowed path roots。

### Fixed

- ShapeLayer recipe transform 现在写入 runtime stream，并在版本迁移时保留静态值和关键帧。

### Security

- chunk 声明尺寸在分配前与输入、父容器和资源预算核对。
- path input 默认关闭；启用时拒绝目录和 symlink 逃逸。

## Release policy

首次版本号、发布日期和许可证将在公开发布前确定。发布后使用 Keep a Changelog 风格，并为每个 tag 增加对应版本段。
