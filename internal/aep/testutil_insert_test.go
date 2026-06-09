package aep

import (
	"github.com/example/aep-parser/internal/rifx"
	"github.com/example/aep-parser/internal/scene"
)

// SetLayerCompForTest assigns the back-ref used by InsertLayer's R5 check.
// Test-only; production code never calls this.
func SetLayerCompForTest(l *Layer, c *Composition) { scene.SetLayerComp(l, c) }

// SetCompProjForTest assigns the project back-ref used by InsertLayer's R3.
// Test-only.
func SetCompProjForTest(c *Composition, p *Project) { scene.SetCompositionProj(c, p) }

func ClearLayerLayrListForTest(l *Layer) {
	if lb := layerBack(l); lb != nil {
		lb.layrList = nil
	}
}

func CorruptSrcLayrFormTypeForTest(l *Layer) {
	if lb := layerBack(l); lb != nil && lb.layrList != nil {
		lb.layrList.FormType = rifx.ChunkID{'X', 'X', 'X', 'X'}
	}
}

// ProjForTest exposes Composition.proj for InsertLayer tests.
func ProjForTest(c *Composition) *Project { return scene.CompositionProj(c) }

// DestFootageByPathForTest exposes destFootageByPath for cross-Project tests.
func DestFootageByPathForTest(p *Project, path string) *Footage { return destFootageByPath(p, path) }

// LocateItemBlockByIDForTest exposes locateItemBlockByID over a Project's root
// Fold, returning (start, end) within the matched container; (-1,-1) if absent.
func LocateItemBlockByIDForTest(p *Project, id uint32) (int, int) {
	pb := projectBack(p)
	if pb == nil || pb.rootFold == nil {
		return -1, -1
	}
	_, s, e := locateItemBlockByID(pb.rootFold, id)
	return s, e
}

// ImportFootageBlockForTest exposes importFootageBlock for cross-Project tests.
func ImportFootageBlockForTest(dest, src *Project, srcID uint32, name string) (uint32, error) {
	return importFootageBlock(dest, src, srcID, name)
}
