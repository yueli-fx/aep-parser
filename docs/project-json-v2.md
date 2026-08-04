# ProjectJSON v2

`Document.ProjectJSON()` 返回与解析器内部对象分离的快照。v2 保留 v1 的 `flat_index`，并以 `property_ref` 关联属性树、兼容的旧扁平属性以及顶层 `property_records`；`property_ref` 不是文件偏移或写回定位符。opaque 叶也有对应 record，不产生悬空引用。

每个树节点只保存结构事实：`property_type`、`elided`、`name_source`、兄弟顺序、重复序号、规范路径、`origin_evidence`，以及叶节点的 `property_ref`。值语义集中保存在对应的 `property_records` record 中，避免在大型工程的树和 record 之间重复整套值、关键帧与状态：

```json
{
  "property_type": "PROPERTY",
  "property_ref": "property-123",
  "match_name": "ADBE Position",
  "name": "ADBE Position",
  "name_source": "fallback",
  "property_value_type": "TwoD_SPATIAL",
  "dimensions": 2,
  "elided": null,
  "elided_status": "unknown",
  "can_vary_over_time": true,
  "is_spatial": true,
  "decode_status": "decoded",
  "temporal_ease_status": "not-present",
  "raw_preserved": true,
  "write_capability": "supported"
}
```

`decode_status` 使用 `decoded`、`partially-decoded`、`unsupported` 或 `failed`。`write_capability` 使用 `supported`、`preserve-only`、`unsupported` 或 `unknown`；细分操作能力位于 `write_capabilities`。这些字段描述解析器事实，不代表产品层是否应允许编辑或浏览器是否能够预览。

当前 AEP 解析模型没有可靠的 `elided` 原始位，因此输出 `null` 和 `elided_status=unknown`，不会把未知冒充成 `false`。写回能力由实际 writer backing（cdat、tdb4、tdbs 或关键帧流）验证；不能证明完整前置条件的 separation 保持 `unknown`。

`origin_evidence` 记录原始命名条目在父组中的序号、由该顺序生成的规范路径及保留状态。它是快照内的原始定位证据，不暴露 RIFX chunk、文件偏移或 writer backref，也不充当业务 ID。无法证明 named/indexed 语义的新组输出 `property_type=UNKNOWN`，不会因未命中 indexed 目录就默认伪装成 `NAMED_GROUP`。

当 Parser 能证明属性来自原始文档、具有稳定规范路径且保留了源顺序时，property record 会额外包含 `write_target`：

```json
{
  "document_ref": "sha256:...",
  "composition_id": 12,
  "layer_id": 34,
  "canonical_path": "/0:ADBE Transform Group#0/1:ADBE Position#0"
}
```

`write_target` 只能交回创建该快照的 `Document.Export`，其中 `document_ref` 将目标绑定到确切的源 AEP。调用方仍需检查 record 的 `write_capabilities`，Export 还会在应用前再次验证 writer backing。Export v1 是全有或全无的数值静态属性变更：它在私有文档副本上修改、写出、重新 Parse 并验证目标值，同时对 writer 声明范围以外的完整 RIFX 树做字节摘要。成功报告 `byte-exact-outside-claimed-ranges` 才表示未知 chunk、opaque tail、节点顺序和所有未声明字节均被保留。

`name_source=decoded` 表示 AEP 中存在独立实例名；`fallback` 表示解析器用稳定 `match_name` 填充；`missing` 表示两者都没有。Parser 不做本地化翻译。

关键帧的 Temporal Ease influence 单位为 `ratio`。每侧、关键帧聚合及 property record 都有验证状态；异常值或因损坏布局无法完整解码的 ease 标为 `invalid-preserved`，同时把 property 的 `decode_status` 降为 `partially-decoded`，不会静默截断或假报成功。JSON 无法表示的 `NaN`/无穷值输出为 `null`，原始类别写入对应的 `speed_raw` 或 `influence_raw`。

顶层 `property_integrity` 对原始命名条目、已保留树条目、属性 record、关联、重复关联和 opaque/unknown 保留计数。未关联属性、缺失属性树和未保留节点进入 `property_diagnostics`；`dropped_unknown_count` 来自 observed/preserved 实际差值，不是固定常量。
