package scene

// Project filter views and lookups — mirror of py-aep's //nolint:jargon
// `project.footages`, `project.root_folder`, `project.layer_by_id`,
// `project.effect_names`. Most are wrappers over existing direct-field
// access (Project.Footage / Project.Folders); LayerByID and EffectNames
// add new functionality.

// Footages returns the project's footage items (alias of the
// Project.Footage slice field, provided as a convenience).
func (p *Project) Footages() []*Footage {
	return p.Footage
}

// RootFolder returns the project's root folder. AE projects always
// have a root folder at the top of the items tree; we return the first
// Folder slice entry, or nil when the project was built without
// folders (synthesized minimal fixtures).
func (p *Project) RootFolder() *Folder {
	if len(p.Folders) == 0 {
		return nil
	}
	return p.Folders[0]
}

// LayerByID searches every composition in the project for a layer
// with the matching ID and returns the first hit. Returns nil for
// id == 0 (the "no layer" sentinel) and when no comp owns this id.
func (p *Project) LayerByID(id uint32) *Layer {
	if id == 0 {
		return nil
	}
	for _, c := range p.Compositions {
		if l := c.LayerByID(id); l != nil {
			return l
		}
	}
	return nil
}

// EffectNames returns the match-names of every effect referenced by
// the project, from the root-level `Pefl` LIST → `pjef` Utf8 entries.
// Returns nil when the project has no Pefl LIST (e.g., projects with
// no effects applied anywhere or builder-synthesized projects).
func (p *Project) EffectNames() []string {
	if p.back == nil {
		return nil
	}
	return p.back.EffectNames()
}
