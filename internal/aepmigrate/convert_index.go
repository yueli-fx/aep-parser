package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

type convertCompIndex struct {
	byID   map[uint32]profile.Composition
	byName map[string]profile.Composition
}

func newConvertCompIndex(prof *profile.Profile) convertCompIndex {
	out := convertCompIndex{
		byID:   map[uint32]profile.Composition{},
		byName: map[string]profile.Composition{},
	}
	if prof == nil {
		return out
	}
	for _, comp := range prof.Comps {
		if comp.ID != 0 {
			out.byID[comp.ID] = comp
		}
		if _, exists := out.byName[comp.Name]; !exists {
			out.byName[comp.Name] = comp
		}
	}
	return out
}

func (idx convertCompIndex) sourceComposition(layer profile.Layer) (profile.Composition, bool) {
	if layer.SourceRef == nil {
		return profile.Composition{}, false
	}
	var comp profile.Composition
	var ok bool
	if layer.SourceRef.ID != 0 {
		comp, ok = idx.byID[layer.SourceRef.ID]
	}
	if !ok && layer.SourceRef.Name != "" {
		comp, ok = idx.byName[layer.SourceRef.Name]
	}
	return comp, ok
}

func (idx convertTargetCompIndex) sourceComposition(layer profile.Layer) (*aep.Composition, bool) {
	if layer.SourceRef == nil {
		return nil, false
	}
	var comp *aep.Composition
	var ok bool
	if layer.SourceRef.ID != 0 {
		comp, ok = idx.bySourceID[layer.SourceRef.ID]
	}
	if !ok && layer.SourceRef.Name != "" {
		comp, ok = idx.byName[layer.SourceRef.Name]
	}
	return comp, ok && comp != nil
}

type convertTargetCompIndex struct {
	bySourceID map[uint32]*aep.Composition
	byName     map[string]*aep.Composition
}

func newConvertTargetCompIndex() convertTargetCompIndex {
	return convertTargetCompIndex{
		bySourceID: map[uint32]*aep.Composition{},
		byName:     map[string]*aep.Composition{},
	}
}

func (idx convertTargetCompIndex) add(source profile.Composition, target *aep.Composition) {
	if source.ID != 0 {
		idx.bySourceID[source.ID] = target
	}
	if _, exists := idx.byName[source.Name]; !exists {
		idx.byName[source.Name] = target
	}
}

type convertFootageIndex struct {
	byID   map[uint32]profile.Item
	byName map[string]profile.Item
}

func newConvertFootageIndex(prof *profile.Profile) convertFootageIndex {
	out := convertFootageIndex{
		byID:   map[uint32]profile.Item{},
		byName: map[string]profile.Item{},
	}
	if prof == nil {
		return out
	}
	for _, item := range prof.Items.Footage {
		if item.ID != 0 {
			out.byID[item.ID] = item
		}
		if _, exists := out.byName[item.Name]; !exists {
			out.byName[item.Name] = item
		}
	}
	return out
}

func (idx convertFootageIndex) solidDetails(layer profile.Layer) (profile.FootageDetails, bool) {
	if layer.SourceRef == nil {
		return profile.FootageDetails{}, false
	}
	var item profile.Item
	var ok bool
	if layer.SourceRef.ID != 0 {
		item, ok = idx.byID[layer.SourceRef.ID]
	}
	if !ok && layer.SourceRef.Name != "" {
		item, ok = idx.byName[layer.SourceRef.Name]
	}
	if !ok || item.Footage == nil || item.Footage.AssetType != "solid" {
		return profile.FootageDetails{}, false
	}
	return *item.Footage, true
}

func targetCompBySource(project *aep.Project, index int, source profile.Composition) (*aep.Composition, error) {
	if index < len(project.Compositions) {
		return project.Compositions[index], nil
	}
	for _, comp := range project.Compositions {
		if comp.Name == source.Name {
			return comp, nil
		}
	}
	return nil, fmt.Errorf("comp %q effects: target comp missing", source.Name)
}

func targetLayerBySourceRef(comp *aep.Composition, ref *profile.LayerRef) *aep.Layer {
	if ref == nil {
		return nil
	}
	if ref.Name != "" {
		if layer := comp.LayerByName(ref.Name); layer != nil {
			return layer
		}
	}
	if ref.Index >= 0 && ref.Index < len(comp.Layers) {
		return comp.Layers[ref.Index]
	}
	if ref.Index > 0 && ref.Index <= len(comp.Layers) {
		return comp.Layers[ref.Index-1]
	}
	return nil
}
