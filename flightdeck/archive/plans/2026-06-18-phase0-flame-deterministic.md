---
status: done
summary: 证命门:纯 Go 手写一条火焰效果栈(固态+Fractal Noise+Turbulent Displace+Tint+evolution 动画)→ AE 双版本实渲 PNG → 看图。结论=狭义命门过(库能确定性造可辨认火焰 + 效果栈/param/animate/mask 全 gated),但产品质量被用户真机否决(只有形态、不像火焰)→ roadmap 改为先 Phase 2 学真实样本再重做。镜像 orbit_demo render harness。
last_updated: 2026-06-18
implements: specs/2026-06-18-procedural-fx-generator.md
---

# Phase 0 — 确定性火焰(make-or-break)Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 纯 Go 确定性地造出一个「看得出是火焰」的 .aep(固态层 + 效果栈),AE 2020+2025 实渲一帧 PNG,人眼确认像火焰——证明这个库能确定性产出好的程序化 FX。

**Architecture:** 镜像已验证的 `orbit_demo_shipgate_test.go` from-scratch render harness:Go 用 `NewSolidLayer`+`AddEffect`+`SetEffectParam`+`AnimateEffectParam` 组装效果栈 → `WriteAEP` → AE JSX `saveFrameToPng` 渲一帧 → Go 解码 PNG 做像素/目视检查。无 AI、无网站。效果栈走 AE 原生 `ADBE Fractal Noise` + `ADBE Turbulent Displace` + `ADBE Tint`(全部 effects 域已 ae-accept gated)。

**Tech Stack:** Go(internal/aep facade)· AE ExtendScript(渲染)· `scripts/ae_run.ps1`(无人值守驱动 AE)· image/png(像素检查)。

**铁律(spec 2026-06-18-procedural-fx-generator):** 红线4 渲染类先看图;每增量 AE 实渲 + Read png 自验,不靠值 round-trip。**make-or-break**:Task 6 渲不出像火焰的东西 → 当场叫停,按 spec fallback 到数据条族,不进 Phase 1。

---

## File Structure

- `test_data/probe_flame_params.jsx`(新建)— AE 侧 dump Fractal Noise/Turbulent Displace/Tint 各参数 matchName↔人类名,产出参数表(Go 侧只见数字 ID,人类名要 AE 解析)。
- `internal/aep/flame_demo_shipgate_test.go`(新建)— 主体:`buildFlameDemo` 组装效果栈 + `runFlameDemoGate` 双版本渲染门。镜像 `orbit_demo_shipgate_test.go`。
- `test_data/verify_flame_demo.jsx`(新建)— AE 打开 + `saveFrameToPng` 渲指定帧 + 写 done。镜像 `test_data/verify_orbit_demo.jsx`。
- `flightdeck/showcase/procedural-fx/`(新建 `INDEX.md` + `gen.go` + `render.jsx`)— 渲染类必出 showcase(rules.md)。Task 7 落。

**复用参考(读但不改)**:`internal/aep/orbit_demo_shipgate_test.go`(render harness 模板)· `test_data/verify_orbit_demo.jsx`(saveFrameToPng 机制)· `internal/aep/example_effect_test.go`(SetEffectParam 用法)· `internal/aep/animate_effect_param_shipgate_test.go`(AnimateEffectParam 用法)· `incidents/procedural-fx-over-vector.md`(火焰技术=fractal noise,非矢量)· `incidents/effect-param-elision-synthesis-lite.md`(param 动画已知缺口)。

---

### Task 1: 探明 Fractal Noise / Turbulent Displace / Tint 的参数 matchName↔人类名

**Files:**
- Create: `test_data/probe_flame_params.jsx`

Go 侧 `fx.Parameters` 的 `.Name` 不解析人类名(只镜像数字 matchName,如 `ADBE Fractal Noise-0007`)。`SetEffectParam` 需要完整 matchName,所以要先在 AE 里建这三个效果、dump 每个参数的 `matchName` + `name`,得到「Contrast/Brightness/Evolution/Complexity/Scale… → ADBE Fractal Noise-NNNN」对照表,供后续 Task 硬编码。

- [ ] **Step 1: 写探测 JSX**

```javascript
// test_data/probe_flame_params.jsx — dump param matchName↔name for flame effects.
(function () {
    var dir = "e:/projects/tools/aep-parser/test_data/";
    var log = [];
    function dump(effMN) {
        var comp = app.project.items.addComp("P", 640, 360, 1, 3, 30);
        var sol = comp.layers.addSolid([0, 0, 0], "S", 640, 360, 1);
        var fx;
        try { fx = sol.property("ADBE Effect Parade").addProperty(effMN); }
        catch (e) { log.push(effMN + ": ADD FAIL " + e.toString()); return; }
        log.push("=== " + effMN + " (" + fx.numProperties + " params) ===");
        for (var i = 1; i <= fx.numProperties; i++) {
            var p = fx.property(i);
            log.push("  " + p.matchName + "  ::  " + p.name);
        }
        comp.remove();
    }
    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); app.newProject();
        dump("ADBE Fractal Noise");
        dump("ADBE Turbulent Displace");
        dump("ADBE Tint");
    } catch (e) { log.push("EXC " + e.toString() + " line=" + e.line); }
    var m = new File(dir + "probe_flame_params.done"); m.open("w"); m.write(log.join("\n")); m.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
```

- [ ] **Step 2: 跑它(AE2025 足够,纯读)**

Run:
```bash
rm -f test_data/probe_flame_params.done
pwsh -NoProfile -File scripts/ae_run.ps1 -AeExe "E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe" -Jsx "E:/projects/tools/aep-parser/test_data/probe_flame_params.jsx" -Done "e:/projects/tools/aep-parser/test_data/probe_flame_params.done" -TimeoutSec 180
cat test_data/probe_flame_params.done
```
Expected: 三个效果各自的参数 matchName↔name 列表。**记下** Fractal Noise 的:Contrast / Brightness / Complexity / Evolution / 以及 Transform 下的 Scale(可能是 `Scale Height`)对应的 `ADBE Fractal Noise-NNNN`;Tint 的 Map-Black-To(`ADBE Tint-0001`)/ Map-White-To(`ADBE Tint-0002`)/ Amount(`ADBE Tint-0003`);Turbulent Displace 的 Amount / Size / Complexity。

- [ ] **Step 3: 提交探测脚本 + 参数表注释**

把得到的对照表作为注释写进 `flame_demo_shipgate_test.go` 顶部(Task 2 建文件时)。本步先只提交 JSX:
```bash
git add -f test_data/probe_flame_params.jsx
git commit -m "chore(flame): Phase0 Task1 — AE probe dumps flame effect param matchName↔name"
```

---

### Task 2: 渲染骨架 — 固态 + Fractal Noise(默认参数)能渲出非黑帧

**Files:**
- Create: `internal/aep/flame_demo_shipgate_test.go`
- Create: `test_data/verify_flame_demo.jsx`

先把「Go 造层 → AE 渲一帧 PNG → Go 检查」整条链路打通,效果只加 Fractal Noise(默认),目标:渲出的帧**不是纯黑/纯空**(证明效果在渲染、harness 通)。

- [ ] **Step 1: 写 verify JSX(镜像 verify_orbit_demo.jsx 的 saveFrameToPng)**

先读 `test_data/verify_orbit_demo.jsx` 抄渲帧机制,改成开 "FLAME" comp、在 comp 中点(duration/2)`saveFrameToPng`:

```javascript
// test_data/verify_flame_demo.jsx — open + render mid-frame PNG + done.
(function () {
    var a = new File("e:/projects/tools/aep-parser/test_data/flame_demo_args.json");
    a.open("r"); var args = eval("(" + a.read() + ")"); a.close();
    var log = []; var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    try {
        app.open(new File(args.input));
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "FLAME") { comp = it; break; }
        }
        if (!comp) fail("comp FLAME not found");
        else { comp.saveFrameToPng(comp.duration / 2, new File(args.png)); log.push("  rendered t=" + (comp.duration / 2)); }
    } catch (e) { fail("EXC " + e.toString() + " line=" + e.line); }
    var d = new File(args.done); d.open("w"); d.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n")); d.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
```
> 注:若 `saveFrameToPng` 在某版本不可用,回退 `verify_orbit_demo.jsx` 用的渲法(读它确认实际机制)。

- [ ] **Step 2: 写 flame_demo_shipgate_test.go 骨架(buildFlameDemo + runFlameDemoGate)**

镜像 `orbit_demo_shipgate_test.go` 结构。`buildFlameDemo` 先只:黑底 1080×1920 竖构图(火焰偏竖)solid + `AddEffect(ADBE Fractal Noise)`:

```go
package aep_test

import (
	"image"
	_ "image/png"
	"fmt"; "os"; "path/filepath"; "strings"; "testing"
	aep "github.com/example/aep-parser/internal/aep"
)

func buildFlameDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "FLAME", 1080, 1920, 30, 4)
	if err != nil { t.Fatalf("NewComposition: %v", err) }
	sol, err := aep.NewSolidLayer(comp, "Flame", 1080, 1920, [3]float64{0, 0, 0})
	if err != nil { t.Fatalf("NewSolidLayer: %v", err) }
	fn, err := aep.AddEffect(sol, aep.EffectFractalNoise)
	if err != nil { t.Fatalf("AddEffect FractalNoise: %v", err) }
	_ = fn
	return p
}

func runFlameDemoGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" { t.Skip("set AE_SHIP_GATE=1") }
	const argsPath = `e:/projects/tools/aep-parser/test_data/flame_demo_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_flame_demo.jsx`
	toFwd := func(s string) string { return strings.ReplaceAll(s, `\`, `/`) }
	p := buildFlameDemo(t, target)
	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "flame_in.aep")
	doneFile := filepath.Join(tempDir, "flame.done")
	framePNG := filepath.Join(tempDir, "flame_frame.png")
	out, err := os.Create(inputAEP); if err != nil { t.Fatal(err) }
	if err := p.WriteAEP(out); err != nil { out.Close(); t.Fatalf("WriteAEP: %v", err) }
	out.Close()
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"png":%q}`, toFwd(inputAEP), toFwd(doneFile), toFwd(framePNG))
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil { t.Fatal(err) }
	defer os.Remove(argsPath); os.Remove(doneFile)
	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 240)
	body, _ := os.ReadFile(doneFile)
	t.Logf("flame %s:\n%s", ver, string(body))
	if l := strings.SplitN(string(body), "\n", 2); len(l) == 0 || strings.TrimSpace(l[0]) != "PASS" {
		t.Fatalf("flame %s render FAIL:\n%s", ver, string(body))
	}
	// non-black check: at least some pixel is meaningfully bright.
	f, err := os.Open(framePNG); if err != nil { t.Fatalf("png: %v", err) }
	defer f.Close()
	img, _, err := image.Decode(f); if err != nil { t.Fatalf("decode: %v", err) }
	bright := 0
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y += 17 {
		for x := b.Min.X; x < b.Max.X; x += 17 {
			r, g, bl, _ := img.At(x, y).RGBA()
			if r>>8 > 40 || g>>8 > 40 || bl>>8 > 40 { bright++ }
		}
	}
	if bright == 0 { t.Errorf("flame %s: frame is all black — effect not rendering", ver) }
	t.Logf("flame %s: %d bright sample points (see %s)", ver, bright, framePNG)
}

func TestFlameDemo_AEShipGate_AE2025(t *testing.T) { runFlameDemoGate(t, ae2025(), "AE2025", aep.TargetAE2025) }
func TestFlameDemo_AEShipGate_AE2020(t *testing.T) { runFlameDemoGate(t, ae2020(), "AE2020", aep.TargetAE2020) }
```

- [ ] **Step 3: vet + Go round-trip(无 AE)先过**

Run: `go vet ./internal/aep/ && go test ./internal/aep/ -run TestFlameDemo -count=1`
Expected: PASS(AE_SHIP_GATE 未设 → Skip;只验编译 + build 不 panic)。

- [ ] **Step 4: AE2025 渲染骨架**

Run: `AE_SHIP_GATE=1 go test ./internal/aep/ -run TestFlameDemo_AEShipGate_AE2025 -count=1 -v`
Expected: PASS,日志报 bright sample points > 0。**Read 那个 flame_frame.png 目视**(从日志路径拷出或在 tempdir):应见灰色噪声纹理(还不是火焰,先证渲染通)。

- [ ] **Step 5: 提交**

```bash
git add internal/aep/flame_demo_shipgate_test.go
git add -f test_data/verify_flame_demo.jsx
git commit -m "feat(flame): Phase0 Task2 — solid+FractalNoise renders non-black frame (harness up)"
```

---

### Task 3: 调 Fractal Noise 成「火焰状噪声」(静态参数)

**Files:**
- Modify: `internal/aep/flame_demo_shipgate_test.go`(`buildFlameDemo`)

用 Task 1 的参数表,把 Fractal Noise 调成竖向拉伸、高对比的火苗状噪声。典型:Fractal Type=Basic、Contrast 高(~200)、Brightness 适中、Complexity ~6、Transform 里 Scale Height >> Scale Width(竖拉)。

- [ ] **Step 1: 在 buildFlameDemo 里 SetEffectParam(用 Task1 实测的 matchName)**

把 `_ = fn` 换成(matchName 以 Task1 实测为准,以下 NNNN 占位需替换成真实 ID):
```go
must := func(label string, _ *aep.Property, err error) {
	t.Helper(); if err != nil { t.Fatalf("%s: %v", label, err) }
}
must("Contrast", aep.SetEffectParam(sol, fn, "ADBE Fractal Noise-<Contrast>", 220.0))
must("Brightness", aep.SetEffectParam(sol, fn, "ADBE Fractal Noise-<Brightness>", 8.0))
must("Complexity", aep.SetEffectParam(sol, fn, "ADBE Fractal Noise-<Complexity>", 6.0))
must("ScaleHeight", aep.SetEffectParam(sol, fn, "ADBE Fractal Noise-<ScaleHeight>", 400.0))
must("ScaleWidth", aep.SetEffectParam(sol, fn, "ADBE Fractal Noise-<ScaleWidth>", 60.0))
```
> 若某参数是「curve/layer-ref 类」被 SetEffectParam refuse(见 incident effect-param-elision-synthesis-lite),换用别的标量参数或跳过该项,boundary 记下。

- [ ] **Step 2: AE2025 渲 + 看图**

Run: `AE_SHIP_GATE=1 go test ./internal/aep/ -run TestFlameDemo_AEShipGate_AE2025 -count=1 -v`
Expected: PASS。**Read flame_frame.png**:应见竖向拉丝的高对比噪声(像火苗的「乱」,但还没颜色/形状)。不像就回 Step 1 调参数(visual loop)。

- [ ] **Step 3: 提交**

```bash
git add internal/aep/flame_demo_shipgate_test.go
git commit -m "feat(flame): Phase0 Task3 — tune FractalNoise to vertical flame-like noise"
```

---

### Task 4: 上色 — Tint 把黑→暗红、白→黄(火焰配色)

**Files:**
- Modify: `internal/aep/flame_demo_shipgate_test.go`

- [ ] **Step 1: AddEffect Tint + SetEffectParam Map-Black-To / Map-White-To**

在 Fractal Noise 之后加(颜色值是 ARGB/RGB,按 SetEffectParam color 类型;以 example_effect_test.go 的 color 写法为准):
```go
tn, err := aep.AddEffect(sol, aep.EffectTint)
if err != nil { t.Fatalf("AddEffect Tint: %v", err) }
must("MapBlack", aep.SetEffectParam(sol, tn, "ADBE Tint-<MapBlackTo>", []float64{0.15, 0.0, 0.0})) // 深红
must("MapWhite", aep.SetEffectParam(sol, tn, "ADBE Tint-<MapWhiteTo>", []float64{1.0, 0.85, 0.2})) // 亮黄
```
> color 参数的具体 Go 类型(`[]float64` len3/4 或别的)以 Task1 + example_effect_test.go 实测为准。

- [ ] **Step 2: 渲 + 像素检查(暖色出现)+ 看图**

加一个像素断言:中部区域应有「红/橙/黄」暖色(R 明显 > B)。在 runFlameDemoGate 的 png 检查里补:
```go
warm := 0
for y := b.Min.Y; y < b.Max.Y; y += 17 {
	for x := b.Min.X; x < b.Max.X; x += 17 {
		r, g, bl, _ := img.At(x, y).RGBA()
		if r>>8 > 80 && int(r>>8)-int(bl>>8) > 40 && int(r>>8) >= int(g>>8) { warm++ }
	}
}
if warm < 3 { t.Errorf("flame %s: too few warm (fire-colored) pixels = %d", ver, warm) }
```
Run AE2025 gate。**Read png**:应见暗红底 + 黄亮芯的火焰配色。

- [ ] **Step 3: 提交**

```bash
git add internal/aep/flame_demo_shipgate_test.go
git commit -m "feat(flame): Phase0 Task4 — Tint maps noise to fire palette (warm-pixel check)"
```

---

### Task 5: 塑形 — Turbulent Displace + 竖向衰减让它像火苗轮廓

**Files:**
- Modify: `internal/aep/flame_demo_shipgate_test.go`

让平铺的彩色噪声收成「下宽上窄、底实顶虚」的火苗形。手段:Turbulent Displace(扭曲边缘)+ 一个竖向亮度衰减(底亮顶暗)。衰减最简做法 = 再叠一个 Fractal/Ramp 或用 mask;先试 **Turbulent Displace 扭形**,形不够再加衰减层。

- [ ] **Step 1: AddEffect Turbulent Displace + 参数**

```go
td, err := aep.AddEffect(sol, aep.EffectTurbulentDisplace)
if err != nil { t.Fatalf("AddEffect TurbulentDisplace: %v", err) }
must("TD-Amount", aep.SetEffectParam(sol, td, "ADBE Turbulent Displace-<Amount>", 60.0))
must("TD-Size", aep.SetEffectParam(sol, td, "ADBE Turbulent Displace-<Size>", 40.0))
```
> 顺序:效果栈应是 FractalNoise → Tint → TurbulentDisplace(扭已上色的火),若 AddEffect 追加在末尾正好。

- [ ] **Step 2: 渲 + 看图(make-or-break 预检)**

Run AE2025 gate。**Read png**:目标是「一眼看出是火焰」。**这是命门的第一道目视**——像 → 进 Task 6;明显不像(平铺噪声/糊成一团)→ 先试加竖向衰减(底亮顶暗的 Ramp/第二 Fractal 做 luma),仍不像 → 触发 spec fallback(停火焰、转数据条),在 plan 末尾「verdict」记原因。

- [ ] **Step 3: 提交**

```bash
git add internal/aep/flame_demo_shipgate_test.go
git commit -m "feat(flame): Phase0 Task5 — Turbulent Displace shapes flame silhouette"
```

---

### Task 6: 动起来 — animate Evolution(火焰翻腾)

**Files:**
- Modify: `internal/aep/flame_demo_shipgate_test.go`

火焰的「活」靠 Fractal Noise 的 Evolution 随时间转。用 `AnimateEffectParam`(已 gated,见 animate_effect_param_shipgate_test.go)给 Evolution 打两个关键帧(0° → N×360°)。

- [ ] **Step 1: AnimateEffectParam Evolution**

```go
if _, err := aep.AnimateEffectParam(sol, fn, "ADBE Fractal Noise-<Evolution>", []aep.ScalarKeyframe{
	{Time: 0, Value: 0},
	{Time: 4, Value: 1440}, // 4 圈 over 4s
}); err != nil {
	t.Fatalf("AnimateEffectParam Evolution: %v", err)
}
```
> **已知缺口风险**(incident effect-param-elision-synthesis-lite):若 AnimateEffectParam 对 elided Evolution refuse,先 `SetEffectParam` 物化 Evolution 静态值再 animate;仍不行则 fallback = 给 Evolution 挂表达式 `time*360`(SetExpression,已 gated),boundary 记下走了哪条。

- [ ] **Step 2: 验运动(渲两帧,比像素)**

改 verify JSX 渲两帧(t=1 和 t=3,各存一个 png),Go 解码两图断言「明显不同」(火在动):
```go
// 在 gate 里渲两帧后:
if framesIdentical(pngA, pngB) { t.Errorf("flame %s: frames at t=1,t=3 identical — evolution not animating", ver) }
```
(framesIdentical = 采样若干点全相等则 true;实现仿 checkOrbitRenderedPixels。)
Run AE2025 gate。**Read 两帧 png**:火苗形态应明显变化。

- [ ] **Step 3: 提交**

```bash
git add internal/aep/flame_demo_shipgate_test.go test_data/verify_flame_demo.jsx
git commit -m "feat(flame): Phase0 Task6 — animate Evolution, verify motion across 2 frames"
```

---

### Task 7: 双版本门 + showcase + make-or-break verdict + 交用户真机验

**Files:**
- Create: `flightdeck/showcase/procedural-fx/INDEX.md`, `gen.go`, `render.jsx`

- [ ] **Step 1: 跑 AE2020(双版本)**

Run: `AE_SHIP_GATE=1 go test ./internal/aep/ -run TestFlameDemo_AEShipGate_AE2020 -count=1 -v`
Expected: PASS + 暖色/运动检查过。**Read AE2020 帧**目视一致。若 AE2020 渲染与 2025 差异大,boundary 记。

- [ ] **Step 2: 落 showcase(rules.md 渲染类必出)**

`flightdeck/showcase/procedural-fx/`:`gen.go`(package main,把 buildFlameDemo 逻辑提成可独立 `go run` 的生成器,产 flame.aep)+ `render.jsx`(渲 png)+ `INDEX.md`(frontmatter 写测哪个方向 + status: 待review)。`*.aep`/`*.png` 已 gitignore,`gen.go`/`render.jsx`/`INDEX.md` tracked,保持 `go build ./...` 绿。

- [ ] **Step 3: verdict + 提交**

在本 plan 末尾追加 `## Verdict` 段:火焰像不像、走了哪些 fallback、Phase 0 过/不过、给 Phase 1 的 carry-over(参数表 + 哪些 param 动画有坑)。
```bash
git add flightdeck/showcase/procedural-fx/ flightdeck/plans/2026-06-18-phase0-flame-deterministic.md
git add -f flightdeck/showcase/procedural-fx/render.jsx
git commit -m "feat(flame): Phase0 Task7 — double-version gate + showcase + verdict"
```

- [ ] **Step 4: 交用户真机验收(agent 眼验 ≠ 用户验收,rules.md showcase review-gate)**

通知用户:在真 AE 打开 showcase 产物复核火焰质量。**用户点头前 showcase 标 `待review`,不自标 complete。** 这也是 Phase 0 make-or-break 的最终判定:用户认可「像火焰且能接受」→ Phase 0 通过,进 Phase 1(参数化);用户否决 → 按 spec fallback 数据条族,记入 verdict。

---

## Self-Review 注记(写计划时已核)

- **Spec 覆盖**:Phase 0 = spec 的 Phase 0 全部(手写火焰 + 渲染 + 看图 + make-or-break + fallback 钩子)。Phase 1-4 不在本 plan(各自独立 plan)。
- **已知缺口显式挂钩**:effect-param 动画缺口(Task 6 fallback 表达式)· curve 类 param refuse(Task 3/4 fallback)· saveFrameToPng 可用性(Task 2 fallback 读 orbit jsx 机制)。这些是真实 unknown,故 Task 1 + 各 fallback 显式写出,不是 placeholder。
- **`<Contrast>`/`<Evolution>` 等尖括号 = Task 1 必须先实测填的真实数字 matchName**,不是可跳过的占位——Task 1 是硬前置。

---

## Verdict(2026-06-18,执行完成)

**Phase 0 = agent 眼验 PASS,待用户真机验收(make-or-break 最终判定)。**

- Task1(probe 参数表 9aee→a1f)· Task2(harness)· Task3-5(静态火焰)· Task6(运动)· Task7(双版本+showcase)全做完。commit:Task1=…Task2=…Task3-5=`3429f18`·Task6=`ce1d312`·Task7=本次。
- **双版本 gate `TestFlameDemo_AEShipGate_AE2020/AE2025` 都 PASS**,渲染指标一致:bright≈1473 / warm≈1141 / 两帧 diff≈1472(运动活)。AE2020 与 AE2025 渲染帧逐像素一致。
- **agent Read png 眼验**:水滴形火苗,竖向橙黄火舌 + 黑间隙,羽化柔边,t=1/t=3 内部纹理明显翻腾 → **一眼看得出是火焰**。命门过。
- **走的路径**:effect-stack(FractalNoise tall + Tint + TurbulentDisplace + 羽化 teardrop mask + AnimateEffectParam Evolution),全程 gated 公共 API,**无从零字节裸写**。
- **已知缺口未撞**:`AnimateEffectParam` 给 elided Fractal Noise/Turbulent Displace 的 Evolution 打关键帧**直接成功**(incident effect-param-elision-synthesis-lite 的 refuse 未发生,本场景标量 angle 参数走得通)→ Phase 1 参数化无需 fallback 表达式。
- **未触发 fallback**(没退数据条):火焰一次成型。
- **carry-over to Phase 1**:① 参数表已在 `flame_demo_shipgate_test.go` 顶部注释 + showcase gen.go;② 可改进项(底更红/芯更亮/加 Glow 泛光)留 Phase 1 参数化时做;③ 配方=「effect-stack recipe」已坐实可行,Phase 2 学样本就是抽这套效果栈 + 参数范围。

**等用户真机打开 `flightdeck/showcase/procedural-fx/flame.aep` 复核 → 认可则 Phase 0 正式过、showcase 翻 complete、进 Phase 1;否决则记原因再议。**

### 用户真机验收结论(2026-06-18):❌ 质量不合格 — 顺序调整

用户真机看后否决:「只有火焰的形态,但和火焰差很多」。**实情**:产物=橙色噪声水滴,缺真火关键——白热芯 / 由内到外色温渐变(白→黄→橙→红→暗尖)/ 泛光(Glow)/ 向上舔的细节。**一眼可辨认是火苗轮廓,但远未达「像火焰」的产品质量。**

**关键教训(已上升到 spec)**:**手搓配方只到「可辨认」、到不了「好」——好视觉得借真实人做的样本**(验证 brainstorm 的核心判断:AI/程序化在有界族上能「认得出」,但「好看」需人造参照)。Phase 0 的**狭义**目标(证库能确定性造可辨认火焰 + 效果栈+param+动画+mask 全 gated 可行)技术上达成;但产品质量门槛未过。

**决策(改 roadmap 顺序)**:不进原 Phase 1(参数化)。**先做 Phase 2(学真实样本)再回头重做火焰**——没造出「好」火焰前参数化无意义。下一步=用户提供真实火焰 .aep(纯 AE 原生效果、非插件、非素材视频)→ 库解析抽「好火焰」的效果栈+参数 → 重做 → 再过用户关。

**carry-over 仍有效**:effect-stack recipe 机制 + AnimateEffectParam 无缺口 + 参数表,都复用;变的只是「配方内容」要照真实样本重定。
