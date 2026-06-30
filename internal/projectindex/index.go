// Package projectindex builds workflow-local lookup indexes over parsed AEP
// projects without making the core Project model maintain persistent maps.
package projectindex

import "github.com/yueli-fx/aep-parser/internal/aep"

// Index is a snapshot lookup view over one Project.
//
// It stores pointers to the parsed project graph observed by Build. It does not
// track later mutations; rebuild after changing indexed fields or after Reopen.
type Index struct {
	project *aep.Project

	compositionsByID map[uint32]*aep.Composition
	footageByID      map[uint32]*aep.Footage
	layersByLayerID  map[uint32][]*aep.Layer
	layersByName     map[string][]*aep.Layer
	layersBySourceID map[uint32][]*aep.Layer
	layersByEffect   map[string][]*aep.Layer
}

// Build returns a non-nil snapshot index for project. Build(nil) returns an
// empty index whose queries return nil or zero results.
func Build(project *aep.Project) *Index {
	idx := &Index{
		project:          project,
		compositionsByID: map[uint32]*aep.Composition{},
		footageByID:      map[uint32]*aep.Footage{},
		layersByLayerID:  map[uint32][]*aep.Layer{},
		layersByName:     map[string][]*aep.Layer{},
		layersBySourceID: map[uint32][]*aep.Layer{},
		layersByEffect:   map[string][]*aep.Layer{},
	}
	if project == nil {
		return idx
	}
	for _, comp := range project.Compositions {
		if comp == nil {
			continue
		}
		if comp.ID != 0 {
			if _, ok := idx.compositionsByID[comp.ID]; !ok {
				idx.compositionsByID[comp.ID] = comp
			}
		}
		for _, layer := range comp.Layers {
			idx.addLayer(layer)
		}
	}
	for _, footage := range project.Footage {
		if footage == nil || footage.ID == 0 {
			continue
		}
		if _, ok := idx.footageByID[footage.ID]; !ok {
			idx.footageByID[footage.ID] = footage
		}
	}
	return idx
}

func (idx *Index) addLayer(layer *aep.Layer) {
	if layer == nil {
		return
	}
	if layer.ID != 0 {
		idx.layersByLayerID[layer.ID] = append(idx.layersByLayerID[layer.ID], layer)
	}
	if layer.Name != "" {
		idx.layersByName[layer.Name] = append(idx.layersByName[layer.Name], layer)
	}
	if layer.SourceID != 0 {
		idx.layersBySourceID[layer.SourceID] = append(idx.layersBySourceID[layer.SourceID], layer)
	}
	seenEffects := map[string]bool{}
	for _, effect := range layer.Effects {
		if effect == nil || effect.MatchName == "" || seenEffects[effect.MatchName] {
			continue
		}
		seenEffects[effect.MatchName] = true
		idx.layersByEffect[effect.MatchName] = append(idx.layersByEffect[effect.MatchName], layer)
	}
}

// CompositionByID returns the first composition with id in Project.Compositions
// slice order, or nil when id is 0 or missing.
func (idx *Index) CompositionByID(id uint32) *aep.Composition {
	if idx == nil || id == 0 {
		return nil
	}
	return idx.compositionsByID[id]
}

// FootageByID returns the first footage item with id in Project.Footage slice
// order, or nil when id is 0 or missing.
func (idx *Index) FootageByID(id uint32) *aep.Footage {
	if idx == nil || id == 0 {
		return nil
	}
	return idx.footageByID[id]
}

// AVItemByID matches Project.AVItemByID semantics: compositions first, then
// footage, each using first match in the corresponding project slice.
func (idx *Index) AVItemByID(id uint32) aep.AVItem {
	if idx == nil || id == 0 {
		return nil
	}
	if comp := idx.compositionsByID[id]; comp != nil {
		return comp
	}
	if footage := idx.footageByID[id]; footage != nil {
		return footage
	}
	return nil
}

// LayersByLayerID returns all layers whose layer ID equals id.
func (idx *Index) LayersByLayerID(id uint32) []*aep.Layer {
	if idx == nil || id == 0 {
		return nil
	}
	return cloneLayers(idx.layersByLayerID[id])
}

// LayersByName returns all layers whose display name equals name.
func (idx *Index) LayersByName(name string) []*aep.Layer {
	if idx == nil || name == "" {
		return nil
	}
	return cloneLayers(idx.layersByName[name])
}

// LayersBySourceID returns all layers whose SourceID equals sourceID.
func (idx *Index) LayersBySourceID(sourceID uint32) []*aep.Layer {
	if idx == nil || sourceID == 0 {
		return nil
	}
	return cloneLayers(idx.layersBySourceID[sourceID])
}

// LayersByEffect returns all layers with an effect whose MatchName equals
// matchName. A layer appears at most once even if it has duplicate effects with
// that match name.
func (idx *Index) LayersByEffect(matchName string) []*aep.Layer {
	if idx == nil || matchName == "" {
		return nil
	}
	return cloneLayers(idx.layersByEffect[matchName])
}

func cloneLayers(layers []*aep.Layer) []*aep.Layer {
	if len(layers) == 0 {
		return nil
	}
	return append([]*aep.Layer(nil), layers...)
}
