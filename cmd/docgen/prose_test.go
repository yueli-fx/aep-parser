package main

import (
	"strings"
	"testing"
)

func TestRenderProse_Markdown(t *testing.T) {
	in := "Widget 是一个示例 widget。\n\n第二段。"
	out := renderProse(in)
	if !strings.Contains(out, "Widget 是一个示例 widget。") {
		t.Fatalf("missing para1: %q", out)
	}
	if !strings.Contains(out, "第二段。") {
		t.Fatalf("missing para2: %q", out)
	}
	if !strings.Contains(out, "\n\n") {
		t.Fatalf("paragraphs not separated: %q", out)
	}
}
