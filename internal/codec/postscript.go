package codec

import (
	"encoding/binary"
	"fmt"
	"unicode/utf16"
)

// ──────────────────────────────────────────────────────────────────
// PostScript-style text-dict mini-parser
// ──────────────────────────────────────────────────────────────────

const (
	PsStr  = 1
	PsNum  = 2
	PsBool = 3
	PsName = 4
	PsDict = 5
	PsArr  = 6
)

// PsValue is a node in the parsed btdk tree. Compact untagged-union;
// only the field matching Kind is meaningful.
type PsValue struct {
	Kind int
	Num  float64
	Bv   bool
	Str  string // tName: bare name; tString: decoded text (UTF-8 if FE FF prefix, else raw bytes as Go string)
	Arr  []*PsValue
	Dict []PsDictEntry
	// SrcStart/SrcEnd locate the value's bytes inside the btdk body.
	// Tracked for every kind (scalars span their literal; dicts and
	// arrays span from `<<` / `[` through the matching close). Used by
	// length-variable PostScript splice writes (see write_text.go).
	SrcStart, SrcEnd int
}

type PsDictEntry struct {
	Key string
	Val *PsValue
}

func (v *PsValue) AsNum() float64 {
	if v == nil || v.Kind != PsNum {
		return 0
	}
	return v.Num
}

// PsPath walks `/key/key/key` (numeric keys also index arrays). Each
// segment matches either a dict key (string match) or an array index
// (parsed as int). Returns nil if any segment is missing.
func PsPath(v *PsValue, path string) *PsValue {
	if v == nil || path == "" || path == "/" {
		return v
	}
	// Skip leading '/'.
	start := 0
	if path[0] == '/' {
		start = 1
	}
	for start < len(path) {
		end := start
		for end < len(path) && path[end] != '/' {
			end++
		}
		seg := path[start:end]
		v = PsStep(v, seg)
		if v == nil {
			return nil
		}
		start = end + 1
	}
	return v
}

// PsStep takes one path segment ("0", "53", "1") and looks it up in
// dict (key match) or array (index match). Dict keys are stringified
// integers in the btdk format.
func PsStep(v *PsValue, seg string) *PsValue {
	if v.Kind == PsDict {
		for _, e := range v.Dict {
			if e.Key == seg {
				return e.Val
			}
		}
		return nil
	}
	if v.Kind == PsArr {
		var idx int
		if _, err := fmt.Sscanf(seg, "%d", &idx); err != nil {
			return nil
		}
		if idx < 0 || idx >= len(v.Arr) {
			return nil
		}
		return v.Arr[idx]
	}
	return nil
}

// PsPathStr is a shortcut that returns the string value at path or "".
func PsPathStr(v *PsValue, path string) string {
	w := PsPath(v, path)
	if w == nil || w.Kind != PsStr {
		return ""
	}
	return w.Str
}

// ExtractBtdkBody walks the raw btds payload bytes and returns the inner
// btdk LIST's body (PostScript text) plus that body's starting offset inside
// raw (so callers can translate body-local offsets back to TextSourceRaw
// offsets). The payload begins with a "LIST tdbs" sub-chunk holding the
// property descriptor; the btdk LIST follows. Both follow the regular RIFX
// LIST header layout (LIST + uint32 BE size + 4-byte formType + payload).
func ExtractBtdkBody(raw []byte) ([]byte, int, error) {
	for off := 0; off+12 <= len(raw); {
		if string(raw[off:off+4]) != "LIST" {
			return nil, 0, fmt.Errorf("expected LIST at %#x, got %q", off, raw[off:off+4])
		}
		size := binary.BigEndian.Uint32(raw[off+4 : off+8])
		if int(size) < 4 || off+8+int(size)-4 > len(raw) {
			return nil, 0, fmt.Errorf("LIST at %#x has bad size %d", off, size)
		}
		bodyStart := off + 12
		bodyEnd := off + 8 + int(size)
		if string(raw[off+8:off+12]) == "btdk" {
			return raw[bodyStart:bodyEnd], bodyStart, nil
		}
		off = bodyEnd
		// LIST chunks have implicit pad to even length.
		if size%2 != 0 {
			off++
		}
	}
	return nil, 0, fmt.Errorf("no btdk LIST in btds payload")
}

// ParsePSDict parses a top-level btdk body as an implicit dict
// (stream of /key value pairs until EOF). Returns nil if the body is
// empty.
func ParsePSDict(body []byte) *PsValue {
	l := &PsLexer{D: body}
	root := &PsValue{Kind: PsDict}
	for {
		tok := l.Next()
		if tok.Kind == PsTokEOF {
			break
		}
		if tok.Kind != PsTokName {
			continue // recovery: ignore stray tokens
		}
		key := tok.Str
		valTok := l.Next()
		v := ParsePSValue(l, valTok)
		root.Dict = append(root.Dict, PsDictEntry{Key: key, Val: v})
	}
	if len(root.Dict) == 0 {
		return nil
	}
	return root
}

// ParsePSValue consumes one value given the already-peeked first token.
// Every returned PsValue carries srcStart / srcEnd in btdk-body
// coordinates so callers can splice exact byte ranges.
func ParsePSValue(l *PsLexer, t PsToken) *PsValue {
	switch t.Kind {
	case PsTokOpenDict:
		d := &PsValue{Kind: PsDict, SrcStart: t.SrcStart}
		for {
			nt := l.Next()
			if nt.Kind == PsTokCloseDict || nt.Kind == PsTokEOF {
				d.SrcEnd = nt.SrcEnd
				if d.SrcEnd == 0 {
					d.SrcEnd = l.Pos
				}
				return d
			}
			if nt.Kind != PsTokName {
				continue
			}
			vt := l.Next()
			v := ParsePSValue(l, vt)
			d.Dict = append(d.Dict, PsDictEntry{Key: nt.Str, Val: v})
		}
	case PsTokOpenArr:
		a := &PsValue{Kind: PsArr, SrcStart: t.SrcStart}
		for {
			nt := l.Next()
			if nt.Kind == PsTokCloseArr || nt.Kind == PsTokEOF {
				a.SrcEnd = nt.SrcEnd
				if a.SrcEnd == 0 {
					a.SrcEnd = l.Pos
				}
				return a
			}
			v := ParsePSValue(l, nt)
			a.Arr = append(a.Arr, v)
		}
	case PsTokName:
		return &PsValue{Kind: PsName, Str: t.Str, SrcStart: t.SrcStart, SrcEnd: t.SrcEnd}
	case PsTokNumber:
		return &PsValue{Kind: PsNum, Num: t.Num, SrcStart: t.SrcStart, SrcEnd: t.SrcEnd}
	case PsTokBool:
		return &PsValue{Kind: PsBool, Bv: t.Bv, SrcStart: t.SrcStart, SrcEnd: t.SrcEnd}
	case PsTokString:
		return &PsValue{Kind: PsStr, Str: t.Str, SrcStart: t.SrcStart, SrcEnd: t.SrcEnd}
	}
	return &PsValue{}
}

// Token kinds for the btdk lexer.
const (
	PsTokEOF = iota
	PsTokOpenDict
	PsTokCloseDict
	PsTokOpenArr
	PsTokCloseArr
	PsTokName
	PsTokNumber
	PsTokBool
	PsTokString
)

type PsToken struct {
	Kind int
	Str  string
	Num  float64
	Bv   bool
	// srcStart/srcEnd bracket the raw token bytes inside the btdk body;
	// for string tokens (PsTokString) this includes the enclosing
	// parentheses. Used by SetText to locate the exact splice site.
	SrcStart, SrcEnd int
}

type PsLexer struct {
	D   []byte
	Pos int
}

func (l *PsLexer) SkipWS() {
	for l.Pos < len(l.D) {
		c := l.D[l.Pos]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == 0 {
			l.Pos++
			continue
		}
		break
	}
}

func isPSDelim(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\r', '[', ']', '<', '>', '/', '(', ')':
		return true
	}
	return false
}

func (l *PsLexer) Next() PsToken {
	l.SkipWS()
	if l.Pos >= len(l.D) {
		return PsToken{Kind: PsTokEOF}
	}
	tokStart := l.Pos
	c := l.D[l.Pos]
	switch {
	case c == '<' && l.Pos+1 < len(l.D) && l.D[l.Pos+1] == '<':
		l.Pos += 2
		return PsToken{Kind: PsTokOpenDict, SrcStart: tokStart, SrcEnd: l.Pos}
	case c == '>' && l.Pos+1 < len(l.D) && l.D[l.Pos+1] == '>':
		l.Pos += 2
		return PsToken{Kind: PsTokCloseDict, SrcStart: tokStart, SrcEnd: l.Pos}
	case c == '[':
		l.Pos++
		return PsToken{Kind: PsTokOpenArr, SrcStart: tokStart, SrcEnd: l.Pos}
	case c == ']':
		l.Pos++
		return PsToken{Kind: PsTokCloseArr, SrcStart: tokStart, SrcEnd: l.Pos}
	case c == '/':
		l.Pos++
		s := l.Pos
		for l.Pos < len(l.D) && !isPSDelim(l.D[l.Pos]) {
			l.Pos++
		}
		return PsToken{Kind: PsTokName, Str: string(l.D[s:l.Pos]), SrcStart: tokStart, SrcEnd: l.Pos}
	case c == '(':
		return l.ReadString()
	default:
		s := l.Pos
		for l.Pos < len(l.D) && !isPSDelim(l.D[l.Pos]) {
			l.Pos++
		}
		w := string(l.D[s:l.Pos])
		switch w {
		case "true":
			return PsToken{Kind: PsTokBool, Bv: true, SrcStart: tokStart, SrcEnd: l.Pos}
		case "false":
			return PsToken{Kind: PsTokBool, Bv: false, SrcStart: tokStart, SrcEnd: l.Pos}
		}
		var n float64
		fmt.Sscanf(w, "%g", &n)
		return PsToken{Kind: PsTokNumber, Num: n, SrcStart: tokStart, SrcEnd: l.Pos}
	}
}

// readString consumes a PostScript string `( … )` with backslash
// escapes and balanced-paren counting, then decodes UTF-16BE if the
// payload begins with the FE FF BOM.
func (l *PsLexer) ReadString() PsToken {
	start := l.Pos
	l.Pos++ // skip '('
	var buf []byte
	depth := 1
	for l.Pos < len(l.D) && depth > 0 {
		b := l.D[l.Pos]
		if b == '\\' && l.Pos+1 < len(l.D) {
			nb := l.D[l.Pos+1]
			switch nb {
			case '(':
				buf = append(buf, '(')
			case ')':
				buf = append(buf, ')')
			case '\\':
				buf = append(buf, '\\')
			case 'n':
				buf = append(buf, '\n')
			case 'r':
				buf = append(buf, '\r')
			case 't':
				buf = append(buf, '\t')
			default:
				buf = append(buf, '\\', nb)
			}
			l.Pos += 2
			continue
		}
		if b == '(' {
			depth++
			buf = append(buf, b)
			l.Pos++
			continue
		}
		if b == ')' {
			depth--
			if depth == 0 {
				l.Pos++
				break
			}
			buf = append(buf, b)
			l.Pos++
			continue
		}
		buf = append(buf, b)
		l.Pos++
	}
	return PsToken{Kind: PsTokString, Str: DecodePSString(buf), SrcStart: start, SrcEnd: l.Pos}
}

// DecodePSString turns the raw byte payload of a PostScript string
// into a Go UTF-8 string. The FE FF BOM marks UTF-16BE; otherwise
// bytes are returned as-is (Windows code-page strings are very rare
// in AE-generated btdk, mostly ASCII font tags).
func DecodePSString(b []byte) string {
	if len(b) >= 2 && b[0] == 0xfe && b[1] == 0xff {
		r := b[2:]
		units := make([]uint16, 0, len(r)/2)
		for i := 0; i+1 < len(r); i += 2 {
			units = append(units, binary.BigEndian.Uint16(r[i:i+2]))
		}
		// utf16.Decode combines surrogate pairs (astral chars round-trip;
		// rune-per-unit would mangle them to U+FFFD).
		return string(utf16.Decode(units))
	}
	return string(b)
}
