package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func materializeLayerRefs(target VersionLabel, layers []convertLayerPair) error {
	bySourceID := map[uint32]*aep.Layer{}
	byName := map[string]*aep.Layer{}
	for _, pair := range layers {
		if pair.source.ID != 0 {
			bySourceID[pair.source.ID] = pair.target
		}
		if _, exists := byName[pair.source.Name]; !exists {
			byName[pair.source.Name] = pair.target
		}
	}
	for _, pair := range layers {
		if pair.source.ParentRef == nil {
			continue
		}
		parent := bySourceID[pair.source.ParentRef.ID]
		if parent == nil && pair.source.ParentRef.Name != "" {
			parent = byName[pair.source.ParentRef.Name]
		}
		if parent == nil {
			return fmt.Errorf("layer %q parent %q not found", pair.source.Name, pair.source.ParentRef.Name)
		}
		if err := pair.target.SetParent(parent.ID); err != nil {
			return fmt.Errorf("layer %q parent %q: %w", pair.source.Name, pair.source.ParentRef.Name, err)
		}
	}
	for _, pair := range layers {
		if pair.source.LightSourceRef == nil {
			continue
		}
		source := bySourceID[pair.source.LightSourceRef.ID]
		if source == nil && pair.source.LightSourceRef.Name != "" {
			source = byName[pair.source.LightSourceRef.Name]
		}
		if source == nil {
			return fmt.Errorf("layer %q light source %q not found", pair.source.Name, pair.source.LightSourceRef.Name)
		}
		if err := pair.target.SetLightSource(source); err != nil {
			return fmt.Errorf("layer %q light source %q: %w", pair.source.Name, pair.source.LightSourceRef.Name, err)
		}
	}
	for _, pair := range layers {
		mode, ok := convertTrackMatteMode(pair.source.Flags.TrackMatte)
		if !ok {
			return fmt.Errorf("layer %q track matte mode %d unsupported", pair.source.Name, pair.source.Flags.TrackMatte)
		}
		if mode == aep.TrackMatteNone {
			continue
		}
		if isExplicitMatteRef(pair.source) && supportsExplicitMatteTarget(target) {
			matte := bySourceID[pair.source.MatteRef.ID]
			if matte == nil && pair.source.MatteRef.Name != "" {
				matte = byName[pair.source.MatteRef.Name]
			}
			if matte == nil {
				return fmt.Errorf("layer %q matte source %q not found", pair.source.Name, pair.source.MatteRef.Name)
			}
			if err := pair.target.SetTrackMatteSource(matte, mode); err != nil {
				return fmt.Errorf("layer %q matte source %q: %w", pair.source.Name, pair.source.MatteRef.Name, err)
			}
			continue
		}
		if err := pair.target.SetTrackMatte(mode); err != nil {
			return fmt.Errorf("layer %q track matte: %w", pair.source.Name, err)
		}
	}
	return nil
}

func convertTrackMatteMode(mode uint8) (aep.TrackMatteType, bool) {
	switch mode {
	case uint8(aep.TrackMatteNone):
		return aep.TrackMatteNone, true
	case uint8(aep.TrackMatteAlpha):
		return aep.TrackMatteAlpha, true
	case uint8(aep.TrackMatteAlphaInverse):
		return aep.TrackMatteAlphaInverse, true
	case uint8(aep.TrackMatteLuma):
		return aep.TrackMatteLuma, true
	case uint8(aep.TrackMatteLumaInverse):
		return aep.TrackMatteLumaInverse, true
	default:
		return aep.TrackMatteNone, false
	}
}
