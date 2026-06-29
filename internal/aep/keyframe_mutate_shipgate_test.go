// internal/aep/keyframe_mutate_shipgate_test.go
//
// AE ship gate for the keyframe MUTATE setters (scene_keyframe_writers.go +
// SetStaticValue / SetFrameTime / InsertKeyframe / DeleteKeyframe). The
// create-path (AddKeyframeLinear / AddKeyframeWithEase) is already render-proven
// by mg_ease_shipgate_test.go; this gate proves the in-place MUTATE setters
// write bytes AE ingests into the correct DOM fields. Verified at the value
// surface (DOM readback) — the setters all produce byte layouts identical to
// what the gated create-path emits, so the render face is transitively covered;
// the mutate gate's distinct job is field-correct AE ingestion.
//
// One isolated shape layer per concern so a single bad setter never masks the
// rest:
//   - EASE: SetOutInterp/SetInInterp (→Bezier) + SetOutTemporalEase/
//     SetInTemporalEase — AE reads keyOut/InInterpolationType=BEZIER and
//     keyOut/InTemporalEase influence.
//   - VALT: SetTime + SetValue + SetFrameTime — AE reads keyTime/keyValue.
//   - TAN:  SetOutSpatialTangent/SetInSpatialTangent (bezier path) — AE reads
//     keyOut/InSpatialTangent.
//   - STAT: Property.SetStaticValue on a non-keyframed Position — AE reads value.
//   - INS:  InsertKeyframe 2→3 — AE reads numKeys.
//   - DEL:  DeleteKeyframe 3→2 — AE reads numKeys.
//
// Gated by AE_SHIP_GATE.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func buildKFMutateDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "KFMUT", 1920, 1080, 30, 4)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	addDot := func(name string, color [4]float64) *aep.ShapeLayer {
		dl, err := aep.NewShapeLayer(comp, name)
		if err != nil {
			t.Fatalf("NewShapeLayer %s: %v", name, err)
		}
		el, err := dl.RootGroup().AddEllipse()
		if err != nil {
			t.Fatalf("AddEllipse %s: %v", name, err)
		}
		if err := el.SetSize([2]float64{80, 80}); err != nil {
			t.Fatalf("%s SetSize: %v", name, err)
		}
		f, err := dl.RootGroup().AddFill()
		if err != nil {
			t.Fatalf("AddFill %s: %v", name, err)
		}
		if err := f.SetColor(color); err != nil {
			t.Fatalf("%s SetColor: %v", name, err)
		}
		return dl
	}

	ease := addDot("EASE", [4]float64{1.0, 0.55, 0.1, 1})
	mustKF := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	mustKF("EASE kf0", ease.Position().AddKeyframeLinear(0, [2]float64{300, 250}))
	mustKF("EASE kf1", ease.Position().AddKeyframeLinear(4, [2]float64{1500, 250}))

	valt := addDot("VALT", [4]float64{0.25, 0.85, 1.0, 1})
	mustKF("VALT kf0", valt.Position().AddKeyframeLinear(0, [2]float64{300, 500}))
	mustKF("VALT kf1", valt.Position().AddKeyframeLinear(2, [2]float64{900, 500}))
	mustKF("VALT kf2", valt.Position().AddKeyframeLinear(4, [2]float64{1500, 500}))

	tan := addDot("TAN", [4]float64{1.0, 0.2, 0.55, 1})
	mustKF("TAN kf0", tan.Position().AddKeyframeLinear(0, [2]float64{300, 800}))
	mustKF("TAN kf1", tan.Position().AddKeyframeLinear(4, [2]float64{1500, 800}))

	stat := addDot("STAT", [4]float64{0.7, 0.7, 0.2, 1})
	mustKF("STAT static", stat.Position().SetStaticValue([2]float64{600, 950}))

	ins := addDot("INS", [4]float64{0.4, 1.0, 0.45, 1})
	mustKF("INS kf0", ins.Position().AddKeyframeLinear(0, [2]float64{300, 150}))
	mustKF("INS kf1", ins.Position().AddKeyframeLinear(4, [2]float64{1500, 150}))

	del := addDot("DEL", [4]float64{0.9, 0.4, 0.9, 1})
	mustKF("DEL kf0", del.Position().AddKeyframeLinear(0, [2]float64{300, 1050}))
	mustKF("DEL kf1", del.Position().AddKeyframeLinear(2, [2]float64{900, 1050}))
	mustKF("DEL kf2", del.Position().AddKeyframeLinear(4, [2]float64{1500, 1050}))

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	cc := rp.Compositions[0]
	mustMut := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}

	// EASE: interpolation + temporal ease on the linear keyframes.
	ek := cc.LayerByName("EASE").Position().Keyframes
	if len(ek) != 2 {
		t.Fatalf("EASE keyframes = %d, want 2", len(ek))
	}
	mustMut("EASE SetOutInterp", ek[0].SetOutInterp(aep.InterpBezier))
	mustMut("EASE SetOutTemporalEase", ek[0].SetOutTemporalEase([]aep.TemporalEase{{Speed: 0, Influence: 0.85}}))
	mustMut("EASE SetInInterp", ek[1].SetInInterp(aep.InterpBezier))
	mustMut("EASE SetInTemporalEase", ek[1].SetInTemporalEase([]aep.TemporalEase{{Speed: 0, Influence: 0.6}}))

	// VALT: time + value + frame-time.
	vk := cc.LayerByName("VALT").Position().Keyframes
	if len(vk) != 3 {
		t.Fatalf("VALT keyframes = %d, want 3", len(vk))
	}
	// Layer Position parses as 3D (z=0) even on a 2D layer; AE's DOM still
	// returns 2-element arrays, so value/tangent slices are length 3 here.
	mustMut("VALT SetTime", vk[1].SetTime(1.6)) // stays between kf0 (0) and kf2
	mustMut("VALT SetValue", vk[1].SetValue([]float64{850, 560, 0}))
	mustMut("VALT SetFrameTime", vk[2].SetFrameTime(90)) // 90/30fps = 3.0s

	// TAN: spatial tangents (require bezier interp on the eased side).
	tk := cc.LayerByName("TAN").Position().Keyframes
	if len(tk) != 2 {
		t.Fatalf("TAN keyframes = %d, want 2", len(tk))
	}
	mustMut("TAN SetOutInterp", tk[0].SetOutInterp(aep.InterpBezier))
	mustMut("TAN SetOutSpatialTangent", tk[0].SetOutSpatialTangent([]float64{300, -250, 0}))
	mustMut("TAN SetInInterp", tk[1].SetInInterp(aep.InterpBezier))
	mustMut("TAN SetInSpatialTangent", tk[1].SetInSpatialTangent([]float64{-300, -250, 0}))

	// STAT: mutate a non-keyframed Position via Property.SetStaticValue.
	sp := cc.LayerByName("STAT").Position()
	if len(sp.Keyframes) != 0 {
		t.Fatalf("STAT should be static, got %d keyframes", len(sp.Keyframes))
	}
	mustMut("STAT SetStaticValue", sp.SetStaticValue([]float64{420, 950, 0}))

	// INS: insert an interior keyframe (2 → 3).
	ip := cc.LayerByName("INS").Position()
	if _, _, err := aep.InsertKeyframe(ip, 2.0, []float64{900, 150, 0}); err != nil {
		t.Fatalf("InsertKeyframe: %v", err)
	}
	if len(ip.Keyframes) != 3 {
		t.Fatalf("INS after insert = %d keyframes, want 3", len(ip.Keyframes))
	}

	// DEL: delete the interior keyframe (3 → 2).
	dp := cc.LayerByName("DEL").Position()
	if err := aep.DeleteKeyframe(dp, 1); err != nil {
		t.Fatalf("DeleteKeyframe: %v", err)
	}
	if len(dp.Keyframes) != 2 {
		t.Fatalf("DEL after delete = %d keyframes, want 2", len(dp.Keyframes))
	}

	return rp
}

func runKFMutateGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/kf_mutate_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_kf_mutate.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildKFMutateDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "kf_mutate_in.aep")
	resavedAEP := filepath.Join(tempDir, "kf_mutate_resaved.aep")
	doneFile := filepath.Join(tempDir, "kf_mutate.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP))
	if err := writeGeneratedArgs(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 240)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	body := string(content)
	t.Logf("kf mutate %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("kf mutate %s ship gate FAIL:\n%s", ver, body)
	}

	// Resave proof: the mutated keyframe data survives AE's re-encode and the
	// Go parser reads it back.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	cc := re.Compositions[0]

	easL := cc.LayerByName("EASE")
	if easL == nil {
		t.Fatal("resaved: EASE missing")
	}
	ek := easL.Position().Keyframes
	if len(ek) != 2 {
		t.Fatalf("resaved EASE keyframes = %d, want 2", len(ek))
	}
	if ek[0].OutInterp != aep.InterpBezier {
		t.Errorf("resaved EASE kf0 OutInterp = %v, want Bezier", ek[0].OutInterp)
	}
	if len(ek[0].OutTemporalEase) == 0 || ek[0].OutTemporalEase[0].Influence < 0.8 {
		t.Errorf("resaved EASE kf0 out ease lost: %+v", ek[0].OutTemporalEase)
	}

	valL := cc.LayerByName("VALT")
	if valL == nil {
		t.Fatal("resaved: VALT missing")
	}
	if vk := valL.Position().Keyframes; len(vk) != 3 {
		t.Errorf("resaved VALT keyframes = %d, want 3", len(vk))
	}
	if dl := cc.LayerByName("DEL"); dl == nil || len(dl.Position().Keyframes) != 2 {
		n := -1
		if dl != nil {
			n = len(dl.Position().Keyframes)
		}
		t.Errorf("resaved DEL keyframes = %d, want 2", n)
	}
	if il := cc.LayerByName("INS"); il == nil || len(il.Position().Keyframes) != 3 {
		n := -1
		if il != nil {
			n = len(il.Position().Keyframes)
		}
		t.Errorf("resaved INS keyframes = %d, want 3", n)
	}
}

func TestKeyframeMutate_AEShipGate_AE2020(t *testing.T) {
	runKFMutateGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestKeyframeMutate_AEShipGate_AE2025(t *testing.T) {
	runKFMutateGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
