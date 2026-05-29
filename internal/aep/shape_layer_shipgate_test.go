// internal/aep/shape_layer_shipgate_test.go
//
// V2.2 Phase 5 Task 5.2 — AE ship gate (mirrors V2.1 runAEShipGate).
//
// Driver Go side:
//   1. Build canonical 3-ShapeLayer project (spec §5.2) + WriteAEP → tempDir
//   2. Write args.json to fixed path verify_v2_2.jsx reads
//   3. scripts/ae_run.ps1 dispatches AE + auto-dismisses modals
//   4. Assert .done first line == "PASS"
//
// Gated by AE_SHIP_GATE env var; CI / no-AE runs SKIP cleanly.
package aep_test

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/rifx"
)

func runV2_2ShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "v2_2_canonical.aep")
	resavedAEP := filepath.Join(tempDir, "v2_2_canonical.resaved.aep")
	doneFile := filepath.Join(tempDir, "v2_2_test.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/v2_2_args.json`

	// 1. Build canonical shape graph (spec §5.2) — must match the layer
	//    names and values verify_v2_2.jsx is checking against.
	p := aep.NewProject(target)
	comp, err := p.NewComposition("Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	a, _ := comp.NewShapeLayer("A_RectFill_Animated")
	rectA, _ := a.RootGroup().AddRect()
	_ = rectA.Size().AddKeyframeLinear(0, [2]float64{50, 50})
	_ = rectA.Size().AddKeyframeLinear(2, [2]float64{300, 200})
	fillA, _ := a.RootGroup().AddFill()
	_ = fillA.Color().AddKeyframeLinear(0, [4]float64{1, 0, 0, 1})
	_ = fillA.Color().AddKeyframeLinear(2, [4]float64{0, 0, 1, 1})
	_ = a.Position().AddKeyframeLinear(0, [2]float64{0, 0})
	_ = a.Position().AddKeyframeLinear(2, [2]float64{500, 300})

	b, _ := comp.NewShapeLayer("B_EllipseStroke_Static")
	ellB, _ := b.RootGroup().AddEllipse()
	_ = ellB.SetSize([2]float64{150, 150})
	strokeB, _ := b.RootGroup().AddStroke()
	_ = strokeB.SetColor([4]float64{0, 0, 1, 1})
	_ = strokeB.SetWidth(5)

	c, _ := comp.NewShapeLayer("C_PathFillStroke_Static")
	pathC, _ := c.RootGroup().AddPath()
	_ = pathC.SetVertices([][2]float64{{0, 0}, {100, 0}, {100, 100}, {0, 100}})
	_ = pathC.SetClosed(true)
	fillC, _ := c.RootGroup().AddFill()
	_ = fillC.SetColor([4]float64{0, 1, 0, 1})
	strokeC, _ := c.RootGroup().AddStroke()
	_ = strokeC.SetWidth(2)

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	// 2. args.json — forward-slashed paths for AE
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP))
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	// 3. ae_run.ps1 (drop-in for AfterFX -r — handles convert / save-changes / data-loss modals)
	jsxPath := `E:/projects/tools/aep-parser/test_data/verify_v2_2.jsx`
	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	// 4. assert PASS (.done guaranteed to exist after ps1 exit 0)
	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.SplitN(string(content), "\n", 2)
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("ship gate FAIL:\n%s", string(content))
	}
}

// The 3-layer canonical ship-gate exercises Path + Stroke, still deferred in
// V2.2.1 (Ellipse shipped separately — see TestV2_2_Ellipse_AEShipGate_*).
// Skip until per-kind extract+embed bytes land for Path/Stroke.
func TestV2_2_AEShipGate_AE2025(t *testing.T) {
	t.Skip("Path/Stroke embed bytes still deferred (Ellipse ships via TestV2_2_Ellipse_AEShipGate_*)")
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2ShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_AEShipGate_AE2020(t *testing.T) {
	t.Skip("Path/Stroke embed bytes still deferred (Ellipse ships via TestV2_2_Ellipse_AEShipGate_*)")
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2ShipGate(t, aep.TargetAE2020, aeExe)
}

// runV2_2EllipseShipGate is the V2.2.1 focused gate: one ShapeLayer with an
// Ellipse + Fill. Asserts AE accepts the Ellipse (no silent-drop) and
// re-derives the Size/Position values Go injected via embed+overwrite. Values
// differ from the tolerance fixture so a no-op overwrite is caught.
func runV2_2EllipseShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "v2_2_ellipse.aep")
	resavedAEP := filepath.Join(tempDir, "v2_2_ellipse.resaved.aep")
	doneFile := filepath.Join(tempDir, "v2_2_ellipse.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/v2_2_ellipse_args.json`

	p := aep.NewProject(target)
	comp, err := p.NewComposition("Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, err := comp.NewShapeLayer("Ellipse_Static")
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	ell, err := l.RootGroup().AddEllipse()
	if err != nil {
		t.Fatalf("AddEllipse: %v", err)
	}
	_ = ell.SetSize([2]float64{260, 140})
	_ = ell.SetPosition([2]float64{70, 90})
	fill, err := l.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("AddFill: %v", err)
	}
	_ = fill.SetColor([4]float64{0.5, 0.5, 0.5, 1})

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP))
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	jsxPath := `E:/projects/tools/aep-parser/test_data/verify_v2_2_ellipse.jsx`
	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.SplitN(string(content), "\n", 2)
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("ellipse ship gate FAIL:\n%s", string(content))
	}

	// Prove AE retained the overwritten Size/Position by parsing the file AE
	// re-saved (sidesteps the .value ExtendScript quirk). AE writes the value
	// at cdat offset 0 (BE f64), same as the native tolerance fixture.
	assertResavedEllipse(t, resavedAEP, [2]float64{260, 140}, [2]float64{70, 90})

	// Fill color round-trips with the V2.2.1 ARGB×255 encoding (fill set to
	// [0.5,0.5,0.5,1] → ARGB×255 [255,127.5,127.5,127.5]).
	root := parseAEP(t, resavedAEP)
	if fc := streamCdat(root, "ADBE Vector Fill Color"); len(fc) >= 32 {
		rd := func(off int) float64 { return math.Float64frombits(binary.BigEndian.Uint64(fc[off : off+8])) }
		want := []float64{255, 127.5, 127.5, 127.5}
		for i, w := range want {
			if got := rd(i * 8); math.Abs(got-w) > 1.0 {
				t.Errorf("resaved fill color[%d] = %.4g, want %.4g (ARGB×255)", i, got, w)
			}
		}
	} else {
		t.Errorf("resaved fill color cdat missing/short")
	}
}

// assertResavedEllipse parses an AE-resaved .aep, locates the Ellipse Size /
// Position cdat, and asserts the leading f64 BE values match want.
func assertResavedEllipse(t *testing.T, path string, wantSize, wantPos [2]float64) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open resaved: %v", err)
	}
	defer f.Close()
	root, err := rifx.Parse(f)
	if err != nil {
		t.Fatalf("parse resaved: %v", err)
	}
	gotSize := readShapeStreamVec2(t, root, "ADBE Vector Ellipse Size")
	gotPos := readShapeStreamVec2(t, root, "ADBE Vector Ellipse Position")
	if math.Abs(gotSize[0]-wantSize[0]) > 0.5 || math.Abs(gotSize[1]-wantSize[1]) > 0.5 {
		t.Errorf("resaved Ellipse Size = %v, want %v", gotSize, wantSize)
	}
	if math.Abs(gotPos[0]-wantPos[0]) > 0.5 || math.Abs(gotPos[1]-wantPos[1]) > 0.5 {
		t.Errorf("resaved Ellipse Position = %v, want %v", gotPos, wantPos)
	}
}

// readShapeStreamVec2 finds the tdmn matching name anywhere in the tree, reads
// the following tdbs cdat, and returns its first two f64 BE values.
func readShapeStreamVec2(t *testing.T, root *rifx.Chunk, name string) [2]float64 {
	t.Helper()
	var out [2]float64
	var found bool
	var walk func(*rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		if found {
			return
		}
		for i := 0; i < len(c.Children); i++ {
			ch := c.Children[i]
			if ch.ID == rifx.IDTdmn && i+1 < len(c.Children) && trimShipNUL(string(ch.Data)) == name {
				tdbs := c.Children[i+1]
				if tdbs.IsList() && tdbs.FormType == rifx.IDTdbs {
					for _, cc := range tdbs.Children {
						if cc.ID == rifx.IDCdat && len(cc.Data) >= 16 {
							out[0] = math.Float64frombits(binary.BigEndian.Uint64(cc.Data[0:8]))
							out[1] = math.Float64frombits(binary.BigEndian.Uint64(cc.Data[8:16]))
							found = true
							return
						}
					}
				}
			}
			if ch.IsList() {
				walk(ch)
				if found {
					return
				}
			}
		}
	}
	walk(root)
	if !found {
		t.Fatalf("%s: not found in resaved aep", name)
	}
	return out
}

func trimShipNUL(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			return s[:i]
		}
	}
	return s
}

func TestV2_2_Ellipse_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2EllipseShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_Ellipse_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2EllipseShipGate(t, aep.TargetAE2020, aeExe)
}

// runV2_2PathShipGate is the V2.2.1 focused Path gate: one ShapeLayer with a
// Path (DISTINCT triangle, not the embed's square, so the geometry splice is
// actually exercised) + Fill. Asserts AE accepts it (from-scratch path CRASHED
// AE 2020) and the re-saved ldat anchors round-trip to the input vertices.
func runV2_2PathShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "v2_2_path.aep")
	resavedAEP := filepath.Join(tempDir, "v2_2_path.resaved.aep")
	doneFile := filepath.Join(tempDir, "v2_2_path.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/v2_2_path_args.json`

	verts := [][2]float64{{10, 20}, {70, 30}, {40, 90}}

	p := aep.NewProject(target)
	comp, err := p.NewComposition("Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, err := comp.NewShapeLayer("Path_Static")
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	pth, err := l.RootGroup().AddPath()
	if err != nil {
		t.Fatalf("AddPath: %v", err)
	}
	if err := pth.SetVertices(verts); err != nil {
		t.Fatalf("SetVertices: %v", err)
	}
	_ = pth.SetClosed(true)
	fill, err := l.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("AddFill: %v", err)
	}
	_ = fill.SetColor([4]float64{0.5, 0.5, 0.5, 1})

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP))
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	jsxPath := `E:/projects/tools/aep-parser/test_data/verify_v2_2_path.jsx`
	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.SplitN(string(content), "\n", 2)
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("path ship gate FAIL:\n%s", string(content))
	}

	assertResavedPathAnchors(t, resavedAEP, verts)
}

// assertResavedPathAnchors parses an AE-resaved .aep, decodes the path shph
// bbox + ldat vertex anchors (slot 0 of each 6-f32 vertex block, bbox-
// de-normalized), and asserts they match the input vertices — proving AE
// retained the geometry the splice injected.
func assertResavedPathAnchors(t *testing.T, path string, want [][2]float64) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open resaved: %v", err)
	}
	defer f.Close()
	root, err := rifx.Parse(f)
	if err != nil {
		t.Fatalf("parse resaved: %v", err)
	}
	shph := findShipChunk(root, rifx.IDShph)
	ldat := findShipChunk(root, rifx.IDLdat)
	if shph == nil || ldat == nil {
		t.Fatalf("resaved path: shph=%v ldat=%v (one missing → path dropped)", shph != nil, ldat != nil)
	}
	rf := func(b []byte, off int) float64 {
		return float64(math.Float32frombits(binary.BigEndian.Uint32(b[off : off+4])))
	}
	minX, minY := rf(shph.Data, 4), rf(shph.Data, 8)
	maxX, maxY := rf(shph.Data, 12), rf(shph.Data, 16)
	rx, ry := maxX-minX, maxY-minY
	if len(ldat.Data) < len(want)*24 {
		t.Fatalf("resaved ldat too short: %d B for %d verts", len(ldat.Data), len(want))
	}
	for i, w := range want {
		ax := rf(ldat.Data, i*24)*rx + minX
		ay := rf(ldat.Data, i*24+4)*ry + minY
		if math.Abs(ax-w[0]) > 0.5 || math.Abs(ay-w[1]) > 0.5 {
			t.Errorf("resaved vertex %d anchor = (%.3g,%.3g), want (%.3g,%.3g)", i, ax, ay, w[0], w[1])
		}
	}
}

func findShipChunk(root *rifx.Chunk, id rifx.ChunkID) *rifx.Chunk {
	for _, ch := range root.Children {
		if ch.ID == id {
			return ch
		}
		if ch.IsList() {
			if g := findShipChunk(ch, id); g != nil {
				return g
			}
		}
	}
	return nil
}

func TestV2_2_Path_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2PathShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_Path_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2PathShipGate(t, aep.TargetAE2020, aeExe)
}

// runV2_2StrokeShipGate is the V2.2.1 focused Stroke gate: one ShapeLayer with
// a Rect (geometry) + Stroke. Asserts AE accepts the stroke (no silent-drop)
// and the re-saved Color/Width/Opacity cdat round-trip. Values differ from the
// embed fixture so the cdat overwrite is genuinely exercised. Color uses AE's
// [A,R,G,B]×255 f64 encoding (RE'd from the stroke tolerance fixture).
func runV2_2StrokeShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "v2_2_stroke.aep")
	resavedAEP := filepath.Join(tempDir, "v2_2_stroke.resaved.aep")
	doneFile := filepath.Join(tempDir, "v2_2_stroke.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/v2_2_stroke_args.json`

	color := [4]float64{1, 0, 0, 1} // red, distinct from embed fixture's [0,0,1,1]
	width, opacity := 4.0, 60.0

	p := aep.NewProject(target)
	comp, err := p.NewComposition("Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, err := comp.NewShapeLayer("Stroke_Static")
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	r, err := l.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect: %v", err)
	}
	_ = r.SetSize([2]float64{200, 100})
	stroke, err := l.RootGroup().AddStroke()
	if err != nil {
		t.Fatalf("AddStroke: %v", err)
	}
	_ = stroke.SetColor(color)
	_ = stroke.SetWidth(width)
	_ = stroke.SetOpacity(opacity)

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP))
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	jsxPath := `E:/projects/tools/aep-parser/test_data/verify_v2_2_stroke.jsx`
	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.SplitN(string(content), "\n", 2)
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("stroke ship gate FAIL:\n%s", string(content))
	}

	// Decode re-saved cdats. Color = [A,R,G,B]×255 f64; Width/Opacity = f64.
	root := parseAEP(t, resavedAEP)
	colCdat := streamCdat(root, "ADBE Vector Stroke Color")
	wCdat := streamCdat(root, "ADBE Vector Stroke Width")
	opCdat := streamCdat(root, "ADBE Vector Stroke Opacity")
	if colCdat == nil || wCdat == nil || opCdat == nil {
		t.Fatalf("resaved stroke cdat missing (color=%v width=%v op=%v) — stroke dropped?",
			colCdat != nil, wCdat != nil, opCdat != nil)
	}
	rdF64 := func(b []byte, off int) float64 {
		return math.Float64frombits(binary.BigEndian.Uint64(b[off : off+8]))
	}
	// Color stored ARGB×255: [0]=A*255 [8]=R*255 [16]=G*255 [24]=B*255.
	wantARGB := []float64{color[3] * 255, color[0] * 255, color[1] * 255, color[2] * 255}
	for i, w := range wantARGB {
		if got := rdF64(colCdat, i*8); math.Abs(got-w) > 1.0 {
			t.Errorf("resaved stroke color[%d] = %.3g, want %.3g (ARGB×255)", i, got, w)
		}
	}
	if got := rdF64(wCdat, 0); math.Abs(got-width) > 0.01 {
		t.Errorf("resaved stroke width = %.3g, want %.3g", got, width)
	}
	if got := rdF64(opCdat, 0); math.Abs(got-opacity) > 0.01 {
		t.Errorf("resaved stroke opacity = %.3g, want %.3g", got, opacity)
	}
}

func parseAEP(t *testing.T, path string) *rifx.Chunk {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	root, err := rifx.Parse(f)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return root
}

// streamCdat finds the tdmn matching name anywhere in the tree and returns the
// cdat inside the following LIST(tdbs), or nil.
func streamCdat(root *rifx.Chunk, name string) []byte {
	var found []byte
	var walk func(*rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		for i := 0; i < len(c.Children); i++ {
			ch := c.Children[i]
			if ch.ID == rifx.IDTdmn && i+1 < len(c.Children) && trimShipNUL(string(ch.Data)) == name {
				tdbs := c.Children[i+1]
				if tdbs.IsList() && tdbs.FormType == rifx.IDTdbs {
					for _, cc := range tdbs.Children {
						if cc.ID == rifx.IDCdat {
							found = cc.Data
							return
						}
					}
				}
			}
			if ch.IsList() {
				walk(ch)
				if found != nil {
					return
				}
			}
		}
	}
	walk(root)
	return found
}

func TestV2_2_Stroke_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2StrokeShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_Stroke_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2StrokeShipGate(t, aep.TargetAE2020, aeExe)
}

// runV2_2RectKfShipGate is the V2.2.1 focused keyframe-persistence gate: a
// ShapeLayer with an animated Rect Size (2 linear keyframes) + Fill. Asserts
// AE accepts it and the re-saved Rect Size keyframe container (lhd3 numKf +
// ldat values) round-trips — proving shape sub-stream keyframes persist (the
// non-spatial Vec2 ldat is byte-identical to AE's own encoding).
func runV2_2RectKfShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "v2_2_rectkf.aep")
	resavedAEP := filepath.Join(tempDir, "v2_2_rectkf.resaved.aep")
	doneFile := filepath.Join(tempDir, "v2_2_rectkf.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/v2_2_rectkf_args.json`

	p := aep.NewProject(target)
	comp, err := p.NewComposition("Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, err := comp.NewShapeLayer("RectKf")
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	r, err := l.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect: %v", err)
	}
	_ = r.Size().AddKeyframeLinear(0, [2]float64{50, 50})
	_ = r.Size().AddKeyframeLinear(2, [2]float64{300, 200})
	fill, err := l.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("AddFill: %v", err)
	}
	_ = fill.SetColor([4]float64{0.5, 0.5, 0.5, 1})

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP))
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	jsxPath := `E:/projects/tools/aep-parser/test_data/verify_v2_2_rectkf.jsx`
	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.SplitN(string(content), "\n", 2)
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("rect-kf ship gate FAIL:\n%s", string(content))
	}

	// Decode the re-saved Rect Size keyframe container.
	root := parseAEP(t, resavedAEP)
	kfl := findShipList(root, "ADBE Vector Rect Size")
	if kfl == nil {
		t.Fatalf("resaved Rect Size has no LIST(list) keyframe container — keyframes dropped")
	}
	lhd3 := findShipChunk(kfl, rifx.IDLhd3)
	ldat := findShipChunk(kfl, rifx.IDLdat)
	if lhd3 == nil || ldat == nil {
		t.Fatalf("resaved Rect Size kf container missing lhd3/ldat")
	}
	numKf := binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C])
	bpk := int(binary.BigEndian.Uint32(lhd3.Data[0x10:0x14]))
	if numKf != 2 {
		t.Errorf("resaved Rect Size numKf = %d, want 2", numKf)
	}
	// Non-spatial Vec2: value at block+0x08 (2 f64). Check kf0=[50,50], kf1=[300,200].
	rdF64 := func(off int) float64 { return math.Float64frombits(binary.BigEndian.Uint64(ldat.Data[off : off+8])) }
	if got0x, got0y := rdF64(0x08), rdF64(0x10); math.Abs(got0x-50) > 0.5 || math.Abs(got0y-50) > 0.5 {
		t.Errorf("resaved Rect Size kf0 = (%.3g,%.3g), want (50,50)", got0x, got0y)
	}
	if got1x, got1y := rdF64(bpk+0x08), rdF64(bpk+0x10); math.Abs(got1x-300) > 0.5 || math.Abs(got1y-200) > 0.5 {
		t.Errorf("resaved Rect Size kf1 = (%.3g,%.3g), want (300,200)", got1x, got1y)
	}
}

// findShipList finds the tdmn matching name, then returns the LIST(list/kfl)
// keyframe container inside the following LIST(tdbs), or nil (static stream).
func findShipList(root *rifx.Chunk, name string) *rifx.Chunk {
	var found *rifx.Chunk
	var walk func(*rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		for i := 0; i+1 < len(c.Children); i++ {
			ch := c.Children[i]
			if ch.ID == rifx.IDTdmn && trimShipNUL(string(ch.Data)) == name {
				tdbs := c.Children[i+1]
				if tdbs.IsList() && tdbs.FormType == rifx.IDTdbs {
					for _, cc := range tdbs.Children {
						if cc.IsList() && cc.FormType == rifx.IDkfl {
							found = cc
							return
						}
					}
				}
			}
			if ch.IsList() {
				walk(ch)
				if found != nil {
					return
				}
			}
		}
	}
	walk(root)
	return found
}

func TestV2_2_RectKf_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2RectKfShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_RectKf_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2RectKfShipGate(t, aep.TargetAE2020, aeExe)
}

// runV2_2FillKfShipGate is the V2.2.1 Color-keyframe gate: a ShapeLayer with a
// Rect + Fill whose Color is animated (2 linear keyframes). Color keyframes use
// the spatial-style ldat block (value@0x38, bpk 152) with [A,R,G,B]×255 values
// — byte-identical to AE's own encoding (verified offline). Confirms AE accepts
// the color-keyframe path end-to-end + the re-saved keyframes round-trip.
func runV2_2FillKfShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "v2_2_fillkf.aep")
	resavedAEP := filepath.Join(tempDir, "v2_2_fillkf.resaved.aep")
	doneFile := filepath.Join(tempDir, "v2_2_fillkf.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/v2_2_fillkf_args.json`

	p := aep.NewProject(target)
	comp, err := p.NewComposition("Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, _ := comp.NewShapeLayer("FillKf")
	r, _ := l.RootGroup().AddRect()
	_ = r.SetSize([2]float64{200, 100})
	fill, _ := l.RootGroup().AddFill()
	_ = fill.Color().AddKeyframeLinear(0, [4]float64{1, 0, 0, 1})
	_ = fill.Color().AddKeyframeLinear(2, [4]float64{0, 0, 1, 1})

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q}`, toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP))
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	runAeRunShipGate(t, aeExe, `E:/projects/tools/aep-parser/test_data/verify_v2_2_fillkf.jsx`, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	if lines := strings.SplitN(string(content), "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("fill-kf ship gate FAIL:\n%s", string(content))
	}

	root := parseAEP(t, resavedAEP)
	kfl := findShipList(root, "ADBE Vector Fill Color")
	if kfl == nil {
		t.Fatalf("resaved Fill Color has no keyframe container — color keyframes dropped")
	}
	lhd3 := findShipChunk(kfl, rifx.IDLhd3)
	if lhd3 == nil || binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C]) != 2 {
		t.Errorf("resaved Fill Color numKf != 2")
	}
	// Color block is spatial-style: value (ARGB×255) at block+0x38.
	ldat := findShipChunk(kfl, rifx.IDLdat)
	if ldat != nil && len(ldat.Data) >= 0x38+32 {
		a := math.Float64frombits(binary.BigEndian.Uint64(ldat.Data[0x38:0x40]))
		rr := math.Float64frombits(binary.BigEndian.Uint64(ldat.Data[0x40:0x48]))
		if math.Abs(a-255) > 1 || math.Abs(rr-255) > 1 { // kf0 [1,0,0,1] → A=255,R=255
			t.Errorf("resaved Fill Color kf0 ARGB = (%.3g,%.3g,..), want A=255,R=255", a, rr)
		}
	}
}

func TestV2_2_FillKf_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2FillKfShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_FillKf_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2FillKfShipGate(t, aep.TargetAE2020, aeExe)
}

// runV2_2EllKfShipGate: Ellipse with animated Size (non-spatial Vec2) +
// Position (spatial Vec2 motion-path, bpk 104). Position ldat differs from AE
// only in the auto-computed ~0 spatial tangent (AE recomputes on load); the
// gate confirms AE accepts it and round-trips the keyframe VALUES.
func runV2_2EllKfShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "v2_2_ellkf.aep")
	resavedAEP := filepath.Join(tempDir, "v2_2_ellkf.resaved.aep")
	doneFile := filepath.Join(tempDir, "v2_2_ellkf.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/v2_2_ellkf_args.json`

	p := aep.NewProject(target)
	comp, _ := p.NewComposition("Main", 1920, 1080, 30, 5)
	l, _ := comp.NewShapeLayer("EllAnim")
	ell, _ := l.RootGroup().AddEllipse()
	_ = ell.Size().AddKeyframeLinear(0, [2]float64{50, 50})
	_ = ell.Size().AddKeyframeLinear(2, [2]float64{300, 200})
	_ = ell.Position().AddKeyframeLinear(0, [2]float64{10, 20})
	_ = ell.Position().AddKeyframeLinear(2, [2]float64{70, 90})

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q}`, toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP))
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	runAeRunShipGate(t, aeExe, `E:/projects/tools/aep-parser/test_data/verify_v2_2_ellkf.jsx`, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	if lines := strings.SplitN(string(content), "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("ell-kf ship gate FAIL:\n%s", string(content))
	}

	root := parseAEP(t, resavedAEP)
	// Position keyframes survive: spatial block value at 0x38 (kf0 [10,20]).
	if kfl := findShipList(root, "ADBE Vector Ellipse Position"); kfl != nil {
		if lhd3 := findShipChunk(kfl, rifx.IDLhd3); lhd3 == nil || binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C]) != 2 {
			t.Errorf("resaved Ellipse Position numKf != 2")
		}
		if ldat := findShipChunk(kfl, rifx.IDLdat); ldat != nil && len(ldat.Data) >= 0x48 {
			x := math.Float64frombits(binary.BigEndian.Uint64(ldat.Data[0x38:0x40]))
			y := math.Float64frombits(binary.BigEndian.Uint64(ldat.Data[0x40:0x48]))
			if math.Abs(x-10) > 0.5 || math.Abs(y-20) > 0.5 {
				t.Errorf("resaved Ellipse Position kf0 = (%.3g,%.3g), want (10,20)", x, y)
			}
		}
	} else {
		t.Errorf("resaved Ellipse Position keyframes dropped")
	}
}

func TestV2_2_EllKf_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2EllKfShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_EllKf_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2EllKfShipGate(t, aep.TargetAE2020, aeExe)
}

// runV2_2LayrPosKfShipGate is the V2.2.1 Layr Transform Position keyframe gate
// (Path B): a ShapeLayer whose combined Transform Position ("ADBE Position") is
// animated (2 linear keyframes). Persisted as a bpk-128 spatial dim-3 block
// (value@0x38 X/Y/Z, Z=0; motion-path marker@0x08) injected into the embedded
// transform-group template. Confirms AE accepts the path end-to-end + the
// re-saved keyframes round-trip their VALUES.
func runV2_2LayrPosKfShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "v2_2_layrposkf.aep")
	resavedAEP := filepath.Join(tempDir, "v2_2_layrposkf.resaved.aep")
	doneFile := filepath.Join(tempDir, "v2_2_layrposkf.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/v2_2_layrposkf_args.json`

	p := aep.NewProject(target)
	comp, err := p.NewComposition("Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, _ := comp.NewShapeLayer("PosAnim")
	r, _ := l.RootGroup().AddRect()
	_ = r.SetSize([2]float64{200, 100})
	_ = l.Position().AddKeyframeLinear(0, [2]float64{100, 200})
	_ = l.Position().AddKeyframeLinear(2, [2]float64{700, 400})

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q}`, toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP))
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	runAeRunShipGate(t, aeExe, `E:/projects/tools/aep-parser/test_data/verify_v2_2_layrposkf.jsx`, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	if lines := strings.SplitN(string(content), "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("layr-pos-kf ship gate FAIL:\n%s", string(content))
	}

	root := parseAEP(t, resavedAEP)
	kfl := findShipList(root, "ADBE Position")
	if kfl == nil {
		t.Fatalf("resaved Layr Position has no keyframe container — position keyframes dropped")
	}
	lhd3 := findShipChunk(kfl, rifx.IDLhd3)
	if lhd3 == nil || binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C]) != 2 {
		t.Errorf("resaved Layr Position numKf != 2")
	}
	// Spatial dim-3 block: value at block+0x38 (X), +0x40 (Y), +0x48 (Z=0).
	ldat := findShipChunk(kfl, rifx.IDLdat)
	if ldat != nil && len(ldat.Data) >= 0x48+8 {
		x := math.Float64frombits(binary.BigEndian.Uint64(ldat.Data[0x38:0x40]))
		y := math.Float64frombits(binary.BigEndian.Uint64(ldat.Data[0x40:0x48]))
		if math.Abs(x-100) > 0.5 || math.Abs(y-200) > 0.5 {
			t.Errorf("resaved Layr Position kf0 = (%.3g,%.3g), want (100,200)", x, y)
		}
	}
}

func TestV2_2_LayrPosKf_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2LayrPosKfShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_LayrPosKf_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2LayrPosKfShipGate(t, aep.TargetAE2020, aeExe)
}
