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
//
//aep:cap domain=project tier=stable verify=ae-accept gate=TestProjectSettings_AEShipGate_AE2020,TestProjectSettings_AEShipGate_AE2025 boundary="presence-encoded toggle(lnrb chunk 增删);双版本 AE gated(project_settings,app.project.linearBlending DOM readback=true)" alias="linear blending,线性混合,lnrb,linear color blending"
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
//
//aep:cap domain=project tier=stable verify=roundtrip boundary="presence-encoded toggle(lnrp chunk 增删);⚠ AE DOM 不可值验(已知):app.project.linearizeWorkingSpace 即便 AE 自存文件 reload 也读 false——OCIO/CMS profile derived state,ScriptingAPI 不绑 lnrp chunk(详 incident project-flag-chunks-lnrb-lnrp § lnrp readback quirk)。封顶 acceptance(byte-preservation,字节同 AE 自存),留 roundtrip" alias="linearize working space,线性化工作空间,lnrp,linear working space"
func (p *Project) SetLinearizeWorkingSpace(v bool) error {
	if p.back == nil {
		return fmt.Errorf("project: cannot toggle %s — no root chunk (built outside parser)", "lnrp")
	}
	return p.back.SetLinearizeWorkingSpace(v)
}
