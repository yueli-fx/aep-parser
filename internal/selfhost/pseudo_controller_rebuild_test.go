package selfhost

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildPseudoControllerRebuildProofMapsFamilyControls(t *testing.T) {
	inventory := pseudoControllerInventoryFixture()
	understanding := pseudoControllerUnderstandingFixture()

	proof, err := BuildPseudoControllerRebuildProof(inventory, understanding, PseudoControllerRebuildOptions{MaxFamilies: 3})
	if err != nil {
		t.Fatalf("BuildPseudoControllerRebuildProof: %v", err)
	}

	if proof.SchemaVersion != 1 {
		t.Fatalf("schema_version = %d, want 1", proof.SchemaVersion)
	}
	if proof.Summary.PseudoFamilies != 2 || proof.Summary.SelectedFamilies != 2 {
		t.Fatalf("summary = %+v", proof.Summary)
	}
	top := proof.Families[0]
	if top.MatchName != "Pseudo/RigA" {
		t.Fatalf("top family = %q, want Pseudo/RigA", top.MatchName)
	}
	if top.SystemParamCount != 1 {
		t.Fatalf("system_param_count = %d, want 1", top.SystemParamCount)
	}
	assertPseudoControlPlan(t, top, "Pseudo/RigA-0001", "slider", "keyframed")
	assertPseudoControlPlan(t, top, "Pseudo/RigA-0002", "color", "")
	assertPseudoControlPlan(t, top, "Pseudo/RigA-0003", "checkbox", "")
	assertPseudoControlPlan(t, top, "Pseudo/RigA-0004", "layer", "")
	if len(top.UnsupportedControls) != 0 {
		t.Fatalf("unsupported controls = %+v", top.UnsupportedControls)
	}
	if proof.Families[1].MatchName != "Pseudo/RigB" || len(proof.Families[1].UnsupportedControls) != 1 {
		t.Fatalf("second family = %+v", proof.Families[1])
	}
}

func TestRunPseudoControllerRebuildProofWritesJSONAndAEP(t *testing.T) {
	dir := t.TempDir()
	inventoryPath := filepath.Join(dir, "inventory.json")
	understandingPath := filepath.Join(dir, "understanding.json")
	outDir := filepath.Join(dir, "proof")
	writeJSONFile(t, inventoryPath, pseudoControllerInventoryFixture())
	writeJSONFile(t, understandingPath, pseudoControllerUnderstandingFixture())

	proof, err := RunPseudoControllerRebuildProof(PseudoControllerRebuildOptions{
		InventoryPath:     inventoryPath,
		UnderstandingPath: understandingPath,
		OutDir:            outDir,
		MaxFamilies:       2,
	})
	if err != nil {
		t.Fatalf("RunPseudoControllerRebuildProof: %v", err)
	}

	if proof.OutputPath != filepath.Join(outDir, "proof.json") {
		t.Fatalf("output_path = %q", proof.OutputPath)
	}
	if _, err := os.Stat(proof.OutputPath); err != nil {
		t.Fatalf("proof.json missing: %v", err)
	}
	if proof.Families[0].GeneratedAEP == "" {
		t.Fatalf("generated AEP missing in top family: %+v", proof.Families[0])
	}
	if _, err := os.Stat(proof.Families[0].GeneratedAEP); err != nil {
		t.Fatalf("generated AEP missing: %v", err)
	}
	if proof.Summary.GeneratedFamilies != 1 || proof.Summary.GeneratedControls == 0 {
		t.Fatalf("summary = %+v", proof.Summary)
	}
}

func TestRunPseudoControllerRebuildProofGeneratesEverySupportedFamily(t *testing.T) {
	dir := t.TempDir()
	inventoryPath := filepath.Join(dir, "inventory.json")
	understandingPath := filepath.Join(dir, "understanding.json")
	outDir := filepath.Join(dir, "proof")
	inventory := pseudoControllerInventoryFixture()
	inventory.Effects[1] = EffectFieldInventoryEffect{
		MatchName:        "Pseudo/RigB",
		Class:            "pseudo",
		Occurrences:      5,
		ProjectSamples:   []string{"samples/b.aep"},
		ParamKinds:       2,
		ParamOccurrences: 10,
		Params: []EffectFieldInventoryParam{
			pseudoParam("Pseudo/RigB-0000", "layer_ref", 5, 0, 0, 5, nil),
			pseudoParam("Pseudo/RigB-0001", "number", 5, 1, 0, 0, []any{10}),
		},
	}
	writeJSONFile(t, inventoryPath, inventory)
	writeJSONFile(t, understandingPath, BuildEffectFieldUnderstanding(inventory, "inventory.json"))

	proof, err := RunPseudoControllerRebuildProof(PseudoControllerRebuildOptions{
		InventoryPath:     inventoryPath,
		UnderstandingPath: understandingPath,
		OutDir:            outDir,
		MaxFamilies:       2,
	})
	if err != nil {
		t.Fatalf("RunPseudoControllerRebuildProof: %v", err)
	}

	if proof.Summary.GeneratedFamilies != 2 {
		t.Fatalf("generated_families = %d, want 2", proof.Summary.GeneratedFamilies)
	}
	for _, family := range proof.Families {
		if family.Status != "generated" || family.GeneratedAEP == "" {
			t.Fatalf("family not generated: %+v", family)
		}
		if _, err := os.Stat(family.GeneratedAEP); err != nil {
			t.Fatalf("generated AEP missing for %s: %v", family.MatchName, err)
		}
	}
}

func TestRunPseudoControllerRebuildProofRejectsMalformedInput(t *testing.T) {
	dir := t.TempDir()
	inventoryPath := filepath.Join(dir, "inventory.json")
	writeFile(t, inventoryPath, `{"schema_version":`)

	_, err := RunPseudoControllerRebuildProof(PseudoControllerRebuildOptions{
		InventoryPath:     inventoryPath,
		UnderstandingPath: filepath.Join(dir, "missing.json"),
		OutDir:            filepath.Join(dir, "proof"),
	})

	if err == nil {
		t.Fatalf("err = nil, want malformed inventory error")
	}
}

func assertPseudoControlPlan(t *testing.T, family PseudoControllerRebuildFamily, paramMatchName, kind, note string) {
	t.Helper()
	for _, control := range family.Controls {
		if control.ParamMatchName != paramMatchName {
			continue
		}
		if control.ControlKind != kind {
			t.Fatalf("%s control kind = %q, want %q", paramMatchName, control.ControlKind, kind)
		}
		if note != "" && !containsString(control.BehaviorNotes, note) {
			t.Fatalf("%s notes = %+v, want %q", paramMatchName, control.BehaviorNotes, note)
		}
		return
	}
	t.Fatalf("control %q missing from %+v", paramMatchName, family.Controls)
}

func pseudoControllerInventoryFixture() EffectFieldInventory {
	return EffectFieldInventory{
		SchemaVersion: 1,
		Root:          "data/samples",
		Summary: EffectFieldInventorySummary{
			ProjectCount:      2,
			EffectKinds:       3,
			EffectOccurrences: 16,
			ParamKinds:        8,
			ParamOccurrences:  80,
		},
		Effects: []EffectFieldInventoryEffect{
			{
				MatchName:        "Pseudo/RigA",
				Class:            "pseudo",
				Occurrences:      10,
				ProjectSamples:   []string{"samples/a.aep"},
				ParamKinds:       5,
				ParamOccurrences: 50,
				Params: []EffectFieldInventoryParam{
					pseudoParam("Pseudo/RigA-0000", "layer_ref", 10, 0, 0, 10, nil),
					pseudoParam("Pseudo/RigA-0001", "number", 10, 2, 0, 0, []any{42}),
					pseudoParam("Pseudo/RigA-0002", "vector4", 10, 0, 0, 0, []any{[]any{1, 0, 0, 1}}),
					pseudoParam("Pseudo/RigA-0003", "bool", 10, 0, 0, 0, []any{true}),
					pseudoParam("Pseudo/RigA-0004", "layer_ref", 10, 0, 0, 10, nil),
				},
			},
			{
				MatchName:        "Pseudo/RigB",
				Class:            "pseudo",
				Occurrences:      5,
				ProjectSamples:   []string{"samples/b.aep"},
				ParamKinds:       2,
				ParamOccurrences: 10,
				Params: []EffectFieldInventoryParam{
					pseudoParam("Pseudo/RigB-0000", "layer_ref", 5, 0, 0, 5, nil),
					pseudoParam("Pseudo/RigB-0001", "string", 5, 0, 0, 0, []any{"mode"}),
				},
			},
			{
				MatchName:        "ADBE Slider Control",
				Class:            "native_supported",
				Occurrences:      1,
				ParamKinds:       1,
				ParamOccurrences: 1,
			},
		},
	}
}

func pseudoControllerUnderstandingFixture() EffectFieldUnderstanding {
	return BuildEffectFieldUnderstanding(pseudoControllerInventoryFixture(), "inventory.json")
}

func pseudoParam(matchName, valueType string, occurrences, keyframes, expressions, layerRefs int, examples []any) EffectFieldInventoryParam {
	return EffectFieldInventoryParam{
		MatchName:              matchName,
		Names:                  []CountRow{{Name: matchName, Count: occurrences}},
		ValueTypes:             []CountRow{{Name: valueType, Count: occurrences}},
		Occurrences:            occurrences,
		StaticValueOccurrences: occurrences,
		KeyframedOccurrences:   keyframes,
		ExpressionOccurrences:  expressions,
		LayerRefOccurrences:    layerRefs,
		ExampleStaticValues:    examples,
	}
}
