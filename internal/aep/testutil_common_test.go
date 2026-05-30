package aep

import "github.com/example/aep-parser/internal/rifx"

// White-box accessors for unexported fields.
// `_test.go` 后缀使这些方法仅在 test build 时编译，不污染 production binary。
func (p *Project) NextItemIDForTest() uint32 { return p.nextItemID }
func (p *Project) RootFoldForTest() *rifx.Chunk {
	if p.back == nil {
		return nil
	}
	return p.back.rootFold
}

// ItemListForTest 暴露 Composition.itemList 给 Phase 5+ golden tests 用。
func (c *Composition) ItemListForTest() *rifx.Chunk {
	if c.back == nil {
		return nil
	}
	return c.back.itemList
}

// RootForTest 暴露 Project.root 给 debug tests 用。
func (p *Project) RootForTest() *rifx.Chunk {
	if p.back == nil {
		return nil
	}
	return p.back.root
}

// LdtaForTest 暴露 Layer.ldta 给 DeleteLayer / structural mutation tests 用
// (验 ldta @0x84 / @0xA0 / @0x6B byte-level writes).
func (l *Layer) LdtaForTest() *rifx.Chunk {
	if l.back == nil {
		return nil
	}
	return l.back.ldta
}

// NewTestProperty creates a Property with synthetic tdb4/tdsb/tdum/tduM
// chunks for white-box testing. All chunk refs are optional (pass nil to omit).
func NewTestProperty(matchName string, components int, tdb4, tdsb, tdum, tduM *rifx.Chunk) *Property {
	p := &Property{MatchName: matchName, Name: matchName, Components: components}
	p.back = &propertyBackrefs{
		tdb4: tdb4,
		tdsb: tdsb,
		tdum: tdum,
		tduM: tduM,
	}
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
