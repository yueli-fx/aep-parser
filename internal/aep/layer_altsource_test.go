package aep_test

import (
	"bytes"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestAlternateSourceReal(t *testing.T) {
	proj, comp := openAlternateSourceAE24(t)

	// AE 25 wraps any AVItem passed to setAlternateSource() in an
	// auto-created precomp named "<slotName>_<originalName> 2", filed
	// under a "媒体替换合成" / "Media Replacement Comps" folder. The
	// persisted blsi points at the wrapper, not at the AVItem the script
	// passed. Our fixture set the alt source to ALT_SRC_B; AE stored the
	// wrapper comp "MediaSlot_ALT_SRC_B 2" instead.
	wrapper := proj.CompositionByName("MediaSlot_ALT_SRC_B 2")
	if wrapper == nil {
		t.Fatalf("expected AE-auto-created wrapper comp 'MediaSlot_ALT_SRC_B 2' missing — fixture or AE version differs?")
	}

	withAlt := layerByName(comp, "with_alt_b")
	baseline := layerByName(comp, "baseline_no_alt")
	if withAlt == nil || baseline == nil {
		t.Fatalf("fixture layers missing (withAlt=%v baseline=%v)", withAlt, baseline)
	}

	if !withAlt.HasAlternateSourceSlot() {
		t.Error("with_alt_b: HasAlternateSourceSlot() = false, want true")
	}
	if !baseline.HasAlternateSourceSlot() {
		t.Error("baseline_no_alt: HasAlternateSourceSlot() = false, want true (EGP slot is persisted even without an override)")
	}

	if withAlt.AlternateSourceID != wrapper.ID {
		t.Errorf("with_alt_b AlternateSourceID = %d, want %d (wrapper)", withAlt.AlternateSourceID, wrapper.ID)
	}
	got := withAlt.AlternateSource()
	if got == nil {
		t.Fatalf("with_alt_b AlternateSource() = nil, want wrapper comp")
	}
	if got.ItemID() != wrapper.ID || got.ItemName() != wrapper.Name {
		t.Errorf("with_alt_b AlternateSource() = {id=%d name=%q}, want {id=%d name=%q}",
			got.ItemID(), got.ItemName(), wrapper.ID, wrapper.Name)
	}

	if baseline.AlternateSourceID != 0 {
		t.Errorf("baseline_no_alt AlternateSourceID = %d, want 0", baseline.AlternateSourceID)
	}
	if baseline.AlternateSource() != nil {
		t.Errorf("baseline_no_alt AlternateSource() = %v, want nil", baseline.AlternateSource())
	}
}

func TestSetAlternateSourceRoundtrip(t *testing.T) {
	proj, comp := openAlternateSourceAE24(t)
	srcA := proj.CompositionByName("ALT_SRC_A")
	srcB := proj.CompositionByName("ALT_SRC_B")
	baseline := layerByName(comp, "baseline_no_alt")
	if srcA == nil || srcB == nil || baseline == nil {
		t.Fatalf("fixture missing (srcA=%v srcB=%v baseline=%v)", srcA, srcB, baseline)
	}
	if baseline.AlternateSourceID != 0 {
		t.Fatalf("baseline starts with override id=%d, want 0", baseline.AlternateSourceID)
	}

	if err := baseline.SetAlternateSource(srcA); err != nil {
		t.Fatalf("SetAlternateSource(srcA): %v", err)
	}
	if baseline.AlternateSourceID != srcA.ID {
		t.Errorf("after Set(srcA): AlternateSourceID = %d, want %d", baseline.AlternateSourceID, srcA.ID)
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	srcA2 := proj2.CompositionByName("ALT_SRC_A")
	srcB2 := proj2.CompositionByName("ALT_SRC_B")
	comp2 := proj2.CompositionByName("RE_ALT_SOURCE_MAIN")
	baseline2 := layerByName(comp2, "baseline_no_alt")
	if baseline2 == nil {
		t.Fatalf("baseline_no_alt missing after roundtrip")
	}
	if baseline2.AlternateSourceID != srcA2.ID {
		t.Errorf("roundtrip AlternateSourceID = %d, want %d", baseline2.AlternateSourceID, srcA2.ID)
	}
	if got := baseline2.AlternateSource(); got == nil || got.ItemID() != srcA2.ID {
		t.Errorf("roundtrip AlternateSource() = %v, want srcA", got)
	}

	// Switch to srcB and roundtrip again.
	if err := baseline2.SetAlternateSource(srcB2); err != nil {
		t.Fatalf("SetAlternateSource(srcB): %v", err)
	}
	var buf2 bytes.Buffer
	if err := proj2.WriteAEP(&buf2); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj3, err := aep.FromReader(bytes.NewReader(buf2.Bytes()))
	if err != nil {
		t.Fatalf("re-parse after switch: %v", err)
	}
	comp3 := proj3.CompositionByName("RE_ALT_SOURCE_MAIN")
	srcB3 := proj3.CompositionByName("ALT_SRC_B")
	baseline3 := layerByName(comp3, "baseline_no_alt")
	if baseline3.AlternateSourceID != srcB3.ID {
		t.Errorf("after switch+roundtrip: AlternateSourceID = %d, want %d", baseline3.AlternateSourceID, srcB3.ID)
	}

	// Clear and roundtrip.
	if err := baseline3.ClearAlternateSource(); err != nil {
		t.Fatalf("ClearAlternateSource: %v", err)
	}
	if baseline3.AlternateSourceID != 0 {
		t.Errorf("after Clear: AlternateSourceID = %d, want 0", baseline3.AlternateSourceID)
	}
	if baseline3.AlternateSource() != nil {
		t.Errorf("after Clear: AlternateSource() != nil")
	}
	var buf3 bytes.Buffer
	if err := proj3.WriteAEP(&buf3); err != nil {
		t.Fatalf("WriteAEP after clear: %v", err)
	}
	proj4, err := aep.FromReader(bytes.NewReader(buf3.Bytes()))
	if err != nil {
		t.Fatalf("re-parse after clear: %v", err)
	}
	comp4 := proj4.CompositionByName("RE_ALT_SOURCE_MAIN")
	baseline4 := layerByName(comp4, "baseline_no_alt")
	if baseline4.AlternateSourceID != 0 {
		t.Errorf("after clear+roundtrip: AlternateSourceID = %d, want 0", baseline4.AlternateSourceID)
	}
}

func TestSetAlternateSourceRejectsInvalid(t *testing.T) {
	proj, comp := openAlternateSourceAE24(t)
	baseline := layerByName(comp, "baseline_no_alt")
	if baseline == nil {
		t.Fatalf("baseline_no_alt missing")
	}

	// Item id not in project.
	bogus := &aep.Composition{ID: 999999, Name: "BOGUS"}
	if err := baseline.SetAlternateSource(bogus); err == nil {
		t.Error("bogus item id: expected error")
	}

	// Layer without an EGP slot: pick any layer inside ALT_SRC_A (the
	// "innerA" solid that has no addToMotionGraphics applied).
	srcAComp := proj.CompositionByName("ALT_SRC_A")
	if srcAComp == nil || len(srcAComp.Layers) == 0 {
		t.Skip("ALT_SRC_A has no layers; can't test no-slot rejection")
	}
	inner := srcAComp.Layers[0]
	if inner.HasAlternateSourceSlot() {
		t.Skip("unexpected: innerA has EGP slot")
	}
	srcB := proj.CompositionByName("ALT_SRC_B")
	if err := inner.SetAlternateSource(srcB); err == nil {
		t.Error("no-slot SetAlternateSource: expected error")
	}
}

// TestLayerAddFontAndUse exercises Layer.AddFont — append a new font
// to the btdk Fonts table, then point an existing style run at the
// new index via SetRunFontIndex. Round-trip via WriteAEP to confirm
// the new font appears in re-parsed TextSource.Fonts.
func TestLayerAddFontAndUse(t *testing.T) {
	proj, err := aep.Open("../../test_data/fixtures/re_text.aep")
	if err != nil {
		t.Skipf("re_text.aep not present")
	}
	var comp *aep.Composition
	for _, c := range proj.Compositions {
		if c.Name == "RE_TEXT" {
			if comp == nil || len(c.Layers) > len(comp.Layers) {
				comp = c
			}
		}
	}
	if comp == nil {
		t.Fatal("RE_TEXT comp missing")
	}
	var layer *aep.Layer
	for _, l := range comp.Layers {
		if l.Name == "baseline_A" {
			layer = l
			break
		}
	}
	if layer == nil || layer.TextSource == nil {
		t.Fatal("baseline_A text layer not found")
	}
	beforeCount := len(layer.TextSource.Fonts)

	newIdx, err := layer.AddFont("Arial-BoldMT")
	if err != nil {
		t.Fatalf("AddFont: %v", err)
	}
	if newIdx != beforeCount {
		t.Errorf("AddFont returned %d, want %d (= old len)", newIdx, beforeCount)
	}
	if len(layer.TextSource.Fonts) != beforeCount+1 {
		t.Errorf("Fonts len = %d, want %d", len(layer.TextSource.Fonts), beforeCount+1)
	}
	if layer.TextSource.Fonts[newIdx] != "Arial-BoldMT" {
		t.Errorf("Fonts[%d] = %q, want %q", newIdx, layer.TextSource.Fonts[newIdx], "Arial-BoldMT")
	}

	// Point run #0 at the new font.
	if err := layer.SetRunFontIndex(0, newIdx); err != nil {
		t.Fatalf("SetRunFontIndex: %v", err)
	}
	if layer.TextSource.Runs[0].FontIndex != newIdx || layer.TextSource.Runs[0].FontName != "Arial-BoldMT" {
		t.Errorf("after SetRunFontIndex: idx=%d name=%q", layer.TextSource.Runs[0].FontIndex, layer.TextSource.Runs[0].FontName)
	}

	// Round-trip
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	var layer2 *aep.Layer
	for _, c := range proj2.Compositions {
		if c.Name != "RE_TEXT" {
			continue
		}
		for _, l := range c.Layers {
			if l.Name == "baseline_A" {
				layer2 = l
				break
			}
		}
	}
	if layer2 == nil || layer2.TextSource == nil {
		t.Fatal("after roundtrip: baseline_A missing or undecoded")
	}
	if len(layer2.TextSource.Fonts) != beforeCount+1 {
		t.Errorf("roundtrip Fonts len = %d, want %d", len(layer2.TextSource.Fonts), beforeCount+1)
	}
	if layer2.TextSource.Fonts[newIdx] != "Arial-BoldMT" {
		t.Errorf("roundtrip Fonts[%d] = %q", newIdx, layer2.TextSource.Fonts[newIdx])
	}
	if layer2.TextSource.Runs[0].FontName != "Arial-BoldMT" {
		t.Errorf("roundtrip run[0].FontName = %q", layer2.TextSource.Runs[0].FontName)
	}
}
