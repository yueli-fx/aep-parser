package main

import (
	"go/doc/comment"
	"regexp"
	"strings"
)

// renderProse 把 Go doc-comment 文本解析成结构再渲成 markdown。
// doc comment 不是 markdown：必须经 comment.Parser/Printer 转换（无原生表格）。
func renderProse(text string) string {
	if strings.TrimSpace(text) == "" {
		return ""
	}
	var p comment.Parser
	parsed := p.Parse(text)
	var pr comment.Printer
	md := string(pr.Markdown(parsed))
	return strings.TrimSpace(tidyMarkdown(md))
}

// mdEscapeBackslash matches the comment printer's defensive escapes of
// markdown-significant punctuation in plain prose. go/doc/comment escapes
// `[` (doc-link bracket), `_` / `*` (emphasis), and `` ` `` (code span) even
// inside ordinary text, so `[]float64` renders as `\[]float64` and
// `Position_0` as `Position\_0`. None of our doc comments rely on markdown
// emphasis, so reversing these escapes is safe and restores the source spelling.
var mdEscapeBackslash = regexp.MustCompile("\\\\([_\\[\\]*`])")

// mdListItem matches a list-item line the printer indents by two spaces
// (top-level `  - x` / `  1. x`). The source comments write flush-left lists,
// so we strip the printer's two-space lead to match.
var mdListItem = regexp.MustCompile(`^  ([-*+]|\d+\.) `)

// tidyMarkdown post-processes comment.Printer.Markdown output to match the
// hand-authored docs style: reverse the over-escaping and flush-left top-level
// list markers. Leaves indented code-block continuations untouched.
func tidyMarkdown(md string) string {
	md = mdEscapeBackslash.ReplaceAllString(md, "$1")
	lines := strings.Split(md, "\n")
	for i, ln := range lines {
		if mdListItem.MatchString(ln) {
			lines[i] = ln[2:]
		}
	}
	return strings.Join(lines, "\n")
}
