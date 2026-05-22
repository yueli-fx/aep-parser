package aep

import "github.com/example/aep-parser/internal/rifx"

// White-box accessors for unexported Project fields.
// `_test.go` 后缀使这些方法仅在 test build 时编译，不污染 production binary。
func (p *Project) NextItemIDForTest() uint32    { return p.nextItemID }
func (p *Project) RootFoldForTest() *rifx.Chunk { return p.rootFold }
