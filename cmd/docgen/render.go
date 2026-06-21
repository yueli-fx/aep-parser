package main

import (
	"fmt"
	"strings"

	"github.com/example/aep-parser/internal/apidoc"
)

// renderType 渲单个 docType 为 markdown（不含文件级 version header / includes）。
func renderType(t *docType) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s object\n", t.name)
	if p := renderProse(t.doc); p != "" {
		fmt.Fprintf(&b, "\n%s\n", p)
	}

	if len(t.attributes) > 0 {
		b.WriteString("\n## Attributes\n")
		for _, a := range t.attributes {
			renderSymbol(&b, t.name, a)
		}
	}
	if len(t.methods) > 0 {
		b.WriteString("\n## Methods\n")
		for _, m := range t.methods {
			renderSymbol(&b, t.name, m)
		}
	}
	if len(t.consts) > 0 {
		b.WriteString("\n## Constants\n")
		for _, c := range t.consts {
			if p := renderProse(c.doc); p != "" {
				fmt.Fprintf(&b, "\n%s\n", p)
			}
			fmt.Fprintf(&b, "\n```go\n%s\n```\n", strings.TrimSpace(c.code))
		}
	}
	return b.String()
}

// renderFuncs 渲一个 "## Functions" 段（包级函数：签名 + prose + 可选 Example）。
func renderFuncs(funcs []symbol) string {
	var b strings.Builder
	b.WriteString("## Functions\n")
	for _, f := range funcs {
		fmt.Fprintf(&b, "\n### %s\n\n```go\n%s\n```\n", f.name, f.signature)
		if f.summary != "" {
			fmt.Fprintf(&b, "\n%s\n", f.summary)
		}
		if p := renderProse(f.doc); p != "" {
			fmt.Fprintf(&b, "\n%s\n", p)
		}
		renderParams(&b, f.params)
		if f.returns != "" {
			fmt.Fprintf(&b, "\n**Returns:** %s\n", f.returns)
		}
		for _, ex := range f.examples {
			label := "Example"
			if ex.suffix != "" {
				label = "Example (" + ex.suffix + ")"
			}
			fmt.Fprintf(&b, "\n**%s:**\n\n```go\n%s\n```\n", label, ex.code)
		}
	}
	return b.String()
}

// renderParams emits a markdown parameter table (nothing when params is empty).
func renderParams(b *strings.Builder, params []apidoc.Param) {
	if len(params) == 0 {
		return
	}
	b.WriteString("\n| Parameter | Description |\n|---|---|\n")
	for _, p := range params {
		fmt.Fprintf(b, "| `%s` | %s |\n", p.Name, p.Desc)
	}
}

func renderSymbol(b *strings.Builder, typeName string, s symbol) {
	fmt.Fprintf(b, "\n### %s.%s\n\n", typeName, s.name)
	if s.kind == kindField {
		fmt.Fprintf(b, "```go\n%s\n```\n", s.fieldDecl)
	} else {
		fmt.Fprintf(b, "```go\n%s\n```\n", s.signature)
	}
	if s.summary != "" {
		fmt.Fprintf(b, "\n%s\n", s.summary)
	}
	if p := renderProse(s.doc); p != "" {
		fmt.Fprintf(b, "\n%s\n", p)
	}
	renderParams(b, s.params)
	if s.returns != "" {
		fmt.Fprintf(b, "\n**Returns:** %s\n", s.returns)
	}
	if s.kind == kindField || s.kind == kindGetter {
		rw := "read-only"
		if s.readWrite {
			rw = "read-write"
		}
		if s.jsonName != "" {
			fmt.Fprintf(b, "\nJSON: `%s` · %s\n", s.jsonName, rw)
		} else {
			fmt.Fprintf(b, "\n%s\n", rw)
		}
	}
	for _, ex := range s.examples {
		label := "Example"
		if ex.suffix != "" {
			label = "Example (" + ex.suffix + ")"
		}
		fmt.Fprintf(b, "\n**%s:**\n\n```go\n%s\n```\n", label, ex.code)
	}
}
