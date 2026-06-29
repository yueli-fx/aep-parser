// internal/aep/text_animator_rotxy_test.go
//
// Round-trip survival test for the write-only Text Animator leaves Rotation X /
// Rotation Y. They are NOT render-gated (per-character 3D rotation is visually
// inert in a plain 2D text layer — see text_animator_neighbor_shipgate_test.go +
// incidents/text-animator-create-re.md), so this asserts only the thing that IS
// true of them: the facade writes the leaf and its value survives a WriteAEP →
// Open round-trip. No AE process involved; runs in the normal suite.
package aep_test

import (
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestTextRotationXY_RoundTrip(t *testing.T) {
	cases := []struct {
		name  string
		add   func(*aep.Layer) error
		match string
	}{
		{"RotationX", func(tl *aep.Layer) error { _, e := aep.AddTextRotationXAnimator(tl, 75, 0, 100, 0); return e }, "ADBE Text Rotation X"},
		{"RotationY", func(tl *aep.Layer) error { _, e := aep.AddTextRotationYAnimator(tl, 75, 0, 100, 0); return e }, "ADBE Text Rotation Y"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := aep.NewProject(aep.TargetAE2020)
			comp, err := aep.NewComposition(p, "RT", 1280, 720, 24, 5)
			if err != nil {
				t.Fatalf("NewComposition: %v", err)
			}
			tl, err := aep.NewTextLayer(comp, "TXT")
			if err != nil {
				t.Fatalf("NewTextLayer: %v", err)
			}
			if err := tl.SetText("ABC"); err != nil {
				t.Fatalf("SetText: %v", err)
			}
			if err := tc.add(tl); err != nil {
				t.Fatalf("add %s animator: %v", tc.name, err)
			}

			path := filepath.Join(t.TempDir(), "rt.aep")
			f, err := os.Create(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := p.WriteAEP(f); err != nil {
				f.Close()
				t.Fatalf("WriteAEP: %v", err)
			}
			f.Close()

			if _, err := aep.Open(path); err != nil {
				t.Fatalf("reopen: %v", err)
			}
			root := parseAEP(t, path)
			cdat := streamCdat(root, tc.match)
			if len(cdat) < 8 {
				t.Fatalf("%s cdat missing after round-trip", tc.match)
			}
			if v := math.Float64frombits(binary.BigEndian.Uint64(cdat[0:8])); math.Abs(v-75) > 0.5 {
				t.Errorf("%s = %v, want 75", tc.match, v)
			}
		})
	}
}
