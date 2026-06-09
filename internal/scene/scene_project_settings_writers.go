package scene

import "fmt"

// Project-level presence-encoded flag chunks (lnrb / lnrp). These sit as
// direct children of the root RIFX LIST; the setting is "on" when the chunk
// is present. Toggling adds/removes the chunk on root, so the serializer-side
// mutation (and the rifx-typed insert-position helper) lives in back_project.go;
// the scene getters/setters here delegate through the ProjectWriter interface.

// LinearBlending reports whether the project uses linear blending
// (presence of lnrb chunk under root).
func (p *Project) LinearBlending() bool {
	if p.back == nil {
		return false
	}
	return p.back.LinearBlendingFlag()
}

// SetLinearBlending toggles the lnrb chunk under root.
func (p *Project) SetLinearBlending(v bool) error {
	if p.back == nil {
		return fmt.Errorf("project: cannot toggle %s — no root chunk (built outside parser)", "lnrb")
	}
	return p.back.SetLinearBlending(v)
}

// LinearizeWorkingSpace reports whether the working color space is
// linearized for blending (presence of lnrp chunk under root).
func (p *Project) LinearizeWorkingSpace() bool {
	if p.back == nil {
		return false
	}
	return p.back.LinearizeWorkingSpaceFlag()
}

// SetLinearizeWorkingSpace toggles the lnrp chunk under root.
func (p *Project) SetLinearizeWorkingSpace(v bool) error {
	if p.back == nil {
		return fmt.Errorf("project: cannot toggle %s — no root chunk (built outside parser)", "lnrp")
	}
	return p.back.SetLinearizeWorkingSpace(v)
}
