package aep

import (
	"encoding/binary"
	"fmt"
)

// ──────────────────────────────────────────────────────────────────
// PostScript-style text-dict mini-parser
// ──────────────────────────────────────────────────────────────────

const (
	psStr  = 1
	psNum  = 2
	psBool = 3
	psName = 4
	psDict = 5
	psArr  = 6
)

// psValue is a node in the parsed btdk tree. Compact untagged-union;
// only the field matching kind is meaningful.
type psValue struct {
	kind int
	num  float64
	bv   bool
	str  string // tName: bare name; tString: decoded text (UTF-8 if FE FF prefix, else raw bytes as Go string)
	arr  []*psValue
	dict []psDictEntry
	// srcStart/srcEnd locate the value's bytes inside the btdk body.
	// Tracked for every kind (scalars span their literal; dicts and
	// arrays span from `<<` / `[` through the matching close). Used by
	// length-variable PostScript splice writes (see write_text.go).
	srcStart, srcEnd int
}

type psDictEntry struct {
	key string
	val *psValue
}

func (v *psValue) asNum() float64 {
	if v == nil || v.kind != psNum {
		return 0
	}
	return v.num
}

// psPath walks `/key/key/key` (numeric keys also index arrays). Each
// segment matches either a dict key (string match) or an array index
// (parsed as int). Returns nil if any segment is missing.
func psPath(v *psValue, path string) *psValue {
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
		v = psStep(v, seg)
		if v == nil {
			return nil
		}
		start = end + 1
	}
	return v
}

// psStep takes one path segment ("0", "53", "1") and looks it up in
// dict (key match) or array (index match). Dict keys are stringified
// integers in the btdk format.
func psStep(v *psValue, seg string) *psValue {
	if v.kind == psDict {
		for _, e := range v.dict {
			if e.key == seg {
				return e.val
			}
		}
		return nil
	}
	if v.kind == psArr {
		var idx int
		if _, err := fmt.Sscanf(seg, "%d", &idx); err != nil {
			return nil
		}
		if idx < 0 || idx >= len(v.arr) {
			return nil
		}
		return v.arr[idx]
	}
	return nil
}

// psPathStr is a shortcut that returns the string value at path or "".
func psPathStr(v *psValue, path string) string {
	w := psPath(v, path)
	if w == nil || w.kind != psStr {
		return ""
	}
	return w.str
}

// parsePSDict parses a top-level btdk body as an implicit dict
// (stream of /key value pairs until EOF). Returns nil if the body is
// empty.
func parsePSDict(body []byte) *psValue {
	l := &psLexer{d: body}
	root := &psValue{kind: psDict}
	for {
		tok := l.next()
		if tok.kind == psTokEOF {
			break
		}
		if tok.kind != psTokName {
			continue // recovery: ignore stray tokens
		}
		key := tok.str
		valTok := l.next()
		v := parsePSValue(l, valTok)
		root.dict = append(root.dict, psDictEntry{key: key, val: v})
	}
	if len(root.dict) == 0 {
		return nil
	}
	return root
}

// parsePSValue consumes one value given the already-peeked first token.
// Every returned psValue carries srcStart / srcEnd in btdk-body
// coordinates so callers can splice exact byte ranges.
func parsePSValue(l *psLexer, t psToken) *psValue {
	switch t.kind {
	case psTokOpenDict:
		d := &psValue{kind: psDict, srcStart: t.srcStart}
		for {
			nt := l.next()
			if nt.kind == psTokCloseDict || nt.kind == psTokEOF {
				d.srcEnd = nt.srcEnd
				if d.srcEnd == 0 {
					d.srcEnd = l.pos
				}
				return d
			}
			if nt.kind != psTokName {
				continue
			}
			vt := l.next()
			v := parsePSValue(l, vt)
			d.dict = append(d.dict, psDictEntry{key: nt.str, val: v})
		}
	case psTokOpenArr:
		a := &psValue{kind: psArr, srcStart: t.srcStart}
		for {
			nt := l.next()
			if nt.kind == psTokCloseArr || nt.kind == psTokEOF {
				a.srcEnd = nt.srcEnd
				if a.srcEnd == 0 {
					a.srcEnd = l.pos
				}
				return a
			}
			v := parsePSValue(l, nt)
			a.arr = append(a.arr, v)
		}
	case psTokName:
		return &psValue{kind: psName, str: t.str, srcStart: t.srcStart, srcEnd: t.srcEnd}
	case psTokNumber:
		return &psValue{kind: psNum, num: t.num, srcStart: t.srcStart, srcEnd: t.srcEnd}
	case psTokBool:
		return &psValue{kind: psBool, bv: t.bv, srcStart: t.srcStart, srcEnd: t.srcEnd}
	case psTokString:
		return &psValue{kind: psStr, str: t.str, srcStart: t.srcStart, srcEnd: t.srcEnd}
	}
	return &psValue{}
}

// Token kinds for the btdk lexer.
const (
	psTokEOF = iota
	psTokOpenDict
	psTokCloseDict
	psTokOpenArr
	psTokCloseArr
	psTokName
	psTokNumber
	psTokBool
	psTokString
)

type psToken struct {
	kind int
	str  string
	num  float64
	bv   bool
	// srcStart/srcEnd bracket the raw token bytes inside the btdk body;
	// for string tokens (psTokString) this includes the enclosing
	// parentheses. Used by SetText to locate the exact splice site.
	srcStart, srcEnd int
}

type psLexer struct {
	d   []byte
	pos int
}

func (l *psLexer) skipWS() {
	for l.pos < len(l.d) {
		c := l.d[l.pos]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == 0 {
			l.pos++
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

func (l *psLexer) next() psToken {
	l.skipWS()
	if l.pos >= len(l.d) {
		return psToken{kind: psTokEOF}
	}
	tokStart := l.pos
	c := l.d[l.pos]
	switch {
	case c == '<' && l.pos+1 < len(l.d) && l.d[l.pos+1] == '<':
		l.pos += 2
		return psToken{kind: psTokOpenDict, srcStart: tokStart, srcEnd: l.pos}
	case c == '>' && l.pos+1 < len(l.d) && l.d[l.pos+1] == '>':
		l.pos += 2
		return psToken{kind: psTokCloseDict, srcStart: tokStart, srcEnd: l.pos}
	case c == '[':
		l.pos++
		return psToken{kind: psTokOpenArr, srcStart: tokStart, srcEnd: l.pos}
	case c == ']':
		l.pos++
		return psToken{kind: psTokCloseArr, srcStart: tokStart, srcEnd: l.pos}
	case c == '/':
		l.pos++
		s := l.pos
		for l.pos < len(l.d) && !isPSDelim(l.d[l.pos]) {
			l.pos++
		}
		return psToken{kind: psTokName, str: string(l.d[s:l.pos]), srcStart: tokStart, srcEnd: l.pos}
	case c == '(':
		return l.readString()
	default:
		s := l.pos
		for l.pos < len(l.d) && !isPSDelim(l.d[l.pos]) {
			l.pos++
		}
		w := string(l.d[s:l.pos])
		switch w {
		case "true":
			return psToken{kind: psTokBool, bv: true, srcStart: tokStart, srcEnd: l.pos}
		case "false":
			return psToken{kind: psTokBool, bv: false, srcStart: tokStart, srcEnd: l.pos}
		}
		var n float64
		fmt.Sscanf(w, "%g", &n)
		return psToken{kind: psTokNumber, num: n, srcStart: tokStart, srcEnd: l.pos}
	}
}

// readString consumes a PostScript string `( … )` with backslash
// escapes and balanced-paren counting, then decodes UTF-16BE if the
// payload begins with the FE FF BOM.
func (l *psLexer) readString() psToken {
	start := l.pos
	l.pos++ // skip '('
	var buf []byte
	depth := 1
	for l.pos < len(l.d) && depth > 0 {
		b := l.d[l.pos]
		if b == '\\' && l.pos+1 < len(l.d) {
			nb := l.d[l.pos+1]
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
			l.pos += 2
			continue
		}
		if b == '(' {
			depth++
			buf = append(buf, b)
			l.pos++
			continue
		}
		if b == ')' {
			depth--
			if depth == 0 {
				l.pos++
				break
			}
			buf = append(buf, b)
			l.pos++
			continue
		}
		buf = append(buf, b)
		l.pos++
	}
	return psToken{kind: psTokString, str: decodePSString(buf), srcStart: start, srcEnd: l.pos}
}

// decodePSString turns the raw byte payload of a PostScript string
// into a Go UTF-8 string. The FE FF BOM marks UTF-16BE; otherwise
// bytes are returned as-is (Windows code-page strings are very rare
// in AE-generated btdk, mostly ASCII font tags).
func decodePSString(b []byte) string {
	if len(b) >= 2 && b[0] == 0xfe && b[1] == 0xff {
		r := b[2:]
		var out []rune
		for i := 0; i+1 < len(r); i += 2 {
			cu := binary.BigEndian.Uint16(r[i : i+2])
			out = append(out, rune(cu))
		}
		return string(out)
	}
	return string(b)
}
