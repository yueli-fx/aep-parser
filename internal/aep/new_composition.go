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

// buildCompIdpc 构造 8-byte idpc chunk。
//
// RE-2 finding (workshop/plans/v2-1-foundation-plan.md Task 0.2): AE 自己写全零，
// 多 comp 同 project 也共享同样 8B 零字节。idpc 不是 per-item UUID — 真正的 Item
// 唯一性走 idta @0x10（由 nextItemID 分配）。
//
// 我们写全零跟 AE 行为一致，不引 crypto/rand。
func buildCompIdpc() *rifx.Chunk {
	return &rifx.Chunk{
		ID:   rifx.ChunkID{'i', 'd', 'p', 'c'},
		Data: make([]byte, 8), // 全 0
	}
}

// idtaCompDefaultBytes: 84 字节 idta default，从 AE 2025 saved 1-comp fixture
// (test_data/fdta_probe/AE2025_1comp.aep) verbatim 拷贝。Builder 只 overwrite
// @0x10 (Item ID) + 显式归零 @0x3A (Label)；其余字节保持 AE-default 不动（含
// @0x14..0x17 const 0x20、@0x3A 在 fixture 中残留=0x0F 我们 reset 为 0、
// @0x50 session token 等 RE-3 未完全语义化字段）。
//
// 注意 ID offset 是 @0x10 (parse.classifyItem reads idta.U32(16))，不是 @0x14。
// 早期文档把 @0x14 当作 ID — 实测 fixture (ID 1, 13) 与 parse 行为一致 @0x10。
var idtaCompDefaultBytes = [idtaSize]byte{
	// @0x00..0x07: type=0x04 (Composition) + padding
	0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	// @0x08..0x0F: padding
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	// @0x10..0x17: ID(@0x10)=1 (overwritten by builder) + const 0x20 (@0x14..0x17)
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x20,
	// @0x18..0x1F: padding
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	// @0x20..0x27: padding
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	// @0x28..0x2F: padding
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	// @0x30..0x37: padding
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	// @0x38..0x3F: @0x3A=0x0F in fixture (residual from saved-state Label?
	//   builder resets to 0 since users haven't called SetLabel)
	0x00, 0x00, 0x0F, 0x00, 0x00, 0x00, 0x00, 0x00,
	// @0x40..0x47: padding
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	// @0x48..0x4F: padding
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	// @0x50..0x53: session token (AE may rewrite; we keep fixture value)
	0xE6, 0x35, 0x5E, 0x03,
}

// buildCompIdta 构造 84-byte idta，template-copy default + overwrite @0x10 Item ID。
// 同时 reset @0x3A (Label) 为 0，确保 NewComposition 的 comp 默认无 label color。
func buildCompIdta(itemID uint32) *rifx.Chunk {
	data := make([]byte, idtaSize)
	copy(data, idtaCompDefaultBytes[:])
	binary.BigEndian.PutUint32(data[idtaItemID:idtaItemID+4], itemID)
	data[idtaLabel] = 0 // 显式归零（fixture 残留 0x0F）
	return &rifx.Chunk{
		ID:   rifx.ChunkID{'i', 'd', 't', 'a'},
		Data: data,
	}
}
