package aep

import "github.com/example/aep-parser/internal/rifx"

// SetLayerCompForTest assigns the back-ref used by InsertLayer's R5 check.
// Test-only; production code never calls this.
func SetLayerCompForTest(l *Layer, c *Composition) { l.comp = c }

// SetCompProjForTest assigns the project back-ref used by InsertLayer's R3.
// Test-only.
func SetCompProjForTest(c *Composition, p *Project) { c.proj = p }

func ClearLayerLayrListForTest(l *Layer) {
	if l.back != nil {
		l.back.layrList = nil
	}
}

func CorruptSrcLayrFormTypeForTest(l *Layer) {
	if l.back != nil && l.back.layrList != nil {
		l.back.layrList.FormType = rifx.ChunkID{'X', 'X', 'X', 'X'}
	}
}

// ProjForTest exposes Composition.proj for InsertLayer tests.
func (c *Composition) ProjForTest() *Project { return c.proj }

// DestFootageByPathForTest exposes destFootageByPath for cross-Project tests.
func DestFootageByPathForTest(p *Project, path string) *Footage { return destFootageByPath(p, path) }

// LocateItemBlockByIDForTest exposes locateItemBlockByID over a Project's root Fold.
func LocateItemBlockByIDForTest(p *Project, id uint32) (int, int) {
	if p.back == nil || p.back.rootFold == nil {
		return -1, -1
	}
	return locateItemBlockByID(p.back.rootFold, id)
}

// ImportFootageBlockForTest exposes importFootageBlock for cross-Project tests.
func ImportFootageBlockForTest(dest, src *Project, srcID uint32, name string) (uint32, error) {
	return importFootageBlock(dest, src, srcID, name)
}
