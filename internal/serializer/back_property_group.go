package serializer

import (
	"github.com/example/aep-parser/internal/rifx"
	"github.com/example/aep-parser/internal/scene"
)

// propertyGroupBackrefs holds the rifx.Chunk reference behind an
// AEPropertyGroup — the underlying tdgp LIST that the structural property
// ops (Remove / Duplicate / MoveTo / SetDimensionsSeparated) splice. Lives
// in a separate shard so AEPropertyGroup stays chunk-free.
//
// Always allocated for groups built by the parser's tree builders (see
// parse_property_group.go) — the synthetic root gets an empty shard
// (chunk == nil), every named subgroup gets its tdgp payload. Structural
// mutate ops reach it via group.back.chunk.
type propertyGroupBackrefs struct {
	// chunk is the group's underlying tdgp LIST; nil for the synthetic root.
	chunk *rifx.Chunk
}

var _ PropertyGroupWriter = (*propertyGroupBackrefs)(nil)

func (b *propertyGroupBackrefs) IsPropertyGroupWriter() {}

// propertyGroupBack returns the concrete backrefs behind an AEPropertyGroup's
// writer interface for serializer-stage (parse_/mutate_) tdgp LIST access.
// Returns nil when the group was built outside the parser's tree builders.
func propertyGroupBack(g *AEPropertyGroup) *propertyGroupBackrefs {
	if b, ok := scene.PropertyGroupBack(g).(*propertyGroupBackrefs); ok {
		return b
	}
	return nil
}
