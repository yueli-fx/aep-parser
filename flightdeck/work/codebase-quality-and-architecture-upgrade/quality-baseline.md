# 质量基线 — 综合升级

## 依赖与并发

- `tools/debug/dag_boundary`：DAG OK，未发现低层反向依赖。
- race：`rifx`、`serializer`、`scene`、`aep_test` 全部通过。
- 公开 `Document` 明确不是并发安全对象；server 每个请求独立 parse，不共享 Document。
- Go toolchain 从 1.25.1 升至 1.25.12；`govulncheck v1.6.0 ./...` 从 18 个可达标准库漏洞降为 0。

## Coverage 快照

以下为 package-local statement coverage，不等同于最终能力证据，也不会统计所有从 `internal/aep_test` 穿过 scene/serializer 的黑盒路径：

| Package | Coverage |
|---|---:|
| rifx | 64.8% |
| serializer | 30.3% |
| scene | 3.1% |
| profile | 82.3% |
| recipe | 78.0% |
| aepmigrate | 67.9% |
| server | 83.8% |
| toolkitcli | 71.3% |

因此不把 scene 3.1% 直接解释为“97% 未测试”：大量行为通过外部 `aep_test`、fixture 和 AE gates 验证。后续改进应建立跨 package coverage 或能力证据映射，而不是追逐单一百分比。

## 性能基线

fixture：`re_solidnull.aep`，77,653 bytes；Windows amd64，Intel i7-9700K。三次结果受 Windows cache/扫描波动影响，当前只作为回归数量级：

| Benchmark | 时间 | 分配字节 | allocations |
|---|---:|---:|---:|
| Parse | 3.25–7.11 ms/op | ~373 KB/op | 9,993/op |
| ProfileJSON | 0.59–0.85 ms/op | ~115 KB/op | 955/op |

基准入口已进入根外部测试 package，后续性能优化必须同时报告时间、bytes/op 与 allocs/op。当前 77 KB 样本不存在阻塞性性能问题；批量大文件和峰值内存仍需单独容量样本，不从小 fixture 外推。

## Interface 结论

- 根 `Document` 是小而深的稳定 inspection/profile/write interface，继续隐藏 scene、serializer 和 chunk。
- 不可信输入通过一个聚合 `Limits` 值对象配置，避免为每项预算扩张方法参数。
- 大型 authoring facade 仍留在 `internal/aep`，不机械暴露到根 package；只有真实外部 workflow 才扩大根 interface。
- `PropertyStream.Keyframes` 的浅拷贝方案被拒绝：generic Value 可能含 slice，浅拷贝不能提供 immutability，且会增加 lowering 热路径分配。完整解决需要 immutable value 或受控遍历 interface，不在本轮用半措施替代。

## 交付证据

- recipe profile：151/151 pass，241 个 profile path，0 validate/compile/profile failure。
- migration：906 total / 900 pass / 0 blocked / 0 failed / 6 skipped；standalone verify profile diff=0；explicit matte AE2025 1/1。
- parser：RIFX fuzz 与带结构化 AEP seed 的 serializer fuzz 均无 panic。
- 全仓 test、vet、capindex 500 项、Windows amd64 / Darwin arm64 / Linux amd64 build 通过。

## 剩余风险

- PropertyStream mutable keyframe view 为 accepted P2，只存在 internal authoring surface；根 SDK 不暴露。
- package-local scene coverage 低，能力证据分散在 `internal/aep_test` 和 AE gates；后续应生成跨 package/能力证据视图，不以单一覆盖率阈值阻塞。
- 大文件/批量峰值内存尚无代表性语料 benchmark；当前小 fixture 仅建立回归入口，不能外推 1 GiB 上限容量。
