package aep

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"strings"

	"github.com/example/aep-parser/internal/rifx"
)

// chunkIDHead 是 root-level "head" chunk 的 ChunkID。包含 project 级 counter
// (max item ID + a save-sequence counter)，AE 25 打开时校验 —— 若 counter
// 低于实际 item IDs，AE 报 "文件数据丢失" (实测 Phase 6 ship gate, 2026-05-22)。
var chunkIDHead = rifx.ChunkID{'h', 'e', 'a', 'd'}

// WriteAEP serializes the (possibly mutated) project back to RIFX binary
// form. Sizes are recomputed from the current chunk data, so mutations
// such as Footage.SetPath that change byte lengths are handled correctly.
//
// This is best-effort write-back. The library only understands a small
// subset of the .aep format; chunks we don't know about pass through
// byte-for-byte. If After Effects rejects the output, file a sample.
func (p *Project) WriteAEP(w io.Writer) error {
	if p.root == nil {
		return fmt.Errorf("aep: project has no underlying RIFX tree (was it built from FromReader?)")
	}
	p.syncHeadCounters()
	return p.root.Write(w)
}

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
func (p *Project) syncHeadCounters() {
	if p.root == nil {
		return
	}
	head := p.root.FindFirst(chunkIDHead)
	if head == nil || len(head.Data) < 20 {
		return
	}
	curA := binary.BigEndian.Uint32(head.Data[12:16])
	curB := binary.BigEndian.Uint32(head.Data[16:20])
	if p.nextItemID > curA {
		binary.BigEndian.PutUint32(head.Data[12:16], p.nextItemID)
	}
	if p.nextItemID > curB {
		binary.BigEndian.PutUint32(head.Data[16:20], p.nextItemID)
	}
}

// SetPath updates the footage's source path. The change is propagated to
// the underlying RIFX chunks (the alas JSON's "fullpath" field is rewritten
// in-place; a legacy Cpth chunk, if any, is fully replaced). The next call
// to Project.WriteAEP will serialize the new path.
//
// Returns an error if no writable path chunk exists for this footage
// (e.g. solids and placeholders never had one).
func (f *Footage) SetPath(newPath string) error {
	if f.aliasChunk == nil && f.cpthChunk == nil {
		return fmt.Errorf("footage %d (%q): no path chunks present (solid/placeholder?)", f.ID, f.Name)
	}
	if f.aliasChunk != nil {
		newData, err := replaceJSONStringField(f.aliasChunk.Data, "fullpath", newPath)
		if err != nil {
			return fmt.Errorf("footage %d (%q): rewrite alas fullpath: %w", f.ID, f.Name, err)
		}
		f.aliasChunk.Data = newData
	}
	if f.cpthChunk != nil {
		// Cpth is NUL-terminated UTF-8 text.
		buf := make([]byte, len(newPath)+1)
		copy(buf, newPath)
		f.cpthChunk.Data = buf
	}
	f.Path = newPath
	if base := filepath.Base(strings.ReplaceAll(newPath, `\`, `/`)); base != "" && base != "." {
		f.Name = base
	}
	return nil
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


// SetBitsPerChannel writes the project's color depth (8 / 16 / 32 bpc)
// to BOTH the nhed @0x0F and nnhd @0x18 header bytes. AE stores the
// enum redundantly; we keep both in sync.
//
// Accepts the existing `BPC8` / `BPC16` / `BPC32` constants. Other
// values are written verbatim (in case AE introduces e.g. half-float
// later) but produce a less obvious AE UI state.
//
// length-preserving (2 bytes total).
func (p *Project) SetBitsPerChannel(bpc BitsPerChannel) error {
	if p.nhedChunk == nil || p.nnhdChunk == nil {
		return fmt.Errorf("project: header chunks missing (built outside parser?)")
	}
	if len(p.nhedChunk.Data) <= 0x0F {
		return fmt.Errorf("project: nhed too short (len=%d) for BitsPerChannel write", len(p.nhedChunk.Data))
	}
	if len(p.nnhdChunk.Data) <= 0x18 {
		return fmt.Errorf("project: nnhd too short (len=%d) for BitsPerChannel write", len(p.nnhdChunk.Data))
	}
	p.nhedChunk.Data[0x0F] = byte(bpc)
	p.nnhdChunk.Data[0x18] = byte(bpc)
	p.BitsPerChannel = bpc
	return nil
}
