# ⚠ SetTimeRemapEnabled false-green: AE needs 2 identity keyframes, not a static value

SUMMARY: SetTimeRemapEnabled false-green: AE needs 2 identity keyframes, not a static value
READ WHEN: 实现/调试 Layer.SetTimeRemapEnabled;启用时间重映射 Go 不报错但 AE 读 timeRemapEnabled=false;以为设静态值 0.0 够;TimeRemapEnabled getter 与 SetTimeRemapEnabled 不自洽

---

## Signature
- symptom: `SetTimeRemapEnabled(true) 不报错,但 AE DOM layer.timeRemapEnabled=false`
- error_type: —  (false-green / 值不被 AE 认,且 getter/setter 表示法不一致)
- where: Layer.SetTimeRemapEnabled (scene_layer_accessors.go) → TimeRemap().SetStaticValue(0.0);probe = TestLayerFlags2_AEShipGate_*(TRM 层,已从 gate 移除)
- trigger: precomp 层(有时长源)SetTimeRemapEnabled(true) 期望 AE 显示启用

## 症状/复现

补验 arc layer-flags2 批想把 SetTimeRemapEnabled 一起升。载体 = MAIN 内 precomp 层(源 SRC 4s,CanSetTimeRemapEnabled=true)。SetTimeRemapEnabled(true) Go 不报错,但 AE2025 读 `layer.timeRemapEnabled=false`(同批 SetEffectsEnabled/SetIsNull 都过)。

## 根因

**AE 用「Time Remap 属性带 ≥1(默认 2 个 identity)关键帧」表示启用时间重映射**,不是一个静态值。本 setter `enable` 分支只 `TimeRemap().SetStaticValue(0.0)` —— 设了静态值、**0 关键帧**,AE 不当启用。

**且本库自身 getter/setter 不自洽**:`TimeRemapEnabled()` 返回 `len(p.Keyframes) > 0`,而 `SetTimeRemapEnabled(true)` 设的是静态值(0 关键帧)→ 设完自查也是 false。doc 注释里 RE 实锤反而印证了正解:"re_wave2_ae24.aep RE_TIMEREMAP:tr_baseline 0 关键帧;tr_remap_on **2** 关键帧"——AE 启用态 = 2 关键帧。

## 修法(未做,留待需求驱动)

正确 SetTimeRemapEnabled(true) 须**合成 2 个 identity 关键帧**(t=inPoint 值=源起始时间、t=outPoint 值=源结束时间,即 value==time 的恒等映射),而非裸静态值;可能还需 ldta enable flag(待 RE 确认是否单靠关键帧即可)。这是属性合成级改动(类同 transform-group default-omission 的 materialize),不是单值写。

当前处理:
- SetTimeRemapEnabled **降级 stable→alpha**,boundary 标 false-green,**不升 ae-accept**(留 roundtrip)。
- 不纳入 layer-flags2 gate(SetEffectsEnabled + SetIsNull 已双版本 ae-accept)。
- 升级路径:enable 改为合成 2 identity 关键帧 + 对齐 getter → precomp 载体 gate 验 AE timeRemapEnabled=true。

## Cases
- 2026-06-17 首次。补验 arc layer-flags2 批,SetTimeRemapEnabled(true) AE 读回 false;同批 SetEffectsEnabled/SetIsNull 正常升 ae-accept。同族 false-green:[[layer-setstretch-ae-recomputes-span]](单写值 AE 不认)· [[comp-setframerate-no-duration-rescale]]。
