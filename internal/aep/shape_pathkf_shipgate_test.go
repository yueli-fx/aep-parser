// internal/aep/shape_pathkf_shipgate_test.go
//
// AE ship gate: animated shape-path (Phase 2 lower) acceptance + persistence.
// Gated by AE_SHIP_GATE env var; CI / no-AE runs SKIP cleanly.
//
// Builds a from-scratch ShapeLayer whose Path stream carries 3 linear keyframes
// (distinct vertex counts), lowered as N shaps + a tdbs time table inside the
// embedded om-s body (spliceAnimatedPath). Proves AE accepts the multi-frame
// om-s without crashing / dropping the layer (from-scratch path CRASHED AE 2020
// historically), then re-decodes the AE-resaved file to confirm the animation
// (N shaps + time table) survived. See incident-reports/path-keyframe-write-re.md.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/rifx"
)

func runV2_2PathKfShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "v2_2_pathkf.aep")
	resavedAEP := filepath.Join(tempDir, "v2_2_pathkf.resaved.aep")
	doneFile := filepath.Join(tempDir, "v2_2_pathkf.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/v2_2_pathkf_args.json`

	frames := []struct {
		t     float64
		verts [][2]float64
	}{
		{0, [][2]float64{{0, 0}, {40, 0}, {40, 40}, {0, 40}}},
		{1, [][2]float64{{0, 0}, {100, 0}, {100, 100}, {0, 100}}},
		{2, [][2]float64{{0, 0}, {200, 0}, {100, 200}}},
	}

	p := aep.NewProject(target)
	comp, err := p.NewComposition("Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, err := aep.NewShapeLayer(comp, "PathKf")
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	pth, err := l.RootGroup().AddPath()
	if err != nil {
		t.Fatalf("AddPath: %v", err)
	}
	for _, f := range frames {
		bp := aep.BezierPath{
			Vertices:    f.verts,
			InTangents:  make([][2]float64, len(f.verts)),
			OutTangents: make([][2]float64, len(f.verts)),
			Closed:      true,
		}
		if err := pth.Path().AddKeyframeLinear(f.t, bp); err != nil {
			t.Fatalf("AddKeyframeLinear(%g): %v", f.t, err)
		}
	}

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

	jsxPath := `E:/projects/tools/aep-parser/test_data/verify_v2_2_pathkf.jsx`
	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.SplitN(string(content), "\n", 2)
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("path-kf ship gate FAIL:\n%s", string(content))
	}

	// Re-decode the AE-resaved path: the om-s must hold N shaps (one per
	// keyframe) + a tdbs time table whose kfl is present. om-s is a LIST whose
	// FormType is "om-s" (not a leaf chunk ID), so search by FormType.
	root := parseAEP(t, resavedAEP)
	oms := findShipListByForm(root, rifx.IDOmS)
	if oms == nil {
		t.Fatalf("resaved: no om-s (path dropped)")
	}
	var nShaps int
	var timeKfl *rifx.Chunk
	for _, ch := range oms.Children {
		if !ch.IsList() {
			continue
		}
		switch ch.FormType {
		case rifx.IDOmks:
			for _, s := range ch.Children {
				if s.IsList() && s.FormType == rifx.IDShap {
					nShaps++
				}
			}
		case rifx.IDTdbs:
			for _, c := range ch.Children {
				if c.IsList() && c.FormType == rifx.IDkfl {
					timeKfl = c
				}
			}
		}
	}
	if nShaps < len(frames) {
		t.Errorf("resaved: om-s shaps = %d, want >=%d (animation collapsed)", nShaps, len(frames))
	}
	if timeKfl == nil {
		t.Errorf("resaved: tdbs has no time-table kfl (animation dropped)")
	}
}

func TestV2_2_PathKf_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2PathKfShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_PathKf_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2PathKfShipGate(t, aep.TargetAE2020, aeExe)
}
