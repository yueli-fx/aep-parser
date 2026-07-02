package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

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
