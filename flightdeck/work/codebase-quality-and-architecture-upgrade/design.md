# 设计 — 代码质量与架构综合升级

## 总目标

把 aep-parser 升级为可长期维护、可安全处理不可信工程、可在不破坏未知数据的前提下扩展、并且能够用证据说明能力边界的代码库。

升级成功不等于 lint 全绿或文件更整齐，而是：

- 错误输入不会造成失控资源消耗、panic 或越权访问；
- 已支持的读写行为满足明确不变量，不会静默损坏工程；
- 核心 module 通过小 interface 隐藏复杂实现，调用方无需理解 chunk/runtime 双重状态；
- 新字段、新版本和新产品 adapter 能在清晰 seam 上扩展；
- 测试证据等级与公开承诺匹配；
- 逆向知识、代码、fixture、capindex 和文档拥有明确真相源。

## 分支目标

### A. 正确性与写回不变量

审查 parse → scene → mutation → lower → write → reparse 链路，消除状态不同步、未知 chunk 丢失、失败后半写入和错误结果静默降级。

### B. 安全性与资源治理

覆盖输入大小、分配、递归、整数溢出、路径、并发、超时、表达式/外部依赖和批处理隔离。所有限制都应有 typed error 与可复现测试。

### C. Module 深度与可扩展性

识别浅 module、透传 facade 和泄漏实现细节的 interface。把复杂行为收进能产生 leverage 与 locality 的深 module；只有真实变化点才建立 seam 和 adapter。

### D. 测试与证据体系

区分单元、round-trip、profile、fixture、AE acceptance 和 render 证据。测试穿过 module interface 观察结果，不依赖可替换 implementation 细节。

### E. 性能与容量

建立典型和极端 AEP 的基准、峰值内存、分配热点、批量吞吐和重复完整字节副本台账，再根据证据决定流式或惰性解析的必要性。

### F. 知识与真相源整合

梳理代码、fixture、capindex、生成文档、registry 和 Flightdeck 的职责，消除需要人工同步的重复台账，让 durable finding 可路由、可验证、可更新。

### G. 工程规范与维护自动化

在行为风险收敛后处理错误风格、命名、注释、gofmt、lint、依赖治理和机械清理。规范门禁必须匹配当前基线并逐步收紧。

## Review 原则

- 风险优先于美观：P0/P1 先于格式和命名。
- interface 是测试面；避免为了测试暴露 implementation。
- 通过删除测试判断浅 module：删掉后若复杂度只散回调用方，才说明 module 有价值。
- 重构和行为修复分开提交；每个修复先建立可复现证据。
- 不把“Go 能 reparse”误写成“AE 能接受”，不把“AE 能打开”误写成“视觉语义正确”。
