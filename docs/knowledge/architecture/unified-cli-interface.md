# Unified CLI interface

新用户工作流统一进入 `cmd/aep`，由 `internal/toolkitcli` 集中提供 JSON envelope、退出码和 inspect/profile/diff/migrate/capabilities；旧命令仅保留兼容与高级用途。

统一 CLI 的进程 adapter 是 `cmd/aep`，实现 module 是 `internal/toolkitcli`。module interface 只有 `Run(args, stdout, stderr) int`，因此命令解析、JSON 输出和退出码可以在不启动子进程的情况下集中测试。

用户级子命令：

- `inspect -in project.aep`
- `profile -in project.aep`
- `diff -expected before.aep -actual after.aep [-ignore rules.json]`
- `migrate -in source.aep -target AE2025 -out migrated.aep`
- `capabilities`

成功与错误都使用 `schema_version: 1` JSON envelope。usage/未知命令退出 2，运行失败或语义 diff/blocked migration 退出 1，成功退出 0。

不要为新的普通用户 workflow 再增加独立 `cmd/*`。优先在 `internal/toolkitcli` 增加深层操作，并让旧命令继续作为兼容或高级维护入口。只有依赖完全不同、生命周期独立且不适合统一 JSON contract 的工具才考虑单独命令。
