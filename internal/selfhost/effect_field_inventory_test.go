package selfhost

import (
	"testing"

	"github.com/yueli-fx/aep-parser/internal/profile"
)

func TestBuildEffectFieldInventoryAggregatesEffectParams(t *testing.T) {
	exprEnabled := true
	prof := &profile.Profile{
		Meta: profile.Meta{Path: "sample.aep"},
		Comps: []profile.Composition{{
			Name: "Main",
			Layers: []profile.Layer{{
				Name: "Layer",
				Effects: []profile.Effect{
					{
						MatchName:       "ADBE Slider Control",
						DisplayName:     "Slider",
						DependencyClass: "native",
						Params: []profile.Property{{
							Name:              "Slider",
							MatchName:         "ADBE Slider Control-0001",
							StaticValue:       42.0,
							Changed:           true,
							Expression:        "time",
							ExpressionEnabled: &exprEnabled,
							Keyframes: []profile.Keyframe{{
								Time:  0,
								Value: 1.0,
							}},
						}},
					},
					{
						MatchName:       "Pseudo/ColorRig",
						DisplayName:     "Color Rig",
						DependencyClass: "pseudo",
						Params: []profile.Property{{
							Name:        "Color",
							MatchName:   "Pseudo/ColorRig-0001",
							StaticValue: []float64{1, 0.5, 0.25, 1},
						}},
					},
					{
						MatchName:       "SC Glow",
						DisplayName:     "Sapphire Glow",
						DependencyClass: "third_party",
						Params: []profile.Property{{
							Name:      "Input Layer",
							MatchName: "SC Glow-0001",
							LayerRef:  &profile.LayerRef{ID: 7, Index: 2, Name: "Map"},
						}},
					},
				},
			}},
		}},
	}

	inventory := BuildEffectFieldInventory([]*profile.Profile{prof})

	if inventory.Summary.ProjectCount != 1 {
		t.Fatalf("project count = %d, want 1", inventory.Summary.ProjectCount)
	}
	if inventory.Summary.EffectKinds != 3 || inventory.Summary.ParamKinds != 3 {
		t.Fatalf("summary = %+v, want 3 effect and param kinds", inventory.Summary)
	}
	assertInventoryEffect(t, inventory, "ADBE Slider Control", "native_supported", []string{"scalar", "slider", "time"})
	assertInventoryParam(t, inventory, "ADBE Slider Control", "ADBE Slider Control-0001", "number", 1, 1, 1, 0)
	assertInventoryEffect(t, inventory, "Pseudo/ColorRig", "pseudo", []string{"color"})
	assertInventoryParam(t, inventory, "Pseudo/ColorRig", "Pseudo/ColorRig-0001", "vector4", 0, 0, 0, 0)
	assertInventoryEffect(t, inventory, "SC Glow", "third_party", []string{"layer_ref"})
	assertInventoryParam(t, inventory, "SC Glow", "SC Glow-0001", "layer_ref", 0, 0, 0, 1)
}

func assertInventoryEffect(t *testing.T, inventory EffectFieldInventory, matchName, class string, capabilities []string) {
	t.Helper()
	for _, effect := range inventory.Effects {
		if effect.MatchName != matchName {
			continue
		}
		if effect.Class != class {
			t.Fatalf("%s class = %q, want %q", matchName, effect.Class, class)
		}
		for _, want := range capabilities {
			if !containsString(effect.InferredCapabilities, want) {
				t.Fatalf("%s capabilities = %+v, want %q", matchName, effect.InferredCapabilities, want)
			}
		}
		return
	}
	t.Fatalf("effect %q missing from %+v", matchName, inventory.Effects)
}

func assertInventoryParam(t *testing.T, inventory EffectFieldInventory, effectName, paramName, valueType string, keyframes, expressions, changed, layerRefs int) {
	t.Helper()
	for _, effect := range inventory.Effects {
		if effect.MatchName != effectName {
			continue
		}
		for _, param := range effect.Params {
			if param.MatchName != paramName {
				continue
			}
			if !countRowsContain(param.ValueTypes, valueType) {
				t.Fatalf("%s value types = %+v, want %q", paramName, param.ValueTypes, valueType)
			}
			if param.KeyframedOccurrences != keyframes || param.ExpressionOccurrences != expressions || param.ChangedOccurrences != changed || param.LayerRefOccurrences != layerRefs {
				t.Fatalf("%s param = %+v", paramName, param)
			}
			return
		}
	}
	t.Fatalf("param %q/%q missing", effectName, paramName)
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func countRowsContain(rows []CountRow, name string) bool {
	for _, row := range rows {
		if row.Name == name {
			return true
		}
	}
	return false
}
