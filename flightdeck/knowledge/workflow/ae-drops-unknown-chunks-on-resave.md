# ⚠ AE drops unknown RIFX chunks on resave (custom metadata stash)

SUMMARY: AE drops unknown RIFX chunks on resave (custom metadata stash)
READ WHEN: 想往 .aep 嵌入自定义数据/作者元数据/版权/任意 stash;纠结自定义 RIFX chunk 会不会被 AE 保留;评估 lib-write→lib-read vs AE-resave 的数据存活;选元数据通道(comment/marker/XMP/自定义chunk)

---

## Signature
- symptom: `自定义 RIFX chunk 注入 .aep,AE 能正常打开,但 AE 保存(resave)后该 chunk 消失`
- error_type: —  (negative-finding / 数据丢失,非异常)
- where: AE 保存路径(AE 从自身 DOM 重序列化整个工程);probe = TestCustomChunkProbe_* (internal/aep/custom_chunk_probe_shipgate_test.go)
- trigger: 往工程里塞 AE schema 之外的任意 chunk,期望它跨 AE 保存存活

## 症状/复现

想给 .aep 嵌作者名/版权/工具水印等自定义元数据,问"AE 会不会拦截"。实测(2026-06-17,双版本 + 两位置)结论:**AE 打开 OK,resave 后丢**。

probe `TestCustomChunkProbe_*`:库造最小工程(NewProject+NewComposition+NewSolidLayer)→ `rifx.Parse` 自身输出 → 注入自定义 leaf chunk `uMD1`(data=作者串)→ `Chunk.Write` 出文件 → AE open+save → `rifx.Parse` resaved 找该 chunk。

| 注入位置 | AE 版本 | AE 打开 | resave 后 |
|---|---|---|---|
| RIFX root(Fold 同级) | AE2025 | ✅ 正常(numItems/comp 都对) | ❌ DROPPED |
| RIFX root | AE2020 | ✅ | ❌ DROPPED |
| Comp Item LIST 内(idta/cdta/Layr 同级) | AE2025 | ✅ | ❌ DROPPED |

(Item×AE2020 未单跑;机制版本无关,root 双版本已坐实 → 同样 DROPPED 高置信。)

## 根因

**AE 保存 = 从它自己的内存对象模型(DOM)整片重序列化工程**,不是 byte-patch。任何不在 AE schema/DOM 里的 chunk——无论塞在 root 还是它建模的容器里——重写时根本不会被写回。**这跟本库的 opaque-preservation 是两码事**:本库的 parser 保留未解 chunk(硬约束 #5),是**本库的** round-trip 保真;AE 没有这个义务。

**关键:AE 打开未知 chunk 不报错、不损坏**(没触发"项目文件似乎已损坏")——它按 RIFX size 跳过。所以:
- **lib-write → lib-read 链路**(AE 只当查看器、从不保存):自定义 chunk **存活**(磁盘文件 AE 不动,只有保存才重写)。任意数据都行,opaque-preservation 保证库读回。
- **要 AE 保存后还在**:自定义 chunk **不行**。

**AE 唯一会保留的"未知数据"是它显式建模为可扩展的槽**:① 缺失插件的 effect 数据(无该插件的机器打开+保存仍保住 effect 字节);② 已知 property descriptor 内 AE 不认识的新版属性字节(见 [[btdk-point-value-needs-formatpsreal]] 邻域 / 批7c 洞察:AE2020 opaque 保住 AE24+ enum)。这些是 AE 框架内的"剩余字节"保留,**不是任意顶层 chunk**。

## 修法

要嵌元数据,按"是否需 AE 保存后存活"分流:

- **只库写库读(AE 不重存)** → 怎么塞都行,含自定义 chunk(`rifx.Parse` → 改 `Chunk.Children` → `Chunk.Write`)。AE 打开不损坏。
- **要 AE 保存后存活、可接受显示为注释** → 用 AE 原生 comment(`Layer.SetComment` / `Marker.SetComment`+Chapter/URL 已双版本 ae-accept gated;`Composition.SetComment` / `Footage.SetComment` 是 item-level idta,目前只 roundtrip 验,from-scratch 可能撞 layer 当年的 ldta/idta has-comment flag 坑 → 用 Layer/Marker comment 最稳)。
- **要正经元数据(作者/版权,AE File Info 面板可编,跨 Adobe 通用)** → XMP(`dc:creator` 是标准作者字段)。**本库目前不支持 XMP**(capindex 无匹配),要做须 RE「AE 把工程 XMP 存哪个 chunk」+ 双版本 ship-gate。属未来 feature。

probe 保留为 negative-finding gate(`AE_SHIP_GATE=1 go test -run TestCustomChunkProbe_*`),日后想验别的注入位置(Layr 内 / Fold 内 / 已知 chunk Data 尾部追加字节)直接改 `inject*` 复跑。

## Cases
- 2026-06-17 首次。用户问"能否在 aep 嵌作者名,会被 AE 拦截么"→ probe 实测 root(双版本)+ Item(AE2025)全 DROPPED;AE 打开均正常。结论:AE resave 从 DOM 重建,任意 chunk 不保;元数据走 native comment / XMP(未支持)。
