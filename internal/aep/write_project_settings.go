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
	if p.back == nil || p.back.root == nil {
		return false
	}
	for _, c := range p.back.root.Children {
		if !c.IsList() && c.ID == id {
			return true
		}
	}
	return false
}

// setRootFlagChunk adds (when on=true) or removes (when on=false) a
// presence-encoded flag chunk on root.
//
// **Layout** (against AE 2025-saved fixture
// `re_linear_blending_on.aep`): the chunk carries **1 byte 0x01**, NOT
// zero-length. Empty payload makes AE reject the file with "文件数据
// 丢失"/"file data missing".
//
// **Position** matters: AE inserts lnrb / lnrp immediately AFTER the
// root `cpid` chunk (color-management profile id) and before `dwga`.
// Append-to-end likewise causes "file data missing". We insert right
// after the existing cpid (or fall back to before dwga, or append if
// neither anchor exists).
func (p *Project) setRootFlagChunk(id rifx.ChunkID, on bool) error {
	if p.back == nil || p.back.root == nil {
		return fmt.Errorf("project: cannot toggle %s — no root chunk (built outside parser)", id)
	}
	idx := -1
	for i, c := range p.back.root.Children {
		if !c.IsList() && c.ID == id {
			idx = i
			break
		}
	}
	if on && idx < 0 {
		insertAt := flagChunkInsertPosition(p.back.root)
		newChunk := &rifx.Chunk{ID: id, Data: []byte{0x01}}
		p.back.root.Children = append(p.back.root.Children, nil)
		copy(p.back.root.Children[insertAt+1:], p.back.root.Children[insertAt:])
		p.back.root.Children[insertAt] = newChunk
	}
	if !on && idx >= 0 {
		p.back.root.Children = append(p.back.root.Children[:idx], p.back.root.Children[idx+1:]...)
	}
	return nil
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
	return p.setRootFlagChunk(rifx.IDLnrb, v)
}

// LinearizeWorkingSpace reports whether the working color space is
// linearized for blending (presence of lnrp chunk under root).
func (p *Project) LinearizeWorkingSpace() bool {
	return p.rootChunkPresent(rifx.IDLnrp)
}

// SetLinearizeWorkingSpace toggles the lnrp chunk under root.
func (p *Project) SetLinearizeWorkingSpace(v bool) error {
	return p.setRootFlagChunk(rifx.IDLnrp, v)
}
