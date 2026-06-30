package projectindex

import (
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestBuildNilReturnsQueryableEmptyIndex(t *testing.T) {
	idx := Build(nil)
	if idx == nil {
		t.Fatal("Build(nil) returned nil")
	}
	if got := idx.CompositionByID(1); got != nil {
		t.Fatalf("CompositionByID(1) = %v, want nil", got)
	}
	if got := idx.FootageByID(1); got != nil {
		t.Fatalf("FootageByID(1) = %v, want nil", got)
	}
	if got := idx.AVItemByID(1); got != nil {
		t.Fatalf("AVItemByID(1) = %v, want nil", got)
	}
	assertNoLayers(t, "LayersByLayerID", idx.LayersByLayerID(1))
	assertNoLayers(t, "LayersByName", idx.LayersByName("L"))
	assertNoLayers(t, "LayersBySourceID", idx.LayersBySourceID(1))
	assertNoLayers(t, "LayersByEffect", idx.LayersByEffect("ADBE Glo2"))
}

func TestIndexProjectItemFirstMatchSemantics(t *testing.T) {
	firstComp := &aep.Composition{ID: 7, Name: "first comp"}
	secondComp := &aep.Composition{ID: 7, Name: "second comp"}
	firstFootage := &aep.Footage{ID: 8, Name: "first footage"}
	secondFootage := &aep.Footage{ID: 8, Name: "second footage"}
	sharedFootage := &aep.Footage{ID: 7, Name: "shared id footage"}
	project := &aep.Project{
		Compositions: []*aep.Composition{firstComp, secondComp},
		Footage:      []*aep.Footage{sharedFootage, firstFootage, secondFootage},
	}

	idx := Build(project)

	if got := idx.CompositionByID(7); got != firstComp {
		t.Fatalf("CompositionByID(7) = %p, want first comp %p", got, firstComp)
	}
	if got := idx.FootageByID(8); got != firstFootage {
		t.Fatalf("FootageByID(8) = %p, want first footage %p", got, firstFootage)
	}
	if got := idx.AVItemByID(7); got != firstComp {
		t.Fatalf("AVItemByID(7) = %p, want composition before footage %p", got, firstComp)
	}
	if got := idx.CompositionByID(0); got != nil {
		t.Fatalf("CompositionByID(0) = %v, want nil", got)
	}
	if got := idx.FootageByID(0); got != nil {
		t.Fatalf("FootageByID(0) = %v, want nil", got)
	}
	if got := idx.AVItemByID(0); got != nil {
		t.Fatalf("AVItemByID(0) = %v, want nil", got)
	}
}

func TestIndexLayerQueriesReturnAllMatchesInProjectOrder(t *testing.T) {
	layerA := &aep.Layer{
		ID:       11,
		Name:     "Title",
		SourceID: 100,
		Effects:  []*aep.Effect{{MatchName: "ADBE Glo2"}},
	}
	layerB := &aep.Layer{
		ID:       11,
		Name:     "Title",
		SourceID: 200,
		Effects:  []*aep.Effect{{MatchName: "ADBE Gaussian Blur 2"}},
	}
	layerC := &aep.Layer{
		ID:       12,
		Name:     "BG",
		SourceID: 100,
		Effects:  []*aep.Effect{{MatchName: "ADBE Glo2"}, {MatchName: "ADBE Glo2"}},
	}
	project := &aep.Project{
		Compositions: []*aep.Composition{
			{Name: "Main", Layers: []*aep.Layer{layerA, nil, layerB}},
			nil,
			{Name: "Precomp", Layers: []*aep.Layer{layerC}},
		},
	}

	idx := Build(project)

	assertLayers(t, "LayersByLayerID", idx.LayersByLayerID(11), layerA, layerB)
	assertLayers(t, "LayersByName", idx.LayersByName("Title"), layerA, layerB)
	assertLayers(t, "LayersBySourceID", idx.LayersBySourceID(100), layerA, layerC)
	assertLayers(t, "LayersByEffect", idx.LayersByEffect("ADBE Glo2"), layerA, layerC)
	assertNoLayers(t, "LayersByEffect missing", idx.LayersByEffect("ADBE Missing"))
	assertNoLayers(t, "LayersByLayerID zero", idx.LayersByLayerID(0))
	assertNoLayers(t, "LayersBySourceID zero", idx.LayersBySourceID(0))
}

func assertLayers(t *testing.T, name string, got []*aep.Layer, want ...*aep.Layer) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s len = %d, want %d", name, len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s[%d] = %p, want %p", name, i, got[i], want[i])
		}
	}
}

func assertNoLayers(t *testing.T, name string, got []*aep.Layer) {
	t.Helper()
	if len(got) != 0 {
		t.Fatalf("%s returned %d layer(s), want zero", name, len(got))
	}
}
