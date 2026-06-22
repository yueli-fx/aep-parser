// flightdeck/showcase/booyah-clone/main.go — rebuild the entire "Booyah Glitch"
// project from scratch via our Go API, DAG leaf-first, the original as a read-only
// value oracle. Plan: flightdeck/plans/2026-06-19-booyah-glitch-replication.md.
//
// Comps are assembled in DAG topological order (leaves first); each gen_<comp>.go
// adds one composition. Run from repo root: `go run ./flightdeck/showcase/booyah-clone`.
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

const outPath = "flightdeck/showcase/booyah-clone/booyah-clone.aep"

// cloneFps matches the original project's frame rate. Every Booyah comp is 29.97
// (NTSC; TickRate 23976). Earlier clones used a whole 30 fps to dodge a NewComposition
// bug that wrote an inconsistent cdta time-base for fractional fps (AE read 30720 while
// we lowered keyframes at 23976). That bug is now fixed — NewComposition writes the NTSC
// legacy marker cdta @0xA8 = round(fps×100) so AE evaluates against 23976 — so the clone
// runs at the true 29.97 and its frames line up with the original frame-for-frame (the
// 30-fps approximation drifted by ~0.1%/frame, visible on a frame-by-frame diff).
const cloneFps = 29.97

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	orc, err := openOriginal()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	p := aep.NewProject(aep.TargetAE2020)

	// --- DAG topological order (leaves first) ---
	// Phase 1: create each comp's base structure (layers / shapes / text).
	buildShapeIiip(p, orc)   // ① シェイイイイプ！！！ (shape API needs no reopen)
	buildTextKomako(p, orc)  // ② テキスト変えるならココ！ (text layer + SetText)
	buildFractalMap(p, orc)  // ③ マップ用フラクタルノイズ (comp + 2 solids)
	buildKakuh(p, orc)       // ④ カクッ (comp + 2 stroked-rect+trim shape layers)
	buildPrecomp1(p, orc)    // ⑤ プリコンポジション 1 (3 precomp layers → comp ①)
	buildShapeKatamari(p, orc) // ⑥ シェイプの塊 (7 precomp layers → comp ⑤)
	buildHiraku(p, orc)        // ⑦ ここは開けない (3 precomp layers → comp ⑥ + Glow)
	buildRgbzure(p, orc)       // ⑧ RGBズレ (3 precomp layers → comp ② + Fill RGB shift)
	buildNanka(p, orc)         // ⑨ なんか周りのやつ (2 precomp ④ + 1 shape層 w/ animated Trim)
	buildGlitchText(p, orc)    // ⑩ グリッチテキスト (27-layer monster: skeleton sources)

	// Phase 2: reopen once, then apply the mutations that require a PARSED layer
	// (SetLayerTransform + AddText*Animator + AddEffect can't run on un-Reopened
	// New*-built layers).
	rp, err := aep.Reopen(p)
	must(err)
	finishTextKomako(rp, orc)  // ②
	finishFractalMap(rp, orc)  // ③
	finishKakuh(rp, orc)       // ④
	finishPrecomp1(rp, orc)    // ⑤
	finishShapeKatamari(rp, orc) // ⑥
	finishHiraku(rp, orc)        // ⑦
	finishRgbzure(rp, orc)       // ⑧
	finishNanka(rp, orc)         // ⑨
	finishGlitchText(rp, orc)    // ⑩ blend/timing/visibility (sources anchored)

	// Phase 3: reopen again so the Position/Opacity channels materialized by
	// SetLayerTransform are now parsed properties, then attach the wiggle
	// expressions the original drives them with (SetExpression needs a parsed
	// *Property). Original uses these for the RGB jitter (⑧) and a flickering
	// precomp copy (⑤ L2); the easeprobe missed them because it only inspected
	// keyframe interpolation, not expressions.
	rp2, err := aep.Reopen(rp)
	must(err)
	finishPrecomp1Expr(rp2, orc) // ⑤ L2 Opacity wiggle
	finishRgbzureExpr(rp2, orc)  // ⑧ Position X-wiggle + Opacity wiggle
	finishNankaExpr(rp2, orc)    // ⑨ Opacity wiggle (L0/L1/L2)

	f, err := os.Create(outPath)
	must(err)
	defer f.Close()
	must(rp2.WriteAEP(f))
	fmt.Println("wrote", outPath)
}
