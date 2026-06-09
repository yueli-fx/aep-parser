package aep

import (
	"github.com/example/aep-parser/internal/rifx"
	"github.com/example/aep-parser/internal/scene"
	"github.com/example/aep-parser/internal/serializer"
)

// White-box accessors for unexported back-ref fields, kept here so the many
// `package aep_test` golden tests keep calling them unqualified. The actual
// back-ref state lives in internal/serializer post package-split, so these
// delegate there. `_test.go` 后缀使这些函数仅在 test build 时编译。
func NextItemIDForTest(p *Project) uint32 { return scene.ProjectNextItemID(p) }

func RootFoldForTest(p *Project) *rifx.Chunk { return serializer.ProjectRootFold(p) }

// ItemListForTest 暴露 Composition.itemList 给 golden tests 用。
func ItemListForTest(c *Composition) *rifx.Chunk { return serializer.CompItemListForTest(c) }

// RootForTest 暴露 Project.root 给 debug tests 用。
func RootForTest(p *Project) *rifx.Chunk { return serializer.ProjectRoot(p) }

// LdtaForTest 暴露 Layer.ldta 给 DeleteLayer / structural mutation tests 用
// (验 ldta @0x84 / @0xA0 / @0x6B byte-level writes).
func LdtaForTest(l *Layer) *rifx.Chunk { return serializer.LayerLdta(l) }

// NewTestProperty creates a Property with synthetic tdb4/tdsb/tdum/tduM
// chunks for white-box testing. All chunk refs are optional (pass nil to omit).
func NewTestProperty(matchName string, components int, tdb4, tdsb, tdum, tduM *rifx.Chunk) *Property {
	return serializer.NewTestProperty(matchName, components, tdb4, tdsb, tdum, tduM)
}

// MakeTdb4 builds a minimal 124-byte tdb4 chunk with the given flags.
// spatialStatic: byte 0x05; cvot: byte 0x0B; noValue: byte 0x39; typeFlags: byte 0x3B;
// dims: bytes 0x02-0x03 (big-endian uint16).
func MakeTdb4(dims uint16, spatialStatic, cvot, noValue, typeFlags byte) *rifx.Chunk {
	return serializer.MakeTdb4(dims, spatialStatic, cvot, noValue, typeFlags)
}
