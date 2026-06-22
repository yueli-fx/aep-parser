---
showcase: booyah-clone
direction: 全工程复刻 — 把整个 Booyah Glitch 真实工程(12 comp/61 层/~100 mask/wiggle+Evolution 表达式/Curves)用咱们的 Go API 从零重建,作为「理解金标准」检验。原工程仅当只读取值神谕,chunk 全由 API 重建。
capabilities: [full-project-replication, precomp-nesting, fractal-noise-params, displacement-map-layer-ref, glow, fill, gaussian-blur, venetian-blinds, exposure, gradient-ramp, mask-many, shape-rect-animated, trim-paths, text-animator, expression-evolution, expression-wiggle, curves]
gates: []
status: 待review
last_updated: 2026-06-22
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
| ① | シェイイイイプ！！！ | ✅**complete**（用户真机验收 2026-06-21）· AE 2025 render 像素级一致(0/0/32px)· ⚠AE 2020 render 待补 | 4 rect 精确 kf + fill 色逐值对账原工程 ✓(色彩:raw [A,R,G,B]0-255 ↔ SetColor)。**AE 2025 接受、shape 不 drop**。**render 三个真 bug 已修**:① 全空白=层 position 默认 (0,0) 非合成中心([[shape-layer-position-default-offscreen]],4f579d1);② t=1 残留杆=层时长默认满 comp,原层仅 [0,0.901](6091016);③ **中间帧错位=`deriveTickRate` 把 NTSC kf 时间读大 3×**(`×1000/scale` 伪修正,cdta @0x08 才是真 tickrate;AE valueAtTime/keyTime 实证 → d03101c,详 [[ntsc-tickrate-derive-3x-off]])。→ **clone vs orig 中间帧 t=0.3/0.5/0.8 像素 diff = 0/0/32px**(32px=sub-pixel 29.97↔30 帧snap;clone 用 30fps 因 NewComposition 分数 fps cdta 时基不一致,另案见同 incident)。⚠AE 2020 侧 render 待补。 |
| ② | テキスト変えるならココ！ | ✅**complete**(用户真机验收 2026-06-22:居中/斜体对、UI 可编辑) | 1 text 层 "GLITCH" 全建成(`gen_text_komako.go`,2add514+本提交)。三缺口**全闭**:①`SetLayerTransform` 物化层 Position(28kf)/Opacity(24kf);②`AddTextTrackingAnimator`+`AnimateTextTracking`(135→0);③`AddTextCharacterOffsetAnimator`+`AnimateTextCharacterOffset`(21→0)——后两者本会话 RE+双版本 render-gate ship([[text-animator-create-re]] 末 Case)。Go round-trip 逐值对账原工程 ✓(Anchor/Position/Opacity/2 动画器 kf 全等)。**AE2020≡AE2025 接受+render 一致**:t=1.0 渲出 "GLITCH"→"PURCLQ"(Character Offset +9 字母环移)+ Tracking 撑开 + Opacity flicker,效果对。**文字样式 + 居中已对齐原工程**(2 轮 review 反馈,RE 自原 text doc DOM):`SetRunFontIndex`(Industry-Demi)+`SetRunFontSize(110)`+`SetRunTracking(65)`+`SetRunFauxItalic(true)`+`SetParagraphJustification(center)`——斜体 + 字体/字号匹配。**竖直居中根因**:原工程 anchor Y=-436 = 其**文字框中心**(原 glyph 在 layer-space top=-502,baseline 被抬高~380px——一个我们 `SetText` 不复刻的 btdk first-baseline),我们的 from-scratch 是正常 point-text(box center≈-27),照抄 -436 会把文字压低~410px(sourceRectAtTime 实测定位)。修法:**按 clone 自己的 box center 设 anchor**(非照抄原值),原 Position kf 即把该中心摆正中 → 渲染 centroid Y=530(原 544,差在 glitch 抖动内)。**fidelity delta**(非阻塞):kf linear 近似原 ease(故 t=1.0 显 "PURCLQ" scramble 中,原已回 "GLITCH");box-center -27 取自替换字体(本机无 Industry),用户机 Industry 度量略异(~10px);字色默认(原色经父 ⑧ RGBズレ Fill 覆写);30fps(同 ①)。 |
| ③ | マップ用フラクタルノイズ | ✅**complete**(用户真机验收 2026-06-22)· 双版本 AE-accept + DOM 对账 PASS | `gen_fractal_map.go`:2 黑 solid 层各 1 Fractal Noise。Go round-trip 逐值对账原工程 ✓(NoiseType=1/Contrast=254/UniformScaling=0/ScaleW 211·2847/ScaleH=12/Complexity=1/Offset Turbulence 2kf/Evolution expr 全等,连 L1 `time*3000\r` 尾 CR 都从 oracle 取)。**AE2020≡AE2025 接受**:comp 2 层不 drop,Fractal Noise 32 props 全在,**Evolution `expr="time*1200"/"time*3000" on`(AE DOM 确认表达式真启用,非假绿)**,Offset Turbulence DOM=960,540(parser fraction[0.5,0.5]→AE 像素中心,映射对),blend Overlay/Normal 对。**结构 delta**(render-neutral,已记账):原 2 层共享 1 solid(footage 67);我们无 from-scratch 共享 footage-item API(仅 SetSource-retarget 会留孤儿 / DuplicateLayer 继承 effect 需脆弱二次 reopen)→ 用各自 1 黑 solid(Fractal Noise 自生成像素,黑 solid 像素无关 → 渲染等同)。**终帧 render-pixel 归 comp ⑫**。30fps(原生即 30,无 ① 的 tickrate fudge)。 |
| ④ | カクッ | ✅**complete**(用户真机验收 2026-06-22:居中对、描边一致)· 双版本 AE-accept + DOM | `gen_kakuh.go`:2 shape 层(「シェイプレイヤー 3」over「2」),各=**stroked rect(1635×810,无 fill,描边 0.8px)+ Trim Paths(Start=97/Offset=5.3)** + 层级 `SetLayerTransform` Scale(2kf 64%→100%)/Opacity(20kf flicker),L0 加 RotateZ=180。Go round-trip 逐值对账 ✓(rect/stroke-width/trim-start-offset/scale-kf/op-kf/rotZ 全等)。**AE2020≡AE2025 接受**:4 comp 全在、④ 2 层不 drop、2×Stroke+2×Trim 被 AE 吃进、dump OK。**两处关键修正**:①`SetLayerTransform.Scale()` 单位=**percent**(非 parser 的 fraction),scale 需 ×100(初版 0.64 写成 0.64%);②**层位置**:原工程 Position 是 separated dimensions,parser 读 Position_0/_1=0,0 但真机居中——照抄 0,0 把图层摆到左上角(用户真机反馈);修法=`SetLayerTransform.Position()` 显式居中 comp 中心(rect 在 shape-(0,0)、anchor (0,0),层居中即 rect 居中),render 复验 rect 居中(两对角 bracket)。详 [[shape-layer-position-default-offscreen]] 第二 Case(separated-position 读侧陷阱)。**结构/保真 delta**(render-neutral,记账):①from-scratch shape 多套一层 `Vector Group→Vectors Group`(同 comp ①,已验像素级一致);②描边色取 AddStroke 默认黑(原工程 elide 了描边色=AE 建时默认,不可恢复)——0.8px 细线;③Scale Z=1 vs 原 0.853(2D 层 Z 无关);④Position 写 unified(0,0) vs 原 separated(0,0),值同;⑤kf linear 近似原 ease。终帧 render-pixel 归 ⑫。30fps。 |
| ⑤ | プリコンポジション 1 | 🔶**待review**(双版本 AE-accept + DOM 对账 PASS) | `gen_precomp1.go`:**首遇 precomp 嵌套**——3 个 `NewPrecompLayer` 全引用 comp ①(staggered 三份 glitch)。L0 Position 2kf[778→1204]/L1 Opacity 13kf+居中/L2 静态居中;层 start 错峰 0.267/0.133/0.167 经 `SetStartTime`(post-reopen ldta edit;**坑**:precomp 是 embed-template clone,pre-reopen 的 scene `StartTime` 字段不生效,须 post-reopen SetStartTime)。**居中陷阱(用户真机反馈"右下角"→bisect 揪出库 bug)**:precomp 居中要 anchor=源中心,但 `SetLayerTransform` 对 **AV/precomp 层的 anchor 按「源尺寸的分数」编码**(0.5=中心),非 shape/text 的像素——写像素 960,540 被 AE 读成 960×1920/540×1080≈(1.84M,0.58M),pivot 飞出去→渲染全黑(anchor 0,0 时 0×N=0 不暴露,故前版渲到右下角)。**修法**:写分数 `anchor(0.5,0.5)`(AE DOM 实测读回 960,540),render 复验居中(t=0.6 三份 staggered glitch 居中,合原版布局)。详 [[setlayertransform-av-anchor-fraction]](库 bug,gen 暂走 workaround)。Go round-trip:3 层 srcID 全=①、pos/op/start 全等。**AE2020≡AE2025 接受**:5 comp、⑤ 3 层不 drop、DOM 三层 source 全绑 ①([[precomp-layer-source-id-re]] empty/wrong 失败面排除)+ **AE2025 render 居中眼验**。verify.jsx 加 `l.source.name` dump。终帧 render-pixel 归 ⑫。无新 API。 |
| ⑥ | シェイプの塊 | ✅**complete**(用户真机验收 2026-06-22) | `gen_shape_katamari.go`:**复用 ⑤ 的 `NewPrecompLayer` 建法**——7 层全引用 comp ⑤ プリコンポジション 1(一簇缩放/错峰的嵌套副本 = "shape 塊")。各层 Position 2kf(**绝对坐标 541→1340 横滑**,非 ⑤ 的 separated 0,0 陷阱,无需显式居中)+ Opacity 34/41/39/39/39/44 kf(L6 无);静态 Scale(L0 24%/L1-5 33%/L6 75%)+ Rotate Z(L0/L3/L4=90°);层 start 错峰 0/0.367/0.334/0.300/0.134/0/0 经 `SetStartTime`(post-reopen)。**带上 ⑤ 的教训**:每层复刻**全 transform 通道(含静态 Scale/Rotation)**,非只 Position/Opacity——否则副本飞散([[layer-replication-drops-static-transform-channels]])。anchor 全 elided→写源中心**分数(0.5,0.5)**(同 ⑤ 的 AV-anchor-fraction 库 bug workaround [[setlayertransform-av-anchor-fraction]])。Go 结构对账:7 层 srcID 全=⑤、Position/Opacity kf 数 + 静态 Scale/Rotation + start 全等(34/41/39/39/39/44 对上 plan)。**AE2020≡AE2025 接受**:6 comp、⑥ 7 层不 drop、DOM 7 层 source 全绑 ⑤、OK 无损坏。**render 眼验**(t=1.0 AE2025):渲出一簇缩小的 glitch shape 切片(青/灰横条+十字)居中偏右下、尺寸/透明度各异,无 blank/飞散/off-screen,簇偏右与 Position 横滑一致。无新 API。终帧 render-pixel 归 ⑫。 |
| ⑦ | ここは開けない方が身のため | 🔶**待review**(双版本 AE-accept + 结构对账 + orig-vs-clone render 匹配) | `gen_hiraku.go`:3 层全 src=⑥ シェイプの塊(复用 `NewPrecompLayer`)。**新维度 = Glow(`ADBE Glo2`) 挂 precomp 层**:L0 静态 Position[1127,592]+Scale 101%+Opacity 25kf(淡出 0)+Glow(Radius=0/Intensity=0.30);L1 静态 Position[822,526]+Scale 101%+Opacity 34kf+Glow(Intensity=0.35);**L2 隐形**(Opacity 静态 0、无 Position/Scale/Glow,为结构保真照建)。Glow Radius=0 是非默认(AE 默认 25),`SetEffectParam(-0003=Radius, -0004=Intensity)` 从 oracle 取。Position 是**静态绝对坐标**(非 ⑥ 的 2kf、非 ⑤ 的 separated 0,0),anchor 源中心分数 0.5。同 ⑥ 复刻全 transform 通道(含静态 Scale)。Go 结构对账:3 层 srcID 全=⑥、Position 静态值/Opacity kf 数(25/34/L2=0)/Glow Radius·Intensity/start 全等。**AE2020≡AE2025 接受**:⑦ 3 层不 drop、2×Glo2 入 DOM(Radius=0/Intensity 0.30·0.35 AE 读回对)、OK 无损坏。**render orig-vs-clone**(t=0.5):clone 与原工程同帧像素布局一致(同十字+横切片+同位置同淡度;⑦ 是中间 comp,孤立看本就淡,留待 ⑩/⑫ 合成叠加)。无新 API。终帧 render-pixel 归 ⑫。 |
| ⑧ | RGBズレ | ⬜待建 | 3 层 Fill 色差(需②) |
| ⑨ | なんか周りのやつ | ⬜待建 | 3 层 + Trim Paths(需④) |
| ⑩ | グリッチテキスト | ⬜待建 | 27 层/~100 mask/11 displacement ← 怪物(需⑦②⑧③) |
| ⑪ | 背景変えるならココ！ | ⬜待建 | 6 层 + wiggle + Ramp |
| ⑫ | メインコンプ！ | ⬜待建 | 3 层 + Curves ← 顶 + 终帧 render-pixel(需⑩⑨⑪) |

## 验证状态

未开建(scaffold 仅 Phase 0 Task 0.1)。showcase review-gate:用户真机验过 `booyah-clone.aep` 才可 `complete`;agent 眼验 ≠ 用户验收。
