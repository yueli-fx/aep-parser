---
status: active
summary: 实现 Booyah Glitch 全工程复刻:Phase0 前置 spike(表达式 Evolution=time*N + wiggle / 单层 ~21 mask / Curves 曲线数据)→ Phase1-4 按 DAG 叶→根逐 comp 建(extract 值→写 gen_<comp>.go→AE 接受+verify.jsx DOM对账→覆盖账本→commit)→ 终帧 メインコンプ 像素对照原工程。
last_updated: 2026-06-19
implements: specs/2026-06-19-booyah-glitch-full-replication.md
---

# Booyah Glitch 复刻 — 实现计划

> **For agentic workers:** 本 plan 用 checkbox(`- [ ]`)跟踪。验证模型 = 本项目实际工作流(`go vet/build` 绿 + AE 双版本接受 + `verify.jsx` DOM 对账 + 终帧 render-pixel),**非** 通用 pytest TDD。每完成一个 comp/spike 即 commit(直推 main,永不 push)。

**Goal:** 用咱们的 Go API 从零重建整个 Booyah Glitch(12 comp),结构保真 + 终帧渲染对照原工程。

**Architecture:** `flightdeck/showcase/booyah-clone/` 一个 `package main`:`oracle.go` 用 `aep.Open` 把原工程当只读取值源;`gen_<comp>.go` 每 comp 一个 build 函数,经 New*/AddEffect/Set*/AddMask/Animate* 重建;`main.go` 按 DAG 拓扑序拼装 → 写 `booyah-clone.aep`。原工程只读、绝不 copy 字节。

**Tech Stack:** Go(internal/aep facade)· AE 2020+2025 ship-gate(`scripts/ae_run.ps1`)· `render.jsx`/`verify.jsx`(ExtendScript)。

**真相源:** 每 comp 结构/值取自 `tmp/booyah_dissect.txt`(完整画像)+ 执行时 `aep.Open` 实读。本 plan 内联的是**已知标量参数 + API 调用骨架**;mask 路径 / 关键帧值这类批量数据由 `oracle.go` 程序化提取,不内联。

---

## Phase 0 — 前置 spike + 基建(先把未知打掉,再建)

de-risk 原则:表达式/many-mask/curves 三个未知一旦 blocked,会改变后续 comp 的建法,必须最前置。

### Task 0.1 — Scaffold booyah-clone 包

**Files:**
- Create: `flightdeck/showcase/booyah-clone/main.go`(package main,空 `func main()` 先写出 `booyah-clone.aep` 一个空 project)
- Create: `flightdeck/showcase/booyah-clone/oracle.go`(`openOriginal()` 用 `aep.Open` 打开 `samples/.../Booyah Glitch.aep`,返回 `*aep.Project`)
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

- [ ] **Step 1:** 对 `グリッチテキスト` L0 跑访问器,打印它的 effects+params+mask 数+keyframe;比对 `tmp/booyah_dissect.txt` L0 值一致。
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

- [ ] **Step 1:** `NewTextLayer` + `SetText("GLITCH")`;Position/Opacity 关键帧(oracle 取值)。
- [ ] **Step 2:** Tracking/Character Offset 动画 → 查 capindex 是否有对应 setter;无 → RE(text-animator)或账本标 `待 RE`。**不绕**:tracking 是 btdk/text-animator,先确认可写性。
- [ ] **Step 3:** 验收。账本 ② = 待review / 部分待RE。Commit。

### Task 1.3 — ③ マップ用フラクタルノイズ(id=69,2 Fractal Noise 层)〔依赖 Task 0.3〕

**结构:** 2 层同 footage(67) 源,各 1 个 Fractal Noise。L0 blend=Overlay,L1 blend=Normal。
- 共有非默认:Noise Type=1,Uniform Scaling=0(off),Offset Turbulence(2kf),Complexity=1,Evolution `expr="time*1200"`(L0)/`"time*3000\r"`(L1)。
- L0:Contrast=254,Scale Width=211,Scale Height=12。L1:Contrast=254,Scale Width=2847,Scale Height=12。

- [ ] **Step 1:** 建 comp + 2 层(footage 源:确认 footage 共享建法 —— 一个 footage item 两层引用,或各自 solid;首遇 footage-share,实测确定)。
- [ ] **Step 2:** 各层 `AddEffect("ADBE Fractal Noise")` + `SetEffectParam` 上述标量(matchName 见 dissect:`ADBE Fractal Noise-0004`=Contrast 等)+ Offset Turbulence `AnimateEffectParam`。
- [ ] **Step 3:** Evolution 按 Task 0.3 结论:SetExpression 或 keyframe 近似。`SetBlendingMode(Overlay)` on L0。
- [ ] **Step 4:** 验收。账本 ③。Commit。

### Task 1.4 — ④ カクッ(id=337,2 shape 层)

**结构:** 2 shape 层,层级 Scale(2kf ease)+ Opacity(20kf)动画;shape 几何由 oracle 提取(静态 path)。

- [ ] **Step 1:** `NewShapeLayer` ×2,oracle 提取各自 shape path → 重建。
- [ ] **Step 2:** 层 Scale/Opacity 关键帧。验收。账本 ④。Commit。

---

## Phase 2 — 中层 comp

### Task 2.1 — ⑤ プリコンポジション 1(id=205,3 层,依赖①)
3 层均 src=① シェイイイイプ。L0 Position(2kf)、L1 Opacity(13kf)、L2 无动画。`NewPrecompLayer(comp, ①, name)` ×3 + 关键帧。验收。Commit。

### Task 2.2 — ⑥ シェイプの塊(id=243,7 层,依赖⑤)
7 层均 src=⑤。各层 Position(2kf)+ Opacity(34/41/39/39/39/44 kf,L6 无)。`NewPrecompLayer` ×7 + 关键帧(oracle 取)。验收。Commit。

### Task 2.3 — ⑦ ここは開けない方が身のため(id=287,3 层,依赖⑥)
3 层 src=⑥。L0 Opacity(25kf)+Glow(Radius=0,Intensity=0.30);L1 Opacity(34kf)+Glow(Radius=0,Intensity=0.35);L2 无。`AddEffect("ADBE Glo2")`+SetEffectParam。验收。Commit。

### Task 2.4 — ⑧ RGBズレ(id=132,3 层,依赖②)
3 层 src=② テキスト。L0"B" Fill Color=[255,0,131,255]+Opacity(23kf);L1"R" Fill(默认色?dissect 未标✎)+Opacity(24kf);L2"G" Fill Color=[255,0,255,86]+Opacity(25kf)。`AddEffect("ADBE Fill")`+`SetEffectParam(Color)`。**这是 showcase/glitch 碰过的 RGB 色差路**,复用经验。验收。Commit。

### Task 2.5 — ⑨ なんか周りのやつ(id=355,3 层,依赖④)
L0/L1 src=④ カクッ(无动画);L2 shape 层 "シェイプレイヤー 1" 带 **Trim Paths** 动画(`ADBE Vector Trim Start` 2kf ease + `ADBE Vector Trim Offset` 2kf ease)。`AddTrim`(`incidents/trim-paths-vector-filter-re.md`)+ animate。验收。Commit。

---

## Phase 3 — 怪物 comp

### Task 3.1 — ⑩ グリッチテキスト(id=85,27 层,~100 mask,依赖⑦②⑧③)

**这是工作量与风险中心。** 子结构(见 dissect L0–L26):
- **L0–L6** 7 个调整层 src=footage(319):各 Transform(Geometry2)+ Scale Height 动画/✎(157/248/144…)+ **每层 9–21 个 mask**(撕裂切片)。
- **L7–L8** 调整层 src=footage(110):Glow(Glo2)×2/×1(Threshold/Radius/Intensity ✎,L7 blend=Add)。
- **L9** src=⑦ ここは開けない。
- **L10–L20** 调整层 src=footage(376/100):各 Displacement Map(Layer 指向 mask 源 + Max H/V Displacement,部分 animated)+ 多 mask。
- **L21** src=② テキスト;**L22** src=⑧ RGBズレ;**L23"横ブラー"** src=② + Gaussian Blur×2;**L24"…さぶ"** Fractal Noise+Gaussian Blur(Evolution expr);**L25** src=③ マップ用ノイズ;**L26** footage(374) Fractal Noise(expr)。

- [ ] **Step 1:** 先建**无 mask 骨架**(27 层 + 源引用 + blend + 层级动画),AE 接受 → 锚定层结构对。
- [ ] **Step 2:** **批量 mask**:oracle 提取 L0–L20 每层 mask bezier paths → `AddMask`/`SetMaskPath` 程序化重建(用 Task 0.4 验过的 many-mask 路径)。这是「小提取器」长出来的地方(`copyMasksFromOriginal(origLayer, newLayer)`)。
- [ ] **Step 3:** effect 链:各 Geometry2/Glo2/Displacement Map(`SetEffectLayerParam` 指 mask/noise 源层)+ Fractal Noise + Gaussian Blur,params 由 oracle 取;animated 的走 AnimateEffectParam。Evolution 按 0.3 结论。
- [ ] **Step 4:** 验收(此 comp 单独 render 一帧 Read png 自查 glitch 切片对)。账本 ⑩ 逐特征(masks/displacement/glow/noise)。Commit(可拆多 commit:骨架 / masks / effects)。

---

## Phase 4 — 顶层 + 终验

### Task 4.1 — ⑪ 背景変えるならココ！(id=43,6 层)〔含 wiggle + Ramp〕
- L0 footage(65) adjustment + Noise2(Amount=11);L1 footage(57) adjustment + Exposure2(**Exposure expr="wiggle(34,0.29)" 9kf** — Task 0.3 结论);L2 footage(63) Overlay + Venetian Blinds(Completion=97,Dir=90,Width=8);L3 footage(59) Venetian Blinds(62/90/12);L4 footage(285) masks=1;L5 footage(41) + **Ramp**(Start Color/End Color/End of Ramp/Ramp Shape=2 radial,Start of Ramp 2kf — 值见 dissect L5)。
- [ ] Noise2/Exposure2/Venetian Blinds/Ramp 各 AddEffect+SetEffectParam;Ramp 颜色+点(`ADBE Ramp-000x`);L4 一个 mask。wiggle 按 0.3。验收。账本 ⑪。Commit。

### Task 4.2 — ⑫ メインコンプ！(id=14,3 层)〔顶 + 终验〕
- L0 src=⑩ グリッチテキスト,blend=Add,Position(3kf ease)+Scale(1kf)+ **Venetian Blinds**(Completion=14,Dir=90,Width=9);L1 src=⑨;L2 src=⑪ + **Exposure2**(Exposure=-1.45)+ **Curves**(Task 0.5 结论)。
- [ ] **Step 1:** 拼 3 层 + effects。验收(AE 接受 + 结构对账)。
- [ ] **Step 2(终帧 render-pixel,红线4):** `render.jsx` 渲染 メインコンプ 同一帧(选 glitch 活跃帧,如 t=1.0s)→ 与原工程同 comp 同帧 render 像素对照。Read 两张 png 目视 + 数值差。Expected:视觉等同(色差/撕裂/辉光/背景一致)。
- [ ] **Step 3:** 双版本 AE ship-gate 全绿。账本 ⑫ + 全表收口(已复刻/blocked 逐项实证)。Commit。

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
