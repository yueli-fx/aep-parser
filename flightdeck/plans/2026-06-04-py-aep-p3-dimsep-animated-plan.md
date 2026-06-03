---
status: active
implements: specs/2026-05-26-py-aep-parity-design.md
summary: P3 DimensionsSeparated animated 子方向 — animated Position 的 keyframe 流拆分/合并（RE-first：先 byte-diff AE animated before/after，再实现 stream split/merge + ship-gate）
---

# P3 §3C — DimensionsSeparated **animated** 子方向

**Spec**: [`../specs/2026-05-26-py-aep-parity-design.md`](../specs/2026-05-26-py-aep-parity-design.md) §3C
**前置**: static Position separate/merge 双向已 ship（`mutate_property_separate.go`，AE 2020+2025 6/6 PASS，`incidents/separate-dimensions-write-mechanics.md`）。**当前 animated 一律 refuse**（leader `StaticValue` 非 `[]float64` / follower 非 scalar 即报错）。
**性质**: 结构性 + keyframe 流迁移 → V2.1 atomic + **AE 2020+2025 双版本 ship-gate**（CLAUDE.md #6）。RE-first。

## 0. 核心未知（gating RE）

static 是值迁移；animated 是 **keyframe 流迁移**：
- **separate**：leader 的一条 3D **spatial** Position keyframe 流（motion-path，spatial in/out tangent，bpk-128，`valueLayout{dim:3,headerByte:0x07,spatial:true}`）→ 拆成 3 条 1D **temporal** 标量流（Position_0/1/2，bpk-48，non-spatial，temporal influence/speed tangent）。
- **merge**：3 条 1D 标量流 → 1 条 3D spatial 流（若 X/Y/Z keyframe 时间不对齐，AE 如何处理？取并集？插值？—— 未知）。
- **tangent 映射**是最大未知：spatial 切线（2D/3D 方向向量）↔ 每轴 temporal 切线（influence%/speed）之间 AE 的换算规则。
- keyframe 字节布局必走 `layoutFor` 分发（`incidents/keyframe-byte-layout-dispatcher.md`）。

## 0.1 RE findings — cut 1（2026-06-04，AE 2020 `re_separate_dims_anim.jsx`）

fixtures：`re_sepdim_anim_{before,after}.aep`（gitignored test_data，本地）。leader 3 keyframe [100,200,50]/[300,400,150]/[500,100,250] @ t=0/1/2，默认 ease（influence 16.667%）。

**结构（我方 parser 读两态，与 static 同构，仅 stream 类型互换）**：
- **separated**：leader `ADBE Position` static=**[960,540,0]（property 默认）**，kf 流清空 — 同 static separate（leader 不再权威，AE 读 followers）。Position_0/1/2 **animated**（StaticValue=nil = 有 kf 流），Position_2 已合成。
- **merged**：leader **animated**（kf 流）；Position_0/1 static=0 预分配占位；无 Position_2。

**per-axis 语义（AE scripting 读回）**：leader 3D spatial kf → 3 条 1D temporal kf，每轴 keyframe = (value=该轴分量, in/out influence, in/out **speed**)。speed = 该轴空间速度分量（Position_1 Y 速度有正负 +3333/-1666/-5000，对应 200→400→100 反向）；influence 端点=0、内部≈1%（默认转换值）。observed speeds（units/sec）：
```
Pos0(X): out k1=+3333.3  in/out k2=+6666.7  in k3=+3333.3
Pos1(Y): out k1=+3333.3  in/out k2=-1666.7  in k3=-5000.0
Pos2(Z): out k1=+1666.7  in/out k2=+3333.3  in k3=+1666.7
```

**∴ separate animated = ①解 leader 3D spatial kf 流 → 每轴 (t,v,in/out speed,influence) ②leader 重置 static 默认（复用 static 路径）③每轴 1D temporal kf 流写 followers（static 占位 → animated stream 结构性转换）④合成 Pos2。** merge = 反向重组。

**剩余硬 RE（cut 2）**：(a) leader 3D spatial kf 流的 ldat byte 解码（已有 `lowerTransformVec2Spatial` 编码端，需对称解码取每 kf 的 spatial in/out tangent）；(b) spatial tangent → per-axis temporal speed/influence 的换算（observed speeds 是否能从 spatial bezile 切线纯几何推出，还是 AE 另存？需 byte-diff leader kf ldat vs follower kf ldat）；(c) per-axis 1D temporal kf 的 ldat 编码（speed/influence 落 keyframe block 哪几个 offset）。**先 byte-diff，勿臆测（path-keyframe 三处误读教训）。**

## 1. RE 计划（Phase 0，先做）

- `test_data/re_separate_dims_anim.jsx`（改自 `re_separate_dims.jsx`）：建 3D 层，Position 打 **2-3 个 keyframe**（不同 time + 不同 3D 值 + 至少一个非 linear easing），save before → `pos.dimensionsSeparated=true` → save after。AE 2020 生成。
- byte-diff before/after（`tmp_debug/list_item_chunks` / 自写 dump 工具）：
  1. leader 分离后 keyframe 流变成什么？（清空回默认？保留？AE 读 followers）
  2. Position_0/1/2 的 keyframe 流：time 表、每 kf 的 value、tangent 字节 —— 与 leader 各轴分量的对应关系。
  3. tangent：取一个已知 easing 的 leader kf，看拆出的 per-axis kf tangent 字节，反推映射。
- 产出：在本 plan 追加「§0 RE findings」+ 升级/新建 incident。**先有 byte 真相，再写代码**（bisection-over-stacking 教训）。

## 2. 实现（Phase 1，RE 后细化）

- 复用现有 keyframe 编解码：`lowerTransformVec2Spatial`（3D spatial）/ `lowerTransformScalar`（1D）+ `valueLayout` + `layoutFor` + parse_keyframe 解码。
- `separatePosition` 扩 animated 分支：解 leader kf 流 → 按 RE 规则生成 3 条 1D kf 流写入 followers（含合成 Position_2）。
- `mergePosition` 扩 animated 分支：解 3 follower kf 流 → 合成 1 条 3D spatial 流写 leader。time 不对齐策略按 RE 定。
- atomic：沿用 static 的 validate-then-commit（fallible 的 re-parse/合成前置）。

## 3. 测试（Phase 2）

1. Go round-trip：自建 animated fixture → separate → WriteAEP → re-parse → 断言 followers kf 流（time/value/tangent）正确 + leader 状态；merge 反向。
2. 新 `TestDimSepAnimated_AEShipGate_{AE2020,AE2025}`：separate/merge animated → AE 读回 keyframe 值/时间一致 + resave 往返。
3. `go vet ./... && go test ./...` 全绿 + PASS 对账。

## 4. 切片顺序

1. **Phase 0 RE**：生成 animated before/after fixture + byte-diff + findings 文档化（incident）。**当前 / 下一步**。
2. separate animated 实现 + Go round-trip 测试。
3. merge animated 实现 + 测试。
4. AE 双版本 ship-gate → 升 stable。

## 5. 风险

- tangent 映射可能 lossy / 不可逆（spatial↔temporal 非双射）→ 若 AE 自身有损，我们对齐 AE 行为即可（ship-gate 为准），不追求数学可逆。
- merge 的 time-misalignment 语义可能复杂 → 首切片可限定「followers keyframe time 对齐」场景，misaligned 暂 refuse。
- 多分量 spatial motion-path 的 bezier 切线编码本就是项目最难布局之一 → 严格 byte-diff，勿肉眼臆测（path-keyframe RE 三处误读教训）。
