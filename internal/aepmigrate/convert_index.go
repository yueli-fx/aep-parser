package aepmigrate

import (
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
