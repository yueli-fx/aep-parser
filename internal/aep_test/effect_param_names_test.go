package aep_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestEffectParamNamesRoundTrip(t *testing.T) {
	for _, name := range []string{"Blurriness", "blurriness", "模糊度"} {
		t.Run(name, func(t *testing.T) {
			project, layer, fx := gbDefaultInstanceLayer(t)
			// Start with an elided parameter, then update the materialized one.
			for _, value := range []float64{30, 12} {
				if _, err := aep.SetEffectParam(layer, fx, name, value); err != nil {
					t.Fatal(err)
				}
			}
			reopened, err := aep.Reopen(project)
			if err != nil {
				t.Fatal(err)
			}
			for _, param := range reopened.Compositions[0].LayerByName("S").Effects[0].Parameters {
				if param.MatchName == "ADBE Gaussian Blur 2-0001" {
					if param.StaticValue != float64(12) {
						t.Fatalf("saved value: %v", param.StaticValue)
					}
					return
				}
			}
			t.Fatal("missing saved blur parameter")
		})
	}
}

func TestEffectParamNamesOtherEffect(t *testing.T) {
	_, layer, _ := gbDefaultInstanceLayer(t)
	fx, err := aep.AddEffect(layer, aep.EffectSliderControl)
	if err != nil {
		t.Fatal(err)
	}
	param, err := aep.AnimateEffectParam(layer, fx, "滑块", []aep.ScalarKeyframe{
		{Time: 0, Value: 1}, {Time: 1, Value: 10},
	})
	if err != nil {
		t.Fatal(err)
	}
	if param.MatchName != "ADBE Slider Control-0001" || len(param.Keyframes) != 2 {
		t.Fatalf("unexpected animated parameter: %+v", param)
	}
}

func TestEffectParamNamesRejectWithoutMutation(t *testing.T) {
	project, layer, fx := gbDefaultInstanceLayer(t)
	var before, after bytes.Buffer
	if err := project.WriteAEP(&before); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"", "Blurrines", "Slider", "ADBE Slider Control-0001", "ADBE Gaussian Blur 2-0001", "adbe gaussian blur 2-0001", "-0001"} {
		if _, err := aep.SetEffectParam(layer, fx, name, 30.0); err == nil {
			t.Fatalf("accepted unknown parameter %q", name)
		}
	}
	if _, err := aep.AnimateEffectParam(layer, fx, "ADBE Gaussian Blur 2-0001", []aep.ScalarKeyframe{
		{Time: 0, Value: 30}, {Time: 1, Value: 0},
	}); err == nil {
		t.Fatal("scalar animation accepted a raw match-name")
	}
	if _, err := aep.AnimateEffectParamVec(layer, fx, "ADBE Gaussian Blur 2-0001", []aep.VectorKeyframe{
		{Time: 0, Value: []float64{0, 0}}, {Time: 1, Value: []float64{1, 1}},
	}); err == nil {
		t.Fatal("vector animation accepted a raw match-name")
	}
	// Simulate two controls sharing a display name. Never choose the first.
	fx.Parameters = append(fx.Parameters,
		&aep.Property{MatchName: "custom-1", Name: "Amount"},
		&aep.Property{MatchName: "custom-2", Name: "Amount"})
	if _, err := aep.SetEffectParam(layer, fx, "Amount", 30.0); err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("expected ambiguity error, got %v", err)
	}
	fx.Parameters = fx.Parameters[:len(fx.Parameters)-2]
	if err := project.WriteAEP(&after); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before.Bytes(), after.Bytes()) {
		t.Fatal("failed name resolution mutated the project")
	}
}
