# Changelog

本文件记录面向外部用户的重要变化。项目尚未发布首个版本。

## Unreleased

### Added

- 5 个独立公开 SDK 示例与 3 个组合工程配方（人物信息条、动态标题、进度条）。

- 详细中文与英文 README，包含 SDK、CLI、Recipe、HTTP、兼容边界和项目知识入口。
- 可从模块外导入的根 Go package：`Open`、`Parse`、`Inspect`、`ProfileJSON`、`Write`。
- 统一 `aep` CLI：`inspect`、`profile`、`diff`、`migrate`、`capabilities`。
- RIFX parser 资源预算、typed format/limit errors、畸形输入测试和 fuzz gate。
- HTTP server 并发/timeout/panic 防护与 allowed path roots。

### Changed

- 公开展示迁至 `showcase/`，151 篇专题知识迁至 `docs/knowledge/`；发布源码包排除内部工作站。

- 首次公开发布前确定采用 PolyForm Noncommercial 1.0.0，提供非商业免费许可与单独书面商业授权路径；同步双语授权说明、贡献许可和发布附件。

### Fixed

- 文档生成不再记录构建工具链版本，避免 Go 安全补丁升级造成快照和索引无关漂移。
- 测试与自托管验证改用带生成来源的固定 fixture，不再依赖本地 showcase 输出；文档换行统一为 LF。

- ShapeLayer recipe transform 现在写入 runtime stream，并在版本迁移时保留静态值和关键帧。

### Security

- 最低 Go 版本提升至 1.25.13，修复 CI 检出的四个可达标准库漏洞（GO-2026-6090、GO-2026-6089、GO-2026-6088、GO-2026-5972）；CI 与发布构建继续统一读取 go.mod。
- chunk 声明尺寸在分配前与输入、父容器和资源预算核对。
- path input 默认关闭；启用时拒绝目录和 symlink 逃逸。

## Release policy

首次版本号和发布日期将在公开发布前确定；许可证已确定为 PolyForm Noncommercial 1.0.0（source-available）。发布后使用 Keep a Changelog 风格，并为每个 tag 增加对应版本段。
