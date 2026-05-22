// internal/aep/new_composition.go
//
// V2: Project.NewComposition + chunk builders for synthesizing comp
// Item LIST from scratch. See workshop/specs/v2-1-foundation-design.md
// and workshop/plans/v2-1-foundation-plan.md for design + RE findings.
package aep

import (
	"encoding/binary"

	"github.com/example/aep-parser/internal/rifx"
)

// buildCompIide 构造 4-byte iide chunk（Item index entry header）。
// 内容来自 AE 2025 saved fixture dump（实测固定 0x01000000）。语义未 RE，但
// AE 跨 fixture 都接受相同字节。
func buildCompIide() *rifx.Chunk {
	return &rifx.Chunk{
		ID:   rifx.ChunkID{'i', 'i', 'd', 'e'},
		Data: []byte{0x01, 0x00, 0x00, 0x00},
	}
}

var _ = binary.BigEndian // keep import for downstream tasks
