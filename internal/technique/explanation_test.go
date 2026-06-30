package technique_test

import (
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/technique"
)

func TestBuildExplanationTurnsPortraitIntoDeterministicTechniqueNotes(t *testing.T) {
	portrait := &technique.Portrait{
		SchemaVersion: technique.SchemaVersion,
		SourcePath:    "fixtures/explain.aep",
		Fingerprint: technique.FingerprintSummary{
			CompCount:          2,
			LayerCount:         8,
			EffectCount:        3,
			TextAnimatorCount:  2,
			ShapeOperatorCount: 6,
			DependencyCount:    4,
			UnknownCount:       1,
			LayerRoleCounts:    map[string]int{"shape": 3, "text": 1, "controller": 1},
		},
		SignalLayers: []technique.SignalLayer{
			{CompName: "Main", LayerName: "Burst", Role: "shape", Score: 4, Signals: []string{"shape_operator:trim", "shape_operator:stroke"}},
			{CompName: "Main", LayerName: "Title", Role: "text", Score: 3, Signals: []string{"text_animator:position", "effect:keyframed"}},
			{CompName: "Main", LayerName: "Controller", Role: "controller", Score: 2, Signals: []string{"dependency:effect_param_layer"}},
			{CompName: "Main", LayerName: "BG", Role: "solid", Score: 1, Signals: []string{"effect:native"}},
		},
		Mechanisms: technique.MechanismSummary{
			EffectClassCounts:           map[string]int{"native": 2, "third_party": 1},
			EffectMatchCounts:           map[string]int{"ADBE Fill": 2, "Plugin Magic": 1},
			TextAnimatorKindCounts:      map[string]int{"position": 1, "opacity": 1},
			ShapeFamilyCounts:           map[string]int{"trim": 2, "stroke": 2, "star": 1},
			ReproducibilityCounts:       map[string]int{"native": 2, "third_party": 1},
			ThirdPartyEffectMatchCounts: map[string]int{"Plugin Magic": 1},
		},
		Graph: technique.GraphSummary{
			EdgeCount:      4,
			RelationCounts: map[string]int{"source": 2, "effect_param_layer": 1, "matte": 1},
		},
		TechniqueHints: []technique.TechniqueHint{
			{ID: "shape_operator_stack", Confidence: "medium", Signals: []string{"layer:Burst"}},
			{ID: "controller_rig", Confidence: "medium", Signals: []string{"controller:Controller"}},
			{ID: "plugin_dependent", Confidence: "high", Signals: []string{"effect:Plugin Magic"}},
		},
		Unknowns: technique.UnknownSummary{Count: 1},
	}

	explanation, err := technique.BuildExplanation(portrait)
	if err != nil {
		t.Fatalf("BuildExplanation: %v", err)
	}

	if explanation.SchemaVersion != technique.SchemaVersion || explanation.SourcePath != portrait.SourcePath {
		t.Fatalf("explanation identity = %+v", explanation)
	}
	if explanation.Portrait.Fingerprint.LayerCount != portrait.Fingerprint.LayerCount {
		t.Fatalf("embedded portrait = %+v", explanation.Portrait.Fingerprint)
	}
	if len(explanation.Overview) == 0 || !strings.Contains(explanation.Overview[0], "2 comps, 8 layers, 3 effects") {
		t.Fatalf("overview = %+v", explanation.Overview)
	}
	assertTechniqueNote(t, explanation, "controller_rig", "controller/null")
	assertTechniqueNote(t, explanation, "plugin_dependent", "Third-party")
	assertTechniqueNote(t, explanation, "shape_operator_stack", "trim")
	assertArchetype(t, explanation, "shape_system")
	assertArchetype(t, explanation, "effect_stack")
	assertArchetype(t, explanation, "controller_rig")
	assertArchetype(t, explanation, "plugin_dependent")
	assertPattern(t, explanation, "effect_controlled_shape_system")
	assertPattern(t, explanation, "plugin_dependent_effect_stack")
	assertPattern(t, explanation, "kinetic_text_system")
	assertRecreationStep(t, explanation, "structure", "source (2)")
	assertRecreationStep(t, explanation, "effects", "Plugin Magic (1)")
	assertRecreationStep(t, explanation, "unknowns", "unknown parsed items: 1")
	if len(explanation.TopSignalLayers) != 3 || explanation.TopSignalLayers[0].LayerName != "Burst" {
		t.Fatalf("top signal layers = %+v", explanation.TopSignalLayers)
	}
	if !containsText(explanation.ReproducibilityNotes, "third-party effects: 1") {
		t.Fatalf("reproducibility notes = %+v", explanation.ReproducibilityNotes)
	}
	if explanation.RecreationReadiness.Status != "needs_reverse_engineering" ||
		!containsText(explanation.RecreationReadiness.Blockers, "unknown parsed items: 1") ||
		!containsText(explanation.RecreationReadiness.Blockers, "Plugin Magic (1)") {
		t.Fatalf("readiness = %+v", explanation.RecreationReadiness)
	}
	if !containsText(explanation.UnknownNotes, "1 unknown") {
		t.Fatalf("unknown notes = %+v", explanation.UnknownNotes)
	}
}

func TestBuildExplanationAllowsEmptyPortrait(t *testing.T) {
	explanation, err := technique.BuildExplanation(&technique.Portrait{
		SchemaVersion: technique.SchemaVersion,
		SourcePath:    "empty.aep",
		Fingerprint: technique.FingerprintSummary{
			LayerRoleCounts: map[string]int{},
		},
		Mechanisms: technique.MechanismSummary{
			EffectClassCounts:           map[string]int{},
			EffectMatchCounts:           map[string]int{},
			TextAnimatorKindCounts:      map[string]int{},
			ShapeFamilyCounts:           map[string]int{},
			ReproducibilityCounts:       map[string]int{},
			ThirdPartyEffectMatchCounts: map[string]int{},
		},
		Graph: technique.GraphSummary{RelationCounts: map[string]int{}},
	})
	if err != nil {
		t.Fatalf("BuildExplanation: %v", err)
	}
	if len(explanation.Overview) == 0 || len(explanation.Techniques) != 0 || len(explanation.TopSignalLayers) != 0 {
		t.Fatalf("empty explanation = %+v", explanation)
	}
	if len(explanation.Archetypes) != 0 {
		t.Fatalf("empty archetypes = %+v", explanation.Archetypes)
	}
	if len(explanation.Patterns) != 0 {
		t.Fatalf("empty patterns = %+v", explanation.Patterns)
	}
	if len(explanation.RecreationSteps) != 0 {
		t.Fatalf("empty recreation steps = %+v", explanation.RecreationSteps)
	}
	if explanation.RecreationReadiness.Status != "analysis_ready" {
		t.Fatalf("empty readiness = %+v", explanation.RecreationReadiness)
	}
	if !containsText(explanation.UnknownNotes, "No unknown") {
		t.Fatalf("unknown notes = %+v", explanation.UnknownNotes)
	}
}

func TestBuildExplanationTextLayerOnlyStepDoesNotClaimZeroAnimators(t *testing.T) {
	explanation, err := technique.BuildExplanation(&technique.Portrait{
		SchemaVersion: technique.SchemaVersion,
		SourcePath:    "text-only.aep",
		Fingerprint: technique.FingerprintSummary{
			LayerCount:      1,
			TextLayerCount:  1,
			LayerRoleCounts: map[string]int{"text": 1},
		},
		Mechanisms: technique.MechanismSummary{
			EffectClassCounts:           map[string]int{},
			EffectMatchCounts:           map[string]int{},
			TextAnimatorKindCounts:      map[string]int{},
			ShapeFamilyCounts:           map[string]int{},
			ReproducibilityCounts:       map[string]int{},
			ThirdPartyEffectMatchCounts: map[string]int{},
		},
		Graph: technique.GraphSummary{RelationCounts: map[string]int{}},
	})
	if err != nil {
		t.Fatalf("BuildExplanation: %v", err)
	}
	for _, step := range explanation.RecreationSteps {
		if step.ID == "text" {
			if strings.Contains(step.Summary, "0 text animators") {
				t.Fatalf("text recreation step = %+v", step)
			}
			return
		}
	}
	t.Fatalf("text recreation step not found in %+v", explanation.RecreationSteps)
}

func assertTechniqueNote(t *testing.T, explanation *technique.Explanation, id, wantText string) {
	t.Helper()
	for _, note := range explanation.Techniques {
		if note.ID == id {
			if !strings.Contains(note.Summary, wantText) {
				t.Fatalf("technique %s summary = %q, want %q", id, note.Summary, wantText)
			}
			return
		}
	}
	t.Fatalf("technique %s not found in %+v", id, explanation.Techniques)
}

func containsText(values []string, want string) bool {
	for _, value := range values {
		if strings.Contains(value, want) {
			return true
		}
	}
	return false
}

func assertArchetype(t *testing.T, explanation *technique.Explanation, id string) {
	t.Helper()
	for _, archetype := range explanation.Archetypes {
		if archetype.ID == id {
			if archetype.Score <= 0 || archetype.Summary == "" {
				t.Fatalf("archetype %s = %+v", id, archetype)
			}
			return
		}
	}
	t.Fatalf("archetype %s not found in %+v", id, explanation.Archetypes)
}

func assertPattern(t *testing.T, explanation *technique.Explanation, id string) {
	t.Helper()
	for _, pattern := range explanation.Patterns {
		if pattern.ID == id {
			if pattern.Score <= 0 || pattern.Summary == "" || len(pattern.Archetypes) == 0 {
				t.Fatalf("pattern %s = %+v", id, pattern)
			}
			return
		}
	}
	t.Fatalf("pattern %s not found in %+v", id, explanation.Patterns)
}

func assertRecreationStep(t *testing.T, explanation *technique.Explanation, id, wantText string) {
	t.Helper()
	lastPriority := 0
	for _, step := range explanation.RecreationSteps {
		if step.Priority < lastPriority {
			t.Fatalf("recreation steps are not sorted by priority: %+v", explanation.RecreationSteps)
		}
		lastPriority = step.Priority
		if step.ID != id {
			continue
		}
		joined := strings.Join(append(append([]string{step.Summary}, step.Inputs...), append(step.Risks, step.Evidence...)...), " ")
		if !strings.Contains(joined, wantText) {
			t.Fatalf("recreation step %s = %+v, want text %q", id, step, wantText)
		}
		return
	}
	t.Fatalf("recreation step %s not found in %+v", id, explanation.RecreationSteps)
}
