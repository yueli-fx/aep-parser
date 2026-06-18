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
| expr-evolution(`time*N`) | ⬜待建 | Task 0.3 spike:AE 是否真求值 Go 写的 Evolution 表达式 |
| expr-wiggle(`wiggle(34,0.29)`) | ⬜待建 | Task 0.3 同 spike |
| many-mask(单层 ~21) | ⬜待建 | Task 0.4 spike:多 mask 不被 silent-drop |
| curves(CurvesCustom 曲线数据) | ⬜待建 | Task 0.5 spike:arbitrary-data 可读/可写性 |

### 12 comp(DAG 叶→根)

| # | comp | status | 备注 |
|---|---|---|---|
| ① | シェイイイイプ！！！ | ⬜待建 | 1 shape 层,4 动画矩形 |
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
