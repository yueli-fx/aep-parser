package scene

import "testing"

func TestCompositionSourceIndexPreservesFirstMatchSemantics(t *testing.T) {
	firstComp := &Composition{ID: 10, Name: "first comp"}
	secondComp := &Composition{ID: 10, Name: "second comp"}
	firstFootage := &Footage{ID: 20, Name: "first footage"}
	secondFootage := &Footage{ID: 20, Name: "second footage"}
	proj := &Project{
		Compositions: []*Composition{firstComp, secondComp},
		Footage:      []*Footage{firstFootage, secondFootage},
	}
	comp := &Composition{proj: proj}

	index := newCompositionSourceIndex(comp)

	if got := index.composition(10); got != firstComp {
		t.Fatalf("composition(10) = %p (%q), want first duplicate %p", got, got.Name, firstComp)
	}
	if got := index.footage(20); got != firstFootage {
		t.Fatalf("footage(20) = %p (%q), want first duplicate %p", got, got.Name, firstFootage)
	}
	if got := index.composition(0); got != nil {
		t.Fatalf("composition(0) = %p, want nil", got)
	}
	if got := index.footage(0); got != nil {
		t.Fatalf("footage(0) = %p, want nil", got)
	}
}

func TestCompositionSourceViewsUseLocalSourceIndex(t *testing.T) {
	precomp := &Composition{ID: 10, Name: "precomp"}
	file := &Footage{ID: 20, Name: "file"}
	solid := &Footage{ID: 30, Name: "solid", IsSolid: true}
	placeholder := &Footage{ID: 40, Name: "placeholder", IsPlaceholder: true}
	proj := &Project{
		Compositions: []*Composition{precomp},
		Footage:      []*Footage{file, solid, placeholder},
	}
	comp := &Composition{Name: "main", proj: proj}
	compLayer := &Layer{Name: "comp", SourceID: precomp.ID, comp: comp}
	fileLayer := &Layer{Name: "file", SourceID: file.ID, comp: comp}
	solidLayer := &Layer{Name: "solid", SourceID: solid.ID, comp: comp}
	placeholderLayer := &Layer{Name: "placeholder", SourceID: placeholder.ID, comp: comp}
	unwiredLayer := &Layer{Name: "unwired", SourceID: file.ID}
	comp.Layers = []*Layer{compLayer, fileLayer, solidLayer, placeholderLayer, unwiredLayer}

	assertLayers(t, "AVLayers", comp.AVLayers(), compLayer, fileLayer, solidLayer, placeholderLayer)
	assertLayers(t, "CompositionLayers", comp.CompositionLayers(), compLayer)
	assertLayers(t, "FootageLayers", comp.FootageLayers(), fileLayer, solidLayer, placeholderLayer)
	assertLayers(t, "FileLayers", comp.FileLayers(), fileLayer)
	assertLayers(t, "SolidLayers", comp.SolidLayers(), solidLayer)
	assertLayers(t, "PlaceholderLayers", comp.PlaceholderLayers(), placeholderLayer)
}

func assertLayers(t *testing.T, name string, got []*Layer, want ...*Layer) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s returned %d layers, want %d", name, len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("%s[%d] = %q, want %q", name, i, got[i].Name, want[i].Name)
		}
	}
}
