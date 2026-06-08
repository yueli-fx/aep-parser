# specs/ — INDEX

<!-- AUTO:specs -->
### 待启动（idea）
- [2026-05-29-path-embed-re-findings.md](2026-05-29-path-embed-re-findings.md) — idea — Path embed bytes — RE findings（V2.2.1 子项②预备，已 ship，留作 Stroke 参考）
- [2026-06-01-renderer-write-re-findings.md](2026-06-01-renderer-write-re-findings.md) — idea — Composition.Renderer W — RE findings（prin 104B / prda 变长，结构性，待实现）
- [deferred-backlog.md](deferred-backlog.md) — idea — Deferred backlog 细项（从退役 logbook §Deferred 迁入）

### 进行中·完成（active·done）
- [2026-06-07-v3-m8-physical-split-design.md](2026-06-07-v3-m8-physical-split-design.md) — active — V3 M8 方案② 真·物理分包设计（A 先行 + B′ back-ref 接口）：scene/serializer/codec 物理拆包，back-ref 作 scene 内 writer 接口（serializer 实现）→ 保留全部方法 API（不破 API）、scene 编译期零 rifx；eager length-preserving patch 经接口；opaque 延后到 C；全程保 byte-exact 回归门
- [2026-05-27-v3-deep-think.md](2026-05-27-v3-deep-think.md) — active — V3 deep think：open questions / risk register / migration strategy
- [2026-05-26-py-aep-parity-design.md](2026-05-26-py-aep-parity-design.md) — active — py-aep parity API 全覆盖路线图（P1/P2 已落；P3 §3A RQ R/W+结构性 + DimensionsSeparated R/W 双向(static+animated) + 3C PropertyBase Remove/MoveTo/Duplicate + 3G comp marker 增删 已落，剩 ValueText）
- [2026-05-22-v3-direction.md](2026-05-22-v3-direction.md) — active — V3 direction：scene-graph IR + capability matrix + serializer split（Phase 1-5 + 包重组方案① + M8 scene→rifx 白名单清零 已落；真·物理分包(方案②接口倒置)执行中 P2 6/10）
<!-- /AUTO -->
