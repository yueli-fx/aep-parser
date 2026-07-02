package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeProjectMasks(project *aep.Project, prof *profile.Profile) (*aep.Project, error) {
	if !hasProfileMasks(prof) {
		return project, nil
	}
	reopened, err := aep.Reopen(project)
	if err != nil {
		return nil, fmt.Errorf("reopen for masks: %w", err)
	}
	for compIndex, sourceComp := range prof.Comps {
		targetComp, err := targetCompBySource(reopened, compIndex, sourceComp)
		if err != nil {
			return nil, err
		}
		for layerIndex, sourceLayer := range sourceComp.Layers {
			if len(sourceLayer.Masks) == 0 {
				continue
			}
			if layerIndex >= len(targetComp.Layers) {
				return nil, fmt.Errorf("comp %q masks: target layer index %d missing", sourceComp.Name, layerIndex)
			}
			targetLayer := targetComp.Layers[layerIndex]
			for maskIndex, sourceMask := range sourceLayer.Masks {
				mask, err := aep.AddMask(targetLayer, sourceMask.Name, maskBezierPathFromProfile(sourceMask))
				if err != nil {
					return nil, fmt.Errorf("comp %q layer %q add mask %d: %w", sourceComp.Name, sourceLayer.Name, maskIndex, err)
				}
				if err := materializeMaskSurface(targetLayer, mask, sourceMask); err != nil {
					return nil, fmt.Errorf("comp %q layer %q mask %q: %w", sourceComp.Name, sourceLayer.Name, sourceMask.Name, err)
				}
			}
		}
	}
	return reopened, nil
}
