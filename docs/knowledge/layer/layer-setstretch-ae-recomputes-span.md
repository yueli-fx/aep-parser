# ⚠ SetStretch false-green: AE recomputes stretch from in/out span

SetStretch false-green: AE recomputes stretch from in/out span

## Signature
- symptom: `Layer.SetStretch(0.5) → Go round-trip 读回 0.5,但 AE 打开后 layer.stretch=100`
- error_type: —  (false-green / 值不被 AE 认,非异常)
- where: Layer.SetStretch (scene_layer_writers.go) + back.SetStretch(@ldta 0x08 分子 / 0x6C 分母);probe = TestLayerBool_AEShipGate_*(STR 层已从 gate 移除)
- trigger: 给层设时间拉伸只写 stretch 分子/分母,不调 in/out span

## 症状/复现

补验 arc layer-bool 批起初想把 SetStretch 一起升 ae-accept。实测(AE2025,2 种载体):

| 载体 | SetStretch(0.5) → AE layer.stretch |
|---|---|
| from-scratch shape 层(无源) | 100(reset) |
| precomp 层(SRC 子合成,有 duration) | **100(仍 reset)** |

→ 换有 duration 的源**也没用**,排除"载体无源"假说。SetStretch **单写 @0x08/0x6C Go round-trip 绿,但 AE 一律读回 100** = 红线1 假绿。

## 根因(强假说,未完全 RE 实锤)

AE 的 time-stretch **不是只读 stretch 分数**——它按 **layer 实际 in/out span ÷ 源 span** 重算 stretch。本 setter 只改 stretch 字节、**不动 outPoint/span**,于是 AE 打开时发现"stretch 说 50% 但 span 还是 100%",判 stretch=100% 覆盖我们的值。

(另一可能:@0x08/0x6C 不是 AE 真读的 stretch 源字段。但本库 parser 从真 AE 文件 RE 出这俩偏移、StretchFrac 能读回真值 → 偏移大概率对,问题在"AE 写时 stretch 与 span 一致,我们破坏了一致性"。)

## 修法(未做,留待需求驱动)

正确 SetStretch 须**协调写**:改 stretch 的同时把 outPoint 调成 `inPoint + 原 span × ratio`(让 span 反映拉伸),AE 才认。即 SetStretch 应是**复合操作**(stretch + 自动 outPoint),不是单字节写。

当前处理:
- SetStretch **降级 stable→alpha**,boundary 标"确认 false-green",**不升 ae-accept**(留 roundtrip)。
- 不纳入 layer-bool gate(AutoOrient + CollapseTransform 已双版本 ae-accept)。
- 升级路径:实现协调 outPoint 的复合 SetStretch → 建 precomp 载体 gate 验 AE layer.stretch=拉伸值。需求驱动再做。

## Cases
- 2026-06-17 首次。补验 arc layer-bool 批,SetStretch(0.5) 双载体均 AE reset 100。AutoOrient/CollapseTransform 同批正常升 ae-accept;SetStretch 隔离为 alpha false-green。同族教训:[comp-setframerate-no-duration-rescale](../composition/comp-setframerate-no-duration-rescale.md)(时间字段交互)· [btdk-point-value-needs-formatpsreal](../text/btdk-point-value-needs-formatpsreal.md)(单写值假绿)。
