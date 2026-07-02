package aepmigrate

import (
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

type convertLayerPair struct {
	source profile.Layer
	target *aep.Layer
}

func materializeLayerMetadata(layer *aep.Layer, source profile.Layer) error {
	if source.Label != 0 && source.Label != layer.Label {
		if err := layer.SetLabel(source.Label); err != nil {
			return err
		}
	}
	if source.Comment != "" {
		if err := layer.SetComment(source.Comment); err != nil {
			return err
		}
	}
	return nil
}

func materializeNullLayer(comp *aep.Composition, source profile.Layer, footage convertFootageIndex) (*aep.Layer, error) {
	if solid, ok := footage.solidDetails(source); ok && solid.Width != 0 && solid.Height != 0 && solid.SolidColor != nil {
		layer, err := aep.NewSolidLayer(comp, source.Name, int(solid.Width), int(solid.Height), *solid.SolidColor)
		if err != nil {
			return nil, err
		}
		if err := layer.SetIsNull(true); err != nil {
			return nil, err
		}
		return layer, nil
	}
	return aep.NewNullLayer(comp, source.Name)
}
