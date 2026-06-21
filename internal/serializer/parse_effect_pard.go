package serializer

import (
	"encoding/binary"
	"math"

	"github.com/example/aep-parser/internal/rifx"
)

// pardParamDef holds the metadata extracted from a single pard
// (effect parameter definition) chunk. Layout depends on the
// property_control_type byte at pard @0x0F.
type pardParamDef struct {
	name        string              // raw name from pard @0x10-0x2F (32 bytes, NUL-padded)
	controlType PropertyControlType // from pard @0x0F
	lastValue   any                 // type-dependent: float64, int, []float64
	defaultVal  any                 // type-dependent
	nbOptions   int                 // enum only (pard s4 @high16)
	minValue    any                 // scalar: s2; color: float64 constant
	maxValue    any                 // scalar: s2; slider: f4; color: float64 constant
}

// parsePardParams extracts parameter definitions from the parT LIST
// inside an effect's sspc wrapper. Returns a map keyed by parameter
// match-name. Returns nil when no parT LIST is found.
//
// parT structure: parn(count) then tdmn+pard pairs. The first tdmn+pard
// pair describes the parent effect itself and is skipped. Additional
// chunks (pdnm, etc.) may appear between pairs — we skip non-tdmn chunks.
func parsePardParams(sspc *rifx.Chunk) map[string]*pardParamDef {
	if sspc == nil {
		return nil
	}
	var parT *rifx.Chunk
	for _, ch := range sspc.Children {
		if ch.IsList() && ch.FormType == rifx.IDparT {
			parT = ch
			break
		}
	}
	if parT == nil {
		return nil
	}

	result := make(map[string]*pardParamDef)
	kids := parT.Children
	idx := 0
	for i := 0; i < len(kids); i++ {
		ch := kids[i]
		if ch.ID != rifx.IDTdmn {
			continue
		}
		name := trimNUL(ch.Data)
		if i+1 >= len(kids) {
			continue
		}
		payload := kids[i+1]
		if payload.ID != rifx.IDpard {
			continue
		}
		// Skip the first tdmn+pard pair (effect header).
		idx++
		if idx == 1 {
			i++ // skip the pard too
			continue
		}
		def := parseSinglePard(name, payload)
		if def != nil {
			result[name] = def
		}
		i++ // skip the pard chunk
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// parseSinglePard extracts metadata from one pard chunk.
// pard layout: 15B pad + 1B control_type + 32B name + 8B pad = 56B prefix,
// then type-dependent payload.
func parseSinglePard(matchName string, pard *rifx.Chunk) *pardParamDef {
	d := pard.Data
	if len(d) < 56 {
		return nil
	}

	controlType := PropertyControlType(d[0x0F])

	// Extract name from 32 bytes at @0x10.
	name := matchName // fallback
	rawName := d[0x10:0x30]
	if n := nullTermIndex(rawName); n > 0 {
		name = string(rawName[:n])
	}

	def := &pardParamDef{
		name:        name,
		controlType: controlType,
	}

	// Payload starts at offset 56 (0x38).
	body := d[0x38:]

	switch controlType {
	case PCTLScalar:
		// s4 last_value + 72B pad + s2 min + 2B pad + s2 max
		if len(body) >= 82 {
			def.lastValue = float64(int32(binary.BigEndian.Uint32(body[0:4]))) / 65536.0
			def.defaultVal = def.lastValue
			def.minValue = float64(int16(binary.BigEndian.Uint16(body[76:78])))
			def.maxValue = float64(int16(binary.BigEndian.Uint16(body[80:82])))
		}

	case PCTLAngle:
		// s4 last_value
		if len(body) >= 4 {
			def.lastValue = float64(int32(binary.BigEndian.Uint32(body[0:4]))) / 65536.0
			def.defaultVal = def.lastValue
		}

	case PCTLBoolean:
		// u4 last_value + u1 default
		if len(body) >= 5 {
			def.lastValue = int(binary.BigEndian.Uint32(body[0:4]))
			def.defaultVal = int(body[4])
		}

	case PCTLTwoD:
		// s4 last_value_x + s4 last_value_y (÷128 for pixels)
		if len(body) >= 8 {
			x := float64(int32(binary.BigEndian.Uint32(body[0:4]))) / 128.0
			y := float64(int32(binary.BigEndian.Uint32(body[4:8]))) / 128.0
			def.lastValue = []float64{x, y}
			def.defaultVal = def.lastValue
		}

	case PCTLEnum:
		// u4 last_value + s4 nb_options (>>16 = count) + s4 default
		if len(body) >= 12 {
			def.lastValue = int(binary.BigEndian.Uint32(body[0:4]))
			rawNb := int32(binary.BigEndian.Uint32(body[4:8]))
			def.nbOptions = int(rawNb >> 16)
			def.defaultVal = int(int32(binary.BigEndian.Uint32(body[8:12])))
			def.minValue = 1.0
			def.maxValue = float64(def.nbOptions)
		}

	case PCTLColor:
		// pard stores colors as [A, R, G, B] in 0-255 range.
		// Look for color data — py-aep uses last_color / default_color //nolint:jargon
		// from a separate ColorPardChunk layout. For now, extract
		// what we can from the body. Color pard has a different prefix
		// layout. We'll handle it if we encounter it in fixtures.

	case PCTLSlider:
		// f8 last_value + 52B pad + f4 max_value
		if len(body) >= 64 {
			def.lastValue = math.Float64frombits(binary.BigEndian.Uint64(body[0:8]))
			def.defaultVal = def.lastValue
			def.maxValue = float64(math.Float32frombits(binary.BigEndian.Uint32(body[60:64])))
		}

	case PCTLThreeD:
		// f8 x + f8 y + f8 z (×512 for pixels)
		if len(body) >= 24 {
			x := math.Float64frombits(binary.BigEndian.Uint64(body[0:8])) * 512.0
			y := math.Float64frombits(binary.BigEndian.Uint64(body[8:16])) * 512.0
			z := math.Float64frombits(binary.BigEndian.Uint64(body[16:24])) * 512.0
			def.lastValue = []float64{x, y, z}
			def.defaultVal = def.lastValue
		}

	default:
		// LAYER, GROUP, MASK, CURVE, UNKNOWN — no typed payload.
	}

	return def
}

// nullTermIndex returns the index of the first NUL byte in b,
// or len(b) if no NUL is found.
func nullTermIndex(b []byte) int {
	for i, v := range b {
		if v == 0 {
			return i
		}
	}
	return len(b)
}
