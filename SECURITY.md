# Security Policy

## 当前支持范围

项目仍处于首次公开发布前阶段。安全修复优先覆盖：

- 当前 `mainline`；
- 统一 `aep` CLI 和根 Go package；
- `internal/rifx` 的畸形输入与资源耗尽问题；
- `aepserver` 的上传、路径输入和进程稳定性问题。

尚未发布的历史 commit、研究脚本和 AE 自动化 worker 不承诺独立安全维护周期。

构建与 release 必须使用 `go.mod` 声明的 Go patch 版本，或同一 release family 中更新且无已知可达漏洞的 patch。CI 固定运行 `govulncheck`、RIFX/serializer fuzz smoke、核心 race tests 和全仓测试；降低工具链版本属于安全敏感变更。

## 报告漏洞

仓库公开后，请使用 GitHub Private Vulnerability Reporting / Security Advisory 私下报告。不要在公开 issue 中附加恶意 `.aep`、宿主机路径、凭证或可直接利用的复现细节。

报告应包含：

- 受影响的 commit、命令或 package；
- 最小复现输入，或可私下获取该输入的方式；
- 预期与实际行为；
- 是否导致 panic、越界读取、资源耗尽、任意文件访问或代码执行；
- 已知的缓解方式。

维护者确认问题前，不承诺具体披露时限。确认后会先给出影响范围和临时缓解，再协调修复与公开披露。

## 部署边界

`aepserver` 已具备 parser 预算、请求体限制、并发限制、HTTP timeout、panic recovery 和 allowed path roots，但多租户公网部署仍应使用独立 worker、容器/操作系统资源限制与最小权限文件系统。进程内 recovery 不能替代内存和 CPU 隔离。
