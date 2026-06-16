// internal/aep/mg_text_style_shipgate_test.go
//
// AE render-pixel ship gate for the text styling family — upgrades SetRunFontSize
// / SetRunTracking / SetParagraphJustification / SetRunCapsOption from
// roundtrip-only to AE-render-proven (delivery-contract red line 4). SetRunLeading
// is deliberately excluded: it round-trips but AE renders default line spacing for
// a from-scratch layer (evidence-based defer — see the incident). This gate caught
// the FormatPSReal bug: AE stores the point-measurement style keys as REALS and
// reads a bare integer as 16.16 fixed-point, so SetRunFontSize used to render
// fontSize/65536 (invisible) despite a clean Go round-trip — see
// incidents/text-style-render-gate-fromscratch-blocked.md.
//
// A from-scratch text-layer Position is default-omitted (not settable), so each
// knob lives in its OWN single-comp project rendered in its own cold-start AE.
//
// CACHE GOTCHA (root-caused 2026-06-16): saveFrameToPng serves frames from AE's
// persistent on-disk cache keyed by (comp-id, render-time). The deterministic
// builder gives every single-comp project the SAME comp-id (=1), so rendering
// them all at time 0 collided on one cache entry — every PNG came out as one
// comp's frame, even across separate AE processes (the cache outlives the app),
// which silently false-greened the gate. Two defenses, both required: (1) the
// verify JSX renders each comp at a UNIQUE job.time (distinct cache key within a
// run); (2) clearAEDiskCache empties the disk cache before the loop (kills stale
// frames left by prior runs at a now-reused key). app.purge(ALL_CACHES) does NOT
// help — it clears RAM caches, not the on-disk frame cache. See
// incidents/text-style-render-gate-fromscratch-blocked.md.
//
// Go measures the rendered-text bounding box per PNG and compares the pair.
// Styling is applied after Reopen (parse-the-clone surfaces Runs[0]/Paragraphs[0]).
//
// Gated by AE_SHIP_GATE.
package aep_test

import (
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

type tsComp struct {
	name  string
	text  string
	style func(t *testing.T, l *aep.Layer)
}

func mgTextStyleComps() []tsComp {
	return []tsComp{
		{"TS_FS_SMALL", "Ag", func(t *testing.T, l *aep.Layer) { mustText(t, "size", l.SetRunFontSize(0, 50)) }},
		{"TS_FS_BIG", "Ag", func(t *testing.T, l *aep.Layer) { mustText(t, "size", l.SetRunFontSize(0, 160)) }},
		{"TS_TRK_TIGHT", "MMMM", func(t *testing.T, l *aep.Layer) { mustText(t, "trk", l.SetRunTracking(0, 0)) }},
		{"TS_TRK_WIDE", "MMMM", func(t *testing.T, l *aep.Layer) { mustText(t, "trk", l.SetRunTracking(0, 1200)) }},
		{"TS_JL", "ABCD", func(t *testing.T, l *aep.Layer) { mustText(t, "j", l.SetParagraphJustification(0, aep.TextJustifyLeft)) }},
		{"TS_JC", "ABCD", func(t *testing.T, l *aep.Layer) { mustText(t, "j", l.SetParagraphJustification(0, aep.TextJustifyCenter)) }},
		// Leading is intentionally NOT render-gated: SetRunLeading round-trips
		// (DOM reads back 70/220) but AE renders default line spacing for a
		// from-scratch text layer regardless (220 vs 70 both render the two
		// lines tightly — verified visually). Evidence-based defer, kept
		// verify=roundtrip; see text-style-render-gate-fromscratch-blocked.md.
		// "ace" is all x-height (no ascenders/descenders), so all-caps "ACE"
		// renders unambiguously taller than lowercase "ace".
		{"TS_CAPS_N", "ace", func(t *testing.T, l *aep.Layer) {}},
		{"TS_CAPS_A", "ace", func(t *testing.T, l *aep.Layer) { mustText(t, "caps", l.SetRunCapsOption(0, aep.TextCapsAll)) }},
	}
}

// buildOneTextStyleProject builds a project containing a SINGLE comp (white BG
// shape + one styled text layer). Each knob gets its own project because in
// headless `AfterFX -r` mode comp.saveFrameToPng renders the project's active
// comp regardless of the receiver and the viewer can't be switched — so the only
// way to render a specific comp deterministically is to make it the sole comp.
func buildOneTextStyleProject(t *testing.T, target aep.AETarget, c tsComp, idx int) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)

	// Distinct comp height per knob. AE's render cache key is too coarse to tell
	// these near-identical single-comp projects apart (same 1920x1080, same
	// BG+text structure), so renders collided — every PNG came out as one comp's
	// frame, even across fresh AE processes (persistent disk cache). A unique
	// height makes each comp structurally distinct (like separate showcases,
	// which never collide). Height doesn't affect the glyph bounding box or the
	// horizontal centroid the gate measures; width stays 1920 so justify cx is
	// comparable across comps.
	comp, err := aep.NewComposition(p, c.name, 1920, uint16(1080+idx), 30, 5)
	if err != nil {
		t.Fatalf("NewComposition %s: %v", c.name, err)
	}
	bg, err := aep.NewShapeLayer(comp, "BG")
	if err != nil {
		t.Fatalf("%s BG: %v", c.name, err)
	}
	br, err := bg.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("%s BG rect: %v", c.name, err)
	}
	if err := br.SetSize([2]float64{2200, 1300}); err != nil {
		t.Fatalf("%s BG size: %v", c.name, err)
	}
	bf, err := bg.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("%s BG fill: %v", c.name, err)
	}
	if err := bf.SetColor([4]float64{1, 1, 1, 1}); err != nil { // white BG; default text reads against it
		t.Fatalf("%s BG color: %v", c.name, err)
	}
	if err := bg.Position().SetStaticValue([2]float64{960, 540}); err != nil {
		t.Fatalf("%s BG pos: %v", c.name, err)
	}
	l, err := aep.NewTextLayer(comp, "T")
	if err != nil {
		t.Fatalf("%s NewTextLayer: %v", c.name, err)
	}
	if err := l.SetText(c.text); err != nil {
		t.Fatalf("%s SetText: %v", c.name, err)
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen %s: %v", c.name, err)
	}
	rcomp := rp.Compositions[0]
	if bg := rcomp.LayerByName("BG"); bg != nil {
		mustText(t, "BG MoveToEnd", aep.MoveToEnd(bg)) // text on top
	}
	rl := rcomp.LayerByName("T")
	if rl == nil {
		t.Fatalf("reopened %s layer T missing", c.name)
	}
	c.style(t, rl)
	return rp
}

func mustText(t *testing.T, label string, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", label, err)
	}
}

// clearAEDiskCache empties AE's persistent disk frame caches before the render
// gate. saveFrameToPng serves cached frames keyed by (comp-id, time); since the
// builder gives every single-comp project the same comp-id, a stale frame from a
// prior run at a reused (id, time) key is served instead of a fresh render —
// the cross-contamination that false-greened this gate. The default disk-cache
// location is %LOCALAPPDATA%\Temp\Adobe\After Effects\<ver>\Disk Cache*.noindex;
// clearing it is exactly what the "Empty Disk Cache" pref button does and the
// cache regenerates on next render. See
// incidents/text-style-render-gate-fromscratch-blocked.md.
func clearAEDiskCache(t *testing.T) {
	t.Helper()
	base := filepath.Join(os.Getenv("LOCALAPPDATA"), "Temp", "Adobe", "After Effects")
	matches, _ := filepath.Glob(filepath.Join(base, "*", "Disk Cache*.noindex"))
	for _, m := range matches {
		if err := os.RemoveAll(m); err != nil {
			t.Logf("clearAEDiskCache: %v (continuing)", err)
			continue
		}
		t.Logf("clearAEDiskCache: emptied %s", m)
	}
}

// inkBoxFull returns the bounding box of "ink" (non-background) pixels over the
// whole image. The BG is a white shape; a from-scratch NewTextLayer renders in
// AE's default fill (observed light-blue on this build, not dark), so the
// detector is color-agnostic: any pixel that deviates enough from white counts
// as ink. (The earlier dark-only threshold scored light-blue text as background
// and reported every box as 0x0 — see text-style-render-gate-fromscratch-blocked.)
func inkBoxFull(img image.Image) (w, h, cx, cy, count int) {
	b := img.Bounds()
	minX, minY := 1<<30, 1<<30
	maxX, maxY := -1, -1
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bb, _ := img.At(x, y).RGBA()
			r8, g8, b8 := int(r>>8), int(g>>8), int(bb>>8)
			if (255-r8)+(255-g8)+(255-b8) > 60 { // deviates from white BG = ink
				count++
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}
	if count == 0 {
		return 0, 0, 0, 0, 0
	}
	return maxX - minX, maxY - minY, (minX + maxX) / 2, (minY + maxY) / 2, count
}

func runMGTextStyleGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	clearAEDiskCache(t)
	const argsPath = `e:/projects/tools/aep-parser/test_data/mg_text_style_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_mg_text_style.jsx`
	pngDir := `e:/projects/tools/aep-parser/tmp_debug`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	tempDir := t.TempDir()
	defer os.Remove(argsPath)

	// One FRESH AE launch per knob. saveFrameToPng races AE's async renderer
	// within a single -r session: rendering multiple single-comp projects via
	// sequential app.open returned stale / cross-contaminated frames (every PNG
	// came out as one comp; neither app.purge nor unique render-times made it
	// deterministic — there is no render-completion signal to wait on). A fresh
	// AE with exactly one comp has nothing to race, so the render is reliable.
	// Slower (one cold start per knob) but deterministic.
	comps := mgTextStyleComps()
	for idx, c := range comps {
		p := buildOneTextStyleProject(t, target, c, idx)
		inputAEP := filepath.Join(tempDir, c.name+".aep")
		out, err := os.Create(inputAEP)
		if err != nil {
			t.Fatal(err)
		}
		if err := p.WriteAEP(out); err != nil {
			out.Close()
			t.Fatalf("WriteAEP %s: %v", c.name, err)
		}
		out.Close()

		pngPath := fmt.Sprintf("%s/mg_text_style_%s_%s.png", toFwd(pngDir), c.name, ver)
		os.Remove(filepath.FromSlash(pngPath))

		// Distinct render time per comp. AE serves a frame cache keyed by
		// (comp-uid, time) that persists even across separate AE processes; every
		// deterministically-built single-comp project reuses the same comp-uid, so
		// rendering them all at the same time returned ONE cached frame for all.
		// A well-separated time per comp (still static text, identical-looking
		// frame) gives each a unique cache key. The text is static so the frame is
		// the same at any time; 5s comp duration leaves ample room.
		renderTime := 0.1 + float64(idx)*0.5

		doneC := filepath.Join(tempDir, c.name+".done")
		argsJSON := fmt.Sprintf(`{"done":%q,"ver":%q,"jobs":[{"name":%q,"input":%q,"png":%q,"time":%g}]}`,
			toFwd(doneC), ver, c.name, toFwd(inputAEP), pngPath, renderTime)
		if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
			t.Fatal(err)
		}
		os.Remove(doneC)

		runAeRunShipGate(t, aeExe, jsxPath, doneC, 150)

		content, err := os.ReadFile(doneC)
		if err != nil {
			t.Fatalf("%s %s done: %v", ver, c.name, err)
		}
		body := string(content)
		t.Logf("%s %s AE readback: %s", ver, c.name, strings.TrimSpace(body))
		if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
			t.Errorf("%s %s render FAIL: %s", ver, c.name, body)
		}
	}

	box := func(comp string) (w, h, cx, cnt int) {
		path := fmt.Sprintf("%s/mg_text_style_%s_%s.png", pngDir, comp, ver)
		f, err := os.Open(path)
		if err != nil {
			t.Errorf("%s %s png missing: %v", ver, comp, err)
			return
		}
		defer f.Close()
		im, _, err := image.Decode(f)
		if err != nil {
			t.Errorf("%s %s decode: %v", ver, comp, err)
			return
		}
		w, h, cx, _, cnt = inkBoxFull(im)
		return
	}

	smW, smH, _, _ := box("TS_FS_SMALL")
	bgW, bgH, _, _ := box("TS_FS_BIG")
	tiW, _, _, _ := box("TS_TRK_TIGHT")
	wiW, _, _, _ := box("TS_TRK_WIDE")
	_, _, jlCx, _ := box("TS_JL")
	_, _, jcCx, _ := box("TS_JC")
	cnW, cnH, _, _ := box("TS_CAPS_N")
	caW, caH, _, _ := box("TS_CAPS_A")

	t.Logf("%s FontSize  SMALL %dx%d  BIG %dx%d", ver, smW, smH, bgW, bgH)
	t.Logf("%s Tracking  TIGHT w=%d  WIDE w=%d", ver, tiW, wiW)
	t.Logf("%s Justify   JLEFT cx=%d  JCENTER cx=%d", ver, jlCx, jcCx)
	t.Logf("%s Caps      NORM %dx%d  ALL %dx%d", ver, cnW, cnH, caW, caH)

	// FontSize: BIG (160pt) glyphs are far taller than SMALL (50pt).
	if bgH < smH*2 {
		t.Errorf("%s BIG h=%d not >= 2x SMALL h=%d — SetRunFontSize did not enlarge", ver, bgH, smH)
	}
	// Tracking: WIDE (1200) spreads the same MMMM far wider than TIGHT (0).
	if wiW < tiW+150 {
		t.Errorf("%s WIDE w=%d not >> TIGHT w=%d — SetRunTracking did not widen", ver, wiW, tiW)
	}
	// Justification: same default anchor; left-justified text sits right of it,
	// centre-justified straddles it, so JLEFT's box centre is well right of JCENTER's.
	if jlCx-jcCx < 40 {
		t.Errorf("%s justify: JLEFT cx=%d JCENTER cx=%d — left should sit right of centred", ver, jlCx, jcCx)
	}
	// Caps: all-caps renders "ace" as "ACE" — capitals are taller than the
	// lowercase x-height ("ace" has no ascenders/descenders, so the gain is
	// unambiguous).
	if caH < cnH+15 {
		t.Errorf("%s CAPSALL h=%d not > CAPSNORM h=%d+15 — SetRunCapsOption all-caps did not render taller", ver, caH, cnH)
	}
}

func TestMGTextStyle_AEShipGate_AE2020(t *testing.T) {
	runMGTextStyleGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGTextStyle_AEShipGate_AE2025(t *testing.T) {
	runMGTextStyleGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
