// internal/aep/ldta_layout.go
package codec

// Layer ldta byte offsets — single source of truth shared by parser
// (parse_layer.go), writer (write_layer.go), and builder (V2.2 lower_layer.go).
//
// ldta canonical total size: 160 bytes (AE 2020 / 2022). AE 2025 emits 164
// bytes with the trailing 4 bytes always zero. V2.2 builder targets the 160-B
// canonical layout; the trailing AE 2025 padding is treated as
// length-preserving passthrough and does NOT enter the capability matrix.
//
// Offsets here are sourced from V1 parse_layer.go header comment block plus
// RE dumps of a canonical empty ShapeLayer.
const (
	LdtaLayerID        = 0x00 // uint32 BE — layer-local ID (referenced by child Layer.ParentID)
	LdtaQuality        = 0x04 // uint16 BE — LayerQuality enum (0=Wireframe / 1=Draft / 2=Best)
	LdtaStretchDivd    = 0x08 // int32  BE — stretch_dividend
	LdtaStartTimeDivd  = 0x0C // int32  BE — start_time_dividend
	LdtaStartTimeDivs  = 0x10 // uint32 BE — start_time_divisor
	LdtaInPointDivd    = 0x14 // int32  BE — source-media in-point dividend
	LdtaInPointDivs    = 0x18 // uint32 BE — source-media in-point divisor
	LdtaOutPointDivd   = 0x1C // int32  BE — source-media out-point dividend
	LdtaOutPointDivs   = 0x20 // uint32 BE — source-media out-point divisor
	LdtaAttrByte0      = 0x25 // uint8  — LayerAttrBits byte 0 (sampling / frame-blend / guide)
	LdtaAttrByte1      = 0x26 // uint8  — LayerAttrBits byte 1 (null / 3D / adjustment / solo / ...)
	LdtaAttrByte2      = 0x27 // uint8  — LayerAttrBits byte 2 (lock / shy / motion-blur / visible / ...)
	LdtaSourceID       = 0x28 // uint32 BE — source item ID
	LdtaLegacyName     = 0x40 // 32 B NUL-terminated (ignored; real name in Utf8 sibling)
	LdtaLabel          = 0x3D // uint8 — label color index
	LdtaBlendingMode   = 0x63 // uint8 — BlendingMode enum
	LdtaPreserveTrans  = 0x67 // uint8 bool
	LdtaTrackMatte     = 0x6B // uint8 — TrackMatteType enum
	LdtaStretchDivs    = 0x6C // uint32 BE — stretch_divisor (paired with LdtaStretchDivd)
	LdtaLayerSubtype   = 0x80 // uint32 BE — LayerSubtype enum (0=AV / 1=Light / 2=Camera / 3=Text / 4=Shape / 5=3DModel)
	LdtaParentID       = 0x84 // uint32 BE — parent layer ID (0 = no parent)
	LdtaLightKind      = 0x88 // uint32 BE — Light layers only: LightKind enum
	LdtaCameraFilmSize = 0x98 // f64    BE — Camera layers only: 102.0472 = 36mm @ 72PPI (RO; see scar)
	LdtaTrackMatteSrc  = 0xA0 // uint32 BE — explicit track-matte source layer ID (AE 23+; 0 = legacy implicit)
)

// ldta total size per AE version.
const (
	LdtaSize2020 = 160 // AE 2020 / 2022 canonical (V2.2 builder output)
	LdtaSize2025 = 164 // AE 2025 — trailing 4 B always zero (length-preserving passthrough)
)
