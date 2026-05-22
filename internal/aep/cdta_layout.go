// internal/aep/cdta_layout.go
package aep

// Composition cdta byte offsets — single source of truth shared by
// parser (parse_composition.go), writer (write_composition.go), and
// builder (new_composition.go).
//
// cdta total size is 204 bytes (0xCC); layout is stable AE 2020 → AE 2025
// (bit-for-bit verified, see workshop/scars/).
const (
	cdtaResolutionFactorX  = 0x00 // uint16 BE
	cdtaResolutionFactorY  = 0x02 // uint16 BE
	cdtaTicksPerFrame      = 0x06 // uint16 BE — fps-derived (1024 for 24/25/30; 512 for 50/60; 800 for 29.97; 400 for 59.94)
	cdtaTickRate           = 0x08 // uint32 BE — ticks_per_frame × fps
	cdtaTimeBaseDivisor    = 0x10 // uint32 BE — always 600 (matches WorkArea divisor)
	cdtaTickRateMirror18   = 0x18 // uint32 BE — mirror of cdtaTickRate
	cdtaWorkAreaStart      = 0x1C // uint32 BE dividend
	cdtaWorkAreaStartDiv   = 0x20 // uint32 BE divisor
	cdtaWorkAreaEnd        = 0x24 // uint32 BE dividend; 0xFFFFFFFF = sentinel
	cdtaWorkAreaEndDiv     = 0x28 // uint32 BE divisor
	cdtaMasterTicks        = 0x2C // uint32 BE — ticks_per_frame × 5 × fps_nominal_whole
	cdtaTickRateMirror30   = 0x30 // uint32 BE — mirror of cdtaTickRate
	cdtaBGColorR           = 0x34 // uint8
	cdtaBGColorG           = 0x35 // uint8
	cdtaBGColorB           = 0x36 // uint8
	cdtaFlagsByte8A        = 0x8A // bit 0 = Draft3D
	cdtaFlagsByte8B        = 0x8B // bit 0/3/4/5/7 = various comp flags
	cdtaWidth              = 0x8C // uint16 BE
	cdtaHeight             = 0x8E // uint16 BE
	cdtaPixelAspectNum     = 0x90 // uint32 BE numerator
	cdtaPixelAspectDen     = 0x94 // uint32 BE denominator
	cdtaFrameRateWhole     = 0x9C // uint16 BE whole part
	cdtaFrameRateFrac      = 0x9E // uint16 BE fractional part (frac / 65536)
	cdtaDisplayStartTime   = 0xA4 // uint32 BE dividend
	cdtaDisplayStartDiv    = 0xA8 // uint32 BE divisor
	cdtaShutterAngle       = 0xAE // uint16 BE
	cdtaDuration           = 0xB0 // uint32 BE frames
	cdtaShutterPhase       = 0xB4 // int32 BE
	cdtaDurationMirror     = 0xB8 // uint32 BE — mirror of cdtaDuration
	cdtaMotionBlurAdaptive = 0xC4 // int32 BE
	cdtaMotionBlurSamples  = 0xC8 // int32 BE
	cdtaSize               = 0xCC // 总长 = 204
)

// Item idta byte offsets (84 bytes total).
// Full layout RE'd in workshop/plans/v2-1-foundation-plan.md RE-3 finding.
// Builder strategy: template-copy 84B + overwrite @0x10 (Item ID).
//
// NOTE: Item ID offset is @0x10 (verified by parse.classifyItem reading
// idta.U32(16) + AE2025_1comp.aep/AE2025_2comp.aep fixture dumps showing
// IDs 1 and 13 at offset 0x10..0x13). Earlier comment said @0x14 — that
// was a documentation error; the constant was declared but unused until
// now, so no behavior change.
const (
	idtaTypeCode = 0x00 // uint16 BE; 0x04 = Composition / 0x01 = Folder
	idtaItemID   = 0x10 // uint32 BE; per-item ID
	idtaLabel    = 0x3A // uint8; label color index 0..16 (0 = none)
	idtaSize     = 84
)
