package serializer

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/example/aep-parser/internal/scene"
)

// syncHeadCounters 把 root head chunk 里的两个 32-bit counter 同步到至少
// nextItemID。AE 25 用这两个 counter 校验文件完整性 (item ID upper bound);
// counter < 实际 item ID 触发 "文件数据丢失" 错误。
//
// 行为: counter = max(currentValue, nextItemID)。parse-then-write 路径不
// 破坏既有文件 (existing counter ≥ nextItemID)；NewProject+NewComposition
// 路径覆盖模板默认 1 → 实际下一个可用 ID。
//
// head chunk layout (20 bytes total):
//
//	[0..3]   flags/version (opaque)
//	[4..7]   svap mirror (file UUID prefix)
//	[8..11]  constant 0x80000000
//	[12..15] counter A (uint32 BE; ~next item ID)
//	[16..19] counter B (uint32 BE; ~save-sequence; pattern not fully RE'd
//	         but ≥ nextItemID empirically opens in AE 25)
//
// Free function (serializer stage): it reaches the concrete project back-ref.
func syncHeadCounters(p *Project) {
	pb := projectBack(p)
	if pb == nil || pb.root == nil {
		return
	}
	head := pb.root.FindFirst(chunkIDHead)
	if head == nil || len(head.Data) < 20 {
		return
	}
	nextItemID := scene.ProjectNextItemID(p)
	curA := binary.BigEndian.Uint32(head.Data[12:16])
	curB := binary.BigEndian.Uint32(head.Data[16:20])
	if nextItemID > curA {
		binary.BigEndian.PutUint32(head.Data[12:16], nextItemID)
	}
	if nextItemID > curB {
		binary.BigEndian.PutUint32(head.Data[16:20], nextItemID)
	}
}

// replaceJSONStringField does a byte-level surgical replacement of a single
// "key":"value" string field's value, preserving the rest of the JSON
// byte-for-byte. We avoid Unmarshal/Marshal to preserve original field
// order and whitespace, which AE's parser may depend on.
func replaceJSONStringField(data []byte, key, newValue string) ([]byte, error) {
	needle := []byte(`"` + key + `":"`)
	start := indexBytes(data, needle)
	if start < 0 {
		return nil, fmt.Errorf("field %q not found", key)
	}
	valueStart := start + len(needle)
	// Find matching closing quote, respecting backslash escapes.
	end := valueStart
	for end < len(data) {
		if data[end] == '\\' && end+1 < len(data) {
			end += 2
			continue
		}
		if data[end] == '"' {
			break
		}
		end++
	}
	if end >= len(data) {
		return nil, fmt.Errorf("field %q: unterminated string", key)
	}
	encoded := jsonEscapeString(newValue)
	out := make([]byte, 0, len(data)-(end-valueStart)+len(encoded))
	out = append(out, data[:valueStart]...)
	out = append(out, encoded...)
	out = append(out, data[end:]...)
	return out, nil
}

func indexBytes(s, sub []byte) int {
	if len(sub) == 0 {
		return 0
	}
outer:
	for i := 0; i+len(sub) <= len(s); i++ {
		for j := range sub {
			if s[i+j] != sub[j] {
				continue outer
			}
		}
		return i
	}
	return -1
}

func writeFloat64(d []byte, offset int, f float64) error {
	if offset < 0 || offset+8 > len(d) {
		return fmt.Errorf("float64 write at offset %d: out of bounds (len=%d)", offset, len(d))
	}
	binary.BigEndian.PutUint64(d[offset:offset+8], math.Float64bits(f))
	return nil
}

// jsonEscapeString returns the JSON-string body for s (no surrounding quotes).
func jsonEscapeString(s string) []byte {
	var b []byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '"':
			b = append(b, '\\', '"')
		case '\\':
			b = append(b, '\\', '\\')
		case '\n':
			b = append(b, '\\', 'n')
		case '\r':
			b = append(b, '\\', 'r')
		case '\t':
			b = append(b, '\\', 't')
		case '\b':
			b = append(b, '\\', 'b')
		case '\f':
			b = append(b, '\\', 'f')
		default:
			if c < 0x20 {
				b = append(b, []byte(fmt.Sprintf(`\u%04x`, c))...)
			} else {
				b = append(b, c)
			}
		}
	}
	return b
}
