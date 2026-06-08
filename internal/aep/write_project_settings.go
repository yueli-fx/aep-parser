package aep

import (
	"fmt"

	"github.com/example/aep-parser/internal/rifx"
)

// Project-level presence-encoded flag chunks (lnrb / lnrp). These sit as
// direct children of the root RIFX LIST; the setting is "on" when the chunk
// is present. Toggling adds/removes the chunk on root, so this lives in the
// write stage (it mutates the chunk tree, unlike the in-place .Data patches
// for nnhd / adfr / dwga which stay in scene_project_settings.go).

// rootChunkPresent reports whether root has a direct child chunk
// (not LIST) with the given id.
func (p *Project) rootChunkPresent(id rifx.ChunkID) bool {
	pb := p.projectBack()
	if pb == nil || pb.root == nil {
		return false
	}
	for _, c := range pb.root.Children {
		if !c.IsList() && c.ID == id {
			return true
		}
	}
	return false
}

// flagChunkInsertPosition returns the index where a new lnrb / lnrp
// flag chunk should be inserted on root. Per RE: after `cpid`,
// otherwise before `dwga`, otherwise at end.
func flagChunkInsertPosition(root *rifx.Chunk) int {
	for i, c := range root.Children {
		if !c.IsList() && c.ID == chunkIDCpid {
			return i + 1
		}
	}
	for i, c := range root.Children {
		if !c.IsList() && c.ID == rifx.IDDwga {
			return i
		}
	}
	return len(root.Children)
}

// chunkIDCpid is the root-level color-profile id chunk, used as the
// insertion anchor for lnrb / lnrp flag chunks. Not exposed in rifx
// since no other code path needs it.
var chunkIDCpid = rifx.ChunkID{'c', 'p', 'i', 'd'}

// LinearBlending reports whether the project uses linear blending
// (presence of lnrb chunk under root).
func (p *Project) LinearBlending() bool {
	return p.rootChunkPresent(rifx.IDLnrb)
}

// SetLinearBlending toggles the lnrb chunk under root.
func (p *Project) SetLinearBlending(v bool) error {
	if p.back == nil {
		return fmt.Errorf("project: cannot toggle %s — no root chunk (built outside parser)", rifx.IDLnrb)
	}
	return p.back.SetLinearBlending(v)
}

// LinearizeWorkingSpace reports whether the working color space is
// linearized for blending (presence of lnrp chunk under root).
func (p *Project) LinearizeWorkingSpace() bool {
	return p.rootChunkPresent(rifx.IDLnrp)
}

// SetLinearizeWorkingSpace toggles the lnrp chunk under root.
func (p *Project) SetLinearizeWorkingSpace(v bool) error {
	if p.back == nil {
		return fmt.Errorf("project: cannot toggle %s — no root chunk (built outside parser)", rifx.IDLnrp)
	}
	return p.back.SetLinearizeWorkingSpace(v)
}
