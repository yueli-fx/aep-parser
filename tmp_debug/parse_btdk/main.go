// Parse btdk payload as Adobe TextEngine PostScript-style dict and
// pretty-print the tree with paths highlighted. Used to map out which
// dict path holds the user text / font / size / color fields.
//
// Usage: go run ./tmp_debug/parse_btdk file.aep layer_name_filter
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"

	aep "github.com/example/aep-parser/internal/aep"
)

// Token kinds.
const (
	tOpenDict  = iota // <<
	tCloseDict        // >>
	tOpenArr          // [
	tCloseArr         // ]
	tName             // /foo
	tNumber           // 123 or -1.5
	tBool             // true/false
	tString           // (...) — UTF-16BE if prefixed with FE FF
	tEOF
)

type token struct {
	kind  int
	str   string  // for tName / tString
	num   float64 // for tNumber
	bv    bool    // for tBool
	bytes []byte  // raw string bytes for tString
}

type lexer struct {
	d   []byte
	pos int
}

func (l *lexer) skipWS() {
	for l.pos < len(l.d) {
		c := l.d[l.pos]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == 0 {
			l.pos++
		} else {
			break
		}
	}
}

func (l *lexer) next() token {
	l.skipWS()
	if l.pos >= len(l.d) {
		return token{kind: tEOF}
	}
	c := l.d[l.pos]
	switch {
	case c == '<' && l.pos+1 < len(l.d) && l.d[l.pos+1] == '<':
		l.pos += 2
		return token{kind: tOpenDict}
	case c == '>' && l.pos+1 < len(l.d) && l.d[l.pos+1] == '>':
		l.pos += 2
		return token{kind: tCloseDict}
	case c == '[':
		l.pos++
		return token{kind: tOpenArr}
	case c == ']':
		l.pos++
		return token{kind: tCloseArr}
	case c == '/':
		l.pos++
		s := l.pos
		for l.pos < len(l.d) {
			b := l.d[l.pos]
			if b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '[' || b == ']' || b == '<' || b == '>' || b == '/' || b == '(' {
				break
			}
			l.pos++
		}
		return token{kind: tName, str: string(l.d[s:l.pos])}
	case c == '(':
		l.pos++
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
		return token{kind: tString, bytes: buf, str: decodeStr(buf)}
	default:
		// number or bool
		s := l.pos
		for l.pos < len(l.d) {
			b := l.d[l.pos]
			if b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '[' || b == ']' || b == '<' || b == '>' || b == '/' || b == '(' {
				break
			}
			l.pos++
		}
		w := string(l.d[s:l.pos])
		switch w {
		case "true":
			return token{kind: tBool, bv: true}
		case "false":
			return token{kind: tBool, bv: false}
		}
		var n float64
		fmt.Sscanf(w, "%g", &n)
		return token{kind: tNumber, num: n}
	}
}

// decodeStr renders a string for display: if FE FF prefix, decode UTF-16BE;
// otherwise quote raw.
func decodeStr(b []byte) string {
	if len(b) >= 2 && b[0] == 0xfe && b[1] == 0xff {
		// UTF-16BE
		r := b[2:]
		var out []rune
		for i := 0; i+1 < len(r); i += 2 {
			cu := binary.BigEndian.Uint16(r[i : i+2])
			out = append(out, rune(cu))
		}
		return fmt.Sprintf("%q (utf16be)", string(out))
	}
	return fmt.Sprintf("%q", string(b))
}

// Value: dict, array, name, number, bool, string.
type value struct {
	kind int
	num  float64
	bv   bool
	str  string
	raw  []byte
	dict []dictEntry
	arr  []value
}

type dictEntry struct {
	key string // unprefixed
	val value
}

func parseValue(l *lexer, lookahead *token) value {
	t := *lookahead
	switch t.kind {
	case tOpenDict:
		d := value{kind: tOpenDict}
		for {
			nt := l.next()
			if nt.kind == tCloseDict || nt.kind == tEOF {
				return d
			}
			if nt.kind != tName {
				// recovery: try to skip
				continue
			}
			vt := l.next()
			v := parseValue(l, &vt)
			d.dict = append(d.dict, dictEntry{key: nt.str, val: v})
		}
	case tOpenArr:
		a := value{kind: tOpenArr}
		for {
			nt := l.next()
			if nt.kind == tCloseArr || nt.kind == tEOF {
				return a
			}
			v := parseValue(l, &nt)
			a.arr = append(a.arr, v)
		}
	case tName:
		return value{kind: tName, str: t.str}
	case tNumber:
		return value{kind: tNumber, num: t.num}
	case tBool:
		return value{kind: tBool, bv: t.bv}
	case tString:
		return value{kind: tString, str: t.str, raw: t.bytes}
	}
	return value{kind: tEOF}
}

func dump(v value, indent string, path string) {
	switch v.kind {
	case tOpenDict:
		fmt.Printf("%s<<  (path=%s)\n", indent, path)
		for _, e := range v.dict {
			subpath := path + "/" + e.key
			fmt.Printf("%s  /%s ", indent, e.key)
			dumpInline(e.val, indent+"  ", subpath)
		}
		fmt.Printf("%s>>\n", indent)
	case tOpenArr:
		fmt.Printf("%s[  (path=%s, n=%d)\n", indent, path, len(v.arr))
		for i, e := range v.arr {
			subpath := fmt.Sprintf("%s[%d]", path, i)
			fmt.Printf("%s  ", indent)
			dumpInline(e, indent+"  ", subpath)
		}
		fmt.Printf("%s]\n", indent)
	case tName:
		fmt.Printf("%s/%s\n", indent, v.str)
	case tNumber:
		fmt.Printf("%s%g\n", indent, v.num)
	case tBool:
		fmt.Printf("%s%v\n", indent, v.bv)
	case tString:
		fmt.Printf("%s%s\n", indent, v.str)
	}
}

func dumpInline(v value, indent string, path string) {
	switch v.kind {
	case tOpenDict:
		fmt.Println()
		dump(v, indent, path)
	case tOpenArr:
		fmt.Println()
		dump(v, indent, path)
	case tName:
		fmt.Printf("/%s\n", v.str)
	case tNumber:
		fmt.Printf("%g\n", v.num)
	case tBool:
		fmt.Printf("%v\n", v.bv)
	case tString:
		fmt.Printf("%s\n", v.str)
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: parse_btdk file.aep layer_filter")
		os.Exit(2)
	}
	p, err := aep.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	filter := os.Args[2]
	for _, c := range p.Compositions {
		// Only the comp with most layers (newest)
		_ = c
	}
	// Find layer matching filter — prefer the comp where it exists most recently.
	var found []byte
	for _, c := range p.Compositions {
		for _, l := range c.Layers {
			if l.TextSourceRaw == nil || !strings.Contains(l.Name, filter) {
				continue
			}
			found = l.TextSourceRaw
		}
	}
	if found == nil {
		fmt.Fprintln(os.Stderr, "no matching layer")
		os.Exit(1)
	}
	// Find the btdk LIST inside the btds payload: it's a sub-LIST chunk.
	idx := bytes.Index(found, []byte("btdk"))
	if idx < 0 {
		fmt.Fprintln(os.Stderr, "no btdk in payload")
		os.Exit(1)
	}
	// LIST header is 4 bytes before "btdk" (the 4-byte size). Take size from
	// 4 bytes before "btdk" header start; payload starts after "btdk".
	listIDStart := idx - 8 // 4 (LIST id) + 4 (size)
	if listIDStart < 0 {
		fmt.Fprintln(os.Stderr, "bad btdk position")
		os.Exit(1)
	}
	size := binary.BigEndian.Uint32(found[listIDStart+4 : listIDStart+8])
	payloadStart := idx + 4
	payloadEnd := payloadStart + int(size) - 4 // size includes the 4-byte formType
	if payloadEnd > len(found) {
		payloadEnd = len(found)
	}
	body := found[payloadStart:payloadEnd]
	fmt.Fprintf(os.Stderr, "btdk payload: %d bytes, starts %q\n", len(body), body[:30])
	l := &lexer{d: body}
	// Top-level is an implicit dict: /key value /key value... until EOF.
	top := value{kind: tOpenDict}
	for {
		nt := l.next()
		if nt.kind == tEOF {
			break
		}
		if nt.kind != tName {
			continue
		}
		vt := l.next()
		v := parseValue(l, &vt)
		top.dict = append(top.dict, dictEntry{key: nt.str, val: v})
	}
	dump(top, "", "")
}
