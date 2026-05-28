package aep

import (
	"fmt"

	"github.com/example/aep-parser/internal/rifx"
)

// InsertLayer deep-clones src (from a sibling comp in the SAME Project)
// into c.Layers at atIdx. ALPHA — pending AE 2020+2025 ship-gate
// (3 modes × 2 versions = 6 PASS) before Stable promotion. See
// flightdeck/specs/2026-05-29-v3-phase5c-insertlayer-design.md.
func (c *Composition) InsertLayer(src *Layer, atIdx int) (*Layer, error) {
	if src == nil {
		return nil, fmt.Errorf("InsertLayer: src cannot be nil")
	}
	if c.back == nil || c.back.itemList == nil {
		return nil, fmt.Errorf("InsertLayer: dest comp %q has no itemList back-ref (built outside parser?)", c.Name)
	}
	if c.proj == nil {
		return nil, fmt.Errorf("InsertLayer: dest comp %q has no project back-ref", c.Name)
	}
	if atIdx < 0 || atIdx > len(c.Layers) {
		return nil, fmt.Errorf("InsertLayer: atIdx %d out of range (have %d layers; %d is append)", atIdx, len(c.Layers), len(c.Layers))
	}
	if src.comp == nil {
		return nil, fmt.Errorf("InsertLayer: src.comp is nil (layer detached from any comp)")
	}
	if src.comp == c {
		return nil, fmt.Errorf("InsertLayer: src and dest are the same comp %q — use DuplicateLayer instead", c.Name)
	}
	if src.comp.proj != c.proj {
		return nil, fmt.Errorf("InsertLayer: src and dest in different Projects — cross-Project insert deferred to Phase 5C.1")
	}
	if src.Type != LayerTypeAV {
		return nil, fmt.Errorf("InsertLayer: refuse non-AV src (Type=%s); only AV layers supported in Phase 5C", src.Type)
	}
	if src.SourceID != 0 && src.SourceID == c.ID {
		return nil, fmt.Errorf("InsertLayer: refuse direct pre-comp loop (src.SourceID=%d == dest.ID=%d)", src.SourceID, c.ID)
	}
	if src.back == nil || src.back.layrList == nil {
		return nil, fmt.Errorf("InsertLayer: src layer %q has no Layr chunk back-ref", src.Name)
	}
	srcChildren := src.comp.back.itemList.Children
	srcLayrIdx := findLayrIndexInItemList(src.comp.back.itemList, src.back.layrList)
	if srcLayrIdx < 0 {
		return nil, fmt.Errorf("InsertLayer: src layer %q Layr chunk not found in its comp's itemList", src.Name)
	}
	if !srcChildren[srcLayrIdx].IsList() || srcChildren[srcLayrIdx].FormType != rifx.IDLayr {
		return nil, fmt.Errorf("InsertLayer: src layer %q backref points to non-Layr chunk (FormType=%s)", src.Name, chunkIDString(srcChildren[srcLayrIdx].FormType))
	}
	if srcLayrIdx+1 >= len(srcChildren) {
		return nil, fmt.Errorf("InsertLayer: src layer %q Layr at end of itemList (no Ewst sibling)", src.Name)
	}
	if !srcChildren[srcLayrIdx+1].IsList() || srcChildren[srcLayrIdx+1].FormType != rifx.IDEwst {
		return nil, fmt.Errorf("InsertLayer: src layer %q expected Ewst sibling after Layr, found %s", src.Name, chunkIDString(srcChildren[srcLayrIdx+1].FormType))
	}
	return nil, fmt.Errorf("InsertLayer: not yet implemented (happy path forthcoming in Phase B)")
}
