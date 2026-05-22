// internal/aep/new_composition.go
//
// V2: Project.NewComposition + chunk builders for synthesizing comp
// Item LIST from scratch. See workshop/specs/v2-1-foundation-design.md
// and workshop/plans/v2-1-foundation-plan.md for design + RE findings.
package aep

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"fmt"
	"math"
	"sync"

	"github.com/example/aep-parser/internal/rifx"
)

//go:embed templates/2025_dummy_comp.aep
var embeddedDummyCompTemplate []byte

var (
	templateInit              sync.Once
	templateCompItemChunks    []*rifx.Chunk
	templateItemSiblingChunks []*rifx.Chunk // 8 chunks AE 25 expects after each Item in Fold (FEE LIST + fvdv/fiop/ftts/foac/fiac/fipc/fifl)
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

// buildCompCdta 构造 204-byte cdta。
// Frame rate 经 encodeFrameRate 走 NTSC canonical 表 (RE-4)。
// fps-derived timing 字段 (TickRate / mirrors / masterTicks) 经 lookupFpsTiming
// 查表 —— 这些字段 parser 不读，但 AE 25 打开做时间轴 sanity check 时必读，
// 全零会让 AE 25 crash (实测 Phase 6 ship gate, 2026-05-22)。
// WorkAreaEnd 写 sentinel 0xFFFFFFFF (per RE-4 finding)。
func buildCompCdta(w, h uint16, fps, duration float64) []byte {
	d := make([]byte, cdtaSize)

	timing := lookupFpsTiming(fps)

	// ResolutionFactor @0x00/0x02 default [1,1]
	binary.BigEndian.PutUint16(d[cdtaResolutionFactorX:cdtaResolutionFactorX+2], 1)
	binary.BigEndian.PutUint16(d[cdtaResolutionFactorY:cdtaResolutionFactorY+2], 1)

	// fps timing prologue @0x06 / @0x08 / @0x30
	binary.BigEndian.PutUint16(d[cdtaTicksPerFrame:cdtaTicksPerFrame+2], timing.ticksPerFrame)
	binary.BigEndian.PutUint32(d[cdtaTickRate:cdtaTickRate+4], timing.tickRate)
	binary.BigEndian.PutUint32(d[cdtaTickRateMirror30:cdtaTickRateMirror30+4], timing.tickRate)

	// TimeBaseDivisor @0x10 — always 600 (matches WorkArea divisor)
	// Secondary divisor @0x18 — 600 for fresh comps (实测 A_baseline + dummy_comp).
	// AE rewrites to TickRate after user mods（E_shutter/F_shutter/RE_fps_*）—
	// builder 出 fresh comp，必须用 600，否则 AE 25 把 ShutterAngle 按 NTSC 因子重算
	// (实测 Phase 6 ship gate: stored 180 → AE display 216 when @0x18=TickRate)。
	binary.BigEndian.PutUint32(d[cdtaTimeBaseDivisor:cdtaTimeBaseDivisor+4], 600)
	binary.BigEndian.PutUint32(d[cdtaSecondaryDivisor18:cdtaSecondaryDivisor18+4], 600)

	// WorkArea @0x1C..@0x2B：start=0/600, end=sentinel/600
	binary.BigEndian.PutUint32(d[cdtaWorkAreaStart:cdtaWorkAreaStart+4], 0)
	binary.BigEndian.PutUint32(d[cdtaWorkAreaStartDiv:cdtaWorkAreaStartDiv+4], 600)
	binary.BigEndian.PutUint32(d[cdtaWorkAreaEnd:cdtaWorkAreaEnd+4], 0xFFFFFFFF) // sentinel "use Duration"
	binary.BigEndian.PutUint32(d[cdtaWorkAreaEndDiv:cdtaWorkAreaEndDiv+4], 600)

	// MasterTicks @0x2C = round(duration_seconds × nominalTickRate)。
	// AE 25 display duration ≈ @0x2C / tickRate (per RE_cdta_probe A_baseline + ship gate
	// 实测：错值导致 AE 显示 错 duration + 触发其它字段误算如 ShutterAngle)。
	masterTicks := uint32(math.Round(duration * float64(timing.nominalTickRate)))
	binary.BigEndian.PutUint32(d[cdtaMasterTicks:cdtaMasterTicks+4], masterTicks)

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

	// Duration @0xB0 + mirror @0xB8 (= round(duration_seconds × fps) frames)
	durationFrames := uint32(math.Round(duration * fps))
	binary.BigEndian.PutUint32(d[cdtaDuration:cdtaDuration+4], durationFrames)
	binary.BigEndian.PutUint32(d[cdtaDurationMirror:cdtaDurationMirror+4], durationFrames)

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

// ensureTemplateCompItemChunks lazy-init template-copy chunks，即 builder 不合成
// 但 AE 期望存在的 sibling chunks (dats / cdrp / comr / PRin / DLay 等)。
// 一次性从 embedded dummy-comp 模板 parse 出 dummy comp 的 Item LIST，剔除
// builder-managed children (iide/idpc/idta/Utf8/cdta/Layr)，余下 deep-cloned
// 存到 templateCompItemChunks。
//
// 同时提取 Item LIST 之后的 8 个 sibling chunks (FEE LIST + 7 small chunks)
// 存到 templateItemSiblingChunks —— AE 25 要求每个 Item 后面都跟这一组 chunks，
// 否则报 "文件数据丢失" (实测 Phase 6 ship gate)。
func ensureTemplateCompItemChunks() {
	templateInit.Do(func() {
		p, err := FromReader(bytes.NewReader(embeddedDummyCompTemplate))
		if err != nil {
			panic(fmt.Sprintf("aep: corrupt dummy-comp template (build bug): %v", err))
		}
		if len(p.Compositions) == 0 {
			panic("aep: dummy-comp template missing comp (build bug)")
		}
		comp := p.Compositions[0]
		if comp.itemList == nil {
			panic("aep: parseComposition didn't wire itemList (build bug)")
		}
		for _, ch := range comp.itemList.Children {
			if isBuilderManagedChunk(ch) {
				continue
			}
			templateCompItemChunks = append(templateCompItemChunks, deepCloneChunk(ch))
		}

		// Walk root → Fold → find Item LIST，提取其后的 8 个 sibling chunks。
		// dummy_comp.aep Fold layout (10 children): fdta / Item / FEE / fvdv /
		// fiop / ftts / foac / fiac / fipc / fifl. 跳过 fdta + Item，取剩余 8 个。
		var foldList *rifx.Chunk
		for _, ch := range p.root.Children {
			if ch.IsList() && ch.FormType == rifx.IDFold {
				foldList = ch
				break
			}
		}
		if foldList == nil {
			panic("aep: dummy-comp template missing Fold LIST (build bug)")
		}
		seenItem := false
		for _, ch := range foldList.Children {
			if !seenItem {
				if ch.IsList() && ch.FormType == rifx.IDItem {
					seenItem = true
				}
				continue
			}
			templateItemSiblingChunks = append(templateItemSiblingChunks, deepCloneChunk(ch))
		}
		if len(templateItemSiblingChunks) == 0 {
			panic("aep: dummy-comp template missing Item sibling chunks (build bug)")
		}
	})
}

// isBuilderManagedChunk: builder 自己合成的 chunk，不从 template 复制。
// Builder-managed: iide / idpc / idta / Utf8 / cdta / LIST(Layr)
// Template-copy:   dats / cdrp / comr / LIST(PRin) / LIST(DLay) / LIST(SLay) 等
func isBuilderManagedChunk(c *rifx.Chunk) bool {
	tag := string(c.ID[:])
	switch tag {
	case "iide", "idpc", "idta", "Utf8", "cdta":
		return true
	}
	if c.IsList() && c.FormType == rifx.IDLayr {
		return true
	}
	return false
}

// deepCloneChunk: 递归深拷贝 *rifx.Chunk（防止 NewComposition 间共享 mutation）。
func deepCloneChunk(c *rifx.Chunk) *rifx.Chunk {
	clone := &rifx.Chunk{
		ID:       c.ID,
		FormType: c.FormType,
		Data:     append([]byte(nil), c.Data...),
	}
	for _, ch := range c.Children {
		clone.Children = append(clone.Children, deepCloneChunk(ch))
	}
	return clone
}

// buildCompItem 包成完整 LIST formType=Item。
// Children 顺序（按 AE 2025 saved fixture 实测，见
// test_data/comp_item_children.golden.txt）：
//
//	iide / idpc / idta / Utf8 / LIST(dats) / cdta / cdrp / comr /
//	LIST(PRin) / LIST(DLay) ... / LIST(Layr)
//
// builder-managed: iide / idpc / idta / Utf8 / cdta / Layr (6 chunks)
// template-copy:   dats / cdrp / comr / PRin / DLay etc.
//
// 关键：fixture 中 LIST(dats) 在 cdta 之前；其余 template chunks 在 cdta 之后；
// LIST(Layr) 永远是最后一个。这个顺序对 AE 重开兼容性很重要。
func buildCompItem(itemID uint32, name string, cdta []byte) *rifx.Chunk {
	ensureTemplateCompItemChunks()

	iide := buildCompIide()
	idpc := buildCompIdpc()
	idta := buildCompIdta(itemID)
	utf8 := &rifx.Chunk{
		ID:   rifx.IDUtf8,
		Data: []byte(name),
	}
	cdtaChunk := &rifx.Chunk{
		ID:   rifx.IDCdta,
		Data: cdta,
	}

	children := []*rifx.Chunk{iide, idpc, idta, utf8}

	// fixture 中 template chunks[0] = LIST(dats)，必须插在 cdta 之前。
	// 其余 template chunks 全部插到 cdta 之后。
	//
	// 不写 empty Layr：AE 自己 saved 的 empty comp 也不带 Layr（dummy template
	// 验证：只有 SLay/CLay/SecL，没有 Layr）。若写 empty Layr，parseLayer 会在
	// ldta 缺失时返回 stub 层，污染 comp.Layers（V2.2 AddLayer 时再 build Layr）。
	if len(templateCompItemChunks) > 0 && isDatsList(templateCompItemChunks[0]) {
		children = append(children, deepCloneChunk(templateCompItemChunks[0]))
		children = append(children, cdtaChunk)
		for _, tc := range templateCompItemChunks[1:] {
			children = append(children, deepCloneChunk(tc))
		}
	} else {
		// fallback: dats 不在 [0]，直接 cdta 然后所有 template chunks
		children = append(children, cdtaChunk)
		for _, tc := range templateCompItemChunks {
			children = append(children, deepCloneChunk(tc))
		}
	}

	return &rifx.Chunk{
		ID:       rifx.IDList,
		FormType: rifx.IDItem,
		Children: children,
	}
}

func isDatsList(c *rifx.Chunk) bool {
	return c.IsList() && string(c.FormType[:]) == "dats"
}

// NewComposition adds an empty composition to the project's root folder.
// 详 spec: workshop/specs/v2-1-foundation-design.md §Public API
//
// Required:
//   name        — non-empty string
//   width/height — > 0 (uint16; AE max 30000)
//   frameRate   — > 0 (Hz; 29.97 etc.; whole+frac/65536 encoding handled internally)
//   duration    — > 0 (seconds; converted to whole frames via fps internally)
//
// Optional fields default to AE-typical (BGColor=0/PAR=1.0/ResFac=1,1/Shutter=180,0/MotionBlur=128,16).
// Override via existing Set* methods after the call.
//
// Composition.ID is auto-assigned (Project.nextItemID++, monotonic).
// New comp appends to the project's root folder.
//
// Atomic mutation (Invariant #10): if chunk parse fails or warnings appear,
// rollback chunk-tree + typed index + warnings to pre-call state.
//
// Warnings-as-failure (Invariant #11): builder must produce zero parser warnings —
// if any appear, that's a builder bug; rollback + return internal error.
func (p *Project) NewComposition(
	name string,
	width, height uint16,
	frameRate, duration float64,
) (*Composition, error) {
	// 1. Validate
	if err := validateNewCompositionInputs(name, width, height, frameRate, duration); err != nil {
		return nil, err
	}

	// 2. Allocate ID (monotonic; never reuses — Invariant #9)
	id := p.allocItemID()

	// 3. Build chunks
	cdtaBytes := buildCompCdta(width, height, frameRate, duration)
	itemList := buildCompItem(id, name, cdtaBytes)

	// 4. Atomic mutation prep
	if p.rootFold == nil {
		return nil, fmt.Errorf("internal: project missing root Fold (template malformed?)")
	}
	oldChildLen := len(p.rootFold.Children)
	oldWarningsLen := len(p.Warnings)

	// 5. Append to rootFold + reparse closed loop。
	// AE 25 在 Fold 里要求每个 Item LIST 后面都跟 8 个 sibling chunks
	// (FEE LIST + fvdv/fiop/ftts/foac/fiac/fipc/fifl) —— 否则报 "文件数据丢失"
	// (实测 Phase 6 ship gate)。从 dummy_comp 模板 deep-clone。
	p.rootFold.Children = append(p.rootFold.Children, itemList)
	for _, sib := range templateItemSiblingChunks {
		p.rootFold.Children = append(p.rootFold.Children, deepCloneChunk(sib))
	}
	comp, err := parseComposition(itemList, id, name, &p.Warnings)
	if err != nil {
		// Rollback
		p.rootFold.Children = p.rootFold.Children[:oldChildLen]
		p.Warnings = p.Warnings[:oldWarningsLen]
		return nil, fmt.Errorf("internal: re-parsing new composition: %w", err)
	}

	// 6. Warnings-as-failure (Invariant #11)
	if len(p.Warnings) != oldWarningsLen {
		newWarnings := append([]string(nil), p.Warnings[oldWarningsLen:]...)
		p.rootFold.Children = p.rootFold.Children[:oldChildLen]
		p.Warnings = p.Warnings[:oldWarningsLen]
		return nil, fmt.Errorf("internal: builder produced %d parser warning(s): %v", len(newWarnings), newWarnings)
	}

	// 7. Wire back-pointer + register in typed index
	comp.proj = p
	// comp.itemList 已由 parseComposition 设置 (Phase 4 wiring)
	p.Compositions = append(p.Compositions, comp)

	return comp, nil
}

// validateNewCompositionInputs returns nil if all inputs are valid, or
// an error naming the offending field + value.
func validateNewCompositionInputs(name string, w, h uint16, fps, duration float64) error {
	if name == "" {
		return fmt.Errorf("composition name cannot be empty")
	}
	if w == 0 || h == 0 {
		return fmt.Errorf("composition size must be > 0 (got %dx%d)", w, h)
	}
	if fps <= 0 {
		return fmt.Errorf("frame rate must be > 0 (got %g)", fps)
	}
	if duration <= 0 {
		return fmt.Errorf("duration must be > 0 (got %g)", duration)
	}
	return nil
}
