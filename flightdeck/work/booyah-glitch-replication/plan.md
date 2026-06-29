# Booyah Glitch 复刻 — 实现计划

> **For agentic workers:** 本 plan 用 checkbox(`- [ ]`)跟踪。验证模型 = 本项目实际工作流(`go vet/build` 绿 + AE 双版本接受 + `verify.jsx` DOM 对账 + 终帧 render-pixel),**非** 通用 pytest TDD。每完成一个 comp/spike 即 commit(直推 main,永不 push)。

**Goal:** 用咱们的 Go API 从零重建整个 Booyah Glitch(12 comp),结构保真 + 终帧渲染对照原工程。

**Architecture:** `flightdeck/showcase/booyah-clone/` 一个 `package main`:`oracle.go` 用 `aep.Open` 把原工程当只读取值源;`gen_<comp>.go` 每 comp 一个 build 函数,经 New*/AddEffect/Set*/AddMask/Animate* 重建;`main.go` 按 DAG 拓扑序拼装 → 写 `booyah-clone.aep`。原工程只读、绝不 copy 字节。

**Tech Stack:** Go(internal/aep facade)· AE 2020+2025 ship-gate(`scripts/ae_run.ps1`)· `render.jsx`/`verify.jsx`(ExtendScript)。

**真相源:** 每 comp 结构/值取自 `data/reference/booyah-glitch/booyah_dissect.txt`(完整画像)+ 执行时 `aep.Open` 实读。本 plan 内联的是**已知标量参数 + API 调用骨架**;mask 路径 / 关键帧值这类批量数据由 `oracle.go` 程序化提取,不内联。

---

## Progress

current: **Phase 3 怪物 comp ⑩ グリッチテキスト ✅ COMPLETE（🔶待 review）**:27 层骨架(04a5e66)+ 131 bbox-rect mask(875fdd6)+ 27 effect 链(88076b1),三 Step 全双版本 AE gated + Go 对账 + render 眼验。副产库 fix `setLdtaFrac` 粗 divisor 截断(2886c5e)。下一 = **Phase 4 ⑪ 背景変えるならココ + ⑫ メインコンプ（顶层 + 终帧 render-pixel）**。Phase 1+2 已全完成（详下）。**Phase 1 叶子 comp 全完成**:1.1 ① ✅ + 1.2 ② ✅(用户验收) + 1.3 ③ ✅(用户验收) + 1.4 ④ ✅(🔶待 review)。**Phase 2 进行中**:Task 2.1 ⑤ プリコンポジション 1 ✅(用户真机验收,首遇 `NewPrecompLayer` 嵌套) + **Task 2.2 ⑥ シェイプの塊 ✅(🔶待 review)**:`gen_shape_katamari.go` 7 层全 src=⑤,复用 ⑤ 建法 + 复刻全 transform 通道(Position 2kf 绝对坐标 + Opacity 34/41/39/39/39/44 kf + 静态 Scale 24/33/75% + RotateZ 90°),4 关全过(结构对账 + AE2020≡AE2025 7 层不 drop + render 眼验 glitch 切片簇)。⑥⑦ 用户真机验收 complete。**Task 2.4 ⑧ RGBズレ ✅(🔶待 review)**:`gen_rgbzure.go` 3 层 src=② + 各层 `AddEffect("ADBE Fill")` 染单一 RGB 通道(B/默认红/G)+ start 错峰 = chromatic aberration;4 关全过含 render 眼验(红绿蓝三份文字重叠)。坑:SetEffectParam 颜色要 []float64 非 [4]float64;L1 默认红靠模板(AE DOM 确认)。下一 = Task 2.5 ⑨ なんか周りのやつ(L0/L1 src=④ 无动画 + L2 shape 层带 Trim Paths 动画)。**Task 1.4 ④ カクッ**:`gen_kakuh.go` 2 shape 层=stroked rect(1635×810)+Trim Paths(97/5.3)+层级 Scale(2kf)/Opacity(20kf)+L0 RotateZ=180,双版本 AE-accept(2×Stroke+2×Trim 入 DOM)。坑:`SetLayerTransform.Scale` 单位 percent 需 ×100;shape 几何是参数化(Rect+Stroke+Trim)非 freeform path。**Task 1.3 ③ マップ用フラクタルノイズ**:`gen_fractal_map.go` 2 黑 solid 各 1 Fractal Noise,**双版本 AE 接受 + DOM 对账 PASS**(Evolution `time*1200/time*3000` 表达式真启用、Offset Turbulence 2kf→DOM 960,540、UniformScaling off、blend Overlay/Normal、2 层不 drop)。footage-share 实测=各自 solid(无共享 API,render-neutral)。值全 oracle 取。verify.jsx 加 effect-parade dump。Task 1.2:两文字动画器 gap #2 Tracking + #3 Character Offset 本会话 RE + 双版本 render-gate ship(commit 2add514),`gen_text_komako.go` 拼成 comp ②（text + `SetLayerTransform` 28kf Position/24kf Opacity + 2 动画器），AE2020≡AE2025 接受+render 一致（GLITCH→PURCLQ 字符环移 + tracking 撑开 + opacity flicker）。✅ 基建(clear_ae_crashstate 67023b6;ae_run PostMessage 根治 foreground-lock b411d09)✅ AE 2025 接受+shape 不 drop ✅ render 三真 bug 全修:层 position→中心(4f579d1)·层时长→[0,0.901](6091016)·**中间帧错位根因=`deriveTickRate` 把 NTSC kf 时间读大 3×**(`×1000/scale` 伪修正;cdta @0x08 才是真 tickrate;AE valueAtTime/keyTime 实证根因,**非** kf 值/插值/分页——前一会话猜错方向 → d03101c;详 incident ntsc-tickrate-derive-3x-off)→ **clone vs orig 中间帧 t=0.3/0.5/0.8 像素 diff=0/0/32px**。残留旁支:AE 2020 侧 render 待补;clone 用 30fps 绕开 NewComposition 分数 fps cdta 时基 bug(另案)。

**Phase 0（前置 de-risk）全收口:**
- ✅ 0.1 scaffold（04e57eb）：`showcase/booyah-clone/` 包 + oracle 只读神谕 + 覆盖账本。
- ✅ 0.2 读侧审计（c3bff18）：mask 顶点/keyframe/params/expr 全可读；Curves 曲线数据 = 读侧 gap。
- ✅ 0.3 表达式 = GO（29b4db6，头号 de-risk 解除）：`SetExpression` stable+双版本 AE gated；Evolution=`time*N`（`ADBE Fractal Noise-0023`）+ `wiggle(34,0.29)`（`ADBE Exposure2-0003`）round-trip 干净。噪声动画+wiggle 可忠实复刻、无需降级。
- 🔶 0.4 many-mask(21)：并入 ⑩ 在位 AE 验（masks 集中在 ⑩；Go round-trip 对 silent-drop 假绿）。
- ⛔ 0.5 Curves：曲线数据物理 blocked（arbitrary-data 无 scripting），实例可加 → ⑫ 降级 + 在位验。

**Phase 1 进行中:**
- ✅ Task 1.1 comp ① シェイイイイプ（e4d266c Go 建 + d03101c tickrate 根因修）：4 rect 精确 kf + fill 色逐值对账 ✓；**AE 2025 render 与原工程像素级一致**（中间帧 0/0/32px）。结构 delta（shape 组嵌套 + transform 默认物化，render-neutral）已记账。残留旁支：AE 2020 render 待补；clone 30fps 绕开 NewComposition 分数 fps cdta bug。
- ⚠ 首次 AE 验证撞 crash-state cascade，受阻（详 `incidents/ae-automation-occlusion-crashstate.md` Case 2c）：run1 verify.jsx 用 JSON.stringify 在 catch 外抛 → 0 字节 done → ae_run 假 PASS → force-kill → 置崩溃标志；run2 safe-mode 框被前台游戏 foreground-lock 挡住关框 → 超时。已修 verify.jsx（纯字符串、末尾一次写）。注：游戏窗口≠用户在用（operator 在另一台机器），不问用户让机器。
- ✅ Task 1.2 ② テキスト（text+animator，用户真机验收 complete）。
- ✅ Task 1.3 ③ マップ用フラクタルノイズ（2 solid + Fractal Noise + Evolution expr，双版本 AE-accept + DOM PASS，complete 用户真机验收）。
- ✅ Task 1.4 ④ カクッ（2 shape 层=stroked rect+Trim Paths+层级 Scale/Opacity+L0 RotateZ，双版本 AE-accept，🔶待 review）。
- → **Phase 1 叶子层全完成**；进 Phase 2 中层 comp（⑤⑥⑦⑧⑨）。

---

## Phase 0 — 前置 spike + 基建(先把未知打掉,再建)

de-risk 原则:表达式/many-mask/curves 三个未知一旦 blocked,会改变后续 comp 的建法,必须最前置。

### Task 0.1 — Scaffold booyah-clone 包

**Files:**
- Create: `flightdeck/showcase/booyah-clone/main.go`(package main,空 `func main()` 先写出 `booyah-clone.aep` 一个空 project)
- Create: `flightdeck/showcase/booyah-clone/oracle.go`(`openOriginal()` 用 `aep.Open` 打开 `data/samples/.../Booyah Glitch.aep`,返回 `*aep.Project`)
- Create: `flightdeck/showcase/booyah-clone/render.jsx`(照 `showcase/glitch/render.jsx` 改 outPath;支持渲染指定 comp 名)
- Create: `flightdeck/showcase/booyah-clone/verify.jsx`(打开 .aep,dump 指定 comp 的 DOM 值为 JSON 到 stdout/文件)
- Create: `flightdeck/showcase/booyah-clone/INDEX.md`(frontmatter `showcase: booyah-clone` / `status: 待建` + **覆盖账本表**:每 comp/每未知特征 一行 status=待建)

- [ ] **Step 1:** 写 `oracle.go` 的 `openOriginal()` + 一个 `dumpCompNames()`,`main.go` 调它打印 12 个 comp 名,确认能读原工程。
- [ ] **Step 2:** `go vet ./flightdeck/showcase/booyah-clone/ && go run ./flightdeck/showcase/booyah-clone`,Expected: 打印 12 comp 名、exit 0。
- [ ] **Step 3:** `INDEX.md` 建覆盖账本骨架(12 comp 行 + 未知前沿 4 行:expr-evolution / expr-wiggle / many-mask-21 / curves)。
- [ ] **Step 4:** Commit `chore(showcase/booyah-clone): scaffold replication package + coverage ledger`。

### Task 0.2 — 取值 oracle(读侧完整性审计)

**目的:** 复刻靠「读原工程的值喂写器」。先确认 reader getter 能 surface 所有需要的值;读不出的(如 Curves)= 读侧 RE 任务,登记账本。

**Files:** Modify `oracle.go`(加 typed 访问:layer 列表/类型/blend/transform、effect matchName+param 值、mask path bezier、keyframe times+values、表达式文本)。

- [ ] **Step 1:** 对 `グリッチテキスト` L0 跑访问器,打印它的 effects+params+mask 数+keyframe;比对 `data/reference/booyah-glitch/booyah_dissect.txt` L0 值一致。
- [ ] **Step 2:** 逐项确认可读:mask bezier 顶点(`Mask` getter)· keyframe 值与缓动 · effect param 标量/颜色/点 · 表达式文本。**读不出的登记 INDEX 账本 `待 RE(read)`**(预期:Curves 曲线数据)。
- [ ] **Step 3:** `go vet` 绿。Commit `feat(showcase/booyah-clone): value-oracle over original (read-side)`。

### Task 0.3 — SPIKE:表达式(Evolution=time*N + wiggle)〔gates ③⑩⑪⑫〕

**为什么先做:** 表达式落在叶子 comp ③(Fractal Noise Evolution)+ ⑪(wiggle)。`SetExpression` 是已知 false-green 高危(`incidents/expression-enable-byte-pair.md`)。必须先实证 AE 是否真求值。

**Files:** Create `tmp_debug/` 或现有 ship-gate test 里加最小用例(NewSolidLayer + AddEffect Fractal Noise + SetExpression on Evolution = `"time*1200"`;另一例 Exposure + `"wiggle(34,0.29)"`)。

- [ ] **Step 1:** 写最小 Go:一层 Fractal Noise,`SetExpression(Evolution, "time*1200")` + `SetExpressionEnabled(true)`,WriteAEP。
- [ ] **Step 2:** AE ship-gate(`scripts/ae_run.ps1`,AE2025 先):打开 + verify.jsx 读 `expressionEnabled` 与逐帧 Evolution 值是否随 time 变化(render 两帧采样或 DOM valueAtTime)。Expected:**判定 PASS / FAIL**。
- [ ] **Step 3(go/no-go):**
  - PASS → 表达式可复刻,账本 `expr-evolution`/`expr-wiggle` = 可做;后续 comp 直接 SetExpression。
  - FAIL(AE 不求值)→ **降级路**(用户已批准):Evolution 用线性 keyframe 近似(0→time*N*dur over 6s),wiggle 用其 9kf 实测值直接复刻;账本标 `blocked(expr)` + 实证理由;若 FAIL 写新 incident(或 append `expression-enable-byte-pair` 一个 Case)。
- [ ] **Step 4:** 双版本(AE2020 + 2025)跑确认结论一致。Commit spike 结论(test + 账本 + 可能的 incident)。

### Task 0.4 — SPIKE:单层 ~21 mask〔gates ⑩〕

**Files:** 最小 Go:一层 + `AddMask` ×21(各一个简单 rect bezier path)+ WriteAEP。

- [ ] **Step 1:** 建 21 mask 层,WriteAEP,Go round-trip 读回 mask 数 == 21。
- [ ] **Step 2:** AE 双版本 ship-gate:打开不报损坏 + verify.jsx 读 `layer.mask.numProperties == 21`(防 silent-drop)。Expected:PASS。
- [ ] **Step 3:** 若 FAIL(分页/silent-drop)→ RE + 新 incident;若 PASS → 账本 `many-mask-21` = 可做。Commit。

### Task 0.5 — SPIKE:Curves 曲线数据〔gates ⑫〕

**Files:** 最小 Go:一层 + `AddEffect("ADBE CurvesCustom")`;探 `SetEffectParam` 能否物化曲线数据(arbitrary-data)。

- [ ] **Step 1:** 读原工程 メインコンプ L2 的 Curves effect,看 reader 能否 surface 曲线点(预期不能 → 读侧 gap)。
- [ ] **Step 2:** 探写侧:Curves 默认实例 AE 接受否(只 AddEffect 不设曲线)。若曲线数据不可读不可写 → 账本 `curves` = `blocked` + 实证理由(arbitrary-data,无 scripting 通道);⑫ 的 Curves 降级为「加默认 Curves 效果实例,曲线标 blocked」。
- [ ] **Step 3:** Commit spike 结论 + 账本。

---

## Phase 1 — 叶子 comp(DAG 第一层)

每 comp 统一**验收四关**(下文不再重复,简称「验收」):①`go vet`+`go build ./...` 绿 ②`go run` 生成 .aep ③ **Go 结构对账**(开原+clone,断言该 comp 层数/类型/blend/effect matchName+params/mask 数/keyframe 数一致)④ AE 双版本接受 + `verify.jsx` DOM 确认 AE 真吃进(防 false-green)。终层另加 render-pixel。

### Task 1.1 — ① シェイイイイプ！！！(id=173,1 shape 层,4 动画矩形)

**结构:** 1 个 shape 层 "シェイプレイヤー 1",含 **4 个矩形**,每个 `ADBE Vector Rect Size` + `ADBE Vector Rect Position` 带关键帧(Size 3/3/11/8 kf,Position 2/2/3/2 kf,linear)。act=[0→0.9]。

**Files:** Create `gen_shape_iiip.go`(`func buildShapeIiip(p *aep.Project, orc *oracle) *aep.Composition`)。

- [ ] **Step 1:** `NewComposition(p,"シェイイイイプ！！！",1920,1080,30,6)` → `NewShapeLayer`。
- [ ] **Step 2:** 加 4 个 rect(`RectNode`/对应 shape primitive API),每个 `AnimateScalarKeyframes`/对应 size+position 关键帧 —— **值由 oracle 从原工程 L0 提取**(rect size/pos 关键帧 times+values)。
- [ ] **Step 3:** 验收(无 render-pixel,叶子非终层;但单独 render 一帧 Read png 自查形状对)。
- [ ] **Step 4:** 账本 ① = 待review。Commit `feat(showcase/booyah-clone): comp① シェイイイイプ (4 animated rects)`。

### Task 1.2 — ② テキスト変えるならココ！(id=27,1 text 层 "GLITCH")

**结构:** text 层 "GLITCH",动画:`ADBE Text Tracking Amount`(2kf ease)·`ADBE Text Character Offset`(2kf ease)·Position(28kf)·Opacity(24kf)。

**风险:** 文字动画器(tracking/char-offset)= text-animator territory(`incidents/text-animator-create-re.md`)。NewTextLayer 仅 SetText。

- [x] **Step 1:** `NewTextLayer` + `SetText("GLITCH")`;Position(28kf)/Opacity(24kf) 经 `SetLayerTransform`（reopen 后,oracle 取值,opacity ×100 转 percent）。
- [x] **Step 2:** Tracking/Character Offset = text-animator leaf,本会话 RE（直接从原工程读 = 1D scalar，companion 自动 materialize）+ ship `AddTextTrackingAnimator`/`AddTextCharacterOffsetAnimator` + `AnimateText*`，双版本 render-gate PASS（commit 2add514）。
- [x] **Step 3:** 验收：Go round-trip 逐值对账 ✓ + AE2020≡AE2025 接受+render（GLITCH→PURCLQ）。账本 ② = 🔶待review。Commit。

### Task 1.3 — ③ マップ用フラクタルノイズ(id=69,2 Fractal Noise 层)〔依赖 Task 0.3〕

**结构:** 2 层同 footage(67) 源,各 1 个 Fractal Noise。L0 blend=Overlay,L1 blend=Normal。
- 共有非默认:Noise Type=1,Uniform Scaling=0(off),Offset Turbulence(2kf),Complexity=1,Evolution `expr="time*1200"`(L0)/`"time*3000\r"`(L1)。
- L0:Contrast=254,Scale Width=211,Scale Height=12。L1:Contrast=254,Scale Width=2847,Scale Height=12。

- [x] **Step 1:** 建 comp + 2 层。**footage-share 实测结论**:footage(67) = 黑 solid 1920×1080;无 from-scratch 共享 footage-item API(`SetSource`-retarget 留孤儿 / `DuplicateLayer` 继承 effect 需脆弱二次 reopen)→ **各自 1 黑 solid**(Fractal Noise 自生成像素 → 渲染等同,render-neutral delta)。`NewSolidLayer` 底部追加,建 A 后 B → A=index0(top)/B=index1,序对原工程。
- [x] **Step 2:** 各层 `AddEffect("ADBE Fractal Noise")` + `SetEffectParam` 标量(`-0002`NoiseType/`-0004`Contrast/`-0009`UniformScaling/`-0011`ScaleW/`-0012`ScaleH/`-0015`Complexity,值从 oracle 取)+ Offset Turbulence(`-0013`,2D 点)`AnimateEffectParamVec`(非 scalar 版)。
- [x] **Step 3:** Evolution(`-0023`)`SetEffectParam(0)`+`SetExpression(oracle 取)`+`SetExpressionEnabled(true)`(Task 0.3 GO,无需降级)。`SetBlendingMode(Overlay)` on L0。
- [x] **Step 4:** 验收 4 关全过(Go round-trip 逐值对账 ✓ + **AE2020≡AE2025 接受+DOM 确认**:2 层不 drop、Fractal Noise 32 props、Evolution `expr on`、Offset DOM=960,540、blend 对)。verify.jsx 加 `dumpEffects`(effect parade + 表达式状态,服务后续 ④⑦⑧⑨⑩⑪⑫)。账本 ③=🔶待review。**终帧 render-pixel 归 ⑫**。Commit。

### Task 1.4 — ④ カクッ(id=337,2 shape 层)

**结构:** 2 shape 层,层级 Scale(2kf ease)+ Opacity(20kf)动画;shape 几何由 oracle 提取(静态 path)。

- [x] **Step 1:** `NewShapeLayer` ×2。读侧实测:shape 几何**不是 freeform path 而是参数化** = Rect(1635×810)+ Stroke(0.8px)+ Trim Paths(Start=97/Offset=5.3);用 `AddRect/AddStroke/AddTrim`(非 path 提取)。Position 原=separated(0,0)=NewShapeLayer 默认,无需 centering。
- [x] **Step 2:** 层 `SetLayerTransform` Scale(2kf 64%→100%)/Opacity(20kf)+L0 RotateZ=180。**坑**:Scale 单位 percent 非 fraction,需 ×100(同 Opacity)。验收 4 关全过(Go round-trip 逐值 ✓ + AE2020≡AE2025 接受:4 comp、④ 2 层不 drop、2×Stroke+2×Trim 入 DOM、OK)。账本 ④=🔶待review。Commit。

---

## Phase 2 — 中层 comp

### Task 2.1 — ⑤ プリコンポジション 1(id=205,3 层,依赖①)✅(🔶待 review)
3 层均 src=① シェイイイイプ。`gen_precomp1.go`:`NewPrecompLayer(⑤, ①, name)` ×3。L0 Position(2kf [778→1204])、L1 Opacity(13kf)+居中、L2 静态居中;层 start 错峰 0.267/0.133/0.167。**两个坑**:①precomp 是 embed-template clone → pre-reopen scene `StartTime` 字段无效,须 post-reopen `SetStartTime`(ldta edit);②L1/L2 又撞 separated-position 读 0,0 → 显式居中(comp ④ 同款)。**验收**:Go round-trip(3 层 srcID 全=①、pos/op/start 全等)+ AE2020≡AE2025(5 comp、⑤ 3 层不 drop、**DOM 三层 source 全绑 comp ①** = precomp 解析正确)。verify.jsx 加 `l.source.name` dump。账本 ⑤=🔶待review。Commit。

### Task 2.2 — ⑥ シェイプの塊(id=243,7 层,依赖⑤)✅(🔶待 review)
7 层均 src=⑤。各层 Position(2kf)+ Opacity(34/41/39/39/39/44 kf,L6 无)。`gen_shape_katamari.go`:`NewPrecompLayer` ×7(复用 ⑤ 建法)+ 关键帧/静态 transform(oracle 取)。**Position 是绝对坐标(541→1340 横滑),非 ⑤ 的 separated 0,0 陷阱**,无需显式居中;只 L6(Position elided)居中。**带上 ⑤ 教训:复刻全 transform 通道(静态 Scale 24/33/75% + Rotate Z 90° on L0/L3/L4),非只 Position/Opacity**([[layer-replication-drops-static-transform-channels]]);anchor 源中心分数(0.5,0.5)([[setlayertransform-av-anchor-fraction]])。**验收 4 关全过**:Go 结构对账(7 层 srcID 全=⑤、kf 数 34/41/39/39/39/44、静态 Scale/Rotation/start 全等)+ AE2020≡AE2025 接受(6 comp、⑥ 7 层不 drop、DOM source 全绑 ⑤)+ render 眼验(t=1.0 渲出缩放 glitch 切片簇,无 blank/飞散)。无新 API。终帧 render-pixel 归 ⑫。账本 ⑥=🔶待review。Commit。

### Task 2.3 — ⑦ ここは開けない方が身のため(id=287,3 层,依赖⑥)✅(🔶待 review)
3 层 src=⑥。L0 Opacity(25kf)+Glow(Radius=0,Intensity=0.30);L1 Opacity(34kf)+Glow(Radius=0,Intensity=0.35);L2 无(隐形 Opacity=0)。`gen_hiraku.go`:`AddEffect("ADBE Glo2")`+`SetEffectParam(-0003=Radius,-0004=Intensity)`。**新维度 = Glow 挂 precomp 层**(首次 effect 上嵌套层)。Position 是**静态绝对坐标**(非 ⑥ 的 2kf、非 ⑤ 的 separated 0,0),L0/L1 Scale 101%;同 ⑥ 复刻全 transform 通道。**验收 4 关全过**:Go 结构对账(3 层 srcID 全=⑥、Position 静态值/Opacity kf 25·34·L2=0/Glow Radius=0·Intensity 0.30·0.35/start 全等)+ AE2020≡AE2025 接受(3 层不 drop、2×Glo2 入 DOM 值对)+ **render orig-vs-clone t=0.5 像素布局一致**(⑦ 中间 comp 孤立看淡=忠实,原工程同帧同样淡)。无新 API。终帧 render-pixel 归 ⑫。账本 ⑦=🔶待review。Commit。

### Task 2.4 — ⑧ RGBズレ(id=132,3 层,依赖②)✅(🔶待 review)
3 层 src=② テキスト。L0"B" Fill Color[A,R,G,B]=[255,0,131,255]+Opacity(23kf);**L1"R" Fill-0002 elided=默认红**(AE DOM 实测 v=1,0,0,1 确认模板默认即红→只 AddEffect 不设色,镜像原 elision);L2"G" Fill Color=[255,0,255,86]+Opacity(25kf)。`gen_rgbzure.go`:`AddEffect("ADBE Fill")`+`SetEffectParam("ADBE Fill-0002", []float64{A,R,G,B})`。**坑**:SetEffectParam 颜色要 `[]float64` len=Components(传 [4]float64 报 unsupported value type);编码 [A,R,G,B] 0-255=parser 读出格式,直接拷 oracle。全 Position 静态[960,540]居中,L2 start=**-0.1335 负值**正确 round-trip。**验收 4 关全过**:Go 结构对账(3 层 srcID=②、Position/Opacity 23·24·25kf/Fill 颜色 L0L2 精确·L1 elided/start 含负全等)+ AE2020≡AE2025 接受(3 层不 drop、3×Fill 入 DOM、颜色 AE 读回对 [A,R,G,B]→AE[R,G,B,A] 验证)+ **render 眼验**(t=1.0 红绿蓝三份文字重叠+相位错峰=RGB 色差)。无新 API。终帧 render-pixel 归 ⑫。账本 ⑧=🔶待review。Commit。

### Task 2.5 — ⑨ なんか周りのやつ(id=355,3 层,依赖④)✅(🔶待 review)
L0/L1 src=④ カクッ(precomp,角括号,L0 横向镜像);L2 = 新建 shape 层 "シェイプレイヤー 1"(Rect 1856×1015)带 **Trim Paths** 动画(Start 2kf 100→0 + Offset 2kf 0→720,**bezier ease 逐值拷**)+ Stroke。`gen_nanka.go`:**首个混合源 comp**(precomp + from-scratch shape)+ **首个非线性关键帧 ease 复刻**(`PropertyStream.AddKeyframeWithEase`,lowerShapeScalar→injectAnimatedStream)。Opacity wiggle(静态 opacity,SetExpression 安全)。**坑/delta**:L0 原 3D Rotate Y=180 → 用 Scale X=-100% 2D 负缩放等价镜像(库有 `SetIs3D`+`SetRotateY` gated 但与 precomp+SetLayerTransform 组合未测,记 mechanism delta)。**验收 4 关全过**:Go round-trip(3 层 src/trim 2kf ease 逐值精确/opacity expr/scale 全等)+ AE2020≡AE2025 接受(9 comp、⑨ 3 层不 drop、3×Trim group 入 DOM、3×Opacity wiggle ON)+ render 眼验(t=0.4 trim 局部 draw-on → t=1.0 闭合 + ④ 角括号镜像)。账本 ⑨=🔶待review。终帧 render-pixel 归 ⑫。Commit。

---

## Phase 3 — 怪物 comp

### Task 3.1 — ⑩ グリッチテキスト(id=85,27 层,~100 mask,依赖⑦②⑧③)

**这是工作量与风险中心。** 子结构(见 dissect L0–L26):
- **L0–L6** 7 个调整层 src=footage(319):各 Transform(Geometry2)+ Scale Height 动画/✎(157/248/144…)+ **每层 9–21 个 mask**(撕裂切片)。
- **L7–L8** 调整层 src=footage(110):Glow(Glo2)×2/×1(Threshold/Radius/Intensity ✎,L7 blend=Add)。
- **L9** src=⑦ ここは開けない。
- **L10–L20** 调整层 src=footage(376/100):各 Displacement Map(Layer 指向 mask 源 + Max H/V Displacement,部分 animated)+ 多 mask。
- **L21** src=② テキスト;**L22** src=⑧ RGBズレ;**L23"横ブラー"** src=② + Gaussian Blur×2;**L24"…さぶ"** Fractal Noise+Gaussian Blur(Evolution expr);**L25** src=③ マップ用ノイズ;**L26** footage(374) Fractal Noise(expr)。

- [x] **Step 1 ✅(commit 04a5e66):** 无 mask 骨架 27 层 — `gen_glitch_text.go` 读原 comp 层驱动(precomp→同名 clone comp / footage-adj→`NewAdjustmentLayer` / footage-av→`NewSolidLayer`),post-reopen 设 blend/start+in/out/visible。**Go 对账 27/27**(源种类/blend/可见/in-out frame-exact)+ **双版本 AE 接受 ⑩=27 层不 drop**。副产:修 `setLdtaFrac` 粗 divisor 截断 bug(commit 2886c5e,见下)。
- [x] **Step 2 ✅(commit 875fdd6):** 131 bbox-rect 撕裂切片 mask(L0–L20)。**读侧发现**:mask 几何 = shph @0x04 bbox(L,T,R,B 源空间分数)+ 归一化 unit-square ldat;`Mask.Vertices` 只 surface ldat → 全 mask 读着一样,真几何在 bbox。`maskBBoxRect` 解 bbox→像素矩形→`AddMask`(全 131 是 4 顶点 Add-mode 轴对齐矩形,审计确认)。**Go 对账 0 mismatch、bbox round-trip byte-exact**(maxErr 0)+ **双版本 AE 接受 131 mask 不 drop(含 21-mask 层)= 解 spike 0.4 many-mask 未知**。verify.jsx 加 per-layer mask count dump。
- [x] **Step 3 ✅(commit 88076b1):** effect 链 27 个 effect。curated per-effect tuned-param set(略 group/marker leaf,SetEffectParam 不碰 group):Geometry2 Scale Height(L0/L1 7kf anim、L2–L6 静态)/ Glo2 Threshold·Radius·Intensity(L7×2 Add、L8×1)/ Displacement Map Max H/V·Edge(L10–L20,L19/L20 animated)/ Gaussian Blur(L23×2、L24×1)/ Fractal Noise 全参 + Brightness/Offset anim + Evolution expr time*3000/time*2000(L24/L26)。**Displacement Map 层引用 RE 结论 = SELF-ref**:原版 12 个 map 的 tdpi == 自身宿主层 id(raw-rifx probe 实证),`AddEffect` 默认把 tdpi 绑宿主 → **无需 `SetEffectLayerParam`**,self-ref 自动复刻(clone probe 确认全 self)。**Go 对账 0 mismatch**(27 effect、match-name、标量值、kf 数、Evolution expr 全等)+ **双版本 AE 接受**(effect tally 12/7/5/3/4/3 一致、4× Evolution expressionEnabled=ON、工程开 clean)+ **render 眼验**(t=1.0 发光 chromatic glitch 文字、displacement 横向撕裂 + Glo2 halo,居中无 blank)。**⑩ COMPLETE（🔶待 review）**。终帧 render-pixel vs 原版归 ⑫。

> **⑩ グリッチテキスト 整体 COMPLETE（🔶待 review）**：27 层 + 131 mask + 27 effect,三 Step 全双版本 AE gated + Go 对账 byte/值 exact。副产库 fix:`setLdtaFrac` 粗 divisor 截断(commit 2886c5e)。读侧发现两条:mask 几何在 shph bbox(`Mask.Vertices` 只 surface 归一化 ldat)· Displacement Map tdpi self-ref。下一 = Phase 4 ⑪⑫。
- [ ] **Step 4:** 验收(此 comp 单独 render 一帧 Read png 自查 glitch 切片对)。账本 ⑩ 逐特征(masks/displacement/glow/noise)。Commit(可拆多 commit:骨架 / masks / effects)。

---

## Phase 4 — 顶层 + 终验

### Task 4.1 — ⑪ 背景変えるならココ！(id=43,6 层)〔含 wiggle + Ramp〕
- L0 footage(65) adjustment + Noise2(Amount=11);L1 footage(57) adjustment + Exposure2(**Exposure expr="wiggle(34,0.29)" 9kf** — Task 0.3 结论);L2 footage(63) Overlay + Venetian Blinds(Completion=97,Dir=90,Width=8);L3 footage(59) Venetian Blinds(62/90/12);L4 footage(285) masks=1;L5 footage(41) + **Ramp**(Start Color/End Color/End of Ramp/Ramp Shape=2 radial,Start of Ramp 2kf — 值见 dissect L5)。
- [x] **✅(commit 687dd12,🔶待 review)** `gen_haikei.go`:6 层全 solid(oracle 读色,L0/L1 adjustment),Exposure2/Venetian Blinds×2/Ramp(径向+2kf+4-comp 色 byte-exact)+L4 bbox-rect mask。Go 对账 BAD=0+双版本 AE 接受+render 眼验。**两条 honest delta**:① **Noise2 not in embedded set**→skip+log(能力 gap);② **Exposure wiggle 舍**(kf+expr→AE drop 层 + 级联,bisection 实证:⑪ 初次只进 1/6 层;= ⑧ keyframed-opacity wiggle 同坑扩到 effect param;保 9kf 主闪烁舍 ±0.29 wiggle)。

### Task 4.2 — ⑫ メインコンプ！(id=14,3 层)〔顶 + 终验〕
- L0 src=⑩ グリッチテキスト,blend=Add,Position(3kf ease)+Scale(1kf)+ **Venetian Blinds**(Completion=14,Dir=90,Width=9);L1 src=⑨;L2 src=⑪ + **Exposure2**(Exposure=-1.45)+ **Curves**(Task 0.5 结论)。
- [x] **Step 1 ✅(commit 33485e5):** 3 precomp 层(⑩/⑨/⑪)+ effects(L0 Venetian Blinds、L2 Exposure2 -1.45 + CurvesCustom 默认实例[曲线 arbitrary-data blocked])+ L0/L1 静态 Opacity wiggle(safe,⑤L2 型)。Go 对账全等 + 双版本 AE 接受(整 12-comp 工程 clean、CurvesCustom+VenetianBlinds 入 DOM、wiggle ON)。
- [x] **Step 2 终帧 render-pixel(红线4)✅:** clone vs 原版 メインコンプ t=1.0 双开渲染对照(tmp_debug)。**hero element 对**:居中 chromatic glitch 文字 + 辉光 + 横向 Venetian-Blinds 扫描线 + 暗背景。**honest delta**:⑪ 背景 flare/blinds 区在 clone 渲成锐利亮矩形 vs 原版柔和融合 → 溯因(a)**mask feather/opacity 未复刻**(锐利满不透明 vs 原柔边)(b)blocked Noise2 grain + 舍弃的 Exposure wiggle 调制原背景。**非像素等同**;结构复刻完成 + AE-gated,余为 documented polish/blocked delta。
- [x] **Step 3 ✅:** 双版本 AE ship-gate 全绿(整工程)。L0 Position kf 写 linear(LayerTransform 无 eased builder)= 75px nudge 微 delta。
- 注:Curves 曲线数据 arbitrary-data blocked(spike 0.5),仅加默认实例;**整 Booyah 12-comp from-scratch 复刻全建成 + 双版本 AE gated**。

### Task 4.3 — 收口 + landing
- [ ] INDEX 覆盖账本最终态(诚实标 blocked + 实证);showcase status=待review(用户真机验收后才 complete)。
- [ ] 新 incident/capindex tag 落地(spike 结论、新写路径)。
- [ ] `/flightdeck:landing` 归档 plan + 同步 cockpit。

---

## Self-Review(对 spec 覆盖)

- spec §3 三关 → 每 comp「验收」+ Task 4.2 终帧 render-pixel ✓
- spec §4 取值神谕/不 copy 字节 → Task 0.2 oracle 只读 + 各 gen 经 API 重建 ✓
- spec §6 DAG 12 comp → Phase1-4 全覆盖(①–⑫)✓
- spec §7 未知前沿 → Phase 0 spike 全覆盖(expr/many-mask/curves)+ text-animator(Task 1.2)/trim(2.5)/footage-share(1.3)/ramp(4.1)即遇即 RE ✓
- spec §8 覆盖账本 → Task 0.1 建骨架 + 各 comp 更新 + 4.3 收口 ✓
- 类型一致:`oracle`/`buildXxx(p, orc)` 命名贯穿;effect matchName 用 dissect 原名 ✓
