# ⚠ AE 接受自建 .aep 的多阶段 gate + canonical seed 策略

SUMMARY: AE 接受自建 .aep 的多阶段 gate + canonical seed 策略
READ WHEN: building .aep from scratch targeting AE 24+/25 acceptance; debugging "AfterFX crash" or "文件数据丢失" on builder output

---

Phase 6 ship gate 2026-05-22。从零构造 .aep 走 AE 的"接受"检查时，遇到 5 道梯度症状。每修一道 AE 给一个新症状信号，最终用 **单一 canonical seed (AE 2020)** 加 fps/duration 计算字段收敛。

## 症状梯度

| Stage | builder 缺什么 | AE 表现 | 信号 |
|---|---|---|---|
| 1 | cdta 二级 timing 字段全零 | **AfterFX 进程崩** | 崩溃 = 内部 sanity check 失败 |
| 2 | head[12..15] / head[16..19] counter < nextItemID | "文件数据丢失"（同 stage 3） | item ID upper bound |
| 3 | Item LIST 后面缺 8 个 Fold-level sibling chunks (FEE/fvdv/fiop/ftts/foac/fiac/fipc/fifl) | **"After Effects 错误: 文件数据丢失"** | item attribute cache 缺 |
| 4 | cdta @0x2C (masterTicks) 错值 | duration 显示错（10s → 5.005s）+ NTSC shutter ScriptingAPI 误算 | AE 用 @0x2C/tickRate 算 duration |
| 5 | Item 内字段跨 AE 版本不同 (AE 24+ Material 10 groups / ldta 160→164B / FEE ppSn) | AE 2020 "unknown exception" 打开 AE 25 出的文件 | per-version Item internals |

## 收敛策略: 单一 canonical seed (AE 2020)

**Escape hatch**: AE 高版本读低版本兼容 → builder 永远输出最小版本 (AE 2020)。一个 dummy_comp 模板 + per-fps timing table 即可跨 AE 2020/2022/2025。

不维护 per-version Item 模板。不在代码里写 "if target == AE2025"。

## Capability traits (替代 hardcoded 版本判定)

V3 方向先记下来 —— Item 内部跨 AE 的差异点：

| trait | AE 2020 (17.7) | AE 2022 (22.6) | AE 2025 |
|---|---|---|---|
| `ldta_size` | 160 B | 160 B | 164 B |
| `fee_has_ppSn` | false (0 children) | true (1 child) | true (1 child) |
| `material_lighting_groups` | 0 | 0 | 10 (Casts/Accepts Shadows, Light Transmission, Ambient/Diffuse/Specular/Shininess/Metal, Shadow Color, Accepts Lights) |
| `default_layer_tdgp_children` | 19 | 19 | 37 |

我们写 AE 2020 trait subset（全 false / minimum size / 0 material groups），AE 25 通过 back-compat 路径读取并补内部缺省。

## Stage 1 详: cdta 二级 timing 字段

cdta 总 204 字节，但 parser 历史只读 ~16 个关键字段（W/H/fps/duration/bg/shutter）。AE 写出的字节里大量"未 RE"slot 实际上是 timing engine 必读。

字段表 (fps-依赖部分):

| offset | 含义 | 30fps | 29.97fps | 50fps | 59.94fps |
|---|---|---|---|---|---|
| @0x06 (u16) | ticks_per_frame | 1024 | 800 | 512 | 400 |
| @0x08 (u32) | tickRate = tpf × fps_actual | 30720 | 23976 | 25600 | 23976 |
| @0x10 (u32) | time-base divisor (constant 600) | 600 | 600 | 600 | 600 |
| @0x18 (u32) | secondary divisor (600 for fresh comps; tickRate after user edit) | 600 | 600 | 600 | 600 |
| @0x2C (u32) | masterTicks = duration_s × nominalTickRate | 5s × 30720 | 10s × 24000 | … | … |
| @0x30 (u32) | tickRate mirror | = @0x08 | | | |
| @0xB8 (u32) | duration_frames mirror of @0xB0 | | | | |

**Two tick rates** for NTSC:
- `tickRate (actual)` = tpf × fps_actual (29.97 → 23976) — 存 @0x08 / @0x30
- `nominalTickRate` = tpf × fps_nominal_whole (29.97 → 24000) — 用作 @0x2C 的 divisor

整数 fps 下两者相等，NTSC 下不等。AE display duration = @0x2C / tickRate 算出来的秒数 —— 跟 @0xB0 / fps_actual 算的对得上是因为 nominalTickRate / tickRate ≈ 1。

**Fixture sources**:
- `test_data/fixtures/re_tickrate.aep` RE_fps_* — 每个 canonical fps 的 cdta 字节
- `test_data/fixtures/re_cdta_probe.aep` A_baseline (29.97, 10s) — masterTicks 公式实证
- `internal/serializer/templates/2020_dummy_comp.aep` — 默认 30fps 12s seed

## Stage 2 详: head counters

Root `head` chunk (20 B) `[12..15]` / `[16..19]` 两个 uint32 BE counter。`NewProject(...)` 从模板继承 `[1, 1]`。AE 期望 counter ≥ max item ID used。

修：`WriteAEP` 入口前 `syncHeadCounters()` 把两个 counter 拉到 `max(currentValue, nextItemID)`。

counter B 语义未完全 RE（AE2025_2comp=49 / dummy_comp=25 / re_batch=5229 —— 跟 save sequence 或总 property registration 数相关）。设为 nextItemID 够 AE 接受。

## Stage 3 详: Item Fold-level siblings

每个 Item LIST 在 Fold 里必须跟 8 个 sibling chunks（Fold 的直接 children，**不是** Item LIST 的 children）：

```
LIST Fold:
  fdta
  LIST Item       ← comp 1
  LIST FEE        ← sibling 1 (AE 2020: 0 children; AE 22/25: 1 child ppSn)
  fvdv (4 B)      ← sibling 2 (= 0x00000003)
  fiop (1 B)      ← sibling 3 (= 0x00)
  ftts (4 B)      ← sibling 4 (= 0x00000000)
  foac (1 B)      ← sibling 5 (= 0x00)
  fiac (1 B)      ← sibling 6 (= 0x00)
  fipc (2 B)      ← sibling 7 (= 0x0000)
  fifl (4 B)      ← sibling 8 (= 0x00000000)
  LIST Item       ← comp 2
  ... 重复 sibling 1-8
```

修：`ensureCompTemplate` 从 dummy_comp 提取 `siblingChunks`。`NewComposition` 在 rootFold append Item 后再 append 8 个 clone。

跨 comp 字节完全相同，无需 per-item 计算。

## Stage 4 详: masterTicks 公式

最初公式 `masterTicks = ticks_per_frame × 5 × fps_nominal_whole` 对 5-second fixture（re_tickrate）巧合正确，但跟 duration 无关。10s 的 A_baseline 数据让公式正确：

```
masterTicks = round(duration_seconds × nominalTickRate)
```

`@0x18` 也跟 fresh-comp vs user-edited 分两种状态:
- Fresh new comp: `@0x18 = 600` (= TimeBaseDivisor)
- User edited timing/shutter: AE 把 `@0x18` 改成 `tickRate`

Builder 出 fresh，写 600。

## Stage 5 详: per-version Item internals → escape hatch

我们最初的做法是给每个 target 一个 `<version>_dummy_comp.aep` 模板（embed 3 个，dispatch 选）。AE 2020 + AE 2025 ship gate 都过。

随后实证：AE 高版本能读 AE 2020 格式 items（back-compat）。所以可以 collapse 3 模板 → 1 (AE 2020 minimum)。同时 ship gate 仍过。

最终代码: 单一 `templates/2020_dummy_comp.aep`。Project.target 字段仍存在 —— 决定 **empty-project skeleton** (root chunks 如 svap/nhed)，不影响 Item items。

## 工具链

`tmp_debug/gen_dummy_comp.jsx` — 跨 AE 版本通用，args.json 指定输出路径。让 AE 自己存一个 1-comp project 用作 RE 源。运行:
```
"E:/adobe/Adobe After Effects <version>/Support Files/AfterFX.exe" -r .../gen_dummy_comp.jsx
```
JSX 末尾写 `.done` marker；Go 端轮询等。

## 教训

- **builder-from-scratch ≠ parser 反向**: parser 容错（缺字段当默认 0），AE 严格（缺字段不开 / crash / 显示错值）。
- **多信号区分根因**: 崩溃 = cdta timing 空；"数据丢失" = item siblings 缺 / head counter 错；duration 错值 = masterTicks 错；shutter 错值 = ScriptingAPI quirk (NTSC stored × 1.2)。
- **canonical seed 策略**: 一个最小合法模板比 N 个 per-version 维护成本低。AE back-compat 是天然 escape hatch。
- **ScriptingAPI ≠ stored value**: 至少 NTSC `shutterAngle` 经 ScriptingAPI 返回时被 × 1.2（stored 180 → API 216）。储存字节跟 AE-saved 文件完全一致；JSX 校验需绕开。

## 关联 commits / 文件

| 文件 | 角色 |
|---|---|
| `internal/codec/cdta_layout.go` | 7 个 offset 常量（Stage 1 + 4） |
| `internal/codec/framerate.go` | `fpsTiming` + `canonicalFpsTiming` + `lookupFpsTiming` |
| `internal/serializer/mutate_composition_new.go` | `compTmpl` (single seed) + builder logic |
| `internal/serializer/write.go` | `syncHeadCounters` (Stage 2) |
| `internal/serializer/templates/2020_dummy_comp.aep` | 唯一 canonical seed |
| `tmp_debug/gen_dummy_comp.jsx` | JSX 工具：让任意 AE 版本写一个 dummy_comp |
| `test_data/generators/verify_v2_1.jsx` | AE-side ship gate driver |
| `internal/aep/new_composition_test.go` | `runAEShipGate` + `TestV2_1_AEShipGate_AE2020/2025` |
