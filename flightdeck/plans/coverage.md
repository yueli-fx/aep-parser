---
status: active
implements: archive/specs/2026-05-26-py-aep-parity-design.md
summary: capindex 不覆盖的残值(暂搁/不可达/negative findings/可探方向);写/做能力真相源已迁 capindex
last_updated: 2026-06-16
---

# Coverage 残值（写/做能力已迁 capindex）

> **写/做能力真相源 = capindex**（源码内 `aep:cap` tag，CI 强制写面零漏标）：
> 查询 `go run ./cmd/capindex -q "<词>"` 或 grep `docs/capabilities.json` / 读 `docs/capabilities.md`。
> · API 签名 + read accessor 清单 → docgen `docs/*.md`
> · byte 级机制 / RE findings → `incidents/`
> · AE-attribute → Go-field 交叉矩阵 → [coverage-detail.md](coverage-detail.md)
>
> 本文件原 `## ✅ 已 ship` 写能力快照（347 行，2026-05~06）已**退役**——内容被 capindex（写能力状态）+ docgen（read/API）+ incidents（机制）吸收且更准、CI 防漂。git history 留存原文。
> 这里只保留 capindex **不覆盖**的曲面:**暂搁 / 不可达 / negative findings / 可探方向**（CLAUDE.md 文档地图指向本文件即为此残值）。

兼容声明:AE 2020 读下限 + AE 24+ 字段渐进写。length-preserving 写约束详 CLAUDE.md #1。

---

## 🗑️ 暂搁（fixture / 环境阻塞，遇到需要再做）

| 项 | 阻塞原因 |
| --- | --- |
| `AVLayer.environmentLayer` | 需 equirectangular 360° 视频素材 |
| `TextDocument.ligature` | 默认字体 ligature=false 设值无 diff；需带 OT `liga` feature 的字体 fixture |
| ~~`maskFeatherFalloff`~~ | ✅ **SHIPPED 2026-06-17**（mkif `@0x03`，旧「mkif 字节零变化」结论错——看漏 @0x03）；双版本 AE gate |
| 4D 颜色 32bpc 范围 (0..1 vs 0..255) | 需 32bpc 项目 fixture 验证 |
| mkif 残余字节 `@0x10 / @0x18 / @0x20-0x27` | 未 RE；roundtrip 走 `Mask.MkifRaw` 保留原字节，不假设 48 字节全已知 |
| **Property metadata (P2b 2D)** | 需 pard chunk reader + specs.py schema table 端口，复杂 RE 工作 |
| **ImportPlaceholder (P2a 2E)** | opti chunk format 未 RE；user fixture (`re_placeholder_ae20.aep`) 显示生成的 opti tag AE 不识别。2026-05-27 删除合成 builder，等真 fixture 出来后再起 |
| **CMS chunk 创建（P2b 2B）** | 文件原本没 CMS chunk 时 setter 拒写。CMS Utf8 在 root 下的精确位置 / 容器（裸 Utf8 vs LIST 包装）未 RE；空挂 AE 可能拒文件 |
| **`DisplayColorSpace`（P2b 2B）** | py-aep 提示 separate chunk，位置未 RE；旧 stub 永远 "None"，2026-05-27 删除 |

## ❌ 不可达（length-preserving 写约束之外 / AE 限制）

| 项 | 原因 |
| --- | --- |
| Layer / Effect / Mask vertex / ShapePrimitive **增删** | 结构性，破坏多个父 LIST 大小 |
| `Property.timeRemapEnabled` toggle | AE 加/删 2 个 identity keyframe（结构性） |
| `lineOrientation` 横/竖排切换 | layer-local 坐标重排 + 多字段连锁 |
| `MaskPropertyGroup.rotoBezier` | 切换重写整个 shape 顶点表示（+16 字节，4500+ byte-diff） |
| Motion Graphics Template / EP 模板 binding（除 `alternateSource`） | 跨 chunk 复杂结构，P3 罕用 |
| Project 渲染设置（`gpuAccel / colorSpace / expressionEngine`） | P3，AE 24+ 大多锁定为 default。（`Composition.renderer` R/W 已 ship — `SetRenderer` 双版本 ship-gate 绿） |
| Adobe World-Ready composer 切换 | P3 |
| 手动 kerning **首次启用** | 结构性添加（需 AE 先 emit `/8` slot） |
| Camera `FilmSize` setter | ldta `@0x98` 持久化但 ScriptingAPI 不暴露写路径（详 `incidents/camera-filmsize-ldta-write-blocked.md`） |

## ⚠ Negative findings（runtime-only / AE ScriptingAPI 限制）

| 项 | 结论 |
| --- | --- |
| `CompItem.dropFrame` | AE 不持久化（脚本可设可读，字节零变化）—— FrameRate NTSC 自动推断 |
| `TextDocument.fontLocation` | AE 不持久化（runtime 从系统字体注册表 join 路径） |
| Variable fonts axes **写** | `TextDocument.fontVariation` 不存在；唯一写途径 = 切已加载的 named-instance |
| `composerEngine` / `everyLineComposer` | AE 24+ 只接受 UNIVERSAL；其它写入抛错 |
| `setAlternateSource(item)` AE 脚本行为 | 自动包 wrapper precomp；我们 setter 不包装 |
| AE 24/25 ldta 加长 | 实测仍 164 字节（与 AE 23+ 一致），未见第三种长度 |
| cdta tail (≥ 0xCC) | 不存在 —— cdta 总长就是 0xCC=204 |
| `Property.valueText` / `propertyParameters`（§3H） | 通用不可达：内置枚举 label 不在文件（pard 只存 nbOptions 计数），需 Adobe 不公开的 per-effect×版本 schema DB；AE 26.0-only API；py-aep 自己也没做。仅自定义 Dropdown Menu Control 的 label 在 `pdnm` chunk 可 RE（有界子集，当前不做）。详 `incidents/valuetext-needs-schema-db.md` |

---

## ❓ 剩余可探方向（非 candidate 列表，需要新发现才动）

可达字段约 99% 已 ship。继续动需要：

1. **ldta `@0x60-0x82` / `@0x8C-0x9F` 零值区 probe** — 高密度 JSX layer-flag 探针，可能挖出 1-2 个零散 flag 或全 negative。
3. **Footage proxy 字段** — 大部分结构性。
4. **Project nhed/nnhd 扩展字段** — 除 BitsPerChannel 外的字节，可能持 ColorSpace / Working Color Profile。
