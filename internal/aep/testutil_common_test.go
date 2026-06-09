package aep

import (
	"github.com/example/aep-parser/internal/rifx"
	"github.com/example/aep-parser/internal/scene"
)

// White-box accessors for unexported fields. Free functions (the receivers are
// scene types post package-split); the call form is `NextItemIDForTest(p)`.
// `_test.go` 后缀使这些函数仅在 test build 时编译，不污染 production binary。
func NextItemIDForTest(p *Project) uint32 { return scene.ProjectNextItemID(p) }

func RootFoldForTest(p *Project) *rifx.Chunk {
	pb := projectBack(p)
	if pb == nil {
		return nil
	}
	return pb.rootFold
}

// ItemListForTest 暴露 Composition.itemList 给 golden tests 用。
func ItemListForTest(c *Composition) *rifx.Chunk {
	cb := compositionBack(c)
	if cb == nil {
		return nil
	}
	return cb.itemList
}

// RootForTest 暴露 Project.root 给 debug tests 用。
func RootForTest(p *Project) *rifx.Chunk {
	pb := projectBack(p)
	if pb == nil {
		return nil
	}
	return pb.root
}

// LdtaForTest 暴露 Layer.ldta 给 DeleteLayer / structural mutation tests 用
// (验 ldta @0x84 / @0xA0 / @0x6B byte-level writes).
func LdtaForTest(l *Layer) *rifx.Chunk {
	lb := layerBack(l)
	if lb == nil {
		return nil
	}
	return lb.ldta
}

// NewTestProperty creates a Property with synthetic tdb4/tdsb/tdum/tduM
// chunks for white-box testing. All chunk refs are optional (pass nil to omit).
func NewTestProperty(matchName string, components int, tdb4, tdsb, tdum, tduM *rifx.Chunk) *Property {
	p := &Property{MatchName: matchName, Name: matchName, Components: components}
	scene.SetPropertyBack(p, &propertyBackrefs{
		tdb4: tdb4,
		tdsb: tdsb,
		tdum: tdum,
		tduM: tduM,
	})
	return p
}

// MakeTdb4 builds a minimal 124-byte tdb4 chunk with the given flags.
// spatialStatic: byte 0x05; cvot: byte 0x0B; noValue: byte 0x39; typeFlags: byte 0x3B;
// dims: bytes 0x02-0x03 (big-endian uint16).
func MakeTdb4(dims uint16, spatialStatic, cvot, noValue, typeFlags byte) *rifx.Chunk {
	d := make([]byte, 124)
	d[0] = 0xDB
	d[1] = 0x99
	d[2] = byte(dims >> 8)
	d[3] = byte(dims)
	d[0x05] = spatialStatic
	d[0x0B] = cvot
	d[0x39] = noValue
	d[0x3B] = typeFlags
	return &rifx.Chunk{ID: rifx.IDtdb4, Size: 124, Data: d}
}
