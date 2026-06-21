package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/example/aep-parser/internal/apidoc"
)

// machineTagLine reports whether a comment line is a machine @-field (@gate /
// @incident / @domain / @stability / @verify / @since / @alias). Those carry
// test identifiers, frozen enum tokens, or CJK search aliases — not lintable
// English prose — so the comment jargon lint skips them.
func machineTagLine(text string) bool {
	t := strings.TrimSpace(text)
	for _, tag := range []string{"@gate", "@incident", "@domain", "@stability", "@verify", "@since", "@alias"} {
		if strings.HasPrefix(t, tag) {
			return true
		}
	}
	return false
}

// lintRepoComments walks the non-test Go sources under root/internal and runs the
// jargon blocklist over every comment line, returning one error per violation.
// It skips _test.go files (test-comment hygiene is out of scope), the jargon
// blocklist definition itself (it names the forbidden tokens), lines carrying a
// //nolint:jargon escape, and machine @-field lines.
func lintRepoComments(root string) ([]error, error) {
	base := filepath.Join(root, "internal")
	var errs []error
	err := filepath.WalkDir(base, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		slash := filepath.ToSlash(path)
		if strings.HasSuffix(slash, "internal/apidoc/jargon.go") {
			return nil // the blocklist source names the forbidden tokens by design
		}
		fset := token.NewFileSet()
		f, perr := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if perr != nil {
			return perr
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		for _, cg := range f.Comments {
			for _, c := range cg.List {
				if strings.Contains(c.Text, "//nolint:jargon") {
					continue
				}
				line := strings.TrimPrefix(c.Text, "//")
				if machineTagLine(line) {
					continue
				}
				for _, v := range apidoc.LintJargon(line) {
					pos := fset.Position(c.Pos())
					errs = append(errs, fmt.Errorf("%s:%d jargon token %q (pattern %s)", rel, pos.Line, v.Token, v.Pattern))
				}
			}
		}
		return nil
	})
	return errs, err
}
