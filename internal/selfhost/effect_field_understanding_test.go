package selfhost

import "testing"

func TestBuildEffectFieldUnderstandingClassifiesReproducibility(t *testing.T) {
	inventory := EffectFieldInventory{
		SchemaVersion: 1,
		OutputPath:    "tmp/effect_field_inventory/inventory.json",
		Summary: EffectFieldInventorySummary{
			EffectKinds:       4,
			EffectOccurrences: 18,
			ParamKinds:        15,
			ParamOccurrences:  40,
		},
		Effects: []EffectFieldInventoryEffect{
			{
				MatchName:            "ADBE Slider Control",
				Class:                "native_supported",
				Occurrences:          3,
				ParamKinds:           2,
				ParamOccurrences:     6,
				InferredCapabilities: []string{"scalar", "slider"},
			},
			{
				MatchName:            "Pseudo/Controller",
				Class:                "pseudo",
				Occurrences:          4,
				ParamKinds:           3,
				ParamOccurrences:     12,
				InferredCapabilities: []string{"color", "layer_ref", "scalar"},
			},
			{
				MatchName:            "tc Particular",
				Class:                "third_party",
				Occurrences:          2,
				ParamKinds:           8,
				ParamOccurrences:     20,
				InferredCapabilities: []string{"color", "expression", "keyframes", "position", "scalar"},
			},
			{
				MatchName:            "Mystery FX",
				Class:                "unknown_or_alias",
				Occurrences:          9,
				ParamKinds:           2,
				ParamOccurrences:     2,
				InferredCapabilities: []string{"scalar"},
			},
		},
	}

	understanding := BuildEffectFieldUnderstanding(inventory, "tmp/effect_field_inventory/inventory.json")

	if understanding.SchemaVersion != 1 {
		t.Fatalf("schema version = %d, want 1", understanding.SchemaVersion)
	}
	if understanding.SourceInventory != "tmp/effect_field_inventory/inventory.json" {
		t.Fatalf("source inventory = %q", understanding.SourceInventory)
	}
	if understanding.Summary.EffectKinds != 4 || understanding.Summary.ParamKinds != 15 {
		t.Fatalf("summary = %+v", understanding.Summary)
	}
	assertUnderstandingEffect(t, understanding, "ADBE Slider Control", "go_native_exact", "emit_native", "native_effect_can_be_emitted_by_go", "")
	assertUnderstandingEffect(t, understanding, "Pseudo/Controller", "pseudo_rebuildable", "rebuild_pseudo_controls", "param_names_and_values_parseable", "Pseudo effect fields are parseable controls")
	assertUnderstandingEffect(t, understanding, "tc Particular", "third_party_plugin_required", "preserve_as_dependency", "param_names_and_values_parseable", "Rendering equivalence requires the plugin")
	assertUnderstandingEffect(t, understanding, "Mystery FX", "unknown_or_alias_pending", "classify_or_alias", "class_or_alias_needs_review", "Effect class or alias needs review")
	assertSummaryHasCount(t, understanding.Summary.Reproducibility, "third_party_plugin_required", 2)
	assertSummaryHasCount(t, understanding.Summary.GenerationPolicies, "rebuild_pseudo_controls", 4)
	assertSummaryHasCount(t, understanding.Summary.StudyActions, "map_controller_inputs", 4)
}

func TestBuildEffectFieldUnderstandingSortsStudyPriority(t *testing.T) {
	inventory := EffectFieldInventory{
		Effects: []EffectFieldInventoryEffect{
			{
				MatchName:            "ADBE Fill",
				Class:                "native_supported",
				Occurrences:          100,
				ParamKinds:           2,
				InferredCapabilities: []string{"color", "scalar"},
			},
			{
				MatchName:            "Pseudo/Controller",
				Class:                "pseudo",
				Occurrences:          8,
				ParamKinds:           10,
				InferredCapabilities: []string{"expression", "keyframes", "layer_ref"},
			},
			{
				MatchName:            "tc Particular",
				Class:                "third_party",
				Occurrences:          8,
				ParamKinds:           300,
				InferredCapabilities: []string{"color", "expression", "keyframes", "layer_ref", "position"},
			},
		},
	}

	understanding := BuildEffectFieldUnderstanding(inventory, "inventory.json")

	if len(understanding.Effects) != 3 {
		t.Fatalf("effects = %d, want 3", len(understanding.Effects))
	}
	if understanding.Effects[0].MatchName != "ADBE Fill" {
		t.Fatalf("first effect = %q, want ADBE Fill", understanding.Effects[0].MatchName)
	}
	if understanding.Effects[1].MatchName != "tc Particular" {
		t.Fatalf("second effect = %q, want tc Particular", understanding.Effects[1].MatchName)
	}
	if understanding.Effects[1].StudyPriority <= understanding.Effects[2].StudyPriority {
		t.Fatalf("third-party priority %d should exceed pseudo priority %d", understanding.Effects[1].StudyPriority, understanding.Effects[2].StudyPriority)
	}
}

func assertUnderstandingEffect(t *testing.T, understanding EffectFieldUnderstanding, matchName, reproducibility, generationPolicy, fieldUnderstanding, boundarySubstring string) {
	t.Helper()
	for _, effect := range understanding.Effects {
		if effect.MatchName != matchName {
			continue
		}
		if effect.Reproducibility != reproducibility || effect.GenerationPolicy != generationPolicy || effect.FieldUnderstanding != fieldUnderstanding {
			t.Fatalf("%s effect = %+v", matchName, effect)
		}
		if boundarySubstring != "" && !containsText(effect.Boundary, boundarySubstring) {
			t.Fatalf("%s boundary = %q, want substring %q", matchName, effect.Boundary, boundarySubstring)
		}
		if effect.StudyPriority <= 0 {
			t.Fatalf("%s study priority = %d, want positive", matchName, effect.StudyPriority)
		}
		return
	}
	t.Fatalf("effect %q missing from %+v", matchName, understanding.Effects)
}

func assertSummaryHasCount(t *testing.T, rows []CountRow, name string, count int) {
	t.Helper()
	for _, row := range rows {
		if row.Name == name {
			if row.Count != count {
				t.Fatalf("%s count = %d, want %d", name, row.Count, count)
			}
			return
		}
	}
	t.Fatalf("summary rows missing %q: %+v", name, rows)
}

func containsText(value, substring string) bool {
	for i := 0; i+len(substring) <= len(value); i++ {
		if value[i:i+len(substring)] == substring {
			return true
		}
	}
	return false
}
