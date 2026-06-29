package serializer

// Test-support accessors exposing serializer-internal back-ref state to
// white-box tests in this package and to the aep facade's test helpers
// (which delegate here). These live in a non-test file so they are visible
// across the package boundary to internal/aep test code; internal/serializer
// is not importable outside the module, so this does not widen the public API.

import (
	"github.com/yueli-fx/aep-parser/internal/rifx"
	"github.com/yueli-fx/aep-parser/internal/scene"
)

// ProjectRootFold returns the project's cached root Fold LIST chunk, or nil.
func ProjectRootFold(p *Project) *rifx.Chunk {
	pb := projectBack(p)
	if pb == nil {
		return nil
	}
	return pb.rootFold
}

// ProjectRoot returns the project's root RIFX chunk, or nil.
func ProjectRoot(p *Project) *rifx.Chunk {
	pb := projectBack(p)
	if pb == nil {
		return nil
	}
	return pb.root
}

// ProjectVersionString decodes the AE version string from the project head
// chunk (empty when absent).
func ProjectVersionString(p *Project) string {
	pb := projectBack(p)
	if pb == nil {
		return ""
	}
	return pb.VersionString()
}

// LayerLdta returns the layer's ldta chunk, or nil.
func LayerLdta(l *Layer) *rifx.Chunk {
	lb := layerBack(l)
	if lb == nil {
		return nil
	}
	return lb.ldta
}

// ClearLayerLayrList nils the layer's cached Layr LIST back-ref (simulates a
// detached layer for InsertLayer's R5 check).
func ClearLayerLayrList(l *Layer) {
	if lb := layerBack(l); lb != nil {
		lb.layrList = nil
	}
}

// CorruptSrcLayrFormType replaces the layer's Layr LIST form type with a
// sentinel so structural-mutation guards reject it.
func CorruptSrcLayrFormType(l *Layer) {
	if lb := layerBack(l); lb != nil && lb.layrList != nil {
		lb.layrList.FormType = rifx.ChunkID{'X', 'X', 'X', 'X'}
	}
}

// DestFootageByPath resolves a footage item in p by its source file path.
func DestFootageByPath(p *Project, path string) *Footage { return destFootageByPath(p, path) }

// LocateItemBlockByID returns the (start, end) span of the item block with the
// given ID within the project's root Fold, or (-1, -1) when absent.
func LocateItemBlockByID(p *Project, id uint32) (int, int) {
	pb := projectBack(p)
	if pb == nil || pb.rootFold == nil {
		return -1, -1
	}
	_, s, e := locateItemBlockByID(pb.rootFold, id)
	return s, e
}

// ImportFootageBlock copies the footage item srcID from src into dest under the
// given name, returning the new item ID.
func ImportFootageBlock(dest, src *Project, srcID uint32, name string) (uint32, error) {
	return importFootageBlock(dest, src, srcID, name)
}

// NewTestProperty builds a Property with synthetic tdb4/tdsb/tdum/tduM chunks
// for white-box testing. All chunk refs are optional (pass nil to omit).
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
