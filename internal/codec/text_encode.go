package codec

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// LightSourceUndefined is the ldta @0x28 sentinel meaning "no explicit light
// source layer" for the SetLightSource path.
const LightSourceUndefined uint32 = 0xFFFFFFFF

// RunStylePath builds the PostScript path to the style-run dict at the given
// run index: /1/1/0/0/6/0/{runIdx}/0/0/6.
func RunStylePath(runIdx int) string {
	return fmt.Sprintf("/1/1/0/0/6/0/%d/0/0/6", runIdx)
}

// ParagraphStylePath builds the PostScript path to the paragraph style dict at
// the given paragraph index: /1/1/0/0/5/0/{paraIdx}/0/0/5.
func ParagraphStylePath(paraIdx int) string {
	return fmt.Sprintf("/1/1/0/0/5/0/%d/0/0/5", paraIdx)
}

// FormatPSNumber writes a float64 in the same shape AE uses: trailing zeros
// trimmed but the integer form kept when fractional == 0 (88 stays "88",
// 1.5 stays "1.5").
func FormatPSNumber(v float64) string {
	if v == float64(int64(v)) {
		return strconv.FormatInt(int64(v), 10)
	}
	return strconv.FormatFloat(v, 'g', -1, 64)
}

// FormatPSColorArray formats a [R, G, B, A] color (each 0..1) as the btdk
// source order [A, R, G, B] inside square brackets.
func FormatPSColorArray(rgba [4]float64) []byte {
	var b bytes.Buffer
	b.WriteByte('[')
	b.WriteByte(' ')
	for _, v := range [4]float64{rgba[3], rgba[0], rgba[1], rgba[2]} {
		b.WriteString(FormatPSNumber(v))
		b.WriteByte(' ')
	}
	b.WriteByte(']')
	return b.Bytes()
}

// Byte/value encoders for AE text + comment payloads. These are pure
// string→bytes transforms (no chunk-tree access), shared by the scene Set*
// wrappers and the serializer back-ref impls, so they live in codec.

// EncodeAEPSText encodes a Go string as the byte sequence AE writes for the
// text-document string field in btdk: open paren, FE FF BOM, UTF-16BE code
// units (with surrogate-pair expansion), '\r' at end of every paragraph (input
// '\n' becomes '\r'; a trailing '\r' is added if absent), then close paren.
// Byte values 0x28/0x29/0x5C anywhere in the resulting stream (high or low byte
// of any code unit) are backslash-escaped per PostScript string rules.
func EncodeAEPSText(s string) []byte {
	s = strings.ReplaceAll(s, "\n", "\r")
	if !strings.HasSuffix(s, "\r") {
		s += "\r"
	}
	var buf bytes.Buffer
	buf.WriteByte('(')
	buf.WriteByte(0xfe)
	buf.WriteByte(0xff)
	for _, r := range s {
		writeUTF16BEEscaped(&buf, r)
	}
	buf.WriteByte(')')
	return buf.Bytes()
}

// EncodeAEPSStringNoCR is EncodeAEPSText without the auto-appended trailing
// \r — used for non-paragraph strings like font names that should be encoded
// literally.
func EncodeAEPSStringNoCR(s string) []byte {
	var buf bytes.Buffer
	buf.WriteByte('(')
	buf.WriteByte(0xfe)
	buf.WriteByte(0xff)
	for _, r := range s {
		writeUTF16BEEscaped(&buf, r)
	}
	buf.WriteByte(')')
	return buf.Bytes()
}

// SerializeFontEntry renders a CoolTypeFont entry in the shape AE writes for
// non-default fonts:
//
//	<< /0 << /99 /CoolTypeFont /0 << /0 (FE FF utf16be name) /2 0 >> >> >>
func SerializeFontEntry(name string) []byte {
	encoded := EncodeAEPSStringNoCR(name)
	var b bytes.Buffer
	b.WriteString("<< /0 << /99 /CoolTypeFont /0 << /0 ")
	b.Write(encoded)
	b.WriteString(" /2 0 >> >> >>")
	return b.Bytes()
}

// EncodeCmta encodes a Go-friendly LF-separated string into AE's cmta payload
// format: CRLF line endings + single NUL terminator.
func EncodeCmta(s string) []byte {
	if s == "" {
		// AE typically writes a single NUL even for the empty case; matches
		// what decodeCmta gracefully reads back as "".
		return []byte{0}
	}
	out := normalizeToCRLF(s) + "\x00"
	return []byte(out)
}

// writeUTF16BEEscaped emits one rune as 2 or 4 UTF-16BE bytes, escaping any
// byte that is a PostScript string special (\, (, )).
func writeUTF16BEEscaped(buf *bytes.Buffer, r rune) {
	if r > 0xffff {
		v := uint32(r) - 0x10000
		hi := uint16(0xD800 | (v >> 10))
		lo := uint16(0xDC00 | (v & 0x3FF))
		writeBEEscaped(buf, hi)
		writeBEEscaped(buf, lo)
		return
	}
	writeBEEscaped(buf, uint16(r))
}

func writeBEEscaped(buf *bytes.Buffer, cu uint16) {
	for _, b := range [2]byte{byte(cu >> 8), byte(cu)} {
		switch b {
		case '\\', '(', ')':
			buf.WriteByte('\\')
		}
		buf.WriteByte(b)
	}
}

// normalizeToCRLF converts every \n that isn't preceded by \r into \r\n.
// \r\n in the input is preserved as-is.
func normalizeToCRLF(s string) string {
	if !containsLoneLF(s) {
		return s
	}
	out := make([]byte, 0, len(s)+8)
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' && (i == 0 || s[i-1] != '\r') {
			out = append(out, '\r', '\n')
			continue
		}
		out = append(out, s[i])
	}
	return string(out)
}

func containsLoneLF(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' && (i == 0 || s[i-1] != '\r') {
			return true
		}
	}
	return false
}
