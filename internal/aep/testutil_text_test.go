package aep_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

// psEncodeUTF16BE returns a PostScript-string-quoted byte sequence for
// the given Go string, prefixed with the FE FF BOM that flags UTF-16BE
// to the btdk decoder. Escapes the three PostScript string specials.
func psEncodeUTF16BE(s string) string {
	var buf bytes.Buffer
	buf.WriteByte(0xfe)
	buf.WriteByte(0xff)
	for _, r := range s {
		// BMP only — tests don't exercise surrogate-pair text.
		buf.WriteByte(byte(r >> 8))
		buf.WriteByte(byte(r))
	}
	raw := buf.String()
	raw = strings.ReplaceAll(raw, "\\", "\\\\")
	raw = strings.ReplaceAll(raw, "(", "\\(")
	raw = strings.ReplaceAll(raw, ")", "\\)")
	return "(" + raw + ")"
}

// buildBtdkPSText synthesizes a minimal CoolType text-engine document
// containing the keys our decoder walks: one font, one paragraph with a
// justification, and one style run with font index / size / fill color.
// AE always appends a trailing CR to each paragraph, so the helper does
// too — that's what decodeTextSource's normalizeTextLines strips.
func buildBtdkPSText(text, fontName string, fontSize float64, just int, fillRGBA [4]float64) []byte {
	body := fmt.Sprintf(
		`/0 << /1 << /0 [ << /0 << /99 /CoolTypeFont /0 << /0 %s >> >> >> ] >> >> `+
			`/1 << /1 [ << /0 << /0 %s `+
			`/5 << /0 [ << /0 << /0 << /5 << /0 %d >> >> >> >> ] >> `+
			`/6 << /0 [ << /0 << /0 << /6 << /0 0 /1 %g `+
			`/53 << /99 /SimplePaint /0 << /0 1 /1 [ %g %g %g %g ] >> >> >> >> >> >> ] >> `+
			`>> >> ] >> >>`,
		psEncodeUTF16BE(fontName),
		psEncodeUTF16BE(text+"\r"),
		just,
		fontSize,
		// AE source order is [A, R, G, B]; tests pass [R, G, B, A] so flip.
		fillRGBA[3], fillRGBA[0], fillRGBA[1], fillRGBA[2],
	)
	return []byte(body)
}

// buildBtdsWrapped wraps a btdk PostScript body into the LIST-inside-LIST
// structure that decodeTextSource expects (LIST tdbs + LIST btdk). The
// tdbs sub-LIST is intentionally minimal (zero-length body) — the decoder
// skips past it to find btdk.
func buildBtdsWrapped(btdkBody []byte) []byte {
	var out bytes.Buffer
	// LIST tdbs with empty body (size = 4 = the formType only)
	out.WriteString("LIST")
	_ = binary.Write(&out, binary.BigEndian, uint32(4))
	out.WriteString("tdbs")
	// LIST btdk with the PostScript body
	out.WriteString("LIST")
	_ = binary.Write(&out, binary.BigEndian, uint32(4+len(btdkBody)))
	out.WriteString("btdk")
	out.Write(btdkBody)
	return out.Bytes()
}
