package aep_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// Golden values cross-checked against py-aep
// flightdeck/charts/py-aep/samples/models/composition/guides.json.
// py-aep's orientationType is the logical enum (HORIZONTAL=0, VERTICAL=1);
// our Orientation carries the binary code (2=horizontal, 1=vertical) and
// .String() bridges the two. Comp "guides_both" 1920x1080: two horizontal
// guides (top offsets 270/810) + one vertical (left offset 960).
func TestGuidesReader(t *testing.T) {
	proj, err := aep.Open("../../test_data/guides.aep")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	cases := []struct {
		comp string
		want []struct {
			orient string
			pos    float64
		}
	}{
		{"guides_both", []struct {
			orient string
			pos    float64
		}{{"horizontal", 270}, {"horizontal", 810}, {"vertical", 960}}},
		{"guides_horizontal", []struct {
			orient string
			pos    float64
		}{{"horizontal", 540}}},
		{"guides_none", nil},
	}

	for _, c := range cases {
		t.Run(c.comp, func(t *testing.T) {
			comp := proj.CompositionByName(c.comp)
			if comp == nil {
				t.Fatalf("comp %q not found", c.comp)
			}
			if len(comp.Guides) != len(c.want) {
				t.Fatalf("len(Guides) = %d, want %d", len(comp.Guides), len(c.want))
			}
			for i, w := range c.want {
				g := comp.Guides[i]
				if g.Orientation.String() != w.orient {
					t.Errorf("Guides[%d].Orientation = %q, want %q", i, g.Orientation.String(), w.orient)
				}
				if g.Position != w.pos {
					t.Errorf("Guides[%d].Position = %v, want %v", i, g.Position, w.pos)
				}
			}
		})
	}
}

func TestGuidesJSON(t *testing.T) {
	proj, err := aep.Open("../../test_data/guides.aep")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	var sb strings.Builder
	if err := proj.WriteJSON(&sb); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	out := sb.String()
	for _, want := range []string{
		`"guides"`,
		`"orientation": "horizontal"`,
		`"orientation": "vertical"`,
		`"position": 960`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("JSON missing %s", want)
		}
	}
}

// Reading guides must not perturb byte-identical round-trip: the Gide subtree
// is untouched leaves/lists (opaque preservation).
func TestGuidesRoundTripByteIdentical(t *testing.T) {
	const path = "../../test_data/guides.aep"
	orig, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	proj, err := aep.FromReader(bytes.NewReader(orig))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), orig) {
		t.Errorf("roundtrip not byte-identical: in=%d out=%d", len(orig), buf.Len())
	}
}

// Length-preserving guide setters: patch position/orientation in place, write,
// re-parse, confirm persisted.
func TestGuidesWriteRoundTrip(t *testing.T) {
	orig, err := os.ReadFile("../../test_data/guides.aep")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	proj, err := aep.FromReader(bytes.NewReader(orig))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	comp := proj.CompositionByName("guides_both")
	if comp == nil || len(comp.Guides) != 3 {
		t.Fatalf("want guides_both with 3 guides")
	}
	comp.Guides[0].SetPosition(123.5)                  // 270 -> 123.5
	comp.Guides[2].SetOrientation(aep.GuideHorizontal) // vertical -> horizontal

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	g := re.CompositionByName("guides_both").Guides
	if g[0].Position != 123.5 {
		t.Errorf("Guides[0].Position = %v, want 123.5", g[0].Position)
	}
	if g[2].Orientation != aep.GuideHorizontal {
		t.Errorf("Guides[2].Orientation = %v, want horizontal", g[2].Orientation)
	}
}
