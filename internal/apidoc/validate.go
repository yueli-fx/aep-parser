package apidoc

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Mode controls strictness during the migration window.
type Mode int

const (
	// ModeWarn skips completeness for un-converted (no-@tag) symbols so CI stays
	// green during migration. A converted symbol is always fully validated.
	ModeWarn Mode = iota
	// ModeStrict additionally flags an un-converted write-surface symbol as
	// missing its @tag annotation — used at the flip.
	ModeStrict
)

// Symbol carries the signature facts Validate cross-checks an Annotation against.
type Symbol struct {
	Name            string   // for error messages
	Pos             string   // "file:line" for error messages
	Params          []string // non-receiver parameter names, in order
	ReturnsNonError bool     // signature returns >= 1 non-error value
	WriteSurface    bool     // capindex requires() predicate result
}

// Context carries external facts for the gate/incident crosschecks.
type Context struct {
	Gates        map[string]bool // test name -> unconditionally-skipped
	IncidentsDir string          // incidents directory (for @incident existence)
}

var sinceRe = regexp.MustCompile(`^AE\d{4}$`)

// Validate checks ann against the schema for sym. An un-converted symbol (no
// @tag) yields no errors in ModeWarn (or a single "missing @tag" in ModeStrict
// for a write-surface symbol); a converted symbol is held to the full schema in
// both modes. Each error is prefixed "file:line symbol:".
func Validate(ann *Annotation, sym Symbol, ctx Context, mode Mode) []error {
	var errs []error
	add := func(format string, args ...any) {
		prefix := fmt.Sprintf("%s %s: ", sym.Pos, sym.Name)
		errs = append(errs, fmt.Errorf(prefix+format, args...))
	}

	if !ann.HasTags {
		if mode == ModeStrict && sym.WriteSurface {
			add("missing @tag annotation")
		}
		return errs
	}

	// Completeness (write surface only).
	if sym.WriteSurface {
		for _, m := range []struct {
			ok   bool
			name string
		}{
			{ann.Summary != "", "@summary"},
			{ann.Domain != "", "@domain"},
			{ann.Stability != "", "@stability"},
			{ann.Verify != "", "@verify"},
			{ann.Since != "", "@since"},
		} {
			if !m.ok {
				add("missing %s", m.name)
			}
		}
	}

	// Enums.
	if ann.Domain != "" && !has(Domains, ann.Domain) {
		add("invalid @domain %q", ann.Domain)
	}
	if ann.Stability != "" && !has(Stabilities, ann.Stability) {
		add("invalid @stability %q", ann.Stability)
	}
	if ann.Verify != "" && !has(Verifies, ann.Verify) {
		add("invalid @verify %q", ann.Verify)
	}
	if ann.Since != "" && (!sinceRe.MatchString(ann.Since) || !has(SinceVersions, ann.Since)) {
		add("invalid @since %q (want one of %v)", ann.Since, SinceVersions)
	}

	// @summary format.
	if ann.Summary != "" {
		if strings.Contains(ann.Summary, "\n") {
			add("@summary must be a single line")
		}
		if n := len([]rune(ann.Summary)); n > 80 {
			add("@summary exceeds 80 chars (%d)", n)
		}
		if strings.HasSuffix(ann.Summary, ".") {
			add("@summary must not end with a period")
		}
	}

	validateParams(ann, sym, add)

	// @returns presence (signature returns a non-error value).
	if sym.ReturnsNonError && ann.Returns == "" {
		add("missing @returns (signature returns a non-error value)")
	}

	// @gate presence (always) + existence (only when the caller supplies the
	// test set — docgen validates well-formedness with a nil Gates map and leaves
	// the gate/incident crosscheck to capindex).
	if (ann.Verify == "ae-accept" || ann.Verify == "render-pixel") && len(ann.Gate) == 0 {
		add("@verify %s requires @gate", ann.Verify)
	}
	if ctx.Gates != nil {
		for _, g := range ann.Gate {
			skipped, ok := ctx.Gates[g]
			if !ok {
				add("@gate %q not found in any *_test.go", g)
			} else if skipped {
				add("@gate %q is unconditionally skipped (disabled)", g)
			}
		}
	}

	// @incident existence.
	for _, slug := range ann.Incident {
		if ctx.IncidentsDir != "" && !incidentExists(ctx.IncidentsDir, slug) {
			add("@incident %q not found under incidents/", slug)
		}
	}

	return errs
}

func validateParams(ann *Annotation, sym Symbol, add func(string, ...any)) {
	want := map[string]bool{}
	for _, p := range sym.Params {
		want[p] = true
	}
	got := map[string]bool{}
	for _, p := range ann.Params {
		got[p.Name] = true
		if !want[p.Name] {
			add("@param %q is not a parameter of the signature", p.Name)
			continue
		}
		switch {
		case strings.TrimSpace(p.Desc) == "":
			add("@param %q has an empty description", p.Name)
		case strings.Contains(p.Desc, "TODO"):
			add("@param %q description is a TODO placeholder", p.Name)
		case len(strings.Fields(p.Desc)) < 3:
			add("@param %q description must be >= 3 words", p.Name)
		}
	}
	for _, p := range sym.Params {
		if !got[p] {
			add("missing @param %q", p)
		}
	}
}

func incidentExists(dir, slug string) bool {
	candidates := []string{
		filepath.Join(dir, slug+".md"),
		filepath.Join(filepath.Dir(dir), "archive", "incidents", slug+".md"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}
