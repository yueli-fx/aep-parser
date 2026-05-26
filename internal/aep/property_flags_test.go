package aep_test

import (
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// TestProperty_Tdb4Flags verifies the tdb4 flag readers against
// re_batch.aep (mix of spatial Position keyframes, color Tritone params,
// scalar Gaussian Blur Blurriness). Treats failures as warnings (Logf)
// when the property cannot be located — fixture content drift should
// not block the suite.
func TestProperty_Tdb4Flags(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_batch.aep")
	if err != nil {
		t.Skipf("re_batch.aep not present: %v", err)
	}

	var (
		position    *aep.Property
		opacity     *aep.Property
		colorEffect *aep.Property // ADBE Tritone-0001 (4D color) or similar
		blurriness  *aep.Property
	)
	for _, c := range proj.Compositions {
		for _, l := range c.Layers {
			for _, p := range l.Properties {
				switch p.MatchName {
				case aep.MatchNamePosition:
					if position == nil {
						position = p
					}
				case aep.MatchNameOpacity:
					if opacity == nil {
						opacity = p
					}
				}
			}
			for _, e := range l.Effects {
				for _, p := range e.Parameters {
					switch p.MatchName {
					case "ADBE Tritone-0001":
						if colorEffect == nil && p.Components == 4 {
							colorEffect = p
						}
					case "ADBE Gaussian Blur 2-0001":
						if blurriness == nil {
							blurriness = p
						}
					}
				}
			}
		}
	}

	if position != nil {
		if !position.IsSpatial() {
			t.Errorf("Position IsSpatial = false; want true")
		}
		if !position.IsVector() {
			t.Errorf("Position IsVector = false; want true")
		}
		if position.IsColor() {
			t.Error("Position IsColor = true; want false")
		}
		if !position.CanVaryOverTime() {
			t.Errorf("Position CanVaryOverTime = false; want true")
		}
	} else {
		t.Log("Position property not found in fixture; skipping spatial checks")
	}

	if opacity != nil {
		if opacity.IsSpatial() {
			t.Errorf("Opacity IsSpatial = true; want false")
		}
		if opacity.IsColor() {
			t.Errorf("Opacity IsColor = true; want false")
		}
		if !opacity.CanVaryOverTime() {
			t.Errorf("Opacity CanVaryOverTime = false; want true")
		}
	}

	if colorEffect != nil {
		if !colorEffect.IsColor() {
			t.Errorf("Tritone-0001 (4D) IsColor = false; want true")
		}
		if colorEffect.IsSpatial() {
			t.Errorf("Tritone-0001 IsSpatial = true; want false")
		}
	} else {
		t.Log("Tritone-0001 4D color not found; skipping color checks")
	}

	if blurriness != nil && len(blurriness.Keyframes) > 0 {
		if !blurriness.IsAnimated() {
			t.Errorf("Gaussian Blur Blurriness (has %d kfs) IsAnimated = false", len(blurriness.Keyframes))
		}
	}
}

func TestProperty_TdbFlags_FallbackDefaults(t *testing.T) {
	// Property built without going through the parser → tdb4 == nil.
	// Flag readers should default to false; IsAnimated falls back to
	// len(Keyframes).
	p := &aep.Property{MatchName: "test", Components: 1}
	if p.IsSpatial() {
		t.Error("IsSpatial on bare Property = true; want false")
	}
	if p.IsColor() || p.IsInteger() || p.IsVector() || p.IsNoValue() {
		t.Error("Type flags on bare Property all should be false")
	}
	if p.CanVaryOverTime() {
		t.Error("CanVaryOverTime on bare Property = true; want false (default)")
	}
	if p.IsAnimated() {
		t.Error("IsAnimated on bare Property (0 keyframes) = true")
	}
}
