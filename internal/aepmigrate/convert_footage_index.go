package aepmigrate

import "github.com/yueli-fx/aep-parser/internal/profile"

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
