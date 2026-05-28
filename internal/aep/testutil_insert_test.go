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
