package profile_test

import (
	"math"
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

func TestBuildSyntheticProjectIncludesTrackMatteRef(t *testing.T) {
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Matte Comp", 1920, 1080, 30, 3)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	matte, err := aep.NewSolidLayer(comp, "Matte Source", 1920, 1080, [3]float64{1, 1, 1})
	if err != nil {
		t.Fatalf("NewSolidLayer matte: %v", err)
	}
	fill, err := aep.NewSolidLayer(comp, "Fill", 1920, 1080, [3]float64{0, 0, 1})
	if err != nil {
		t.Fatalf("NewSolidLayer fill: %v", err)
	}
	if err := fill.SetTrackMatte(aep.TrackMatteAlpha); err != nil {
		t.Fatalf("SetTrackMatte: %v", err)
	}

	prof, err := profile.Build(project, profile.Options{})
	if err != nil {
		t.Fatalf("build profile: %v", err)
	}

	layer := findProfileLayer(t, prof, "Fill")
	if layer.Flags.TrackMatte != uint8(aep.TrackMatteAlpha) {
		t.Fatalf("TrackMatte = %d, want %d", layer.Flags.TrackMatte, aep.TrackMatteAlpha)
	}
	if layer.MatteRef == nil {
		t.Fatal("MatteRef = nil, want matte source layer")
	}
	if layer.MatteRef.ID != matte.ID || layer.MatteRef.Name != matte.Name {
		t.Fatalf("MatteRef = %+v, want ID=%d Name=%q", layer.MatteRef, matte.ID, matte.Name)
	}
}

func TestBuildSyntheticProjectIncludesExplicitTrackMatteRef(t *testing.T) {
	project := aep.NewProject(aep.TargetAE2025)
	comp, err := aep.NewComposition(project, "Explicit Matte Comp", 1920, 1080, 30, 3)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	fill, err := aep.NewSolidLayer(comp, "Fill", 1920, 1080, [3]float64{0, 0, 1})
	if err != nil {
		t.Fatalf("NewSolidLayer fill: %v", err)
	}
	matte, err := aep.NewSolidLayer(comp, "Explicit Matte", 1920, 1080, [3]float64{1, 1, 1})
	if err != nil {
		t.Fatalf("NewSolidLayer matte: %v", err)
	}
	if err := fill.SetTrackMatteSource(matte, aep.TrackMatteLuma); err != nil {
		t.Fatalf("SetTrackMatteSource: %v", err)
	}

	prof, err := profile.Build(project, profile.Options{})
	if err != nil {
		t.Fatalf("build profile: %v", err)
	}

	layer := findProfileLayer(t, prof, "Fill")
	if layer.Flags.TrackMatte != uint8(aep.TrackMatteLuma) {
		t.Fatalf("TrackMatte = %d, want %d", layer.Flags.TrackMatte, aep.TrackMatteLuma)
	}
	if layer.MatteRef == nil {
		t.Fatal("MatteRef = nil, want explicit matte source layer")
	}
	if layer.MatteRef.ID != matte.ID || layer.MatteRef.Name != matte.Name {
		t.Fatalf("MatteRef = %+v, want ID=%d Name=%q", layer.MatteRef, matte.ID, matte.Name)
	}
}

func TestBuildSyntheticProjectUsesParsedInOutPoints(t *testing.T) {
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Timing Comp", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	layer, err := aep.NewSolidLayer(comp, "Timed", 1920, 1080, [3]float64{0.5, 0.5, 0.5})
	if err != nil {
		t.Fatalf("NewSolidLayer: %v", err)
	}
	if err := layer.SetStartTime(1.25); err != nil {
		t.Fatalf("SetStartTime: %v", err)
	}
	if err := layer.SetInPoint(0.5); err != nil {
		t.Fatalf("SetInPoint: %v", err)
	}
	if err := layer.SetOutPoint(3.75); err != nil {
		t.Fatalf("SetOutPoint: %v", err)
	}

	prof, err := profile.Build(project, profile.Options{})
	if err != nil {
		t.Fatalf("build profile: %v", err)
	}

	got := findProfileLayer(t, prof, "Timed").Timing
	if math.Abs(got.InPoint-0.5) > 1e-4 {
		t.Fatalf("InPoint = %g, want 0.5", got.InPoint)
	}
	if math.Abs(got.OutPoint-3.75) > 1e-4 {
		t.Fatalf("OutPoint = %g, want 3.75", got.OutPoint)
	}
	if math.Abs(got.StartTime-1.25) > 1e-4 {
		t.Fatalf("StartTime = %g, want 1.25", got.StartTime)
	}
}

func findProfileLayer(t *testing.T, prof *profile.Profile, name string) *profile.Layer {
	t.Helper()
	for ci := range prof.Comps {
		for li := range prof.Comps[ci].Layers {
			layer := &prof.Comps[ci].Layers[li]
			if layer.Name == name {
				return layer
			}
		}
	}
	t.Fatalf("profile layer %q not found", name)
	return nil
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
