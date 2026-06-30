package profile_test

import (
	"math"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
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

func TestBuildCopiesStructuredParseWarnings(t *testing.T) {
	project := &aep.Project{
		Warnings: []string{"short keyframe stream"},
		ParseWarnings: []aep.ParseWarning{{
			Chunk:   "lhd3",
			Offset:  12,
			Message: "short keyframe stream",
		}},
	}

	prof, err := profile.Build(project, profile.Options{})
	if err != nil {
		t.Fatalf("build profile: %v", err)
	}
	if got, want := prof.Meta.ParseWarnings, []string{"short keyframe stream"}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("ParseWarnings = %v, want %v", got, want)
	}
	if got, want := prof.Meta.StructuredParseWarnings, []aep.ParseWarning{{Chunk: "lhd3", Offset: 12, Message: "short keyframe stream"}}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("StructuredParseWarnings = %+v, want %+v", got, want)
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

func TestBuildSyntheticProjectIncludesEffectParamLayerRefs(t *testing.T) {
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Layer Ref", 640, 360, 24, 3)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if _, err := aep.NewSolidLayer(comp, "Matte", 640, 360, [3]float64{1, 1, 1}); err != nil {
		t.Fatalf("NewSolidLayer Matte: %v", err)
	}
	if _, err := aep.NewSolidLayer(comp, "FX", 640, 360, [3]float64{1, 0, 0}); err != nil {
		t.Fatalf("NewSolidLayer FX: %v", err)
	}
	project, err = aep.Reopen(project)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	comp = project.Compositions[0]
	matte := comp.LayerByName("Matte")
	fxLayer := comp.LayerByName("FX")
	if matte == nil || fxLayer == nil {
		t.Fatal("Matte/FX layers missing after reopen")
	}
	fx, err := aep.AddEffect(fxLayer, "ADBE Set Matte3")
	if err != nil {
		t.Fatalf("AddEffect: %v", err)
	}
	if err := aep.SetEffectLayerParam(fxLayer, fx, "ADBE Set Matte3-0001", matte); err != nil {
		t.Fatalf("SetEffectLayerParam: %v", err)
	}

	prof, err := profile.Build(project, profile.Options{})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	var fxEffect *profile.Effect
	var param *profile.Property
	for li := range prof.Comps[0].Layers {
		layer := &prof.Comps[0].Layers[li]
		if layer.Name != "FX" {
			continue
		}
		for ei := range layer.Effects {
			effect := &layer.Effects[ei]
			if effect.MatchName == "ADBE Set Matte3" {
				fxEffect = effect
			}
			for pi := range effect.Params {
				if effect.Params[pi].MatchName == "ADBE Set Matte3-0001" {
					param = &effect.Params[pi]
				}
			}
		}
	}
	if fxEffect == nil {
		t.Fatal("Set Matte effect not found")
	}
	if param == nil {
		t.Fatal("Set Matte layer param not found")
	}
	if param.LayerRef == nil || param.LayerRef.Name != "Matte" {
		t.Fatalf("LayerRef = %+v, want Matte", param.LayerRef)
	}
	if !containsString(fxEffect.TunedParams, "ADBE Set Matte3-0001") {
		t.Fatalf("TunedParams = %v, want ADBE Set Matte3-0001", fxEffect.TunedParams)
	}
}

func TestBuildMarkerFixtureIncludesCompMarkers(t *testing.T) {
	path := repoPath(t, "test_data", "fixtures", "re_compmarker.aep")
	project, err := aep.Open(path)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}

	prof, err := profile.Build(project, profile.Options{Path: path})
	if err != nil {
		t.Fatalf("build profile: %v", err)
	}

	comp := findProfileComp(t, prof, "RE_CM")
	if len(comp.Markers) != 2 {
		t.Fatalf("markers = %d, want 2", len(comp.Markers))
	}
	if comp.Markers[0].Comment != "comp marker A" {
		t.Fatalf("marker[0].comment = %q", comp.Markers[0].Comment)
	}
	if math.Abs(comp.Markers[0].Time-1.0) > 1e-3 {
		t.Fatalf("marker[0].time = %g, want 1.0", comp.Markers[0].Time)
	}
	if comp.Markers[1].Comment != "second marker" {
		t.Fatalf("marker[1].comment = %q", comp.Markers[1].Comment)
	}
	if math.Abs(comp.Markers[1].Time-2.5) > 1e-3 {
		t.Fatalf("marker[1].time = %g, want 2.5", comp.Markers[1].Time)
	}
	if comp.Markers[1].Chapter != "chap-X" {
		t.Fatalf("marker[1].chapter = %q, want chap-X", comp.Markers[1].Chapter)
	}
	if comp.Markers[1].Path.Path == "" || comp.Markers[1].Evidence.Level != profile.EvidenceL1Parsed {
		t.Fatalf("marker[1] path/evidence missing: %+v", comp.Markers[1])
	}
}

func TestBuildGuidesFixtureIncludesCompGuides(t *testing.T) {
	path := repoPath(t, "test_data", "fixtures", "guides.aep")
	project, err := aep.Open(path)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}

	prof, err := profile.Build(project, profile.Options{Path: path})
	if err != nil {
		t.Fatalf("build profile: %v", err)
	}

	comp := findProfileComp(t, prof, "guides_both")
	if len(comp.Guides) != 3 {
		t.Fatalf("guides = %d, want 3", len(comp.Guides))
	}

	want := []struct {
		orientation string
		position    float64
	}{
		{"horizontal", 270},
		{"horizontal", 810},
		{"vertical", 960},
	}
	for i, w := range want {
		got := comp.Guides[i]
		if got.Orientation != w.orientation {
			t.Fatalf("guide[%d].orientation = %q, want %q", i, got.Orientation, w.orientation)
		}
		if math.Abs(got.Position-w.position) > 1e-3 {
			t.Fatalf("guide[%d].position = %g, want %g", i, got.Position, w.position)
		}
		if got.Path.Path == "" || got.Evidence.Level != profile.EvidenceL1Parsed {
			t.Fatalf("guide[%d] path/evidence missing: %+v", i, got)
		}
	}
}

func TestBuildEssentialGraphicsFixtureIncludesCompControllers(t *testing.T) {
	path := repoPath(t, "test_data", "fixtures", "eg_multiple_controllers.aep")
	project, err := aep.Open(path)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}

	prof, err := profile.Build(project, profile.Options{Path: path})
	if err != nil {
		t.Fatalf("build profile: %v", err)
	}

	comp := findProfileComp(t, prof, "primary")
	if comp.MotionGraphicsTemplateName != "Untitled" {
		t.Fatalf("motion graphics template name = %q, want Untitled", comp.MotionGraphicsTemplateName)
	}
	if len(comp.EssentialGraphics) != 3 {
		t.Fatalf("essential graphics controllers = %d, want 3", len(comp.EssentialGraphics))
	}

	want := []struct {
		name string
		typ  string
	}{
		{"Brightness", "slider"},
		{"Layer Opacity", "slider"},
		{"Background Color", "color"},
	}
	for i, w := range want {
		got := comp.EssentialGraphics[i]
		if got.Name != w.name {
			t.Fatalf("essential_graphics[%d].name = %q, want %q", i, got.Name, w.name)
		}
		if got.Type != w.typ {
			t.Fatalf("essential_graphics[%d].type = %q, want %q", i, got.Type, w.typ)
		}
		if got.UUID == "" {
			t.Fatalf("essential_graphics[%d].uuid is empty", i)
		}
		if got.Path.Path == "" || got.Evidence.Level != profile.EvidenceL1Parsed {
			t.Fatalf("essential_graphics[%d] path/evidence missing: %+v", i, got)
		}
	}
}

func TestBuildRenderQueueFixtureIncludesQueueSummary(t *testing.T) {
	path := repoPath(t, "test_data", "fixtures", "rq_numitems_1.aep")
	project, err := aep.Open(path)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}

	prof, err := profile.Build(project, profile.Options{Path: path})
	if err != nil {
		t.Fatalf("build profile: %v", err)
	}

	if prof.RenderQueue == nil {
		t.Fatal("RenderQueue = nil, want summary")
	}
	if prof.RenderQueue.NumItems != 1 {
		t.Fatalf("RenderQueue.NumItems = %d, want 1", prof.RenderQueue.NumItems)
	}
	if len(prof.RenderQueue.Items) != 1 {
		t.Fatalf("RenderQueue.Items = %d, want 1", len(prof.RenderQueue.Items))
	}

	item := prof.RenderQueue.Items[0]
	if item.CompName != "TestComp" {
		t.Fatalf("render_queue.items[0].comp_name = %q, want TestComp", item.CompName)
	}
	if item.OutputModuleCount != 1 {
		t.Fatalf("render_queue.items[0].output_module_count = %d, want 1", item.OutputModuleCount)
	}
	if len(item.OutputModules) != 1 {
		t.Fatalf("render_queue.items[0].output_modules = %d, want 1", len(item.OutputModules))
	}
	output := item.OutputModules[0]
	if output.Name != "H.264 - Match Render Settings - 15 Mbps" {
		t.Fatalf("output_modules[0].name = %q", output.Name)
	}
	if output.FileTemplate != "[compName].[fileextension]" {
		t.Fatalf("output_modules[0].file_template = %q", output.FileTemplate)
	}
	if output.Path.Path == "" || output.Evidence.Level != profile.EvidenceL1Parsed {
		t.Fatalf("output module path/evidence missing: %+v", output)
	}
	if math.Abs(item.TimeSpanStart) > 1e-3 {
		t.Fatalf("render_queue.items[0].time_span_start = %g, want 0", item.TimeSpanStart)
	}
	if math.Abs(item.TimeSpanDuration-10) > 1e-3 {
		t.Fatalf("render_queue.items[0].time_span_duration = %g, want 10", item.TimeSpanDuration)
	}
	if item.Path.Path == "" || item.Evidence.Level != profile.EvidenceL1Parsed {
		t.Fatalf("render queue item path/evidence missing: %+v", item)
	}
}

func TestBuildSyntheticProjectIncludesLayerSourceRefs(t *testing.T) {
	project := aep.NewProject(aep.TargetAE2020)
	sourceComp, err := aep.NewComposition(project, "Source Comp", 640, 360, 30, 2)
	if err != nil {
		t.Fatalf("NewComposition source: %v", err)
	}
	mainComp, err := aep.NewComposition(project, "Main Comp", 1920, 1080, 30, 3)
	if err != nil {
		t.Fatalf("NewComposition main: %v", err)
	}
	precomp, err := aep.NewPrecompLayer(mainComp, sourceComp, "Nested Source")
	if err != nil {
		t.Fatalf("NewPrecompLayer: %v", err)
	}
	solid, err := aep.NewSolidLayer(mainComp, "Solid Source", 1920, 1080, [3]float64{0.2, 0.4, 0.8})
	if err != nil {
		t.Fatalf("NewSolidLayer: %v", err)
	}
	solidFootage := solid.SourceFootage()
	if solidFootage == nil {
		t.Fatal("solid source footage is nil")
	}

	prof, err := profile.Build(project, profile.Options{})
	if err != nil {
		t.Fatalf("build profile: %v", err)
	}

	precompProfile := findProfileLayer(t, prof, "Nested Source")
	if precompProfile.SourceRef == nil {
		t.Fatal("precomp SourceRef = nil")
	}
	if got := *precompProfile.SourceRef; got.ID != precomp.SourceID || got.Kind != "composition" || got.Name != "Source Comp" {
		t.Fatalf("precomp SourceRef = %+v, want ID=%d Kind=composition Name=%q", got, precomp.SourceID, "Source Comp")
	}

	solidProfile := findProfileLayer(t, prof, "Solid Source")
	if solidProfile.SourceRef == nil {
		t.Fatal("solid SourceRef = nil")
	}
	if got := *solidProfile.SourceRef; got.ID != solid.SourceID || got.Kind != "footage" || got.Name != solidFootage.Name {
		t.Fatalf("solid SourceRef = %+v, want ID=%d Kind=footage Name=%q", got, solid.SourceID, solidFootage.Name)
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

func findProfileComp(t *testing.T, prof *profile.Profile, name string) *profile.Composition {
	t.Helper()
	for i := range prof.Comps {
		comp := &prof.Comps[i]
		if comp.Name == name {
			return comp
		}
	}
	t.Fatalf("profile comp %q not found", name)
	return nil
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
