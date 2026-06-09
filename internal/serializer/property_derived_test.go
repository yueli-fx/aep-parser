package serializer

import (
	"testing"

	"github.com/example/aep-parser/internal/rifx"
	"github.com/example/aep-parser/internal/scene"
)

// Property derived-state tests — IsModified / Active / Enabled / Elided /
// IsNameSet. Internal-test so we can wire tdsb / tdb4 chunks directly.

func TestProperty_Enabled_DefaultTrue(t *testing.T) {
	p := &Property{MatchName: "test"}
	if !p.Enabled() {
		t.Error("Enabled() with no tdsb = false; want true (py-aep _enable_flags default = 1)")
	}
	if !p.Active() {
		t.Error("Active() with no tdsb = false; want true (alias of Enabled)")
	}
}

func TestProperty_Enabled_TdsbByte3Bit0(t *testing.T) {
	// tdsb byte 3 bit 0 = 1 → enabled
	tdsb := &rifx.Chunk{ID: rifx.IDTdsb, Data: []byte{0, 0, 0, 0x01}}
	p := &Property{MatchName: "test"}
	scene.SetPropertyBack(p, &propertyBackrefs{tdsb: tdsb})
	if !p.Enabled() {
		t.Error("Enabled() with tdsb byte3 bit0 set = false; want true")
	}
	// Clear bit 0 → disabled
	tdsb.Data[3] = 0x00
	if p.Enabled() {
		t.Error("Enabled() with tdsb byte3 bit0 clear = true; want false")
	}
}

func TestProperty_IsModified_Animated(t *testing.T) {
	p := &Property{
		MatchName: "test",
		Keyframes: []*Keyframe{{Time: 0}},
	}
	if !p.IsModified() {
		t.Error("IsModified() with keyframes = false; want true")
	}
}

func TestProperty_IsModified_Expression(t *testing.T) {
	p := &Property{
		MatchName:  "test",
		Expression: "time * 100",
	}
	if !p.IsModified() {
		t.Error("IsModified() with expression = false; want true")
	}
}

func TestProperty_IsModified_ValueDiffersFromDefault(t *testing.T) {
	p := &Property{
		MatchName:    "test",
		StaticValue:  50.0,
		DefaultValue: 100.0,
	}
	if !p.IsModified() {
		t.Error("IsModified() with value != default = false; want true")
	}
}

func TestProperty_IsModified_ValueEqualsDefault(t *testing.T) {
	p := &Property{
		MatchName:    "test",
		StaticValue:  100.0,
		DefaultValue: 100.0,
	}
	if p.IsModified() {
		t.Error("IsModified() with value == default = true; want false")
	}
}

func TestProperty_IsModified_ValueEqualsDefault_Vector(t *testing.T) {
	p := &Property{
		MatchName:    "test",
		StaticValue:  []float64{0, 0, 0},
		DefaultValue: []float64{0, 0, 0},
	}
	if p.IsModified() {
		t.Error("IsModified() with vector value == default = true; want false")
	}
	// Mutate one component → modified.
	p.StaticValue = []float64{1, 0, 0}
	if !p.IsModified() {
		t.Error("IsModified() with vector value differs in one component = false; want true")
	}
}

func TestProperty_IsModified_NoDefault(t *testing.T) {
	// No default + no keyframes + no expression → reports false (we can't
	// know if it's been modified without a baseline).
	p := &Property{
		MatchName:   "test",
		StaticValue: 42.0,
	}
	if p.IsModified() {
		t.Error("IsModified() with no default + no keyframes + no expression = true; want false (no baseline)")
	}
}

func TestProperty_Elided_AlwaysFalse(t *testing.T) {
	p := &Property{MatchName: "test"}
	if p.Elided() {
		t.Error("Elided() = true; want false (placeholder impl; py-aep marks synthesized groups only, we don't synthesize yet)")
	}
}

func TestProperty_IsNameSet_Default(t *testing.T) {
	// Parser sets Name = MatchName by default → not user-set.
	p := &Property{MatchName: "ADBE Position", Name: "ADBE Position"}
	if p.IsNameSet() {
		t.Error("IsNameSet() with Name == MatchName = true; want false")
	}
}

func TestProperty_IsNameSet_Explicit(t *testing.T) {
	p := &Property{MatchName: "ADBE Position", Name: "My Custom Position"}
	if !p.IsNameSet() {
		t.Error("IsNameSet() with Name != MatchName = false; want true")
	}
}

func TestAEPropertyGroup_IsModified_NilSafe(t *testing.T) {
	var g *AEPropertyGroup
	if g.IsModified() {
		t.Error("nil.IsModified() = true; want false")
	}
}

func TestAEPropertyGroup_IsModified_IndexedGroupWithChildren(t *testing.T) {
	parade := &AEPropertyGroup{
		MatchName: MatchNameGroupEffectParade,
		Children:  []PropertyBase{&AEPropertyGroup{MatchName: "ADBE Gaussian Blur 2"}},
	}
	if !parade.IsModified() {
		t.Error("EffectParade with 1 child = not modified; want true (indexed groups modify on insert)")
	}
}

func TestAEPropertyGroup_IsModified_IndexedGroupEmpty(t *testing.T) {
	parade := &AEPropertyGroup{MatchName: MatchNameGroupEffectParade}
	if parade.IsModified() {
		t.Error("empty EffectParade = modified; want false")
	}
}

func TestAEPropertyGroup_IsModified_RecurseIntoLeaf(t *testing.T) {
	leaf := &Property{
		MatchName:    "ADBE Opacity",
		StaticValue:  50.0,
		DefaultValue: 100.0,
	}
	group := &AEPropertyGroup{
		MatchName: MatchNameGroupTransform,
		Children:  []PropertyBase{leaf},
	}
	if !group.IsModified() {
		t.Error("TransformGroup containing modified Opacity = not modified; want true")
	}
}

func TestAEPropertyGroup_IsModified_AllUnchanged(t *testing.T) {
	leaf := &Property{
		MatchName:    "ADBE Opacity",
		StaticValue:  100.0,
		DefaultValue: 100.0,
	}
	group := &AEPropertyGroup{
		MatchName: MatchNameGroupTransform,
		Children:  []PropertyBase{leaf},
	}
	if group.IsModified() {
		t.Error("TransformGroup with unmodified Opacity = modified; want false")
	}
}
