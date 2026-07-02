package aepmigrate

import (
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

type convertProjectMaterializer struct {
	name  string
	apply func(*aep.Project, *profile.Profile) (*aep.Project, error)
}

func materializeRebuiltProject(project *aep.Project, prof *profile.Profile) (*aep.Project, error) {
	var err error
	for _, step := range convertProjectMaterializers() {
		project, err = step.apply(project, prof)
		if err != nil {
			return nil, err
		}
	}
	return project, nil
}

func convertProjectMaterializers() []convertProjectMaterializer {
	return []convertProjectMaterializer{
		{name: "comp_metadata", apply: applyNoLayerCompMetadata},
		{name: "transform_expressions", apply: materializeProjectTransformExpressions},
		{name: "text_animators", apply: materializeProjectTextAnimators},
		{name: "masks", apply: materializeProjectMasks},
		{name: "effects", apply: materializeProjectEffects},
	}
}
