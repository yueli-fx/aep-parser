package technique_test

import (
	"testing"

	"github.com/yueli-fx/aep-parser/internal/technique"
)

func TestBuildPortraitSummarizesMechanismsGraphSignalsAndHints(t *testing.T) {
	facts := &technique.FactSet{
		SchemaVersion: technique.SchemaVersion,
		SourcePath:    "fixtures/portrait.aep",
		Summary: technique.Summary{
			CompCount:       1,
			LayerCount:      5,
			EffectCount:     2,
			TextLayerCount:  1,
			ShapeLayerCount: 1,
			UnknownCount:    1,
			MainCompName:    "Main",
		},
		Layers: []technique.LayerFact{
			{CompName: "Main", ID: 10, Name: "Title", Role: "text", Type: "text", Path: pathRef(`layers[10]`, `Title`), Evidence: evidence()},
			{CompName: "Main", ID: 11, Name: "Burst", Role: "shape", Type: "shape", Path: pathRef(`layers[11]`, `Burst`), Evidence: evidence()},
			{CompName: "Main", ID: 12, Name: "Nested", Role: "precomp", Type: "precomp", Path: pathRef(`layers[12]`, `Nested`), Evidence: evidence()},
			{CompName: "Main", ID: 13, Name: "Controller", Role: "controller", Type: "null", Path: pathRef(`layers[13]`, `Controller`), Evidence: evidence()},
			{CompName: "Main", ID: 14, Name: "Plate", Role: "solid", Type: "solid", Path: pathRef(`layers[14]`, `Plate`), Evidence: evidence()},
		},
		Effects: []technique.EffectFact{
			{
				CompName:          "Main",
				LayerName:         "Title",
				MatchName:         "ADBE Gaussian Blur 2",
				DependencyClass:   "native",
				ChangedParamCount: 1,
				TunedParamCount:   1,
				HasExpression:     true,
				HasKeyframes:      true,
				HasLayerRef:       true,
				Path:              pathRef(`effects.blur`, `Blur`),
				Evidence:          evidence(),
			},
			{
				CompName:        "Main",
				LayerName:       "Plate",
				MatchName:       "Plugin Magic",
				DependencyClass: "third_party",
				Path:            pathRef(`effects.plugin`, `Plugin Magic`),
				Evidence:        evidence(),
			},
		},
		TextAnimators: []technique.TextAnimatorFact{
			{CompName: "Main", LayerName: "Title", PropertyKind: "position", MatchName: "ADBE Text Position 3D", HasKeyframes: true, Path: pathRef(`text.position`, `Position`), Evidence: evidence()},
			{CompName: "Main", LayerName: "Title", PropertyKind: "opacity", MatchName: "ADBE Text Opacity", HasStaticValue: true, Path: pathRef(`text.opacity`, `Opacity`), Evidence: evidence()},
		},
		ShapeOperators: []technique.ShapeOperatorFact{
			{CompName: "Main", LayerName: "Burst", Family: "star", Source: "shape", MatchName: "star", Path: pathRef(`shape.star`, `Star`), Evidence: evidence()},
			{CompName: "Main", LayerName: "Burst", Family: "stroke", Source: "property", MatchName: "ADBE Vector Graphic - Stroke", Path: pathRef(`shape.stroke`, `Stroke`), Evidence: evidence()},
			{CompName: "Main", LayerName: "Burst", Family: "trim", Source: "property", MatchName: "ADBE Vector Filter - Trim", Path: pathRef(`shape.trim`, `Trim`), Evidence: evidence()},
		},
		Dependencies: []technique.DependencyFact{
			{CompName: "Main", SourceName: "Nested", SourceID: 12, Relation: "source", TargetName: "Precomp", TargetKind: "composition", Path: pathRef(`dep.source`, `source`), Evidence: evidence()},
			{CompName: "Main", SourceName: "Burst", SourceID: 11, Relation: "matte", TargetName: "Title", TargetID: 10, Path: pathRef(`dep.matte`, `matte`), Evidence: evidence()},
			{CompName: "Main", SourceName: "Plate", SourceID: 14, Relation: "parent", TargetName: "Controller", TargetID: 13, Path: pathRef(`dep.parent`, `parent`), Evidence: evidence()},
			{CompName: "Main", SourceName: "Title", SourceID: 10, Relation: "effect_param_layer", TargetName: "Controller", TargetID: 13, Property: "ADBE Set Matte3-0001", Path: pathRef(`dep.effect`, `effect`), Evidence: evidence()},
		},
		Unknowns: []technique.UnknownFact{{Path: "unknown", Reason: "test", Evidence: evidence()}},
	}

	portrait, err := technique.BuildPortrait(facts)
	if err != nil {
		t.Fatalf("BuildPortrait: %v", err)
	}

	if portrait.SchemaVersion != technique.SchemaVersion || portrait.SourcePath != facts.SourcePath {
		t.Fatalf("portrait identity = %+v", portrait)
	}
	if portrait.Fingerprint.CompCount != 1 ||
		portrait.Fingerprint.LayerCount != 5 ||
		portrait.Fingerprint.EffectCount != 2 ||
		portrait.Fingerprint.TextAnimatorCount != 2 ||
		portrait.Fingerprint.ShapeOperatorCount != 3 ||
		portrait.Fingerprint.DependencyCount != 4 ||
		portrait.Fingerprint.UnknownCount != 1 {
		t.Fatalf("fingerprint = %+v", portrait.Fingerprint)
	}
	if portrait.Fingerprint.LayerRoleCounts["text"] != 1 || portrait.Fingerprint.LayerRoleCounts["shape"] != 1 || portrait.Fingerprint.LayerRoleCounts["controller"] != 1 {
		t.Fatalf("role counts = %+v", portrait.Fingerprint.LayerRoleCounts)
	}
	if portrait.Mechanisms.EffectClassCounts["native"] != 1 || portrait.Mechanisms.EffectClassCounts["third_party"] != 1 {
		t.Fatalf("effect class counts = %+v", portrait.Mechanisms.EffectClassCounts)
	}
	if portrait.Mechanisms.ReproducibilityCounts["native"] != 1 || portrait.Mechanisms.ReproducibilityCounts["third_party"] != 1 {
		t.Fatalf("reproducibility counts = %+v", portrait.Mechanisms.ReproducibilityCounts)
	}
	if portrait.Mechanisms.EffectMatchCounts["ADBE Gaussian Blur 2"] != 1 || portrait.Mechanisms.TextAnimatorKindCounts["position"] != 1 || portrait.Mechanisms.ShapeFamilyCounts["trim"] != 1 {
		t.Fatalf("mechanisms = %+v", portrait.Mechanisms)
	}
	if portrait.Graph.EdgeCount != 4 || portrait.Graph.RelationCounts["source"] != 1 || portrait.Graph.RelationCounts["matte"] != 1 || portrait.Graph.RelationCounts["effect_param_layer"] != 1 {
		t.Fatalf("graph = %+v", portrait.Graph)
	}
	assertSignalLayer(t, portrait, "Title", "effect:keyframed", "text_animator:position", "dependency:effect_param_layer")
	assertSignalLayer(t, portrait, "Burst", "shape_operator:trim", "dependency:matte")
	assertSignalLayer(t, portrait, "Nested", "dependency:source")
	assertSignalLayer(t, portrait, "Plate", "effect:third_party", "dependency:parent")
	assertHint(t, portrait, "kinetic_text")
	assertHint(t, portrait, "shape_operator_stack")
	assertHint(t, portrait, "precomp_assembly")
	assertHint(t, portrait, "effect_driven_layer")
	assertHint(t, portrait, "matte_composite")
	assertHint(t, portrait, "controller_rig")
	assertHint(t, portrait, "plugin_dependent")
	if portrait.Unknowns.Count != 1 {
		t.Fatalf("unknown summary = %+v", portrait.Unknowns)
	}
}

func TestBuildPortraitAllowsEmptyFactSet(t *testing.T) {
	portrait, err := technique.BuildPortrait(&technique.FactSet{
		SchemaVersion: technique.SchemaVersion,
		SourcePath:    "empty.aep",
	})
	if err != nil {
		t.Fatalf("BuildPortrait: %v", err)
	}
	if portrait.SchemaVersion != technique.SchemaVersion || portrait.SourcePath != "empty.aep" {
		t.Fatalf("portrait = %+v", portrait)
	}
	if portrait.Fingerprint.LayerCount != 0 || len(portrait.SignalLayers) != 0 || len(portrait.TechniqueHints) != 0 {
		t.Fatalf("empty portrait = %+v", portrait)
	}
	if portrait.Fingerprint.LayerRoleCounts == nil ||
		portrait.Mechanisms.EffectClassCounts == nil ||
		portrait.Mechanisms.EffectMatchCounts == nil ||
		portrait.Mechanisms.TextAnimatorKindCounts == nil ||
		portrait.Mechanisms.ShapeFamilyCounts == nil ||
		portrait.Mechanisms.ReproducibilityCounts == nil ||
		portrait.Graph.RelationCounts == nil {
		t.Fatalf("empty portrait has nil maps: %+v", portrait)
	}
}

func assertSignalLayer(t *testing.T, portrait *technique.Portrait, layerName string, signals ...string) {
	t.Helper()
	for _, layer := range portrait.SignalLayers {
		if layer.LayerName != layerName {
			continue
		}
		for _, signal := range signals {
			if !contains(layer.Signals, signal) {
				t.Fatalf("signal layer %s = %+v, missing %q", layerName, layer, signal)
			}
		}
		return
	}
	t.Fatalf("signal layer %s not found in %+v", layerName, portrait.SignalLayers)
}

func assertHint(t *testing.T, portrait *technique.Portrait, id string) {
	t.Helper()
	for _, hint := range portrait.TechniqueHints {
		if hint.ID == id {
			return
		}
	}
	t.Fatalf("hint %s not found in %+v", id, portrait.TechniqueHints)
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
