package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func rebuildProject(target VersionLabel, prof *profile.Profile) (*aep.Project, error) {
	project := aep.NewProject(aepTarget(target))
	footage := newConvertFootageIndex(prof)
	comps := newConvertCompIndex(prof)
	targetComps := newConvertTargetCompIndex()
	for _, comp := range prof.Comps {
		next, err := aep.NewComposition(project, comp.Name, comp.Width, comp.Height, comp.FrameRate, comp.Duration)
		if err != nil {
			return nil, err
		}
		if err := applyStableCompSettings(next, comp); err != nil {
			return nil, err
		}
		targetComps.add(comp, next)
	}
	for _, comp := range prof.Comps {
		next, ok := targetComps.bySourceID[comp.ID]
		if !ok {
			next = targetComps.byName[comp.Name]
		}
		if next == nil {
			return nil, fmt.Errorf("comp %q: target comp missing after creation", comp.Name)
		}
		if err := rebuildCompLayers(target, comp, next, footage, comps, targetComps); err != nil {
			return nil, err
		}
	}
	return project, nil
}

func aepTarget(target VersionLabel) aep.AETarget {
	switch target {
	case VersionAE2021:
		return aep.TargetAE2021
	case VersionAE2022:
		return aep.TargetAE2022
	case VersionAE2023:
		return aep.TargetAE2023
	case VersionAE2024:
		return aep.TargetAE2024
	case VersionAE2025:
		return aep.TargetAE2025
	default:
		return aep.TargetAE2020
	}
}
