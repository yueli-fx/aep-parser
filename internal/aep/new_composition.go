// internal/aep/new_composition.go
//
// V2: Project.NewComposition + chunk builders for synthesizing comp
// Item LIST from scratch. See workshop/specs/v2-1-foundation-design.md
// and workshop/plans/v2-1-foundation-plan.md for design + RE findings.
package aep

import (
	"encoding/binary"
	"math"

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

// buildCompCdta 构造 204-byte cdta，5 个必填 + AE 默认值。
// Frame rate 经 encodeFrameRate 走 NTSC canonical 表（RE-4）。
// WorkAreaEnd 写 sentinel 0xFFFFFFFF（per RE-4 finding）。
func buildCompCdta(w, h uint16, fps, duration float64) []byte {
	d := make([]byte, cdtaSize)

	// ResolutionFactor @0x00/0x02 default [1,1]
	binary.BigEndian.PutUint16(d[cdtaResolutionFactorX:cdtaResolutionFactorX+2], 1)
	binary.BigEndian.PutUint16(d[cdtaResolutionFactorY:cdtaResolutionFactorY+2], 1)

	// TickRate @0x08 — AE 25 新 comp 写 0x400=1024（实测 AE2025 fixture）；fallback safe value
	binary.BigEndian.PutUint32(d[cdtaTickRate:cdtaTickRate+4], 1024)

	// WorkArea @0x1C..@0x2B：start=0/600, end=sentinel/600
	binary.BigEndian.PutUint32(d[cdtaWorkAreaStart:cdtaWorkAreaStart+4], 0)
	binary.BigEndian.PutUint32(d[cdtaWorkAreaStartDiv:cdtaWorkAreaStartDiv+4], 600)
	binary.BigEndian.PutUint32(d[cdtaWorkAreaEnd:cdtaWorkAreaEnd+4], 0xFFFFFFFF) // sentinel "use Duration"
	binary.BigEndian.PutUint32(d[cdtaWorkAreaEndDiv:cdtaWorkAreaEndDiv+4], 600)

	// BGColor @0x34..@0x36 default {0,0,0}（bytes 已 0）

	// Width @0x8C / Height @0x8E
	binary.BigEndian.PutUint16(d[cdtaWidth:cdtaWidth+2], w)
	binary.BigEndian.PutUint16(d[cdtaHeight:cdtaHeight+2], h)

	// PixelAspect @0x90/0x94 default 1/1
	binary.BigEndian.PutUint32(d[cdtaPixelAspectNum:cdtaPixelAspectNum+4], 1)
	binary.BigEndian.PutUint32(d[cdtaPixelAspectDen:cdtaPixelAspectDen+4], 1)

	// FrameRate @0x9C/0x9E — canonical encoding
	enc := encodeFrameRate(fps)
	binary.BigEndian.PutUint16(d[cdtaFrameRateWhole:cdtaFrameRateWhole+2], enc.whole)
	binary.BigEndian.PutUint16(d[cdtaFrameRateFrac:cdtaFrameRateFrac+2], enc.frac)

	// DisplayStartTime @0xA4/0xA8 default 0/1
	binary.BigEndian.PutUint32(d[cdtaDisplayStartTime:cdtaDisplayStartTime+4], 0)
	binary.BigEndian.PutUint32(d[cdtaDisplayStartDiv:cdtaDisplayStartDiv+4], 1)

	// ShutterAngle @0xAE default 180
	binary.BigEndian.PutUint16(d[cdtaShutterAngle:cdtaShutterAngle+2], 180)

	// Duration @0xB0 = round(duration_seconds * fps) frames
	durationFrames := uint32(math.Round(duration * fps))
	binary.BigEndian.PutUint32(d[cdtaDuration:cdtaDuration+4], durationFrames)

	// ShutterPhase @0xB4 default 0（已是 0）

	// MotionBlurAdaptive @0xC4 default 128
	binary.BigEndian.PutUint32(d[cdtaMotionBlurAdaptive:cdtaMotionBlurAdaptive+4], 128)

	// MotionBlurSamples @0xC8 default 16
	binary.BigEndian.PutUint32(d[cdtaMotionBlurSamples:cdtaMotionBlurSamples+4], 16)

	return d
}

// buildEmptyLayrList 构造空 LIST formType=Layr（0 children）。
func buildEmptyLayrList() *rifx.Chunk {
	return &rifx.Chunk{
		ID:       rifx.IDList,
		FormType: rifx.IDLayr,
		Children: nil,
	}
}
