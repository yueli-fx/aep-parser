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

func TestCorpusBuilderExtractsExpressionFacts(t *testing.T) {
	position := &aep.Property{
		MatchName:  "ADBE Position",
		Name:       "Position",
		Expression: "wiggle(2, 20)",
	}
	blurAmount := &aep.Property{
		MatchName:  "ADBE Gaussian Blur 2-0001",
		Name:       "Blurriness",
		Expression: "time * 12",
	}
	project := &aep.Project{
		Compositions: []*aep.Composition{
			{ID: 7, Name: "Main", Layers: []*aep.Layer{
				{
					ID:         11,
					Index:      0,
					Name:       "Title",
					Properties: []*aep.Property{position},
					Effects: []*aep.Effect{
						{MatchName: "ADBE Gaussian Blur 2", Name: "Gaussian Blur", Parameters: []*aep.Property{blurAmount}},
					},
				},
			}},
		},
	}
	builder := NewCorpusBuilder()

	if err := builder.AddProject("samples/expressions.aep", project); err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	corpus, err := builder.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	projectID := corpus.Projects[0].ID
	expressionFacts := factsByKind(corpus.Facts, FactExpression)
	if len(expressionFacts) != 2 {
		t.Fatalf("len(expressionFacts) = %d, want 2; facts=%+v", len(expressionFacts), corpus.Facts)
	}
	assertCorpusFact(t, expressionFacts[0], projectID, FactExpression, "property.expression", "wiggle(2, 20)", "Title")
	if expressionFacts[0].Location.PropertyMatchName != "ADBE Position" || expressionFacts[0].Location.PropertyPath != "layers[].properties[1]" {
		t.Fatalf("first expression location = %+v, want layer position property", expressionFacts[0].Location)
	}
	assertCorpusFact(t, expressionFacts[1], projectID, FactExpression, "property.expression", "time * 12", "Title")
	if expressionFacts[1].Location.EffectMatchName != "ADBE Gaussian Blur 2" || expressionFacts[1].Location.PropertyMatchName != "ADBE Gaussian Blur 2-0001" || expressionFacts[1].Location.PropertyPath != "layers[].effects[1].params[1]" {
		t.Fatalf("second expression location = %+v, want effect param property", expressionFacts[1].Location)
	}
}

func TestCorpusBuilderExtractsTextStyleFacts(t *testing.T) {
	project := &aep.Project{
		Compositions: []*aep.Composition{
			{ID: 7, Name: "Main", Layers: []*aep.Layer{
				{
					ID:    21,
					Index: 0,
					Name:  "Title",
					Type:  aep.LayerTypeText,
					TextSource: &aep.TextSource{
						Fonts: []string{"Inter-Regular", "Helvetica-Bold"},
						Runs: []aep.TextStyleRun{
							{FontIndex: 0, FontName: "Inter-Regular"},
							{FontIndex: 1},
						},
					},
				},
			}},
		},
	}
	builder := NewCorpusBuilder()

	if err := builder.AddProject("samples/text.aep", project); err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	corpus, err := builder.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	projectID := corpus.Projects[0].ID
	textFacts := factsByKind(corpus.Facts, FactTextStyle)
	if len(textFacts) != 2 {
		t.Fatalf("len(textFacts) = %d, want 2; facts=%+v", len(textFacts), corpus.Facts)
	}
	assertCorpusFact(t, textFacts[0], projectID, FactTextStyle, "text.font", "Inter-Regular", "Title")
	if textFacts[0].Location.PropertyPath != "layers[].text.runs[1]" {
		t.Fatalf("first text fact location = %+v, want first run", textFacts[0].Location)
	}
	assertCorpusFact(t, textFacts[1], projectID, FactTextStyle, "text.font", "Helvetica-Bold", "Title")
	if textFacts[1].Location.PropertyPath != "layers[].text.runs[2]" {
		t.Fatalf("second text fact location = %+v, want second run", textFacts[1].Location)
	}
}

func TestCorpusBuilderExtractsShapeUsageFacts(t *testing.T) {
	project := &aep.Project{
		Compositions: []*aep.Composition{
			{ID: 7, Name: "Main", Layers: []*aep.Layer{
				{
					ID:    31,
					Index: 0,
					Name:  "Shape Layer",
					Type:  aep.LayerTypeShape,
					ShapePrimitives: []*aep.ShapePrimitive{
						{Kind: aep.ShapePrimitiveRect, GroupName: "Box"},
						{Kind: aep.ShapePrimitiveStar, GroupName: "Spark"},
					},
					ShapePaths: []*aep.ShapePath{
						{Name: "Logo Path", Closed: true},
					},
				},
			}},
		},
	}
	builder := NewCorpusBuilder()

	if err := builder.AddProject("samples/shapes.aep", project); err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	corpus, err := builder.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	projectID := corpus.Projects[0].ID
	shapeFacts := factsByKind(corpus.Facts, FactShapeUsage)
	if len(shapeFacts) != 3 {
		t.Fatalf("len(shapeFacts) = %d, want 3; facts=%+v", len(shapeFacts), corpus.Facts)
	}
	assertCorpusFact(t, shapeFacts[0], projectID, FactShapeUsage, "shape.kind", "rect", "Shape Layer")
	if shapeFacts[0].Location.PropertyPath != "layers[].shapes.primitives[1]" {
		t.Fatalf("first shape location = %+v, want first primitive", shapeFacts[0].Location)
	}
	assertCorpusFact(t, shapeFacts[1], projectID, FactShapeUsage, "shape.kind", "star", "Shape Layer")
	if shapeFacts[1].Location.PropertyPath != "layers[].shapes.primitives[2]" {
		t.Fatalf("second shape location = %+v, want second primitive", shapeFacts[1].Location)
	}
	assertCorpusFact(t, shapeFacts[2], projectID, FactShapeUsage, "shape.kind", "path", "Shape Layer")
	if shapeFacts[2].Location.PropertyPath != "layers[].shapes.paths[1]" {
		t.Fatalf("third shape location = %+v, want first path", shapeFacts[2].Location)
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

func factsByKind(facts []CorpusFact, kind CorpusFactKind) []CorpusFact {
	out := []CorpusFact{}
	for _, fact := range facts {
		if fact.Kind == kind {
			out = append(out, fact)
		}
	}
	return out
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
