package main

import (
	"go/doc/comment"
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
	return strings.TrimSpace(string(pr.Markdown(parsed)))
}
