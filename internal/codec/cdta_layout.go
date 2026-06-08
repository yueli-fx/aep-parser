// internal/aep/cdta_layout.go
package codec

// Composition cdta byte offsets — single source of truth shared by
// parser (parse_composition.go), writer (write_composition.go), and
// builder (new_composition.go).
//
// cdta total size is 204 bytes (0xCC); layout is stable AE 2020 → AE 2025
// (bit-for-bit verified).
const (
	CdtaResolutionFactorX  = 0x00 // uint16 BE
	CdtaResolutionFactorY  = 0x02 // uint16 BE
	CdtaTicksPerFrame      = 0x06 // uint16 BE — fps-derived (1024 for 24/25/30; 512 for 50/60; 800 for 29.97; 400 for 59.94)
	CdtaTickRate           = 0x08 // uint32 BE — ticks_per_frame × fps
	CdtaTimeBaseDivisor    = 0x10 // uint32 BE — always 600 (matches WorkArea divisor)
	CdtaSecondaryDivisor18 = 0x18 // uint32 BE — 600 for fresh comps; AE rewrites to TickRate on user mod

	CdtaWorkAreaStart      = 0x1C // uint32 BE dividend
	CdtaWorkAreaStartDiv   = 0x20 // uint32 BE divisor
	CdtaWorkAreaEnd        = 0x24 // uint32 BE dividend; 0xFFFFFFFF = sentinel
	CdtaWorkAreaEndDiv     = 0x28 // uint32 BE divisor
	CdtaMasterTicks        = 0x2C // uint32 BE — ticks_per_frame × 5 × fps_nominal_whole
	CdtaTickRateMirror30   = 0x30 // uint32 BE — mirror of CdtaTickRate
	CdtaBGColorR           = 0x34 // uint8
	CdtaBGColorG           = 0x35 // uint8
	CdtaBGColorB           = 0x36 // uint8
	CdtaFlagsByte8A        = 0x8A // bit 0 = Draft3D
	CdtaFlagsByte8B        = 0x8B // bit 0/3/4/5/7 = various comp flags
	CdtaWidth              = 0x8C // uint16 BE
	CdtaHeight             = 0x8E // uint16 BE
	CdtaPixelAspectNum     = 0x90 // uint32 BE numerator
	CdtaPixelAspectDen     = 0x94 // uint32 BE denominator
	CdtaFrameRateWhole     = 0x9C // uint16 BE whole part
	CdtaFrameRateFrac      = 0x9E // uint16 BE fractional part (frac / 65536)
	CdtaDisplayStartTime   = 0xA4 // uint32 BE dividend
	CdtaDisplayStartDiv    = 0xA8 // uint32 BE divisor
	CdtaShutterAngle       = 0xAE // uint16 BE
	CdtaDuration           = 0xB0 // uint32 BE frames
	CdtaShutterPhase       = 0xB4 // int32 BE
	CdtaDurationMirror     = 0xB8 // uint32 BE — mirror of CdtaDuration
	CdtaMotionBlurAdaptive = 0xC4 // int32 BE
	CdtaMotionBlurSamples  = 0xC8 // int32 BE
	CdtaSize               = 0xCC // 总长 = 204
)

// Item idta byte offsets (84 bytes total).
// Builder strategy: template-copy 84B + overwrite @0x10 (Item ID).
//
// NOTE: Item ID offset is @0x10 (verified by parse.classifyItem reading
// idta.U32(16) + AE2025_1comp.aep/AE2025_2comp.aep fixture dumps showing
// IDs 1 and 13 at offset 0x10..0x13). Earlier comment said @0x14 — that
// was a documentation error; the constant was declared but unused until
// now, so no behavior change.
const (
	IdtaTypeCode = 0x00 // uint16 BE; 0x04 = Composition / 0x01 = Folder
	IdtaItemID   = 0x10 // uint32 BE; per-item ID
	IdtaLabel    = 0x3A // uint8; label color index 0..16 (0 = none)
	IdtaSize     = 84
)
