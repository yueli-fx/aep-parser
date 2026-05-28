package aep

// SetLayerCompForTest assigns the back-ref used by InsertLayer's R5 check.
// Test-only; production code never calls this.
func SetLayerCompForTest(l *Layer, c *Composition) { l.comp = c }

// SetCompProjForTest assigns the project back-ref used by InsertLayer's R3.
// Test-only.
func SetCompProjForTest(c *Composition, p *Project) { c.proj = p }
