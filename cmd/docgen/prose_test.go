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

func TestRenderProse_TidiesEscapesAndLists(t *testing.T) {
	in := "Value rules:\n  - 1D property: pass float64\n  - multi-component: pass []float64 (Position_0/1/2)"
	out := renderProse(in)
	// over-escaped markdown punctuation is reversed
	if strings.Contains(out, `\[`) || strings.Contains(out, `\_`) {
		t.Fatalf("escapes not reversed: %q", out)
	}
	if !strings.Contains(out, "[]float64") || !strings.Contains(out, "Position_0/1/2") {
		t.Fatalf("source spelling not restored: %q", out)
	}
	// top-level list markers are flush-left, not 2-space indented
	for ln := range strings.SplitSeq(out, "\n") {
		if strings.HasPrefix(ln, "  - ") {
			t.Fatalf("list item still indented: %q", ln)
		}
	}
	if !strings.Contains(out, "\n- 1D property:") && !strings.HasPrefix(out, "- 1D property:") {
		// list should render with flush-left markers somewhere
		if !strings.Contains(out, "- 1D property:") {
			t.Fatalf("flush-left list marker missing: %q", out)
		}
	}
}
