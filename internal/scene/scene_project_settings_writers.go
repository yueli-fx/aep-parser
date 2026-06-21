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

// @summary     Toggle the project's linear blending setting
// @param       v  true to enable linear blending, false to disable it
// @domain      project
// @stability   stable
// @verify      ae-accept
// @gate        TestProjectSettings_AEShipGate_AE2020,TestProjectSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    presence-encoded toggle: enabling adds the lnrb chunk under
//   root, disabling removes it. Verified on both AE versions, with the
//   DOM's app.project.linearBlending reading back true after a round-trip.
// @alias       linear blending,线性混合,lnrb,linear color blending
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

// @summary     Toggle the project's linearize-working-space setting
// @param       v  true to enable the linearized working space, false to disable it
// @domain      project
// @stability   stable
// @verify      roundtrip
// @since       AE2020
// @boundary    presence-encoded toggle: enabling adds the lnrp chunk under
//   root, disabling removes it. The DOM's app.project.linearizeWorkingSpace
//   is not a reliable verification surface — it reads false even on a file
//   AE itself saved and reloaded, because the value is derived from the
//   color-management profile rather than bound to the lnrp chunk. Acceptance
//   is therefore capped at byte-preservation against AE's own save, hence
//   roundtrip rather than ae-accept.
// @incident    project-flag-chunks-lnrb-lnrp
// @alias       linearize working space,线性化工作空间,lnrp,linear working space
func (p *Project) SetLinearizeWorkingSpace(v bool) error {
	if p.back == nil {
		return fmt.Errorf("project: cannot toggle %s — no root chunk (built outside parser)", "lnrp")
	}
	return p.back.SetLinearizeWorkingSpace(v)
}
