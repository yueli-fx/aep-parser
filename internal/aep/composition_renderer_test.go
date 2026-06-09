package aep_test

import (
	"bytes"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// rendererFixtures maps a test_data fixture to its expected baseline renderer
// match-name (RE'd from py-aep's renderer samples).
var rendererFixtures = map[string]string{
	"../../test_data/renderer_classic_3d.aep":  "ADBE Escher",
	"../../test_data/renderer_advanced_3d.aep": "ADBE Calder",
	"../../test_data/renderer_cinema_4d.aep":   "ADBE Ernst",
	"../../test_data/renderer_ray_traced.aep":  "ADBE Picasso",
}

// expected prda lengths per renderer (template sizes).
var rendererPrdaLen = map[string]int{
	"ADBE Escher":  12,
	"ADBE Calder":  52,
	"ADBE Ernst":   20,
	"ADBE Picasso": 16,
}

// TestCompositionRendererRead confirms the baseline match-name decode on each
// renderer fixture.
func TestCompositionRendererRead(t *testing.T) {
	for path, want := range rendererFixtures {
		proj, err := aep.Open(path)
		if err != nil {
			t.Skipf("%s not present: %v", path, err)
		}
		if len(proj.Compositions) == 0 {
			t.Fatalf("%s: no compositions", path)
		}
		if got := proj.Compositions[0].Renderer; got != want {
			t.Errorf("%s: Renderer = %q, want %q", path, got, want)
		}
	}
}

// TestSetRendererRoundTrip switches each fixture to every other engine and
// verifies the match-name + prda length persist through WriteAEP + re-parse.
func TestSetRendererRoundTrip(t *testing.T) {
	const src = "../../test_data/renderer_classic_3d.aep"
	targets := []string{"ADBE Calder", "ADBE Ernst", "ADBE Picasso", "ADBE Escher"}

	for _, target := range targets {
		proj, err := aep.Open(src)
		if err != nil {
			t.Skipf("%s not present: %v", src, err)
		}
		comp := proj.Compositions[0]
		if err := aep.SetRenderer(comp, target); err != nil {
			t.Fatalf("SetRenderer(%q): %v", target, err)
		}
		if comp.Renderer != target {
			t.Errorf("in-mem Renderer = %q, want %q", comp.Renderer, target)
		}

		var buf bytes.Buffer
		if err := proj.WriteAEP(&buf); err != nil {
			t.Fatalf("WriteAEP: %v", err)
		}
		proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
		if err != nil {
			t.Fatalf("re-parse: %v", err)
		}
		c2 := proj2.Compositions[0]
		if c2.Renderer != target {
			t.Errorf("after roundtrip: Renderer = %q, want %q", c2.Renderer, target)
		}
		if got := len(c2.PrdaRawBytes()); got != rendererPrdaLen[target] {
			t.Errorf("after roundtrip to %q: prda len = %d, want %d", target, got, rendererPrdaLen[target])
		}
	}
}

// TestSetRendererExtendscriptAlias accepts the ExtendScript module name
// "ADBE Advanced 3d" and normalizes it to the binary match_name "ADBE Escher".
func TestSetRendererExtendscriptAlias(t *testing.T) {
	const src = "../../test_data/renderer_cinema_4d.aep"
	proj, err := aep.Open(src)
	if err != nil {
		t.Skipf("%s not present: %v", src, err)
	}
	comp := proj.Compositions[0]
	if err := aep.SetRenderer(comp, "ADBE Advanced 3d"); err != nil {
		t.Fatalf("SetRenderer(\"ADBE Advanced 3d\"): %v", err)
	}
	if comp.Renderer != "ADBE Escher" {
		t.Errorf("Renderer = %q, want binary name \"ADBE Escher\"", comp.Renderer)
	}
	if got := len(comp.PrdaRawBytes()); got != 12 {
		t.Errorf("prda len = %d, want 12 (Escher template)", got)
	}
}

// TestSetRendererUnknown rejects an unknown match-name without mutating.
func TestSetRendererUnknown(t *testing.T) {
	const src = "../../test_data/renderer_classic_3d.aep"
	proj, err := aep.Open(src)
	if err != nil {
		t.Skipf("%s not present: %v", src, err)
	}
	comp := proj.Compositions[0]
	before := comp.Renderer
	if err := aep.SetRenderer(comp, "ADBE Nonexistent"); err == nil {
		t.Fatal("SetRenderer(unknown) = nil, want error")
	}
	if comp.Renderer != before {
		t.Errorf("Renderer mutated on rejected call: %q (was %q)", comp.Renderer, before)
	}
}

// TestSetRendererNoBackref errors on a comp built outside the parser.
func TestSetRendererNoBackref(t *testing.T) {
	comp := &aep.Composition{Name: "synthetic"}
	if err := aep.SetRenderer(comp, "ADBE Escher"); err == nil {
		t.Fatal("SetRenderer on synthetic comp = nil, want error")
	}
}
