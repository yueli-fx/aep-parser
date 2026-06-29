// internal/aep/text_animator_struct_shipgate_test.go
//
// Promotes the Text Animator structural ops (Remove / MoveTo / Duplicate on the
// "ADBE Text Animators" indexed group) from Alpha to shipped. The ops already
// share the generic rebuildIndexedGroupChunk machinery that the Effect Parade
// ship-gated (incidents/property-indexed-group-structural-re.md); this gate
// proves the SAME ops on a Text Animators group: AE 2020 + AE 2025 ACCEPT the
// restructured group (no data-loss / corrupt) and read back the expected
// animator order. Structural — not a rendering-class capability — so the gate
// asserts on AE's read-back order (mirroring the Effect Parade gate), not pixels.
//
// Fixture: a from-scratch text layer with three animators, each carrying one
// distinct driven leaf so order is identifiable — Opacity, Skew, Fill Color.
// Variants (× AE 2020 + AE 2025):
//   - remove middle (Skew)        -> [Opacity, Fill Color]
//   - move last to front (Color)  -> [Fill Color, Opacity, Skew]
//   - duplicate first (Opacity)   -> [Opacity, Opacity, Skew, Fill Color]
//
// A non-AE round-trip subtest (TestTextAnimatorStruct_RoundTrip) asserts the
// same order survives WriteAEP → Open without an AE process. Gated by AE_SHIP_GATE.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

const (
	leafOpacity = "ADBE Text Opacity"
	leafSkew    = "ADBE Text Skew"
	leafColor   = "ADBE Text Fill Color"
)

// buildTextStruct3 builds a from-scratch text layer carrying three animators
// (Opacity, Skew, Fill Color), reopened so the property tree is parsed with the
// back-refs the structural ops need.
func buildTextStruct3(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "TXSTRUCT", 1280, 720, 24, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	tl, err := aep.NewTextLayer(comp, "TXT")
	if err != nil {
		t.Fatalf("NewTextLayer: %v", err)
	}
	if err := tl.SetText("ABCDEF"); err != nil {
		t.Fatalf("SetText: %v", err)
	}
	if _, err := aep.AddTextOpacityAnimator(tl, 50, 0, 100, 0); err != nil {
		t.Fatalf("AddTextOpacityAnimator: %v", err)
	}
	if _, err := aep.AddTextSkewAnimator(tl, 20, 0, 100, 0); err != nil {
		t.Fatalf("AddTextSkewAnimator: %v", err)
	}
	if _, err := aep.AddTextColorAnimator(tl, 0, 0, 1, 1, 0, 100, 0); err != nil {
		t.Fatalf("AddTextColorAnimator: %v", err)
	}
	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	return rp
}

// textAnimatorsGroup returns the layer's "ADBE Text Animators" indexed group.
func textAnimatorsGroup(t *testing.T, tl *aep.Layer) *aep.AEPropertyGroup {
	t.Helper()
	tp := tl.TextPropertiesGroup()
	if tp == nil {
		t.Fatal("layer has no Text Properties group")
	}
	g := tp.Group("ADBE Text Animators")
	if g == nil {
		t.Fatal("layer has no Text Animators group")
	}
	return g
}

// animatorTags returns each animator's driven-leaf match-name, in order.
func animatorTags(t *testing.T, tl *aep.Layer) []string {
	t.Helper()
	animators := textAnimatorsGroup(t, tl)
	var tags []string
	for i := 0; ; i++ {
		c := animators.ChildByIndex(i)
		if c == nil {
			break
		}
		ag, ok := c.(*aep.AEPropertyGroup)
		if !ok {
			continue
		}
		props := ag.Group("ADBE Text Animator Properties")
		if props == nil {
			tags = append(tags, "?")
			continue
		}
		leaf := props.ChildByIndex(0)
		if leaf == nil {
			tags = append(tags, "?")
			continue
		}
		tags = append(tags, leaf.PropertyMatchName())
	}
	return tags
}

type structVariant struct {
	name   string
	mutate func(t *testing.T, tl *aep.Layer) error
	expect []string
}

func structVariants() []structVariant {
	return []structVariant{
		{
			name: "RemoveMiddle",
			mutate: func(t *testing.T, tl *aep.Layer) error {
				return aep.RemovePropertyGroup(textAnimatorsGroup(t, tl).ChildByIndex(1).(*aep.AEPropertyGroup))
			},
			expect: []string{leafOpacity, leafColor},
		},
		{
			name: "MoveLastToFront",
			mutate: func(t *testing.T, tl *aep.Layer) error {
				return aep.MovePropertyGroup(textAnimatorsGroup(t, tl).ChildByIndex(2).(*aep.AEPropertyGroup), 0)
			},
			expect: []string{leafColor, leafOpacity, leafSkew},
		},
		{
			name: "DuplicateFirst",
			mutate: func(t *testing.T, tl *aep.Layer) error {
				_, err := aep.DuplicatePropertyGroup(textAnimatorsGroup(t, tl).ChildByIndex(0).(*aep.AEPropertyGroup))
				return err
			},
			expect: []string{leafOpacity, leafOpacity, leafSkew, leafColor},
		},
	}
}

func TestTextAnimatorStruct_RoundTrip(t *testing.T) {
	for _, v := range structVariants() {
		t.Run(v.name, func(t *testing.T) {
			rp := buildTextStruct3(t, aep.TargetAE2020)
			tl := rp.Compositions[0].LayerByName("TXT")
			if tl == nil {
				t.Fatal("TXT layer missing after build")
			}
			if got := animatorTags(t, tl); len(got) != 3 {
				t.Fatalf("pre-mutation animator count = %d (%v), want 3", len(got), got)
			}
			if err := v.mutate(t, tl); err != nil {
				t.Fatalf("mutate: %v", err)
			}

			path := filepath.Join(t.TempDir(), "txstruct.aep")
			f, err := os.Create(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := rp.WriteAEP(f); err != nil {
				f.Close()
				t.Fatalf("WriteAEP: %v", err)
			}
			f.Close()

			re, err := aep.Open(path)
			if err != nil {
				t.Fatalf("reopen: %v", err)
			}
			rtl := re.Compositions[0].LayerByName("TXT")
			if rtl == nil {
				t.Fatal("TXT layer missing after reopen")
			}
			got := animatorTags(t, rtl)
			if strings.Join(got, ",") != strings.Join(v.expect, ",") {
				t.Errorf("animator order = %v, want %v", got, v.expect)
			}
		})
	}
}

func runTextAnimatorStructGate(t *testing.T, aeExe, ver string, target aep.AETarget, v structVariant) {
	const argsPath = `e:/projects/tools/aep-parser/test_data/text_animator_struct_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_text_animator_struct.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	rp := buildTextStruct3(t, target)
	tl := rp.Compositions[0].LayerByName("TXT")
	if tl == nil {
		t.Fatal("TXT layer missing after build")
	}
	if err := v.mutate(t, tl); err != nil {
		t.Fatalf("mutate: %v", err)
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "txstruct_in.aep")
	resavedAEP := filepath.Join(tempDir, "txstruct_resaved.aep")
	doneFile := filepath.Join(tempDir, "txstruct.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	expectJSON := `["` + strings.Join(v.expect, `","`) + `"]`
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"expect":%s}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP), expectJSON)
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 300)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	body := string(content)
	t.Logf("%s %s AE readback:\n%s", v.name, ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("%s %s ship gate FAIL:\n%s", v.name, ver, body)
	}

	// Resave proof: AE kept the restructured group on re-encode.
	if _, err := aep.Open(resavedAEP); err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
}

func runTextAnimatorStructSuite(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	for _, v := range structVariants() {
		t.Run(v.name, func(t *testing.T) {
			runTextAnimatorStructGate(t, aeExe, ver, target, v)
		})
	}
}

func TestTextAnimatorStruct_AEShipGate_AE2020(t *testing.T) {
	runTextAnimatorStructSuite(t, ae2020(), "AE2020", aep.TargetAE2020)
}

func TestTextAnimatorStruct_AEShipGate_AE2025(t *testing.T) {
	runTextAnimatorStructSuite(t, ae2025(), "AE2025", aep.TargetAE2025)
}
