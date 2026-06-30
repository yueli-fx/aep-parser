package projectindex

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestCorpusBuilderExtractsDurableLayerSourceAndEffectFacts(t *testing.T) {
	glow := &aep.Effect{MatchName: "ADBE Glo2", Name: "Glow"}
	blur := &aep.Effect{MatchName: "ADBE Gaussian Blur 2", Name: "Gaussian Blur"}
	layerA := &aep.Layer{ID: 11, Index: 0, Name: "Title", SourceID: 100, Effects: []*aep.Effect{glow}}
	layerB := &aep.Layer{ID: 12, Index: 1, Name: "BG", SourceID: 200, Effects: []*aep.Effect{blur}}
	project := &aep.Project{
		Compositions: []*aep.Composition{
			{ID: 7, Name: "Main", Layers: []*aep.Layer{layerA, layerB}},
		},
		Footage: []*aep.Footage{
			{ID: 100, Name: "Logo.png"},
			{ID: 200, Name: "Plate.mov"},
		},
	}
	builder := NewCorpusBuilder()

	if err := builder.AddProject("samples/project.aep", project); err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	corpus, err := builder.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if len(corpus.Projects) != 1 {
		t.Fatalf("len(Projects) = %d, want 1", len(corpus.Projects))
	}
	projectMeta := corpus.Projects[0]
	if projectMeta.ID == "" || projectMeta.Fingerprint == "" {
		t.Fatalf("project ID/fingerprint must be populated: %+v", projectMeta)
	}
	if projectMeta.Path != "samples/project.aep" || projectMeta.CompCount != 1 || projectMeta.LayerCount != 2 || projectMeta.EffectCount != 2 {
		t.Fatalf("project meta = %+v, want path/counts", projectMeta)
	}
	if len(corpus.Facts) != 4 {
		t.Fatalf("len(Facts) = %d, want 4", len(corpus.Facts))
	}
	assertCorpusFact(t, corpus.Facts[0], projectMeta.ID, FactLayerSource, "source_id", "100", "Title")
	assertCorpusFact(t, corpus.Facts[1], projectMeta.ID, FactEffectUsage, "effect.match_name", "ADBE Glo2", "Title")
	assertCorpusFact(t, corpus.Facts[2], projectMeta.ID, FactLayerSource, "source_id", "200", "BG")
	assertCorpusFact(t, corpus.Facts[3], projectMeta.ID, FactEffectUsage, "effect.match_name", "ADBE Gaussian Blur 2", "BG")
	if corpus.Facts[0].Location.ItemName != "Logo.png" || corpus.Facts[2].Location.ItemName != "Plate.mov" {
		t.Fatalf("source item names = %q/%q, want footage names", corpus.Facts[0].Location.ItemName, corpus.Facts[2].Location.ItemName)
	}
}

func TestCorpusJSONDoesNotExposeProjectPointers(t *testing.T) {
	builder := NewCorpusBuilder()
	if err := builder.AddProject("", &aep.Project{
		Compositions: []*aep.Composition{
			{ID: 1, Name: "Main", Layers: []*aep.Layer{
				{ID: 2, Index: 0, Name: "Layer", Effects: []*aep.Effect{{MatchName: "ADBE Glo2"}}},
			}},
		},
	}); err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	corpus, err := builder.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	data, err := json.Marshal(corpus)
	if err != nil {
		t.Fatalf("json.Marshal(Corpus): %v", err)
	}
	got := string(data)
	for _, want := range []string{
		`"projects":[`,
		`"facts":[`,
		`"kind":"effect_usage"`,
		`"location":`,
		`"match":{"field":"effect.match_name","value":"ADBE Glo2"}`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("JSON %s missing %s", got, want)
		}
	}
	for _, forbidden := range []string{"Pointers", "pointers", "Compositions", "Footage"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("JSON %s should not contain %s", got, forbidden)
		}
	}
}

func TestCorpusBuilderAllowsNilProjects(t *testing.T) {
	builder := NewCorpusBuilder()
	if err := builder.AddProject("missing.aep", nil); err != nil {
		t.Fatalf("AddProject(nil): %v", err)
	}
	corpus, err := builder.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(corpus.Projects) != 1 {
		t.Fatalf("len(Projects) = %d, want 1", len(corpus.Projects))
	}
	if corpus.Projects[0].Path != "missing.aep" || corpus.Projects[0].CompCount != 0 {
		t.Fatalf("nil project meta = %+v, want path and zero counts", corpus.Projects[0])
	}
	if len(corpus.Facts) != 0 {
		t.Fatalf("len(Facts) = %d, want 0", len(corpus.Facts))
	}
}

func assertCorpusFact(t *testing.T, got CorpusFact, projectID string, kind CorpusFactKind, field, value, layerName string) {
	t.Helper()
	if got.ProjectID != projectID || got.Kind != kind {
		t.Fatalf("fact identity = project %q kind %q, want %q/%q", got.ProjectID, got.Kind, projectID, kind)
	}
	if got.Match.Field != field || got.Match.Value != value {
		t.Fatalf("fact match = %+v, want %s=%s", got.Match, field, value)
	}
	if got.Location.LayerName != layerName {
		t.Fatalf("fact layer name = %q, want %q", got.Location.LayerName, layerName)
	}
	if got.Summary == "" {
		t.Fatalf("fact summary is empty: %+v", got)
	}
}
