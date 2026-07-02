package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeProjectEffects(project *aep.Project, prof *profile.Profile) (*aep.Project, error) {
	if !hasProfileEffects(prof) {
		return project, nil
	}
	reopened, err := aep.Reopen(project)
	if err != nil {
		return nil, fmt.Errorf("reopen for effects: %w", err)
	}
	for compIndex, sourceComp := range prof.Comps {
		targetComp, err := targetCompBySource(reopened, compIndex, sourceComp)
		if err != nil {
			return nil, err
		}
		for layerIndex, sourceLayer := range sourceComp.Layers {
			if len(sourceLayer.Effects) == 0 {
				continue
			}
			if layerIndex >= len(targetComp.Layers) {
				return nil, fmt.Errorf("comp %q effects: target layer index %d missing", sourceComp.Name, layerIndex)
			}
			targetLayer := targetComp.Layers[layerIndex]
			for _, sourceEffect := range sourceLayer.Effects {
				targetEffect, err := aep.AddEffect(targetLayer, sourceEffect.MatchName)
				if err != nil {
					return nil, fmt.Errorf("comp %q layer %q add effect %q: %w", sourceComp.Name, sourceLayer.Name, sourceEffect.MatchName, err)
				}
				if err := materializeEffectParams(targetComp, targetLayer, sourceLayer, targetEffect, sourceEffect); err != nil {
					return nil, fmt.Errorf("comp %q layer %q effect %q params: %w", sourceComp.Name, sourceLayer.Name, sourceEffect.MatchName, err)
				}
			}
		}
	}
	return reopened, nil
}
