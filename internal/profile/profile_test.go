package profile_test

import (
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/profile"
)

func repoPath(t *testing.T, parts ...string) string {
	t.Helper()
	all := append([]string{"..", ".."}, parts...)
	return filepath.Join(all...)
}

func TestBuildTextFixtureIncludesSchemaPathsEvidenceAndText(t *testing.T) {
	path := repoPath(t, "flightdeck", "showcase", "text", "text.aep")
	project, err := aep.Open(path)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}

	prof, err := profile.Build(project, profile.Options{Path: path})
	if err != nil {
		t.Fatalf("build profile: %v", err)
	}

	if prof.SchemaVersion != 1 {
		t.Fatalf("SchemaVersion = %d, want 1", prof.SchemaVersion)
	}
	if prof.Meta.Path != path {
		t.Fatalf("Meta.Path = %q, want %q", prof.Meta.Path, path)
	}
	if len(prof.Comps) != 1 {
		t.Fatalf("len(Comps) = %d, want 1", len(prof.Comps))
	}

	comp := prof.Comps[0]
	if comp.Path.Path == "" || comp.Path.DisplayPath == "" {
		t.Fatalf("comp path not populated: %+v", comp.Path)
	}
	if comp.Evidence.Level != profile.EvidenceL1Parsed {
		t.Fatalf("comp evidence = %q, want %q", comp.Evidence.Level, profile.EvidenceL1Parsed)
	}
	if len(comp.Layers) == 0 {
		t.Fatal("expected layers")
	}

	var textLayer *profile.Layer
	for i := range comp.Layers {
		layer := &comp.Layers[i]
		if layer.Path.Path == "" || layer.Path.DisplayPath == "" {
			t.Fatalf("layer path not populated: %+v", layer.Path)
		}
		if layer.Evidence.Level != profile.EvidenceL1Parsed {
			t.Fatalf("layer evidence = %q, want %q", layer.Evidence.Level, profile.EvidenceL1Parsed)
		}
		if layer.Type == "text" {
			textLayer = layer
		}
	}
	if textLayer == nil {
		t.Fatal("expected text layer")
	}
	if textLayer.Text == nil || textLayer.Text.Text == "" {
		t.Fatalf("text source not surfaced: %+v", textLayer.Text)
	}
}

func TestBuildEffectsFixtureIncludesEffectUsageAndTunedParams(t *testing.T) {
	path := repoPath(t, "flightdeck", "showcase", "effects", "effects.aep")
	project, err := aep.Open(path)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	dict, err := profile.LoadEffectDictionary(repoPath(t, "data", "effects-dict", "effects_en_US_25.1x68.json"))
	if err != nil {
		t.Fatalf("load dictionary: %v", err)
	}

	prof, err := profile.Build(project, profile.Options{Path: path, Dict: dict})
	if err != nil {
		t.Fatalf("build profile: %v", err)
	}

	if got := prof.Fingerprint.EffectUsage["ADBE Gaussian Blur 2"]; got != 1 {
		t.Fatalf("Gaussian Blur usage = %d, want 1", got)
	}

	var blur *profile.Effect
	for ci := range prof.Comps {
		for li := range prof.Comps[ci].Layers {
			for ei := range prof.Comps[ci].Layers[li].Effects {
				effect := &prof.Comps[ci].Layers[li].Effects[ei]
				if effect.MatchName == "ADBE Gaussian Blur 2" {
					blur = effect
				}
			}
		}
	}
	if blur == nil {
		t.Fatal("Gaussian Blur effect not found")
	}
	if blur.DependencyClass != "native" {
		t.Fatalf("DependencyClass = %q, want native", blur.DependencyClass)
	}
	if !containsString(blur.TunedParams, "ADBE Gaussian Blur 2-0001") {
		t.Fatalf("TunedParams = %v, want ADBE Gaussian Blur 2-0001", blur.TunedParams)
	}
	var blurriness *profile.Property
	for i := range blur.Params {
		if blur.Params[i].MatchName == "ADBE Gaussian Blur 2-0001" {
			blurriness = &blur.Params[i]
		}
	}
	if blurriness == nil {
		t.Fatal("Blurriness param not found")
	}
	if !strings.Contains(blurriness.Path.Path, ".params.by_match_name") {
		t.Fatalf("param path = %q, want effect params path", blurriness.Path.Path)
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
