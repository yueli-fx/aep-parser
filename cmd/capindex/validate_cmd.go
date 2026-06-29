package main

import (
	"fmt"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/apidoc"
)

// runValidate validates the parsed entries against the apidoc schema. A converted
// symbol (Ann != nil) is held to the full schema in either mode; an un-converted
// write-surface symbol passes in warn mode and fails only in strict mode. It also
// runs the jargon blocklist over every converted symbol's annotation prose.
func runValidate(entries []Entry, surface *publicSurface, ctx apidoc.Context, mode apidoc.Mode) []error {
	var errs []error
	for _, e := range entries {
		if e.parseErr != nil {
			errs = append(errs, fmt.Errorf("%s %s: %v", e.Pos, symbolName(e), e.parseErr))
			continue
		}
		sym := apidoc.Symbol{
			Name: symbolName(e), Pos: e.Pos, Params: e.Params,
			ReturnsNonError: e.ReturnsNonError, WriteSurface: surface.requires(e),
		}
		ann := e.Ann
		if ann == nil {
			ann = &apidoc.Annotation{} // un-converted: HasTags false
		}
		errs = append(errs, apidoc.Validate(ann, sym, ctx, mode)...)
		if e.Ann != nil {
			for _, v := range apidoc.LintJargon(jargonText(e.Ann)) {
				errs = append(errs, fmt.Errorf("%s %s: jargon token %q (pattern %s)", sym.Pos, sym.Name, v.Token, v.Pattern))
			}
		}
	}
	return errs
}

func symbolName(e Entry) string {
	if e.Recv != "" {
		return strings.TrimPrefix(e.Recv, "*") + "." + e.Symbol
	}
	return e.Symbol
}

// jargonText joins the human-facing annotation prose for the jargon lint
// (machine fields like @domain/@gate are not prose).
func jargonText(a *apidoc.Annotation) string {
	parts := []string{a.Summary, a.Description, a.Returns, a.Boundary}
	for _, p := range a.Params {
		parts = append(parts, p.Desc)
	}
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, "\n")
}
