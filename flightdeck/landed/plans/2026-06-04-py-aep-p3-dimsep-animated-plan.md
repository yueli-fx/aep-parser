---
status: done
implements: specs/2026-05-26-py-aep-parity-design.md
last_updated: 2026-06-04
summary: P3 DimensionsSeparated animated 子方向 — animated Position 的 keyframe 流拆分/合并（RE-first：先 byte-diff AE animated before/after，再实现 stream split/merge + ship-gate）
---

# P3 §3C — DimensionsSeparated **animated** 子方向

**Spec**: [`../specs/2026-05-26-py-aep-parity-design.md`](../specs/2026-05-26-py-aep-parity-design.md) §3C
**前置**: static Position separate/merge 双向已 ship（`mutate_property_separate.go`，AE 2020+2025 6/6 PASS，`incidents/separate-dimensions-write-mechanics.md`）。**当前 animated 一律 refuse**（leader `StaticValue` 非 `[]float64` / follower 非 scalar 即报错）。
**性质**: 结构性 + keyframe 流迁移 → V2.1 atomic + **AE 2020+2025 双版本 ship-gate**（CLAUDE.md #6）。RE-first。

## 接力锚点（下次会话 Phase 4 ship-gate 起点）

**状态**：Phase 0 RE + **Phase 1 separate + Phase 3 merge animated 实现全部完成**（2026-06-04，`separatePositionAnimated`/`mergePositionAnimated`，3D 首切片，round-trip 双向过，byte-exact vs AE 自存 fixture）。**下一步 = Phase 4 AE 双版本 ship-gate**：
1. 写 animated-readback verify JSX（模仿 `verify_separate_dims.jsx`，读 leader/follower keyframe 的 time/value/speed/influence 回吐 PASS/FAIL + resave）。
2. 加 `TestDimSepAnimated_AEShipGate_{AE2020,AE2025}`（separate + merge 两态 × 双版本，仿 `property_separate_shipgate_test.go`，gated by AE_SHIP_GATE）。
3. 第二 fixture（不同时距/值/4 kf）验 100/0.01 + 弧长常数非 fixture-specific。
4. 全过 → 升 stable，更新 cockpit/coverage。
separate 输出 byte-identical AE、merge 值/切线 byte-exact（仅 @0x08/@0x10 缓存字段差，AE 不校验），ship-gate 预期稳过。

**起手步骤**：
1. **先 byte-check after fixture 的 follower animated tdbs 结构**——用 `tmp_debug`（已有 `dump_sepdim_kf` dump 高层字段；如需原始 chunk 树用 `tmp_debug/list_item_chunks`）看 `re_sepdim_anim_after.aep` 的 `ADBE Position_0` tdbs：static 占位（`{tdsb,tdsn,tdb4,cdat,tdum,tduM}`）→ animated 后变成什么（cdat 换成 `kfl{lhd3,ldat}`？tdb4 @0x05 flag 值？）。**这是唯一剩的结构未知**，impl 前必看。
2. 在 `mutate_property_separate.go` 的 `separatePosition` 加 animated 分支（`p.StaticValue` 非 `[]float64` 即走 animated；leader `p.Keyframes` 非空判定）。复用 leader 的已解码 `p.Keyframes`（含 `InSpatialTangent`/`OutSpatialTangent`/`Value`）。
3. 每轴构造 1D temporal kf 流，套 §0.2 映射（`out_speed=outSpatTan[axis]×100`，`in_speed=−inSpatTan[axis]×100`，`inf=0.01`/边界 0，bezier）。编码复用 `lowerTransformScalar` + `valueLayout{dim:1,headerByte:0x00}`（bpk=48，offset 见 `parse_keyframe.go` non-spatial：value@0x08 / in spd@0x10 inf@0x18 / out spd@0x20 inf@0x28）。
4. 3 个 static follower → animated（替 cdat 为 kfl 流 + tdb4 flag，按步骤 1 的 byte-check）；leader 重置 static 默认（**复用现有 static separate 的 tdsb/cdat 改动**）；合成 Pos2（leader 同样，但其 kf 流拆 Z 轴）。
5. atomic：沿用 static 的 validate-then-commit（fallible 的 stream 构造/re-parse 前置）。
6. **测试先行**：自建/用 `re_sepdim_anim_before.aep` → `SetDimensionsSeparated(true)` → WriteAEP → re-parse → 断言 3 follower 各 3 kf 的 time/value/speed/influence == after fixture（§0.2 数值）+ leader static 默认。
7. 然后 merge 分支 + 双版本 ship-gate（**第二 fixture 不同时距/值验 100/0.01 常数泛化**）。

**首切片限定**：leader path temporal ease ≈ linear（默认）；leader 自身有非默认 temporal ease 时暂 refuse（§0.2 未决①）。

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

## 0.2 RE findings — cut 2（2026-06-04，`tmp_debug/dump_sepdim_kf` 解码 before/after kf）

我方 parser 已全解码两布局（`parse_keyframe.go`）：leader **spatial-style** bpk=128（path 单 ease @0x18-0x30 + 3D value @0x38 + 3D spatial in-tan @0x50 + out-tan @0x68）；follower **non-spatial 1D** bpk=48（value @0x08 + in speed/inf @0x10/0x18 + out speed/inf @0x20/0x28）。

**leader（merged）实测**：in/out interp = linear，path ease = {spd 0, inf 0}，但 **spatial tangent 非零**（auto-bezier）。每 kf 的 `outSpatTan[axis] = (P_next−P_prev)/6`（内部）、端点 /6 的边段；in-tan = −out-tan（对称 auto-bezier）。

**映射（leader spatial tangent → follower per-axis temporal，9/9 kf-side 全对上含符号）**：
```
follower.value[axis]    = leader.value[axis]                 // 直接拷
follower.out_speed      = leader.outSpatTan[axis] × 100
follower.in_speed       = −leader.inSpatTan[axis]  × 100
follower.influence      = 0.01（有相邻段的 side）, 0（首 kf in-side / 末 kf out-side）
follower.in/out interp  = bezier
```
常数 `100 = 1/0.01`，即 `spatialTan = speed × influence`。leader 的 spatial tangent 是 AE 存盘的现成值，**直接读不重算**。

**merge 反向**：`outSpatTan[axis] = follower.out_speed/100`，`inSpatTan[axis] = −follower.in_speed/100`，value 合并 3 轴；leader interp 设 linear + path ease 0。

**实现含义**：
- separate animated = 解 leader kf 流（已有 `parse_keyframe` 给全字段）→ 对每轴每 kf 套上式生成 1D temporal kf → 把 3 个 static 占位 follower **转 animated stream**（建 tdbs→kfl→{lhd3,ldat(bpk=48)}，复用 `lowerTransformScalar` 编码 + `layoutFor`）→ leader 重置 static 默认（同 static 路径 tdsb/cdat）→ 合成 Pos2。
- **未决小项**（impl 时 byte-check）：① path ease 非默认（leader 自身有 temporal ease）时映射是否仍成立——首切片限定 leader path-ease≈linear，非默认暂 refuse；② 100/0.01 常数是否随 keyframe 时距变——**ship-gate 用不同时距/值的第二 fixture 验证**（防 fixture-specific）；③ follower 转 animated 后 tdb4 flag（@0x05 等）取值需对 after fixture 的 follower tdbs byte-check。

**∴ cut-2 完成：核心 tangent 映射破解，实现路径清晰，剩 byte-level 编码细节在 impl 时对 after fixture 逐字段校验。**

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

1. ✅ **Phase 0 RE**（cut-1 结构同构 + cut-2 tangent 映射破解，见 §0.1/§0.2；incident 已升级）。`b41359f` + 本次。
2. ✅ **separate animated 实现** + Go round-trip 测试（2026-06-04）。`separatePositionAnimated`（3D linear-path-ease 首切片）：解 leader kf 流 → per-axis 映射 → 3 follower static→animated stream（直接建 kfl，bpk=48 hdr07=0x08）→ leader 重置 static 默认（72B cdat=[def, kf0 in/out spatTan]）→ 合成 Pos2。`TestSetDimensionsSeparated_Animated` 通过；separate 后 4 个 Position* tdbs **byte-identical AE after fixture**（`tmp_debug/verify_sep_anim`）。字节细节升级进 incident。
3. ✅ **merge animated 实现 + 测试**（2026-06-04，`mergePositionAnimated`）。`mergePosition` 加 animated follower 检测 → 解 3 follower 1D temporal kf → 合成 1 条 3D spatial leader 流（`outSpatTan=out_speed/100`、`inSpatTan=−in_speed/100`），leader static→animated flag 反向，followers 删除。`TestSetDimensionsSeparated_Merge_Animated` 通过；leader block value+in/out tangent **byte-exact AE 自存 merge fixture**（`tmp_debug/verify_merge_anim`），仅 @0x08/@0x10 缓存字段（AE 不校验）+ 末 kf outTan 不同。AE merge ground-truth fixture `re_sepdim_anim_merge_after.aep` 由 `re_sepdim_anim_merge.jsx` 生成（AE2020 self-serve）。
4. ✅ **AE 双版本 ship-gate 8/8 PASS**（2026-06-04，AE2020+2025 × separate/merge × uniform/非均匀）。`property_separate_anim_shipgate_test.go` + `verify_sepdim_anim.jsx`。
5. ✅ **cut-3 修正**（第二 fixture `re_sepdim_anim2` 非均匀 0.5/1.0/1.5s 抓出 fixture-specific bug）：speed = **值中心差分 ×(100/6)**（时距无关，非 leader 存盘 spatTan×100），influence = **0.01/segDur**（非常数 0.01）。time/value/influence byte-identical AE，speed ≤1 ULP（AE tick 量化）。第三 fixture `re_sepdim_anim3` uniform 0.5s 确认 speed 时距无关。`aeSepDimSpeedFactor` 常数 = `0x4030aaaaaaac192b`。**→ 可升 stable。**

## 5. 风险

- tangent 映射可能 lossy / 不可逆（spatial↔temporal 非双射）→ 若 AE 自身有损，我们对齐 AE 行为即可（ship-gate 为准），不追求数学可逆。
- merge 的 time-misalignment 语义可能复杂 → 首切片可限定「followers keyframe time 对齐」场景，misaligned 暂 refuse。
- 多分量 spatial motion-path 的 bezier 切线编码本就是项目最难布局之一 → 严格 byte-diff，勿肉眼臆测（path-keyframe RE 三处误读教训）。
