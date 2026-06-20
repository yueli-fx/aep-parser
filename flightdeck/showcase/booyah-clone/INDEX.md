---
showcase: booyah-clone
direction: 全工程复刻 — 把整个 Booyah Glitch 真实工程(12 comp/61 层/~100 mask/wiggle+Evolution 表达式/Curves)用咱们的 Go API 从零重建,作为「理解金标准」检验。原工程仅当只读取值神谕,chunk 全由 API 重建。
capabilities: [full-project-replication, precomp-nesting, fractal-noise-params, displacement-map-layer-ref, glow, fill, gaussian-blur, venetian-blinds, exposure, gradient-ramp, mask-many, shape-rect-animated, trim-paths, text-animator, expression-evolution, expression-wiggle, curves]
gates: []
status: 待建
last_updated: 2026-06-19
regenerate: "go run ./flightdeck/showcase/booyah-clone  +  scripts/ae_run.ps1 render.jsx / verify.jsx"
---

# booyah-clone — Booyah Glitch 全工程复刻(理解金标准)

照 `specs/2026-06-19-booyah-glitch-full-replication.md` + `plans/2026-06-19-booyah-glitch-replication.md`。原工程 = `samples/motionbox/glitch/booyah-glitch/Booyah Glitch.aep`。

## 产物

| 文件 | 类型 | 说明 |
|---|---|---|
| `main.go` / `oracle.go` / `gen_<comp>.go` | 生成器(Go, tracked) | oracle 只读原工程;main 按 DAG 拼装;每 comp 一个 build 函数 → `booyah-clone.aep` |
| `render.jsx` / `verify.jsx` | AE 脚本(tracked) | render 终帧 png;verify dump 指定 comp DOM 值对账 |
| `booyah-clone.aep` / `*.png` / `*.json` | 产出 (gitignored) | |

## 覆盖账本(诚实标:✅复刻 / 🔧待RE / ⛔物理blocked / ⬜待建)

### 未知前沿(Phase 0 spike)

| 特征 | status | 备注 |
|---|---|---|
| expr-evolution(`time*N`) | ✅ | **Task 0.3 GO**:`SetExpression` stable+双版本 AE gated(`TestExprEffect` effect-param+`time*40`);time* idiom 已 gated(`TestExpression` time*90);`ADBE Fractal Noise-0023` Evolution=`time*1200` round-trip 实证干净。无需降级。 |
| expr-wiggle(`wiggle(34,0.29)`) | ✅ | **Task 0.3 GO**:wiggle idiom 已 gated(expr_vocab `wiggle(2,250)`);`ADBE Exposure2-0003` Exposure=`wiggle(34,0.29)` round-trip 干净。写入端字节路径与表达式内容无关(closed decision 2026-06-17)。 |
| many-mask(单层 ~21) | 🔶在位验 | 0.4 standalone spike **并入 ⑩ 在位验**(masks 集中在 ⑩,①-⑨ 仅背景 L4 有 1 个;Go round-trip 对 silent-drop 无意义=假绿,只能 AE 验,故 ⑩ 建时 AE-accept 实地验)。mask=parade list,非 keyframe lhd3 分页,Go 侧无 count 上限。 |
| curves(CurvesCustom 曲线数据) | ⛔曲线blocked / 实例可加 | **Task 0.2+0.5**:曲线数据读侧即 gap(只 `-0000` master)+ 无写 param = arbitrary-data 物理 blocked(无 scripting 通道)。但 effect **实例可加**(`ADBE CurvesCustom` 原生在库)→ ⑫ 降级:加默认 Curves 实例、曲线标 ⛔,实例 AE-accept 在 ⑫ 验。 |

> **Task 0.2 读侧审计(2026-06-19,probe 已删)**:mask 顶点 ✅(グリッチテキスト L0 = 16 mask,`Vertices`/`Closed`/`Mode`/`PathKeyframes` 可读)· keyframe 值 ✅(`Time`/`Value`)· effect params + expression ✅。读侧唯一 gap = Curves 曲线数据(见上行)。gen 文件直接用 raw API(`l.Masks`/`l.Effects`/`pr.Keyframes`)即可,oracle 暂不需额外 typed accessor(YAGNI)。

### 12 comp(DAG 叶→根)

| # | comp | status | 备注 |
|---|---|---|---|
| ① | シェイイイイプ！！！ | ✅AE-accept · 🔶render残留 | 4 rect 精确 kf + fill 色逐值对账原工程 ✓(色彩:raw [A,R,G,B]0-255 ↔ SetColor [R,G,B,A]0-1)。**AE 2025 接受、shape 不 silent-drop**(verify.jsx 实证 4 rect+fill)。**render**:曾全空白(红线4:值对渲染空)——根因=**层 position 默认 (0,0) 非合成中心**,rect 负坐标全落画外([[shape-layer-position-default-offscreen]],fix 4f579d1:gen 经 `Transform().Position()` 拷原层 position)→ **t=0 帧与原工程字节完全一致、t=0.5 差 4B**。🔶残留:>4kf 的 rect#3(11)/#4(8)在 t=0.5 位置略偏、t=1 有原工程没有的残留杆(疑 lhd3 关键帧分页 [[lhd3-keyframe-capacity-pages]])。⚠结构 delta(rect 包进 Vector Group/Vectors Group,原 flat)render-neutral 已证。 |
| ② | テキスト変えるならココ！ | ⬜待建 | 1 text 层 "GLITCH" + tracking/char-offset 动画器 |
| ③ | マップ用フラクタルノイズ | ⬜待建 | 2 Fractal Noise 层 + Evolution expr(依赖 0.3) |
| ④ | カクッ | ⬜待建 | 2 shape 层 |
| ⑤ | プリコンポジション 1 | ⬜待建 | 3 层(需①) |
| ⑥ | シェイプの塊 | ⬜待建 | 7 层(需⑤) |
| ⑦ | ここは開けない方が身のため | ⬜待建 | 3 层 + Glow(需⑥) |
| ⑧ | RGBズレ | ⬜待建 | 3 层 Fill 色差(需②) |
| ⑨ | なんか周りのやつ | ⬜待建 | 3 层 + Trim Paths(需④) |
| ⑩ | グリッチテキスト | ⬜待建 | 27 层/~100 mask/11 displacement ← 怪物(需⑦②⑧③) |
| ⑪ | 背景変えるならココ！ | ⬜待建 | 6 层 + wiggle + Ramp |
| ⑫ | メインコンプ！ | ⬜待建 | 3 层 + Curves ← 顶 + 终帧 render-pixel(需⑩⑨⑪) |

## 验证状态

未开建(scaffold 仅 Phase 0 Task 0.1)。showcase review-gate:用户真机验过 `booyah-clone.aep` 才可 `complete`;agent 眼验 ≠ 用户验收。
