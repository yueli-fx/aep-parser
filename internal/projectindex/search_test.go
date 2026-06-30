package projectindex

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestSearchLayersBySourceIDReturnsStableLayerHits(t *testing.T) {
	layerA := &aep.Layer{ID: 11, Index: 0, Name: "Title", SourceID: 100}
	layerB := &aep.Layer{ID: 12, Index: 2, Name: "BG", SourceID: 200}
	layerC := &aep.Layer{ID: 13, Index: 1, Name: "Overlay", SourceID: 100}
	project := &aep.Project{
		Compositions: []*aep.Composition{
			{ID: 7, Name: "Main", Layers: []*aep.Layer{layerA, layerB}},
			{ID: 8, Name: "Precomp", Layers: []*aep.Layer{layerC}},
		},
		Footage: []*aep.Footage{{ID: 100, Name: "Logo.png"}},
	}

	hits := Build(project).SearchLayersBySourceID(100)

	if len(hits) != 2 {
		t.Fatalf("len(hits) = %d, want 2", len(hits))
	}
	assertHit(t, hits[0], HitLayer, "source_id", "100", Location{
		ItemID:     100,
		ItemKind:   "footage",
		ItemName:   "Logo.png",
		CompID:     7,
		CompName:   "Main",
		LayerID:    11,
		LayerIndex: 0,
		LayerName:  "Title",
	})
	assertHit(t, hits[1], HitLayer, "source_id", "100", Location{
		ItemID:     100,
		ItemKind:   "footage",
		ItemName:   "Logo.png",
		CompID:     8,
		CompName:   "Precomp",
		LayerID:    13,
		LayerIndex: 1,
		LayerName:  "Overlay",
	})
	if hits[0].Pointers.Comp != project.Compositions[0] || hits[0].Pointers.Layer != layerA {
		t.Fatalf("first hit pointers = %+v, want comp/layer pointers", hits[0].Pointers)
	}
	if hits[1].Pointers.Comp != project.Compositions[1] || hits[1].Pointers.Layer != layerC {
		t.Fatalf("second hit pointers = %+v, want comp/layer pointers", hits[1].Pointers)
	}
	if hits[0].Pointers.Item != project.Footage[0] || hits[1].Pointers.Item != project.Footage[0] {
		t.Fatalf("source hit item pointers = %p/%p, want %p", hits[0].Pointers.Item, hits[1].Pointers.Item, project.Footage[0])
	}
}

func TestSearchEffectsByMatchNameReturnsEffectOccurrenceHits(t *testing.T) {
	glowA := &aep.Effect{MatchName: "ADBE Glo2", Name: "Glow A"}
	blur := &aep.Effect{MatchName: "ADBE Gaussian Blur 2", Name: "Gaussian Blur"}
	glowA2 := &aep.Effect{MatchName: "ADBE Glo2", Name: "Glow A Duplicate"}
	glowB := &aep.Effect{MatchName: "ADBE Glo2", Name: "Glow B"}
	layerA := &aep.Layer{ID: 21, Index: 0, Name: "Title", Effects: []*aep.Effect{glowA, blur, glowA2}}
	layerB := &aep.Layer{ID: 22, Index: 1, Name: "BG", Effects: []*aep.Effect{glowB}}
	project := &aep.Project{
		Compositions: []*aep.Composition{
			{ID: 9, Name: "Main", Layers: []*aep.Layer{layerA, nil, layerB}},
		},
	}

	hits := Build(project).SearchEffectsByMatchName("ADBE Glo2")

	if len(hits) != 3 {
		t.Fatalf("len(hits) = %d, want 3", len(hits))
	}
	assertHit(t, hits[0], HitEffect, "effect.match_name", "ADBE Glo2", Location{
		CompID:           9,
		CompName:         "Main",
		LayerID:          21,
		LayerIndex:       0,
		LayerName:        "Title",
		EffectMatchName:  "ADBE Glo2",
		EffectName:       "Glow A",
		EffectOccurrence: 1,
	})
	assertHit(t, hits[1], HitEffect, "effect.match_name", "ADBE Glo2", Location{
		CompID:           9,
		CompName:         "Main",
		LayerID:          21,
		LayerIndex:       0,
		LayerName:        "Title",
		EffectMatchName:  "ADBE Glo2",
		EffectName:       "Glow A Duplicate",
		EffectOccurrence: 2,
	})
	assertHit(t, hits[2], HitEffect, "effect.match_name", "ADBE Glo2", Location{
		CompID:           9,
		CompName:         "Main",
		LayerID:          22,
		LayerIndex:       1,
		LayerName:        "BG",
		EffectMatchName:  "ADBE Glo2",
		EffectName:       "Glow B",
		EffectOccurrence: 1,
	})
	if hits[0].Pointers.Effect != glowA || hits[1].Pointers.Effect != glowA2 || hits[2].Pointers.Effect != glowB {
		t.Fatalf("effect pointers = %p/%p/%p, want %p/%p/%p", hits[0].Pointers.Effect, hits[1].Pointers.Effect, hits[2].Pointers.Effect, glowA, glowA2, glowB)
	}
}

func TestSearchPropertiesByMatchNameReturnsLayerAndEffectParamHits(t *testing.T) {
	layerOpacity := &aep.Property{MatchName: "ADBE Opacity", Name: "Opacity"}
	layerOpacity2 := &aep.Property{MatchName: "ADBE Opacity", Name: "Opacity Copy"}
	blurAmount := &aep.Property{MatchName: "ADBE Gaussian Blur 2-0001", Name: "Blurriness"}
	blur := &aep.Effect{
		MatchName:  "ADBE Gaussian Blur 2",
		Name:       "Gaussian Blur",
		Parameters: []*aep.Property{blurAmount},
	}
	layer := &aep.Layer{
		ID:         31,
		Index:      0,
		Name:       "Title",
		Properties: []*aep.Property{layerOpacity, layerOpacity2},
		Effects:    []*aep.Effect{blur},
	}
	project := &aep.Project{
		Compositions: []*aep.Composition{{ID: 10, Name: "Main", Layers: []*aep.Layer{layer}}},
	}

	opacityHits := Build(project).SearchPropertiesByMatchName("ADBE Opacity")
	if len(opacityHits) != 2 {
		t.Fatalf("len(opacityHits) = %d, want 2", len(opacityHits))
	}
	assertHit(t, opacityHits[0], HitProperty, "property.match_name", "ADBE Opacity", Location{
		CompID:            10,
		CompName:          "Main",
		LayerID:           31,
		LayerIndex:        0,
		LayerName:         "Title",
		PropertyMatchName: "ADBE Opacity",
		PropertyName:      "Opacity",
		PropertyPath:      "layers[].properties[1]",
	})
	assertHit(t, opacityHits[1], HitProperty, "property.match_name", "ADBE Opacity", Location{
		CompID:            10,
		CompName:          "Main",
		LayerID:           31,
		LayerIndex:        0,
		LayerName:         "Title",
		PropertyMatchName: "ADBE Opacity",
		PropertyName:      "Opacity Copy",
		PropertyPath:      "layers[].properties[2]",
	})
	if opacityHits[0].Pointers.Property != layerOpacity || opacityHits[1].Pointers.Property != layerOpacity2 {
		t.Fatalf("opacity property pointers = %p/%p, want %p/%p", opacityHits[0].Pointers.Property, opacityHits[1].Pointers.Property, layerOpacity, layerOpacity2)
	}

	paramHits := Build(project).SearchPropertiesByMatchName("ADBE Gaussian Blur 2-0001")
	if len(paramHits) != 1 {
		t.Fatalf("len(paramHits) = %d, want 1", len(paramHits))
	}
	assertHit(t, paramHits[0], HitProperty, "property.match_name", "ADBE Gaussian Blur 2-0001", Location{
		CompID:            10,
		CompName:          "Main",
		LayerID:           31,
		LayerIndex:        0,
		LayerName:         "Title",
		EffectMatchName:   "ADBE Gaussian Blur 2",
		EffectName:        "Gaussian Blur",
		EffectOccurrence:  1,
		PropertyMatchName: "ADBE Gaussian Blur 2-0001",
		PropertyName:      "Blurriness",
		PropertyPath:      "layers[].effects[1].params[1]",
	})
	if paramHits[0].Pointers.Effect != blur || paramHits[0].Pointers.Property != blurAmount {
		t.Fatalf("param hit pointers = effect %p property %p, want %p/%p", paramHits[0].Pointers.Effect, paramHits[0].Pointers.Property, blur, blurAmount)
	}
}

func TestSearchExpressionsContainingReturnsLayerAndEffectParamHits(t *testing.T) {
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
	blur := &aep.Effect{
		MatchName:  "ADBE Gaussian Blur 2",
		Name:       "Gaussian Blur",
		Parameters: []*aep.Property{blurAmount},
	}
	layer := &aep.Layer{
		ID:         32,
		Index:      0,
		Name:       "Title",
		Properties: []*aep.Property{position},
		Effects:    []*aep.Effect{blur},
	}
	project := &aep.Project{
		Compositions: []*aep.Composition{{ID: 11, Name: "Main", Layers: []*aep.Layer{layer}}},
	}

	wiggleHits := Build(project).SearchExpressionsContaining("wiggle")
	if len(wiggleHits) != 1 {
		t.Fatalf("len(wiggleHits) = %d, want 1", len(wiggleHits))
	}
	assertHit(t, wiggleHits[0], HitExpression, "property.expression", "wiggle(2, 20)", Location{
		CompID:            11,
		CompName:          "Main",
		LayerID:           32,
		LayerIndex:        0,
		LayerName:         "Title",
		PropertyMatchName: "ADBE Position",
		PropertyName:      "Position",
		PropertyPath:      "layers[].properties[1]",
	})

	timeHits := Build(project).SearchExpressionsContaining("time")
	if len(timeHits) != 1 {
		t.Fatalf("len(timeHits) = %d, want 1", len(timeHits))
	}
	assertHit(t, timeHits[0], HitExpression, "property.expression", "time * 12", Location{
		CompID:            11,
		CompName:          "Main",
		LayerID:           32,
		LayerIndex:        0,
		LayerName:         "Title",
		EffectMatchName:   "ADBE Gaussian Blur 2",
		EffectName:        "Gaussian Blur",
		EffectOccurrence:  1,
		PropertyMatchName: "ADBE Gaussian Blur 2-0001",
		PropertyName:      "Blurriness",
		PropertyPath:      "layers[].effects[1].params[1]",
	})
	if timeHits[0].Pointers.Effect != blur || timeHits[0].Pointers.Property != blurAmount {
		t.Fatalf("expression param pointers = %p/%p, want %p/%p", timeHits[0].Pointers.Effect, timeHits[0].Pointers.Property, blur, blurAmount)
	}
}

func TestSearchNilOrInvalidInputReturnsNoHits(t *testing.T) {
	var idx *Index
	assertNoHits(t, "nil SearchLayersBySourceID", idx.SearchLayersBySourceID(100))
	assertNoHits(t, "nil SearchEffectsByMatchName", idx.SearchEffectsByMatchName("ADBE Glo2"))
	assertNoHits(t, "nil SearchPropertiesByMatchName", idx.SearchPropertiesByMatchName("ADBE Opacity"))
	assertNoHits(t, "nil SearchExpressionsContaining", idx.SearchExpressionsContaining("time"))

	empty := Build(nil)
	assertNoHits(t, "empty SearchLayersBySourceID", empty.SearchLayersBySourceID(100))
	assertNoHits(t, "empty SearchEffectsByMatchName", empty.SearchEffectsByMatchName("ADBE Glo2"))
	assertNoHits(t, "empty SearchPropertiesByMatchName", empty.SearchPropertiesByMatchName("ADBE Opacity"))
	assertNoHits(t, "empty SearchExpressionsContaining", empty.SearchExpressionsContaining("time"))
	assertNoHits(t, "zero source ID", empty.SearchLayersBySourceID(0))
	assertNoHits(t, "empty effect match name", empty.SearchEffectsByMatchName(""))
	assertNoHits(t, "empty property match name", empty.SearchPropertiesByMatchName(""))
	assertNoHits(t, "empty expression query", empty.SearchExpressionsContaining(""))
}

func TestSearchHitJSONUsesStableSchemaNames(t *testing.T) {
	hit := Hit{
		Kind:  HitEffect,
		Match: Match{Field: "effect.match_name", Value: "ADBE Glo2"},
		Location: Location{
			CompID:           9,
			CompName:         "Main",
			LayerID:          21,
			LayerIndex:       0,
			LayerName:        "Title",
			EffectMatchName:  "ADBE Glo2",
			EffectOccurrence: 1,
		},
	}

	data, err := json.Marshal(hit)
	if err != nil {
		t.Fatalf("json.Marshal(Hit): %v", err)
	}
	got := string(data)
	for _, want := range []string{
		`"kind":"effect"`,
		`"match":{"field":"effect.match_name","value":"ADBE Glo2"}`,
		`"location":`,
		`"comp_id":9`,
		`"layer_name":"Title"`,
		`"effect_match_name":"ADBE Glo2"`,
		`"effect_occurrence":1`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("JSON %s missing %s", got, want)
		}
	}
	if strings.Contains(got, "Pointers") || strings.Contains(got, "pointers") {
		t.Fatalf("JSON %s should not include pointers by default", got)
	}
}

func assertHit(t *testing.T, got Hit, kind HitKind, field, value string, location Location) {
	t.Helper()
	if got.Kind != kind {
		t.Fatalf("Kind = %q, want %q", got.Kind, kind)
	}
	if got.Match.Field != field || got.Match.Value != value {
		t.Fatalf("Match = %+v, want field=%q value=%q", got.Match, field, value)
	}
	if got.Location != location {
		t.Fatalf("Location = %+v, want %+v", got.Location, location)
	}
}

func assertNoHits(t *testing.T, name string, got []Hit) {
	t.Helper()
	if len(got) != 0 {
		t.Fatalf("%s returned %d hit(s), want zero", name, len(got))
	}
}
