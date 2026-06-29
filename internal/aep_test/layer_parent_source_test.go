package aep_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestLayerParent(t *testing.T) {
	// Three sibling layers in one comp: L10 (no parent), L20 (parent=L10),
	// L30 (parent=999 — orphan, no such layer).
	cases := []struct {
		name       string
		layerIndex int
		wantNil    bool
		wantParent uint32 // expected parent's ID when wantNil == false
	}{
		{"root layer has no parent", 0, true, 0},
		{"child resolves to parent", 1, false, 10},
		{"orphan reference returns nil", 2, true, 0},
	}

	rb := &rifxBuilder{}
	mk := func(id, parentID uint32) []byte {
		return buildLdta160(ldtaOpts{
			LayerID: id, ParentID: parentID, SourceID: 1,
			StretchDividend: 1, StretchDivisor: 1,
			OutPoint: 1.0,
		})
	}
	data := buildMultiLayerComp(rb, [][]byte{
		mk(10, 0),
		mk(20, 10),
		mk(30, 999),
	})
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	layers := proj.Compositions[0].Layers
	if len(layers) != 3 {
		t.Fatalf("expected 3 layers, got %d", len(layers))
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := layers[c.layerIndex].Parent()
			if c.wantNil {
				if got != nil {
					t.Errorf("Parent() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatalf("Parent() = nil, want layer with ID %d", c.wantParent)
			}
			if got.ID != c.wantParent {
				t.Errorf("Parent().ID = %d, want %d", got.ID, c.wantParent)
			}
		})
	}
}

func TestLayerParentNoComp(t *testing.T) {
	// Synthetic layer with no owning comp — Parent() must safely return nil.
	l := &aep.Layer{ID: 1, ParentID: 1}
	if got := l.Parent(); got != nil {
		t.Errorf("Parent() = %v, want nil", got)
	}
}

func TestLayerSourceComposition(t *testing.T) {
	// Project layout:
	//   Comp1 (id=1) — holds 3 layers
	//     L0: SourceID=2   → points to Comp2 (valid pre-comp)
	//     L1: SourceID=3   → points to Footage (non-comp source)
	//     L2: SourceID=999 → missing item
	//   Comp2 (id=2) — empty pre-comp target
	//   Footage (id=3)
	rb := &rifxBuilder{}
	mk := func(id, sourceID uint32) []byte {
		return buildLdta160(ldtaOpts{
			LayerID: id, SourceID: sourceID,
			StretchDividend: 1, StretchDivisor: 1,
			OutPoint: 1.0,
		})
	}
	data := buildMultiCompProject(rb, 2, [][]byte{
		mk(10, 2),
		mk(20, 3),
		mk(30, 999),
	})
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	if len(proj.Compositions) != 2 {
		t.Fatalf("expected 2 comps, got %d", len(proj.Compositions))
	}
	if len(proj.Footage) != 1 {
		t.Fatalf("expected 1 footage, got %d", len(proj.Footage))
	}
	layers := proj.Compositions[0].Layers
	if len(layers) != 3 {
		t.Fatalf("expected 3 layers in comp1, got %d", len(layers))
	}

	cases := []struct {
		name       string
		layerIndex int
		wantNil    bool
		wantCompID uint32
	}{
		{"valid pre-comp source", 0, false, 2},
		{"footage source returns nil", 1, true, 0},
		{"missing source returns nil", 2, true, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := layers[c.layerIndex].SourceComposition()
			if c.wantNil {
				if got != nil {
					t.Errorf("SourceComposition() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatalf("SourceComposition() = nil, want comp with ID %d", c.wantCompID)
			}
			if got.ID != c.wantCompID {
				t.Errorf("SourceComposition().ID = %d, want %d", got.ID, c.wantCompID)
			}
		})
	}
}

func TestLayerSourceCompositionNoOwner(t *testing.T) {
	// Synthetic layer with no owning comp/project — must return nil safely.
	l := &aep.Layer{ID: 1, SourceID: 1}
	if got := l.SourceComposition(); got != nil {
		t.Errorf("SourceComposition() = %v, want nil", got)
	}
}

func TestJSONParentName(t *testing.T) {
	// Same scaffolding as TestLayerParent: L10 (root), L20 (parent=L10),
	// L30 (parent=999, orphan). After JSON marshal:
	//   L10 → no parent_id, no parent_name
	//   L20 → parent_id=10, parent_name="" (no Utf8 name on L10 here)
	//   L30 → parent_id=999, no parent_name (orphan)
	// Then a second pass with a Utf8 name on L10 to verify parent_name
	// actually resolves to the parent's Name.
	rb := &rifxBuilder{}
	mk := func(id, parentID uint32) []byte {
		return buildLdta160(ldtaOpts{
			LayerID: id, ParentID: parentID, SourceID: 1,
			StretchDividend: 1, StretchDivisor: 1,
			OutPoint: 1.0,
		})
	}
	data := buildMultiLayerComp(rb, [][]byte{mk(10, 0), mk(20, 10), mk(30, 999)})
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	// Assign a Name to L10 directly on the parsed model — equivalent to a
	// Utf8 sibling in the Layr LIST, but simpler for the table here.
	proj.Compositions[0].Layers[0].Name = "background"

	jp := proj.ToJSON()
	jl := jp.Compositions[0].Layers

	cases := []struct {
		idx            int
		wantParentID   uint32
		wantParentName string
	}{
		{0, 0, ""},            // root: no parent
		{1, 10, "background"}, // child: resolved
		{2, 999, ""},          // orphan: parent_id present, parent_name empty
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("layer_%d", c.idx), func(t *testing.T) {
			if jl[c.idx].ParentID != c.wantParentID {
				t.Errorf("ParentID = %d, want %d", jl[c.idx].ParentID, c.wantParentID)
			}
			if jl[c.idx].ParentName != c.wantParentName {
				t.Errorf("ParentName = %q, want %q", jl[c.idx].ParentName, c.wantParentName)
			}
		})
	}
}
