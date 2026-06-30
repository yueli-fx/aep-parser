package technique_test

import (
	"testing"

	"github.com/yueli-fx/aep-parser/internal/profile"
	"github.com/yueli-fx/aep-parser/internal/technique"
)

func TestBuildSummarizesCompsLayersEffectsDependenciesAndUnknowns(t *testing.T) {
	enabled := true
	prof := &profile.Profile{
		SchemaVersion: profile.SchemaVersion,
		Meta: profile.Meta{
			Path: "fixtures/example.aep",
		},
		Unknowns: []profile.Unknown{{
			Path:   `comps.by_name["Main"].layers[0].mystery`,
			Reason: "unsupported test field",
			Evidence: profile.Evidence{
				Level:      profile.EvidenceL1Parsed,
				Source:     "profile",
				Confidence: "low",
			},
		}},
		Comps: []profile.Composition{
			{
				ID:        1,
				Name:      "Main",
				Width:     1920,
				Height:    1080,
				FrameRate: 30,
				Duration:  5,
				Path:      pathRef(`comps.by_id[1]`, `Main`),
				Evidence:  evidence(),
				Layers: []profile.Layer{
					{
						ID:    10,
						Index: 1,
						Name:  "Title",
						Type:  "text",
						Path:  pathRef(`comps.by_id[1].layers.by_id[10]`, `Main/Title`),
						Effects: []profile.Effect{{
							MatchName:       "ADBE Gaussian Blur 2",
							DisplayName:     "Gaussian Blur",
							DependencyClass: "native",
							Occurrence:      1,
							TunedParams:     []string{"ADBE Gaussian Blur 2-0001"},
							Path:            pathRef(`effects[0]`, `Gaussian Blur`),
							Evidence:        evidence(),
							Params: []profile.Property{
								{
									MatchName:         "ADBE Gaussian Blur 2-0001",
									Changed:           true,
									Expression:        "time * 5",
									ExpressionEnabled: &enabled,
									Keyframes: []profile.Keyframe{
										{Time: 0, Value: 0},
										{Time: 1, Value: 15},
									},
									Path:     pathRef(`params[0]`, `Blurriness`),
									Evidence: evidence(),
								},
								{
									MatchName: "ADBE Set Matte3-0001",
									LayerRef:  &profile.LayerRef{ID: 11, Index: 2, Name: "Matte"},
									Path:      pathRef(`params[1]`, `Take Matte From Layer`),
									Evidence:  evidence(),
								},
							},
							UnknownParams: []profile.Unknown{{
								Path:     `effects[0].params[99]`,
								Reason:   "unknown effect param",
								Evidence: evidence(),
							}},
						}},
						Evidence: evidence(),
					},
					{
						ID:        11,
						Index:     2,
						Name:      "Matte",
						Type:      "shape",
						ParentRef: &profile.LayerRef{ID: 10, Index: 1, Name: "Title"},
						Path:      pathRef(`comps.by_id[1].layers.by_id[11]`, `Main/Matte`),
						Evidence:  evidence(),
					},
					{
						ID:        12,
						Index:     3,
						Name:      "Nested",
						Type:      "precomp",
						SourceRef: &profile.ItemRef{ID: 2, Kind: "composition", Name: "Precomp"},
						MatteRef:  &profile.LayerRef{ID: 11, Index: 2, Name: "Matte"},
						Path:      pathRef(`comps.by_id[1].layers.by_id[12]`, `Main/Nested`),
						Evidence:  evidence(),
					},
					{
						ID:        13,
						Index:     4,
						Name:      "Solid",
						Type:      "solid",
						SourceRef: &profile.ItemRef{ID: 100, Kind: "footage", Name: "Dark Blue Solid"},
						Path:      pathRef(`comps.by_id[1].layers.by_id[13]`, `Main/Solid`),
						Evidence:  evidence(),
					},
					{
						ID:             14,
						Index:          5,
						Name:           "Environment",
						Type:           "light",
						LightSourceRef: &profile.LayerRef{ID: 12, Index: 3, Name: "Nested"},
						Path:           pathRef(`comps.by_id[1].layers.by_id[14]`, `Main/Environment`),
						Evidence:       evidence(),
					},
					{
						ID:       15,
						Index:    6,
						Name:     "Controller",
						Type:     "null",
						Flags:    profile.LayerFlags{IsNull: true},
						Path:     pathRef(`comps.by_id[1].layers.by_id[15]`, `Main/Controller`),
						Evidence: evidence(),
					},
				},
			},
			{
				ID:        2,
				Name:      "Precomp",
				Width:     640,
				Height:    360,
				FrameRate: 24,
				Duration:  2,
				Path:      pathRef(`comps.by_id[2]`, `Precomp`),
				Evidence:  evidence(),
			},
		},
	}

	facts, err := technique.Build(prof)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if facts.SchemaVersion != 1 {
		t.Fatalf("SchemaVersion = %d, want 1", facts.SchemaVersion)
	}
	if facts.SourcePath != "fixtures/example.aep" {
		t.Fatalf("SourcePath = %q, want fixtures/example.aep", facts.SourcePath)
	}
	if facts.Summary.CompCount != 2 || facts.Summary.LayerCount != 6 || facts.Summary.EffectCount != 1 {
		t.Fatalf("Summary = %+v, want comp/layer/effect counts", facts.Summary)
	}
	if facts.Summary.MainCompName != "Main" {
		t.Fatalf("MainCompName = %q, want Main", facts.Summary.MainCompName)
	}
	if len(facts.Comps) != 2 {
		t.Fatalf("Comps = %d, want 2", len(facts.Comps))
	}
	if facts.Comps[0].MainCandidate != true || facts.Comps[0].LayerCount != 6 {
		t.Fatalf("main comp fact = %+v", facts.Comps[0])
	}
	assertLayerRole(t, facts, "Title", "text", "high")
	assertLayerRole(t, facts, "Matte", "shape", "high")
	assertLayerRole(t, facts, "Nested", "precomp", "high")
	assertLayerRole(t, facts, "Solid", "solid", "high")
	assertLayerRole(t, facts, "Controller", "controller", "medium")

	if len(facts.Effects) != 1 {
		t.Fatalf("Effects = %d, want 1", len(facts.Effects))
	}
	effect := facts.Effects[0]
	if effect.MatchName != "ADBE Gaussian Blur 2" || effect.ChangedParamCount != 1 || !effect.HasExpression || !effect.HasKeyframes || !effect.HasLayerRef || effect.UnknownParamCount != 1 {
		t.Fatalf("effect fact = %+v", effect)
	}

	assertDependency(t, facts, "source", "Nested", "Precomp")
	assertDependency(t, facts, "source", "Solid", "Dark Blue Solid")
	assertDependency(t, facts, "parent", "Matte", "Title")
	assertDependency(t, facts, "matte", "Nested", "Matte")
	assertDependency(t, facts, "light_source", "Environment", "Nested")
	assertDependency(t, facts, "effect_param_layer", "Title", "Matte")

	if facts.Summary.UnknownCount != 2 || len(facts.Unknowns) != 2 {
		t.Fatalf("unknowns summary=%+v unknowns=%+v, want 2", facts.Summary, facts.Unknowns)
	}
}

func TestBuildClassifiesTextAnimatorsAndShapeOperators(t *testing.T) {
	prof := &profile.Profile{
		SchemaVersion: profile.SchemaVersion,
		Comps: []profile.Composition{{
			ID:       1,
			Name:     "Main",
			Path:     pathRef(`comps.by_id[1]`, `Main`),
			Evidence: evidence(),
			Layers: []profile.Layer{
				{
					ID:    10,
					Index: 1,
					Name:  "Headline",
					Type:  "text",
					Path:  pathRef(`comps.by_id[1].layers.by_id[10]`, `Main/Headline`),
					Properties: []profile.Property{
						{
							Name:        "Opacity",
							MatchName:   "ADBE Text Opacity",
							StaticValue: 30,
							Path:        pathRef(`text.props.opacity`, `Text Opacity`),
							Evidence:    evidence(),
						},
						{
							Name:      "Position",
							MatchName: "ADBE Text Position 3D",
							Keyframes: []profile.Keyframe{{Time: 0, Value: []float64{0, 0, 0}}},
							Path:      pathRef(`text.props.position`, `Text Position`),
							Evidence:  evidence(),
						},
						{
							Name:       "Fill Color",
							MatchName:  "ADBE Text Fill Color",
							Expression: "time > 1 ? [1,0,0,1] : value",
							Path:       pathRef(`text.props.fill`, `Text Fill`),
							Evidence:   evidence(),
						},
					},
				},
				{
					ID:    11,
					Index: 2,
					Name:  "Burst",
					Type:  "shape",
					Path:  pathRef(`comps.by_id[1].layers.by_id[11]`, `Main/Burst`),
					Shapes: []profile.Shape{
						{
							Kind:     "star",
							Name:     "Star 1",
							Path:     pathRef(`shape.star`, `Star 1`),
							Evidence: evidence(),
							Properties: []profile.Property{
								{
									Name:      "Points",
									MatchName: "ADBE Vector Star Points",
									Path:      pathRef(`shape.star.points`, `Star Points`),
									Evidence:  evidence(),
								},
							},
						},
						{
							Kind:     "path",
							Name:     "Path 1",
							Path:     pathRef(`shape.path`, `Path 1`),
							Evidence: evidence(),
						},
					},
					Properties: []profile.Property{
						{
							Name:      "Stroke",
							MatchName: "ADBE Vector Graphic - Stroke",
							Path:      pathRef(`shape.stroke`, `Stroke`),
							Evidence:  evidence(),
						},
						{
							Name:      "Trim Paths",
							MatchName: "ADBE Vector Filter - Trim",
							Path:      pathRef(`shape.trim`, `Trim Paths`),
							Evidence:  evidence(),
						},
						{
							Name:      "Repeater",
							MatchName: "ADBE Vector Filter - Repeater",
							Path:      pathRef(`shape.repeater`, `Repeater`),
							Evidence:  evidence(),
						},
					},
				},
			},
		}},
	}

	facts, err := technique.Build(prof)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	assertTextAnimator(t, facts, "Headline", "ADBE Text Opacity", "opacity", true, false, false)
	assertTextAnimator(t, facts, "Headline", "ADBE Text Position 3D", "position", false, false, true)
	assertTextAnimator(t, facts, "Headline", "ADBE Text Fill Color", "fill_color", false, true, false)
	assertShapeOperator(t, facts, "Burst", "star", "shape", "star")
	assertShapeOperator(t, facts, "Burst", "path", "shape", "path")
	assertShapeOperator(t, facts, "Burst", "stroke", "property", "ADBE Vector Graphic - Stroke")
	assertShapeOperator(t, facts, "Burst", "trim", "property", "ADBE Vector Filter - Trim")
	assertShapeOperator(t, facts, "Burst", "repeater", "property", "ADBE Vector Filter - Repeater")
}

func pathRef(path, display string) profile.PathRef {
	return profile.PathRef{Path: path, DisplayPath: display}
}

func evidence() profile.Evidence {
	return profile.Evidence{Level: profile.EvidenceL1Parsed, Source: "profile", Confidence: "high"}
}

func assertLayerRole(t *testing.T, facts *technique.FactSet, name, role, confidence string) {
	t.Helper()
	for _, layer := range facts.Layers {
		if layer.Name == name {
			if layer.Role != role || layer.Confidence != confidence {
				t.Fatalf("layer %q role/confidence = %s/%s, want %s/%s", name, layer.Role, layer.Confidence, role, confidence)
			}
			return
		}
	}
	t.Fatalf("layer %q not found in %+v", name, facts.Layers)
}

func assertDependency(t *testing.T, facts *technique.FactSet, relation, source, target string) {
	t.Helper()
	for _, dep := range facts.Dependencies {
		if dep.Relation == relation && dep.SourceName == source && dep.TargetName == target {
			return
		}
	}
	t.Fatalf("dependency %s %s -> %s not found in %+v", relation, source, target, facts.Dependencies)
}

func assertTextAnimator(t *testing.T, facts *technique.FactSet, layerName, matchName, kind string, staticValue, expression, keyframes bool) {
	t.Helper()
	for _, fact := range facts.TextAnimators {
		if fact.LayerName == layerName && fact.MatchName == matchName {
			if fact.PropertyKind != kind || fact.HasStaticValue != staticValue || fact.HasExpression != expression || fact.HasKeyframes != keyframes {
				t.Fatalf("text animator %s = %+v, want kind/static/expression/keyframes %s/%v/%v/%v", matchName, fact, kind, staticValue, expression, keyframes)
			}
			return
		}
	}
	t.Fatalf("text animator %s on %s not found in %+v", matchName, layerName, facts.TextAnimators)
}

func assertShapeOperator(t *testing.T, facts *technique.FactSet, layerName, family, source, matchName string) {
	t.Helper()
	for _, fact := range facts.ShapeOperators {
		if fact.LayerName == layerName && fact.Family == family && fact.Source == source && fact.MatchName == matchName {
			return
		}
	}
	t.Fatalf("shape operator %s/%s/%s on %s not found in %+v", family, source, matchName, layerName, facts.ShapeOperators)
}
